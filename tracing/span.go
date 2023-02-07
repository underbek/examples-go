package tracing

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

func StartCustomSpan( // nolint:ireturn
	ctx context.Context,
	kind trace.SpanKind,
	pkgName, methodName string,
) (context.Context, trace.Span) {
	tracer := Tracer(pkgName)

	return tracer.Start(ctx, methodName, trace.WithSpanKind(kind))
}

func StartSpan(ctx context.Context, pkgName, methodName string) (context.Context, trace.Span) { //nolint:ireturn
	tracer := Tracer(pkgName)
	return tracer.Start(ctx, methodName, trace.WithSpanKind(trace.SpanKindInternal))
}

func GetSpanContextFromContext(ctx context.Context) trace.SpanContext {
	return trace.SpanContextFromContext(ctx)
}
