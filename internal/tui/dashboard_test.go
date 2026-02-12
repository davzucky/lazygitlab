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
		return IssueResult{Items: []ListItem{{ID: 12, Title: "Issue two", Issue: &IssueDetails{IID: 102, State: "opened", Description: "second issue"}}}, HasNextPage: false}, nil
	}
	return IssueResult{Items: []ListItem{{ID: 11, Title: "Issue one", Issue: &IssueDetails{IID: 101, State: "opened", Description: "first issue"}}}, HasNextPage: true}, nil
}

func (s *stubProvider) LoadMergeRequests(context.Context) ([]ListItem, error) {
	return []ListItem{{ID: 21, Title: "MR one"}}, nil
}

func (s *stubProvider) LoadIssueDetailData(context.Context, int64) (IssueDetailData, error) {
	return IssueDetailData{
		Activities: []IssueActivity{{Actor: "alice", CreatedAt: "2026-01-02 10:00 UTC", Action: "closed"}},
		Comments:   []IssueComment{{Author: "bob", CreatedAt: "2026-01-02 10:05 UTC", Body: "**hello**"}},
	}, nil
}

func TestDashboardViewSwitches(t *testing.T) {
	t.Parallel()

	m := NewDashboardModel(&stubProvider{}, DashboardContext{})
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	model := updated.(DashboardModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(DashboardModel)

	if model.view != IssuesView {
		t.Fatalf("view = %v want %v", model.view, IssuesView)
	}
}

func TestDashboardStartsOnPrimaryView(t *testing.T) {
	t.Parallel()

	m := NewDashboardModel(&stubProvider{}, DashboardContext{})
	if m.view != PrimaryView {
		t.Fatalf("view = %v want %v", m.view, PrimaryView)
	}
}

func TestDashboardEscReturnsToPrimaryFromIssues(t *testing.T) {
	t.Parallel()

	m := NewDashboardModel(&stubProvider{}, DashboardContext{})
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	model := updated.(DashboardModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(DashboardModel)
	if model.view != IssuesView {
		t.Fatalf("view = %v want %v", model.view, IssuesView)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(DashboardModel)
	if model.view != PrimaryView {
		t.Fatalf("view = %v want %v", model.view, PrimaryView)
	}
}

func TestDashboardPrimaryRoutesToMergeRequests(t *testing.T) {
	t.Parallel()

	m := NewDashboardModel(&stubProvider{}, DashboardContext{})
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")})
	model := updated.(DashboardModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(DashboardModel)

	if model.view != MergeRequestsView {
		t.Fatalf("view = %v want %v", model.view, MergeRequestsView)
	}
}

func TestDashboardEscReturnsToPrimaryFromMergeRequests(t *testing.T) {
	t.Parallel()

	m := NewDashboardModel(&stubProvider{}, DashboardContext{})
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")})
	model := updated.(DashboardModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(DashboardModel)
	if model.view != MergeRequestsView {
		t.Fatalf("view = %v want %v", model.view, MergeRequestsView)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(DashboardModel)
	if model.view != PrimaryView {
		t.Fatalf("view = %v want %v", model.view, PrimaryView)
	}
}

func TestDashboardLeftFromIssuesReturnsPrimary(t *testing.T) {
	t.Parallel()

	m := NewDashboardModel(&stubProvider{}, DashboardContext{})
	m.view = IssuesView
	m.loading = false

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	model := updated.(DashboardModel)
	if model.view != PrimaryView {
		t.Fatalf("view = %v want %v", model.view, PrimaryView)
	}
}

func TestDashboardTabCyclesThroughPrimaryIssueMergeRequest(t *testing.T) {
	t.Parallel()

	m := NewDashboardModel(&stubProvider{}, DashboardContext{})

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	model := updated.(DashboardModel)
	if model.view != IssuesView {
		t.Fatalf("view after first tab = %v want %v", model.view, IssuesView)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(DashboardModel)
	if model.view != MergeRequestsView {
		t.Fatalf("view after second tab = %v want %v", model.view, MergeRequestsView)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(DashboardModel)
	if model.view != PrimaryView {
		t.Fatalf("view after third tab = %v want %v", model.view, PrimaryView)
	}
}

func TestDashboardShiftTabCyclesBackFromPrimaryToMergeRequest(t *testing.T) {
	t.Parallel()

	m := NewDashboardModel(&stubProvider{}, DashboardContext{})

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	model := updated.(DashboardModel)
	if model.view != MergeRequestsView {
		t.Fatalf("view = %v want %v", model.view, MergeRequestsView)
	}
}

func TestDashboardPrimaryThreeSelectsMergeRequestWithoutRouting(t *testing.T) {
	t.Parallel()

	m := NewDashboardModel(&stubProvider{}, DashboardContext{})

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("3")})
	model := updated.(DashboardModel)
	if model.view != PrimaryView {
		t.Fatalf("view = %v want %v", model.view, PrimaryView)
	}
	if model.primaryIndex != 1 {
		t.Fatalf("primaryIndex = %d want %d", model.primaryIndex, 1)
	}
}

func TestDashboardIssueStateTabReloads(t *testing.T) {
	t.Parallel()

	provider := &stubProvider{}
	m := NewDashboardModel(provider, DashboardContext{})
	m.view = IssuesView
	m.loading = false
	m.items = []ListItem{{ID: 11, Title: "Issue one", Issue: &IssueDetails{IID: 101, State: "opened", Description: "first issue"}}}

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

func TestDashboardIssueAllTabReloads(t *testing.T) {
	t.Parallel()

	provider := &stubProvider{}
	m := NewDashboardModel(provider, DashboardContext{})
	m.view = IssuesView
	m.loading = false
	m.issueState = IssueStateClosed
	m.items = []ListItem{{ID: 11, Title: "Issue one", Issue: &IssueDetails{IID: 101, State: "closed", Description: "first issue"}}}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	model := updated.(DashboardModel)
	if model.issueState != IssueStateAll {
		t.Fatalf("state = %v want %v", model.issueState, IssueStateAll)
	}
	if cmd == nil {
		t.Fatal("expected reload command")
	}

	_ = cmd()
	if len(provider.issueCalls) == 0 {
		t.Fatal("expected issue load call")
	}
	if provider.issueCalls[0].State != IssueStateAll {
		t.Fatalf("call state = %v want %v", provider.issueCalls[0].State, IssueStateAll)
	}
}

func TestDashboardLoadsNextIssuePageNearEnd(t *testing.T) {
	t.Parallel()

	provider := &stubProvider{}
	m := NewDashboardModel(provider, DashboardContext{})
	m.view = IssuesView
	m.loading = false
	m.items = []ListItem{
		{ID: 11, Title: "Issue one", Issue: &IssueDetails{IID: 101, State: "opened", Description: "first issue"}},
		{ID: 13, Title: "Issue three", Issue: &IssueDetails{IID: 103, State: "opened", Description: "third issue"}},
	}
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
	m.view = IssuesView
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

func TestDashboardInitialRequestIDAcceptsFirstLoad(t *testing.T) {
	t.Parallel()

	m := NewDashboardModel(&stubProvider{}, DashboardContext{})
	m.view = IssuesView
	updated, _ := m.Update(loadedMsg{view: IssuesView, items: []ListItem{{ID: 99, Title: "Loaded", Issue: &IssueDetails{IID: 199, State: "opened", Description: "loaded issue"}}}, requestID: 1, replace: true})
	model := updated.(DashboardModel)
	if len(model.items) == 0 {
		t.Fatal("expected first load message to be accepted")
	}
}

func TestDashboardIssueDetailOpensAndCloses(t *testing.T) {
	t.Parallel()

	m := NewDashboardModel(&stubProvider{}, DashboardContext{})
	m.view = IssuesView
	m.loading = false
	m.items = []ListItem{{ID: 11, Title: "Issue one", Issue: &IssueDetails{IID: 101, State: "opened", Description: "first issue"}}}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(DashboardModel)
	if !model.issueDetail {
		t.Fatal("expected issue detail view to open")
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(DashboardModel)
	if model.issueDetail {
		t.Fatal("expected issue detail view to close")
	}
}

func TestDashboardIssueDetailEnterDoesNotClose(t *testing.T) {
	t.Parallel()

	m := NewDashboardModel(&stubProvider{}, DashboardContext{})
	m.view = IssuesView
	m.loading = false
	m.items = []ListItem{{ID: 11, Title: "Issue one", Issue: &IssueDetails{IID: 101, State: "opened", Description: "first issue"}}}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(DashboardModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(DashboardModel)

	if !model.issueDetail {
		t.Fatal("expected issue detail view to remain open on Enter")
	}
}

func TestDashboardIssueDetailScrollDoesNotMoveSelection(t *testing.T) {
	t.Parallel()

	m := NewDashboardModel(&stubProvider{}, DashboardContext{})
	m.view = IssuesView
	m.loading = false
	m.width = 100
	m.height = 20
	m.items = []ListItem{{
		ID:    11,
		Title: "Issue one",
		Issue: &IssueDetails{
			IID:         101,
			State:       "opened",
			Description: "line one line two line three line four line five line six line seven line eight line nine line ten",
		},
	}}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(DashboardModel)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	model = updated.(DashboardModel)

	if model.selected != 0 {
		t.Fatalf("selected = %d want %d", model.selected, 0)
	}
	if model.detailScroll == 0 {
		t.Fatal("expected detail scroll to advance")
	}
}

func TestDashboardIssueDetailTabSwitches(t *testing.T) {
	t.Parallel()

	m := NewDashboardModel(&stubProvider{}, DashboardContext{})
	m.view = IssuesView
	m.loading = false
	m.items = []ListItem{{ID: 11, Title: "Issue one", Issue: &IssueDetails{IID: 101, State: "opened", Description: "first issue"}}}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(DashboardModel)
	if cmd == nil {
		t.Fatal("expected detail data load command")
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model = updated.(DashboardModel)
	if model.detailTab != issueDetailTabActivities {
		t.Fatalf("detail tab = %v want %v", model.detailTab, issueDetailTabActivities)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	model = updated.(DashboardModel)
	if model.detailTab != issueDetailTabOverview {
		t.Fatalf("detail tab = %v want %v", model.detailTab, issueDetailTabOverview)
	}
}

func TestDashboardIssueDetailMnemonicTabKeys(t *testing.T) {
	t.Parallel()

	m := NewDashboardModel(&stubProvider{}, DashboardContext{})
	m.view = IssuesView
	m.loading = false
	m.items = []ListItem{{ID: 11, Title: "Issue one", Issue: &IssueDetails{IID: 101, State: "opened", Description: "first issue"}}}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(DashboardModel)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")})
	model = updated.(DashboardModel)
	if model.detailTab != issueDetailTabActivities {
		t.Fatalf("detail tab = %v want %v", model.detailTab, issueDetailTabActivities)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	model = updated.(DashboardModel)
	if model.detailTab != issueDetailTabComments {
		t.Fatalf("detail tab = %v want %v", model.detailTab, issueDetailTabComments)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("d")})
	model = updated.(DashboardModel)
	if model.detailTab != issueDetailTabOverview {
		t.Fatalf("detail tab = %v want %v", model.detailTab, issueDetailTabOverview)
	}
}
