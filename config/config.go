package config

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type App struct {
	Name        string `env:"APP_NAME" envDefault:"app"`
	MetricsPort int    `env:"METRICS_PORT" envDefault:"8877"`
	Debug       bool   `env:"DEBUG" envDefault:"false"`
}

type HTTPServer struct {
	Port         int           `env:"HTTP_SERVER_PORT" envDefault:"8181"`
	WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"15s"`
	ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"15s"`
}

type Storage struct {
	DSN string `env:"POSTGRES_DSN" valid:"required"`
}

type Redis struct {
	Addr       string `env:"REDIS_ADDRESS" valid:"required"`
	TLSEnabled bool   `env:"REDIS_ENABLE_TLS" envDefault:"true"`
	Password   string `env:"REDIS_PASSWORD"`
}

func New[T any]() (T, error) {
	var cfg T

	if envFilePath, ok := os.LookupEnv("ENV_FILE_PATH"); ok {
		if err := godotenv.Load(envFilePath); err != nil {
			return cfg, err
		}
	}

	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}

	v := reflect.ValueOf(&cfg).Elem()

	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).CanSet() {
			return cfg, fmt.Errorf("config is not provide unexported fields")
		}

		data := reflect.New(v.Field(i).Type())

		if data.Elem().Kind() != reflect.Struct {
			continue
		}

		if err := env.Parse(data.Interface()); err != nil {
			return cfg, err
		}

		v.Field(i).Set(data.Elem())
	}

	if ok, err := govalidator.ValidateStruct(cfg); !ok {
		return cfg, err
	}

	return cfg, nil
}
