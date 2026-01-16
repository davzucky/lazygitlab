package project

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func DetectProjectPath(overridePath, host string) (string, error) {
	if overridePath != "" {
		if !isValidProjectPath(overridePath) {
			return "", fmt.Errorf("invalid project path format: %s (expected: group/project or group/subgroup/project)", overridePath)
		}
		return overridePath, nil
	}

	projectPath, err := detectFromGitRemote(host)
	if err != nil {
		return "", err
	}

	return projectPath, nil
}

func detectFromGitRemote(host string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Env = os.Environ()
	rootOutput, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository (or any parent up to mount point)")
	}

	repoRoot := strings.TrimSpace(string(rootOutput))
	if repoRoot == "" {
		return "", fmt.Errorf("not a git repository (or any parent up to mount point)")
	}

	cmd = exec.Command("git", "remote", "-v")
	cmd.Dir = repoRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get git remotes: %w", err)
	}

	remoteURL, err := parseRemoteURL(string(output), host)
	if err != nil {
		return "", err
	}

	projectPath, err := extractProjectPath(remoteURL, host)
	if err != nil {
		return "", err
	}

	return projectPath, nil
}

func parseRemoteURL(output, host string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	var remoteURL string

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		url := parts[1]

		if strings.HasPrefix(line, "origin") {
			if host == "" {
				remoteURL = url
				break
			}

			matchedHost := extractHost(url)
			if matchedHost != "" && matchedHost == normalizeHost(host) {
				remoteURL = url
				break
			}
		}
	}

	if remoteURL == "" {
		return "", fmt.Errorf("no GitLab remote URL found in git remotes")
	}

	return remoteURL, nil
}

func extractProjectPath(remoteURL string, host string) (string, error) {
	if remoteURL == "" {
		return "", fmt.Errorf("remote URL is empty")
	}

	if strings.HasPrefix(remoteURL, "git@") {
		re := regexp.MustCompile(`^git@([^:]+):(.+?)(?:\.git)?$`)
		matches := re.FindStringSubmatch(remoteURL)
		if len(matches) < 3 {
			return "", fmt.Errorf("invalid SSH remote URL format: %s", remoteURL)
		}

		hostPart := matches[1]
		normalizedHost := normalizeHost(host)
		if normalizedHost != "" && hostPart != normalizedHost {
			return "", fmt.Errorf("remote host %s does not match configured host %s", hostPart, normalizedHost)
		}

		projectPath := strings.TrimSuffix(matches[2], ".git")
		if !isValidProjectPath(projectPath) {
			return "", fmt.Errorf("invalid project path: %s", projectPath)
		}
		return projectPath, nil
	}

	if strings.HasPrefix(remoteURL, "http://") || strings.HasPrefix(remoteURL, "https://") {
		re := regexp.MustCompile(`^https?://([^/]+)/(.+)$`)
		matches := re.FindStringSubmatch(remoteURL)
		if len(matches) < 3 {
			return "", fmt.Errorf("invalid HTTPS remote URL format: %s", remoteURL)
		}

		remoteHost := matches[1]
		normalizedHost := normalizeHost(host)
		if normalizedHost != "" && remoteHost != normalizedHost {
			return "", fmt.Errorf("remote host %s does not match configured host %s", remoteHost, normalizedHost)
		}

		projectPath := strings.TrimSuffix(matches[2], ".git")
		if !isValidProjectPath(projectPath) {
			return "", fmt.Errorf("invalid project path: %s", projectPath)
		}
		return projectPath, nil
	}

	return "", fmt.Errorf("unsupported remote URL format: %s", remoteURL)
}

func normalizeHost(host string) string {
	normalizedHost := strings.TrimPrefix(strings.TrimPrefix(host, "https://"), "http://")
	return strings.TrimSuffix(normalizedHost, "/")
}

func extractHost(remoteURL string) string {
	remoteURL = strings.TrimSpace(remoteURL)
	if remoteURL == "" {
		return ""
	}

	if strings.HasPrefix(remoteURL, "git@") {
		re := regexp.MustCompile(`^git@([^:]+):`)
		matches := re.FindStringSubmatch(remoteURL)
		if len(matches) > 1 {
			return matches[1]
		}
		return ""
	}

	if !strings.HasPrefix(remoteURL, "http://") && !strings.HasPrefix(remoteURL, "https://") {
		return ""
	}

	parsed, err := url.Parse(remoteURL)
	if err != nil {
		return ""
	}

	host := parsed.Hostname()
	if host == "" {
		return ""
	}

	return host
}

func isValidProjectPath(path string) bool {
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return false
	}

	for _, part := range parts {
		if part == "" {
			return false
		}
	}

	return true
}
