package gitlab

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	gl "gitlab.com/gitlab-org/api/client-go"
)

const (
	defaultPerPage = 50
	maxRetries     = 3
)

type Client interface {
	GetCurrentUser(ctx context.Context) (*gl.User, error)
	GetProject(ctx context.Context, projectPath string) (*gl.Project, error)
	ListProjects(ctx context.Context, search string) ([]*gl.Project, error)
	ListIssues(ctx context.Context, projectPath string, state string) ([]*gl.Issue, error)
	ListMergeRequests(ctx context.Context, projectPath string, state string) ([]*gl.BasicMergeRequest, error)
}

type client struct {
	api    *gl.Client
	logger *log.Logger
}

func NewClient(token string, host string, logger *log.Logger) (Client, error) {
	if token == "" {
		return nil, fmt.Errorf("gitlab token is required")
	}
	if host == "" {
		return nil, fmt.Errorf("gitlab host is required")
	}

	api, err := gl.NewClient(token, gl.WithBaseURL(host))
	if err != nil {
		return nil, fmt.Errorf("create GitLab client: %w", err)
	}

	return &client{api: api, logger: logger}, nil
}

func (c *client) GetCurrentUser(ctx context.Context) (*gl.User, error) {
	var user *gl.User
	err := c.withRetry(ctx, "GetCurrentUser", func() (*gl.Response, error) {
		var err error
		user, _, err = c.api.Users.CurrentUser(gl.WithContext(ctx))
		return nil, err
	})
	if err != nil {
		return nil, fmt.Errorf("get current user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("gitlab returned empty user")
	}
	return user, nil
}

func (c *client) GetProject(ctx context.Context, projectPath string) (*gl.Project, error) {
	var project *gl.Project
	err := c.withRetry(ctx, "GetProject", func() (*gl.Response, error) {
		var err error
		project, _, err = c.api.Projects.GetProject(projectPath, nil, gl.WithContext(ctx))
		return nil, err
	})
	if err != nil {
		return nil, fmt.Errorf("get project %q: %w", projectPath, err)
	}
	if project == nil {
		return nil, fmt.Errorf("project not found: %s", projectPath)
	}

	return project, nil
}

func (c *client) ListProjects(ctx context.Context, search string) ([]*gl.Project, error) {
	all := make([]*gl.Project, 0, defaultPerPage)
	page := int64(1)

	for {
		opts := &gl.ListProjectsOptions{
			Search: gl.Ptr(search),
			ListOptions: gl.ListOptions{
				Page:    page,
				PerPage: defaultPerPage,
			},
			Membership: gl.Ptr(true),
		}

		var projects []*gl.Project
		var resp *gl.Response
		err := c.withRetry(ctx, "ListProjects", func() (*gl.Response, error) {
			var err error
			projects, resp, err = c.api.Projects.ListProjects(opts, gl.WithContext(ctx))
			return resp, err
		})
		if err != nil {
			return nil, fmt.Errorf("list projects: %w", err)
		}

		all = append(all, projects...)
		if resp == nil || resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	return all, nil
}

func (c *client) ListIssues(ctx context.Context, projectPath string, state string) ([]*gl.Issue, error) {
	all := make([]*gl.Issue, 0, defaultPerPage)
	page := int64(1)

	for {
		opts := &gl.ListProjectIssuesOptions{
			ListOptions: gl.ListOptions{Page: page, PerPage: defaultPerPage},
		}
		if state != "" {
			opts.State = gl.Ptr(state)
		}

		var issues []*gl.Issue
		var resp *gl.Response
		err := c.withRetry(ctx, "ListIssues", func() (*gl.Response, error) {
			var err error
			issues, resp, err = c.api.Issues.ListProjectIssues(projectPath, opts, gl.WithContext(ctx))
			return resp, err
		})
		if err != nil {
			return nil, fmt.Errorf("list issues for project %q: %w", projectPath, err)
		}

		all = append(all, issues...)
		if resp == nil || resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	return all, nil
}

func (c *client) ListMergeRequests(ctx context.Context, projectPath string, state string) ([]*gl.BasicMergeRequest, error) {
	all := make([]*gl.BasicMergeRequest, 0, defaultPerPage)
	page := int64(1)

	for {
		opts := &gl.ListProjectMergeRequestsOptions{
			ListOptions: gl.ListOptions{Page: page, PerPage: defaultPerPage},
		}
		if state != "" {
			opts.State = gl.Ptr(state)
		}

		var mrs []*gl.BasicMergeRequest
		var resp *gl.Response
		err := c.withRetry(ctx, "ListMergeRequests", func() (*gl.Response, error) {
			var err error
			mrs, resp, err = c.api.MergeRequests.ListProjectMergeRequests(projectPath, opts, gl.WithContext(ctx))
			return resp, err
		})
		if err != nil {
			return nil, fmt.Errorf("list merge requests for project %q: %w", projectPath, err)
		}

		all = append(all, mrs...)
		if resp == nil || resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	return all, nil
}

func (c *client) withRetry(ctx context.Context, operation string, fn func() (*gl.Response, error)) error {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		resp, err := fn()
		if err == nil {
			return nil
		}

		lastErr = err
		if !isRetryable(resp) || attempt == maxRetries-1 {
			break
		}

		delay := retryDelay(resp, attempt)
		c.logger.Printf("retrying %s attempt=%d delay=%s err=%v", operation, attempt+2, delay, err)
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}

	return lastErr
}

func isRetryable(resp *gl.Response) bool {
	if resp == nil || resp.Response == nil {
		return true
	}

	status := resp.StatusCode
	return status == http.StatusTooManyRequests || status >= http.StatusInternalServerError
}

func retryDelay(resp *gl.Response, attempt int) time.Duration {
	if resp != nil && resp.Response != nil {
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter != "" {
			if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds > 0 {
				return time.Duration(seconds) * time.Second
			}
		}
	}

	base := 300 * time.Millisecond
	return base * time.Duration(1<<attempt)
}
