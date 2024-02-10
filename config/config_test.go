package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/underbek/examples-go/transport/grpcclient"
)

func Test_ParseConfig(t *testing.T) {
	type CustomConfig struct {
		Custom string `env:"CUSTOM" envDefault:"custom"`
	}

	type Config struct {
		CustomConfig
		AppConfig App
	}

	cfg, err := New[Config]()
	assert.NoError(t, err)
	assert.Equal(t, "custom", cfg.Custom)
	assert.Equal(t, "app", cfg.AppConfig.Name)
}

func Test_FieldInBaseConfig(t *testing.T) {
	type Config struct {
		App
		Field string `env:"FIELD" envDefault:"field"`
	}

	cfg, err := New[Config]()
	assert.NoError(t, err)
	assert.Equal(t, "field", cfg.Field)
	assert.Equal(t, "app", cfg.Name)
}

func Test_Validation(t *testing.T) {
	type CustomConfig struct {
		CheckValidation string `env:"VALIDATION" valid:"required"`
	}

	type Config struct {
		CustomConfig
	}

	_, err := New[Config]()
	assert.Error(t, err)
}

func Test_PrivateValueInConfig(t *testing.T) {
	type Config struct {
		appConfig App
	}

	err := os.Setenv("APP_NAME", "TEST-APP")
	require.NoError(t, err)
	defer os.Clearenv()

	cfg, err := New[Config]()
	require.NoError(t, err)
	assert.Empty(t, cfg.appConfig.Name)
}

func Test_ParseByFile(t *testing.T) {
	envFilePath := "/tmp/test_env_file"
	data := []byte("APP_NAME=test\nDEBUG=true")
	err := os.WriteFile(envFilePath, data, 0600)
	require.NoError(t, err)

	defer func() {
		err = os.Remove(envFilePath)
		require.NoError(t, err)
	}()

	err = os.Setenv("ENV_FILE_PATH", envFilePath)
	require.NoError(t, err)
	defer os.Clearenv()

	type Config struct {
		App
	}

	cfg, err := New[Config]()
	require.NoError(t, err)
	assert.Equal(t, "test", cfg.Name)
	assert.True(t, cfg.Debug)
}

func Test_GRPCClients(t *testing.T) {
	type Config struct {
		ClientA grpcclient.Config `envPrefix:"SERVICE_A"`
		ClientB grpcclient.Config `envPrefix:"SERVICE_B"`
	}

	err := os.Setenv("SERVICE_A_DSN", "core.dsn")
	require.NoError(t, err)
	err = os.Setenv("SERVICE_A_WITH_TLS", "true")
	require.NoError(t, err)
	err = os.Setenv("SERVICE_B_DSN", "cs.dsn")
	require.NoError(t, err)

	defer os.Clearenv()

	cfg, err := New[Config]()
	require.NoError(t, err)

	assert.Equal(t, "core.dsn", cfg.ClientA.DSN)
	assert.True(t, cfg.ClientA.WithTls)
	assert.Equal(t, "cs.dsn", cfg.ClientB.DSN)
	assert.False(t, cfg.ClientB.WithTls)
}
