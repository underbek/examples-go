package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/underbek/examples-go/logger"
	"github.com/underbek/examples-go/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
)

type HandleFunc = func(context.Context, amqp.Delivery)

type Consumer struct {
	logger    *logger.Logger
	channel   Channel
	queueName string
	metrics   metrics
}

func NewConsumer(
	l *logger.Logger,
	ch Channel,
	queue string,
	enableMetrics bool,
) *Consumer {
	return &Consumer{
		logger:    l,
		channel:   ch,
		queueName: queue,
		metrics:   newConsumerMetrics(queue, enableMetrics),
	}
}

func (c *Consumer) Consume(ctx context.Context, cfg Consume, handler HandleFunc) error {
	deliveryCh, err := c.channel.Consume(
		c.queueName,
		cfg.ConsumerTag,
		cfg.NoAck,
		cfg.Exclusive,
		cfg.NoLocal,
		cfg.NoWait,
		cfg.Arguments,
	)
	if err != nil {
		c.logger.WithCtx(ctx).WithError(err).Error("consume message")
		return fmt.Errorf("consume message: %w", err)
	}

	gr, ctx := errgroup.WithContext(ctx)

	gr.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()

			case msg, ok := <-deliveryCh:
				if !ok {
					return errors.New("delivery channel was closed")
				}

				c.metrics.observeLatency(msg)

				msgCtx, span := tracing.StartCustomSpan(
					parseAMQPHeaders(ctx, msg.Headers),
					trace.SpanKindConsumer, "rabbitmq", "Read",
					trace.WithAttributes(attribute.String("queue", c.queueName)),
				)

				c.logger.WithCtx(msgCtx).
					With("rmq_consumed_message", msg).
					With("queue", c.queueName).
					Debug("message consumed from rabbitmq")

				c.recoverHandler(msgCtx, handler, msg)

				span.End()
			}
		}
	})

	return gr.Wait()
}

func (c *Consumer) recoverHandler(ctx context.Context, handler HandleFunc, msg amqp.Delivery) {
	defer func() {
		if r := recover(); r != nil {
			c.logger.WithCtx(ctx).
				With("rabbitmq_message", msg).
				With("panic", r).
				With("trace", string(debug.Stack())).
				Error(fmt.Sprintf("Recovered from rabbitmq panic: %v", r))
		}
	}()

	handler(ctx, msg)
}
