package gui

import "github.com/charmbracelet/lipgloss"

type Style struct {
	Sidebar            lipgloss.Style
	SidebarActive      lipgloss.Style
	MainPanel          lipgloss.Style
	MainPanelHeader    lipgloss.Style
	DetailsPanel       lipgloss.Style
	DetailsPanelHeader lipgloss.Style
	StatusBar          lipgloss.Style
	Title              lipgloss.Style
}

func NewStyle() *Style {
	s := &Style{}

	s.Sidebar = lipgloss.NewStyle().
		Width(20).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	s.SidebarActive = lipgloss.NewStyle().
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("62")).
		Padding(0, 1)

	s.MainPanel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Width(60).
		Height(15)

	s.MainPanelHeader = lipgloss.NewStyle().
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("62")).
		Padding(0, 1).
		Width(58).
		Align(lipgloss.Left)

	s.DetailsPanel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Width(80).
		Height(10)

	s.DetailsPanelHeader = lipgloss.NewStyle().
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("62")).
		Padding(0, 1).
		Width(78).
		Align(lipgloss.Left)

	s.StatusBar = lipgloss.NewStyle().
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("63")).
		Padding(0, 1).
		Width(80).
		Align(lipgloss.Left)

	s.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("230")).
		Background(lipgloss.Color("62")).
		Padding(0, 1).
		Bold(true)

	return s
}
