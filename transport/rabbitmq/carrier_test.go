package rabbitmq

import (
	"context"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
	"github.com/underbek/examples-go/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
)

func Test_injectAMQPHeaders(t *testing.T) {
	l, err := logger.New(true)
	require.NoError(t, err)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.Baggage{},
		),
	)

	tests := []struct {
		name        string
		ctxProvider func() context.Context
		want        map[string]interface{}
	}{
		{
			name: "Empty context",
			ctxProvider: func() context.Context {
				return context.Background()
			},
			want: make(map[string]interface{}),
		},
		{
			name: "Context with logger meta",
			ctxProvider: func() context.Context {
				return logger.AddCtxMetaValues(context.Background(), map[string]string{
					"key1": "val1",
					"key2": "val2",
				})
			},
			want: map[string]interface{}{
				logger.Meta: `{"key1":"val1","key2":"val2"}`,
			},
		},
		{
			name: "Context with otel baggage",
			ctxProvider: func() context.Context {
				member, err := baggage.NewMember("mem_key", "mem_val")
				require.NoError(t, err)
				bag, err := baggage.New(member)
				require.NoError(t, err)

				return baggage.ContextWithBaggage(context.Background(), bag)
			},
			want: map[string]interface{}{
				"baggage": "mem_key=mem_val",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, injectAMQPHeaders(tt.ctxProvider(), l))
		})
	}
}

func Test_parseAMQPHeaders(t *testing.T) {
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.Baggage{},
		),
	)

	tests := []struct {
		name         string
		headers      amqp.Table
		wantProvider func() context.Context
	}{
		{
			name:    "Empty headers",
			headers: amqp.Table{},
			wantProvider: func() context.Context {
				return context.Background()
			},
		},
		{
			name: "Headers with logger meta",
			headers: amqp.Table{
				logger.Meta: `{"key1":"val1","key2":"val2"}`,
			},
			wantProvider: func() context.Context {
				return logger.AddCtxMetaValues(context.Background(), map[string]string{
					"key1": "val1",
					"key2": "val2",
				})
			},
		},
		{
			name: "Headers with otel baggage",
			headers: amqp.Table{
				"baggage": "mem_key=mem_val",
			},
			wantProvider: func() context.Context {
				member, err := baggage.NewMember("mem_key", "mem_val")
				require.NoError(t, err)
				bag, err := baggage.New(member)
				require.NoError(t, err)

				return baggage.ContextWithBaggage(context.Background(), bag)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run(tt.name, func(t *testing.T) {
				require.Equal(t, tt.wantProvider(), parseAMQPHeaders(context.Background(), tt.headers))
			})
		})
	}
}
