package tracing

import (
	trace "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
)

type (
	SpanTag struct {
		Key   string
		Value any
	}

	Params struct {
		ExtraOpts []trace.Option
	}
)

func NewParams(extraOpts ...trace.Option) Params {
	return Params{
		ExtraOpts: extraOpts,
	}
}
