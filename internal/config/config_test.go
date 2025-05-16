package config

import (
	"os"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	// Clean environment before testing
	if err := os.Unsetenv("ICEPANEL_API_URL"); err != nil {
		t.Fatalf("Failed to unset env var: %v", err)
	}
	if err := os.Unsetenv("ICEPANEL_TIMEOUT_SECONDS"); err != nil {
		t.Fatalf("Failed to unset env var: %v", err)
	}
	if err := os.Unsetenv("ICEPANEL_TOKEN"); err != nil {
		t.Fatalf("Failed to unset env var: %v", err)
	}

	// Test with default values
	cfg := NewConfig()
	if cfg.APIBaseURL != "https://api.icepanel.io/v1" {
		t.Errorf("Expected default API URL, got %s", cfg.APIBaseURL)
	}
	if cfg.RequestTimeout != 30*time.Second {
		t.Errorf("Expected default timeout of 30s, got %v", cfg.RequestTimeout)
	}
	if cfg.DefaultToken != "" {
		t.Errorf("Expected empty default token, got %s", cfg.DefaultToken)
	}
}

func TestConfigWithEnvironment(t *testing.T) {
	// Set environment variables
	if err := os.Setenv("ICEPANEL_API_URL", "https://custom.api.com"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	if err := os.Setenv("ICEPANEL_TOKEN", "env-token"); err != nil {
		t.Fatalf("Failed to set env var: %v", err)
	}
	// Cleanup after test
	defer func() {
		if err := os.Unsetenv("ICEPANEL_API_URL"); err != nil {
			t.Fatalf("Failed to unset env var: %v", err)
		}
		if err := os.Unsetenv("ICEPANEL_TOKEN"); err != nil {
			t.Fatalf("Failed to unset env var: %v", err)
		}
	}()

	// Test with environment values
	cfg := NewConfig()
	if cfg.APIBaseURL != "https://custom.api.com" {
		t.Errorf("Expected custom API URL, got %s", cfg.APIBaseURL)
	}
	if cfg.DefaultToken != "env-token" {
		t.Errorf("Expected token from env, got %s", cfg.DefaultToken)
	}
}

func TestGetEnvOr(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue string
		want         string
	}{
		{
			name:         "existing environment variable",
			envKey:       "TEST_KEY1",
			envValue:     "test-value",
			defaultValue: "default-value",
			want:         "test-value",
		},
		{
			name:         "non-existing environment variable",
			envKey:       "TEST_KEY2",
			envValue:     "",
			defaultValue: "default-value",
			want:         "default-value",
		},
		{
			name:         "empty environment variable",
			envKey:       "TEST_KEY3",
			envValue:     "",
			defaultValue: "default-value",
			want:         "default-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				if err := os.Setenv(tt.envKey, tt.envValue); err != nil {
					t.Fatalf("Failed to set env var: %v", err)
				}
				defer func() {
					if err := os.Unsetenv(tt.envKey); err != nil {
						t.Fatalf("Failed to unset env var: %v", err)
					}
				}()
			} else {
				if err := os.Unsetenv(tt.envKey); err != nil {
					t.Fatalf("Failed to unset env var: %v", err)
				}
			}

			if got := getEnvOr(tt.envKey, tt.defaultValue); got != tt.want {
				t.Errorf("getEnvOr() = %v, want %v", got, tt.want)
			}
		})
	}
}
