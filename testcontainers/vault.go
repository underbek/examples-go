package testcontainer

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/go-connections/nat"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// VaultContainer wraps testcontainers.Container with extra methods.
type (
	VaultContainer struct {
		testcontainers.Container
		Config VaultContainerConfig
	}

	VaultContainerOption func(c *VaultContainerConfig)

	VaultContainerConfig struct {
		ImageTag   string
		RootToken  string
		MappedPort string
		Host       string
	}
)

// GetDSN returns Vault connection URL.
func (c VaultContainer) GetDSN() string {
	return fmt.Sprintf("http://%s:%s", c.Config.Host, c.Config.MappedPort)
}

func (c VaultContainer) GetToken() string {
	return c.Config.RootToken
}

func WithVaultTag(tag string) VaultContainerOption {
	return func(c *VaultContainerConfig) {
		c.ImageTag = tag
	}
}

func WithVaultRootToken(token string) VaultContainerOption {
	return func(c *VaultContainerConfig) {
		c.RootToken = token
	}
}

// NewVaultContainer creates and starts a Vault container.
func NewVaultContainer(ctx context.Context, opts ...VaultContainerOption) (*VaultContainer, error) {
	registryCred()

	const (
		image = "vault"
		port  = "8200"
	)

	// Define container ENVs
	config := VaultContainerConfig{
		ImageTag:  "latest",
		RootToken: "vault-plaintext-root-token",
	}
	for _, opt := range opts {
		opt(&config)
	}

	containerPort := fmt.Sprintf("%s/tcp", port)

	// Build testcontainer request
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Env: map[string]string{
				"VAULT_ADDR":              fmt.Sprintf("http://0.0.0.0:%s", port),
				"VAULT_DEV_ROOT_TOKEN_ID": config.RootToken,
			},
			ExposedPorts: []string{
				containerPort,
				//"8200:8200", // for local development, do not remove it
			},
			Image: fmt.Sprintf("%s:%s", image, config.ImageTag),
			WaitingFor: wait.ForAll(
				wait.ForExec([]string{"vault", "status"}).
					WithPollInterval(time.Millisecond*100).
					WithExitCodeMatcher(func(exitCode int) bool {
						return exitCode == 0
					}),
				wait.ForLog(".*Vault server started.*").AsRegexp(),
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

	fmt.Println("Vault run by host:", config.Host, config.MappedPort)

	return &VaultContainer{
		Container: container,
		Config:    config,
	}, nil
}
