package tui

import tea "github.com/charmbracelet/bubbletea"

var primaryKeyHints = []string{
	"j/k or arrows: move",
	"enter: open selected screen",
	"1/2/3: quick screen select",
}

func (m DashboardModel) handlePrimaryScreenKey(key string) (tea.Model, tea.Cmd, bool) {
	if m.view != PrimaryView {
		return m, nil, false
	}

	switch key {
	case "j", "down":
		if m.primaryIndex < 1 {
			m.primaryIndex++
		}
		return m, nil, true
	case "k", "up":
		if m.primaryIndex > 0 {
			m.primaryIndex--
		}
		return m, nil, true
	case "enter":
		if m.primaryIndex == 0 {
			m.view = IssuesView
		} else {
			m.view = MergeRequestsView
		}
		m.selected = 0
		model, cmd := m.startLoadCurrentView()
		return model, cmd, true
	case "1":
		m.primaryIndex = 0
		return m, nil, true
	case "2":
		m.primaryIndex = 0
		return m, nil, true
	case "3":
		m.primaryIndex = 1
		return m, nil, true
	}

	return m, nil, false
}

func (m DashboardModel) renderPrimaryBody(width int) []string {
	lines := []string{" " + m.styles.dim.Render("Choose a screen"), ""}
	entries := []struct {
		label string
		hint  string
	}{
		{label: "Issues", hint: "List, filter, search, and inspect issue details"},
		{label: "Merge Requests", hint: "Browse merge requests"},
	}
	for i, entry := range entries {
		prefix := "  "
		rowStyle := m.styles.normalRow
		if i == m.primaryIndex {
			prefix = "› "
			rowStyle = m.styles.selectedRow
		}
		lines = append(lines, rowStyle.Render(prefix+fitLine(entry.label, max(10, width-8))))
		lines = append(lines, m.styles.dim.Render("  "+fitLine(entry.hint, max(10, width-8))))
		lines = append(lines, "")
	}
	for _, hint := range primaryKeyHints {
		lines = append(lines, m.styles.dim.Render(" "+hint))
	}
	return lines
}
