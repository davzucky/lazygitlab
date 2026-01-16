package config

import (
	"os"
	"testing"
)

func TestLoadFromEnv(t *testing.T) {
	os.Setenv("GITLAB_TOKEN", "test-token")
	os.Setenv("GITLAB_HOST", "https://gitlab.example.com")
	defer func() {
		os.Unsetenv("GITLAB_TOKEN")
		os.Unsetenv("GITLAB_HOST")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Token != "test-token" {
		t.Errorf("Expected token 'test-token', got '%s'", cfg.Token)
	}

	if cfg.Host != "https://gitlab.example.com" {
		t.Errorf("Expected host 'https://gitlab.example.com', got '%s'", cfg.Host)
	}
}

func TestLoadNoConfig(t *testing.T) {
	os.Unsetenv("GITLAB_TOKEN")
	os.Unsetenv("GITLAB_HOST")

	cfg, err := Load()

	if err == nil {
		t.Log("Note: Configuration loaded from glab config file (this is expected if glab is configured)")
		return
	}

	if cfg != nil {
		t.Error("Expected nil config when no configuration is present")
	}
}

func TestConfigIsValid(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected bool
	}{
		{
			name:     "valid config",
			config:   Config{Token: "token", Host: "https://gitlab.com"},
			expected: true,
		},
		{
			name:     "missing token",
			config:   Config{Token: "", Host: "https://gitlab.com"},
			expected: false,
		},
		{
			name:     "missing host",
			config:   Config{Token: "token", Host: ""},
			expected: false,
		},
		{
			name:     "missing both",
			config:   Config{Token: "", Host: ""},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.isValid()
			if result != tt.expected {
				t.Errorf("isValid() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
