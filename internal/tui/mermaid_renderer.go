package tui

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	mermaidMinNodeWidth  = 5
	mermaidNodeHeight    = 3
	mermaidHorizontalGap = 6
	mermaidVerticalGap   = 3
)

var (
	mermaidHeaderPattern = regexp.MustCompile(`(?i)^flowchart\s+(LR|RL|TB|BT)$`)
	mermaidNodePattern   = regexp.MustCompile(`^(\w+)(?:\[(.*?)\])?$`)
)

func isMermaidIgnorableLine(line string) bool {
	lower := strings.ToLower(strings.TrimSpace(line))
	if lower == "" {
		return true
	}
	return strings.HasPrefix(lower, "%%") ||
		strings.HasPrefix(lower, "subgraph ") ||
		lower == "end" ||
		strings.HasPrefix(lower, "direction ") ||
		strings.HasPrefix(lower, "classdef ") ||
		strings.HasPrefix(lower, "class ")
}

type mermaidGraph struct {
	direction string
	nodes     map[string]*mermaidNode
	edges     []mermaidEdge
}

type mermaidNode struct {
	id     string
	label  string
	width  int
	height int
	x      int
	y      int
}

type mermaidEdge struct {
	from string
	to   string
}

type mermaidSourceLine struct {
	text   string
	lineNo int
}

func renderMermaidDiagram(input string, maxWidth int) ([]string, error) {
	graph, err := parseMermaidFlowchart(input)
	if err != nil {
		return nil, err
	}
	layoutMermaidGraph(graph)
	lines := renderMermaidGraph(graph)
	if maxWidth > 0 && widestMermaidLine(lines) > maxWidth {
		switched, ok := tryVerticalMermaidDirection(graph.direction)
		if ok {
			graph.direction = switched
			layoutMermaidGraph(graph)
			lines = renderMermaidGraph(graph)
		}
	}
	return lines, nil
}

func widestMermaidLine(lines []string) int {
	maxWidth := 0
	for _, line := range lines {
		if len([]rune(line)) > maxWidth {
			maxWidth = len([]rune(line))
		}
	}
	return maxWidth
}

func tryVerticalMermaidDirection(direction string) (string, bool) {
	switch direction {
	case "LR":
		return "TB", true
	case "RL":
		return "BT", true
	default:
		return "", false
	}
}

func parseMermaidFlowchart(input string) (*mermaidGraph, error) {
	lines := strings.Split(strings.ReplaceAll(input, "\r\n", "\n"), "\n")
	filtered := make([]mermaidSourceLine, 0, len(lines))
	for i, raw := range lines {
		line := strings.TrimSpace(raw)
		if isMermaidIgnorableLine(line) {
			continue
		}
		filtered = append(filtered, mermaidSourceLine{text: line, lineNo: i + 1})
	}
	if len(filtered) == 0 {
		return nil, fmt.Errorf("empty mermaid content")
	}

	header := mermaidHeaderPattern.FindStringSubmatch(filtered[0].text)
	if len(header) != 2 {
		return nil, fmt.Errorf("only flowchart LR/RL/TB/BT is supported")
	}

	graph := &mermaidGraph{
		direction: strings.ToUpper(header[1]),
		nodes:     make(map[string]*mermaidNode),
		edges:     make([]mermaidEdge, 0, len(filtered)),
	}

	for i := 1; i < len(filtered); i++ {
		line := filtered[i].text
		parts := strings.Split(line, "-->")
		if len(parts) > 1 {
			prev := ""
			for _, part := range parts {
				node, err := parseMermaidNode(strings.TrimSpace(part))
				if err != nil {
					return nil, fmt.Errorf("line %d: %w", filtered[i].lineNo, err)
				}
				graph.addOrUpdateNode(node)
				if prev != "" {
					graph.edges = append(graph.edges, mermaidEdge{from: prev, to: node.id})
				}
				prev = node.id
			}
			continue
		}

		node, err := parseMermaidNode(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", filtered[i].lineNo, err)
		}
		graph.addOrUpdateNode(node)
	}

	if len(graph.nodes) == 0 {
		return nil, fmt.Errorf("no nodes found in mermaid flowchart")
	}

	if hasMermaidCycle(graph) {
		return nil, fmt.Errorf("cyclic graphs are not supported yet")
	}

	return graph, nil
}

