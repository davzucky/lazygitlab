package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type loadedMsg struct {
	view        ViewMode
	items       []ListItem
	err         error
	requestID   int
	replace     bool
	hasNextPage bool
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
	searchInput  textinput.Model
	searchMode   bool
	issueState   IssueState
	issueSearch  string
	issuePage    int
	issueHasNext bool
	loadingMore  bool
	requestSeq   int
	requestID    int
}

func NewDashboardModel(provider DataProvider, ctx DashboardContext) DashboardModel {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))

	search := textinput.New()
	search.Prompt = "Search: "
	search.Placeholder = "type and press Enter"
	search.CharLimit = 120
	search.Width = 30

	return DashboardModel{
		provider:    provider,
		ctx:         ctx,
		styles:      newStyles(),
		view:        IssuesView,
		width:       100,
		height:      40,
		loading:     true,
		spinner:     sp,
		searchInput: search,
		issueState:  IssueStateOpened,
	}
}

func (m DashboardModel) Init() tea.Cmd {
	m.requestSeq++
	m.requestID = m.requestSeq
	return tea.Batch(m.spinner.Tick, m.loadCurrentViewCmd(m.requestID, true, 1))
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
		if msg.view != m.view || msg.requestID != m.requestID {
			return m, nil
		}
		m.loading = false
		m.loadingMore = false
		if msg.err != nil {
			m.errorMessage = msg.err.Error()
			if msg.replace {
				m.items = nil
				m.selected = 0
			}
			return m, nil
		}
		m.errorMessage = ""
		if msg.replace {
			m.items = msg.items
		} else {
			m.items = append(m.items, msg.items...)
		}
		if m.view == IssuesView {
			m.issueHasNext = msg.hasNextPage
			if msg.replace {
				m.issuePage = 1
			} else {
				m.issuePage++
			}
		}
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

		if m.searchMode {
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			switch msg.String() {
			case "enter":
				m.searchMode = false
				m.searchInput.Blur()
				m.issueSearch = strings.TrimSpace(m.searchInput.Value())
				m.selected = 0
				return m, m.startLoadCurrentView()
			case "esc":
				m.searchMode = false
				m.searchInput.Blur()
				m.searchInput.SetValue(m.issueSearch)
				return m, nil
			}
			return m, cmd
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "j", "down":
			if m.selected < len(m.items)-1 {
				m.selected++
				if m.shouldLoadMoreIssues() {
					return m, m.startLoadMoreIssues()
				}
			}
		case "k", "up":
			if m.selected > 0 {
				m.selected--
			}
		case "/":
			if m.view == IssuesView {
				m.searchMode = true
				m.searchInput.Focus()
				m.searchInput.SetValue(m.issueSearch)
				m.searchInput.CursorEnd()
				return m, nil
			}
		case "[":
			if m.view == IssuesView {
				m.issueState = prevIssueState(m.issueState)
				m.selected = 0
				return m, m.startLoadCurrentView()
			}
		case "]":
			if m.view == IssuesView {
				m.issueState = nextIssueState(m.issueState)
				m.selected = 0
				return m, m.startLoadCurrentView()
			}
		case "o":
			if m.view == IssuesView {
				m.issueState = IssueStateOpened
				m.selected = 0
				return m, m.startLoadCurrentView()
			}
		case "c":
			if m.view == IssuesView {
				m.issueState = IssueStateClosed
				m.selected = 0
				return m, m.startLoadCurrentView()
			}
		case "a":
			if m.view == IssuesView {
				m.issueState = IssueStateAll
				m.selected = 0
				return m, m.startLoadCurrentView()
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

	lines := []string{header}
	if m.view == IssuesView {
		lines = append(lines,
			" "+m.renderIssueTabs(max(20, width-6)),
			" "+m.renderIssueSearch(max(20, width-6)),
			m.styles.dim.Render(" sort: updated newest first"),
			"",
		)
	} else {
		lines = append(lines, "")
	}
	bodyRows := max(1, height-len(lines)-2)
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
		if m.view == IssuesView && m.loadingMore {
			lines = append(lines, m.styles.dim.Render("  "+m.spinner.View()+" Loading next page..."))
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
	} else if m.view == IssuesView {
		status += fmt.Sprintf(" | issues: %s", issueStateLabel(m.issueState))
		if strings.TrimSpace(m.issueSearch) != "" {
			status += fmt.Sprintf(" | search=%q", m.issueSearch)
		}
		if m.loadingMore {
			status += " | loading more"
		}
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
  [,] or o/c/a        Issue state tabs
  /                   Search issues
  r                   Retry after error
  q                   Quit
  ?                   Toggle help
`
	return m.styles.helpPopup.Render(content)
}

func (m DashboardModel) startLoadCurrentView() tea.Cmd {
	m.loading = true
	m.loadingMore = false
	m.requestSeq++
	m.requestID = m.requestSeq
	if m.view == IssuesView {
		m.issuePage = 1
	}
	return m.loadCurrentViewCmd(m.requestID, true, m.issuePage)
}

func (m DashboardModel) startLoadMoreIssues() tea.Cmd {
	if !m.shouldLoadMoreIssues() {
		return nil
	}
	m.loadingMore = true
	m.requestSeq++
	m.requestID = m.requestSeq
	nextPage := m.issuePage + 1
	return m.loadCurrentViewCmd(m.requestID, false, nextPage)
}

func (m DashboardModel) loadCurrentViewCmd(requestID int, replace bool, issuesPage int) tea.Cmd {
	view := m.view
	provider := m.provider
	issueState := m.issueState
	issueSearch := m.issueSearch
	return func() tea.Msg {
		ctx := context.Background()
		var (
			items       []ListItem
			err         error
			hasNextPage bool
		)

		switch view {
		case IssuesView:
			result, issueErr := provider.LoadIssues(ctx, IssueQuery{
				State:   issueState,
				Search:  issueSearch,
				Page:    issuesPage,
				PerPage: 25,
			})
			err = issueErr
			items = result.Items
			hasNextPage = result.HasNextPage
		case MergeRequestsView:
			items, err = provider.LoadMergeRequests(ctx)
		}

		return loadedMsg{view: view, items: items, err: err, requestID: requestID, replace: replace, hasNextPage: hasNextPage}
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

func (m DashboardModel) shouldLoadMoreIssues() bool {
	if m.view != IssuesView || m.loading || m.loadingMore || !m.issueHasNext {
		return false
	}
	if len(m.items) == 0 {
		return false
	}
	return m.selected >= len(m.items)-2
}

func (m DashboardModel) renderIssueTabs(width int) string {
	tabs := []IssueState{IssueStateOpened, IssueStateClosed, IssueStateAll}
	parts := make([]string, 0, len(tabs))
	for _, tab := range tabs {
		label := issueStateLabel(tab)
		if tab == m.issueState {
			parts = append(parts, m.styles.selectedRow.Render("["+label+"]"))
			continue
		}
		parts = append(parts, m.styles.dim.Render(label))
	}
	return fitLine(strings.Join(parts, "  "), width)
}

func (m DashboardModel) renderIssueSearch(width int) string {
	if m.searchMode {
		return fitLine(m.searchInput.View(), width)
	}
	if strings.TrimSpace(m.issueSearch) == "" {
		return fitLine("Search: (press /)", width)
	}
	return fitLine(fmt.Sprintf("Search: %s (press / to edit)", m.issueSearch), width)
}

func issueStateLabel(state IssueState) string {
	switch state {
	case IssueStateClosed:
		return "Closed"
	case IssueStateAll:
		return "All"
	default:
		return "Open"
	}
}

func nextIssueState(current IssueState) IssueState {
	switch current {
	case IssueStateOpened:
		return IssueStateClosed
	case IssueStateClosed:
		return IssueStateAll
	default:
		return IssueStateOpened
	}
}

func prevIssueState(current IssueState) IssueState {
	switch current {
	case IssueStateOpened:
		return IssueStateAll
	case IssueStateClosed:
		return IssueStateOpened
	default:
		return IssueStateClosed
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
