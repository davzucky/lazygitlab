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

func TestRenderMermaidDiagramSupportsSubgraphBlocks(t *testing.T) {
	input := "flowchart LR\n" +
		"subgraph ST1[\"Stage 1\"]\n" +
		"direction TB\n" +
		"J1[\"job one\"]\n" +
		"J2[\"job two\"]\n" +
		"end\n" +
		"subgraph ST2[\"Stage 2\"]\n" +
		"J3[\"job three\"]\n" +
		"end\n" +
		"J1 --> J2 --> J3\n" +
		"classDef gate fill:#fff7ed,stroke:#f97316;\n" +
		"class J2 gate"

	lines, err := renderMermaidDiagram(input)
	if err != nil {
		t.Fatalf("expected subgraph diagram to render, got error: %v", err)
	}
	output := strings.Join(lines, "\n")
	if !strings.Contains(output, "job one") || !strings.Contains(output, "job two") || !strings.Contains(output, "job three") {
		t.Fatalf("expected subgraph job labels in output, got %q", output)
	}
	if !strings.Contains(output, "▶") {
		t.Fatalf("expected rendered edges in output, got %q", output)
	}
}
