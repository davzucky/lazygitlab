package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type StartupAction string

const (
	StartupActionSelectContext  StartupAction = "select_context"
	StartupActionUseLastProject StartupAction = "use_last_project"
)

type StartupContextChoice struct {
	Action StartupAction
}

type startupOption struct {
	action StartupAction
	label  string
	hint   string
}

type startupContextModel struct {
	options   []startupOption
	selected  int
	choice    StartupContextChoice
	cancelled bool
}

func RunStartupContextScreen(lastProject string) (StartupContextChoice, error) {
	m := newStartupContextModel(lastProject)
	p := tea.NewProgram(m)
	out, err := p.Run()
	if err != nil {
		return StartupContextChoice{}, err
	}

	final := out.(startupContextModel)
	if final.cancelled {
		return StartupContextChoice{}, ErrCancelled
	}
	if final.choice.Action == "" {
		return StartupContextChoice{}, fmt.Errorf("no startup action selected")
	}

	return final.choice, nil
}

func newStartupContextModel(lastProject string) startupContextModel {
	options := []startupOption{{
		action: StartupActionSelectContext,
		label:  "Select GitLab server and project",
		hint:   "Choose an instance, then choose a project",
	}}

	trimmedLastProject := strings.TrimSpace(lastProject)
	if trimmedLastProject != "" {
		options = append(options, startupOption{
			action: StartupActionUseLastProject,
			label:  fmt.Sprintf("Use last project: %s", trimmedLastProject),
			hint:   "Skip pickers and continue immediately",
		})
	}

	return startupContextModel{options: options}
}

func (m startupContextModel) Init() tea.Cmd {
	return nil
}

func (m startupContextModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.cancelled = true
			return m, tea.Quit
		case "down", "j", "tab":
			if m.selected < len(m.options)-1 {
				m.selected++
			}
			return m, nil
		case "up", "k", "shift+tab":
			if m.selected > 0 {
				m.selected--
			}
			return m, nil
		case "enter":
			if len(m.options) == 0 {
				return m, nil
			}
			m.choice.Action = m.options[m.selected].action
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m startupContextModel) View() string {
	rows := []string{
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("Select startup context"),
		"",
		"No GitLab repository context was detected for the current directory.",
		"Choose how to continue:",
		"",
	}

	for i, option := range m.options {
		prefix := "  "
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
		if i == m.selected {
			prefix = "> "
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		}

		rows = append(rows, style.Render(prefix+option.label))
		if strings.TrimSpace(option.hint) != "" {
			rows = append(rows, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("  "+option.hint))
		}
		rows = append(rows, "")
	}

	rows = append(rows, "Enter to select, j/k or arrows to move, Tab/Shift+Tab to cycle, q or Esc to cancel")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Width(84)

	return box.Render(strings.Join(rows, "\n"))
}
