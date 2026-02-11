package app

import (
	"context"
	"fmt"

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

func (p *MockProvider) LoadIssues(context.Context) ([]tui.ListItem, error) {
	items := make([]tui.ListItem, 0, 40)
	for i := 40; i >= 1; i-- {
		items = append(items, tui.ListItem{
			ID:       int64(2000 + i),
			Title:    fmt.Sprintf("Mock issue %02d with long title to validate clipping and stable panel width behavior", i),
			Subtitle: fmt.Sprintf("#%d • opened", 3000+i),
			URL:      fmt.Sprintf("https://mock.gitlab.local/mock/group/project/-/issues/%d", 3000+i),
		})
	}

	return items, nil
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
