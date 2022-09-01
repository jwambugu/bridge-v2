package config

import (
	"testing"
)

func TestGet(t *testing.T) {
	var (
		wantPort = 8000
		wantEnv  = "test"
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
	var wantEnv = Test

	got := GetEnvironment()
	if got != wantEnv {
		t.Fatalf("want environment %v, got %v", wantEnv, got)
	}

	if short := got.Short(); short != "t" {
		t.Fatalf("want short environment %v, got %v", "t", short)
	}
}
