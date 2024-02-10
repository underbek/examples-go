package httpserver

import "time"

type Config struct {
	ShowHealthLogs  bool          `env:"SHOW_HEALTH_LOGS" envDefault:"false"`
	ShowPayloadLogs bool          `env:"SHOW_PAYLOAD_LOGS" envDefault:"true"`
	Port            int           `env:"HTTP_SERVER_PORT" envDefault:"8181"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"15s"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"15s"`
}
