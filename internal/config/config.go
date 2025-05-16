// Package config handles application configuration with environment variable support.
package config

import (
	"os"
	"time"
)

// Config holds application configuration with defaults from environment variables.
type Config struct {
	APIBaseURL     string
	RequestTimeout time.Duration
	DefaultToken   string
}

// NewConfig creates a Config with values from environment or defaults.
func NewConfig() *Config {
	return &Config{
		APIBaseURL:     getEnvOr("ICEPANEL_API_URL", "https://api.icepanel.io/v1"),
		RequestTimeout: time.Duration(getEnvIntOr("ICEPANEL_TIMEOUT_SECONDS", 30)) * time.Second,
		DefaultToken:   os.Getenv("ICEPANEL_TOKEN"),
	}
}

// Helper function to get environment variable with default.
func getEnvOr(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}

// Helper function to get environment variable as int with default.
func getEnvIntOr(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		if intVal, err := time.ParseDuration(value); err == nil {
			return int(intVal)
		}
	}
	return defaultValue
}
