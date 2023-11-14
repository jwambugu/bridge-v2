package testutils

// VaultConfig provides config for a testing vault instance
type VaultConfig struct {
	Address string
	Path    string
	Token   string
}

const (
	vaultAddress = "0.0.0.0:8300"
	vaultPath    = "secret"
	vaultToken   = "dev-only-token"
	vaultTag     = "1.13.3"
)

//func newVaultClient() (*VaultConfig, error) {
//	pool, err := dockertest.NewPool("")
//	if err != nil {
//		return nil, fmt.Errorf("error connecting to docker: %w", err)
//	}
//
//	pool.MaxWait = 5 * time.Second
//
//	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
//		Repository: "vault",
//		Tag:        vaultTag,
//		Env: []string{
//			`VAULT_DEV_ROOT_TOKEN_ID=` + vaultToken,
//			`VAULT_DEV_LISTEN_ADDRESS=` + vaultAddress,
//		},
//		ExposedPorts: []string{"8200/tcp"},
//		PortBindings: map[docker.Port][]docker.PortBinding{},
//		CapAdd:       []string{"IPC_LOCK"},
//	}, func(config *docker.HostConfig) {
//		config.AutoRemove = true
//		config.RestartPolicy = docker.RestartPolicy{
//			Name: "no",
//		}
//	})
//
//	if err != nil {
//		return nil, fmt.Errorf("error starting vault container: %w", err)
//	}
//
//	var (
//		containerPort = resource.GetPort("8300/tcp")
//		address       = `http://0.0.0.0:` + containerPort
//		vaultConfig   = vault.DefaultConfig()
//	)
//
//	vaultConfig.Address = address
//
//	client, err := vault.NewClient(vaultConfig)
//	if err != nil {
//		return nil, fmt.Errorf("error creating vault client: %w", err)
//	}
//
//	client.SetToken(vaultToken)
//
//	var (
//		kv         = client.KVv2(vaultPath)
//		ctx        = context.Background()
//		secretPath = os.Getenv("APP_NAME")
//	)
//
//	if secretPath == "" {
//		secretPath = "bridge"
//	}
//
//	err = pool.Retry(func() error {
//		secrets := map[string]interface{}{
//			"database:user": "root",
//		}
//		return nil
//	})
//
//	return &VaultConfig{
//		Address: address,
//		Path:    vaultPath,
//		Token:   vaultToken,
//	}, nil
//}
