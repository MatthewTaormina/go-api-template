package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/user/go-api-template/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	// Unset all env vars to ensure defaults apply.
	vars := []string{"HOST", "PORT", "READ_TIMEOUT", "WRITE_TIMEOUT",
		"IDLE_TIMEOUT", "SHUTDOWN_TIMEOUT", "TRUSTED_PROXY"}
	for _, v := range vars {
		t.Setenv(v, "")
	}

	cfg := config.Load()

	if cfg.Port != "8080" {
		t.Errorf("Port: got %q, want %q", cfg.Port, "8080")
	}
	if cfg.ReadTimeout != 5*time.Second {
		t.Errorf("ReadTimeout: got %v, want %v", cfg.ReadTimeout, 5*time.Second)
	}
	if cfg.TrustedProxy {
		t.Error("TrustedProxy: got true, want false")
	}
	if cfg.Addr() != ":8080" {
		t.Errorf("Addr: got %q, want %q", cfg.Addr(), ":8080")
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("HOST", "127.0.0.1")
	t.Setenv("PORT", "9090")
	t.Setenv("READ_TIMEOUT", "30s")
	t.Setenv("TRUSTED_PROXY", "true")

	cfg := config.Load()

	if cfg.Host != "127.0.0.1" {
		t.Errorf("Host: got %q, want %q", cfg.Host, "127.0.0.1")
	}
	if cfg.Port != "9090" {
		t.Errorf("Port: got %q, want %q", cfg.Port, "9090")
	}
	if cfg.ReadTimeout != 30*time.Second {
		t.Errorf("ReadTimeout: got %v, want %v", cfg.ReadTimeout, 30*time.Second)
	}
	if !cfg.TrustedProxy {
		t.Error("TrustedProxy: got false, want true")
	}
	if cfg.Addr() != "127.0.0.1:9090" {
		t.Errorf("Addr: got %q, want %q", cfg.Addr(), "127.0.0.1:9090")
	}
}

func TestLoad_InvalidDurationFallsBack(t *testing.T) {
	os.Setenv("READ_TIMEOUT", "not-a-duration")
	defer os.Unsetenv("READ_TIMEOUT")

	cfg := config.Load()
	if cfg.ReadTimeout != 5*time.Second {
		t.Errorf("ReadTimeout: got %v, want default 5s", cfg.ReadTimeout)
	}
}

func TestLoad_InvalidBoolFallsBack(t *testing.T) {
	os.Setenv("TRUSTED_PROXY", "maybe")
	defer os.Unsetenv("TRUSTED_PROXY")

	cfg := config.Load()
	if cfg.TrustedProxy {
		t.Error("TrustedProxy: expected false fallback on invalid value")
	}
}
