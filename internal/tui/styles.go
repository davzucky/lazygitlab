package tui

import "github.com/charmbracelet/lipgloss"

type styles struct {
	app            lipgloss.Style
	sidebar        lipgloss.Style
	sidebarActive  lipgloss.Style
	panel          lipgloss.Style
	header         lipgloss.Style
	issueID        lipgloss.Style
	status         lipgloss.Style
	helpPopup      lipgloss.Style
	errorPopup     lipgloss.Style
	selectedRow    lipgloss.Style
	normalRow      lipgloss.Style
	secondary      lipgloss.Style
	title          lipgloss.Style
	dim            lipgloss.Style
	topLevelBorder lipgloss.Border
}

func newStyles() styles {
	accent := lipgloss.Color("39")
	muted := lipgloss.Color("245")
	bg := lipgloss.Color("236")

	return styles{
		app: lipgloss.NewStyle().Padding(0, 1),
		sidebar: lipgloss.NewStyle().
			Padding(1, 1).
			Foreground(lipgloss.Color("252")),
		sidebarActive: lipgloss.NewStyle().
			PaddingLeft(0).
			Foreground(accent).
			Bold(true),
		panel: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 1),
		header: lipgloss.NewStyle().
			Foreground(accent).
			Bold(true),
		issueID: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true),
		status: lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(accent).
			Padding(0, 1),
		helpPopup: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(accent).
			Background(bg).
			Padding(1, 2).
			Width(70),
		errorPopup: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196")).
			Background(lipgloss.Color("52")).
			Foreground(lipgloss.Color("255")).
			Padding(1, 2).
			Width(70),
		selectedRow: lipgloss.NewStyle().
			Foreground(accent).
			Bold(true),
		normalRow: lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		secondary: lipgloss.NewStyle().Foreground(muted),
		title:     lipgloss.NewStyle().Bold(true).Foreground(accent),
		dim:       lipgloss.NewStyle().Foreground(muted),
	}
}
