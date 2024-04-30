package config

import (
	"time"

	goKitConfig "github.com/underbek/examples-go/config"
	"github.com/underbek/examples-go/storage/pgx"
	"github.com/underbek/examples-go/tracing"
	"github.com/underbek/examples-go/transport/grpcserver"
)

type Config struct {
	goKitConfig.App
	Storage    pgx.Config
	GRPCServer grpcserver.Config
	Jaeger     tracing.Config

	LimitServiceConfig
	Scheduler          Scheduler          `envPrefix:"SCHEDULER_"`
	StorageTransaction StorageTransaction `envPrefix:"POSTGRES_TRANSACTION_"`
}

type StorageTransaction struct {
	RetryAmount int           `env:"RETRY_AMOUNT" envDefault:"5"`
	RetryDelay  time.Duration `env:"RETRY_DELAY" envDefault:"1s"`
}

type LimitServiceConfig struct {
	MaxLimit int `env:"MAX_LIMIT" envDefault:"500"`
}

type Scheduler struct {
	Cleanup Cleanup `envPrefix:"CLEANUP_"`
}

type Cleanup struct {
	RunInterval     time.Duration `env:"RUN_INTERVAL" envDefault:"24h"`
	OutdateInterval time.Duration `env:"OUTDATE_INTERVAL" envDefault:"24h"`
}
