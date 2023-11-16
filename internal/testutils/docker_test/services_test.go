package docker_test_test

import (
	"bridge/internal/testutils/docker_test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVaultClient(t *testing.T) {
	t.Parallel()

	asserts := assert.New(t)

	cl, cleanup, err := docker_test.NewVaultClient()
	defer func() {
		asserts.NoError(cleanup())
	}()

	asserts.NoError(err)
	asserts.NotNil(cl)
}

func TestPostgresSrv(t *testing.T) {
	t.Parallel()

	asserts := assert.New(t)

	pgSrv, cleanup, err := docker_test.NewPostgresSrv()
	defer func() {
		asserts.NoError(cleanup())
	}()

	asserts.NoError(err)
	asserts.NotEmpty(pgSrv.DSN)
}
