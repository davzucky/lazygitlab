package tui

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

const maxMarkdownRenderChars = 12000

var (
	ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)
)

func renderMarkdownParagraphs(input string, width int, mdStyles markdownStyles) []string {
	content := strings.TrimSpace(strings.ReplaceAll(input, "\r\n", "\n"))
	if content == "" {
		return []string{""}
	}
	if utf8.RuneCountInString(content) > maxMarkdownRenderChars {
		return wrapParagraphs(content, width)
	}

	if rendered := renderMarkdownStructured(content, width, mdStyles); len(rendered) > 0 {
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

func renderMarkdownStructured(input string, width int, mdStyles markdownStyles) []string {
	if width <= 0 {
		return nil
	}
	source := []byte(input)
	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(source))
	lines := renderMarkdownBlocks(doc, source, max(8, width), "", mdStyles)
	if len(lines) == 0 {
		return nil
	}
	return trimBlankEdges(lines)
}

func renderMarkdownBlocks(parent ast.Node, source []byte, width int, prefix string, mdStyles markdownStyles) []string {
	out := make([]string, 0, 16)
	for node := parent.FirstChild(); node != nil; node = node.NextSibling() {
		switch typed := node.(type) {
		case *ast.Heading:
			headingPrefix := strings.Repeat("#", typed.Level) + " "
			headingText := mdStyles.heading.Render(renderInlineChildren(typed, source, mdStyles))
			out = append(out, wrapPrefixedText(headingText, width, prefix+headingPrefix, strings.Repeat(" ", len(prefix)+len(headingPrefix)), mdStyles)...)
			out = append(out, "")
		case *ast.Paragraph:
			out = append(out, wrapPrefixedText(renderInlineChildren(typed, source, mdStyles), width, prefix, prefix, mdStyles)...)
			out = append(out, "")
		case *ast.Blockquote:
			out = append(out, renderMarkdownBlocks(typed, source, width, prefix+"> ", mdStyles)...)
			out = append(out, "")
		case *ast.List:
			out = append(out, renderMarkdownList(typed, source, width, prefix, mdStyles)...)
			out = append(out, "")
		case *ast.FencedCodeBlock:
			lang := strings.TrimSpace(string(typed.Language(source)))
			if strings.EqualFold(lang, "mermaid") {
				out = append(out, renderMermaidBlockLines(typed.Lines(), source, prefix, width, mdStyles)...)
				out = append(out, "")
				continue
			}
			if lang == "" {
				out = append(out, mdStyles.fence.Render(prefix+"```"))
			} else {
				out = append(out, mdStyles.fence.Render(prefix+"```"+lang))
			}
			out = append(out, renderHighlightedCodeBlockLines(typed.Lines(), source, prefix, lang, mdStyles)...)
			out = append(out, mdStyles.fence.Render(prefix+"```"))
			out = append(out, "")
		case *ast.CodeBlock:
			out = append(out, mdStyles.fence.Render(prefix+"```"))
			out = append(out, renderHighlightedCodeBlockLines(typed.Lines(), source, prefix, "", mdStyles)...)
			out = append(out, mdStyles.fence.Render(prefix+"```"))
			out = append(out, "")
		case *ast.ThematicBreak:
			out = append(out, mdStyles.fence.Render(prefix+"---"))
			out = append(out, "")
		default:
			textValue := strings.TrimSpace(renderInlineChildren(node, source, mdStyles))
			if textValue != "" {
				out = append(out, wrapPrefixedText(textValue, width, prefix, prefix, mdStyles)...)
				out = append(out, "")
			}
		}
	}
	return out
}

func renderCodeBlockLines(lines *text.Segments, source []byte, prefix string, mdStyles markdownStyles) []string {
	out := make([]string, 0, lines.Len())
	for i := 0; i < lines.Len(); i++ {
		segment := lines.At(i)
		line := strings.TrimRight(string(segment.Value(source)), "\r\n")
		out = append(out, prefix+mdStyles.code.Render(line))
	}
	return out
}

func renderHighlightedCodeBlockLines(lines *text.Segments, source []byte, prefix string, language string, mdStyles markdownStyles) []string {
	codeLines := make([]string, 0, lines.Len())
	for i := 0; i < lines.Len(); i++ {
		segment := lines.At(i)
		codeLines = append(codeLines, strings.TrimRight(string(segment.Value(source)), "\r\n"))
	}
	code := strings.Join(codeLines, "\n")
	if strings.TrimSpace(code) == "" {
		return renderCodeBlockLines(lines, source, prefix, mdStyles)
	}

	var buf bytes.Buffer
	lang := language
	if strings.TrimSpace(lang) == "" {
		lang = "plaintext"
	}
	if err := quick.Highlight(&buf, code, lang, "terminal16m", "monokai"); err != nil {
		return renderCodeBlockLines(lines, source, prefix, mdStyles)
	}
	highlighted := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	out := make([]string, 0, len(highlighted))
	for _, line := range highlighted {
		out = append(out, prefix+line)
	}
	return out
}

func renderMermaidBlockLines(lines *text.Segments, source []byte, prefix string, width int, mdStyles markdownStyles) []string {
	out := []string{
		mdStyles.mermaid.Render(prefix + "```mermaid"),
	}
	sourceLines := extractCodeBlockLines(lines, source)
	diagram, err := renderMermaidDiagram(strings.Join(sourceLines, "\n"), width-lipgloss.Width(prefix))
	if err != nil {
		out = append(out, prefix+mdStyles.warn.Render("Mermaid not supported in this format; showing source."))
		for _, line := range sourceLines {
			out = append(out, prefix+line)
		}
		out = append(out, mdStyles.mermaid.Render(prefix+"```"))
		return out
	}
	for _, line := range diagram {
		out = append(out, prefix+line)
	}
	out = centerMermaidLines(out, prefix, width)
	out = append(out, mdStyles.mermaid.Render(prefix+"```"))
	return out
}

func centerMermaidLines(lines []string, prefix string, width int) []string {
	if len(lines) == 0 {
		return lines
	}
	contentWidth := width - lipgloss.Width(prefix)
	if contentWidth <= 0 {
		return lines
	}
	centered := make([]string, 0, len(lines))
	for i, line := range lines {
		if i == 0 {
			centered = append(centered, line)
			continue
		}
		if !strings.HasPrefix(line, prefix) {
			centered = append(centered, line)
			continue
		}
		raw := strings.TrimPrefix(line, prefix)
		lineWidth := lipgloss.Width(raw)
		if lineWidth <= 0 || lineWidth >= contentWidth {
			centered = append(centered, line)
			continue
		}
		leftPad := (contentWidth - lineWidth) / 2
		centered = append(centered, prefix+strings.Repeat(" ", leftPad)+raw)
	}
	return centered
}

func extractCodeBlockLines(lines *text.Segments, source []byte) []string {
	out := make([]string, 0, lines.Len())
	for i := 0; i < lines.Len(); i++ {
		segment := lines.At(i)
		out = append(out, strings.TrimRight(string(segment.Value(source)), "\r\n"))
	}
	return out
}

func renderMarkdownList(list *ast.List, source []byte, width int, prefix string, mdStyles markdownStyles) []string {
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
		out = append(out, renderMarkdownListItem(listItem, source, width, prefix, marker, mdStyles)...)
		index++
	}
	return out
}

