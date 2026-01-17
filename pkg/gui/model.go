package gui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	spin "github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gl "github.com/davzucky/lazygitlab/pkg/gitlab"
	"github.com/davzucky/lazygitlab/pkg/utils"
	"gitlab.com/gitlab-org/api/client-go"
)

type ViewMode int

const (
	ProjectsView ViewMode = iota
	IssuesView
	MergeRequestsView
)

type IssueFilterState int

const (
	FilterAll IssueFilterState = iota
	FilterOpen
	FilterClosed
)

type DetailSection int

const (
	DescriptionSection DetailSection = iota
	CommentsSection
)

func (s DetailSection) String() string {
	switch s {
	case DescriptionSection:
		return "Description"
	case CommentsSection:
		return "Comments"
	default:
		return "Description"
	}
}

func (f IssueFilterState) String() string {
	switch f {
	case FilterAll:
		return "All"
	case FilterOpen:
		return "Open"
	case FilterClosed:
		return "Closed"
	default:
		return "All"
	}
}

func (f IssueFilterState) ToAPIState() string {
	switch f {
	case FilterAll:
		return ""
	case FilterOpen:
		return "opened"
	case FilterClosed:
		return "closed"
	default:
		return ""
	}
}

type issuesLoadedMsg struct {
	issues []*gitlab.Issue
}

type issuesLoadedErrMsg struct {
	err error
}

type issueDetailLoadedMsg struct {
	issue *gitlab.Issue
}

type commentsLoadedMsg struct {
	comments []*gitlab.Note
}

type commentsLoadedErrMsg struct {
	err error
}

type issueCreatedMsg struct {
	issue *gitlab.Issue
}

type issueCreatedErrMsg struct {
	err error
}

type issueUpdatedMsg struct {
	issue *gitlab.Issue
}

type issueUpdatedErrMsg struct {
	err error
}

type commentCreatedMsg struct {
	comment *gitlab.Note
}

type commentCreatedErrMsg struct {
	err error
}

type clipboardClearMsg struct{}

type Model struct {
	currentView         ViewMode
	items               []ListItem
	selectedItem        int
	projectPath         string
	connection          string
	width               int
	height              int
	styles              *Style
	showHelp            bool
	showError           bool
	errorMessage        string
	isLoading           bool
	spinner             spin.Model
	selectedIssue       *gitlab.Issue
	issueFilter         IssueFilterState
	detailSection       DetailSection
	selectedComments    []*gitlab.Note
	client              gitlabClient
	showCreateIssueForm bool
	issueFormTitle      string
	issueFormDesc       string
	issueFormField      string
	showConfirmPopup    bool
	confirmAction       string
	confirmIssueIID     int64
	clipboardMessage    string
	showCommentForm     bool
	commentFormBody     string
}

type gitlabClient interface {
	GetIssueNotes(projectPath string, issueIID int64, opts *gl.GetIssueNotesOptions) ([]*gitlab.Note, error)
	CreateIssueNote(projectPath string, issueIID int64, opts *gl.CreateIssueNoteOptions) (*gitlab.Note, error)
	CreateIssue(projectPath string, opts *gl.CreateIssueOptions) (*gitlab.Issue, error)
	UpdateIssue(projectPath string, issueIID int64, opts *gl.UpdateIssueOptions) (*gitlab.Issue, error)
}

type ListItem struct {
	ID        int
	Title     string
	State     string
	UpdatedAt time.Time
	CreatedAt time.Time
	Author    string
	Desc      string
	Labels    []string
	Assignees []string
	Milestone string
}

