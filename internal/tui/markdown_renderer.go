package tui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
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

	if rendered := renderMarkdownStructured(content, width); len(rendered) > 0 {
		return rendered
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

func renderMarkdownStructured(input string, width int) []string {
	if width <= 0 {
		return nil
	}
	source := []byte(input)
	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(source))
	lines := renderMarkdownBlocks(doc, source, max(8, width), "")
	if len(lines) == 0 {
		return nil
	}
	return trimBlankEdges(lines)
}

func renderMarkdownBlocks(parent ast.Node, source []byte, width int, prefix string) []string {
	out := make([]string, 0, 16)
	for node := parent.FirstChild(); node != nil; node = node.NextSibling() {
		switch typed := node.(type) {
		case *ast.Heading:
			headingPrefix := strings.Repeat("#", typed.Level) + " "
			out = append(out, wrapPrefixedText(extractNodeText(typed, source), width, prefix+headingPrefix)...)
			out = append(out, "")
		case *ast.Paragraph:
			out = append(out, wrapPrefixedText(extractNodeText(typed, source), width, prefix)...)
			out = append(out, "")
		case *ast.Blockquote:
			out = append(out, renderMarkdownBlocks(typed, source, width, prefix+"> ")...)
			out = append(out, "")
		case *ast.List:
			out = append(out, renderMarkdownList(typed, source, width, prefix)...)
			out = append(out, "")
		case *ast.FencedCodeBlock:
			lang := strings.TrimSpace(string(typed.Language(source)))
			if lang == "" {
				out = append(out, prefix+"```")
			} else {
				out = append(out, prefix+"```"+lang)
			}
			for i := 0; i < typed.Lines().Len(); i++ {
				segment := typed.Lines().At(i)
				line := strings.TrimRight(string(segment.Value(source)), "\r\n")
				out = append(out, prefix+line)
			}
			out = append(out, prefix+"```")
			out = append(out, "")
		case *ast.CodeBlock:
			out = append(out, prefix+"```")
			for i := 0; i < typed.Lines().Len(); i++ {
				segment := typed.Lines().At(i)
				line := strings.TrimRight(string(segment.Value(source)), "\r\n")
				out = append(out, prefix+line)
			}
			out = append(out, prefix+"```")
			out = append(out, "")
		case *ast.ThematicBreak:
			out = append(out, prefix+"---")
			out = append(out, "")
		default:
			textValue := strings.TrimSpace(extractNodeText(node, source))
			if textValue != "" {
				out = append(out, wrapPrefixedText(textValue, width, prefix)...)
				out = append(out, "")
			}
		}
	}
	return out
}

func renderMarkdownList(list *ast.List, source []byte, width int, prefix string) []string {
	out := make([]string, 0, 8)
	index := 0
	for item := list.FirstChild(); item != nil; item = item.NextSibling() {
		listItem, ok := item.(*ast.ListItem)
		if !ok {
			continue
		}
		marker := "- "
		if list.IsOrdered() {
			marker = fmt.Sprintf("%d. ", list.Start+index)
		}
		out = append(out, renderMarkdownListItem(listItem, source, width, prefix, marker)...)
		index++
	}
	return out
}

func renderMarkdownListItem(item *ast.ListItem, source []byte, width int, prefix string, marker string) []string {
	itemLines := make([]string, 0, 4)
	firstPrefix := prefix + marker
	continuationPrefix := prefix + strings.Repeat(" ", len(marker))
	for child := item.FirstChild(); child != nil; child = child.NextSibling() {
		switch typed := child.(type) {
		case *ast.Paragraph:
			itemLines = append(itemLines, wrapPrefixedText(extractNodeText(typed, source), width, firstPrefix)...)
			firstPrefix = continuationPrefix
		case *ast.List:
			itemLines = append(itemLines, renderMarkdownList(typed, source, width, continuationPrefix)...)
		default:
			textValue := strings.TrimSpace(extractNodeText(child, source))
			if textValue != "" {
				itemLines = append(itemLines, wrapPrefixedText(textValue, width, firstPrefix)...)
				firstPrefix = continuationPrefix
			}
		}
	}
	return itemLines
}

func extractNodeText(node ast.Node, source []byte) string {
	var buffer bytes.Buffer
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch typed := child.(type) {
		case *ast.Text:
			buffer.Write(typed.Segment.Value(source))
			if typed.HardLineBreak() || typed.SoftLineBreak() {
				buffer.WriteByte('\n')
			}
		case *ast.CodeSpan:
			buffer.WriteByte('`')
			buffer.WriteString(strings.TrimSpace(string(typed.Text(source))))
			buffer.WriteByte('`')
		case *ast.AutoLink:
			buffer.WriteString(string(typed.Label(source)))
		case *ast.RawHTML:
			for i := 0; i < typed.Segments.Len(); i++ {
				segment := typed.Segments.At(i)
				buffer.Write(segment.Value(source))
			}
		default:
			buffer.WriteString(extractNodeText(typed, source))
		}
	}
	if buffer.Len() == 0 {
		if textNode, ok := node.(*ast.Text); ok {
			return string(textNode.Segment.Value(source))
		}
	}
	return strings.TrimSpace(buffer.String())
}

func wrapPrefixedText(input string, width int, prefix string) []string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil
	}
	effectiveWidth := max(1, width-len(prefix))
	wrapped := wrapParagraphs(trimmed, effectiveWidth)
	out := make([]string, 0, len(wrapped))
	for i, line := range wrapped {
		line = strings.TrimSpace(line)
		if i == 0 {
			out = append(out, prefix+line)
			continue
		}
		out = append(out, strings.Repeat(" ", len(prefix))+line)
	}
	return out
}

func trimBlankEdges(lines []string) []string {
	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	end := len(lines)
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	return lines[start:end]
}
