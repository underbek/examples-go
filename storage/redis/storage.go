package redis

import (
	"context"
	"crypto/tls"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Addr               string `env:"REDIS_ADDRESS" valid:"required"`
	TLSEnabled         bool   `env:"REDIS_ENABLE_TLS" envDefault:"true"`
	Password           string `env:"REDIS_PASSWORD"`
	DB                 int    `env:"REDIS_DB" envDefault:"0"`
	InsecureSkipVerify bool   `env:"REDIS_INSECURE_SKIP_VERIFY" envDefault:"false"`
}

type Storage redis.UniversalClient

type Option = func(pool *redis.Client, st Storage) Storage

func New(ctx context.Context, cfg Config, opts ...Option) (Storage, error) {
	rOpts := &redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	if cfg.TLSEnabled {
		rOpts.TLSConfig = &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: cfg.InsecureSkipVerify, // nolint: gosec
		}
	}

	rCli := redis.NewClient(rOpts)

	var st Storage = rCli

	for _, opt := range opts {
		st = opt(rCli, st)
	}

	err := st.Ping(ctx).Err()
	if err != nil {
		return nil, err
	}

	return st, nil
}
