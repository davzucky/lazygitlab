package tui

import tea "github.com/charmbracelet/bubbletea"

var mergeRequestKeyHints = []string{
	"enter: open merge request details",
	"[: prev state",
	"]: next state",
	"o/m/c/a: open/merged/closed/all",
}

func (m DashboardModel) handleMergeRequestScreenKey(key string) (tea.Model, tea.Cmd, bool) {
	if m.view != MergeRequestsView {
		return m, nil, false
	}

	switch key {
	case "enter":
		if m.hasMergeRequestDetailsSelection() {
			m.mergeRequestDetail = true
			m.mergeRequestDetailScroll = 0
			return m, nil, true
		}
	case "[":
		m.mergeRequestState = prevMergeRequestState(m.mergeRequestState)
		m.selected = 0
		model, cmd := m.startLoadCurrentView()
		return model, cmd, true
	case "]":
		m.mergeRequestState = nextMergeRequestState(m.mergeRequestState)
		m.selected = 0
		model, cmd := m.startLoadCurrentView()
		return model, cmd, true
	case "o":
		m.mergeRequestState = MergeRequestStateOpened
		m.selected = 0
		model, cmd := m.startLoadCurrentView()
		return model, cmd, true
	case "m":
		m.mergeRequestState = MergeRequestStateMerged
		m.selected = 0
		model, cmd := m.startLoadCurrentView()
		return model, cmd, true
	case "c":
		m.mergeRequestState = MergeRequestStateClosed
		m.selected = 0
		model, cmd := m.startLoadCurrentView()
		return model, cmd, true
	case "a":
		m.mergeRequestState = MergeRequestStateAll
		m.selected = 0
		model, cmd := m.startLoadCurrentView()
		return model, cmd, true
	}

	return m, nil, false
}

func (m DashboardModel) renderMergeRequestBody(width int) []string {
	lines := []string{
		" " + m.renderMergeRequestTabs(max(20, width-8)),
		m.styles.dim.Render(" sort: updated newest first"),
		"",
	}
	for _, hint := range mergeRequestKeyHints {
		lines = append(lines, m.styles.dim.Render(" "+hint))
	}
	return lines
}
