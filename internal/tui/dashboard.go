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
	issueDetail  bool
	detailScroll int
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
		requestSeq:  1,
		requestID:   1,
	}
}

func (m DashboardModel) Init() tea.Cmd {
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
		if !m.hasIssueDetailsSelection() {
			m.issueDetail = false
			m.detailScroll = 0
		}
		return m, nil

	case tea.KeyMsg:
		if m.errorMessage != "" {
			switch msg.String() {
			case "r":
				m.errorMessage = ""
				return m.startLoadCurrentView()
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
				return m.startLoadCurrentView()
			case "esc":
				m.searchMode = false
				m.searchInput.Blur()
				m.searchInput.SetValue(m.issueSearch)
				return m, nil
			}
			return m, cmd
		}

		if m.issueDetail {
			switch msg.String() {
			case "esc":
				m.issueDetail = false
				m.detailScroll = 0
				return m, nil
			case "q", "ctrl+c":
				return m, tea.Quit
			case "j", "down":
				m.detailScroll = m.clampDetailScroll(m.detailScroll + 1)
				return m, nil
			case "k", "up":
				m.detailScroll = m.clampDetailScroll(m.detailScroll - 1)
				return m, nil
			case "pgdown":
				m.detailScroll = m.clampDetailScroll(m.detailScroll + 8)
				return m, nil
			case "pgup":
				m.detailScroll = m.clampDetailScroll(m.detailScroll - 8)
				return m, nil
			case "?":
				m.showHelp = true
				return m, nil
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.view == IssuesView && m.hasIssueDetailsSelection() {
				m.issueDetail = true
				m.detailScroll = 0
				return m, nil
			}
		case "j", "down":
			if m.selected < len(m.items)-1 {
				m.selected++
				if m.shouldLoadMoreIssues() {
					return m.startLoadMoreIssues()
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
				return m.startLoadCurrentView()
			}
		case "]":
			if m.view == IssuesView {
				m.issueState = nextIssueState(m.issueState)
				m.selected = 0
				return m.startLoadCurrentView()
			}
		case "o":
			if m.view == IssuesView {
				m.issueState = IssueStateOpened
				m.selected = 0
				return m.startLoadCurrentView()
			}
		case "c":
			if m.view == IssuesView {
				m.issueState = IssueStateClosed
				m.selected = 0
				return m.startLoadCurrentView()
			}
		case "a":
			if m.view == IssuesView {
				m.issueState = IssueStateAll
				m.selected = 0
				return m.startLoadCurrentView()
			}
		case "h", "left":
			if m.view > IssuesView {
				m.view--
				m.selected = 0
				return m.startLoadCurrentView()
			}
		case "l", "right":
			if m.view < MergeRequestsView {
				m.view++
				m.selected = 0
				return m.startLoadCurrentView()
			}
		case "tab":
			if m.view < MergeRequestsView {
				m.view++
			} else {
				m.view = IssuesView
			}
			m.selected = 0
			return m.startLoadCurrentView()
		case "shift+tab":
			if m.view > IssuesView {
				m.view--
			} else {
				m.view = MergeRequestsView
			}
			m.selected = 0
			return m.startLoadCurrentView()
		case "1":
			m.view = IssuesView
			m.selected = 0
			return m.startLoadCurrentView()
		case "2":
			m.view = MergeRequestsView
			m.selected = 0
			return m.startLoadCurrentView()
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
	contentHeight := max(8, m.height-3)
	status := m.renderStatusBar(totalWidth)

	if m.issueDetail {
		detail := m.renderIssueDetailFullscreen(totalWidth, contentHeight)
		return m.styles.app.Render(lipgloss.JoinVertical(lipgloss.Left, detail, status))
	}

	navWidth := minInt(28, max(22, totalWidth/4))
	mainWidth := max(36, totalWidth-navWidth)
	sidebar := m.renderSidebar(navWidth, contentHeight)
	main := m.renderMain(mainWidth, contentHeight)
	content := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, main)
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

	return renderSizedBox(m.styles.panel, width, height, strings.Join(items, "\n"))
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
			m.styles.dim.Render(" enter: open issue details"),
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
		rowsPerItem := 1
		if m.view == IssuesView {
			rowsPerItem = 2
		}
		visibleItems := max(1, bodyRows/rowsPerItem)
		start, end := visibleRange(len(m.items), m.selected, visibleItems)
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
			if m.view == IssuesView {
				meta := "  " + fitLine(issueListMeta(item), contentWidth)
				lines = append(lines, m.styles.dim.Render(meta))
			}
		}

		if len(m.items) > visibleItems {
			footer := fmt.Sprintf("  %d-%d of %d", start+1, end, len(m.items))
			lines = append(lines, m.styles.dim.Render(fitLine(footer, contentWidth)))
		}
		if m.view == IssuesView && m.loadingMore {
			lines = append(lines, m.styles.dim.Render("  "+m.spinner.View()+" Loading next page..."))
		}
	}

	innerHeight := max(1, height-m.styles.panel.GetVerticalFrameSize())
	lines = fitHeight(lines, innerHeight)
	return renderSizedBox(m.styles.panel, width, height, strings.Join(lines, "\n"))
}

