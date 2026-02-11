package app

import (
	"context"
	"fmt"

	"github.com/davzucky/lazygitlab/internal/gitlab"
	"github.com/davzucky/lazygitlab/internal/tui"
)

type Provider struct {
	client      gitlab.Client
	projectPath string
}

func NewProvider(client gitlab.Client, projectPath string) *Provider {
	return &Provider{client: client, projectPath: projectPath}
}

func (p *Provider) LoadProjects(ctx context.Context) ([]tui.ListItem, error) {
	projects, err := p.client.ListProjects(ctx, "")
	if err != nil {
		return nil, err
	}

	items := make([]tui.ListItem, 0, len(projects))
	for _, project := range projects {
		items = append(items, tui.ListItem{
			ID:       project.ID,
			Title:    project.PathWithNamespace,
			Subtitle: project.Description,
			URL:      project.WebURL,
		})
	}

	return items, nil
}

func (p *Provider) LoadIssues(ctx context.Context, query tui.IssueQuery) (tui.IssueResult, error) {
	if p.projectPath == "" {
		return tui.IssueResult{}, fmt.Errorf("no project context selected")
	}

	issues, hasNextPage, err := p.client.ListIssues(ctx, p.projectPath, gitlab.IssueListOptions{
		State:   string(query.State),
		Search:  query.Search,
		Page:    int64(query.Page),
		PerPage: query.PerPage,
	})
	if err != nil {
		return tui.IssueResult{}, err
	}

	items := make([]tui.ListItem, 0, len(issues))
	for _, issue := range issues {
		subtitle := fmt.Sprintf("#%d • %s", issue.IID, issue.State)
		items = append(items, tui.ListItem{
			ID:       issue.ID,
			Title:    issue.Title,
			Subtitle: subtitle,
			URL:      issue.WebURL,
		})
	}

	return tui.IssueResult{Items: items, HasNextPage: hasNextPage}, nil
}

func (p *Provider) LoadMergeRequests(ctx context.Context) ([]tui.ListItem, error) {
	if p.projectPath == "" {
		return nil, fmt.Errorf("no project context selected")
	}

	mrs, err := p.client.ListMergeRequests(ctx, p.projectPath, "opened")
	if err != nil {
		return nil, err
	}

	items := make([]tui.ListItem, 0, len(mrs))
	for _, mr := range mrs {
		subtitle := fmt.Sprintf("!%d • %s", mr.IID, mr.State)
		items = append(items, tui.ListItem{
			ID:       mr.ID,
			Title:    mr.Title,
			Subtitle: subtitle,
			URL:      mr.WebURL,
		})
	}

	return items, nil
}
