package grpcmiddleware

import (
	"context"

	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcZap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/underbek/examples-go/logger"
	trace "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func UnaryInterceptors(logger *logger.Logger) grpc.ServerOption {
	grpcPrometheus.EnableHandlingTimeHistogram()

	return grpc.UnaryInterceptor(
		grpcMiddleware.ChainUnaryServer(
			trace.UnaryServerInterceptor(
				trace.WithPropagators(
					propagation.NewCompositeTextMapPropagator(
						propagation.TraceContext{},
						propagation.Baggage{},
					),
				),
			),

			grpcRecovery.UnaryServerInterceptor(),
			grpcPrometheus.UnaryServerInterceptor,

			grpcZap.UnaryServerInterceptor(
				logger.Named("grpc-middleware").Internal().(*zap.Logger),
				grpcZap.WithLevels(func(code codes.Code) zapcore.Level {
					return zapcore.DebugLevel
				}),
			),
			traceIDToLoggerCtxInterceptor(),
			grpcZap.PayloadUnaryServerInterceptor(
				logger.Named("grpc-payload").Internal().(*zap.Logger),
				func(ctx context.Context, fullMethodName string, servingObject interface{}) bool {
					return logger.Internal().(*zap.Logger).Core().Enabled(zapcore.DebugLevel)
				},
			),
		),
	)
}
