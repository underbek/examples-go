package grpcclient

//nolint:staticcheck
import (
	"bytes"
	"context"
	"fmt"
	"path"

	"github.com/golang/protobuf/proto"
	grpc_logging "github.com/grpc-ecosystem/go-grpc-middleware/logging"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

// CustomPayloadUnaryClientInterceptor returns a new unary client interceptor that logs the payloads of requests and responses.
// func is like original from https://github.com/grpc-ecosystem/go-grpc-middleware/blob/master/logging/zap/payload_interceptors.go#L59
// but there is only difference at logger
func CustomPayloadUnaryClientInterceptor(logger *zap.Logger, decider grpc_logging.ClientPayloadLoggingDecider) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if !decider(ctx, method) {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		logEntry := logger.With(newClientLoggerFields(method)...)
		// ATTENTION!!! here is the difference from original function (is didn't have data from context at logger, now it have)
		logEntry = logEntry.With(ctxzap.TagsToFields(ctx)...)
		logProtoMessageAsJson(logEntry, req, "grpc.request.content", "client request payload logged as grpc.request.content")
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
		grpc_zap.SystemField,
		grpc_zap.ClientField,
		zap.String("grpc.service", service),
		zap.String("grpc.method", method),
	}
}

func logProtoMessageAsJson(logger *zap.Logger, pbMsg interface{}, key string, msg string) {
	if p, ok := pbMsg.(proto.Message); ok {
		logger.Check(zapcore.DebugLevel, msg).Write(zap.Object(key, &jsonpbObjectMarshaler{pb: p}))
	}
}

type jsonpbObjectMarshaler struct {
	pb proto.Message
}

func (j *jsonpbObjectMarshaler) MarshalLogObject(e zapcore.ObjectEncoder) error {
	// ZAP jsonEncoder deals with AddReflect by using json.MarshalObject. The same thing applies for consoleEncoder.
	return e.AddReflected("msg", j)
}

func (j *jsonpbObjectMarshaler) MarshalJSON() ([]byte, error) {
	b := &bytes.Buffer{}
	if err := grpc_zap.JsonPbMarshaller.Marshal(b, j.pb); err != nil {
		return nil, fmt.Errorf("jsonpb serializer failed: %v", err)
	}
	return b.Bytes(), nil
}
