package grpcmiddleware

//nolint:staticcheck
import (
	"context"
	"encoding/json"
	"path"

	grpcLogging "github.com/grpc-ecosystem/go-grpc-middleware/logging"
	grpcZap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/underbek/examples-go/logger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// CustomPayloadUnaryClientInterceptor returns a new unary client interceptor that logs the payloads of requests and responses.
// func is like original from https://github.com/grpc-ecosystem/go-grpc-middleware/blob/master/logging/zap/payload_interceptors.go#L59
// but there is only difference at logger
func CustomPayloadUnaryClientInterceptor(log *zap.Logger, decider grpcLogging.ClientPayloadLoggingDecider) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if !decider(ctx, method) {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		logEntry := log.With(newClientLoggerFields(method)...)
		// ATTENTION!!! here is the difference from original function (is didn't have data from context at logger, now it have)
		logEntry = logEntry.With(logger.GetFieldsFromContext(ctx)...)
		logProtoMessageAsJson(logEntry, req, "grpc.request.content", "client request payload logged as grpc.request.content")

		requestMetadata, _ := metadata.FromOutgoingContext(ctx)
		metadataCopy := requestMetadata.Copy()

		if meta := logger.ParseCtxMeta(ctx); meta != nil {
			data, err := json.Marshal(meta)
			if err != nil {
				log.Check(zapcore.ErrorLevel, "json.Marshal meta for grpc headers").Write(zap.Error(err))
			}

			metadataCopy.Set(logger.Meta, string(data))
		}

		ctx = metadata.NewOutgoingContext(ctx, metadataCopy)

		err := invoker(ctx, method, req, reply, cc, opts...)
		if err == nil {
			logProtoMessageAsJson(logEntry, reply, "grpc.response.content", "client response payload logged as grpc.response.content")
		}
		return err
	}
}

func newClientLoggerFields(fullMethodString string) []zapcore.Field {
	service := path.Dir(fullMethodString)[1:]
	method := path.Base(fullMethodString)
	return []zapcore.Field{
		grpcZap.SystemField,
		grpcZap.ClientField,
		zap.String("grpc.service", service),
		zap.String("grpc.method", method),
	}
}
