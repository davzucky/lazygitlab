package tui

import "context"

type ViewMode int

const (
	ProjectsView ViewMode = iota
	IssuesView
	MergeRequestsView
)

type ListItem struct {
	ID       int64
	Title    string
	Subtitle string
	URL      string
}

type DataProvider interface {
	LoadProjects(ctx context.Context) ([]ListItem, error)
	LoadIssues(ctx context.Context) ([]ListItem, error)
	LoadMergeRequests(ctx context.Context) ([]ListItem, error)
}

type DashboardContext struct {
	ProjectPath string
	Connection  string
	Host        string
}