func (m DashboardModel) renderIssueDetailFullscreen(width int, height int) string {
	contentWidth := max(10, width-6)
	lines := []string{m.styles.header.Render("Issue Detail"), m.styles.dim.Render("Esc to return • j/k to scroll"), ""}
	detailLines := m.issueDetailLines(contentWidth)
	if len(detailLines) == 0 {
		lines = append(lines, m.styles.dim.Render("No issue details available"))
		innerHeight := max(1, height-m.styles.panel.GetVerticalFrameSize())
		lines = fitHeight(lines, innerHeight)
		return renderSizedBox(m.styles.panel, width, height, strings.Join(lines, "\n"))
	}

	bodyRows := max(1, height-len(lines)-2)
	maxScroll := max(0, len(detailLines)-bodyRows)
	start := m.detailScroll
	if start > maxScroll {
		start = maxScroll
	}
	if start < 0 {
		start = 0
	}
	end := minInt(len(detailLines), start+bodyRows)
	lines = append(lines, detailLines[start:end]...)
	if len(detailLines) > bodyRows {
		footer := fmt.Sprintf("%d-%d of %d", start+1, end, len(detailLines))
		lines = append(lines, m.styles.dim.Render(fitLine(footer, contentWidth)))
	}

	innerHeight := max(1, height-m.styles.panel.GetVerticalFrameSize())
	lines = fitHeight(lines, innerHeight)
	return renderSizedBox(m.styles.panel, width, height, strings.Join(lines, "\n"))
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
	innerWidth := max(1, width-m.styles.status.GetHorizontalFrameSize())
	return m.styles.status.Width(innerWidth).Render(fitLine(status, innerWidth))
}

func (m DashboardModel) renderHelp() string {
	content := `Keybindings

Navigation:
  j/k or up/down      Move in list
  h/l or left/right   Switch view
  tab/shift+tab       Cycle views

Actions:
  1,2                 Jump to view
  enter               Open issue detail panel
  esc                 Close issue detail panel
  [,] or o/c/a        Issue state tabs
  /                   Search issues
  r                   Retry after error
  q                   Quit
  ?                   Toggle help
`
	return m.styles.helpPopup.Render(content)
}

func (m DashboardModel) startLoadCurrentView() (tea.Model, tea.Cmd) {
	m.loading = true
	m.loadingMore = false
	m.issueDetail = false
	m.detailScroll = 0
	m.requestSeq++
	m.requestID = m.requestSeq
	if m.view == IssuesView {
		m.issuePage = 1
	}
	return m, m.loadCurrentViewCmd(m.requestID, true, m.issuePage)
}

func (m DashboardModel) startLoadMoreIssues() (tea.Model, tea.Cmd) {
	if !m.shouldLoadMoreIssues() {
		return m, nil
	}
	m.loadingMore = true
	m.requestSeq++
	m.requestID = m.requestSeq
	nextPage := m.issuePage + 1
	return m, m.loadCurrentViewCmd(m.requestID, false, nextPage)
}

func (m DashboardModel) hasIssueDetailsSelection() bool {
	if m.view != IssuesView {
		return false
	}
	if len(m.items) == 0 || m.selected < 0 || m.selected >= len(m.items) {
		return false
	}
	return m.items[m.selected].Issue != nil
}

func (m DashboardModel) clampDetailScroll(next int) int {
	contentWidth, bodyRows := m.issueDetailViewport()
	lines := m.issueDetailLines(contentWidth)
	maxScroll := max(0, len(lines)-bodyRows)
	if next < 0 {
		return 0
	}
	if next > maxScroll {
		return maxScroll
	}
	return next
}

