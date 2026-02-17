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

func TestRenderMarkdownStructuredHeadingAndList(t *testing.T) {
	input := "# Title\n\n- first item\n- second item"
	lines := renderMarkdownStructured(input, 40)
	if len(lines) == 0 {
		t.Fatal("expected structured markdown output")
	}
	if lines[0] != "# Title" {
		t.Fatalf("expected heading line, got %q", lines[0])
	}
	if !containsLineWithPrefix(lines, "- first") {
		t.Fatalf("expected first bullet line in output, got %#v", lines)
	}
	if !containsLineWithPrefix(lines, "- second") {
		t.Fatalf("expected second bullet line in output, got %#v", lines)
	}
}

func TestRenderMarkdownStructuredBlockquote(t *testing.T) {
	lines := renderMarkdownStructured("> quoted line one\n> quoted line two", 32)
	if len(lines) == 0 {
		t.Fatal("expected structured markdown output")
	}
	if !containsLineWithPrefix(lines, "> quoted") {
		t.Fatalf("expected quoted prefix in output, got %#v", lines)
	}
}

func containsLineWithPrefix(lines []string, prefix string) bool {
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			return true
		}
	}
	return false
}