func NewModel(projectPath string, connection string, client gitlabClient) Model {
	styles := NewStyle()
	spinner := spin.New()
	spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("62"))

	return Model{
		currentView:         ProjectsView,
		items:               []ListItem{},
		selectedItem:        0,
		projectPath:         projectPath,
		connection:          connection,
		width:               80,
		height:              24,
		styles:              styles,
		showError:           false,
		errorMessage:        "",
		isLoading:           false,
		spinner:             spinner,
		issueFilter:         FilterAll,
		detailSection:       DescriptionSection,
		selectedComments:    []*gitlab.Note{},
		client:              client,
		showCreateIssueForm: false,
		issueFormTitle:      "",
		issueFormDesc:       "",
		issueFormField:      "title",
		showConfirmPopup:    false,
		confirmAction:       "",
		confirmIssueIID:     0,
	}
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func loadCommentsCmd(projectPath string, issueIID int64, client gitlabClient) tea.Cmd {
	return func() tea.Msg {
		comments, err := client.GetIssueNotes(projectPath, issueIID, nil)
		if err != nil {
			return commentsLoadedErrMsg{err: err}
		}
		return commentsLoadedMsg{comments: comments}
	}
}

func createIssueCmd(projectPath, title, description string, client gitlabClient) tea.Cmd {
	return func() tea.Msg {
		opts := &gl.CreateIssueOptions{
			Title:       title,
			Description: description,
		}
		issue, err := client.CreateIssue(projectPath, opts)
		if err != nil {
			return issueCreatedErrMsg{err: err}
		}
		return issueCreatedMsg{issue: issue}
	}
}

func updateIssueCmd(projectPath string, issueIID int64, stateEvent string, client gitlabClient) tea.Cmd {
	return func() tea.Msg {
		opts := &gl.UpdateIssueOptions{
			StateEvent: stateEvent,
		}
		issue, err := client.UpdateIssue(projectPath, issueIID, opts)
		if err != nil {
			return issueUpdatedErrMsg{err: err}
		}
		return issueUpdatedMsg{issue: issue}
	}
}

