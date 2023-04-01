package testcontainer

import (
	"context"
	"fmt"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// RabbitMQContainer wraps testcontainers.Container with extra methods
type (
	RabbitMQContainer struct {
		testcontainers.Container
		Config RabbitMQContainerConfig
	}

	RabbitMQContainerOption func(c *RabbitMQContainerConfig)

	RabbitMQContainerConfig struct {
		ImageTag   string
		User       string
		Password   string
		MappedPort string
		Host       string
	}
)

// GetDSN returns AMQP connection URL.
func (c RabbitMQContainer) GetDSN() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/",
		c.Config.User,
		c.Config.Password,
		c.Config.Host,
		c.Config.MappedPort)
}

func WithRMQTag(tag string) RabbitMQContainerOption {
	return func(c *RabbitMQContainerConfig) {
		c.ImageTag = tag
	}
}

// NewRabbitMQContainer creates and starts a RabbitMQ container.
func NewRabbitMQContainer(ctx context.Context, opts ...RabbitMQContainerOption) (*RabbitMQContainer, error) {
	const (
		rmqImage = "heidiks/rabbitmq-delayed-message-exchange"
		rmqPort  = "5672"
	)

	// Define container ENVs
	config := RabbitMQContainerConfig{
		ImageTag: "3.10.2-management",
		User:     "guest",
		Password: "guest",
	}
	for _, opt := range opts {
		opt(&config)
	}

	containerPort := rmqPort + "/tcp"

	// Build testcontainer request
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			ExposedPorts: []string{
				containerPort,
				//"5672:5672",   // for local development, do not remove it pls
				//"15672:15672", // for local development, do not remove it pls
			},
			Image:      fmt.Sprintf("%s:%s", rmqImage, config.ImageTag),
			WaitingFor: wait.ForListeningPort(nat.Port(containerPort)),
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

	mappedPort, err := container.MappedPort(ctx, nat.Port(containerPort))
	if err != nil {
		return nil, fmt.Errorf("getting mapped port for (%s): %w", containerPort, err)
	}
	config.MappedPort = mappedPort.Port()
	config.Host = host

	return &RabbitMQContainer{
		Container: container,
		Config:    config,
	}, nil
}
