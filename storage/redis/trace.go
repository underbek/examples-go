package redis

import (
	"github.com/redis/go-redis/v9"
)

type DBTracer struct {
	Storage
}

func WithTrace() Option {
	return func(_ *redis.Client, st Storage) Storage {
		return &DBTracer{
			Storage: st,
		}
	}
}
