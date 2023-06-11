package grpcmiddleware

import (
	"context"
	"strings"

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

func showHealthCheckConditions(path string, showHealthCheck bool) bool {
	//if param showHealthCheck is true or full requested method does not contain grpc health check path
	return showHealthCheck || !strings.Contains(path, "grpc.health.v1.Health")
}

func UnaryInterceptors(logger *logger.Logger, showHealthCheck, showPayloadLogs bool) grpc.ServerOption {
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
				grpcZap.WithDecider(func(fullMethodName string, err error) bool {
					return showHealthCheckConditions(fullMethodName, showHealthCheck)
				}),
			),
			traceIDToLoggerCtxInterceptor(),
			grpcZap.PayloadUnaryServerInterceptor(
				logger.Named("grpc-payload").Internal().(*zap.Logger),
				func(ctx context.Context, fullMethodName string, servingObject interface{}) bool {
					return showPayloadLogs &&
						showHealthCheckConditions(fullMethodName, showHealthCheck) &&
						logger.Internal().(*zap.Logger).Core().Enabled(zapcore.DebugLevel)
				},
			),
		),
	)
}
