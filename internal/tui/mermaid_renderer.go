package tui

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
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

func renderMermaidDiagram(input string) ([]string, error) {
	graph, err := parseMermaidFlowchart(input)
	if err != nil {
		return nil, err
	}
	layoutMermaidGraph(graph)
	return renderMermaidGraph(graph), nil
}

func parseMermaidFlowchart(input string) (*mermaidGraph, error) {
	lines := strings.Split(strings.ReplaceAll(input, "\r\n", "\n"), "\n")
	filtered := make([]string, 0, len(lines))
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "%%") {
			continue
		}
		filtered = append(filtered, line)
	}
	if len(filtered) == 0 {
		return nil, fmt.Errorf("empty mermaid content")
	}

	header := mermaidHeaderPattern.FindStringSubmatch(filtered[0])
	if len(header) != 2 {
		return nil, fmt.Errorf("only flowchart LR/RL/TB/BT is supported")
	}

	graph := &mermaidGraph{
		direction: strings.ToUpper(header[1]),
		nodes:     make(map[string]*mermaidNode),
		edges:     make([]mermaidEdge, 0, len(filtered)),
	}

	for i := 1; i < len(filtered); i++ {
		line := filtered[i]
		parts := strings.Split(line, "-->")
		if len(parts) > 1 {
			prev := ""
			for _, part := range parts {
				node, err := parseMermaidNode(strings.TrimSpace(part))
				if err != nil {
					return nil, fmt.Errorf("line %d: %w", i+1, err)
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
			return nil, fmt.Errorf("line %d: %w", i+1, err)
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
		node.width = max(mermaidMinNodeWidth, len(node.label)+2)
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
			queue = append(queue, id)
		}
	}
	sort.Strings(queue)

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
				queue = append(queue, next)
				sort.Strings(queue)
			}
		}
	}
	return layer
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

	if graph.direction == "LR" || graph.direction == "RL" {
		x := 0
		for i := 0; i <= maxLayer; i++ {
			layerIndex := i
			if graph.direction == "RL" {
				layerIndex = maxLayer - i
			}
			y := 0
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
		x := 0
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
}

func makeMermaidGrid(width int, height int) *mermaidGrid {
	rows := make([][]rune, height)
	for y := 0; y < height; y++ {
		rows[y] = make([]rune, width)
		for x := 0; x < width; x++ {
			rows[y][x] = ' '
		}
	}
	return &mermaidGrid{width: width, height: height, rows: rows}
}

func (g *mermaidGrid) set(x int, y int, ch rune) {
	if x < 0 || y < 0 || x >= g.width || y >= g.height {
		return
	}
	if ch == ' ' {
		return
	}
	g.rows[y][x] = ch
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
	grid.set(x+w-1, y, '┐')
	grid.set(x, y+h-1, '└')
	grid.set(x+w-1, y+h-1, '┘')
	for i := 1; i < w-1; i++ {
		grid.set(x+i, y, '─')
		grid.set(x+i, y+h-1, '─')
	}
	for i := 1; i < h-1; i++ {
		grid.set(x, y+i, '│')
		grid.set(x+w-1, y+i, '│')
	}
	labelX := x + max(1, (w-len(node.label))/2)
	for i, r := range []rune(node.label) {
		grid.set(labelX+i, y+1, r)
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
		midX := startX + (endX-startX)/2
		drawMermaidHorizontal(grid, startX, midX, startY)
		drawMermaidVertical(grid, midX, startY, endY)
		drawMermaidHorizontal(grid, midX, endX, endY)
		grid.set(endX, endY, arrow)
		return
	}

	midY := startY + (endY-startY)/2
	drawMermaidVertical(grid, startX, startY, midY)
	drawMermaidHorizontal(grid, startX, endX, midY)
	drawMermaidVertical(grid, endX, midY, endY)
	grid.set(endX, endY, arrow)
}

func drawMermaidHorizontal(grid *mermaidGrid, start int, end int, y int) {
	if start > end {
		start, end = end, start
	}
	for x := start; x <= end; x++ {
		grid.set(x, y, '─')
	}
}

func drawMermaidVertical(grid *mermaidGrid, x int, start int, end int) {
	if start > end {
		start, end = end, start
	}
	for y := start; y <= end; y++ {
		grid.set(x, y, '│')
	}
}
