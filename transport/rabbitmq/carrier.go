package rabbitmq

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/underbek/examples-go/logger"
	"go.opentelemetry.io/otel"
)

type AmqpHeadersCarrier map[string]interface{}

func (a AmqpHeadersCarrier) Get(key string) string {
	v, ok := a[key]
	if !ok {
		return ""
	}
	return v.(string)
}

func (a AmqpHeadersCarrier) Set(key string, value string) {
	a[key] = value
}

func (a AmqpHeadersCarrier) Keys() []string {
	i := 0
	r := make([]string, len(a))

	for k := range a {
		r[i] = k
		i++
	}

	return r
}

func injectAMQPHeaders(ctx context.Context, log *logger.Logger) map[string]interface{} {
	h := make(AmqpHeadersCarrier)
	otel.GetTextMapPropagator().Inject(ctx, h)

	if meta := logger.ParseCtxMeta(ctx); meta != nil {
		data, err := json.Marshal(meta)
		if err != nil {
			log.WithCtx(ctx).WithError(err).Error("json.Marshal meta for rabbitmq headers")
			return h
		}

		h.Set(logger.Meta, string(data))
	}

	return h
}

func parseAMQPHeaders(ctx context.Context, headers amqp.Table) context.Context {
	carrier := AmqpHeadersCarrier(headers)

	if data := carrier.Get(logger.Meta); len(data) != 0 {
		var meta map[string]string

		if errM := json.Unmarshal([]byte(data), &meta); errM == nil {
			ctx = logger.AddCtxMetaValues(ctx, meta)
		}
	}

	return otel.GetTextMapPropagator().Extract(ctx, carrier)
}
