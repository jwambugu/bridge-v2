package docker_test

import (
	"context"
	"fmt"
	vault "github.com/hashicorp/vault/api"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"time"
)

// VaultClient is a testing vault instance
type VaultClient struct {
	Client  *vault.Client
	Address string
	Path    string
	Token   string
}

const (
	vaultAddress = "0.0.0.0:8300"
	vaultPath    = "secret"
	vaultToken   = "dev-only-token"
)

func NewVaultClient() (*VaultClient, func() error, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, fmt.Errorf("error constructing pool: %w", err)
	}

	if err := pool.Client.Ping(); err != nil {
		return nil, nil, fmt.Errorf("error connecting to docker: %w", err)
	}

	pool.MaxWait = 5 * time.Second

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "vault",
		Tag:        "1.13.3",
		Env: []string{
			`VAULT_DEV_ROOT_TOKEN_ID=` + vaultToken,
			`VAULT_DEV_LISTEN_ADDRESS=` + vaultAddress,
		},
		ExposedPorts: []string{"8300/tcp"},
		PortBindings: map[docker.Port][]docker.PortBinding{},
		CapAdd:       []string{"IPC_LOCK"},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})

	if err != nil {
		return nil, nil, fmt.Errorf("error starting resource: %w", err)
	}

	_ = resource.Expire(60)

	//goland:noinspection HttpUrlsUsage
	var (
		address     = `http://0.0.0.0:` + resource.GetPort("8300/tcp")
		vaultConfig = vault.DefaultConfig()
	)

	vaultConfig.Address = address

	client, err := vault.NewClient(vaultConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating vault client: %w", err)
	}

	client.SetToken(vaultToken)

	var (
		kv         = client.KVv2(vaultPath)
		ctx        = context.Background()
		secretPath = "bridge"
	)

	if err = resource.Expire(60); err != nil {
		return nil, nil, fmt.Errorf("error setting timer to remove container: %w", err)
	}

	err = pool.Retry(func() error {
		_, err := kv.Put(ctx, secretPath, map[string]interface{}{
			"database:dsn":  "postgres://postgres:secret@localhost:5432/bridge?sslmode=disable",
			"database:user": "root",
			"jwt_key":       "86f5778df1b11e35caf8bc793391bfd1",
		})

		if err != nil {
			return fmt.Errorf("failed to set secrets: %w", err)
		}

		secret, err := kv.Get(ctx, secretPath)
		if err != nil {
			return fmt.Errorf("failed to get secrets: %w", err)
		}

		if secret.Data == nil {
			return fmt.Errorf("no secrets found")
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	v := &VaultClient{
		Address: address,
		Client:  client,
		Path:    vaultPath,
		Token:   vaultToken,
	}

	return v, func() error {
		if err := pool.Purge(resource); err != nil {
			return fmt.Errorf("error purging container: %w", err)
		}
		return nil
	}, nil
}
