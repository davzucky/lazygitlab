package tui

import (
	"strings"
	"testing"
)

func TestRenderMermaidDiagramFlowchartLR(t *testing.T) {
	input := "flowchart LR\nA[Start] --> B[End]"
	lines, err := renderMermaidDiagram(input)
	if err != nil {
		t.Fatalf("expected mermaid render success, got error: %v", err)
	}
	output := strings.Join(lines, "\n")
	if !strings.Contains(output, "Start") {
		t.Fatalf("expected Start label in diagram, got %q", output)
	}
	if !strings.Contains(output, "End") {
		t.Fatalf("expected End label in diagram, got %q", output)
	}
	if !strings.Contains(output, "▶") {
		t.Fatalf("expected right arrow in LR diagram, got %q", output)
	}
}

func TestRenderMermaidDiagramFlowchartTB(t *testing.T) {
	input := "flowchart TB\nA[Start] --> B[End]"
	lines, err := renderMermaidDiagram(input)
	if err != nil {
		t.Fatalf("expected mermaid render success, got error: %v", err)
	}
	output := strings.Join(lines, "\n")
	if !strings.Contains(output, "▼") {
		t.Fatalf("expected down arrow in TB diagram, got %q", output)
	}
}

func TestRenderMermaidDiagramUnsupportedSyntax(t *testing.T) {
	input := "sequenceDiagram\nA->>B: hi"
	_, err := renderMermaidDiagram(input)
	if err == nil {
		t.Fatal("expected unsupported syntax error")
	}
}

func TestRenderMermaidDiagramCycleNotSupported(t *testing.T) {
	input := "flowchart LR\nA --> B\nB --> A"
	_, err := renderMermaidDiagram(input)
	if err == nil {
		t.Fatal("expected cycle not supported error")
	}
	if !strings.Contains(err.Error(), "not supported") {
		t.Fatalf("expected not supported error, got %v", err)
	}
}
