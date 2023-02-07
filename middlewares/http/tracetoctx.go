package httpmiddleware

import (
	"net/http"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/underbek/examples-go/logger"
	"github.com/underbek/examples-go/tracing"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func JaegerTraceMiddleware(next http.Handler, cfg tracing.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cfg.Enabled {
			newCtx, span := tracing.StartCustomSpan(r.Context(), trace.SpanKindServer, "httpserver", "jaegerTraceMW")
			defer span.End()

			newCtx = logger.AddCtxValue(newCtx, "traceID", span.SpanContext().TraceID().String())
			ctxzap.AddFields(newCtx, zap.Any("traceID", span.SpanContext().TraceID()))

			r = r.WithContext(newCtx)
		}

		next.ServeHTTP(w, r)
	})
}
