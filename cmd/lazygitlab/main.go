package main

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct{}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	return "LazyGitLab\n\nPress q to quit"
}

func main() {
	p := tea.NewProgram(model{})
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
