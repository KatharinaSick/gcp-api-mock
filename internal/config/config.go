// Package config provides configuration loading and management for the GCP API Mock.
package config

import (
	"fmt"
	"os"
)

// Config holds the application configuration.
type Config struct {
	// Host is the server host address.
	Host string

	// Port is the server port.
	Port string

	// Environment is the runtime environment (development, production).
	Environment string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		Host:        getEnv("GCP_MOCK_HOST", "0.0.0.0"),
		Port:        getEnv("GCP_MOCK_PORT", "8080"),
		Environment: getEnv("GCP_MOCK_ENV", "development"),
	}
}

// Address returns the full server address (host:port).
func (c *Config) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
