package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/davzucky/lazygitlab/internal/gitlab"
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
	parsed := gitlab.ParseSearchQuery(query.Search)
	needle := strings.ToLower(strings.TrimSpace(parsed.Text))

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
		authorLogin := "mock-author"
		assigneeLogin := "mock-assignee"
		if needle != "" && !strings.Contains(strings.ToLower(title), needle) {
			continue
		}
		if parsed.Author != "" && parsed.Author != authorLogin {
			continue
		}
		if parsed.Assignee != "" && parsed.Assignee != assigneeLogin {
			continue
		}
		if parsed.Milestone != "" && parsed.Milestone != "Iteration 1" {
			continue
		}
		if len(parsed.Labels) > 0 && !containsAll([]string{"mock", "ui"}, parsed.Labels) {
			continue
		}

		filtered = append(filtered, tui.ListItem{
			ID:       int64(2000 + i),
			Title:    title,
			Subtitle: fmt.Sprintf("#%d • %s", 3000+i, issueState),
			URL:      fmt.Sprintf("https://mock.gitlab.local/mock/group/project/-/issues/%d", 3000+i),
			Issue: &tui.IssueDetails{
				IID:            int64(3000 + i),
				State:          issueState,
				Author:         "Mock Author",
				AuthorLogin:    authorLogin,
				Assignees:      []string{"Mock Assignee"},
				AssigneeLogins: []string{assigneeLogin},
				Labels:         []string{"mock", "ui"},
				Milestone:      "Iteration 1",
				CreatedAt:      "2026-01-01 10:00 UTC",
				UpdatedAt:      "2026-01-02 11:00 UTC",
				URL:            fmt.Sprintf("https://mock.gitlab.local/mock/group/project/-/issues/%d", 3000+i),
				Description:    "Mock issue description for validating wrapped and scrollable issue detail rendering in the dashboard.",
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

func (p *MockProvider) LoadMergeRequests(_ context.Context, query tui.MergeRequestQuery) (tui.MergeRequestResult, error) {
	state := query.State
	if state == "" {
		state = tui.MergeRequestStateOpened
	}
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PerPage <= 0 {
		query.PerPage = 25
	}
	parsed := gitlab.ParseSearchQuery(query.Search)
	needle := strings.ToLower(strings.TrimSpace(parsed.Text))

	items := make([]tui.ListItem, 0, 20)
	for i := 20; i >= 1; i-- {
		mrState := "opened"
		switch {
		case i%5 == 0:
			mrState = "merged"
		case i%4 == 0:
			mrState = "closed"
		}
		if state == tui.MergeRequestStateOpened && mrState != "opened" {
			continue
		}
		if state == tui.MergeRequestStateMerged && mrState != "merged" {
			continue
		}
		if state == tui.MergeRequestStateClosed && mrState != "closed" {
			continue
		}

		authorLogin := "mock-author"
		assigneeLogin := "mock-assignee"
		title := fmt.Sprintf("Mock merge request %02d with extended title for rendering checks", i)
		if needle != "" && !strings.Contains(strings.ToLower(title), needle) {
			continue
		}
		if parsed.Author != "" && parsed.Author != authorLogin {
			continue
		}
		if parsed.Assignee != "" && parsed.Assignee != assigneeLogin {
			continue
		}
		if parsed.Milestone != "" && parsed.Milestone != "Iteration 1" {
			continue
		}
		if len(parsed.Labels) > 0 && !containsAll([]string{"mock", "ui"}, parsed.Labels) {
			continue
		}

		iid := int64(6000 + i)
		url := fmt.Sprintf("https://mock.gitlab.local/mock/group/project/-/merge_requests/%d", iid)
		items = append(items, tui.ListItem{
			ID:       int64(5000 + i),
			Title:    title,
			Subtitle: fmt.Sprintf("!%d • %s", iid, mrState),
			URL:      url,
			MergeRequest: &tui.MergeRequestDetails{
				IID:            iid,
				State:          mrState,
				Author:         "Mock Author",
				AuthorLogin:    authorLogin,
				Assignees:      []string{"Mock Assignee"},
				AssigneeLogins: []string{assigneeLogin},
				Labels:         []string{"mock", "ui"},
				Milestone:      "Iteration 1",
				SourceBranch:   fmt.Sprintf("feature/mock-%02d", i),
				TargetBranch:   "main",
				CreatedAt:      "2026-01-01 10:00 UTC",
				UpdatedAt:      "2026-01-02 11:00 UTC",
				URL:            url,
				Description:    "Mock merge request description for validating detail rendering and scroll behavior.",
			},
		})
	}

	start := (query.Page - 1) * query.PerPage
	if start >= len(items) {
		return tui.MergeRequestResult{Items: []tui.ListItem{}, HasNextPage: false}, nil
	}
	end := start + query.PerPage
	if end > len(items) {
		end = len(items)
	}

	return tui.MergeRequestResult{Items: items[start:end], HasNextPage: end < len(items)}, nil
}

func (p *MockProvider) LoadIssueDetailData(_ context.Context, issueIID int64) (tui.IssueDetailData, error) {
	if issueIID <= 0 {
		return tui.IssueDetailData{}, fmt.Errorf("invalid issue IID: %d", issueIID)
	}

	return tui.IssueDetailData{
		Activities: []tui.IssueActivity{
			{Actor: "Mock Author", CreatedAt: "2026-01-02 11:30 UTC", Action: "closed"},
			{Actor: "Mock Assignee", CreatedAt: "2026-01-02 11:20 UTC", Action: "reopened"},
			{Actor: "Mock Author", CreatedAt: "2026-01-01 10:10 UTC", Action: "added label ~ui"},
		},
		Comments: []tui.IssueComment{
			{Author: "Mock Reviewer", CreatedAt: "2026-01-02 11:10 UTC", Body: "Looks good overall.\n\n- Please update the loading copy\n- Add one more test"},
			{Author: "Mock Author", CreatedAt: "2026-01-01 10:05 UTC", Body: "Initial report with **markdown** content and a code block:\n\n```go\nfmt.Println(\"hello\")\n```"},
		},
	}, nil
}

func (p *MockProvider) LoadSearchMetadata(context.Context, tui.ViewMode) (tui.SearchMetadata, error) {
	return tui.SearchMetadata{
		Authors:    []tui.SearchUser{{Name: "Mock Author", Username: "mock-author"}, {Name: "Mock Reviewer", Username: "mock-reviewer"}},
		Assignees:  []tui.SearchUser{{Name: "Mock Assignee", Username: "mock-assignee"}, {Name: "Mock Author", Username: "mock-author"}},
		Labels:     []string{"mock", "ui", "backend"},
		Milestones: []string{"Iteration 1", "Iteration 2"},
	}, nil
}

func containsAll(values []string, required []string) bool {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[strings.ToLower(strings.TrimSpace(value))] = struct{}{}
	}
	for _, req := range required {
		if _, ok := set[strings.ToLower(strings.TrimSpace(req))]; !ok {
			return false
		}
	}
	return true
}
