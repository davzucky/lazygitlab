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

func TestParseColorFGBG(t *testing.T) {
	t.Parallel()

	fg, bg, ok := parseColorFGBG("15;0")
	if !ok {
		t.Fatal("expected COLORFGBG parse to succeed")
	}
	if fg != "#ffffff" {
		t.Fatalf("fg = %q want %q", fg, "#ffffff")
	}
	if bg != "#000000" {
		t.Fatalf("bg = %q want %q", bg, "#000000")
	}
}

func TestParseColorFGBGSupportsColonFormat(t *testing.T) {
	t.Parallel()

	fg, bg, ok := parseColorFGBG("0;15:0")
	if !ok {
		t.Fatal("expected COLORFGBG mixed format parse to succeed")
	}
	if fg != "#ffffff" || bg != "#000000" {
		t.Fatalf("parsed colors = (%q, %q), expected (#ffffff, #000000)", fg, bg)
	}
}

func TestResolveThemeFGAndBGUsesExplicitEnv(t *testing.T) {
	t.Setenv(envThemeForeground, "#e6e6e6")
	t.Setenv(envThemeBackground, "#111111")
	t.Setenv("COLORFGBG", "15;0")

	fg, bg, ok := resolveThemeFGAndBG()
	if !ok {
		t.Fatal("expected explicit theme env values to resolve")
	}
	if fg != "#e6e6e6" || bg != "#111111" {
		t.Fatalf("resolved colors = (%q, %q), expected (#e6e6e6, #111111)", fg, bg)
	}
}
