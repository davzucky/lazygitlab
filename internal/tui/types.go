package tui

import "context"

type ViewMode int

const (
	IssuesView ViewMode = iota
	MergeRequestsView
)

type ListItem struct {
	ID       int64
	Title    string
	Subtitle string
	URL      string
}

type IssueState string

const (
	IssueStateOpened IssueState = "opened"
	IssueStateClosed IssueState = "closed"
	IssueStateAll    IssueState = "all"
)

type IssueQuery struct {
	State   IssueState
	Search  string
	Page    int
	PerPage int
}

type IssueResult struct {
	Items       []ListItem
	HasNextPage bool
}

type DataProvider interface {
	LoadIssues(ctx context.Context, query IssueQuery) (IssueResult, error)
	LoadMergeRequests(ctx context.Context) ([]ListItem, error)
}

type DashboardContext struct {
	ProjectPath string
	Connection  string
	Host        string
}