func parseMermaidNode(input string) (*mermaidNode, error) {
	match := mermaidNodePattern.FindStringSubmatch(input)
	if len(match) != 3 {
		return nil, fmt.Errorf("unsupported syntax %q", input)
	}
	id := match[1]
	label := match[2]
	if label == "" {
		label = id
	}
	return &mermaidNode{id: id, label: label}, nil
}

func (g *mermaidGraph) addOrUpdateNode(node *mermaidNode) {
	existing, ok := g.nodes[node.id]
	if !ok {
		g.nodes[node.id] = node
		return
	}
	if node.label != "" && node.label != node.id {
		existing.label = node.label
	}
}

func hasMermaidCycle(graph *mermaidGraph) bool {
	inDegree := make(map[string]int, len(graph.nodes))
	adj := make(map[string][]string, len(graph.nodes))
	for id := range graph.nodes {
		inDegree[id] = 0
	}
	for _, edge := range graph.edges {
		inDegree[edge.to]++
		adj[edge.from] = append(adj[edge.from], edge.to)
	}

	queue := make([]string, 0, len(graph.nodes))
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}
	processed := 0
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		processed++
		for _, next := range adj[id] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}
	return processed != len(graph.nodes)
}

func layoutMermaidGraph(graph *mermaidGraph) {
	for _, node := range graph.nodes {
		node.width = max(mermaidMinNodeWidth, utf8.RuneCountInString(node.label)+2)
		node.height = mermaidNodeHeight
	}
	layers := assignMermaidLayers(graph)
	assignMermaidCoordinates(graph, layers)
}

func assignMermaidLayers(graph *mermaidGraph) map[string]int {
	inDegree := make(map[string]int, len(graph.nodes))
	adj := make(map[string][]string, len(graph.nodes))
	for id := range graph.nodes {
		inDegree[id] = 0
	}
	for _, edge := range graph.edges {
		inDegree[edge.to]++
		adj[edge.from] = append(adj[edge.from], edge.to)
	}

	queue := make([]string, 0, len(graph.nodes))
	for id, degree := range inDegree {
		if degree == 0 {
			queue = insertSorted(queue, id)
		}
	}

	layer := make(map[string]int, len(graph.nodes))
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		neighbors := append([]string(nil), adj[id]...)
		sort.Strings(neighbors)
		for _, next := range neighbors {
			if layer[next] < layer[id]+1 {
				layer[next] = layer[id] + 1
			}
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = insertSorted(queue, next)
			}
		}
	}
	return layer
}

func insertSorted(items []string, value string) []string {
	index := sort.SearchStrings(items, value)
	items = append(items, "")
	copy(items[index+1:], items[index:])
	items[index] = value
	return items
}

func assignMermaidCoordinates(graph *mermaidGraph, layers map[string]int) {
	byLayer := make(map[int][]*mermaidNode)
	maxLayer := 0
	for id, node := range graph.nodes {
		value := layers[id]
		byLayer[value] = append(byLayer[value], node)
		if value > maxLayer {
			maxLayer = value
		}
	}
	for i := 0; i <= maxLayer; i++ {
		sort.Slice(byLayer[i], func(a int, b int) bool {
			return byLayer[i][a].id < byLayer[i][b].id
		})
	}

	layerWidths := make(map[int]int, maxLayer+1)
	layerHeights := make(map[int]int, maxLayer+1)
	maxLayerWidth := 0
	maxLayerHeight := 0
	for i := 0; i <= maxLayer; i++ {
		nodes := byLayer[i]
		if len(nodes) == 0 {
			continue
		}
		maxNodeWidth := 0
		totalNodeWidth := 0
		maxNodeHeight := 0
		totalNodeHeight := 0
		for _, node := range nodes {
			if node.width > maxNodeWidth {
				maxNodeWidth = node.width
			}
			totalNodeWidth += node.width
			if node.height > maxNodeHeight {
				maxNodeHeight = node.height
			}
			totalNodeHeight += node.height
		}
		if graph.direction == "LR" || graph.direction == "RL" {
			layerWidths[i] = maxNodeWidth
			layerHeights[i] = totalNodeHeight + max(0, len(nodes)-1)*mermaidVerticalGap
		} else {
			layerWidths[i] = totalNodeWidth + max(0, len(nodes)-1)*mermaidHorizontalGap
			layerHeights[i] = maxNodeHeight
		}
		if layerWidths[i] > maxLayerWidth {
			maxLayerWidth = layerWidths[i]
		}
		if layerHeights[i] > maxLayerHeight {
			maxLayerHeight = layerHeights[i]
		}
	}

	if graph.direction == "LR" || graph.direction == "RL" {
		x := 0
		for i := 0; i <= maxLayer; i++ {
			layerIndex := i
			if graph.direction == "RL" {
				layerIndex = maxLayer - i
			}
			y := max(0, (maxLayerHeight-layerHeights[layerIndex])/2)
			maxWidth := mermaidMinNodeWidth
			for _, node := range byLayer[layerIndex] {
				node.x = x
				node.y = y
				y += node.height + mermaidVerticalGap
				if node.width > maxWidth {
					maxWidth = node.width
				}
			}
			x += maxWidth + mermaidHorizontalGap
		}
		return
	}

	y := 0
	for i := 0; i <= maxLayer; i++ {
		layerIndex := i
		if graph.direction == "BT" {
			layerIndex = maxLayer - i
		}
		x := max(0, (maxLayerWidth-layerWidths[layerIndex])/2)
		maxHeight := mermaidNodeHeight
		for _, node := range byLayer[layerIndex] {
			node.x = x
			node.y = y
			x += node.width + mermaidHorizontalGap
			if node.height > maxHeight {
				maxHeight = node.height
			}
		}
		y += maxHeight + mermaidVerticalGap
	}
}

