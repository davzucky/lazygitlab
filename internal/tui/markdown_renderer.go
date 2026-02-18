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
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

const maxMarkdownRenderChars = 12000

var (
	markdownHeadingStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	markdownEmStyle      = lipgloss.NewStyle().Italic(true)
	markdownStrongStyle  = lipgloss.NewStyle().Bold(true)
	markdownCodeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("236"))
	markdownLinkStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Underline(true)
	markdownQuoteStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	markdownFenceStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	markdownMermaidStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
	markdownMermaidWarn  = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Italic(true)
	ansiPattern          = regexp.MustCompile(`\x1b\[[0-9;]*m`)
)

func renderMarkdownParagraphs(input string, width int) []string {
	content := strings.TrimSpace(strings.ReplaceAll(input, "\r\n", "\n"))
	if content == "" {
		return []string{""}
	}
	if utf8.RuneCountInString(content) > maxMarkdownRenderChars {
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
	md := goldmark.New(
		goldmark.WithExtensions(extension.Table),
	)
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
			headingText := markdownHeadingStyle.Render(renderInlineChildren(typed, source))
			out = append(out, wrapPrefixedText(headingText, width, prefix+headingPrefix, strings.Repeat(" ", len(prefix)+len(headingPrefix)))...)
			out = append(out, "")
		case *ast.Paragraph:
			out = append(out, wrapPrefixedText(renderInlineChildren(typed, source), width, prefix, prefix)...)
			out = append(out, "")
		case *ast.Blockquote:
			out = append(out, renderMarkdownBlocks(typed, source, width, prefix+"> ")...)
			out = append(out, "")
		case *ast.List:
			out = append(out, renderMarkdownList(typed, source, width, prefix)...)
			out = append(out, "")
		case *extast.Table:
			out = append(out, renderMarkdownTable(typed, source, width, prefix)...)
			out = append(out, "")
		case *ast.FencedCodeBlock:
			lang := strings.TrimSpace(string(typed.Language(source)))
			if strings.EqualFold(lang, "mermaid") {
				out = append(out, renderMermaidBlockLines(typed.Lines(), source, prefix, width)...)
				out = append(out, "")
				continue
			}
			if lang == "" {
				out = append(out, markdownFenceStyle.Render(prefix+"```"))
			} else {
				out = append(out, markdownFenceStyle.Render(prefix+"```"+lang))
			}
			out = append(out, renderHighlightedCodeBlockLines(typed.Lines(), source, prefix, lang)...)
			out = append(out, markdownFenceStyle.Render(prefix+"```"))
			out = append(out, "")
		case *ast.CodeBlock:
			out = append(out, markdownFenceStyle.Render(prefix+"```"))
			out = append(out, renderHighlightedCodeBlockLines(typed.Lines(), source, prefix, "")...)
			out = append(out, markdownFenceStyle.Render(prefix+"```"))
			out = append(out, "")
		case *ast.ThematicBreak:
			out = append(out, markdownFenceStyle.Render(prefix+"---"))
			out = append(out, "")
		default:
			textValue := strings.TrimSpace(renderInlineChildren(node, source))
			if textValue != "" {
				out = append(out, wrapPrefixedText(textValue, width, prefix, prefix)...)
				out = append(out, "")
			}
		}
	}
	return out
}

func renderCodeBlockLines(lines *text.Segments, source []byte, prefix string) []string {
	out := make([]string, 0, lines.Len())
	for i := 0; i < lines.Len(); i++ {
		segment := lines.At(i)
		line := strings.TrimRight(string(segment.Value(source)), "\r\n")
		out = append(out, prefix+markdownCodeStyle.Render(line))
	}
	return out
}

func renderHighlightedCodeBlockLines(lines *text.Segments, source []byte, prefix string, language string) []string {
	codeLines := make([]string, 0, lines.Len())
	for i := 0; i < lines.Len(); i++ {
		segment := lines.At(i)
		codeLines = append(codeLines, strings.TrimRight(string(segment.Value(source)), "\r\n"))
	}
	code := strings.Join(codeLines, "\n")
	if strings.TrimSpace(code) == "" {
		return renderCodeBlockLines(lines, source, prefix)
	}

	var buf bytes.Buffer
	lang := language
	if strings.TrimSpace(lang) == "" {
		lang = "plaintext"
	}
	if err := quick.Highlight(&buf, code, lang, "terminal16m", "monokai"); err != nil {
		return renderCodeBlockLines(lines, source, prefix)
	}
	highlighted := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	out := make([]string, 0, len(highlighted))
	for _, line := range highlighted {
		out = append(out, prefix+line)
	}
	return out
}

