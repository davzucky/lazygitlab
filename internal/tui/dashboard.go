package tui

import (
	"context"
	"fmt"
	"hash/fnv"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
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

type issueDetailLoadedMsg struct {
	issueIID  int64
	data      IssueDetailData
	err       error
	requestID int
}

type markdownRenderedMsg struct {
	cacheKey string
	lines    []string
}

type issueDetailTab int

const (
	issueDetailTabOverview issueDetailTab = iota
	issueDetailTabActivities
	issueDetailTabComments
)

const maxMarkdownRenderChars = 12000

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
	detailTab    issueDetailTab
	detailData   map[int64]IssueDetailData
	detailCache  map[string][]string
	markdownBody map[string][]string
	detailLoad   bool
	detailErr    string
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
		provider:     provider,
		ctx:          ctx,
		styles:       newStyles(),
		view:         IssuesView,
		width:        100,
		height:       40,
		loading:      true,
		spinner:      sp,
		searchInput:  search,
		issueState:   IssueStateOpened,
		detailData:   make(map[int64]IssueDetailData),
		detailCache:  make(map[string][]string),
		markdownBody: make(map[string][]string),
		requestSeq:   1,
		requestID:    1,
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
		m.clearDetailCache()
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
			m.detailTab = issueDetailTabOverview
			m.detailLoad = false
			m.detailErr = ""
		}
		return m, nil

	case issueDetailLoadedMsg:
		if msg.requestID != m.requestID || !m.issueDetail {
			return m, nil
		}
		item, ok := m.selectedIssueItem()
		if !ok || item.Issue == nil || item.Issue.IID != msg.issueIID {
			return m, nil
		}
		m.detailLoad = false
		if msg.err != nil {
			m.detailErr = msg.err.Error()
			return m, nil
		}
		m.detailErr = ""
		m.detailData[msg.issueIID] = msg.data
		m.invalidateDetailCacheForIssue(msg.issueIID)
		return m, m.preloadMarkdownCmd()

	case markdownRenderedMsg:
		if msg.cacheKey == "" || len(msg.lines) == 0 {
			return m, nil
		}
		m.markdownBody[msg.cacheKey] = msg.lines
		m.clearDetailCache()
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
				m.detailTab = issueDetailTabOverview
				m.detailLoad = false
				m.detailErr = ""
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
			case "r":
				item, ok := m.selectedIssueItem()
				if ok && item.Issue != nil {
					delete(m.detailData, item.Issue.IID)
					m.invalidateMarkdownCacheForIssue(item.Issue.IID)
					m.clearDetailCache()
				}
				cmd := m.loadIssueDetailDataCmd()
				if cmd != nil {
					m.detailLoad = true
					m.detailErr = ""
				}
				return m, cmd
			case "tab", "l", "right":
				m.detailTab = nextIssueDetailTab(m.detailTab)
				m.detailScroll = 0
				cmd := m.loadIssueDetailDataCmd()
				if cmd != nil {
					m.detailLoad = true
					m.detailErr = ""
				}
				return m, tea.Batch(cmd, m.preloadMarkdownCmd())
			case "shift+tab", "h", "left":
				m.detailTab = prevIssueDetailTab(m.detailTab)
				m.detailScroll = 0
				cmd := m.loadIssueDetailDataCmd()
				if cmd != nil {
					m.detailLoad = true
					m.detailErr = ""
				}
				return m, tea.Batch(cmd, m.preloadMarkdownCmd())
			case "d":
				m.detailTab = issueDetailTabOverview
				m.detailScroll = 0
				return m, m.preloadMarkdownCmd()
			case "a":
				m.detailTab = issueDetailTabActivities
				m.detailScroll = 0
				cmd := m.loadIssueDetailDataCmd()
				if cmd != nil {
					m.detailLoad = true
					m.detailErr = ""
				}
				return m, tea.Batch(cmd, m.preloadMarkdownCmd())
			case "c":
				m.detailTab = issueDetailTabComments
				m.detailScroll = 0
				cmd := m.loadIssueDetailDataCmd()
				if cmd != nil {
					m.detailLoad = true
					m.detailErr = ""
				}
				return m, tea.Batch(cmd, m.preloadMarkdownCmd())
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
				m.detailTab = issueDetailTabOverview
				m.detailErr = ""
				cmd := m.loadIssueDetailDataCmd()
				if cmd != nil {
					m.detailLoad = true
					m.detailErr = ""
				}
				return m, tea.Batch(cmd, m.preloadMarkdownCmd())
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
	contentHeight := max(8, m.height-5)
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
			" "+m.renderIssueTabs(max(20, width-8)),
			" "+m.renderIssueSearch(max(20, width-8)),
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
		contentWidth := max(10, width-8)
		rowWidth := max(1, contentWidth-2)
		rowsPerItem := 1
		if m.view == IssuesView {
			rowWidth = minInt(rowWidth, 52)
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
			line := prefix + fitLine(item.Title, rowWidth)
			lines = append(lines, rowStyle.Render(line))
			if m.view == IssuesView {
				meta := "  " + fitLine(issueListMeta(item), rowWidth)
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
	viewportWidth := max(8, contentWidth-2)
	lines := []string{
		m.styles.header.Render("Issue Detail"),
		m.styles.dim.Render("Esc return | j/k scroll | tab shift+tab or d/a/c tabs"),
		m.renderIssueDetailTabs(contentWidth),
		"",
	}
	detailLines := m.issueDetailLines(viewportWidth)
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
	lines = append(lines, withVerticalScroll(detailLines[start:end], viewportWidth, start, bodyRows, len(detailLines))...)
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
  tab/shift+tab       Cycle issue detail tabs
  d/a/c               Jump Detail/Activities/Comments
  [,] or o/c/a        Issue state tabs
  /                   Search issues
  r                   Retry load (errors)
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
	m.detailTab = issueDetailTabOverview
	m.detailLoad = false
	m.detailErr = ""
	m.clearDetailCache()
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

func (m DashboardModel) selectedIssueItem() (ListItem, bool) {
	if !m.hasIssueDetailsSelection() {
		return ListItem{}, false
	}
	return m.items[m.selected], true
}

func (m DashboardModel) clearDetailCache() {
	for key := range m.detailCache {
		delete(m.detailCache, key)
	}
}

func (m DashboardModel) invalidateDetailCacheForIssue(issueIID int64) {
	prefix := fmt.Sprintf("%d:", issueIID)
	for key := range m.detailCache {
		if strings.HasPrefix(key, prefix) {
			delete(m.detailCache, key)
		}
	}
}

func (m DashboardModel) invalidateMarkdownCacheForIssue(issueIID int64) {
	prefix := fmt.Sprintf("%d:", issueIID)
	for key := range m.markdownBody {
		if strings.HasPrefix(key, prefix) {
			delete(m.markdownBody, key)
		}
	}
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
	item, ok := m.selectedIssueItem()
	if !ok {
		return nil
	}
	details := item.Issue
	if details == nil {
		return nil
	}

	cacheKey := fmt.Sprintf("%d:%d:%d", details.IID, m.detailTab, width)
	if cached, found := m.detailCache[cacheKey]; found {
		return cached
	}

	if m.detailTab != issueDetailTabOverview {
		if m.detailLoad {
			return wrapLines([]string{"Loading issue detail data..."}, width)
		}
		if m.detailErr != "" {
			return wrapLines([]string{
				fmt.Sprintf("Failed to load issue detail data: %s", m.detailErr),
				"Press r to retry.",
			}, width)
		}
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

	metadata := []string{
		fmt.Sprintf("Title: %s", fallbackValue(item.Title, "-")),
		fmt.Sprintf("IID: %s", iid),
		fmt.Sprintf("State: %s", state),
		fmt.Sprintf("Author: %s", author),
		fmt.Sprintf("Assignees: %s", assignees),
		fmt.Sprintf("Labels: %s", labels),
		fmt.Sprintf("Created: %s", createdAt),
		fmt.Sprintf("Updated: %s", updatedAt),
		fmt.Sprintf("URL: %s", url),
	}

	switch m.detailTab {
	case issueDetailTabActivities:
		computed := m.issueActivityLines(width, details.IID)
		m.detailCache[cacheKey] = computed
		return computed
	case issueDetailTabComments:
		computed := m.issueCommentLines(width, details.IID)
		m.detailCache[cacheKey] = computed
		return computed
	}

	lines := append([]string{"Info:"}, metadata...)
	lines = append(lines, "", "Description:")

	description := strings.TrimSpace(details.Description)
	if description == "" {
		lines = append(lines, "No description provided.")
		return wrapLines(lines, width)
	}

	wrappedMeta := wrapLines(lines, width)
	wrappedDescription := m.markdownOrWrapped(details.IID, "description", 0, description, width)
	computed := append(wrappedMeta, wrappedDescription...)
	m.detailCache[cacheKey] = computed
	return computed
}

func (m DashboardModel) issueCommentLines(width int, issueIID int64) []string {
	data, ok := m.detailData[issueIID]
	if !ok || len(data.Comments) == 0 {
		return wrapLines([]string{"No comments available."}, width)
	}

	lines := make([]string, 0, len(data.Comments)*6)
	for i, comment := range data.Comments {
		if i > 0 {
			lines = append(lines, "")
		}
		header := fmt.Sprintf("%s • %s", fallbackValue(comment.Author, "-"), fallbackValue(comment.CreatedAt, "-"))
		lines = append(lines, wrapLine(header, width)...)
		body := strings.TrimSpace(comment.Body)
		if body == "" {
			lines = append(lines, m.styles.dim.Render("(empty comment)"))
			continue
		}
		lines = append(lines, m.markdownOrWrapped(issueIID, "comment", i, body, width)...)
	}
	return lines
}

func (m DashboardModel) markdownOrWrapped(issueIID int64, section string, index int, content string, width int) []string {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return []string{""}
	}
	key := markdownCacheKey(issueIID, section, index, width, trimmed)
	if rendered, ok := m.markdownBody[key]; ok {
		return rendered
	}
	return wrapParagraphs(trimmed, width)
}

func (m DashboardModel) issueActivityLines(width int, issueIID int64) []string {
	data, ok := m.detailData[issueIID]
	if !ok || len(data.Activities) == 0 {
		return wrapLines([]string{"No activities available."}, width)
	}

	lines := make([]string, 0, len(data.Activities))
	for _, activity := range data.Activities {
		line := fmt.Sprintf("%s • %s • %s", fallbackValue(activity.CreatedAt, "-"), fallbackValue(activity.Actor, "-"), fallbackValue(activity.Action, "-"))
		lines = append(lines, line)
	}
	return wrapLines(lines, width)
}

func (m DashboardModel) renderIssueDetailTabs(_ int) string {
	tabs := []issueDetailTab{issueDetailTabOverview, issueDetailTabActivities, issueDetailTabComments}
	parts := make([]string, 0, len(tabs))
	for _, tab := range tabs {
		label := issueDetailTabLabel(tab)
		runes := []rune(label)
		if len(runes) > 0 {
			mnemonic := string(runes[0])
			rest := string(runes[1:])
			if tab == m.detailTab {
				letter := m.styles.selectedRow.Underline(true).Render(mnemonic)
				parts = append(parts, m.styles.selectedRow.Render(letter+rest))
				continue
			}
			letter := m.styles.dim.Underline(true).Render(mnemonic)
			parts = append(parts, m.styles.dim.Render(letter+rest))
			continue
		}
		if tab == m.detailTab {
			parts = append(parts, m.styles.selectedRow.Render(label))
			continue
		}
		parts = append(parts, m.styles.dim.Render(label))
	}
	return strings.Join(parts, "  ")
}

func withVerticalScroll(lines []string, width int, start int, rows int, total int) []string {
	out := make([]string, 0, len(lines))
	thumbStart, thumbEnd := scrollbarThumb(rows, total, start)
	for i, line := range lines {
		rail := "|"
		if i >= thumbStart && i < thumbEnd {
			rail = "#"
		}
		out = append(out, padToWidth(line, width)+" "+rail)
	}
	for len(out) < rows {
		idx := len(out)
		rail := "|"
		if idx >= thumbStart && idx < thumbEnd {
			rail = "#"
		}
		out = append(out, strings.Repeat(" ", width)+" "+rail)
	}
	return out
}

func scrollbarThumb(rows int, total int, start int) (int, int) {
	if rows <= 0 {
		return 0, 0
	}
	if total <= rows {
		return 0, rows
	}
	thumbSize := max(1, int(math.Round(float64(rows*rows)/float64(total))))
	if thumbSize > rows {
		thumbSize = rows
	}
	maxStart := rows - thumbSize
	scrollRange := total - rows
	if scrollRange <= 0 || maxStart <= 0 {
		return 0, thumbSize
	}
	ratio := float64(start) / float64(scrollRange)
	thumbStart := int(math.Round(ratio * float64(maxStart)))
	if thumbStart < 0 {
		thumbStart = 0
	}
	if thumbStart > maxStart {
		thumbStart = maxStart
	}
	return thumbStart, thumbStart + thumbSize
}

func padToWidth(input string, width int) string {
	if width <= 0 {
		return ""
	}
	visible := lipgloss.Width(input)
	if visible >= width {
		return input
	}
	return input + strings.Repeat(" ", width-visible)
}

func (m DashboardModel) preloadMarkdownCmd() tea.Cmd {
	if !m.issueDetail {
		return nil
	}
	item, ok := m.selectedIssueItem()
	if !ok || item.Issue == nil {
		return nil
	}
	issueIID := item.Issue.IID
	width, _ := m.issueDetailViewport()
	cmds := make([]tea.Cmd, 0, 8)

	if m.detailTab == issueDetailTabOverview {
		description := strings.TrimSpace(item.Issue.Description)
		if description != "" {
			if cmd := m.markdownCmdForContent(issueIID, "description", 0, description, width); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}

	if m.detailTab == issueDetailTabComments {
		if data, ok := m.detailData[issueIID]; ok {
			for i, comment := range data.Comments {
				body := strings.TrimSpace(comment.Body)
				if body == "" {
					continue
				}
				if cmd := m.markdownCmdForContent(issueIID, "comment", i, body, width); cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}
	}

	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

func (m DashboardModel) markdownCmdForContent(issueIID int64, section string, index int, content string, width int) tea.Cmd {
	key := markdownCacheKey(issueIID, section, index, width, content)
	if _, ok := m.markdownBody[key]; ok {
		return nil
	}
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return nil
	}
	return func() tea.Msg {
		lines := renderMarkdownParagraphs(trimmed, width)
		return markdownRenderedMsg{cacheKey: key, lines: lines}
	}
}

func markdownCacheKey(issueIID int64, section string, index int, width int, content string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(content))
	return fmt.Sprintf("%d:%s:%d:%d:%x", issueIID, section, index, width, h.Sum64())
}

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

func (m DashboardModel) loadIssueDetailDataCmd() tea.Cmd {
	item, ok := m.selectedIssueItem()
	if !ok || item.Issue == nil || item.Issue.IID <= 0 {
		return nil
	}
	if _, exists := m.detailData[item.Issue.IID]; exists {
		return nil
	}
	if m.detailLoad {
		return nil
	}
	requestID := m.requestID
	issueIID := item.Issue.IID
	provider := m.provider
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		data, err := provider.LoadIssueDetailData(ctx, issueIID)
		return issueDetailLoadedMsg{issueIID: issueIID, data: data, err: err, requestID: requestID}
	}
}

func issueDetailTabLabel(tab issueDetailTab) string {
	switch tab {
	case issueDetailTabActivities:
		return "Activities"
	case issueDetailTabComments:
		return "Comments"
	default:
		return "Detail"
	}
}

func nextIssueDetailTab(tab issueDetailTab) issueDetailTab {
	switch tab {
	case issueDetailTabOverview:
		return issueDetailTabActivities
	case issueDetailTabActivities:
		return issueDetailTabComments
	default:
		return issueDetailTabOverview
	}
}

func prevIssueDetailTab(tab issueDetailTab) issueDetailTab {
	switch tab {
	case issueDetailTabOverview:
		return issueDetailTabComments
	case issueDetailTabActivities:
		return issueDetailTabOverview
	default:
		return issueDetailTabActivities
	}
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
	contentHeight := max(8, m.height-5)
	contentWidth := max(10, totalWidth-m.styles.panel.GetHorizontalFrameSize()-2)
	bodyRows := max(1, contentHeight-m.styles.panel.GetVerticalFrameSize()-4)
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
