package config

import (
	"bridge/internal/config/vault"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// defaultEnvFile is the default file name used for storing environment variables if none is provided
const defaultEnvFile = ".env"

// ProductionEnvironment is the release environment.
const ProductionEnvironment = "production"

// Provider provides methods for interacting with the configuration provider
type Provider interface {
	Get(ctx context.Context, key string) (string, error)
	Put(ctx context.Context, key string, value string) error
}

// envKey stores the environment variables keys
type envKey struct {
	Debug           bool   `env:"DEBUG"`
	Env             string `env:"ENV"`
	GrpcGatewayPort string `env:"GW_PORT"`
	Name            string `env:"NAME"`
	Port            uint16 `env:"PORT"`
	URL             string `env:"URL"`

	DbDsn  string `env:"DB_DSN" secured:"true"`
	JwtKey string `env:"JWT_KEY" secured:"true"`
}

// EnvKey stores parsed env keys values
var EnvKey envKey

// Config represents a configuration container that utilizes a Provider for retrieving configuration values.
// It encapsulates the mechanism for accessing configuration information and serves as a central entry point
// for managing configuration settings within an application.
type Config struct {
	provider Provider
}

// AbsPath returns the project absolute path from the root project dir
func AbsPath(filename string) string {
	var (
		_, b, _, _ = runtime.Caller(0)
		path       = filepath.Join(filepath.Dir(b), "../..")
	)

	if filename == "" {
		return path
	}

	return filepath.Join(path, filename)
}

// Load reads the provided env file and maps the env key value to Key which is exported
// and can be used across the app.
func (c *Config) Load(ctx context.Context, filename string) (err error) {
	if filename == "" {
		filename = defaultEnvFile
	}

	if err := godotenv.Load(AbsPath(filename)); err != nil {
		return fmt.Errorf("load env: %v", err)
	}

	var key envKey

	keyType := reflect.TypeOf(key)

	for i := 0; i < keyType.NumField(); i++ {
		var (
			field      = keyType.Field(i)
			envTag     = field.Tag.Get("env")
			securedTag = field.Tag.Get("secured")
		)

		envValue, ok := os.LookupEnv(envTag)
		if !ok {
			return fmt.Errorf("env variable %q not found", envTag)
		}

		if securedTag != "" {
			envValue, err = c.provider.Get(ctx, envValue)
			if err != nil {
				return err
			}
		}

		var (
			fieldValue = reflect.ValueOf(&key).Elem().FieldByName(field.Name)
			kind       = field.Type.Kind()
		)

		switch kind {
		case reflect.Bool:
			isTrue := strings.ToLower(envValue) == "true"
			fieldValue.SetBool(isTrue)
		case reflect.Uint16:
			val, err := strconv.Atoi(envValue)
			if err != nil {
				return fmt.Errorf("atoi: %v", err)
			}
			fieldValue.SetUint(uint64(val))
		case reflect.String:
			fieldValue.SetString(envValue)
		default:
			return fmt.Errorf("unsupported type %q for %q", kind, envTag)
		}
	}

	EnvKey = key
	return nil
}

// NewConfig initializes new Config
func NewConfig(p Provider) *Config {
	return &Config{
		provider: p,
	}
}

// NewDefaultConfig creates a new Config instance with using vault as the default provider
func NewDefaultConfig(ctx context.Context) (*Config, error) {
	var (
		vaultAddr  = os.Getenv("VAULT_ADDR")
		vaultPath  = os.Getenv("VAULT_PATH")
		vaultToken = os.Getenv("VAULT_TOKEN")
	)

	vaultProvider, err := vault.NewProvider(vaultAddr, vaultPath, vaultToken)
	if err != nil {
		return nil, err
	}

	config := NewConfig(vaultProvider)

	if err = config.Load(ctx, ""); err != nil {
		return nil, err
	}

	return config, nil
}
