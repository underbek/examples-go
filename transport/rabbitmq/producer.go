package rabbitmq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/underbek/examples-go/logger"
	"github.com/underbek/examples-go/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"golang.org/x/exp/maps"
)

type Producer interface {
	Publish(ctx context.Context, msg PublishMessage) error
	Close() error
}

type producer struct {
	logger       *logger.Logger
	channel      *amqp.Channel
	exchangeName string
}

func NewProducer(logger *logger.Logger, conn Connection, ed ExchangeDeclare) (Producer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	err = ch.ExchangeDeclare(
		ed.Exchange,
		string(ed.Type),
		ed.Durable,
		ed.AutoDelete,
		ed.Internal,
		ed.NoWait,
		ed.Arguments,
	)
	if err != nil {
		return nil, fmt.Errorf("exchange declare: %w", err)
	}
	logger.With("ExchangeDeclare", ed).Info("exchange declared")

	return &producer{
		logger:       logger,
		channel:      ch,
		exchangeName: ed.Exchange,
	}, nil
}

func (p *producer) Publish(ctx context.Context, msg PublishMessage) error {
	ctx, span := tracing.StartCustomSpan(ctx, trace.SpanKindProducer, "rabbitmq", "Write",
		trace.WithAttributes(attribute.String("exchange", p.exchangeName)),
		trace.WithAttributes(attribute.String("routingKey", msg.RoutingKey)),
	)
	defer span.End()

	h := injectAMQPHeaders(ctx)
	maps.Copy(h, msg.Message.Headers)
	msg.Message.Headers = h

	err := p.channel.PublishWithContext(
		ctx,
		p.exchangeName,
		msg.RoutingKey,
		msg.Mandatory,
		msg.Immediate,
		msg.Message,
	)
	if err != nil {
		p.logger.WithCtx(ctx).WithError(err).Error("publish message")
		return fmt.Errorf("publish message: %w", err)
	}

	p.logger.WithCtx(ctx).
		With("message", msg).
		With("exchange", p.exchangeName).
		Debug("message published into rabbitmq")

	return nil
}

func (p *producer) Close() error {
	return p.channel.Close()
}
