package grpcmiddleware

import (
	"context"

	grpcCtxTags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/underbek/examples-go/tracing"
	"go.opentelemetry.io/otel/trace"
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
			// // adds traceID and spanID to the request payload
			tags := grpcCtxTags.NewTags()
			tags.Set(tracing.TraceID, span.SpanContext().TraceID())
			tags.Set(tracing.SpanID, span.SpanContext().SpanID())

			ctx = grpcCtxTags.SetInContext(ctx, tags)
		}
		return handler(ctx, req)
	}
}
