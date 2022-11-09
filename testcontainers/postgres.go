package testcontainer

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/go-connections/nat"
	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/lib/pq"
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
		ImageTag   string
		User       string
		Password   string
		MappedPort string
		Database   string
		Host       string
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

// NewPostgresContainer creates and starts a Postgres container.
func NewPostgresContainer(ctx context.Context, opts ...PostgresContainerOption) (*PostgresContainer, error) {
	const (
		psqlImage = "postgres"
		psqlPort  = "5432"
	)

	// Define container ENVs
	config := PostgresContainerConfig{
		ImageTag: "11.5",
		User:     "user",
		Password: "password",
		Database: "db_test",
	}
	for _, opt := range opts {
		opt(&config)
	}

	containerPort := psqlPort + "/tcp"

	// Build testcontainer request
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Env: map[string]string{
				"POSTGRES_USER":     config.User,
				"POSTGRES_PASSWORD": config.Password,
				"POSTGRES_DB":       config.Database,
			},
			ExposedPorts: []string{
				containerPort,
				//"5432:5432", // for local development, do not remove it
			},
			Image: fmt.Sprintf("%s:%s", psqlImage, config.ImageTag),
			WaitingFor: wait.ForExec([]string{"psql", "-d", config.Database, "-U", config.User, "-c", "SELECT 1;"}).
				WithPollInterval(time.Millisecond * 100).
				WithExitCodeMatcher(func(exitCode int) bool {
					return exitCode == 0
				}),
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