func renderMermaidGraph(graph *mermaidGraph) []string {
	maxX := 0
	maxY := 0
	for _, node := range graph.nodes {
		maxX = max(maxX, node.x+node.width)
		maxY = max(maxY, node.y+node.height)
	}
	grid := makeMermaidGrid(maxX+2, maxY+2)
	for _, node := range graph.nodes {
		drawMermaidNode(grid, node)
	}
	for _, edge := range graph.edges {
		from, okFrom := graph.nodes[edge.from]
		to, okTo := graph.nodes[edge.to]
		if okFrom && okTo {
			drawMermaidEdge(grid, from, to, graph.direction)
		}
	}
	return grid.lines()
}

type mermaidGrid struct {
	width  int
	height int
	rows   [][]rune
	locked [][]bool
}

func makeMermaidGrid(width int, height int) *mermaidGrid {
	rows := make([][]rune, height)
	locked := make([][]bool, height)
	for y := 0; y < height; y++ {
		rows[y] = make([]rune, width)
		locked[y] = make([]bool, width)
		for x := 0; x < width; x++ {
			rows[y][x] = ' '
		}
	}
	return &mermaidGrid{width: width, height: height, rows: rows, locked: locked}
}

func (g *mermaidGrid) set(x int, y int, ch rune) {
	if x < 0 || y < 0 || x >= g.width || y >= g.height {
		return
	}
	// Ignore spaces so drawing never erases existing content.
	if ch == ' ' {
		return
	}
	g.rows[y][x] = ch
}

func (g *mermaidGrid) setIfEmptyOrEdge(x int, y int, ch rune) {
	if x < 0 || y < 0 || x >= g.width || y >= g.height {
		return
	}
	if ch == ' ' {
		return
	}
	current := g.rows[y][x]
	if current == ' ' {
		g.rows[y][x] = ch
		return
	}
	if g.locked[y][x] {
		return
	}
	if isMermaidArrowRune(ch) {
		g.rows[y][x] = ch
		return
	}
	if merged, ok := mergeMermaidEdgeRunes(current, ch); ok {
		g.rows[y][x] = merged
	}
}

func isMermaidArrowRune(ch rune) bool {
	switch ch {
	case '▶', '◀', '▼', '▲':
		return true
	default:
		return false
	}
}

func mergeMermaidEdgeRunes(current rune, next rune) (rune, bool) {
	if current == next {
		return current, true
	}
	hasHorizontal := current == '─' || current == '┼'
	hasVertical := current == '│' || current == '┼'
	if next == '─' {
		hasHorizontal = true
	}
	if next == '│' {
		hasVertical = true
	}
	if hasHorizontal && hasVertical {
		return '┼', true
	}
	if hasHorizontal {
		return '─', true
	}
	if hasVertical {
		return '│', true
	}
	return current, false
}

