package testutils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVaultClient(t *testing.T) {
	t.Parallel()

	asserts := assert.New(t)

	cl, cleanup, err := newVaultClient()
	defer func() {
		asserts.NoError(cleanup())
	}()

	asserts.NoError(err)
	asserts.NotNil(cl)
}

func TestPostgresSrv(t *testing.T) {
	t.Parallel()

	asserts := assert.New(t)

	_, cleanup, err := postgresSrv()
	defer func() {
		asserts.NoError(cleanup())
	}()

	asserts.NoError(err)
}
