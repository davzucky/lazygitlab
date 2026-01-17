package gui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	spin "github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

type Model struct {
	currentView   ViewMode
	items         []ListItem
	selectedItem  int
	projectPath   string
	connection    string
	width         int
	height        int
	styles        *Style
	showHelp      bool
	showError     bool
	errorMessage  string
	isLoading     bool
	spinner       spin.Model
	selectedIssue *gitlab.Issue
	issueFilter   IssueFilterState
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

func NewModel(projectPath string, connection string) Model {
	styles := NewStyle()
	spinner := spin.New()
	spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("62"))

	return Model{
		currentView:  ProjectsView,
		items:        []ListItem{},
		selectedItem: 0,
		projectPath:  projectPath,
		connection:   connection,
		width:        80,
		height:       24,
		styles:       styles,
		showError:    false,
		errorMessage: "",
		isLoading:    false,
		spinner:      spinner,
		issueFilter:  FilterAll,
	}
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case spin.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		if m.showError {
			switch msg.String() {
			case "r":
				m.showError = false
			case "esc", "q":
				m.showError = false
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
			if m.currentView < MergeRequestsView {
				m.currentView++
				m.selectedItem = 0
			} else {
				m.currentView = ProjectsView
				m.selectedItem = 0
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
					m.selectedIssue = nil
				}
			}
		case "esc":
		case "?":
			m.showHelp = true
		case "f":
			if m.currentView == IssuesView {
				m.issueFilter = (m.issueFilter + 1) % 3
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
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
				items = append(items, prefix+stateIcon+" "+displayText)
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

	status := projectInfo + " | " + connInfo + " | " + helpInfo
	if m.currentView == IssuesView {
		status += " | Filter: " + m.issueFilter.String()
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
  3              Merge Requests view

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
