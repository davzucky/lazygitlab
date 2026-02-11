package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type loadedMsg struct {
	view  ViewMode
	items []ListItem
	err   error
}

type DashboardModel struct {
	provider     DataProvider
	ctx          DashboardContext
	styles       styles
	view         ViewMode
	items        []ListItem
	selected     int
	width        int
	height       int
	loading      bool
	errorMessage string
	showHelp     bool
	spinner      spinner.Model
}

func NewDashboardModel(provider DataProvider, ctx DashboardContext) DashboardModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))

	return DashboardModel{
		provider: provider,
		ctx:      ctx,
		styles:   newStyles(),
		view:     IssuesView,
		width:    100,
		height:   40,
		loading:  true,
		spinner:  sp,
	}
}

func (m DashboardModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.loadCurrentViewCmd())
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case loadedMsg:
		if msg.view != m.view {
			return m, nil
		}
		m.loading = false
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			m.items = nil
			m.selected = 0
			return m, nil
		}
		m.errorMessage = ""
		m.items = msg.items
		if m.selected >= len(m.items) {
			m.selected = 0
		}
		return m, nil

	case tea.KeyMsg:
		if m.errorMessage != "" {
			switch msg.String() {
			case "r":
				m.errorMessage = ""
				return m, m.startLoadCurrentView()
			case "esc", "q":
				m.errorMessage = ""
				return m, nil
			}
			return m, nil
		}

		if m.showHelp {
			switch msg.String() {
			case "?", "esc", "q":
				m.showHelp = false
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "j", "down":
			if m.selected < len(m.items)-1 {
				m.selected++
			}
		case "k", "up":
			if m.selected > 0 {
				m.selected--
			}
		case "h", "left":
			if m.view > IssuesView {
				m.view--
				m.selected = 0
				return m, m.startLoadCurrentView()
			}
		case "l", "right":
			if m.view < MergeRequestsView {
				m.view++
				m.selected = 0
				return m, m.startLoadCurrentView()
			}
		case "tab":
			if m.view < MergeRequestsView {
				m.view++
			} else {
				m.view = IssuesView
			}
			m.selected = 0
			return m, m.startLoadCurrentView()
		case "shift+tab":
			if m.view > IssuesView {
				m.view--
			} else {
				m.view = MergeRequestsView
			}
			m.selected = 0
			return m, m.startLoadCurrentView()
		case "1":
			m.view = IssuesView
			m.selected = 0
			return m, m.startLoadCurrentView()
		case "2":
			m.view = MergeRequestsView
			m.selected = 0
			return m, m.startLoadCurrentView()
		case "?":
			m.showHelp = true
		}
	}

	return m, nil
}

func (m DashboardModel) View() string {
	if m.errorMessage != "" {
		return m.styles.errorPopup.Render(fmt.Sprintf("Error\n\n%s\n\nPress r to retry\nPress q or Esc to close", m.errorMessage))
	}

	if m.showHelp {
		return m.renderHelp()
	}

	totalWidth := max(60, m.width-2)
	navWidth := minInt(28, max(22, totalWidth/4))
	mainWidth := max(36, totalWidth-navWidth)
	mainHeight := max(10, (m.height*2)/3)
	detailsHeight := max(6, m.height-mainHeight-4)

	sidebar := m.renderSidebar(navWidth, mainHeight+detailsHeight+2)
	main := m.renderMain(mainWidth, mainHeight)
	details := m.renderDetails(mainWidth, detailsHeight)
	status := m.renderStatusBar(totalWidth)

	content := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, lipgloss.JoinVertical(lipgloss.Left, main, details))
	return m.styles.app.Render(lipgloss.JoinVertical(lipgloss.Left, content, status))
}

func (m DashboardModel) renderSidebar(width int, height int) string {
	items := []string{
		m.styles.title.Render("Navigation"),
		"",
		m.navLabel(IssuesView, fitLine("1. Issues", width-6)),
		m.navLabel(MergeRequestsView, fitLine("2. Merge Requests", width-6)),
		"",
		m.styles.dim.Render("j/k or arrows to move"),
		m.styles.dim.Render("h/l tab to switch"),
		m.styles.dim.Render("q quit, ? help"),
	}

	return m.styles.panel.Width(width).Height(height).Render(strings.Join(items, "\n"))
}

func (m DashboardModel) navLabel(view ViewMode, label string) string {
	if m.view == view {
		return m.styles.sidebarActive.Render("› " + label)
	}
	return m.styles.sidebar.Render("  " + label)
}

