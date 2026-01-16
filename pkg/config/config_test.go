package config

import (
	"os"
	"path/filepath"
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

func TestLoadFromGlabConfig(t *testing.T) {
	homeDir := t.TempDir()
	if err := os.Setenv("HOME", homeDir); err != nil {
		t.Fatalf("failed to set HOME: %v", err)
	}
	t.Setenv("HOME", homeDir)

	configPath := filepath.Join(homeDir, ".config", "glab-cli")
	if err := os.MkdirAll(configPath, 0o755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	config := `host: gitlab.example.io
hosts:
  gitlab.example.io:
    token: glpat-test
    api_host: gitlab.example.io
    api_protocol: https
`

	if err := os.WriteFile(filepath.Join(configPath, "config.yml"), []byte(config), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	os.Unsetenv("GITLAB_TOKEN")
	os.Unsetenv("GITLAB_HOST")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.Token != "glpat-test" {
		t.Errorf("Expected token 'glpat-test', got '%s'", cfg.Token)
	}

	if cfg.Host != "https://gitlab.example.io" {
		t.Errorf("Expected host 'https://gitlab.example.io', got '%s'", cfg.Host)
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
