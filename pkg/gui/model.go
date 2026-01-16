package gui

import (
	"fmt"

	"github.com/charmbracelet/bubbletea"
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
		}
	}

	return m, nil
}

func (m Model) View() string {
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
