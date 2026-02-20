package tui

import "context"

type ViewMode int

const (
	PrimaryView ViewMode = iota
	IssuesView
	MergeRequestsView
)

type ListItem struct {
	ID           int64
	Title        string
	Subtitle     string
	URL          string
	Issue        *IssueDetails
	MergeRequest *MergeRequestDetails
}

type IssueDetails struct {
	IID            int64
	State          string
	Author         string
	AuthorLogin    string
	Assignees      []string
	AssigneeLogins []string
	Labels         []string
	Milestone      string
	CreatedAt      string
	UpdatedAt      string
	URL            string
	Description    string
}

type IssueComment struct {
	Author    string
	CreatedAt string
	Body      string
}

type IssueActivity struct {
	Actor     string
	CreatedAt string
	Action    string
}

type IssueDetailData struct {
	Comments   []IssueComment
	Activities []IssueActivity
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

type MergeRequestState string

const (
	MergeRequestStateOpened MergeRequestState = "opened"
	MergeRequestStateMerged MergeRequestState = "merged"
	MergeRequestStateClosed MergeRequestState = "closed"
	MergeRequestStateAll    MergeRequestState = "all"
)

type MergeRequestQuery struct {
	State   MergeRequestState
	Search  string
	Page    int
	PerPage int
}

type MergeRequestResult struct {
	Items       []ListItem
	HasNextPage bool
}

type MergeRequestDetails struct {
	IID            int64
	State          string
	Author         string
	AuthorLogin    string
	Assignees      []string
	AssigneeLogins []string
	Labels         []string
	Milestone      string
	SourceBranch   string
	TargetBranch   string
	CreatedAt      string
	UpdatedAt      string
	URL            string
	Description    string
}

type DataProvider interface {
	LoadIssues(ctx context.Context, query IssueQuery) (IssueResult, error)
	LoadMergeRequests(ctx context.Context, query MergeRequestQuery) (MergeRequestResult, error)
	LoadIssueDetailData(ctx context.Context, issueIID int64) (IssueDetailData, error)
	LoadSearchMetadata(ctx context.Context, view ViewMode) (SearchMetadata, error)
}

type SearchMetadata struct {
	Authors    []string
	Assignees  []string
	Labels     []string
	Milestones []string
}

type DashboardContext struct {
	ProjectPath string
	Connection  string
	Host        string
}
