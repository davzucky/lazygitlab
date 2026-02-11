package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gl "gitlab.com/gitlab-org/api/client-go"
)

type pickerModel struct {
	input     textinput.Model
	projects  []*gl.Project
	filtered  []*gl.Project
	selected  int
	chosen    string
	cancelled bool
}

func RunProjectPicker(projects []*gl.Project) (string, error) {
	m := newPickerModel(projects)
	p := tea.NewProgram(m)
	out, err := p.Run()
	if err != nil {
		return "", err
	}

	final := out.(pickerModel)
	if final.cancelled {
		return "", ErrCancelled
	}
	if final.chosen == "" {
		return "", fmt.Errorf("no project selected")
	}

	return final.chosen, nil
}

func newPickerModel(projects []*gl.Project) pickerModel {
	input := textinput.New()
	input.Prompt = "Search: "
	input.Focus()
	input.Width = 50

	copyProjects := make([]*gl.Project, 0, len(projects))
	for _, p := range projects {
		if p == nil {
			continue
		}
		copyProjects = append(copyProjects, p)
	}

	return pickerModel{
		input:    input,
		projects: copyProjects,
		filtered: copyProjects,
	}
}

func (m pickerModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.cancelled = true
			return m, tea.Quit
		case "down", "j":
			if m.selected < len(m.filtered)-1 {
				m.selected++
			}
			return m, nil
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
			return m, nil
		case "enter":
			if len(m.filtered) == 0 {
				return m, nil
			}
			m.chosen = m.filtered[m.selected].PathWithNamespace
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.applyFilter()

	return m, cmd
}

func (m *pickerModel) applyFilter() {
	q := strings.ToLower(strings.TrimSpace(m.input.Value()))
	if q == "" {
		m.filtered = m.projects
		if m.selected >= len(m.filtered) {
			m.selected = 0
		}
		return
	}

	matches := make([]*gl.Project, 0)
	for _, p := range m.projects {
		if strings.Contains(strings.ToLower(p.PathWithNamespace), q) || strings.Contains(strings.ToLower(p.Name), q) {
			matches = append(matches, p)
		}
	}

	m.filtered = matches
	if m.selected >= len(m.filtered) {
		m.selected = 0
	}
}

func (m pickerModel) View() string {
	rows := []string{
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("Select a GitLab project"),
		"",
		m.input.View(),
		"",
	}

	if len(m.filtered) == 0 {
		rows = append(rows, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("No matching projects"))
	} else {
		limit := min(12, len(m.filtered))
		for i := 0; i < limit; i++ {
			p := m.filtered[i]
			prefix := "  "
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
			if i == m.selected {
				prefix = "› "
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
			}
			rows = append(rows, style.Render(prefix+p.PathWithNamespace))
		}
	}

	rows = append(rows, "", "Enter to select, j/k to move, q or Esc to cancel")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Width(80)

	return box.Render(strings.Join(rows, "\n"))
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
