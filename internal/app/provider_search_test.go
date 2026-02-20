package app

import (
	"context"
	"testing"

	"github.com/davzucky/lazygitlab/internal/gitlab"
	"github.com/davzucky/lazygitlab/internal/tui"
	gl "gitlab.com/gitlab-org/api/client-go"
)

type captureClient struct {
	issueOpts gitlab.IssueListOptions
	mrOpts    gitlab.MergeRequestListOptions
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
