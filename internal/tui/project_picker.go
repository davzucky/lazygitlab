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

type InstanceOption struct {
	Host  string
	Label string
}

type instancePickerModel struct {
	instances []InstanceOption
	selected  int
	chosen    InstanceOption
	cancelled bool
}

func RunInstancePicker(instances []InstanceOption) (InstanceOption, error) {
	m := newInstancePickerModel(instances)
	p := tea.NewProgram(m)
	out, err := p.Run()
	if err != nil {
		return InstanceOption{}, err
	}

	final := out.(instancePickerModel)
	if final.cancelled {
		return InstanceOption{}, ErrCancelled
	}
	if final.chosen.Host == "" {
		return InstanceOption{}, fmt.Errorf("no instance selected")
	}

	return final.chosen, nil
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
		case "down", "j", "tab":
			if m.selected < len(m.filtered)-1 {
				m.selected++
			}
			return m, nil
		case "up", "k", "shift+tab":
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

func newInstancePickerModel(instances []InstanceOption) instancePickerModel {
	copyInstances := make([]InstanceOption, 0, len(instances))
	for _, instance := range instances {
		if strings.TrimSpace(instance.Host) == "" {
			continue
		}
		copyInstances = append(copyInstances, instance)
	}

	return instancePickerModel{instances: copyInstances}
}

func (m instancePickerModel) Init() tea.Cmd {
	return nil
}

func (m instancePickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.cancelled = true
			return m, tea.Quit
		case "down", "j", "tab":
			if m.selected < len(m.instances)-1 {
				m.selected++
			}
			return m, nil
		case "up", "k", "shift+tab":
			if m.selected > 0 {
				m.selected--
			}
			return m, nil
		case "enter":
			if len(m.instances) == 0 {
				return m, nil
			}
			m.chosen = m.instances[m.selected]
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m instancePickerModel) View() string {
	s := newStyles()
	rows := []string{
		s.title.Render("Select a GitLab instance"),
		"",
	}

	if len(m.instances) == 0 {
		rows = append(rows, s.dim.Render("No configured instances found"))
	} else {
		for i, instance := range m.instances {
			prefix := "  "
			style := s.normalRow
			if i == m.selected {
				prefix = "› "
				style = s.selectedRow
			}

			label := instance.Label
			if strings.TrimSpace(label) == "" {
				label = instance.Host
			}

			rows = append(rows, style.Render(prefix+label))
		}
	}

	rows = append(rows, "", "Enter to select, j/k or arrows to move, Tab/Shift+Tab to cycle, q or Esc to cancel")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(resolveAccentColor()).
		Padding(1, 2).
		Width(80)

	return box.Render(strings.Join(rows, "\n"))
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
	s := newStyles()
	rows := []string{
		s.title.Render("Select a GitLab project"),
		"",
		m.input.View(),
		"",
	}

	if len(m.filtered) == 0 {
		rows = append(rows, s.dim.Render("No matching projects"))
	} else {
		limit := min(12, len(m.filtered))
		for i := 0; i < limit; i++ {
			p := m.filtered[i]
			prefix := "  "
			style := s.normalRow
			if i == m.selected {
				prefix = "› "
				style = s.selectedRow
			}
			rows = append(rows, style.Render(prefix+p.PathWithNamespace))
		}
	}

	rows = append(rows, "", "Enter to select, j/k or arrows to move, Tab/Shift+Tab to cycle, q or Esc to cancel")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(resolveAccentColor()).
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
