package tracing

import (
	"context"

	"go.opentelemetry.io/otel/trace"

	"github.com/underbek/examples-go/logger"
)

const (
	TraceID = "trace-id"
	SpanID  = "span-id"
)

func StartCustomSpan( // nolint:ireturn
	ctx context.Context,
	kind trace.SpanKind,
	pkgName, methodName string,
	opts ...trace.SpanStartOption,
) (context.Context, trace.Span) {
	tracer := Tracer(pkgName)

	opts = append(opts, trace.WithSpanKind(kind))
	ctx, span := tracer.Start(ctx, methodName, opts...)
	if span.IsRecording() {
		ctx = logger.AddCtxValue(ctx, TraceID, span.SpanContext().TraceID())
		ctx = logger.AddCtxValue(ctx, SpanID, span.SpanContext().SpanID())
	}

	return ctx, span
}

func StartSpan(
	ctx context.Context,
	pkgName, methodName string,
	opts ...trace.SpanStartOption,
) (context.Context, trace.Span) { //nolint:ireturn
	tracer := Tracer(pkgName)

	opts = append(opts, trace.WithSpanKind(trace.SpanKindInternal))
	ctx, span := tracer.Start(ctx, methodName, opts...)
	if span.IsRecording() {
		ctx = logger.AddCtxValue(ctx, TraceID, span.SpanContext().TraceID())
		ctx = logger.AddCtxValue(ctx, SpanID, span.SpanContext().SpanID())
	}

	return ctx, span
}

func PutTraceInfoIntoContext(ctx context.Context, traceID [16]byte, spanID [8]byte) context.Context {
	spanCtx := trace.NewSpanContext(
		trace.SpanContextConfig{
			TraceID: traceID,
			SpanID:  spanID,
		},
	)

	ctx = trace.ContextWithSpanContext(ctx, spanCtx)

	if trace.SpanFromContext(ctx).IsRecording() {
		ctx = logger.AddCtxValue(ctx, TraceID, traceID)
		ctx = logger.AddCtxValue(ctx, SpanID, spanID)
	}

	return ctx
}

func GetSpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

func GetSpanContextFromContext(ctx context.Context) trace.SpanContext {
	return trace.SpanContextFromContext(ctx)
}
