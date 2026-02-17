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
	if stripANSI(lines[0]) != "# Title" {
		t.Fatalf("expected heading line, got %q", stripANSI(lines[0]))
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

func TestRenderMarkdownStructuredBlockquoteWrap(t *testing.T) {
	input := "> this is a very long blockquote line that should wrap to multiple lines"
	lines := renderMarkdownStructured(input, 24)
	if len(lines) < 2 {
		t.Fatalf("expected wrapped blockquote output, got %#v", lines)
	}
	for _, line := range lines {
		clean := stripANSI(line)
		if strings.TrimSpace(clean) == "" {
			continue
		}
		if !strings.HasPrefix(clean, "> ") {
			t.Fatalf("expected blockquote continuation to keep prefix, got line %q in %#v", line, lines)
		}
	}
}

func TestRenderMarkdownStructuredCodeBlockIsHighlighted(t *testing.T) {
	input := "```go\nfmt.Println(\"hi\")\n```"
	lines := renderMarkdownStructured(input, 80)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "\x1b[") {
		t.Fatalf("expected ANSI color escapes in rendered code block, got %#v", lines)
	}
	if !containsLineWithPrefix(lines, "```") {
		t.Fatalf("expected code fence lines in output, got %#v", lines)
	}
}

func TestRenderMarkdownStructuredNestedLists(t *testing.T) {
	input := "- parent\n  - child\n1. top\n   1. nested"
	lines := renderMarkdownStructured(input, 40)
	if !containsLineWithPrefix(lines, "- parent") {
		t.Fatalf("expected parent bullet, got %#v", lines)
	}
	if !containsLineWithPrefix(lines, "- child") {
		t.Fatalf("expected nested bullet, got %#v", lines)
	}
	if !containsLineWithPrefix(lines, "1. top") {
		t.Fatalf("expected ordered parent item, got %#v", lines)
	}
	if !containsLineWithPrefix(lines, "1. nested") {
		t.Fatalf("expected nested ordered item, got %#v", lines)
	}
}

func containsLineWithPrefix(lines []string, prefix string) bool {
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(stripANSI(line)), prefix) {
			return true
		}
	}
	return false
}
