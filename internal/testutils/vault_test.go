package testutils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVaultClient(t *testing.T) {
	asserts := assert.New(t)

	cl, cleanup, err := newVaultClient()
	defer func() {
		asserts.NoError(cleanup())
	}()

	asserts.NoError(err)
	asserts.NotNil(cl)
}
