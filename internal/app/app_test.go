package app

import (
	"bytes"
	"strings"
	"testing"
)

func TestIsInteractiveSessionFalseWhenNil(t *testing.T) {
	t.Parallel()

	if isInteractiveSession(nil, nil) {
		t.Fatal("expected non-interactive session for nil files")
	}
}

func TestRenderNonInteractiveSummary(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	renderNonInteractiveSummary(&out, "https://gitlab.example.com", "group/project", "alice")

	text := out.String()
	checks := []string{
		"lazygitlab non-interactive summary",
		"host: https://gitlab.example.com",
		"project: group/project",
		"user: alice",
		"tip: run in an interactive terminal to open the TUI",
	}
	for _, check := range checks {
		if !strings.Contains(text, check) {
			t.Fatalf("expected output to contain %q, got %q", check, text)
		}
	}
}
