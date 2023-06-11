package redis

import (
	"context"
	"crypto/tls"

	"github.com/go-redis/redis/v8"
)

type Config struct {
	Addr               string `env:"REDIS_ADDRESS" valid:"required"`
	TLSEnabled         bool   `env:"REDIS_ENABLE_TLS" envDefault:"true"`
	Password           string `env:"REDIS_PASSWORD"`
	DB                 int    `env:"REDIS_DB" envDefault:"0"`
	InsecureSkipVerify bool   `env:"REDIS_INSECURE_SKIP_VERIFY" envDefault:"false"`
}

type storage struct {
	*redis.Client
}

func New(ctx context.Context, cfg Config) (redis.Cmdable, error) {
	opts := &redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	if cfg.TLSEnabled {
		opts.TLSConfig = &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: cfg.InsecureSkipVerify, // nolint: gosec
		}
	}

	rCli := redis.NewClient(opts)

	err := rCli.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}

	return &storage{
		rCli,
	}, nil
}
