package grpccache

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/underbek/examples-go/logger"
	redisPkg "github.com/underbek/examples-go/storage/redis"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"time"
)

var ErrMarshall = errors.New("error while marshaling")

type service struct {
	cli    redisPkg.Storage
	ttl    time.Duration
	logger *logger.Logger
}

func NewCache(cli redisPkg.Storage, ttl time.Duration, logger *logger.Logger) *service {
	return &service{cli: cli, ttl: ttl, logger: logger}
}

func (s *service) UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, reqRaw, replyRaw interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = logger.AddCtxValue(ctx, "method", method)
		s.logger.WithCtx(ctx).Debug("grpcCache: start trying get from cache")

		req, ok := reqRaw.(proto.Message)
		if !ok {
			s.logger.WithCtx(ctx).Error("grpcCache: request is not proto message")

			return invoker(ctx, method, reqRaw, replyRaw, cc, opts...)
		}

		s.logger.WithCtx(ctx).Debug("grpcCache: req: proto.Marshal")

		reqMsg, err := proto.Marshal(req)
		if err != nil {
			s.logger.WithCtx(ctx).WithError(err).Error("grpcCache: proto.Marshal: request")

			return invoker(ctx, method, reqRaw, replyRaw, cc, opts...)
		}

		key := makeKey(method, hex.EncodeToString(reqMsg))

		s.logger.WithCtx(ctx).Debug("grpcCache: cli.Get")

		cmd := s.cli.Get(ctx, key)
		dataRaw, err := cmd.Bytes()
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				s.logger.WithCtx(ctx).WithError(err).Error("grpcCache: get data from cache error, invoke next")
			}

			s.logger.WithCtx(ctx).Debug("grpcCache: entity not found")

			errInv := invoker(ctx, method, reqRaw, replyRaw, cc, opts...)
			if errInv != nil {
				return errInv
			}

			if errStore := s.storeData(ctx, key, replyRaw); errStore != nil {
				s.logger.WithCtx(ctx).WithError(errStore).Error("grpcCache: store data error")
			}

			s.logger.WithCtx(ctx).Debug("grpcCache: entity cached")

			return nil
		}

		s.logger.WithCtx(ctx).Debug("grpcCache: hex.Decode")

		data := make([]byte, hex.DecodedLen(len(dataRaw)))

		_, err = hex.Decode(data, dataRaw)
		if err != nil {
			s.logger.WithCtx(ctx).WithError(err).Error("grpcCache: hex.Decode")

			errInv := invoker(ctx, method, reqRaw, replyRaw, cc, opts...)
			if errInv != nil {
				return errInv
			}
		}

		s.logger.WithCtx(ctx).Debug("grpcCache: resp: proto.Marshal")

		reply, ok := replyRaw.(proto.Message)
		if !ok {
			s.logger.WithCtx(ctx).WithError(err).Error("grpcCache: response is not proto message")

			if cmd := s.cli.Set(ctx, key, nil, 0); cmd.Err() != nil {
				s.logger.WithCtx(ctx).WithError(err).Error("grpcCache: storeData")
			}

			errInv := invoker(ctx, method, reqRaw, replyRaw, cc, opts...)
			if errInv != nil {
				return errInv
			}

			if errStore := s.storeData(ctx, key, replyRaw); errStore != nil {
				s.logger.WithCtx(ctx).WithError(errStore).Error("grpcCache: storeData")
			}

			return nil
		}

		s.logger.WithCtx(ctx).Debug("grpcCache: resp: unmarshal")

		err = proto.Unmarshal(data, reply)
		if err != nil {
			s.logger.WithCtx(ctx).WithError(err).Error("grpcCache: cached data is corrupt")

			err := invoker(ctx, method, reqRaw, replyRaw, cc, opts...)
			if err != nil {
				return err
			}

			if err := s.storeData(ctx, key, replyRaw); err != nil {
				s.logger.WithCtx(ctx).WithError(err).Error("grpcCache: storeData")
			}
		}

		s.logger.WithCtx(ctx).Debug("grpcCache: entity found, return")

		return nil
	}
}

func (s *service) storeData(ctx context.Context, key string, replyRaw interface{}) error {
	s.logger.WithCtx(ctx).Debug("grpcCache: store: storing data")

	reply, ok := replyRaw.(proto.Message)
	if !ok {
		return ErrMarshall
	}

	s.logger.WithCtx(ctx).Debug("grpcCache: store: proto.Marshal")

	replyMsg, err := proto.Marshal(reply)
	if err != nil {
		return errors.Wrap(err, "proto.Marshal")
	}

	s.logger.WithCtx(ctx).Debug("grpcCache: store: cli.SetNX")

	if cmd := s.cli.SetNX(ctx, key, hex.EncodeToString(replyMsg), s.ttl); cmd.Err() != nil {
		return errors.Wrap(err, "cli.SetNX")
	}

	s.logger.WithCtx(ctx).Debug("grpcCache: store: success")

	return nil
}

func makeKey(method, data string) string {
	return fmt.Sprintf("grpccache:%s:%s", method, data)
}
