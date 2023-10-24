package vault

import (
	"context"
	"errors"
	"fmt"
	vault "github.com/hashicorp/vault/api"
	"strings"
)

const providerName = `vault`

var ErrInvalidProvider = errors.New("invalid provider given")
var ErrInvalidKey = errors.New("invalid key provided")

type Provider struct {
	kv      *vault.KVv2
	secrets map[string]map[string]string
}

// Get retrieves a value from vault using the KV engine. The actual key selected is determined by the value
// separated by the forward slash. For example "vault//secret-path/database:user" will retrieve the key "database:user"
// from the path "secret-path" using the engine "vault".
func (p *Provider) Get(ctx context.Context, key string) (string, error) {
	providerWithKeys := strings.Split(key, "//")
	if len(providerWithKeys) != 2 {
		return "", ErrInvalidKey
	}

	if providerWithKeys[0] != providerName {
		return "", ErrInvalidProvider
	}

	var (
		keyWithPath = strings.Split(providerWithKeys[1], "/")
		secretPath  = keyWithPath[0]
		keyName     = keyWithPath[1]
	)

	if v, ok := p.secrets[secretPath]; ok {
		secretValue, ok := v[keyName]
		if !ok {
			return "", fmt.Errorf("key not found on cached data")
		}

		return secretValue, nil
	}

	kvSecret, err := p.kv.Get(ctx, secretPath)
	if err != nil {
		return "", fmt.Errorf("error reading secrets from vault - %w", err)
	}

	secrets := make(map[string]string, len(kvSecret.Data))
	for k, v := range kvSecret.Data {
		val, ok := v.(string)
		if !ok {
			return "", fmt.Errorf("secret value in vault is not a string")
		}

		secrets[k] = val
	}

	p.secrets[secretPath] = secrets

	val, ok := secrets[keyName]
	if !ok {
		return "", fmt.Errorf("key not found on secrets data")
	}

	return val, nil
}

func (p *Provider) Put(ctx context.Context, key string, value string) error {
	providerWithKeys := strings.Split(key, "//")
	if len(providerWithKeys) != 2 {
		return ErrInvalidKey
	}

	if providerWithKeys[0] != providerName {
		return ErrInvalidProvider
	}

	var (
		keyWithPath = strings.Split(providerWithKeys[1], "/")
		secretPath  = keyWithPath[0]
		keyName     = keyWithPath[1]
	)

	kvSecret, err := p.kv.Get(ctx, secretPath)
	if err != nil {
		return fmt.Errorf("error reading secrets from vault - %w", err)
	}

	kvSecret.Data[keyName] = value

	if _, err := p.kv.Put(ctx, secretPath, kvSecret.Data); err != nil {
		return fmt.Errorf("error inserting key - %w", err)
	}

	// Refresh the cache :)
	if _, err = p.Get(ctx, key); err != nil {
		return err
	}

	return nil
}

// NewProvider creates a new vault provider
func NewProvider(addr, mountPath, token string) (*Provider, error) {
	config := vault.DefaultConfig()
	config.Address = addr

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("error creating vault client - %w", err)
	}

	client.SetToken(token)

	return &Provider{
		kv:      client.KVv2(mountPath),
		secrets: make(map[string]map[string]string),
	}, nil
}
