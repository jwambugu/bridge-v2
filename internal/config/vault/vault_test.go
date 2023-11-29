package vault_test

import (
	"bridge/internal/config"
	"bridge/internal/config/vault"
	"bridge/internal/testutils/docker_test"
	"context"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

var testVaultClient *docker_test.VaultClient

func testMain(m *testing.M) (code int, err error) {
	client, cleanup, err := docker_test.NewVaultClient()
	if err != nil {
		return 0, err
	}

	defer func() {
		err = cleanup()
	}()

	testVaultClient = client

	return m.Run(), err
}

func TestMain(m *testing.M) {
	code, err := testMain(m)
	if err != nil {
		log.Fatalln(err)
	}

	os.Exit(code)
}

func TestProvider_Get_Put(t *testing.T) {
	t.Parallel()

	var (
		asserts = assert.New(t)
		ctx     = context.Background()
	)

	provider, err := vault.NewProvider(testVaultClient.Address, testVaultClient.Path, testVaultClient.Token)
	asserts.NoError(err)

	var (
		keyPrefix = config.ProviderKeySeparator
		wantKey   = keyPrefix + "bridger/test:key"
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
