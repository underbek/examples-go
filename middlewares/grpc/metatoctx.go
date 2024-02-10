package grpcmiddleware

import (
	"context"
	"encoding/json"

	"github.com/underbek/examples-go/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func metaToLoggerCtxInterceptor(log *logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		requestMetadata, _ := metadata.FromIncomingContext(ctx)
		metadataCopy := requestMetadata.Copy()

		metaData := metadataCopy.Get(logger.Meta)

		if len(metaData) != 0 {
			var meta map[string]string

			if errM := json.Unmarshal([]byte(metaData[0]), &meta); errM == nil {
				ctx = logger.AddCtxMetaValues(ctx, meta)
			}
		}

		return handler(ctx, req)
	}
}
