package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davzucky/lazygitlab/pkg/config"
	"github.com/davzucky/lazygitlab/pkg/gitlab"
	"github.com/davzucky/lazygitlab/pkg/gui"
	"github.com/davzucky/lazygitlab/pkg/project"
	"github.com/davzucky/lazygitlab/pkg/utils"
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
	projectFlag := flag.String("project", "", "Manually specify GitLab project path (e.g., group/project)")
	debugFlag := flag.Bool("debug", false, "Enable verbose debugging")
	flag.Parse()

	if err := utils.InitLogger(*debugFlag); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer utils.Close()

	cfg, err := config.Load()
	if err != nil {
		utils.Error("Failed to load config: %v", err)
		p := tea.NewProgram(errorModel{error: err.Error()})
		if _, runErr := p.Run(); runErr != nil {
			utils.Error("Failed to run error UI: %v", runErr)
		}
		os.Exit(1)
	}

	utils.Debug("Loaded config: host=%s, hasToken=%v", cfg.Host, cfg.Token != "")

	if err := cfg.Validate(); err != nil {
		utils.Error("Invalid config: %v", err)
		p := tea.NewProgram(errorModel{error: fmt.Sprintf("Invalid GitLab token: %v", err)})
		if _, runErr := p.Run(); runErr != nil {
			utils.Error("Failed to run error UI: %v", runErr)
		}
		os.Exit(1)
	}

	utils.Info("Configuration validated successfully")

	projectPath, err := project.DetectProjectPath(*projectFlag, cfg.Host)
	if err != nil {
		utils.Error("Failed to detect project path: %v", err)
		projectPath = fmt.Sprintf("Detection failed: %v", err)
	} else {
		utils.Info("Detected project path: %s", projectPath)
	}

	connection := "Connected to " + cfg.Host

	glClient, err := gitlab.NewClient(cfg.Token, cfg.Host)
	if err != nil {
		utils.Error("Failed to create GitLab client: %v", err)
		p := tea.NewProgram(errorModel{error: fmt.Sprintf("Failed to create GitLab client: %v", err)})
		if _, runErr := p.Run(); runErr != nil {
			utils.Error("Failed to run error UI: %v", runErr)
		}
		os.Exit(1)
	}
	defer glClient.Close()

	model := gui.NewModel(projectPath, cfg.Host, connection, glClient)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		os.Exit(1)
	}
}
