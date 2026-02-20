package app

import (
	"context"
	"testing"

	"github.com/davzucky/lazygitlab/internal/gitlab"
	"github.com/davzucky/lazygitlab/internal/tui"
	gl "gitlab.com/gitlab-org/api/client-go"
)

type captureClient struct {
	issueOpts  gitlab.IssueListOptions
	mrOpts     gitlab.MergeRequestListOptions
	members    []*gl.ProjectMember
	labels     []*gl.Label
	milestones []*gl.Milestone
}

func (c *captureClient) GetCurrentUser(context.Context) (*gl.User, error) {
	return nil, nil
}

func (c *captureClient) GetProject(context.Context, string) (*gl.Project, error) {
	return nil, nil
}

func (c *captureClient) ListProjects(context.Context, string) ([]*gl.Project, error) {
	return nil, nil
}

func (c *captureClient) ListProjectMembers(context.Context, string, string) ([]*gl.ProjectMember, error) {
	return c.members, nil
}

func (c *captureClient) ListProjectLabels(context.Context, string, string) ([]*gl.Label, error) {
	return c.labels, nil
}

func (c *captureClient) ListProjectMilestones(context.Context, string, string) ([]*gl.Milestone, error) {
	return c.milestones, nil
}

func (c *captureClient) ListIssues(_ context.Context, _ string, opts gitlab.IssueListOptions) ([]*gl.Issue, bool, error) {
	c.issueOpts = opts
	return nil, false, nil
}

func (c *captureClient) ListIssueNotes(context.Context, string, int64) ([]*gl.Note, error) {
	return nil, nil
}

func (c *captureClient) ListIssueStateEvents(context.Context, string, int64) ([]*gl.StateEvent, error) {
	return nil, nil
}

func (c *captureClient) ListMergeRequests(_ context.Context, _ string, opts gitlab.MergeRequestListOptions) ([]*gl.BasicMergeRequest, bool, error) {
	c.mrOpts = opts
	return nil, false, nil
}

func TestProviderParsesIssueMetadataSearch(t *testing.T) {
	t.Parallel()

	client := &captureClient{}
	provider := NewProvider(client, "group/project")

	_, err := provider.LoadIssues(context.Background(), tui.IssueQuery{State: tui.IssueStateOpened, Search: "login author:alice label:backend milestone:v1"})
	if err != nil {
		t.Fatalf("LoadIssues() error = %v", err)
	}

	if client.issueOpts.Search != "login" {
		t.Fatalf("Search = %q want %q", client.issueOpts.Search, "login")
	}
	if client.issueOpts.AuthorUsername != "alice" {
		t.Fatalf("AuthorUsername = %q want %q", client.issueOpts.AuthorUsername, "alice")
	}
	if len(client.issueOpts.Labels) != 1 || client.issueOpts.Labels[0] != "backend" {
		t.Fatalf("Labels = %#v want %#v", client.issueOpts.Labels, []string{"backend"})
	}
	if client.issueOpts.Milestone != "v1" {
		t.Fatalf("Milestone = %q want %q", client.issueOpts.Milestone, "v1")
	}
}

func TestProviderParsesMergeRequestMetadataSearch(t *testing.T) {
	t.Parallel()

	client := &captureClient{}
	provider := NewProvider(client, "group/project")

	_, err := provider.LoadMergeRequests(context.Background(), tui.MergeRequestQuery{State: tui.MergeRequestStateOpened, Search: "cleanup author:alice label:frontend"})
	if err != nil {
		t.Fatalf("LoadMergeRequests() error = %v", err)
	}

	if client.mrOpts.Search != "cleanup" {
		t.Fatalf("Search = %q want %q", client.mrOpts.Search, "cleanup")
	}
	if client.mrOpts.AuthorUsername != "alice" {
		t.Fatalf("AuthorUsername = %q want %q", client.mrOpts.AuthorUsername, "alice")
	}
	if len(client.mrOpts.Labels) != 1 || client.mrOpts.Labels[0] != "frontend" {
		t.Fatalf("Labels = %#v want %#v", client.mrOpts.Labels, []string{"frontend"})
	}
}

func TestProviderLoadsSearchMetadataFromGitLab(t *testing.T) {
	t.Parallel()

	client := &captureClient{
		members:    []*gl.ProjectMember{{Username: "alice"}, {Username: "bob"}, {Username: "alice"}},
		labels:     []*gl.Label{{Name: "backend"}, {Name: "ui"}},
		milestones: []*gl.Milestone{{Title: "Sprint 1"}},
	}
	provider := NewProvider(client, "group/project")

	metadata, err := provider.LoadSearchMetadata(context.Background(), tui.IssuesView)
	if err != nil {
		t.Fatalf("LoadSearchMetadata() error = %v", err)
	}

	if len(metadata.Authors) != 2 || metadata.Authors[0].Username != "alice" || metadata.Authors[1].Username != "bob" {
		t.Fatalf("Authors = %#v", metadata.Authors)
	}
	if len(metadata.Assignees) != 2 || metadata.Assignees[0].Username != "alice" || metadata.Assignees[1].Username != "bob" {
		t.Fatalf("Assignees = %#v", metadata.Assignees)
	}
	if len(metadata.Labels) != 2 || metadata.Labels[0] != "backend" || metadata.Labels[1] != "ui" {
		t.Fatalf("Labels = %#v want %#v", metadata.Labels, []string{"backend", "ui"})
	}
	if len(metadata.Milestones) != 1 || metadata.Milestones[0] != "Sprint 1" {
		t.Fatalf("Milestones = %#v want %#v", metadata.Milestones, []string{"Sprint 1"})
	}
}

func TestProviderResolvesAuthorNameToUsername(t *testing.T) {
	t.Parallel()

	client := &captureClient{members: []*gl.ProjectMember{{Name: "Alice Doe", Username: "d2db03f2-aaaa-bbbb-cccc-f59ec9974ed2"}}}
	provider := NewProvider(client, "group/project")

	_, err := provider.LoadIssues(context.Background(), tui.IssueQuery{State: tui.IssueStateOpened, Search: "author:Alice Doe label:backend"})
	if err != nil {
		t.Fatalf("LoadIssues() error = %v", err)
	}

	if client.issueOpts.AuthorUsername != "d2db03f2-aaaa-bbbb-cccc-f59ec9974ed2" {
		t.Fatalf("AuthorUsername = %q", client.issueOpts.AuthorUsername)
	}
}
