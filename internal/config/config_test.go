package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Save original env vars and restore after test
	originalHost := os.Getenv("GCP_MOCK_HOST")
	originalPort := os.Getenv("GCP_MOCK_PORT")
	originalEnv := os.Getenv("GCP_MOCK_ENV")
	defer func() {
		os.Setenv("GCP_MOCK_HOST", originalHost)
		os.Setenv("GCP_MOCK_PORT", originalPort)
		os.Setenv("GCP_MOCK_ENV", originalEnv)
	}()

	t.Run("default values", func(t *testing.T) {
		os.Unsetenv("GCP_MOCK_HOST")
		os.Unsetenv("GCP_MOCK_PORT")
		os.Unsetenv("GCP_MOCK_ENV")

		cfg := Load()

		if cfg.Host != "0.0.0.0" {
			t.Errorf("expected Host '0.0.0.0', got '%s'", cfg.Host)
		}
		if cfg.Port != "8080" {
			t.Errorf("expected Port '8080', got '%s'", cfg.Port)
		}
		if cfg.Environment != "development" {
			t.Errorf("expected Environment 'development', got '%s'", cfg.Environment)
		}
	})

	t.Run("custom values from environment", func(t *testing.T) {
		os.Setenv("GCP_MOCK_HOST", "127.0.0.1")
		os.Setenv("GCP_MOCK_PORT", "9090")
		os.Setenv("GCP_MOCK_ENV", "production")

		cfg := Load()

		if cfg.Host != "127.0.0.1" {
			t.Errorf("expected Host '127.0.0.1', got '%s'", cfg.Host)
		}
		if cfg.Port != "9090" {
			t.Errorf("expected Port '9090', got '%s'", cfg.Port)
		}
		if cfg.Environment != "production" {
			t.Errorf("expected Environment 'production', got '%s'", cfg.Environment)
		}
	})
}

func TestConfig_Address(t *testing.T) {
	cfg := &Config{Host: "localhost", Port: "3000"}
	expected := "localhost:3000"

	if got := cfg.Address(); got != expected {
		t.Errorf("Address() = %s, want %s", got, expected)
	}
}

func TestConfig_IsDevelopment(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		want        bool
	}{
		{"development environment", "development", true},
		{"production environment", "production", false},
		{"staging environment", "staging", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Environment: tt.environment}
			if got := cfg.IsDevelopment(); got != tt.want {
				t.Errorf("IsDevelopment() = %v, want %v", got, tt.want)
			}
		})
	}
}
