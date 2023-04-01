package rabbitmq

import (
	"context"
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/underbek/examples-go/logger"
	"github.com/underbek/examples-go/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
)

type HandleFunc = func(ctx context.Context, msg amqp.Delivery)

type Consumer interface {
	Consume(ctx context.Context, cfg Consume, handler HandleFunc) error
	Close() error
}

type consumer struct {
	logger    *logger.Logger
	channel   *amqp.Channel
	queueName string
}

func NewConsumer(logger *logger.Logger, conn Connection, qd QueueDeclare, qb QueueBind) (Consumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	_, err = ch.QueueDeclare(
		qd.Queue,
		qd.Durable,
		qd.AutoDelete,
		qd.Exclusive,
		qd.NoWait,
		qd.Arguments,
	)
	if err != nil {
		return nil, fmt.Errorf("decalre queue: %w", err)
	}

	logger.With("QueueDeclare", qd).Info("queue declared")

	err = ch.QueueBind(qb.Queue, qb.RoutingKey, qb.Exchange, qb.NoWait, qb.Arguments)
	if err != nil {
		return nil, fmt.Errorf("bind queue: %w", err)
	}
	logger.With("QueueBind", qb).Info("queue bound")

	return &consumer{
		logger:    logger,
		channel:   ch,
		queueName: qd.Queue,
	}, nil
}

func (c *consumer) Consume(ctx context.Context, cfg Consume, handler HandleFunc) error {
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

				msgCtx := otel.GetTextMapPropagator().Extract(ctx, AmqpHeadersCarrier(msg.Headers))
				msgCtx, span := tracing.StartCustomSpan(msgCtx, trace.SpanKindConsumer, "rabbitmq", "Read",
					trace.WithAttributes(attribute.String("queue", c.queueName)))

				c.logger.WithCtx(msgCtx).
					With("message", msg).
					With("queue", c.queueName).
					Debug("message consumed from rabbitmq")

				handler(msgCtx, msg)

				span.End()
			}
		}
	})

	return gr.Wait()
}

func (c *consumer) Close() error {
	return c.channel.Close()
}
