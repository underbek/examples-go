package grpcclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	grpcZap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	loggerPkg "github.com/underbek/examples-go/logger"
	"github.com/underbek/examples-go/metrics"
	mw "github.com/underbek/examples-go/middlewares/grpc"
	"github.com/underbek/examples-go/middlewares/grpc/grpccache"
	redisPkg "github.com/underbek/examples-go/storage/redis"
	trace "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type Option = func([]grpc.UnaryClientInterceptor) []grpc.UnaryClientInterceptor

func NewConnection(logger *loggerPkg.Logger, cfg Config, options ...Option) (*grpc.ClientConn, error) {
	var transportCredentials credentials.TransportCredentials
	if cfg.WithTls {
		transportCredentials = credentials.NewTLS(&tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: false,
		})
	} else {
		transportCredentials = insecure.NewCredentials()
	}

	opts := []grpc.DialOption{
		// nolint:staticcheck
		grpc.WithStreamInterceptor(trace.StreamClientInterceptor()),
		// nolint:staticcheck
		grpc.WithUnaryInterceptor(trace.UnaryClientInterceptor()),
		grpc.WithTransportCredentials(transportCredentials),
		grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`),
	}

	unaries := []grpc.UnaryClientInterceptor{
		mw.CustomPayloadUnaryClientInterceptor(
			logger.Named("grpc-client-payload").Internal().(*zap.Logger),
			func(ctx context.Context, fullMethodName string) bool {
				return cfg.ShowPayloadLogs &&
					logger.Internal().(*zap.Logger).Core().Enabled(zapcore.DebugLevel)
			},
		),
		grpcZap.UnaryClientInterceptor(
			logger.Named("grpc-client").Internal().(*zap.Logger),
			grpcZap.WithLevels(func(code codes.Code) zapcore.Level {
				return zapcore.DebugLevel
			}),
			grpcZap.WithMessageProducer(func(ctx context.Context, msg string, level zapcore.Level, code codes.Code, err error, duration zapcore.Field) {
				ctxzap.Extract(ctx).With(loggerPkg.GetFieldsFromContext(ctx)...).Check(level, msg).Write(
					zap.Error(err),
					zap.String("grpc.code", code.String()),
					duration,
				)
			}),
		),
	}

	for _, fn := range options {
		unaries = fn(unaries)
	}

	opts = append(opts, grpc.WithChainUnaryInterceptor(unaries...))

	conn, err := grpc.Dial(
		cfg.DSN,
		opts...,
	)

	if err != nil {
		return nil, fmt.Errorf("create connnection: %w", err)
	}

	return conn, nil
}

func WithCacheOption(rCli redisPkg.Storage, enabled bool, ttl time.Duration, logger *loggerPkg.Logger) Option {
	return func(interceptors []grpc.UnaryClientInterceptor) []grpc.UnaryClientInterceptor {
		if enabled {
			return append(interceptors, grpccache.NewCache(rCli, ttl, logger).UnaryClientInterceptor())
		}

		return interceptors
	}
}

func WithMetricsOption() Option {
	return func(interceptors []grpc.UnaryClientInterceptor) []grpc.UnaryClientInterceptor {
		grpcPrometheus.EnableClientHandlingTimeHistogram(grpcPrometheus.WithHistogramBuckets(metrics.DefBuckets))

		return append(interceptors, grpcPrometheus.UnaryClientInterceptor)
	}
}
