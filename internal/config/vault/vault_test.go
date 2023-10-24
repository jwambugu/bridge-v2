package vault_test

import (
	"bridge/internal/config/vault"
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestProvider_Get(t *testing.T) {
	t.Skip("Requires dockertest")

	var (
		asserts = assert.New(t)
		ctx     = context.Background()

		addr  = os.Getenv("VAULT_ADDR")
		path  = os.Getenv("VAULT_PATH")
		token = os.Getenv("VAULT_TOKEN")
	)

	provider, err := vault.NewProvider(addr, path, token)
	asserts.NoError(err)

	key, err := provider.Get(ctx, "vault//bridger/database:password")
	asserts.NoError(err)
	asserts.NotNil(key)
}

func TestProvider_Put(t *testing.T) {
	t.Skip("Requires dockertest")

	var (
		asserts = assert.New(t)
		ctx     = context.Background()

		addr  = os.Getenv("VAULT_ADDR")
		path  = os.Getenv("VAULT_PATH")
		token = os.Getenv("VAULT_TOKEN")
	)

	provider, err := vault.NewProvider(addr, path, token)
	asserts.NoError(err)

	err = provider.Put(ctx, "vault//bridger/database:host", "localhost:5432")
	asserts.NoError(err)

	key, err := provider.Get(ctx, "vault//bridger/database:host")
	asserts.NoError(err)
	asserts.NotNil(key)

	key, err = provider.Get(ctx, "vault//bridger/database:host")
	asserts.NoError(err)
	asserts.NotNil(key)
}
