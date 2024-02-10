package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"github.com/underbek/examples-go/logger"
	"github.com/underbek/examples-go/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
)

type Handler func(ctx context.Context, msg kafka.Message) error

type Consumer interface {
	Consume(ctx context.Context, handler Handler) error
	Close() error
}

type consumer struct {
	logger              *logger.Logger
	client              *kafka.Reader
	manualReties        int
	manualRetryDuration time.Duration
	metrics             consumerMetrics
}

func NewConsumer(logger *logger.Logger, cfg ConsumerConfig) (Consumer, error) {
	dialer := &kafka.Dialer{
		ClientID:  cfg.AppName,
		Timeout:   cfg.ConnTimeout,
		DualStack: true,
	}
	if err := checkConn(dialer, cfg.Brokers); err != nil {
		return nil, err
	}
	if cfg.ManualRetries == 0 {
		cfg.ManualRetries = 1
	}

	if cfg.ManualRetryDuration == 0 {
		cfg.ManualRetryDuration = time.Minute
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		GroupID:        cfg.GroupID,
		Topic:          cfg.Topic,
		GroupTopics:    cfg.GroupTopics,
		CommitInterval: cfg.CommitInterval,
		Dialer:         dialer,
	})

	logger.Info("kafka connected successfully")

	return &consumer{
		client:              reader,
		manualReties:        cfg.ManualRetries,
		manualRetryDuration: cfg.ManualRetryDuration,
		logger:              logger,
		metrics:             newConsumerMetrics(cfg),
	}, nil
}

// Close waits for all writes to complete and then gracefully closes the connection
func (c *consumer) Close() error {
	return c.client.Close()
}

func (c *consumer) Consume(ctx context.Context, handler Handler) error {
	gr, ctx := errgroup.WithContext(ctx)
	gr.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				err := c.run(ctx, handler)
				if errors.Is(err, continueError) {
					continue
				}

				if err != nil {
					return err
				}
			}
		}
	})

	return gr.Wait()
}

var continueError = errors.New("continue")

func (c *consumer) run(ctx context.Context, handler Handler) error {
	defer func() {
		if r := recover(); r != nil {
			c.logger.WithCtx(ctx).
				With("panic", r).
				With("trace", string(debug.Stack())).
				Error(fmt.Sprintf("Recovered from kafka panic: %v", r))
		}
	}()

	msg, err := c.client.FetchMessage(ctx)
	if err != nil {
		c.logger.WithError(err).Error("fetch message failed")

		c.metrics.incError("fetch", msg)

		time.Sleep(c.manualRetryDuration)

		return continueError
	}

	c.metrics.incMessage(msg).observeLatency(msg)

	carrier := HeadersCarrier(msg.Headers)

	msgCtx := otel.GetTextMapPropagator().Extract(ctx, &carrier)
	msgCtx, span := tracing.StartCustomSpan(msgCtx, trace.SpanKindConsumer, "kafka", "FetchMessage",
		trace.WithAttributes(attribute.String("topic", msg.Topic)))
	defer span.End()

	if data := carrier.Get(logger.Meta); len(data) != 0 {
		var meta map[string]string

		if errM := json.Unmarshal([]byte(data), &meta); errM == nil {
			msgCtx = logger.AddCtxMetaValues(msgCtx, meta)
		}
	}

	c.logger.WithCtx(msgCtx).
		With("topic", msg.Topic).
		With("partition", msg.Partition).
		With("time", msg.Time).
		With("offset", msg.Offset).
		With("headers", msg.Headers).
		With("key", string(msg.Key)).
		With("value", string(msg.Value)).
		Debug("read message from kafka")

	if err = handler(msgCtx, msg); err != nil {
		c.logger.WithCtx(msgCtx).
			WithError(err).
			Error("consume with handler failed")

		c.metrics.incError("handle", msg)

		time.Sleep(c.manualRetryDuration)

		return continueError
	}

	err = c.commit(msgCtx, msg)

	if err != nil {
		return err
	}

	return nil
}

func (c *consumer) commit(ctx context.Context, msg kafka.Message) error {
	for i := 0; i < c.manualReties; i++ {
		if err := c.client.CommitMessages(ctx, msg); err != nil {
			c.logger.WithCtx(ctx).
				With("topic", msg.Topic).
				With("partition", msg.Partition).
				With("offset", msg.Offset).
				With("key", string(msg.Key)).
				WithError(err).
				Error("commit message failed")

			c.metrics.incError("commit", msg)

			time.Sleep(c.manualRetryDuration)
			continue
		}

		c.logger.WithCtx(ctx).
			With("topic", msg.Topic).
			With("partition", msg.Partition).
			With("offset", msg.Offset).
			With("key", string(msg.Key)).
			Debug("commit message successfully")

		return nil
	}

	return fmt.Errorf("commit message failed after %d retries", c.manualReties)
}

// checkConn checks whether the connection can be established with the given configuration
func checkConn(dialer *kafka.Dialer, brokers []string) error {
	if len(brokers) < 1 {
		return errors.New("empty brokers list")
	}

	_, err := dialer.Dial("tcp", brokers[0])
	if err != nil {
		return err
	}

	return nil
}