func (g *mermaidGrid) markLocked(x int, y int) {
	if x < 0 || y < 0 || x >= g.width || y >= g.height {
		return
	}
	g.locked[y][x] = true
}

func (g *mermaidGrid) lines() []string {
	out := make([]string, 0, g.height)
	for _, row := range g.rows {
		line := strings.TrimRight(string(row), " ")
		out = append(out, line)
	}
	return trimBlankEdges(out)
}

func drawMermaidNode(grid *mermaidGrid, node *mermaidNode) {
	x := node.x
	y := node.y
	w := node.width
	h := node.height
	grid.set(x, y, '┌')
	grid.markLocked(x, y)
	grid.set(x+w-1, y, '┐')
	grid.markLocked(x+w-1, y)
	grid.set(x, y+h-1, '└')
	grid.markLocked(x, y+h-1)
	grid.set(x+w-1, y+h-1, '┘')
	grid.markLocked(x+w-1, y+h-1)
	for i := 1; i < w-1; i++ {
		grid.set(x+i, y, '─')
		grid.markLocked(x+i, y)
		grid.set(x+i, y+h-1, '─')
		grid.markLocked(x+i, y+h-1)
	}
	for i := 1; i < h-1; i++ {
		grid.set(x, y+i, '│')
		grid.markLocked(x, y+i)
		grid.set(x+w-1, y+i, '│')
		grid.markLocked(x+w-1, y+i)
	}
	labelWidth := utf8.RuneCountInString(node.label)
	labelX := x + max(1, (w-labelWidth)/2)
	for i, r := range []rune(node.label) {
		grid.set(labelX+i, y+1, r)
		grid.markLocked(labelX+i, y+1)
	}
}

func drawMermaidEdge(grid *mermaidGrid, from *mermaidNode, to *mermaidNode, direction string) {
	startX := from.x + from.width
	startY := from.y + 1
	endX := to.x - 1
	endY := to.y + 1
	arrow := '▶'

	switch direction {
	case "RL":
		startX = from.x - 1
		startY = from.y + 1
		endX = to.x + to.width
		endY = to.y + 1
		arrow = '◀'
	case "TB":
		startX = from.x + from.width/2
		startY = from.y + from.height
		endX = to.x + to.width/2
		endY = to.y - 1
		arrow = '▼'
	case "BT":
		startX = from.x + from.width/2
		startY = from.y - 1
		endX = to.x + to.width/2
		endY = to.y + to.height
		arrow = '▲'
	}

	if direction == "LR" || direction == "RL" {
		if startY == endY {
			drawMermaidHorizontal(grid, startX, endX, startY)
			grid.setIfEmptyOrEdge(endX, endY, arrow)
			return
		}
		midX := startX + max(1, (endX-startX)/2)
		drawMermaidHorizontal(grid, startX, midX, startY)
		drawMermaidVertical(grid, midX, startY, endY)
		drawMermaidHorizontal(grid, midX, endX, endY)
		grid.setIfEmptyOrEdge(endX, endY, arrow)
		return
	}

	if startX == endX {
		drawMermaidVertical(grid, startX, startY, endY)
		grid.setIfEmptyOrEdge(endX, endY, arrow)
		return
	}
	if abs(startX-endX) <= 1 {
		columnX := startX
		if endX < startX {
			columnX = endX
		}
		drawMermaidVertical(grid, columnX, startY, endY)
		grid.setIfEmptyOrEdge(columnX, endY, arrow)
		return
	}
	midY := startY + max(1, (endY-startY)/2)
	drawMermaidVertical(grid, startX, startY, midY)
	drawMermaidHorizontal(grid, startX, endX, midY)
	drawMermaidVertical(grid, endX, midY, endY)
	grid.setIfEmptyOrEdge(endX, endY, arrow)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func drawMermaidHorizontal(grid *mermaidGrid, start int, end int, y int) {
	if start > end {
		start, end = end, start
	}
	for x := start; x <= end; x++ {
		grid.setIfEmptyOrEdge(x, y, '─')
	}
}

func drawMermaidVertical(grid *mermaidGrid, x int, start int, end int) {
	if start > end {
		start, end = end, start
	}
	for y := start; y <= end; y++ {
		grid.setIfEmptyOrEdge(x, y, '│')
	}
}
