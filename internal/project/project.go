package project

import (
	"fmt"
	"net/url"
	"os/exec"
	"strings"

	"github.com/davzucky/lazygitlab/internal/config"
)

func DetectCurrentProject(configuredHost string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("read git origin remote: %w", err)
	}

	remote := strings.TrimSpace(string(out))
	host, projectPath, err := ParseRemoteURL(remote)
	if err != nil {
		return "", err
	}

	if !hostMatches(configuredHost, host) {
		return "", fmt.Errorf("remote host %q does not match configured GitLab host", host)
	}

	return projectPath, nil
}

func ParseRemoteURL(raw string) (string, string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", "", fmt.Errorf("remote URL is empty")
	}

	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") || strings.HasPrefix(trimmed, "ssh://") {
		u, err := url.Parse(trimmed)
		if err != nil {
			return "", "", fmt.Errorf("parse remote URL: %w", err)
		}

		path := strings.TrimPrefix(u.Path, "/")
		path = strings.TrimSuffix(path, ".git")
		if u.Host == "" || path == "" {
			return "", "", fmt.Errorf("invalid remote URL: %q", raw)
		}

		return u.Host, path, nil
	}

	if strings.Contains(trimmed, "@") && strings.Contains(trimmed, ":") {
		parts := strings.SplitN(trimmed, ":", 2)
		userHost := parts[0]
		path := strings.TrimSuffix(parts[1], ".git")
		host := userHost
		if idx := strings.Index(userHost, "@"); idx >= 0 {
			host = userHost[idx+1:]
		}
		if host == "" || path == "" {
			return "", "", fmt.Errorf("invalid SCP-like remote URL: %q", raw)
		}

		return host, path, nil
	}

	return "", "", fmt.Errorf("unsupported remote URL format: %q", raw)
}

func hostMatches(configuredHost string, remoteHost string) bool {
	if strings.TrimSpace(configuredHost) == "" {
		return true
	}

	normalized, err := config.NormalizeHost(configuredHost)
	if err != nil {
		return false
	}

	u, err := url.Parse(normalized)
	if err != nil {
		return false
	}

	return strings.EqualFold(u.Host, remoteHost)
}
