package grpcserver

import "time"

type Config struct {
	ShowHealthLogs  bool          `env:"SHOW_HEALTH_LOGS" envDefault:"false"`
	ShowPayloadLogs bool          `env:"SHOW_PAYLOAD_LOGS" envDefault:"true"`
	Port            int           `env:"GRPC_SERVER_PORT" envDefault:"8080"`
	Timeout         time.Duration `env:"GRPC_SERVER_HANDLER_TIMEOUT" envDefault:"5s"`
	KeepAlive       time.Duration `env:"GRPC_SERVER_KEEPALIVE" envDefault:"60s"`
}
