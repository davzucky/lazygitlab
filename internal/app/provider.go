package app

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

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
		author := "-"
		if issue.Author != nil {
			author = displayName(issue.Author.Name, issue.Author.Username)
		}
		assignees := make([]string, 0, len(issue.Assignees))
		for _, assignee := range issue.Assignees {
			if assignee == nil {
				continue
			}
			assignees = append(assignees, displayName(assignee.Name, assignee.Username))
		}
		labels := make([]string, 0, len(issue.Labels))
		for _, label := range issue.Labels {
			if strings.TrimSpace(label) == "" {
				continue
			}
			labels = append(labels, label)
		}
		items = append(items, tui.ListItem{
			ID:       issue.ID,
			Title:    issue.Title,
			Subtitle: subtitle,
			URL:      issue.WebURL,
			Issue: &tui.IssueDetails{
				IID:         issue.IID,
				State:       issue.State,
				Author:      author,
				Assignees:   assignees,
				Labels:      labels,
				CreatedAt:   formatIssueTime(issue.CreatedAt),
				UpdatedAt:   formatIssueTime(issue.UpdatedAt),
				URL:         issue.WebURL,
				Description: issue.Description,
			},
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

func (p *Provider) LoadIssueDetailData(ctx context.Context, issueIID int64) (tui.IssueDetailData, error) {
	if p.projectPath == "" {
		return tui.IssueDetailData{}, fmt.Errorf("no project context selected")
	}
	if issueIID <= 0 {
		return tui.IssueDetailData{}, fmt.Errorf("invalid issue IID: %d", issueIID)
	}

	notes, err := p.client.ListIssueNotes(ctx, p.projectPath, issueIID)
	if err != nil {
		return tui.IssueDetailData{}, fmt.Errorf("load issue notes: %w", err)
	}
	stateEvents, err := p.client.ListIssueStateEvents(ctx, p.projectPath, issueIID)
	if err != nil {
		return tui.IssueDetailData{}, fmt.Errorf("load issue state events: %w", err)
	}

	comments := make([]tui.IssueComment, 0, len(notes))
	activities := make([]tui.IssueActivity, 0, len(notes)+len(stateEvents))

	for _, note := range notes {
		if note == nil {
			continue
		}
		author := displayName(note.Author.Name, note.Author.Username)
		createdAt := formatIssueTime(note.CreatedAt)
		body := strings.TrimSpace(note.Body)
		if note.System {
			if body == "" {
				body = "System activity"
			}
			activities = append(activities, tui.IssueActivity{Actor: author, CreatedAt: createdAt, Action: body})
			continue
		}
		if body == "" {
			continue
		}
		comments = append(comments, tui.IssueComment{Author: author, CreatedAt: createdAt, Body: body})
	}

	for _, event := range stateEvents {
		if event == nil {
			continue
		}
		actor := "-"
		if event.User != nil {
			actor = displayName(event.User.Name, event.User.Username)
		}
		action := strings.TrimSpace(string(event.State))
		if action == "" {
			action = "state changed"
		}
		activities = append(activities, tui.IssueActivity{
			Actor:     actor,
			CreatedAt: formatIssueTime(event.CreatedAt),
			Action:    action,
		})
	}

	sort.SliceStable(comments, func(i int, j int) bool {
		return comments[i].CreatedAt > comments[j].CreatedAt
	})
	sort.SliceStable(activities, func(i int, j int) bool {
		return activities[i].CreatedAt > activities[j].CreatedAt
	})

	return tui.IssueDetailData{Comments: comments, Activities: activities}, nil
}

func displayName(name string, username string) string {
	trimmedName := strings.TrimSpace(name)
	trimmedUser := strings.TrimSpace(username)
	if trimmedName == "" && trimmedUser == "" {
		return "-"
	}
	if trimmedName == "" {
		return trimmedUser
	}
	return trimmedName
}

func formatIssueTime(value *time.Time) string {
	if value == nil {
		return "-"
	}
	return value.Local().Format("2006-01-02 15:04 MST")
}