func createCommentCmd(projectPath string, issueIID int64, body string, client gitlabClient) tea.Cmd {
	return func() tea.Msg {
		opts := &gl.CreateIssueNoteOptions{
			Body: body,
		}
		comment, err := client.CreateIssueNote(projectPath, issueIID, opts)
		if err != nil {
			return commentCreatedErrMsg{err: err}
		}
		return commentCreatedMsg{comment: comment}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case commentsLoadedMsg:
		m.selectedComments = msg.comments
		m.isLoading = false
		return m, nil
	case commentsLoadedErrMsg:
		m.SetError(fmt.Sprintf("Failed to load comments: %v", msg.err))
		m.isLoading = false
		return m, nil
	case issueCreatedMsg:
		m.showCreateIssueForm = false
		m.isLoading = false
		m.issueFormTitle = ""
		m.issueFormDesc = ""
		m.issueFormField = "title"
		return m, nil
	case issueCreatedErrMsg:
		m.SetError(fmt.Sprintf("Failed to create issue: %v", msg.err))
		m.showCreateIssueForm = false
		m.isLoading = false
		return m, nil
	case issueUpdatedMsg:
		m.showConfirmPopup = false
		m.confirmAction = ""
		m.confirmIssueIID = 0
		m.isLoading = false
		for i, item := range m.items {
			if item.ID == int(msg.issue.IID) {
				m.items[i].State = msg.issue.State
				m.items[i].UpdatedAt = *msg.issue.UpdatedAt
			}
		}
		if m.selectedIssue != nil && m.selectedIssue.IID == msg.issue.IID {
			m.selectedIssue.State = msg.issue.State
		}
		return m, nil
	case issueUpdatedErrMsg:
		m.SetError(fmt.Sprintf("Failed to update issue: %v", msg.err))
		m.showConfirmPopup = false
		m.confirmAction = ""
		m.confirmIssueIID = 0
		m.isLoading = false
		return m, nil
	case commentCreatedMsg:
		m.showCommentForm = false
		m.commentFormBody = ""
		m.isLoading = false
		if len(m.items) > 0 && m.selectedItem < len(m.items) && m.currentView == IssuesView {
			item := m.items[m.selectedItem]
			return m, loadCommentsCmd(m.projectPath, int64(item.ID), m.client)
		}
		return m, nil
	case commentCreatedErrMsg:
		m.SetError(fmt.Sprintf("Failed to create comment: %v", msg.err))
		m.showCommentForm = false
		m.commentFormBody = ""
		m.isLoading = false
		return m, nil
	case clipboardClearMsg:
		m.clipboardMessage = ""
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case spin.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		if m.showCreateIssueForm {
			switch msg.String() {
			case "esc":
				m.showCreateIssueForm = false
				m.issueFormTitle = ""
				m.issueFormDesc = ""
				m.issueFormField = "title"
			case "tab":
				if m.issueFormField == "title" {
					m.issueFormField = "description"
				} else {
					m.issueFormField = "title"
				}
			case "ctrl+enter":
				if m.issueFormTitle == "" {
					m.SetError("Title is required")
					return m, nil
				}
				m.isLoading = true
				return m, createIssueCmd(m.projectPath, m.issueFormTitle, m.issueFormDesc, m.client)
			case "enter":
				if m.issueFormField == "title" {
					m.issueFormField = "description"
				}
			case "ctrl+c":
				return m, tea.Quit
			}
			if len(msg.String()) == 1 && msg.String() != "c" {
				if m.issueFormField == "title" {
					if msg.Type == tea.KeyBackspace {
						if len(m.issueFormTitle) > 0 {
							m.issueFormTitle = m.issueFormTitle[:len(m.issueFormTitle)-1]
						}
					} else if msg.Type == tea.KeyRunes {
						m.issueFormTitle += string(msg.Runes)
					}
				} else {
					if msg.Type == tea.KeyBackspace {
						if len(m.issueFormDesc) > 0 {
							m.issueFormDesc = m.issueFormDesc[:len(m.issueFormDesc)-1]
						}
					} else if msg.Type == tea.KeyRunes {
						m.issueFormDesc += string(msg.Runes)
					}
				}
			}
			return m, nil
		}
		if m.showCommentForm {
			switch msg.String() {
			case "esc":
				m.showCommentForm = false
				m.commentFormBody = ""
			case "ctrl+enter":
				if m.commentFormBody == "" {
					m.SetError("Comment body is required")
					return m, nil
				}
				if len(m.items) > 0 && m.selectedItem < len(m.items) {
					item := m.items[m.selectedItem]
					m.isLoading = true
					return m, createCommentCmd(m.projectPath, int64(item.ID), m.commentFormBody, m.client)
				}
			case "ctrl+c":
				return m, tea.Quit
			}
			if msg.Type == tea.KeyBackspace {
				if len(m.commentFormBody) > 0 {
					m.commentFormBody = m.commentFormBody[:len(m.commentFormBody)-1]
				}
			} else if msg.Type == tea.KeyRunes {
				m.commentFormBody += string(msg.Runes)
			}
			return m, nil
		}
		if m.showError {
			switch msg.String() {
			case "r":
				m.showError = false
			case "esc", "q":
				m.showError = false
			}
			return m, nil
		}
		if m.showConfirmPopup {
			switch msg.String() {
			case "y":
				m.isLoading = true
				stateEvent := "close"
				if m.confirmAction == "reopen" {
					stateEvent = "reopen"
				}
				return m, updateIssueCmd(m.projectPath, m.confirmIssueIID, stateEvent, m.client)
			case "n", "esc":
				m.showConfirmPopup = false
				m.confirmAction = ""
				m.confirmIssueIID = 0
			}
			return m, nil
		}

		if m.showHelp {
			switch msg.String() {
			case "esc", "q", "?":
				m.showHelp = false
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "j", "down":
			if m.selectedItem < len(m.items)-1 {
				m.selectedItem++
			}
		case "k", "up":
			if m.selectedItem > 0 {
				m.selectedItem--
			}
		case "1":
			m.currentView = ProjectsView
			m.selectedItem = 0
		case "2":
			m.currentView = IssuesView
			m.selectedItem = 0
		case "3":
			m.currentView = MergeRequestsView
			m.selectedItem = 0
		case "h", "left":
			if m.currentView > ProjectsView {
				m.currentView--
				m.selectedItem = 0
			}
		case "l", "right":
			if m.currentView < MergeRequestsView {
				m.currentView++
				m.selectedItem = 0
			}
		case "tab":
			if m.currentView == IssuesView && len(m.items) > 0 && m.selectedItem < len(m.items) {
				m.detailSection = (m.detailSection + 1) % 2
			} else {
				if m.currentView < MergeRequestsView {
					m.currentView++
					m.selectedItem = 0
				} else {
					m.currentView = ProjectsView
					m.selectedItem = 0
				}
			}
		case "shift+tab":
			if m.currentView > ProjectsView {
				m.currentView--
				m.selectedItem = 0
			} else {
				m.currentView = MergeRequestsView
				m.selectedItem = 0
			}
		case "enter":
			if len(m.items) > 0 && m.selectedItem < len(m.items) {
				if m.currentView == IssuesView {
					m.detailSection = DescriptionSection
					m.selectedComments = []*gitlab.Note{}
					item := m.items[m.selectedItem]
					m.isLoading = true
					return m, loadCommentsCmd(m.projectPath, int64(item.ID), m.client)
				}
			}
		case "esc":
		case "?":
			m.showHelp = true
		case "f":
			if m.currentView == IssuesView {
				m.issueFilter = (m.issueFilter + 1) % 3
			}
		case "c":
			if m.currentView == IssuesView {
				m.showCreateIssueForm = true
				m.issueFormTitle = ""
				m.issueFormDesc = ""
				m.issueFormField = "title"
			}
		case "o":
			if m.currentView == IssuesView && len(m.items) > 0 && m.selectedItem < len(m.items) {
				item := m.items[m.selectedItem]
				action := "close"
				if item.State == "closed" {
					action = "reopen"
				}
				m.confirmAction = action
				m.confirmIssueIID = int64(item.ID)
				m.showConfirmPopup = true
			}
		case "y":
			if m.currentView == IssuesView && len(m.items) > 0 && m.selectedItem < len(m.items) {
				item := m.items[m.selectedItem]
				url := fmt.Sprintf("https://gitlab.com/%s/-/issues/%d", m.projectPath, item.ID)
				if err := utils.CopyToClipboard(url); err != nil {
					m.SetError(fmt.Sprintf("Failed to copy URL: %v", err))
				} else {
					m.clipboardMessage = "URL copied to clipboard"
					return m, tea.Tick(3*time.Second, func(time.Time) tea.Msg {
						return clipboardClearMsg{}
					})
				}
			}
		case "b":
			if m.currentView == IssuesView && len(m.items) > 0 && m.selectedItem < len(m.items) {
				item := m.items[m.selectedItem]
				url := fmt.Sprintf("https://gitlab.com/%s/-/issues/%d", m.projectPath, item.ID)
				if err := utils.OpenInBrowser(url); err != nil {
					m.SetError(fmt.Sprintf("Failed to open browser: %v", err))
				}
			}
		case "r":
			if m.currentView == IssuesView && len(m.items) > 0 && m.selectedItem < len(m.items) && m.detailSection == CommentsSection {
				m.showCommentForm = true
				m.commentFormBody = ""
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	if m.showCreateIssueForm {
		return m.renderCreateIssueForm()
	}

	if m.showCommentForm {
		return m.renderCommentForm()
	}

	if m.showConfirmPopup {
		return m.renderConfirmPopup()
	}

	if m.showError {
		return m.renderErrorPopup()
	}

	if m.showHelp {
		return m.renderHelpPopup()
	}

	sidebar := m.renderSidebar()
	mainPanel := m.renderMainPanel()
	detailsPanel := m.renderDetailsPanel()
	statusBar := m.renderStatusBar()

	return sidebar + "\n" + mainPanel + "\n" + detailsPanel + "\n" + statusBar
}

func (m Model) renderSidebar() string {
	var items []string

	items = append(items, m.styles.Title.Render("Navigation"))
	items = append(items, "")

	projectsStyle := m.styles.SidebarActive
	issuesStyle := m.styles.Sidebar
	mrStyle := m.styles.Sidebar

	switch m.currentView {
	case ProjectsView:
		projectsStyle = m.styles.SidebarActive
		issuesStyle = m.styles.Sidebar
		mrStyle = m.styles.Sidebar
	case IssuesView:
		projectsStyle = m.styles.Sidebar
		issuesStyle = m.styles.SidebarActive
		mrStyle = m.styles.Sidebar
	case MergeRequestsView:
		projectsStyle = m.styles.Sidebar
		issuesStyle = m.styles.Sidebar
		mrStyle = m.styles.SidebarActive
	}

	items = append(items, projectsStyle.Render("1. Projects"))
	items = append(items, issuesStyle.Render("2. Issues"))
	items = append(items, mrStyle.Render("3. Merge Requests"))
	items = append(items, "")
	items = append(items, m.styles.Sidebar.Render("Press q to quit"))

	sidebarContent := ""
	for _, item := range items {
		sidebarContent += item + "\n"
	}

	return m.styles.Sidebar.Width(m.styles.Sidebar.GetWidth()).Render(sidebarContent)
}

func (m Model) renderMainPanel() string {
	var title string
	var items []string

	switch m.currentView {
	case ProjectsView:
		title = "Projects"
	case IssuesView:
		title = "Issues"
	case MergeRequestsView:
		title = "Merge Requests"
	}

	if m.isLoading {
		loadingText := "  " + m.spinner.View() + " Loading..."
		content := m.styles.MainPanelHeader.Render(title) + "\n\n" + loadingText
		return m.styles.MainPanel.Render(content)
	}

	if len(m.items) == 0 {
		if m.currentView == IssuesView {
			items = append(items, "  No issues found")
		} else {
			items = append(items, "  No items to display")
		}
	} else {
		for i, item := range m.items {
			prefix := "  "
			if i == m.selectedItem {
				prefix = "> "
			}

			if m.currentView == IssuesView {
				displayText := fmt.Sprintf("#%d %s", item.ID, item.Title)
				stateIcon := "●"
				if item.State == "closed" {
					stateIcon = "○"
				}
				labelText := ""
				if len(item.Labels) > 0 {
					maxLabels := 3
					displayLabels := item.Labels
					if len(displayLabels) > maxLabels {
						displayLabels = displayLabels[:maxLabels]
						labelText = fmt.Sprintf(" [%s +%d]", strings.Join(displayLabels, ", "), len(item.Labels)-maxLabels)
					} else {
						labelText = fmt.Sprintf(" [%s]", strings.Join(displayLabels, ", "))
					}
				}
				items = append(items, prefix+stateIcon+" "+displayText+labelText)
			} else {
				items = append(items, prefix+item.Title)
			}
		}
	}

	content := m.styles.MainPanelHeader.Render(title) + "\n\n"
	for _, item := range items {
		content += item + "\n"
	}

	return m.styles.MainPanel.Render(content)
}

func (m Model) renderDetailsPanel() string {
	title := "Details"
	var content string

	if len(m.items) == 0 || m.selectedItem >= len(m.items) {
		content = "  No item selected"
	} else {
		item := m.items[m.selectedItem]
		if m.currentView == IssuesView {
			if m.detailSection == DescriptionSection {
				content = fmt.Sprintf("  #%d - %s\n\n", item.ID, item.Title)

				content += fmt.Sprintf("  State: %s\n", strings.Title(item.State))

				if item.Author != "" {
					content += fmt.Sprintf("  Author: %s\n", item.Author)
				}

				content += fmt.Sprintf("  Created: %s\n", item.CreatedAt.Format("2006-01-02 15:04:05"))
				content += fmt.Sprintf("  Updated: %s\n", item.UpdatedAt.Format("2006-01-02 15:04:05"))

				if len(item.Assignees) > 0 {
					content += fmt.Sprintf("  Assignees: %s\n", strings.Join(item.Assignees, ", "))
				}

				if item.Milestone != "" {
					content += fmt.Sprintf("  Milestone: %s\n", item.Milestone)
				}

				if len(item.Labels) > 0 {
					content += fmt.Sprintf("  Labels: %s\n", strings.Join(item.Labels, ", "))
				}

				if item.Desc != "" {
					content += "\n  Description:\n"
					maxDescLines := 8
					descLines := strings.Split(item.Desc, "\n")
					if len(descLines) > maxDescLines {
						for i := 0; i < maxDescLines; i++ {
							content += "  " + descLines[i] + "\n"
						}
						content += fmt.Sprintf("  ... (%d more lines)\n", len(descLines)-maxDescLines)
					} else {
						for _, line := range descLines {
							content += "  " + line + "\n"
						}
					}
				}
			} else {
				content = fmt.Sprintf("  #%d - %s\n\n", item.ID, item.Title)

				if m.isLoading {
					content += "  " + m.spinner.View() + " Loading comments..."
				} else if len(m.selectedComments) == 0 {
					content += "  No comments yet"
				} else {
					content += fmt.Sprintf("  %d comments\n\n", len(m.selectedComments))
					maxCommentLines := 8
					for i, comment := range m.selectedComments {
						if i >= maxCommentLines {
							content += fmt.Sprintf("  ... (%d more comments)\n", len(m.selectedComments)-maxCommentLines)
							break
						}
						author := "Unknown"
						if comment.Author.Username != "" {
							author = comment.Author.Username
						}
						date := ""
						if comment.CreatedAt != nil {
							date = comment.CreatedAt.Format("2006-01-02 15:04:05")
						}
						content += fmt.Sprintf("  %s @ %s:\n", author, date)
						bodyLines := strings.Split(comment.Body, "\n")
						for _, line := range bodyLines {
							content += "    " + line + "\n"
						}
						content += "\n"
					}
				}
			}
		} else {
			content = "  ID: " + fmt.Sprintf("%d", item.ID) + "\n"
			content += "  Title: " + item.Title
		}
	}

	panelContent := m.styles.DetailsPanelHeader.Render(title) + "\n\n" + content
	return m.styles.DetailsPanel.Render(panelContent)
}

func (m Model) renderStatusBar() string {
	projectInfo := "Project: " + m.projectPath
	connInfo := "Connection: " + m.connection
	helpInfo := "? for help"

	if m.clipboardMessage != "" {
		return m.styles.StatusBar.Render(m.clipboardMessage)
	}

	status := projectInfo + " | " + connInfo + " | " + helpInfo
	if m.currentView == IssuesView {
		status += " | Filter: " + m.issueFilter.String()
		if len(m.items) > 0 && m.selectedItem < len(m.items) {
			status += " | View: " + m.detailSection.String()
			if m.detailSection == CommentsSection {
				status += fmt.Sprintf(" (%d)", len(m.selectedComments))
			}
		}
	}

	return m.styles.StatusBar.Render(status)
}

func (m Model) renderHelpPopup() string {
	helpContent := `Keybindings

Navigation:
  j/k or ↑/↓    Move up/down in list
  h/l or ←/→    Switch to previous/next view
  Tab/Shift+Tab Cycle through views
  Enter          Select/open item
  Esc            Close popup / go back

View Switching:
  1              Projects view
  2              Issues view
  3              Merge Requests view`

	if m.currentView == IssuesView {
		helpContent += `

Issues:
  c              Create new issue
  f              Filter issues (All/Open/Closed)
  o              Toggle issue open/closed
  y              Copy issue URL to clipboard
  b              Open issue in browser
  r              Add comment (in Comments view)
  Tab            Toggle between Description/Comments`
	}

	helpContent += `

Other:
  q              Quit
  ?              Show this help

Press Esc or ? to close
`

	helpStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Background(lipgloss.Color("235")).
		Foreground(lipgloss.Color("255")).
		Padding(1, 2).
		Width(60).
		Align(lipgloss.Center)

	return helpStyle.Render(helpContent)
}

func (m Model) renderErrorPopup() string {
	errorContent := fmt.Sprintf("Error\n\n%s\n\nPress r to retry\nPress q or Esc to close", m.errorMessage)

	errorStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Background(lipgloss.Color("52")).
		Foreground(lipgloss.Color("255")).
		Padding(1, 2).
		Width(60).
		Align(lipgloss.Center)

	return errorStyle.Render(errorContent)
}

func (m *Model) SetLoading(loading bool) {
	m.isLoading = loading
}

func (m *Model) SetError(message string) {
	m.errorMessage = message
	m.showError = true
}

func (m *Model) ClearError() {
	m.errorMessage = ""
	m.showError = false
}

func (m *Model) SetItems(items []ListItem) {
	m.items = items
	m.selectedItem = 0
}

func IssuesToListItems(issues []*gitlab.Issue) []ListItem {
	items := make([]ListItem, len(issues))
	for i, issue := range issues {
		authorName := "Unknown"
		if issue.Author != nil {
			authorName = issue.Author.Username
		}

		labels := make([]string, len(issue.Labels))
		for j, label := range issue.Labels {
			labels[j] = label
		}

		assignees := make([]string, len(issue.Assignees))
		for j, assignee := range issue.Assignees {
			if assignee != nil {
				assignees[j] = assignee.Username
			}
		}

		milestoneName := ""
		if issue.Milestone != nil {
			milestoneName = issue.Milestone.Title
		}

		items[i] = ListItem{
			ID:        int(issue.IID),
			Title:     truncateString(issue.Title, 60),
			State:     issue.State,
			UpdatedAt: *issue.UpdatedAt,
			CreatedAt: *issue.CreatedAt,
			Author:    authorName,
			Desc:      issue.Description,
			Labels:    labels,
			Assignees: assignees,
			Milestone: milestoneName,
		}
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
	})
	return items
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func (m Model) renderCreateIssueForm() string {
	titleFieldStyle := m.styles.Sidebar
	descFieldStyle := m.styles.Sidebar
	if m.issueFormField == "title" {
		titleFieldStyle = m.styles.SidebarActive
	} else {
		descFieldStyle = m.styles.SidebarActive
	}

	titleInput := titleFieldStyle.Render("Title (required):") + "\n"
	if m.issueFormField == "title" {
		titleInput += "  " + m.issueFormTitle + "_"
	} else {
		titleInput += "  " + m.issueFormTitle
	}

	descInput := descFieldStyle.Render("Description (optional):") + "\n"
	maxDescLines := 5
	descLines := strings.Split(m.issueFormDesc, "\n")
	for i, line := range descLines {
		if i >= maxDescLines {
			break
		}
		if i == len(descLines)-1 && m.issueFormField == "description" && len(descLines) < maxDescLines {
			descInput += "  " + line + "_\n"
		} else {
			descInput += "  " + line + "\n"
		}
	}
	if m.issueFormField == "description" && len(descLines) >= maxDescLines {
		descInput += "  _\n"
	}

	formContent := "Create New Issue\n\n" +
		titleInput + "\n\n" +
		descInput + "\n" +
		"  Press Tab to switch fields\n" +
		"  Press Ctrl+Enter to submit\n" +
		"  Press Esc to cancel"

	formStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Background(lipgloss.Color("235")).
		Foreground(lipgloss.Color("255")).
		Padding(1, 2).
		Width(80).
		Align(lipgloss.Center)

	return formStyle.Render(formContent)
}

func (m Model) renderConfirmPopup() string {
	confirmText := fmt.Sprintf("Confirm %s\n\n", strings.Title(m.confirmAction))
	confirmText += fmt.Sprintf("Are you sure you want to %s issue #%d?\n\n", m.confirmAction, m.confirmIssueIID)
	confirmText += "Press y to confirm\n"
	confirmText += "Press n or Esc to cancel"

	confirmStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Background(lipgloss.Color("235")).
		Foreground(lipgloss.Color("255")).
		Padding(1, 2).
		Width(50).
		Align(lipgloss.Center)

	return confirmStyle.Render(confirmText)
}

func (m Model) renderCommentForm() string {
	maxBodyLines := 10
	bodyLines := strings.Split(m.commentFormBody, "\n")

	bodyInput := ""
	for i, line := range bodyLines {
		if i >= maxBodyLines {
			break
		}
		if i == len(bodyLines)-1 {
			bodyInput += "  " + line + "_\n"
		} else {
			bodyInput += "  " + line + "\n"
		}
	}
	if len(bodyLines) >= maxBodyLines {
		bodyInput += "  _\n"
	}

	formContent := "Add Comment\n\n" +
		"Body:\n\n" +
		bodyInput + "\n" +
		"  Press Ctrl+Enter to submit\n" +
		"  Press Esc to cancel"

	formStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Background(lipgloss.Color("235")).
		Foreground(lipgloss.Color("255")).
		Padding(1, 2).
		Width(80).
		Align(lipgloss.Center)

	return formStyle.Render(formContent)
}
