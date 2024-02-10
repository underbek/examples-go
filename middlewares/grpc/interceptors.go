package grpcmiddleware

import (
	"context"
	"runtime/debug"
	"strings"
	"time"

	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcZap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	grpcRecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	exLogger "github.com/underbek/examples-go/logger"
	trace "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func showHealthCheckConditions(path string, showHealthCheck bool) bool {
	//if param showHealthCheck is true or full requested method does not contain grpc health check path
	return showHealthCheck || !strings.Contains(path, "grpc.health.v1.Health")
}

func recoveryHandler(logger *exLogger.Logger) grpcRecovery.RecoveryHandlerFuncContext {
	return func(ctx context.Context, p interface{}) error {
		logger.WithCtx(ctx).
			With("panic", p).
			With("trace", string(debug.Stack())).
			Error("panic occurred")

		return status.Errorf(codes.Internal, "%v", p)
	}
}

func UnaryInterceptors(logger *exLogger.Logger, showHealthCheck, showPayloadLogs bool, timeout time.Duration) grpc.ServerOption {
	grpcPrometheus.EnableHandlingTimeHistogram()
	return grpc.UnaryInterceptor(
		grpcMiddleware.ChainUnaryServer(
			func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				ctxNew, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()

				return handler(ctxNew, req)
			},
			// nolint:staticcheck
			trace.UnaryServerInterceptor(
				trace.WithPropagators(
					propagation.NewCompositeTextMapPropagator(
						propagation.TraceContext{},
						propagation.Baggage{},
					),
				),
			),
			traceIDToLoggerCtxInterceptor(),
			metaToLoggerCtxInterceptor(logger),
			grpcRecovery.UnaryServerInterceptor(
				grpcRecovery.WithRecoveryHandlerContext(recoveryHandler(logger)),
			),
			grpcPrometheus.UnaryServerInterceptor,

			grpcZap.UnaryServerInterceptor(
				logger.Named("grpc-middleware").Internal().(*zap.Logger),
				grpcZap.WithLevels(func(code codes.Code) zapcore.Level {
					return zapcore.DebugLevel
				}),
				grpcZap.WithDecider(func(fullMethodName string, err error) bool {
					return showHealthCheckConditions(fullMethodName, showHealthCheck)
				}),
				grpcZap.WithMessageProducer(func(ctx context.Context, msg string, level zapcore.Level, code codes.Code, err error, duration zapcore.Field) {
					ctxzap.Extract(ctx).With(exLogger.GetMetaFieldsFromContext(ctx)...).Check(level, msg).Write(
						zap.Error(err),
						zap.String("grpc.code", code.String()),
						duration,
					)
				}),
			),
			CustomPayloadUnaryServerInterceptor(
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
