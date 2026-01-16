package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/davzucky/lazygitlab/pkg/gitlab"
	"github.com/spf13/viper"
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
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return
	}

	defaultHost := v.GetString("host")
	hosts := v.GetStringMap("hosts")
	selectedHost, settings := pickGlabHost(defaultHost, hosts)

	if selectedHost != "" {
		token := getHostField(settings, "token")
		apiHost := getHostField(settings, "api_host")
		apiProtocol := getHostField(settings, "api_protocol")

		if cfg.Token == "" && token != "" {
			cfg.Token = token
		}

		if cfg.Host == "" {
			if apiHost == "" {
				apiHost = selectedHost
			}
			cfg.Host = ensureHostScheme(apiHost, apiProtocol)
		}
	}

	if cfg.Token == "" {
		if token := v.GetString("token"); token != "" {
			cfg.Token = token
		}
	}

	if cfg.Host == "" {
		if host := v.GetString("host"); host != "" {
			cfg.Host = ensureHostScheme(host, v.GetString("api_protocol"))
		}
	}
}

func pickGlabHost(defaultHost string, hosts map[string]interface{}) (string, interface{}) {
	if defaultHost != "" {
		if settings, ok := hosts[defaultHost]; ok {
			return defaultHost, settings
		}
	}

	if len(hosts) == 1 {
		for host, settings := range hosts {
			return host, settings
		}
	}

	for host, settings := range hosts {
		if getHostField(settings, "token") != "" {
			return host, settings
		}
	}

	return "", nil
}

func getHostField(settings interface{}, key string) string {
	if settings == nil {
		return ""
	}

	switch typed := settings.(type) {
	case map[string]interface{}:
		if value, ok := typed[key]; ok {
			return fmt.Sprint(value)
		}
	case map[interface{}]interface{}:
		if value, ok := typed[key]; ok {
			return fmt.Sprint(value)
		}
	}

	return ""
}

func ensureHostScheme(host, protocol string) string {
	if host == "" {
		return ""
	}

	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		return strings.TrimSuffix(host, "/")
	}

	if protocol == "" {
		protocol = "https"
	}

	return fmt.Sprintf("%s://%s", protocol, strings.TrimSuffix(host, "/"))
}

func (c *Config) isValid() bool {
	return c.Token != "" && c.Host != ""
}

func (c *Config) Validate() error {
	client, err := gitlab.NewClient(c.Token, c.Host)
	if err != nil {
		return fmt.Errorf("failed to create GitLab client: %w", err)
	}
	defer client.Close()

	user, err := client.GetCurrentUser()
	if err != nil {
		return fmt.Errorf("failed to validate token: %w", err)
	}

	if user == nil {
		return errors.New("invalid token: could not fetch user")
	}

	return nil
}
