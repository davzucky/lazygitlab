package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davzucky/lazygitlab/pkg/config"
	"github.com/davzucky/lazygitlab/pkg/gui"
)

type errorModel struct {
	error string
}

func (m errorModel) Init() tea.Cmd {
	return nil
}

func (m errorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m errorModel) View() string {
	return fmt.Sprintf("Error: %s\n\nPress q to quit", m.error)
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		p := tea.NewProgram(errorModel{error: err.Error()})
		if _, err := p.Run(); err != nil {
			os.Exit(1)
		}
		return
	}

	if err := cfg.Validate(); err != nil {
		p := tea.NewProgram(errorModel{error: fmt.Sprintf("Invalid GitLab token: %v", err)})
		if _, err := p.Run(); err != nil {
			os.Exit(1)
		}
		return
	}

	projectPath := "Not detected"
	connection := "Connected to " + cfg.Host

	model := gui.NewModel(projectPath, connection)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
