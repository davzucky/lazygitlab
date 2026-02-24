package tui

import "fmt"

import tea "github.com/charmbracelet/bubbletea"

var mergeRequestKeyHints = []string{
	"enter: open merge request details",
	"/: search",
	"tab: autocomplete",
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
	case "/":
		m = m.openSearch(MergeRequestsView)
		return m, m.loadSearchMetadataCmd(MergeRequestsView), true
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
	resultCount := fmt.Sprintf("results: %d", len(m.items))
	if m.loading {
		resultCount = "results: loading"
	}
	if m.loadingMore {
		resultCount += " (+more)"
	}
	lines := []string{
		" " + m.renderMergeRequestTabs(max(20, width-8)),
		" " + m.renderMergeRequestSearch(max(20, width-8)),
		m.styles.dim.Render(" sort: updated newest first | " + resultCount),
		m.styles.dim.Render(" keys: / search, tab complete, [ ] state, enter details, ? help"),
		"",
	}
	return lines
}
