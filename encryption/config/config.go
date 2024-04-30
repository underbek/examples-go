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
	Vault      VaultENV `envPrefix:"VAULT"`
	Pool       PoolENV  `envPrefix:"POOL"`
}

type VaultENV struct {
	DSN       string `env:"_DSN" valid:"required"`
	TokenPath string `env:"_TOKEN_PATH" valid:"required"`
	PoolSize  uint   `env:"_POOL_SIZE" envDefault:"4"`
}

type PoolENV struct {
	PoolMode      bool          `env:"_MODE" envDefault:"true"`
	RetryAttempts uint          `env:"_RETRY_ATTEMPTS"  envDefault:"5"`
	RetryDuration time.Duration `env:"_RETRY_DURATION" envDefault:"1s"`
}
