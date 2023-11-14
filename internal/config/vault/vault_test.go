package vault_test

import (
	"bridge/internal/config/vault"
	"bridge/internal/testutils"
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProvider_Get_Put(t *testing.T) {
	t.Parallel()

	var (
		asserts = assert.New(t)
		ctx     = context.Background()
	)

	vaultClient := testutils.VaultClient(t)

	provider, err := vault.NewProvider(vaultClient.Address, vaultClient.Path, vaultClient.Token)
	asserts.NoError(err)

	var (
		wantKey   = "vault//bridger/test:key"
		wantValue = "test:key"
	)

	err = provider.Put(ctx, wantKey, wantValue)
	asserts.NoError(err)

	gotValue, err := provider.Get(ctx, wantKey)
	asserts.NoError(err)
	asserts.NotNil(gotValue)
	asserts.Equal(wantValue, gotValue)

	gotValue, err = provider.Get(ctx, wantKey)
	asserts.NoError(err)
	asserts.NotNil(gotValue)
	asserts.Equal(wantValue, gotValue)
}
