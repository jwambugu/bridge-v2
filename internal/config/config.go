package config

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/joho/godotenv"
)

// Environment is the current running environment.
type Environment string

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
func Get[V EnvKey](key Key, fallback V) V {
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
	return value.(V)
}

// GetEnvironment returns the current running environment the application is running on.
func GetEnvironment() Environment {
	if env := Get[string]("APP_ENV", ""); env != "" {
		return Environment(env)
	}
	return Test
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
