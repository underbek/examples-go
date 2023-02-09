package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

type JaegerConfig struct {
	Endpoint       string `env:"JAEGER_ADDRESS"`
	Enabled        bool   `env:"JAEGER_ENABLED" envDefault:"false"`
	ServiceName    string `env:"APP_NAME" envDefault:"app"`
	ServiceVersion string `env:"SERVICE_VERSION"`
}

type Provider struct {
	provider trace.TracerProvider
}

func NewProvider(config JaegerConfig) (*Provider, error) {
	if !config.Enabled {
		return &Provider{
			provider: trace.NewNoopTracerProvider(),
		}, nil
	}

	endpoint := jaeger.WithEndpoint(config.Endpoint)

	collection := jaeger.WithCollectorEndpoint(
		endpoint,
	)

	exp, err := jaeger.New(collection)
	if err != nil {
		return nil, fmt.Errorf("failed to create Jaeger exporter: %w", err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(config.ServiceName),
			semconv.ServiceVersionKey.String(config.ServiceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	return &Provider{
		provider: provider,
	}, nil
}

type shutdownable interface {
	Shutdown(ctx context.Context) error
}

// Shutdown shuts down the span processors
func (p Provider) Shutdown(ctx context.Context) error {
	if p.provider == nil {
		return nil
	}

	prv, ok := p.provider.(shutdownable)
	if !ok {
		return nil
	}

	if err := prv.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown tracing provider: %w", err)
	}

	return nil
}

func Tracer(name string) trace.Tracer {
	return otel.GetTracerProvider().Tracer(name)
}
