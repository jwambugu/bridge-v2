package docker_test

import (
	"bridge/internal/db"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"github.com/pressly/goose"
	"path/filepath"
	"runtime"
	"time"
)

// PostgresSrv is a testing postgres instance
type PostgresSrv struct {
	DSN string
	DB  *sqlx.DB
}

func NewPostgresSrv() (*PostgresSrv, func() error, error) {
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
		dsn         = fmt.Sprintf("postgres://postgres:secret@%s/bridge?sslmode=disable", hostAndPort)
		pgSrv       = &PostgresSrv{DSN: dsn}
	)

	if err = resource.Expire(60); err != nil {
		return nil, nil, fmt.Errorf("error setting timer to remove container: %w", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 20 * time.Second

	err = pool.Retry(func() error {
		conn, err := db.NewConnection(pgSrv.DSN)
		if err != nil {
			return err
		}

		if err := conn.Ping(); err != nil {
			return err
		}

		pgSrv.DB = conn

		_, pwd, _, _ := runtime.Caller(0)
		migrationsDir := filepath.Join(pwd, "../../..")
		migrationsDir = migrationsDir + "/db/migrations"

		return goose.Run("up", conn.DB, migrationsDir)
	})

	if err != nil {
		return nil, nil, fmt.Errorf("error connecting to docker: %w", err)
	}

	return pgSrv, func() error {
		if err := pool.Purge(resource); err != nil {
			return fmt.Errorf("error purging container: %w", err)
		}
		return nil
	}, nil
}