func renderMermaidBlockLines(lines *text.Segments, source []byte, prefix string, width int) []string {
	out := []string{
		markdownMermaidStyle.Render(prefix + "```mermaid"),
	}
	sourceLines := extractCodeBlockLines(lines, source)
	diagram, err := renderMermaidDiagram(strings.Join(sourceLines, "\n"), width-lipgloss.Width(prefix))
	if err != nil {
		out = append(out, prefix+markdownMermaidWarn.Render("Mermaid not supported in this format; showing source."))
		for _, line := range sourceLines {
			out = append(out, prefix+line)
		}
		out = append(out, markdownMermaidStyle.Render(prefix+"```"))
		return out
	}
	for _, line := range diagram {
		out = append(out, prefix+line)
	}
	out = centerMermaidLines(out, prefix, width)
	out = append(out, markdownMermaidStyle.Render(prefix+"```"))
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

func renderMarkdownTable(table *extast.Table, source []byte, width int, prefix string) []string {
	rows := make([][]string, 0, 4)
	for child := table.FirstChild(); child != nil; child = child.NextSibling() {
		switch typed := child.(type) {
		case *extast.TableHeader:
			rows = append(rows, renderMarkdownTableRow(typed, source))
		case *extast.TableRow:
			rows = append(rows, renderMarkdownTableRow(typed, source))
		}
	}
	if len(rows) == 0 {
		return nil
	}

	columnCount := maxTableColumns(rows, len(table.Alignments))
	if columnCount <= 0 {
		return nil
	}
	normalized := normalizeTableRows(rows, columnCount)
	columnWidths := fitTableColumnWidths(normalized, width-lipgloss.Width(prefix), columnCount)
	if len(columnWidths) == 0 {
		return nil
	}

	out := make([]string, 0, len(normalized)+1)
	out = append(out, renderTableLine(normalized[0], columnWidths, table.Alignments, prefix))
	out = append(out, renderTableSeparator(columnWidths, table.Alignments, prefix))
	for _, row := range normalized[1:] {
		out = append(out, renderTableLine(row, columnWidths, table.Alignments, prefix))
	}
	return out
}

func renderMarkdownTableRow(row ast.Node, source []byte) []string {
	columns := make([]string, 0, 4)
	for child := row.FirstChild(); child != nil; child = child.NextSibling() {
		cell, ok := child.(*extast.TableCell)
		if !ok {
			continue
		}
		value := strings.ReplaceAll(renderInlineChildren(cell, source), "\n", " ")
		columns = append(columns, strings.TrimSpace(value))
	}
	return columns
}

func maxTableColumns(rows [][]string, alignments int) int {
	maxColumns := alignments
	for _, row := range rows {
		if len(row) > maxColumns {
			maxColumns = len(row)
		}
	}
	return maxColumns
}

func normalizeTableRows(rows [][]string, columns int) [][]string {
	normalized := make([][]string, 0, len(rows))
	for _, row := range rows {
		current := make([]string, columns)
		for i := 0; i < columns; i++ {
			if i < len(row) {
				current[i] = row[i]
				continue
			}
			current[i] = ""
		}
		normalized = append(normalized, current)
	}
	return normalized
}

func fitTableColumnWidths(rows [][]string, totalWidth int, columns int) []int {
	if columns <= 0 {
		return nil
	}
	available := totalWidth - (3*columns + 1)
	if available <= 0 {
		available = columns
	}

	widths := make([]int, columns)
	for col := 0; col < columns; col++ {
		maxWidth := 1
		for _, row := range rows {
			cellWidth := lipgloss.Width(stripANSI(row[col]))
			if cellWidth > maxWidth {
				maxWidth = cellWidth
			}
		}
		widths[col] = maxWidth
	}

	for sumInt(widths) > available {
		index := widestShrinkableColumn(widths)
		if index < 0 {
			break
		}
		widths[index]--
	}
	return widths
}

func widestShrinkableColumn(widths []int) int {
	index := -1
	maxWidth := 1
	for i, width := range widths {
		if width <= 1 {
			continue
		}
		if index == -1 || width > maxWidth {
			index = i
			maxWidth = width
		}
	}
	return index
}

func sumInt(values []int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
}

func renderTableLine(row []string, widths []int, alignments []extast.Alignment, prefix string) string {
	parts := make([]string, 0, len(widths))
	for i := 0; i < len(widths); i++ {
		alignment := extast.AlignNone
		if i < len(alignments) {
			alignment = alignments[i]
		}
		cell := ""
		if i < len(row) {
			cell = row[i]
		}
		parts = append(parts, alignTableCell(cell, widths[i], alignment))
	}
	return prefix + "| " + strings.Join(parts, " | ") + " |"
}

func renderTableSeparator(widths []int, alignments []extast.Alignment, prefix string) string {
	parts := make([]string, 0, len(widths))
	for i, width := range widths {
		alignment := extast.AlignNone
		if i < len(alignments) {
			alignment = alignments[i]
		}
		parts = append(parts, tableSeparatorSegment(width, alignment))
	}
	return prefix + "| " + strings.Join(parts, " | ") + " |"
}

func alignTableCell(cell string, width int, alignment extast.Alignment) string {
	trimmed := strings.TrimSpace(cell)
	truncated := fitLine(trimmed, width)
	padding := width - lipgloss.Width(stripANSI(truncated))
	if padding <= 0 {
		return truncated
	}
	leftPad := 0
	rightPad := 0
	switch alignment {
	case extast.AlignRight:
		leftPad = padding
	case extast.AlignCenter:
		leftPad = padding / 2
		rightPad = padding - leftPad
	default:
		rightPad = padding
	}
	return strings.Repeat(" ", leftPad) + truncated + strings.Repeat(" ", rightPad)
}

func tableSeparatorSegment(width int, alignment extast.Alignment) string {
	if width <= 1 {
		if alignment == extast.AlignCenter || alignment == extast.AlignLeft || alignment == extast.AlignRight {
			return ":"
		}
		return "-"
	}
	segment := []rune(strings.Repeat("-", width))
	switch alignment {
	case extast.AlignLeft:
		segment[0] = ':'
	case extast.AlignRight:
		segment[len(segment)-1] = ':'
	case extast.AlignCenter:
		segment[0] = ':'
		segment[len(segment)-1] = ':'
	}
	return string(segment)
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
			itemLines = append(itemLines, wrapPrefixedText(extractNodeText(typed, source), width, firstPrefix, continuationPrefix)...)
			firstPrefix = continuationPrefix
		case *ast.List:
			itemLines = append(itemLines, renderMarkdownList(typed, source, width, continuationPrefix)...)
		default:
			textValue := strings.TrimSpace(extractNodeText(child, source))
			if textValue != "" {
				itemLines = append(itemLines, wrapPrefixedText(textValue, width, firstPrefix, continuationPrefix)...)
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

func renderInlineChildren(node ast.Node, source []byte) string {
	var buffer bytes.Buffer
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		buffer.WriteString(renderInlineNode(child, source))
	}
	return strings.TrimSpace(buffer.String())
}

func renderInlineNode(node ast.Node, source []byte) string {
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
		return markdownCodeStyle.Render("`" + code.String() + "`")
	case *ast.Emphasis:
		content := renderInlineChildren(typed, source)
		if typed.Level == 2 {
			return markdownStrongStyle.Render(content)
		}
		return markdownEmStyle.Render(content)
	case *ast.Link:
		label := renderInlineChildren(typed, source)
		if label == "" {
			label = string(typed.Destination)
		}
		return markdownLinkStyle.Render(label)
	case *ast.AutoLink:
		return markdownLinkStyle.Render(string(typed.Label(source)))
	case *ast.RawHTML:
		var html bytes.Buffer
		for i := 0; i < typed.Segments.Len(); i++ {
			segment := typed.Segments.At(i)
			html.Write(segment.Value(source))
		}
		return html.String()
	default:
		return renderInlineChildren(typed, source)
	}
}

func wrapPrefixedText(input string, width int, prefix string, continuationPrefix string) []string {
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
			out = append(out, markdownQuoteStyle.Render(prefix)+line)
			continue
		}
		out = append(out, markdownQuoteStyle.Render(continuationPrefix)+line)
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
