package testcontainer

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// RedisContainer wraps testcontainers.Container with extra methods.
type (
	RedisContainer struct {
		testcontainers.Container
		Config RedisContainerConfig
	}

	RedisContainerOption func(c *RedisContainerConfig)

	RedisContainerConfig struct {
		ImageTag   string
		Password   string
		MappedPort string
		Host       string
	}
)

func (c RedisContainer) GetHost() string {
	return c.Config.Host + ":" + c.Config.MappedPort
}

func (c RedisContainer) GetPassword() string {
	return c.Config.Password
}

func WithRedisTag(tag string) RedisContainerOption {
	return func(c *RedisContainerConfig) {
		c.ImageTag = tag
	}
}

// NewRedisContainer creates and starts a Redis container.
func NewRedisContainer(ctx context.Context, opts ...RedisContainerOption) (*RedisContainer, error) {
	registryCred()

	const (
		redisImage = "redis"
		redisPort  = "6379"
	)

	// Define container ENVs
	config := RedisContainerConfig{
		ImageTag: "7.0.5",
		Password: "password",
	}

	for _, opt := range opts {
		opt(&config)
	}

	containerPort := fmt.Sprintf("%s/tcp", redisPort)

	// Build testcontainer request
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Env: map[string]string{
				"REDIS_PASSWORD": config.Password,
			},
			ExposedPorts: []string{
				containerPort,
				//"6379:6379", // for local development, do not remove it
			},
			Image: fmt.Sprintf("%s:%s", redisImage, config.ImageTag),
			Cmd:   []string{"/bin/sh", "-c", "redis-server --requirepass $REDIS_PASSWORD"},
			WaitingFor: wait.ForAll(
				wait.ForExec([]string{"redis-cli", "-a", config.Password, "PING"}).
					WithPollInterval(time.Millisecond*100).
					WithExitCodeMatcher(func(exitCode int) bool {
						return exitCode == 0
					}),
				wait.ForLog(".*Ready to accept connections.*").AsRegexp(),
			),
		},
		Started: true,
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("getting request provider: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting host for: %w", err)
	}

	// Get mapped port for 5432/tcp
	mappedPort, err := container.MappedPort(ctx, nat.Port(containerPort))
	if err != nil {
		return nil, fmt.Errorf("getting mapped port for (%s): %w", containerPort, err)
	}
	config.MappedPort = mappedPort.Port()
	config.Host = host

	fmt.Println("Redis run by host:", config.Host, config.MappedPort)

	return &RedisContainer{
		Container: container,
		Config:    config,
	}, nil
}
