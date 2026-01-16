package project

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const (
	gitlabHost = "gitlab.com"
)

func DetectProjectPath(overridePath string) (string, error) {
	if overridePath != "" {
		if !isValidProjectPath(overridePath) {
			return "", fmt.Errorf("invalid project path format: %s (expected: group/project or group/subgroup/project)", overridePath)
		}
		return overridePath, nil
	}

	projectPath, err := detectFromGitRemote()
	if err != nil {
		return "", err
	}

	return projectPath, nil
}

func detectFromGitRemote() (string, error) {
	gitDir := ".git"
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return "", fmt.Errorf("not a git repository (or any parent up to mount point)")
	}

	cmd := exec.Command("git", "remote", "-v")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get git remotes: %w", err)
	}

	remoteURL, err := parseRemoteURL(string(output))
	if err != nil {
		return "", err
	}

	projectPath, err := extractProjectPath(remoteURL)
	if err != nil {
		return "", err
	}

	return projectPath, nil
}

func parseRemoteURL(output string) (string, error) {
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

		if strings.Contains(url, gitlabHost) && strings.HasPrefix(line, "origin") {
			remoteURL = url
			break
		}
	}

	if remoteURL == "" {
		return "", fmt.Errorf("no GitLab remote URL found in git remotes")
	}

	return remoteURL, nil
}

func extractProjectPath(url string) (string, error) {
	var projectPath string

	if strings.HasPrefix(url, "git@") {
		re := regexp.MustCompile(`git@` + regexp.QuoteMeta(gitlabHost) + `:(.+)\.git$`)
		matches := re.FindStringSubmatch(url)
		if len(matches) < 2 {
			return "", fmt.Errorf("invalid SSH remote URL format: %s", url)
		}
		projectPath = matches[1]
	} else if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		re := regexp.MustCompile(`https?://` + regexp.QuoteMeta(gitlabHost) + `/([^/]+/.+)$`)
		matches := re.FindStringSubmatch(url)
		if len(matches) < 2 {
			return "", fmt.Errorf("invalid HTTPS remote URL format: %s", url)
		}
		projectPath = strings.TrimSuffix(matches[1], ".git")
	} else {
		return "", fmt.Errorf("unsupported remote URL format: %s", url)
	}

	if !isValidProjectPath(projectPath) {
		return "", fmt.Errorf("invalid project path: %s", projectPath)
	}

	return projectPath, nil
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
