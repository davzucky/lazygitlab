package tui

import tea "github.com/charmbracelet/bubbletea"

func (m DashboardModel) handleIssueScreenKey(key string) (tea.Model, tea.Cmd, bool) {
	if m.view != IssuesView {
		return m, nil, false
	}

	switch key {
	case "enter":
		if m.hasIssueDetailsSelection() {
			m.issueDetail = true
			m.detailScroll = 0
			m.detailTab = issueDetailTabOverview
			m.detailErr = ""
			cmd := m.loadIssueDetailDataCmd()
			if cmd != nil {
				m.detailLoad = true
				m.detailErr = ""
			}
			return m, tea.Batch(cmd, m.preloadMarkdownCmd()), true
		}
	case "/":
		m.searchMode = true
		m.searchInput.Focus()
		m.searchInput.SetValue(m.issueSearch)
		m.searchInput.CursorEnd()
		return m, nil, true
	case "[":
		m.issueState = prevIssueState(m.issueState)
		m.selected = 0
		model, cmd := m.startLoadCurrentView()
		return model, cmd, true
	case "]":
		m.issueState = nextIssueState(m.issueState)
		m.selected = 0
		model, cmd := m.startLoadCurrentView()
		return model, cmd, true
	case "o":
		m.issueState = IssueStateOpened
		m.selected = 0
		model, cmd := m.startLoadCurrentView()
		return model, cmd, true
	case "c":
		m.issueState = IssueStateClosed
		m.selected = 0
		model, cmd := m.startLoadCurrentView()
		return model, cmd, true
	case "a":
		m.issueState = IssueStateAll
		m.selected = 0
		model, cmd := m.startLoadCurrentView()
		return model, cmd, true
	}

	return m, nil, false
}

func (m DashboardModel) renderIssueBody(width int) []string {
	lines := []string{
		" " + m.renderIssueTabs(max(20, width-8)),
		" " + m.renderIssueSearch(max(20, width-8)),
		m.styles.dim.Render(" enter: open issue details"),
		m.styles.dim.Render(" sort: updated newest first"),
		"",
	}
	return lines
}
