package gitlab

import (
	"fmt"

	"gitlab.com/gitlab-org/api/client-go"
)

type Client interface {
	GetCurrentUser() (*gitlab.User, error)
	GetProject(projectPath string) (*gitlab.Project, error)
	GetIssues(projectPath string, opts *GetIssuesOptions) ([]*gitlab.Issue, error)
	GetProjectIssue(projectPath string, issueIID int64) (*gitlab.Issue, error)
	GetMergeRequests(projectPath string, opts *GetMergeRequestsOptions) ([]*gitlab.BasicMergeRequest, error)
	Close() error
}

type GetIssuesOptions struct {
	State   string
	Page    int64
	PerPage int64
}

type GetMergeRequestsOptions struct {
	State   string
	Page    int64
	PerPage int64
}

type client struct {
	client *gitlab.Client
}

func NewClient(token, host string) (Client, error) {
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}
	if host == "" {
		return nil, fmt.Errorf("host is required")
	}

	glClient, err := gitlab.NewClient(token, gitlab.WithBaseURL(host))
	if err != nil {
		return nil, fmt.Errorf("failed to create GitLab client: %w", err)
	}

	return &client{client: glClient}, nil
}

func (c *client) GetCurrentUser() (*gitlab.User, error) {
	user, _, err := c.client.Users.CurrentUser()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("received nil user from API")
	}
	return user, nil
}

func (c *client) GetProject(projectPath string) (*gitlab.Project, error) {
	project, _, err := c.client.Projects.GetProject(projectPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %s: %w", projectPath, err)
	}
	if project == nil {
		return nil, fmt.Errorf("project not found: %s", projectPath)
	}
	return project, nil
}

func (c *client) GetIssues(projectPath string, opts *GetIssuesOptions) ([]*gitlab.Issue, error) {
	options := &gitlab.ListProjectIssuesOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 20,
		},
	}

	if opts != nil {
		if opts.State != "" {
			options.State = &opts.State
		}
		if opts.PerPage > 0 {
			options.ListOptions.PerPage = opts.PerPage
		}
	}

	var allIssues []*gitlab.Issue
	var page int64 = 1
	if opts != nil && opts.Page > 0 {
		page = int64(opts.Page)
	}

	for {
		options.ListOptions.Page = page
		issues, resp, err := c.client.Issues.ListProjectIssues(projectPath, options)
		if err != nil {
			return nil, fmt.Errorf("failed to list issues for project %s: %w", projectPath, err)
		}

		allIssues = append(allIssues, issues...)

		if resp.NextPage == 0 || (opts != nil && opts.Page > 0) {
			break
		}

		page = resp.NextPage
	}

	return allIssues, nil
}

func (c *client) GetProjectIssue(projectPath string, issueIID int64) (*gitlab.Issue, error) {
	issue, _, err := c.client.Issues.GetIssue(projectPath, issueIID)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue %d from project %s: %w", issueIID, projectPath, err)
	}
	if issue == nil {
		return nil, fmt.Errorf("issue not found: %d", issueIID)
	}
	return issue, nil
}

func (c *client) GetMergeRequests(projectPath string, opts *GetMergeRequestsOptions) ([]*gitlab.BasicMergeRequest, error) {
	options := &gitlab.ListProjectMergeRequestsOptions{
		ListOptions: gitlab.ListOptions{
			PerPage: 20,
		},
	}

	if opts != nil {
		if opts.State != "" {
			options.State = &opts.State
		}
		if opts.PerPage > 0 {
			options.ListOptions.PerPage = opts.PerPage
		}
	}

	var allMRs []*gitlab.BasicMergeRequest
	var page int64 = 1
	if opts != nil && opts.Page > 0 {
		page = int64(opts.Page)
	}

	for {
		options.ListOptions.Page = page
		mrs, resp, err := c.client.MergeRequests.ListProjectMergeRequests(projectPath, options)
		if err != nil {
			return nil, fmt.Errorf("failed to list merge requests for project %s: %w", projectPath, err)
		}

		allMRs = append(allMRs, mrs...)

		if resp.NextPage == 0 || (opts != nil && opts.Page > 0) {
			break
		}

		page = resp.NextPage
	}

	return allMRs, nil
}

func (c *client) Close() error {
	return nil
}
