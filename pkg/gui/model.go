package gui

import (
	"fmt"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ViewMode int

const (
	ProjectsView ViewMode = iota
	IssuesView
	MergeRequestsView
)

type Model struct {
	currentView  ViewMode
	items        []ListItem
	selectedItem int
	projectPath  string
	connection   string
	width        int
	height       int
	styles       *Style
	showHelp     bool
}

type ListItem struct {
	ID    int
	Title string
}

func NewModel(projectPath string, connection string) Model {
	styles := NewStyle()

	return Model{
		currentView:  ProjectsView,
		items:        []ListItem{},
		selectedItem: 0,
		projectPath:  projectPath,
		connection:   connection,
		width:        80,
		height:       24,
		styles:       styles,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
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
			}
		case "esc":
		case "?":
			m.showHelp = true
		}
	}

	return m, nil
}

func (m Model) View() string {
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

	if len(m.items) == 0 {
		items = append(items, "  No items to display")
	} else {
		for i, item := range m.items {
			prefix := "  "
			if i == m.selectedItem {
				prefix = "> "
			}
			items = append(items, prefix+item.Title)
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
		content = "  ID: " + fmt.Sprintf("%d", item.ID) + "\n"
		content += "  Title: " + item.Title
	}

	panelContent := m.styles.DetailsPanelHeader.Render(title) + "\n\n" + content
	return m.styles.DetailsPanel.Render(panelContent)
}

func (m Model) renderStatusBar() string {
	projectInfo := "Project: " + m.projectPath
	connInfo := "Connection: " + m.connection
	helpInfo := "? for help"

	return m.styles.StatusBar.Render(projectInfo + " | " + connInfo + " | " + helpInfo)
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
