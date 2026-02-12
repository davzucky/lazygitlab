package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/davzucky/lazygitlab/internal/tui"
)

type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (p *MockProvider) LoadProjects(context.Context) ([]tui.ListItem, error) {
	return []tui.ListItem{
		{ID: 101, Title: "mock/group/project", Subtitle: "Mock project for CI validation", URL: "https://mock.gitlab.local/mock/group/project"},
		{ID: 102, Title: "mock/group/another-project", Subtitle: "Secondary mock project", URL: "https://mock.gitlab.local/mock/group/another-project"},
	}, nil
}

func (p *MockProvider) LoadIssues(_ context.Context, query tui.IssueQuery) (tui.IssueResult, error) {
	state := query.State
	if state == "" {
		state = tui.IssueStateOpened
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PerPage <= 0 {
		query.PerPage = 25
	}

	filtered := make([]tui.ListItem, 0, 120)
	needle := strings.ToLower(strings.TrimSpace(query.Search))

	for i := 120; i >= 1; i-- {
		issueState := "opened"
		if i%3 == 0 {
			issueState = "closed"
		}
		if state == tui.IssueStateOpened && issueState != "opened" {
			continue
		}
		if state == tui.IssueStateClosed && issueState != "closed" {
			continue
		}

		title := fmt.Sprintf("Mock issue %03d with long title to validate clipping and stable panel width behavior", i)
		if needle != "" && !strings.Contains(strings.ToLower(title), needle) {
			continue
		}

		filtered = append(filtered, tui.ListItem{
			ID:       int64(2000 + i),
			Title:    title,
			Subtitle: fmt.Sprintf("#%d • %s", 3000+i, issueState),
			URL:      fmt.Sprintf("https://mock.gitlab.local/mock/group/project/-/issues/%d", 3000+i),
			Issue: &tui.IssueDetails{
				IID:         int64(3000 + i),
				State:       issueState,
				Author:      "Mock Author",
				Assignees:   []string{"Mock Assignee"},
				Labels:      []string{"mock", "ui"},
				CreatedAt:   "2026-01-01 10:00 UTC",
				UpdatedAt:   "2026-01-02 11:00 UTC",
				URL:         fmt.Sprintf("https://mock.gitlab.local/mock/group/project/-/issues/%d", 3000+i),
				Description: "Mock issue description for validating wrapped and scrollable issue detail rendering in the dashboard.",
			},
		})
	}

	start := (query.Page - 1) * query.PerPage
	if start >= len(filtered) {
		return tui.IssueResult{Items: []tui.ListItem{}, HasNextPage: false}, nil
	}
	end := start + query.PerPage
	if end > len(filtered) {
		end = len(filtered)
	}

	return tui.IssueResult{Items: filtered[start:end], HasNextPage: end < len(filtered)}, nil
}

func (p *MockProvider) LoadMergeRequests(context.Context) ([]tui.ListItem, error) {
	items := make([]tui.ListItem, 0, 20)
	for i := 20; i >= 1; i-- {
		items = append(items, tui.ListItem{
			ID:       int64(5000 + i),
			Title:    fmt.Sprintf("Mock merge request %02d with extended title for rendering checks", i),
			Subtitle: fmt.Sprintf("!%d • opened", 6000+i),
			URL:      fmt.Sprintf("https://mock.gitlab.local/mock/group/project/-/merge_requests/%d", 6000+i),
		})
	}

	return items, nil
}
