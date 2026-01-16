package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davzucky/lazygitlab/pkg/config"
)

type model struct {
	initialized bool
	error       string
}

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
	if m.error != "" {
		return fmt.Sprintf("Error: %s\n\nPress q to quit", m.error)
	}
	return "LazyGitLab\n\nPress q to quit"
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		p := tea.NewProgram(model{error: err.Error()})
		if _, err := p.Run(); err != nil {
			os.Exit(1)
		}
		return
	}

	if err := cfg.Validate(); err != nil {
		p := tea.NewProgram(model{error: fmt.Sprintf("Invalid GitLab token: %v", err)})
		if _, err := p.Run(); err != nil {
			os.Exit(1)
		}
		return
	}

	p := tea.NewProgram(model{initialized: true})
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
