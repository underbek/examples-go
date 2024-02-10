package pgx

import "time"

type Config struct {
	DSN     string        `env:"POSTGRES_DSN" valid:"required"`
	Timeout time.Duration `env:"POSTGRES_QUERY_TIMEOUT" envDefault:"5s"`
}
