package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeHost(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "plain host", input: "gitlab.com", want: "https://gitlab.com/api/v4"},
		{name: "already api", input: "https://gitlab.example.com/api/v4", want: "https://gitlab.example.com/api/v4"},
		{name: "with path", input: "https://gitlab.example.com/custom", want: "https://gitlab.example.com/custom/api/v4"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NormalizeHost(tt.input)
			if err != nil {
				t.Fatalf("NormalizeHost() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("NormalizeHost() = %q want %q", got, tt.want)
			}
		})
	}
}

func TestLoadPrecedence(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	glabDir := filepath.Join(home, ".config", "glab-cli")
	if err := os.MkdirAll(glabDir, 0o755); err != nil {
		t.Fatal(err)
	}
	glabConfig := "hosts:\n  gitlab.com:\n    token: glab-token\n"
	if err := os.WriteFile(filepath.Join(glabDir, "config.yml"), []byte(glabConfig), 0o600); err != nil {
		t.Fatal(err)
	}

	lazyDir := filepath.Join(home, ".config", "lazygitlab")
	if err := os.MkdirAll(lazyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	lazyConfig := "host: https://gitlab.self/api/v4\ntoken: lazy-token\n"
	if err := os.WriteFile(filepath.Join(lazyDir, "config.yml"), []byte(lazyConfig), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv(EnvGitLabHost, "env.gitlab.local")
	t.Setenv(EnvGitLabToken, "env-token")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Token != "env-token" {
		t.Fatalf("Token = %q want env-token", cfg.Token)
	}
	if cfg.Host != "https://env.gitlab.local/api/v4" {
		t.Fatalf("Host = %q want https://env.gitlab.local/api/v4", cfg.Host)
	}
}
