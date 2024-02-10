package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
	"github.com/underbek/examples-go/logger"
	"go.opentelemetry.io/otel"
)

type HeadersCarrier []kafka.Header

func (k HeadersCarrier) Get(key string) string {
	for _, h := range k {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

func (k *HeadersCarrier) Set(key string, value string) {
	*k = append(*k, kafka.Header{
		Key:   key,
		Value: []byte(value),
	})
}

func (k HeadersCarrier) Keys() []string {
	r := make([]string, len(k))
	for _, h := range k {
		r = append(r, h.Key)
	}

	return r
}

func injectKafkaHeaders(ctx context.Context, log *logger.Logger) []kafka.Header {
	h := make(HeadersCarrier, 0)
	otel.GetTextMapPropagator().Inject(ctx, &h)

	if meta := logger.ParseCtxMeta(ctx); meta != nil {
		data, err := json.Marshal(meta)
		if err != nil {
			log.WithCtx(ctx).WithError(err).Error("json.Marshal meta for kafka headers")
		}

		h.Set(logger.Meta, string(data))
	}

	return h
}
