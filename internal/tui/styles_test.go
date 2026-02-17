package tui

import (
	"fmt"
	"testing"
)

func TestIsLikelyTerminalColor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  bool
	}{
		{input: "39", want: true},
		{input: "255", want: true},
		{input: "#00aaff", want: true},
		{input: "#0af", want: false},
		{input: "blue", want: false},
		{input: "", want: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			if got := isLikelyTerminalColor(tc.input); got != tc.want {
				t.Fatalf("isLikelyTerminalColor(%q) = %v want %v", tc.input, got, tc.want)
			}
		})
	}
}

func TestResolveAccentColorUsesEnvOverride(t *testing.T) {
	t.Setenv(envAccentColor, "#00aaff")

	got := fmt.Sprint(resolveAccentColor())
	if got != "#00aaff" {
		t.Fatalf("resolveAccentColor() = %q want %q", got, "#00aaff")
	}
}

func TestResolveAccentColorIgnoresInvalidOverride(t *testing.T) {
	t.Setenv(envAccentColor, "invalid")

	got := fmt.Sprint(resolveAccentColor())
	if got == "invalid" {
		t.Fatal("expected invalid accent override to be ignored")
	}
}
