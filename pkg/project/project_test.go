package project

import (
	"testing"
)

func TestDetectProjectPath_Override(t *testing.T) {
	tests := []struct {
		name        string
		override    string
		expected    string
		expectError bool
	}{
		{
			name:        "valid override",
			override:    "group/project",
			expected:    "group/project",
			expectError: false,
		},
		{
			name:        "valid override with subgroups",
			override:    "group/subgroup/project",
			expected:    "group/subgroup/project",
			expectError: false,
		},
		{
			name:        "invalid override single part",
			override:    "project",
			expected:    "",
			expectError: true,
		},
		{
			name:        "invalid override empty string",
			override:    "",
			expected:    "",
			expectError: false, // Will fall back to git detection
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DetectProjectPath(tt.override)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil && tt.override != "" {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.expected && tt.override != "" {
					t.Errorf("expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestExtractProjectPath(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expected    string
		expectError bool
	}{
		{
			name:        "SSH URL",
			url:         "git@gitlab.com:group/project.git",
			expected:    "group/project",
			expectError: false,
		},
		{
			name:        "SSH URL with subgroups",
			url:         "git@gitlab.com:group/subgroup/project.git",
			expected:    "group/subgroup/project",
			expectError: false,
		},
		{
			name:        "HTTPS URL",
			url:         "https://gitlab.com/group/project.git",
			expected:    "group/project",
			expectError: false,
		},
		{
			name:        "HTTPS URL with subgroups",
			url:         "https://gitlab.com/group/subgroup/project.git",
			expected:    "group/subgroup/project",
			expectError: false,
		},
		{
			name:        "HTTPS URL without .git",
			url:         "https://gitlab.com/group/project",
			expected:    "group/project",
			expectError: false,
		},
		{
			name:        "invalid format",
			url:         "https://github.com/user/repo",
			expected:    "",
			expectError: true,
		},
		{
			name:        "malformed SSH URL",
			url:         "git@gitlab.com/project.git",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractProjectPath(tt.url)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestIsValidProjectPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "valid group/project",
			path:     "group/project",
			expected: true,
		},
		{
			name:     "valid with subgroups",
			path:     "group/subgroup/project",
			expected: true,
		},
		{
			name:     "valid multiple subgroups",
			path:     "group/subgroup/subsubgroup/project",
			expected: true,
		},
		{
			name:     "invalid single part",
			path:     "project",
			expected: false,
		},
		{
			name:     "invalid empty string",
			path:     "",
			expected: false,
		},
		{
			name:     "invalid empty segment",
			path:     "group//project",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidProjectPath(tt.path)
			if result != tt.expected {
				t.Errorf("expected %v, got %v for path %s", tt.expected, result, tt.path)
			}
		})
	}
}

func TestParseRemoteURL(t *testing.T) {
	tests := []struct {
		name        string
		output      string
		expected    string
		expectError bool
	}{
		{
			name: "valid SSH output",
			output: `origin	git@gitlab.com:group/project.git (fetch)
origin	git@gitlab.com:group/project.git (push)`,
			expected:    "git@gitlab.com:group/project.git",
			expectError: false,
		},
		{
			name: "valid HTTPS output",
			output: `origin	https://gitlab.com/group/project.git (fetch)
origin	https://gitlab.com/group/project.git (push)`,
			expected:    "https://gitlab.com/group/project.git",
			expectError: false,
		},
		{
			name: "multiple remotes",
			output: `origin	git@gitlab.com:group/project.git (fetch)
origin	git@gitlab.com:group/project.git (push)
upstream	git@gitlab.com:other/repo.git (fetch)
upstream	git@gitlab.com:other/repo.git (push)`,
			expected:    "git@gitlab.com:group/project.git",
			expectError: false,
		},
		{
			name:        "no GitLab remote",
			output:      `origin	https://github.com/user/repo.git (fetch)`,
			expected:    "",
			expectError: true,
		},
		{
			name:        "empty output",
			output:      "",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseRemoteURL(tt.output)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("expected %s, got %s", tt.expected, result)
				}
			}
		})
	}
}
