package testcontainer

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/phayes/freeport"
	gokafka "github.com/segmentio/kafka-go"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// KafkaContainer wraps testcontainers.Container with extra methods.
type (
	KafkaContainer struct {
		testcontainers.Container
		Config KafkaContainerConfig
	}

	KafkaContainerOption func(c *KafkaContainerConfig)

	KafkaContainerConfig struct {
		ImageTag     string
		MappedPort   string
		Host         string
		StartTimeout time.Duration
		PollInterval time.Duration
	}
)

func (c KafkaContainer) GetBrokers() []string {
	return []string{c.Config.Host + ":" + c.Config.MappedPort}
}

func WithKafkaTag(tag string) KafkaContainerOption {
	return func(c *KafkaContainerConfig) {
		c.ImageTag = tag
	}
}

func WithKafkaStartTimeout(timeout time.Duration) KafkaContainerOption {
	return func(c *KafkaContainerConfig) {
		c.StartTimeout = timeout
	}
}

func WithKafkaPollInterval(interval time.Duration) KafkaContainerOption {
	return func(c *KafkaContainerConfig) {
		c.PollInterval = interval
	}
}

// NewKafkaContainer creates and starts a Kafka container.
func NewKafkaContainer(ctx context.Context, opts ...KafkaContainerOption) (*KafkaContainer, error) {
	registryCred()

	const (
		image = "krisgeus/docker-kafka"
	)

	// Define container ENVs
	config := KafkaContainerConfig{
		ImageTag: "latest",
	}
	for _, opt := range opts {
		opt(&config)
	}

	if config.StartTimeout == 0 {
		config.StartTimeout = time.Minute
	}

	if config.PollInterval == 0 {
		config.PollInterval = time.Millisecond * 500
	}

	port, err := freeport.GetFreePort()
	if err != nil {
		return nil, err
	}

	containerPort := fmt.Sprintf("%d/tcp", port)

	// Build testcontainer request
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Env: map[string]string{
				"ADVERTISED_LISTENERS":  fmt.Sprintf("PLAINTEXT://localhost:%d,INTERNAL://localhost:9093", port),
				"LISTENERS":             fmt.Sprintf("PLAINTEXT://0.0.0.0:%d,INTERNAL://0.0.0.0:9093", port),
				"SECURITY_PROTOCOL_MAP": "PLAINTEXT:PLAINTEXT,INTERNAL:PLAINTEXT",
				"INTER_BROKER":          "INTERNAL",
				"AUTO_CREATE_TOPICS":    "true",
			},
			ExposedPorts: []string{
				containerPort,
				fmt.Sprintf("%d:%d", port, port),
			},
			Image: fmt.Sprintf("%s:%s", image, config.ImageTag),
			WaitingFor: wait.ForAll(
				wait.ForLog(".*kafka entered RUNNING state.*").AsRegexp(),
				wait.ForListeningPort(nat.Port(containerPort)),
				wait.ForNop(func(ctx context.Context, target wait.StrategyTarget) error {
					testcontainers.Logger.Printf("⏰ Start internal kafka check")

					host, errN := target.Host(ctx)
					if errN != nil {
						return errN
					}

					port, errN := target.MappedPort(ctx, nat.Port(containerPort))
					if errN != nil {
						return errN
					}

					testcontainers.Logger.Printf("⚠ Kafka port: %v", port.Port())

					address := fmt.Sprintf("%s:%s", host, port.Port())

					ticker := time.NewTicker(config.PollInterval)
					defer ticker.Stop()

					for {
						select {
						case <-ctx.Done():
							return ctx.Err()
						case <-ticker.C:
							_, errN = gokafka.DialLeader(ctx, "tcp", address, "startup-tmp-topic", 0)
							if errN != nil {
								continue
							}

							return nil
						}
					}

				}).WithStartupTimeout(config.StartTimeout),
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

	mappedPort, err := container.MappedPort(ctx, nat.Port(containerPort))
	if err != nil {
		return nil, fmt.Errorf("getting mapped port for (%s): %w", containerPort, err)
	}

	config.MappedPort = mappedPort.Port()
	config.Host = host

	return &KafkaContainer{
		Container: container,
		Config:    config,
	}, nil
}
