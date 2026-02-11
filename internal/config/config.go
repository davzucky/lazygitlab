package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	EnvGitLabToken = "GITLAB_TOKEN"
	EnvGitLabHost  = "GITLAB_HOST"
)

var errHomeNotFound = errors.New("home directory not found")

type Config struct {
	Host        string `yaml:"host"`
	Token       string `yaml:"token"`
	LastProject string `yaml:"last_project,omitempty"`
	Debug       bool   `yaml:"debug,omitempty"`
}

type glabConfig struct {
	Hosts map[string]glabHostConfig `yaml:"hosts"`
}

type glabHostConfig struct {
	Token string `yaml:"token"`
}

func (c Config) NeedsSetup() bool {
	return strings.TrimSpace(c.Token) == "" || strings.TrimSpace(c.Host) == ""
}

func Load() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, errHomeNotFound
	}

	cfg := Config{}

	glabPath := filepath.Join(home, ".config", "glab-cli", "config.yml")
	if glabCfg, err := loadFromGlab(glabPath); err == nil {
		cfg = merge(cfg, glabCfg)
	}

	lazyPath := filepath.Join(home, ".config", "lazygitlab", "config.yml")
	if lazyCfg, err := loadFromLazyConfig(lazyPath); err == nil {
		cfg = merge(cfg, lazyCfg)
	}

	envCfg := Config{
		Host:  strings.TrimSpace(os.Getenv(EnvGitLabHost)),
		Token: strings.TrimSpace(os.Getenv(EnvGitLabToken)),
	}
	cfg = merge(cfg, envCfg)

	if cfg.Host != "" {
		normalized, err := NormalizeHost(cfg.Host)
		if err != nil {
			return Config{}, fmt.Errorf("invalid host: %w", err)
		}
		cfg.Host = normalized
	}

	return cfg, nil
}

func Save(cfg Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return errHomeNotFound
	}

	normalized := cfg
	if normalized.Host != "" {
		host, err := NormalizeHost(normalized.Host)
		if err != nil {
			return fmt.Errorf("normalize host: %w", err)
		}
		normalized.Host = host
	}

	configDir := filepath.Join(home, ".config", "lazygitlab")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	path := filepath.Join(configDir, "config.yml")
	data, err := yaml.Marshal(normalized)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

func NormalizeHost(host string) (string, error) {
	trimmed := strings.TrimSpace(host)
	if trimmed == "" {
		return "", errors.New("host is empty")
	}

	if !strings.Contains(trimmed, "://") {
		trimmed = "https://" + trimmed
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("parse host: %w", err)
	}

	if parsed.Host == "" {
		return "", errors.New("missing host")
	}

	path := strings.TrimSuffix(parsed.Path, "/")
	if path == "" || path == "/" {
		path = "/api/v4"
	} else if !strings.HasSuffix(path, "/api/v4") {
		path += "/api/v4"
	}
	parsed.Path = path

	return parsed.String(), nil
}

func loadFromGlab(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	parsed := glabConfig{}
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return Config{}, err
	}

	if len(parsed.Hosts) == 0 {
		return Config{}, errors.New("no hosts in glab config")
	}

	if hostCfg, ok := parsed.Hosts["gitlab.com"]; ok {
		if hostCfg.Token != "" {
			return Config{Host: "https://gitlab.com/api/v4", Token: hostCfg.Token}, nil
		}
	}

	for host, hostCfg := range parsed.Hosts {
		if strings.TrimSpace(hostCfg.Token) == "" {
			continue
		}

		normalizedHost, err := NormalizeHost(host)
		if err != nil {
			continue
		}

		return Config{Host: normalizedHost, Token: strings.TrimSpace(hostCfg.Token)}, nil
	}

	return Config{}, errors.New("no valid glab host entries")
}

func loadFromLazyConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func merge(base Config, override Config) Config {
	merged := base
	if strings.TrimSpace(override.Host) != "" {
		merged.Host = strings.TrimSpace(override.Host)
	}
	if strings.TrimSpace(override.Token) != "" {
		merged.Token = strings.TrimSpace(override.Token)
	}
	if strings.TrimSpace(override.LastProject) != "" {
		merged.LastProject = strings.TrimSpace(override.LastProject)
	}
	merged.Debug = merged.Debug || override.Debug
	return merged
}
