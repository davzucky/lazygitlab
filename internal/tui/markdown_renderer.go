package tui

import (
	"strings"

	"github.com/charmbracelet/glamour"
)

const maxMarkdownRenderChars = 12000

func renderMarkdownParagraphs(input string, width int) []string {
	content := strings.TrimSpace(strings.ReplaceAll(input, "\r\n", "\n"))
	if content == "" {
		return []string{""}
	}
	if len([]rune(content)) > maxMarkdownRenderChars {
		return wrapParagraphs(content, width)
	}

	rendered, err := renderMarkdown(content, width)
	if err != nil {
		return wrapParagraphs(content, width)
	}

	return strings.Split(strings.TrimRight(rendered, "\n"), "\n")
}

func renderMarkdown(input string, width int) (string, error) {
	renderWidth := max(20, width)
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(renderWidth),
	)
	if err != nil {
		return "", err
	}
	return renderer.Render(input)
}
