package grpcclient

import (
	"context"
	"crypto/tls"
	"fmt"

	grpcZap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	trace "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/underbek/examples-go/logger"
)

type Config struct {
	ShowPayloadLogs bool   `env:"SHOW_PAYLOAD_LOGS" envDefault:"true"`
	DSN             string `env:"_DSN" valid:"required"`
	WithTLS         bool   `env:"_WITH_TLS" envDefault:"false"`
}

func NewConnection(logger *logger.Logger, cfg Config) (*grpc.ClientConn, error) {
	var transportCredentials credentials.TransportCredentials
	if cfg.WithTLS {
		transportCredentials = credentials.NewTLS(&tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: false,
		})
	} else {
		transportCredentials = insecure.NewCredentials()
	}

	opts := []grpc.DialOption{
		grpc.WithStreamInterceptor(trace.StreamClientInterceptor()),
		grpc.WithUnaryInterceptor(trace.UnaryClientInterceptor()),
		grpc.WithTransportCredentials(transportCredentials),
		grpc.WithChainUnaryInterceptor(
			grpcPrometheus.UnaryClientInterceptor,
			grpcZap.UnaryClientInterceptor(
				logger.Named("grpc-client").Internal().(*zap.Logger),
				grpcZap.WithLevels(func(code codes.Code) zapcore.Level {
					return zapcore.DebugLevel
				}),
			),
			CustomPayloadUnaryClientInterceptor(
				logger.Named("grpc-client-payload").Internal().(*zap.Logger),
				func(ctx context.Context, fullMethodName string) bool {
					return cfg.ShowPayloadLogs &&
						logger.Internal().(*zap.Logger).Core().Enabled(zapcore.DebugLevel)
				},
			),
		),
	}

	conn, err := grpc.Dial(
		cfg.DSN,
		opts...,
	)

	if err != nil {
		return nil, fmt.Errorf("create connnection: %w", err)
	}

	return conn, nil
}
