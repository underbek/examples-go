package kafka

import (
	"context"
	"fmt"
	"net"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/underbek/examples-go/logger"
	"github.com/underbek/examples-go/tracing"
)

type Producer interface {
	Publish(ctx context.Context, msg kafka.Message) error
	Close() error
}

type producer struct {
	logger *logger.Logger
	writer *kafka.Writer
}

func NewProducer(logger *logger.Logger, cfg ProducerConfig) (Producer, error) {
	conn, err := (&kafka.Dialer{
		ClientID:  cfg.AppName,
		Timeout:   cfg.ConnTimeout,
		DualStack: true,
	}).DialLeader(
		context.Background(),
		"tcp",
		cfg.Brokers[0],
		cfg.Topic,
		0,
	)
	if err != nil {
		return nil, err
	}

	if err = conn.Close(); err != nil {
		return nil, err
	}

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Topic:                  cfg.Topic,
		Balancer:               &kafka.RoundRobin{},
		MaxAttempts:            cfg.MaxAttempts,
		WriteBackoffMin:        cfg.WriteBackoffMin,
		WriteBackoffMax:        cfg.WriteBackoffMax,
		BatchSize:              cfg.BatchSize,
		BatchBytes:             cfg.BatchBytes,
		BatchTimeout:           cfg.BatchTimeout,
		ReadTimeout:            cfg.ReadTimeout,
		WriteTimeout:           cfg.WriteTimeout,
		RequiredAcks:           cfg.RequiredAcks,
		Async:                  cfg.Async,
		Compression:            cfg.Compression,
		AllowAutoTopicCreation: cfg.AllowAutoTopicCreation,
		Transport: &kafka.Transport{
			ClientID: cfg.AppName,
			Dial: (&net.Dialer{
				Timeout: cfg.ConnTimeout,
			}).DialContext,
		},
	}

	return &producer{
		logger: logger,
		writer: writer,
	}, nil
}

func (p *producer) Close() error {
	return p.writer.Close()
}

func (p *producer) Publish(ctx context.Context, msg kafka.Message) error {
	ctx, span := tracing.StartCustomSpan(ctx, trace.SpanKindProducer, "kafka", "WriteMessages",
		trace.WithAttributes(attribute.String("topic", msg.Topic)))
	defer span.End()

	h := injectKafkaHeaders(ctx)
	msg.Headers = append(h, msg.Headers...)

	err := p.writer.WriteMessages(ctx, msg)
	if err != nil {
		p.logger.WithCtx(ctx).
			With("topic", msg.Topic).
			With("partition", msg.Partition).
			With("offset", msg.Offset).
			With("key", string(msg.Key)).
			WithError(err).
			Error("publish message to kafka failed")

		return fmt.Errorf("publish message: %w", err)
	}

	p.logger.WithCtx(ctx).
		With("topic", msg.Topic).
		With("partition", msg.Partition).
		With("time", msg.Time).
		With("offset", msg.Offset).
		With("headers", msg.Headers).
		With("key", string(msg.Key)).
		With("value", string(msg.Value)).
		Debug("published message to kafka")

	return nil
}
