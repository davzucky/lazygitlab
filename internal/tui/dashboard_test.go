package tui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type stubProvider struct{}

func (s stubProvider) LoadIssues(context.Context) ([]ListItem, error) {
	return []ListItem{{ID: 11, Title: "Issue one"}}, nil
}

func (s stubProvider) LoadMergeRequests(context.Context) ([]ListItem, error) {
	return []ListItem{{ID: 21, Title: "MR one"}}, nil
}

func TestDashboardViewSwitches(t *testing.T) {
	t.Parallel()

	m := NewDashboardModel(stubProvider{}, DashboardContext{})
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	model := updated.(DashboardModel)

	if model.view != IssuesView {
		t.Fatalf("view = %v want %v", model.view, IssuesView)
	}
}
