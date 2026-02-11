package tui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type issueCall struct {
	State  IssueState
	Search string
	Page   int
}

type stubProvider struct {
	issueCalls []issueCall
}

func (s *stubProvider) LoadIssues(_ context.Context, query IssueQuery) (IssueResult, error) {
	s.issueCalls = append(s.issueCalls, issueCall{State: query.State, Search: query.Search, Page: query.Page})
	if query.Page == 2 {
		return IssueResult{Items: []ListItem{{ID: 12, Title: "Issue two"}}, HasNextPage: false}, nil
	}
	return IssueResult{Items: []ListItem{{ID: 11, Title: "Issue one"}}, HasNextPage: true}, nil
}

func (s *stubProvider) LoadMergeRequests(context.Context) ([]ListItem, error) {
	return []ListItem{{ID: 21, Title: "MR one"}}, nil
}

func TestDashboardViewSwitches(t *testing.T) {
	t.Parallel()

	m := NewDashboardModel(&stubProvider{}, DashboardContext{})
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	model := updated.(DashboardModel)

	if model.view != IssuesView {
		t.Fatalf("view = %v want %v", model.view, IssuesView)
	}
}

func TestDashboardIssueStateTabReloads(t *testing.T) {
	t.Parallel()

	provider := &stubProvider{}
	m := NewDashboardModel(provider, DashboardContext{})
	m.loading = false
	m.items = []ListItem{{ID: 11, Title: "Issue one"}}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("]")})
	model := updated.(DashboardModel)
	if model.issueState != IssueStateClosed {
		t.Fatalf("state = %v want %v", model.issueState, IssueStateClosed)
	}
	if cmd == nil {
		t.Fatal("expected reload command")
	}

	_ = cmd()
	if len(provider.issueCalls) == 0 {
		t.Fatal("expected issue load call")
	}
	if provider.issueCalls[0].State != IssueStateClosed {
		t.Fatalf("call state = %v want %v", provider.issueCalls[0].State, IssueStateClosed)
	}
}

func TestDashboardLoadsNextIssuePageNearEnd(t *testing.T) {
	t.Parallel()

	provider := &stubProvider{}
	m := NewDashboardModel(provider, DashboardContext{})
	m.loading = false
	m.items = []ListItem{{ID: 11, Title: "Issue one"}, {ID: 13, Title: "Issue three"}}
	m.selected = 0
	m.issuePage = 1
	m.issueHasNext = true

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	model := updated.(DashboardModel)
	if model.selected != 1 {
		t.Fatalf("selected = %d want %d", model.selected, 1)
	}
	if cmd == nil {
		t.Fatal("expected next-page load command")
	}

	_ = cmd()
	if len(provider.issueCalls) == 0 {
		t.Fatal("expected issue load call")
	}
	if provider.issueCalls[0].Page != 2 {
		t.Fatalf("call page = %d want %d", provider.issueCalls[0].Page, 2)
	}
}

func TestDashboardIssueSearchAppliesOnEnter(t *testing.T) {
	t.Parallel()

	provider := &stubProvider{}
	m := NewDashboardModel(provider, DashboardContext{})
	m.loading = false

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")})
	model := updated.(DashboardModel)
	if !model.searchMode {
		t.Fatal("expected search mode")
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("bug")})
	model = updated.(DashboardModel)
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(DashboardModel)
	if model.issueSearch != "bug" {
		t.Fatalf("search = %q want %q", model.issueSearch, "bug")
	}
	if cmd == nil {
		t.Fatal("expected search reload command")
	}

	_ = cmd()
	if len(provider.issueCalls) == 0 {
		t.Fatal("expected issue load call")
	}
	if provider.issueCalls[0].Search != "bug" {
		t.Fatalf("call search = %q want %q", provider.issueCalls[0].Search, "bug")
	}
}
