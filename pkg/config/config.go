package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gitlab.com/gitlab-org/api/client-go"
)

type Config struct {
	Host  string
	Token string
}

func Load() (*Config, error) {
	var cfg Config

	loadFromEnv(&cfg)
	if cfg.isValid() {
		return &cfg, nil
	}

	loadFromLazyGitLabConfig(&cfg)
	if cfg.isValid() {
		return &cfg, nil
	}

	loadFromGlabConfig(&cfg)
	if cfg.isValid() {
		return &cfg, nil
	}

	return nil, errors.New("no valid GitLab configuration found. Please set GITLAB_TOKEN and GITLAB_HOST environment variables, create a config file at ~/.config/lazygitlab/config.yml, or configure glab CLI")
}

func loadFromEnv(cfg *Config) {
	if token := os.Getenv("GITLAB_TOKEN"); token != "" {
		cfg.Token = token
	}
	if host := os.Getenv("GITLAB_HOST"); host != "" {
		cfg.Host = host
	}
}

func loadFromLazyGitLabConfig(cfg *Config) {
	if cfg.isValid() {
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(homeDir + "/.config/lazygitlab/")

	if err := v.ReadInConfig(); err != nil {
		return
	}

	if token := v.GetString("token"); token != "" && cfg.Token == "" {
		cfg.Token = token
	}
	if host := v.GetString("host"); host != "" && cfg.Host == "" {
		cfg.Host = host
	}
}

func loadFromGlabConfig(cfg *Config) {
	if cfg.isValid() {
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	configPath := filepath.Join(homeDir, ".config", "glab-cli", "config.yml")
	file, err := os.ReadFile(configPath)
	if err != nil {
		return
	}

	lines := strings.Split(string(file), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "host:") {
			host := strings.TrimSpace(strings.TrimPrefix(line, "host:"))
			if host != "" && cfg.Host == "" {
				cfg.Host = host
			}
		}

		if strings.HasPrefix(line, "token:") || strings.Contains(line, "token:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				token := strings.TrimSpace(strings.Join(parts[1:], ":"))
				if token != "" && cfg.Token == "" {
					cfg.Token = token
				}
			}
		}
	}
}

func (c *Config) isValid() bool {
	return c.Token != "" && c.Host != ""
}

func (c *Config) Validate() error {
	client, err := gitlab.NewClient(c.Token, gitlab.WithBaseURL(c.Host))
	if err != nil {
		return fmt.Errorf("failed to create GitLab client: %w", err)
	}

	user, _, err := client.Users.CurrentUser()
	if err != nil {
		return fmt.Errorf("failed to validate token: %w", err)
	}

	if user == nil {
		return errors.New("invalid token: could not fetch user")
	}

	return nil
}
