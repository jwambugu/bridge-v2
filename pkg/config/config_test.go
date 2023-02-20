package config

import (
	"os"
	"testing"
)

func TestGet(t *testing.T) {
	t.Parallel()

	if err := os.Setenv("APP_ENV", string(CiCd)); err != nil {
		t.Fatalf("want no error, got %v", err)
	}

	var (
		wantPort = 8000
		wantEnv  = string(CiCd)
	)

	env := Get[string]("APP_ENV", "local")
	if env != wantEnv {
		t.Fatalf("want value %v, got %v", wantEnv, env)
	}

	port := Get[int]("APP_PORT", 3000)
	if port != wantPort {
		t.Fatalf("want value %v, got %v", wantPort, port)
	}
}

func TestGetEnvironment(t *testing.T) {
	t.Parallel()

	if err := os.Setenv("APP_ENV", string(CiCd)); err != nil {
		t.Fatalf("want no error, got %v", err)
	}

	var wantEnv = CiCd

	got := GetEnvironment()
	if got != wantEnv {
		t.Fatalf("want environment %v, got %v", wantEnv, got)
	}

	if short := got.Short(); short != wantEnv.Short() {
		t.Fatalf("want short environment %v, got %v", "t", short)
	}
}