func renderMarkdownListItem(item *ast.ListItem, source []byte, width int, prefix string, marker string, mdStyles markdownStyles) []string {
	itemLines := make([]string, 0, 4)
	firstPrefix := prefix + marker
	continuationPrefix := prefix + strings.Repeat(" ", len(marker))
	for child := item.FirstChild(); child != nil; child = child.NextSibling() {
		switch typed := child.(type) {
		case *ast.Paragraph:
			itemLines = append(itemLines, wrapPrefixedText(extractNodeText(typed, source), width, firstPrefix, continuationPrefix, mdStyles)...)
			firstPrefix = continuationPrefix
		case *ast.List:
			itemLines = append(itemLines, renderMarkdownList(typed, source, width, continuationPrefix, mdStyles)...)
		default:
			textValue := strings.TrimSpace(extractNodeText(child, source))
			if textValue != "" {
				itemLines = append(itemLines, wrapPrefixedText(textValue, width, firstPrefix, continuationPrefix, mdStyles)...)
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
			for codeChild := typed.FirstChild(); codeChild != nil; codeChild = codeChild.NextSibling() {
				if textNode, ok := codeChild.(*ast.Text); ok {
					buffer.Write(textNode.Segment.Value(source))
				}
			}
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

func renderInlineChildren(node ast.Node, source []byte, mdStyles markdownStyles) string {
	var buffer bytes.Buffer
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		buffer.WriteString(renderInlineNode(child, source, mdStyles))
	}
	return strings.TrimSpace(buffer.String())
}

func renderInlineNode(node ast.Node, source []byte, mdStyles markdownStyles) string {
	switch typed := node.(type) {
	case *ast.Text:
		textValue := string(typed.Segment.Value(source))
		if typed.HardLineBreak() || typed.SoftLineBreak() {
			return textValue + "\n"
		}
		return textValue
	case *ast.CodeSpan:
		var code bytes.Buffer
		for child := typed.FirstChild(); child != nil; child = child.NextSibling() {
			if textNode, ok := child.(*ast.Text); ok {
				code.Write(textNode.Segment.Value(source))
			}
		}
		return mdStyles.code.Render("`" + code.String() + "`")
	case *ast.Emphasis:
		content := renderInlineChildren(typed, source, mdStyles)
		if typed.Level == 2 {
			return mdStyles.strong.Render(content)
		}
		return mdStyles.em.Render(content)
	case *ast.Link:
		label := renderInlineChildren(typed, source, mdStyles)
		if label == "" {
			label = string(typed.Destination)
		}
		return mdStyles.link.Render(label)
	case *ast.AutoLink:
		return mdStyles.link.Render(string(typed.Label(source)))
	case *ast.RawHTML:
		var html bytes.Buffer
		for i := 0; i < typed.Segments.Len(); i++ {
			segment := typed.Segments.At(i)
			html.Write(segment.Value(source))
		}
		return html.String()
	default:
		return renderInlineChildren(typed, source, mdStyles)
	}
}

func wrapPrefixedText(input string, width int, prefix string, continuationPrefix string, mdStyles markdownStyles) []string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return nil
	}
	plainPrefixLen := lipgloss.Width(stripANSI(prefix))
	effectiveWidth := max(1, width-plainPrefixLen)
	wrappedText := strings.TrimRight(wordwrap.String(trimmed, effectiveWidth), "\n")
	wrapped := strings.Split(wrappedText, "\n")
	out := make([]string, 0, len(wrapped))
	for i, line := range wrapped {
		line = strings.TrimRight(line, " ")
		if i == 0 {
			out = append(out, mdStyles.quote.Render(prefix)+line)
			continue
		}
		out = append(out, mdStyles.quote.Render(continuationPrefix)+line)
	}
	return out
}

func stripANSI(input string) string {
	return ansiPattern.ReplaceAllString(input, "")
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
