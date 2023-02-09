package grpcmiddleware

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/underbek/examples-go/logger"
	"github.com/underbek/examples-go/tracing"
	"go.uber.org/zap"

	"go.opentelemetry.io/otel/trace"

	grpcCtxTags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"google.golang.org/grpc"
)

func traceIDToLoggerCtxInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		span := trace.SpanFromContext(ctx)
		if span.IsRecording() {
			ctx = logger.AddCtxValue(ctx, tracing.TraceID, span.SpanContext().TraceID())
			// adds traceID to the request result
			ctxzap.AddFields(ctx, zap.Any(tracing.TraceID, span.SpanContext().TraceID()))
			// adds traceID to the request payload
			tags := grpcCtxTags.NewTags()
			tags.Set(tracing.TraceID, span.SpanContext().TraceID())

			ctx = grpcCtxTags.SetInContext(ctx, tags)
		}
		return handler(ctx, req)
	}
}
