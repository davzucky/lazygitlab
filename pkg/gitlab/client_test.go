package gitlab

import (
	"testing"

	"gitlab.com/gitlab-org/api/client-go"
)

type mockClient struct {
	user         *gitlab.User
	project      *gitlab.Project
	issues       []*gitlab.Issue
	mergeReqs    []*gitlab.BasicMergeRequest
	userErr      error
	projectErr   error
	issuesErr    error
	mergeReqsErr error
}

func (m *mockClient) GetCurrentUser() (*gitlab.User, error) {
	return m.user, m.userErr
}

func (m *mockClient) GetProject(projectPath string) (*gitlab.Project, error) {
	return m.project, m.projectErr
}

func (m *mockClient) GetIssues(projectPath string, opts *GetIssuesOptions) ([]*gitlab.Issue, error) {
	return m.issues, m.issuesErr
}

func (m *mockClient) GetMergeRequests(projectPath string, opts *GetMergeRequestsOptions) ([]*gitlab.BasicMergeRequest, error) {
	return m.mergeReqs, m.mergeReqsErr
}

func (m *mockClient) Close() error {
	return nil
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		host    string
		wantErr bool
	}{
		{
			name:    "valid token and host",
			token:   "test-token",
			host:    "https://gitlab.com",
			wantErr: false,
		},
		{
			name:    "empty token",
			token:   "",
			host:    "https://gitlab.com",
			wantErr: true,
		},
		{
			name:    "empty host",
			token:   "test-token",
			host:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.token, tt.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetIssuesOptions(t *testing.T) {
	opts := &GetIssuesOptions{
		State:   "opened",
		Page:    1,
		PerPage: 50,
	}

	if opts.State != "opened" {
		t.Errorf("Expected State to be 'opened', got '%s'", opts.State)
	}
	if opts.Page != 1 {
		t.Errorf("Expected Page to be 1, got %d", opts.Page)
	}
	if opts.PerPage != 50 {
		t.Errorf("Expected PerPage to be 50, got %d", opts.PerPage)
	}
}

func TestGetMergeRequestsOptions(t *testing.T) {
	opts := &GetMergeRequestsOptions{
		State:   "opened",
		Page:    1,
		PerPage: 50,
	}

	if opts.State != "opened" {
		t.Errorf("Expected State to be 'opened', got '%s'", opts.State)
	}
	if opts.Page != 1 {
		t.Errorf("Expected Page to be 1, got %d", opts.Page)
	}
	if opts.PerPage != 50 {
		t.Errorf("Expected PerPage to be 50, got %d", opts.PerPage)
	}
}

func TestMockClient(t *testing.T) {
	mock := &mockClient{
		user: &gitlab.User{
			ID:       1,
			Username: "testuser",
		},
	}

	user, err := mock.GetCurrentUser()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("Expected user to be non-nil")
	}
	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}

	mock.userErr = Error("test error")
	_, err = mock.GetCurrentUser()
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func Error(s string) error {
	return &errorString{s}
}

type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}
