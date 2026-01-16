# GitLab API Client

## pkg/gitlab/ Package

The gitlab package provides an abstraction layer over GitLab's Go SDK (gitlab.com/gitlab-org/api/client-go).

### API Patterns

- Use the `Client` interface for all GitLab interactions (enables mocking in tests)
- Call `client.Close()` when done to clean up resources
- Pagination is automatic for list endpoints (issues, merge requests) unless you specify a specific page
- All API errors are wrapped with context using `fmt.Errorf("context: %w", err)`

### Type Notes

- `ListProjectMergeRequests` returns `[]*gitlab.BasicMergeRequest`, not `[]*gitlab.MergeRequest`
- Pagination fields (Page, PerPage) are `int64`, not `int`
- String options need pointer to string (`&str`), not a helper function like `gitlab.String()`

### Testing

- Use the `mockClient` struct in client_test.go as a template for testing code that uses the GitLab API
- The mock implements the same interface as the real client for easy swapping