func (m DashboardModel) issueDetailLines(width int) []string {
	if !m.hasIssueDetailsSelection() {
		return nil
	}
	item := m.items[m.selected]
	details := item.Issue
	if details == nil {
		return nil
	}

	author := fallbackValue(details.Author, "-")
	assignees := joinOrFallback(details.Assignees, "Unassigned")
	labels := joinOrFallback(details.Labels, "None")
	createdAt := fallbackValue(details.CreatedAt, "-")
	updatedAt := fallbackValue(details.UpdatedAt, "-")
	url := fallbackValue(details.URL, "-")
	state := fallbackValue(details.State, "-")
	iid := "-"
	if details.IID > 0 {
		iid = fmt.Sprintf("#%d", details.IID)
	}

	lines := []string{
		fmt.Sprintf("Title: %s", fallbackValue(item.Title, "-")),
		fmt.Sprintf("IID: %s", iid),
		fmt.Sprintf("State: %s", state),
		fmt.Sprintf("Author: %s", author),
		fmt.Sprintf("Assignees: %s", assignees),
		fmt.Sprintf("Labels: %s", labels),
		fmt.Sprintf("Created: %s", createdAt),
		fmt.Sprintf("Updated: %s", updatedAt),
		fmt.Sprintf("URL: %s", url),
		"",
		"Description:",
	}

	description := strings.TrimSpace(details.Description)
	if description == "" {
		lines = append(lines, "No description provided.")
		return wrapLines(lines, width)
	}

	wrappedMeta := wrapLines(lines, width)
	wrappedDescription := wrapParagraphs(description, width)
	return append(wrappedMeta, wrappedDescription...)
}

func wrapLines(lines []string, width int) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		parts := wrapLine(line, width)
		if len(parts) == 0 {
			out = append(out, "")
			continue
		}
		out = append(out, parts...)
	}
	return out
}

func wrapParagraphs(input string, width int) []string {
	paragraphs := strings.Split(strings.ReplaceAll(input, "\r\n", "\n"), "\n")
	out := make([]string, 0, len(paragraphs))
	for _, paragraph := range paragraphs {
		trimmed := strings.TrimSpace(paragraph)
		if trimmed == "" {
			out = append(out, "")
			continue
		}
		out = append(out, wrapLine(trimmed, width)...)
	}
	return out
}

func wrapLine(input string, width int) []string {
	if width <= 1 {
		return []string{fitLine(input, max(1, width))}
	}
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return []string{""}
	}

	words := strings.Fields(trimmed)
	if len(words) == 0 {
		return []string{""}
	}

	lines := make([]string, 0, len(words)/2+1)
	current := words[0]
	for _, word := range words[1:] {
		candidate := current + " " + word
		if len([]rune(candidate)) <= width {
			current = candidate
			continue
		}
		lines = append(lines, current)
		current = word
	}
	lines = append(lines, current)

	for i := range lines {
		lines[i] = fitLine(lines[i], width)
	}
	return lines
}

func joinOrFallback(values []string, fallback string) string {
	filtered := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		filtered = append(filtered, trimmed)
	}
	if len(filtered) == 0 {
		return fallback
	}
	return strings.Join(filtered, ", ")
}

func fallbackValue(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func (m DashboardModel) issueDetailViewport() (int, int) {
	totalWidth := max(60, m.width-2)
	contentHeight := max(8, m.height-3)
	contentWidth := max(10, totalWidth-m.styles.panel.GetHorizontalFrameSize()-2)
	bodyRows := max(1, contentHeight-m.styles.panel.GetVerticalFrameSize()-3)
	return contentWidth, bodyRows
}

func renderSizedBox(style lipgloss.Style, width int, height int, content string) string {
	innerWidth := max(1, width-style.GetHorizontalFrameSize())
	innerHeight := max(1, height-style.GetVerticalFrameSize())
	return style.Width(innerWidth).Height(innerHeight).Render(content)
}

func issueListMeta(item ListItem) string {
	const (
		authorColWidth   = 22
		assigneeColWidth = 22
	)

	if item.Issue == nil {
		if strings.TrimSpace(item.Subtitle) != "" {
			return item.Subtitle
		}
		return fmt.Sprintf(
			"by %s | to %s | created %s",
			padOrTrimRight("-", authorColWidth),
			padOrTrimRight("Unassigned", assigneeColWidth),
			"-",
		)
	}
	author := padOrTrimRight(fallbackValue(item.Issue.Author, "-"), authorColWidth)
	assignee := padOrTrimRight(joinOrFallback(item.Issue.Assignees, "Unassigned"), assigneeColWidth)
	created := fallbackValue(item.Issue.CreatedAt, "-")
	return fmt.Sprintf(
		"by %s | to %s | created %s",
		author,
		assignee,
		created,
	)
}

func padOrTrimRight(input string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(input))
	if len(runes) >= width {
		return string(runes[:width])
	}
	return string(runes) + strings.Repeat(" ", width-len(runes))
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
