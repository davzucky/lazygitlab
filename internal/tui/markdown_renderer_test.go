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

func TestRenderMarkdownStructuredMermaidBlockRendersDiagram(t *testing.T) {
	input := "```mermaid\nflowchart LR\nA[Start] --> B[End]\n```"
	lines := renderMarkdownStructured(input, 100)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "Start") || !strings.Contains(joined, "End") {
		t.Fatalf("expected rendered mermaid labels, got %#v", lines)
	}
	if !strings.Contains(joined, "▶") {
		t.Fatalf("expected diagram arrow output, got %#v", lines)
	}
	if strings.Contains(joined, "Mermaid not supported in this format") {
		t.Fatalf("did not expect fallback warning for supported diagram, got %#v", lines)
	}
}

func TestRenderMarkdownStructuredMermaidBlockFallsBackToSource(t *testing.T) {
	input := "```mermaid\nsequenceDiagram\nA->>B: hi\n```"
	lines := renderMarkdownStructured(input, 80)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "Mermaid not supported in this format; showing source.") {
		t.Fatalf("expected fallback warning, got %#v", lines)
	}
	if !strings.Contains(joined, "sequenceDiagram") || !strings.Contains(joined, "A->>B: hi") {
		t.Fatalf("expected raw mermaid source in fallback, got %#v", lines)
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

func TestRenderMarkdownStructuredTable(t *testing.T) {
	input := strings.Join([]string{
		"| Name | Status |",
		"| :--- | ---: |",
		"| parser | done |",
		"| tests | pending |",
	}, "\n")

	lines := renderMarkdownStructured(input, 60)
	if len(lines) == 0 {
		t.Fatal("expected structured markdown output")
	}
	if !containsLineWithPrefix(lines, "| Name") {
		t.Fatalf("expected table header row, got %#v", lines)
	}
	if !containsLineWithPrefix(lines, "| :") {
		t.Fatalf("expected table separator row with alignment markers, got %#v", lines)
	}
	if !containsLineWithPrefix(lines, "| parser") {
		t.Fatalf("expected table body row, got %#v", lines)
	}
}

func TestRenderMarkdownStructuredTableNarrowWidth(t *testing.T) {
	input := strings.Join([]string{
		"| Column One | Column Two | Column Three |",
		"| :--- | ---: | :---: |",
		"| alpha value | beta value | gamma value |",
	}, "\n")

	width := 22
	lines := renderMarkdownStructured(input, width)
	if len(lines) == 0 {
		t.Fatal("expected structured markdown output")
	}
	if !containsLineWithPrefix(lines, "|") {
		t.Fatalf("expected table output with pipe-delimited lines, got %#v", lines)
	}
	for _, line := range lines {
		clean := stripANSI(line)
		if strings.TrimSpace(clean) == "" {
			continue
		}
		if len([]rune(clean)) > width {
			t.Fatalf("expected rendered table line width <= %d, got %d for line %q", width, len([]rune(clean)), clean)
		}
	}
}

func TestRenderMarkdownStructuredTableCodeCellKeepsContentWhenSpaceAllows(t *testing.T) {
	input := strings.Join([]string{
		"| Package | Type | Update | Change |",
		"|---|---|---|---|",
		"| gcc |  | patch | `15.2.0-r8` → `15.2.0-r9` |",
	}, "\n")

	lines := renderMarkdownStructured(input, 120)
	joined := strings.Join(lines, "\n")
	if strings.Contains(joined, "15.2...") {
		t.Fatalf("expected code cell not to be aggressively truncated, got %#v", lines)
	}
	if !strings.Contains(joined, "`15.2.0-r8` → `15.2.0-r9`") {
		t.Fatalf("expected full code change content in table cell, got %#v", lines)
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
