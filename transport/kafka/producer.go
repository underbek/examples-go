package kafka

import (
	"context"
	"fmt"
	"net"

	"github.com/segmentio/kafka-go"
	"github.com/underbek/examples-go/logger"
	"github.com/underbek/examples-go/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Producer interface {
	Publish(ctx context.Context, msg kafka.Message) error
	Close() error
}

type producer struct {
	logger  *logger.Logger
	writer  *kafka.Writer
	metrics producerMetrics
}

func NewProducer(logger *logger.Logger, cfg ProducerConfig) (Producer, error) {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Balancer:               &kafka.CRC32Balancer{},
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
		Completion:             cfg.Completion,
		Compression:            cfg.Compression,
		AllowAutoTopicCreation: cfg.AllowAutoTopicCreation,
		Transport: &kafka.Transport{
			ClientID: cfg.AppName,
			Dial: (&net.Dialer{
				Timeout: cfg.ConnTimeout,
			}).DialContext,
		},
	}

	dialer := &kafka.Dialer{
		ClientID:  cfg.AppName,
		Timeout:   cfg.ConnTimeout,
		DualStack: true,
	}

	var conn *kafka.Conn
	var err error

	if cfg.AllowManualTopic {
		conn, err = dialer.DialContext(
			context.Background(),
			"tcp",
			cfg.Brokers[0],
		)
	} else {
		writer.Topic = cfg.Topic
		conn, err = dialer.DialLeader(
			context.Background(),
			"tcp",
			cfg.Brokers[0],
			cfg.Topic,
			0,
		)
	}

	if err != nil {
		return nil, err
	}

	if err = conn.Close(); err != nil {
		return nil, err
	}

	return &producer{
		logger:  logger,
		writer:  writer,
		metrics: newProducerMetrics(cfg),
	}, nil
}

func (p *producer) Close() error {
	return p.writer.Close()
}

func (p *producer) Publish(ctx context.Context, msg kafka.Message) error {
	ctx, span := tracing.StartCustomSpan(ctx, trace.SpanKindProducer, "kafka", "WriteMessages",
		trace.WithAttributes(attribute.String("topic", msg.Topic)))
	defer span.End()

	h := injectKafkaHeaders(ctx, p.logger)
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

		p.metrics.incError(msg)

		return fmt.Errorf("publish message: %w", err)
	}

	p.metrics.incMessage(msg)

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
