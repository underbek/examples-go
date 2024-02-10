package rabbitmq

import (
	"context"
	"fmt"
	"time"

	"github.com/underbek/examples-go/logger"
	"github.com/underbek/examples-go/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/exp/maps"
)

type Producer struct {
	logger       *logger.Logger
	channel      Channel
	exchangeName string
}

func NewProducer(l *logger.Logger, ch Channel, exchange string) *Producer {
	return &Producer{
		logger:       l,
		channel:      ch,
		exchangeName: exchange,
	}
}

func (p *Producer) Publish(ctx context.Context, msg PublishMessage) error {
	ctx, span := tracing.StartCustomSpan(
		ctx, trace.SpanKindProducer, "rabbitmq", "Write",
		trace.WithAttributes(attribute.String("exchange", p.exchangeName)),
		trace.WithAttributes(attribute.String("routingKey", msg.RoutingKey)),
	)
	defer span.End()

	h := injectAMQPHeaders(ctx, p.logger)
	maps.Copy(h, msg.Message.Headers)
	msg.Message.Headers = h

	if msg.Message.Timestamp.IsZero() {
		msg.Message.Timestamp = time.Now()
	}

	if err := p.channel.PublishWithContext(
		ctx,
		p.exchangeName,
		msg.RoutingKey,
		msg.Mandatory,
		msg.Immediate,
		msg.Message,
	); err != nil {
		p.logger.WithCtx(ctx).WithError(err).Error("publish message")
		return fmt.Errorf("publish message: %w", err)
	}

	p.logger.WithCtx(ctx).
		With("rmq_published_message", msg).
		With("exchange", p.exchangeName).
		Debug("message published into rabbitmq")

	return nil
}