func (m DashboardModel) renderMain(width int, height int) string {
	header := m.styles.header.Render(m.viewTitle())

	lines := []string{header, ""}
	bodyRows := max(1, height-4)
	if m.loading {
		lines = append(lines, "  "+m.spinner.View()+" Loading...")
	} else if len(m.items) == 0 {
		lines = append(lines, "  No items")
	} else {
		contentWidth := max(10, width-6)
		start, end := visibleRange(len(m.items), m.selected, bodyRows)
		for i := start; i < end; i++ {
			item := m.items[i]
			prefix := "  "
			rowStyle := m.styles.normalRow
			if i == m.selected {
				prefix = "› "
				rowStyle = m.styles.selectedRow
			}
			line := prefix + fitLine(item.Title, contentWidth)
			lines = append(lines, rowStyle.Render(line))
		}

		if len(m.items) > bodyRows {
			footer := fmt.Sprintf("  %d-%d of %d", start+1, end, len(m.items))
			lines = append(lines, m.styles.dim.Render(fitLine(footer, contentWidth)))
		}
	}

	lines = fitHeight(lines, height-2)
	return m.styles.panel.Width(width).Height(height).Render(strings.Join(lines, "\n"))
}

func (m DashboardModel) renderDetails(width int, height int) string {
	lines := []string{m.styles.header.Render("Details"), ""}

	if len(m.items) == 0 || m.selected >= len(m.items) {
		lines = append(lines, m.styles.dim.Render("No selection"))
	} else {
		contentWidth := max(10, width-8)
		item := m.items[m.selected]
		lines = append(lines,
			fmt.Sprintf("ID: %d", item.ID),
			fmt.Sprintf("Title: %s", fitLine(item.Title, contentWidth)),
		)
		if item.Subtitle != "" {
			lines = append(lines, fmt.Sprintf("Info: %s", fitLine(item.Subtitle, contentWidth)))
		}
		if item.URL != "" {
			lines = append(lines, fmt.Sprintf("URL: %s", fitLine(item.URL, contentWidth)))
		}
	}

	return m.styles.panel.Width(width).Height(height).Render(strings.Join(lines, "\n"))
}

func (m DashboardModel) renderStatusBar(width int) string {
	status := fmt.Sprintf("Project: %s | Host: %s | %s", m.ctx.ProjectPath, m.ctx.Host, m.ctx.Connection)
	if m.loading {
		status += " | loading"
	}
	return m.styles.status.Width(width).Render(fitLine(status, max(10, width-2)))
}

func (m DashboardModel) renderHelp() string {
	content := `Keybindings

Navigation:
  j/k or up/down      Move in list
  h/l or left/right   Switch view
  tab/shift+tab       Cycle views

Actions:
  1,2                 Jump to view
  r                   Retry after error
  q                   Quit
  ?                   Toggle help
`
	return m.styles.helpPopup.Render(content)
}

func (m DashboardModel) startLoadCurrentView() tea.Cmd {
	m.loading = true
	return m.loadCurrentViewCmd()
}

func (m DashboardModel) loadCurrentViewCmd() tea.Cmd {
	view := m.view
	provider := m.provider
	return func() tea.Msg {
		ctx := context.Background()
		var (
			items []ListItem
			err   error
		)

		switch view {
		case IssuesView:
			items, err = provider.LoadIssues(ctx)
		case MergeRequestsView:
			items, err = provider.LoadMergeRequests(ctx)
		}

		return loadedMsg{view: view, items: items, err: err}
	}
}

func (m DashboardModel) viewTitle() string {
	switch m.view {
	case IssuesView:
		return "Issues"
	case MergeRequestsView:
		return "Merge Requests"
	default:
		return "Unknown"
	}
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func fitLine(input string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(input)
	if len(runes) <= width {
		return input
	}
	if width <= 1 {
		return string(runes[:width])
	}
	return string(runes[:width-1]) + "…"
}

func visibleRange(total int, selected int, capacity int) (int, int) {
	if total <= 0 || capacity <= 0 {
		return 0, 0
	}
	if total <= capacity {
		return 0, total
	}

	if selected < 0 {
		selected = 0
	}
	if selected >= total {
		selected = total - 1
	}

	half := capacity / 2
	start := selected - half
	if start < 0 {
		start = 0
	}
	end := start + capacity
	if end > total {
		end = total
		start = end - capacity
	}

	return start, end
}

func fitHeight(lines []string, maxLines int) []string {
	if maxLines <= 0 {
		return []string{}
	}
	if len(lines) >= maxLines {
		return lines[:maxLines]
	}

	out := make([]string, 0, maxLines)
	out = append(out, lines...)
	for len(out) < maxLines {
		out = append(out, "")
	}

	return out
}
