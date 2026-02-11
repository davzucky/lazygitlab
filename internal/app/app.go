package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/davzucky/lazygitlab/internal/config"
	"github.com/davzucky/lazygitlab/internal/gitlab"
	"github.com/davzucky/lazygitlab/internal/logging"
	"github.com/davzucky/lazygitlab/internal/project"
	"github.com/davzucky/lazygitlab/internal/tui"
)

type Options struct {
	ProjectOverride string
	Debug           bool
}

func Run(ctx context.Context, opts Options) error {
	logger, closeLogger, err := logging.New(opts.Debug)
	if err != nil {
		return fmt.Errorf("initialize logger: %w", err)
	}
	defer closeLogger()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	if opts.Debug {
		cfg.Debug = true
	}

	if cfg.NeedsSetup() {
		setupResult, setupErr := tui.RunSetupWizard(cfg.Host)
		if setupErr != nil {
			return fmt.Errorf("first-run setup failed: %w", setupErr)
		}

		cfg.Host = setupResult.Host
		cfg.Token = setupResult.Token

		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("save configuration: %w", err)
		}
	}

	client, err := gitlab.NewClient(cfg.Token, cfg.Host, logger)
	if err != nil {
		return err
	}

	authCtx, cancelAuth := context.WithTimeout(ctx, 12*time.Second)
	defer cancelAuth()

	user, err := client.GetCurrentUser(authCtx)
	if err != nil {
		return fmt.Errorf("validate token: %w", err)
	}

	projectPath := strings.TrimSpace(opts.ProjectOverride)
	if projectPath == "" {
		projectPath, err = project.DetectCurrentProject(cfg.Host)
		if err != nil {
			logger.Printf("project autodetect failed: %v", err)
		}
	}

	if projectPath == "" {
		projectPath = strings.TrimSpace(cfg.LastProject)
		if projectPath != "" {
			logger.Printf("using last known project: %s", projectPath)
		}
	}

	if projectPath == "" {
		listCtx, cancelList := context.WithTimeout(ctx, 20*time.Second)
		defer cancelList()

		projects, listErr := client.ListProjects(listCtx, "")
		if listErr != nil {
			return fmt.Errorf("load projects for picker (timeout 20s): %w", listErr)
		}

		projectPath, err = tui.RunProjectPicker(projects)
		if err != nil {
			if errors.Is(err, tui.ErrCancelled) {
				return errors.New("no project selected")
			}
			return fmt.Errorf("project picker failed: %w", err)
		}
	}

	cfg.LastProject = projectPath
	if err := config.Save(cfg); err != nil {
		logger.Printf("failed to persist last project: %v", err)
	}

	provider := NewProvider(client, projectPath)
	model := tui.NewDashboardModel(provider, tui.DashboardContext{
		ProjectPath: projectPath,
		Connection:  fmt.Sprintf("Connected as %s", user.Username),
		Host:        cfg.Host,
	})

	program := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("run TUI: %w", err)
	}

	return nil
}
