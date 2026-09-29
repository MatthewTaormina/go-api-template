// Package config loads application settings from environment variables.
// All values have sane defaults so the server runs out of the box.
// No third-party config library is used.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds every tunable server setting.
// Add new fields here rather than scattering os.Getenv calls across packages.
type Config struct {
	// Host is the network interface to bind. Empty string means all interfaces.
	Host string

	// Port is the TCP port the HTTP server listens on.
	Port string

	// ReadTimeout is the max duration for reading the entire request, including body.
	ReadTimeout time.Duration

	// WriteTimeout is the max duration before timing out writes of the response.
	WriteTimeout time.Duration

	// IdleTimeout is the max amount of time to wait for the next request on a keep-alive connection.
	IdleTimeout time.Duration

	// ShutdownTimeout is how long graceful shutdown waits for in-flight requests to finish.
	ShutdownTimeout time.Duration

	// TrustedProxy controls whether the server reads X-Forwarded-For / X-Real-IP headers.
	// Set to true only when the server sits behind a known reverse proxy (nginx, Caddy, etc.).
	// Leaving it false prevents IP spoofing from untrusted clients.
	TrustedProxy bool
}

// Load reads configuration from environment variables and returns a Config.
// Every variable has a default so no env setup is required for local development.
//
// Environment variables:
//
//	HOST              Network interface to bind (default: "" = all interfaces)
//	PORT              TCP port (default: "8080")
//	READ_TIMEOUT      Max request read duration, e.g. "5s" (default: 5s)
//	WRITE_TIMEOUT     Max response write duration, e.g. "10s" (default: 10s)
//	IDLE_TIMEOUT      Keep-alive idle duration, e.g. "120s" (default: 120s)
//	SHUTDOWN_TIMEOUT  Graceful shutdown wait, e.g. "15s" (default: 15s)
//	TRUSTED_PROXY     Trust X-Forwarded-For header, "true"/"false" (default: false)
func Load() Config {
	return Config{
		Host:            getEnv("HOST", ""),
		Port:            getEnv("PORT", "8080"),
		ReadTimeout:     getDuration("READ_TIMEOUT", 5*time.Second),
		WriteTimeout:    getDuration("WRITE_TIMEOUT", 10*time.Second),
		IdleTimeout:     getDuration("IDLE_TIMEOUT", 120*time.Second),
		ShutdownTimeout: getDuration("SHUTDOWN_TIMEOUT", 15*time.Second),
		TrustedProxy:    getBool("TRUSTED_PROXY", false),
	}
}

// Addr returns the combined "host:port" string for net.Listen / http.Server.Addr.
func (c Config) Addr() string {
	return c.Host + ":" + c.Port
}

// getEnv returns the environment variable named key, or fallback if not set or empty.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getDuration parses a time.Duration string from the environment (e.g. "30s", "1m").
// Returns fallback if the variable is absent or unparseable.
func getDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

// getBool parses a boolean string ("true"/"1"/"false"/"0") from the environment.
// Returns fallback if the variable is absent or unparseable.
func getBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
