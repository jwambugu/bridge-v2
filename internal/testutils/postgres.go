package testutils

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"testing"
	"time"
)

type PostgresConfig struct {
	DSN string
}

func postgresSrv() (*PostgresConfig, func() error, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, fmt.Errorf("error constructing pool: %w", err)
	}

	if err := pool.Client.Ping(); err != nil {
		return nil, nil, fmt.Errorf("error connecting to docker: %w", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15.2-alpine3.17",
		Env: []string{
			"POSTGRES_USER=postgres",
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_DB=bridge",
			"listen_addresses='*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})

	if err != nil {
		return nil, nil, fmt.Errorf("error starting resource: %w", err)
	}

	var (
		hostAndPort = resource.GetHostPort("5432/tcp")
		databaseUrl = fmt.Sprintf("postgres://postgres:secret@%s/bridge?sslmode=disable", hostAndPort)
	)

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 120 * time.Second
	err = pool.Retry(func() error {
		// TODO: attempt to connect to the DB
		db, err := sqlx.Connect("postgres", databaseUrl)
		if err != nil {
			return err
		}

		return db.Ping()
	})

	if err != nil {
		return nil, nil, fmt.Errorf("error connecting to docker: %w", err)
	}

	p := &PostgresConfig{
		DSN: databaseUrl,
	}

	return p, func() error {
		if err := pool.Purge(resource); err != nil {
			return fmt.Errorf("error purging container: %w", err)
		}
		return nil
	}, nil
}

func PostgresSrv(t testing.TB) *PostgresConfig {
	postgresConfig, cleanup, err := postgresSrv()
	if err != nil {
		t.Fatalf("%v\n", err)
	}

	t.Cleanup(func() {
		if err = cleanup(); err != nil {
			t.Fatalf("%v\n", err)
		}
	})

	return postgresConfig
}
