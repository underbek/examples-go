package tracing

import (
	"context"

	"github.com/underbek/examples-go/logger"
	"go.opentelemetry.io/otel/trace"
)

const TraceID = "traceID"

func StartCustomSpan( // nolint:ireturn
	ctx context.Context,
	kind trace.SpanKind,
	pkgName, methodName string,
) (context.Context, trace.Span) {
	tracer := Tracer(pkgName)

	ctx, span := tracer.Start(ctx, methodName, trace.WithSpanKind(kind))
	if span.IsRecording() {
		ctx = logger.AddCtxValue(ctx, TraceID, span.SpanContext().TraceID())
	}

	return ctx, span
}

func StartSpan(ctx context.Context, pkgName, methodName string) (context.Context, trace.Span) { //nolint:ireturn
	tracer := Tracer(pkgName)

	ctx, span := tracer.Start(ctx, methodName, trace.WithSpanKind(trace.SpanKindInternal))
	if span.IsRecording() {
		ctx = logger.AddCtxValue(ctx, TraceID, span.SpanContext().TraceID())
	}

	return ctx, span
}

func GetSpanContextFromContext(ctx context.Context) trace.SpanContext {
	return trace.SpanContextFromContext(ctx)
}
