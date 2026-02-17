package tui

import (
	"strings"
	"testing"
)

func TestRenderMarkdownParagraphsEmpty(t *testing.T) {
	lines := renderMarkdownParagraphs("   ", 40)
	if len(lines) != 1 || lines[0] != "" {
		t.Fatalf("expected single empty line, got %#v", lines)
	}
}

func TestRenderMarkdownParagraphsLongInputFallsBackToWrapping(t *testing.T) {
	content := strings.Repeat("word ", maxMarkdownRenderChars)
	lines := renderMarkdownParagraphs(content, 20)
	if len(lines) == 0 {
		t.Fatal("expected wrapped output for long markdown input")
	}
	if len(lines[0]) > 20 {
		t.Fatalf("expected wrapped line width <= 20, got %d", len(lines[0]))
	}
}
