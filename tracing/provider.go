package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	agent     = "agent"
	collector = "collector"
)

type Config struct {
	JaegerUploader string `env:"JAEGER_UPLOADER" envDefault:"agent"`
	Endpoint       string `env:"JAEGER_ADDRESS"`
	AgentHost      string `env:"JAEGER_AGENT_HOST"`
	AgentPort      string `env:"JAEGER_AGENT_PORT"`
	Enabled        bool   `env:"JAEGER_ENABLED" envDefault:"false"`
	ServiceName    string `env:"APP_NAME" envDefault:"app"`
	ServiceVersion string `env:"SERVICE_VERSION"`
}

type Provider struct {
	provider trace.TracerProvider
}

func NewProvider(config Config) (*Provider, error) {
	if !config.Enabled {
		return &Provider{
			// nolint:staticcheck
			provider: trace.NewNoopTracerProvider(),
		}, nil
	}

	var endpointOption jaeger.EndpointOption
	switch config.JaegerUploader {
	case agent:
		endpointOption = jaeger.WithAgentEndpoint(
			jaeger.WithAgentHost(config.AgentHost),
			jaeger.WithAgentPort(config.AgentPort),
		)
	case collector:
		endpointOption = jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(config.Endpoint),
		)
	default:
		return nil, fmt.Errorf("undefined jaeger uploader type: \"%s\"", config.JaegerUploader)
	}

	exp, err := jaeger.New(endpointOption)
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

	provider := sdkTrace.NewTracerProvider(
		sdkTrace.WithSampler(sdkTrace.AlwaysSample()),
		sdkTrace.WithBatcher(exp),
		sdkTrace.WithResource(res),
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
