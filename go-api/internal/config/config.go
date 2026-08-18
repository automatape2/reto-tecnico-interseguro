// Package config loads and validates this service's configuration from
// environment variables. Loading fails fast at startup (returned as an
// error from Load, surfaced via log.Fatal in main) rather than at
// first-request time, so a misconfigured deployment never serves traffic.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds every environment-derived setting this service needs.
type Config struct {
	Port               string
	JWTSecret          string
	JWTExpiry          time.Duration
	GoAPIPresharedKey  string
	NodeAPIBaseURL     string
	NodeAPITimeout     time.Duration
	CORSAllowedOrigins string
}

// Load reads configuration from the environment, applying defaults for
// optional settings and returning an error if any required setting is
// missing.
func Load() (*Config, error) {
	cfg := &Config{
		Port:               getEnv("PORT", "8080"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		GoAPIPresharedKey:  os.Getenv("GO_API_PRESHARED_KEY"),
		NodeAPIBaseURL:     getEnv("NODE_API_BASE_URL", "http://localhost:4000"),
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "*"),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.GoAPIPresharedKey == "" {
		return nil, fmt.Errorf("GO_API_PRESHARED_KEY is required")
	}

	expiryMinutes, err := getEnvInt("JWT_EXPIRY_MINUTES", 60)
	if err != nil {
		return nil, err
	}
	cfg.JWTExpiry = time.Duration(expiryMinutes) * time.Minute

	timeoutMs, err := getEnvInt("NODE_API_TIMEOUT_MS", 5000)
	if err != nil {
		return nil, err
	}
	cfg.NodeAPITimeout = time.Duration(timeoutMs) * time.Millisecond

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer, got %q: %w", key, v, err)
	}
	return n, nil
}
