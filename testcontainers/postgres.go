package testcontainer

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/go-connections/nat"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer wraps testcontainers.Container with extra methods.
type (
	PostgresContainer struct {
		testcontainers.Container
		Config PostgresContainerConfig
	}

	PostgresContainerOption func(c *PostgresContainerConfig)

	PostgresContainerConfig struct {
		ImageTag     string
		User         string
		Password     string
		MappedPort   string
		Database     string
		Host         string
		EnvPortKey   string
		StartTimeout time.Duration
	}
)

// GetDSN returns DB connection URL.
func (c PostgresContainer) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", c.Config.User, c.Config.Password, c.Config.Host, c.Config.MappedPort, c.Config.Database)
}

func WithPostgresTag(tag string) PostgresContainerOption {
	return func(c *PostgresContainerConfig) {
		c.ImageTag = tag
	}
}

func WithPostgresDatabaseName(dbName string) PostgresContainerOption {
	return func(c *PostgresContainerConfig) {
		c.Database = dbName
	}
}

func WithPostgresStartTimeout(timeout time.Duration) PostgresContainerOption {
	return func(c *PostgresContainerConfig) {
		c.StartTimeout = timeout
	}
}

// NewPostgresContainer creates and starts a Postgres container.
func NewPostgresContainer(ctx context.Context, opts ...PostgresContainerOption) (*PostgresContainer, error) {
	registryCred()

	const (
		psqlImage = "postgres"
		psqlPort  = "5432"
	)

	// Define container ENVs
	config := PostgresContainerConfig{
		ImageTag:   "14.7",
		User:       "user",
		Password:   "password",
		Database:   "db_test",
		EnvPortKey: "TEST_POSTGRES_PORT",
	}
	for _, opt := range opts {
		opt(&config)
	}

	if config.StartTimeout == 0 {
		config.StartTimeout = time.Minute
	}

	containerPort := fmt.Sprintf("%s/tcp", psqlPort)

	// Build testcontainer request
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Env: map[string]string{
				"POSTGRES_USER":     config.User,
				"POSTGRES_PASSWORD": config.Password,
				"POSTGRES_DB":       config.Database,
			},
			ExposedPorts: []string{
				setContainerPortByEnv(containerPort, config.EnvPortKey),
			},
			Image: fmt.Sprintf("%s:%s", psqlImage, config.ImageTag),
			WaitingFor: wait.ForAll(
				wait.ForLog(".*PostgreSQL init process complete; ready for start up.*").AsRegexp(),
				wait.ForSQL(nat.Port(containerPort), "pgx", func(host string, port nat.Port) string {
					return fmt.Sprintf("postgresql://%v:%v@%s:%v/%v", config.User, config.Password, host, port.Port(), config.Database)
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

	// Get mapped port for 5432/tcp
	mappedPort, err := container.MappedPort(ctx, nat.Port(containerPort))
	if err != nil {
		return nil, fmt.Errorf("getting mapped port for (%s): %w", containerPort, err)
	}
	config.MappedPort = mappedPort.Port()
	config.Host = host

	fmt.Println("Postgres run by host:", config.Host, config.MappedPort)

	return &PostgresContainer{
		Container: container,
		Config:    config,
	}, nil
}
