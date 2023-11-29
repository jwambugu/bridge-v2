package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/joho/godotenv"
)

// Environment is the current running environment.
type Environment string

const ProviderKeySeparator = "secret://"

const (
	// Local is the environment used to run the application on the local machine.
	Local Environment = "local"
	// Test is the environment used during testing.
	Test Environment = "test"
	// Staging is the testing environment before releasing to Production.
	Staging Environment = "staging"
	// Production is the release environment.
	Production Environment = "production"
	// CiCd is the deployment and testing environment.
	CiCd Environment = "cicd"
)

// Provider provides methods for interacting with the configuration provider
type Provider interface {
	Get(ctx context.Context, key string) (string, error)
	Put(ctx context.Context, key string, value string) error
}

type Configuration struct {
	provider Provider
}

func (c *Configuration) Get(ctx context.Context, key string) (string, error) {
	envKey := os.Getenv(key)
	securedEnvKey := os.Getenv(key + "_SECURE")

	if securedEnvKey != "" {
		securedKey, err := c.provider.Get(ctx, securedEnvKey)
		if err != nil {
			return "", fmt.Errorf("provider get: %w", err)
		}

		envKey = securedKey
	}

	return envKey, nil
}

// NewConfig initializes new Configuration
func NewConfig(p Provider) *Configuration {
	return &Configuration{
		provider: p,
	}
}

var _ = loadConfig()

// Short retrieves the shortname for the Environment
func (e Environment) Short() string {
	switch e {
	case Local:
		return "l"
	case Test:
		return "t"
	case Production:
		return "p"
	case Staging:
		return "s"
	case CiCd:
		return "c"
	default:
		return "x"
	}
}

type EnvKey interface {
	string | int
}

// Get returns the value of the environment variable named by the key, If the Key is not set, it returns the fallback.
func Get[T EnvKey](key Key, fallback T) T {
	val := os.Getenv(string(key))
	if val == "" {
		return fallback
	}

	var value any
	switch any(fallback).(type) {
	case string:
		value = val
	case int:
		i, err := strconv.Atoi(val)
		if err != nil {
			log.Printf("config get %s: %v", key, err)
			return fallback
		}

		value = i
	default:
		return fallback
	}
	return value.(T)
}

// GetEnvironment returns the current running environment the application is running on.
func GetEnvironment() Environment {
	if env := Get[string](AppEnv, ""); env != "" {
		return Environment(env)
	}
	return Test
}

func GetDBDsn() string {
	var (
		user     = Get[string](DbUser, "")
		password = Get[string](DbPassword, "")
		host     = Get[string](DbHost, "")
		name     = Get[string](DbName, "")
		url      = Key(fmt.Sprintf(`postgres://%v:%v@%v/%v?sslmode=disable`, user, password, host, name))
	)

	return string(url)
}

func loadConfig() error {
	_, b, _, _ := runtime.Caller(0)

	var (
		pwd     = filepath.Dir(b)
		env     = GetEnvironment()
		envFile = pwd + `/.` + string(env) + `.env`
	)

	if err := godotenv.Load(envFile); err != nil {
		log.Fatalf("load config %s: %s", envFile, err.Error())
	}

	return nil
}
