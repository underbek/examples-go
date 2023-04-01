package httpmiddleware

import (
	"net/http"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/underbek/examples-go/logger"
	"github.com/underbek/examples-go/tracing"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func JaegerTraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newCtx, span := tracing.StartCustomSpan(r.Context(), trace.SpanKindServer, "httpserver", "jaegerTraceMW")
		defer span.End()
		if span.IsRecording() {
			newCtx = logger.AddCtxValue(newCtx, tracing.TraceID, span.SpanContext().TraceID().String())
			newCtx = logger.AddCtxValue(newCtx, tracing.SpanID, span.SpanContext().SpanID().String())

			ctxzap.AddFields(newCtx, zap.Any(tracing.TraceID, span.SpanContext().TraceID()))
			ctxzap.AddFields(newCtx, zap.Any(tracing.SpanID, span.SpanContext().SpanID()))

			r = r.WithContext(newCtx)
		}

		next.ServeHTTP(w, r)
	})
}
