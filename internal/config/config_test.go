package config_test

import (
	"bridge/internal/config"
	"bridge/internal/config/vault"
	"bridge/internal/testutils/docker_test"
	"context"
	"github.com/stretchr/testify/require"
	"log"
	"os"
	"testing"
)

var testVaultProvider *vault.Provider

func testMain(m *testing.M) int {
	vaultClient, vaultCleanup, err := docker_test.NewVaultClient()
	if err != nil {
		log.Fatalln(err)
	}

	defer func() {
		if err = vaultCleanup(); err != nil {
			log.Fatalln(err)
		}
	}()

	vaultProvider, err := vault.NewProvider(vaultClient.Address, vaultClient.Path, vaultClient.Token)
	if err != nil {
		log.Fatalln(err)
	}

	_ = os.Setenv("VAULT_ADDR", vaultClient.Address)
	_ = os.Setenv("VAULT_PATH", vaultClient.Path)
	_ = os.Setenv("VAULT_TOKEN", vaultClient.Token)

	testVaultProvider = vaultProvider

	appConfig := config.NewConfig(vaultProvider)
	if err = appConfig.Load(context.Background(), ""); err != nil {
		log.Fatalln(err)
	}

	return m.Run()
}

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func TestConfig(t *testing.T) {
	appConfig := config.NewConfig(testVaultProvider)
	ctx := context.Background()

	err := appConfig.Load(ctx, "")
	require.NoError(t, err)

	appConfig, err = config.NewDefaultConfig(ctx)
	require.NoError(t, err)
	require.NotNil(t, appConfig)
}
