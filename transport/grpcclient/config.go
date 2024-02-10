package grpcclient

type Config struct {
	ShowPayloadLogs bool   `env:"SHOW_PAYLOAD_LOGS" envDefault:"true"`
	DSN             string `env:"_DSN" valid:"required"`
	WithTls         bool   `env:"_WITH_TLS" envDefault:"false"`
}
