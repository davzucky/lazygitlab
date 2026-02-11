package app

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
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
	if os.Getenv("LAZYGITLAB_MOCK_DATA") == "1" {
		provider := NewMockProvider()
		model := tui.NewDashboardModel(provider, tui.DashboardContext{
			ProjectPath: "mock/group/project",
			Connection:  "Connected as mock-user",
			Host:        "https://mock.gitlab.local/api/v4",
		})

		program := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
		if _, err := program.Run(); err != nil {
			return fmt.Errorf("run mock TUI: %w", err)
		}
		return nil
	}

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

	selected := config.Instance{Host: cfg.Host, Token: cfg.Token}
	if projectPath == "" {
		instances, err := config.LoadInstances()
		if err != nil {
			return fmt.Errorf("load configured instances: %w", err)
		}

		if len(instances) == 0 {
			instances = append(instances, selected)
		}

		if len(instances) > 1 {
			options := make([]tui.InstanceOption, 0, len(instances))
			for _, instance := range instances {
				options = append(options, tui.InstanceOption{
					Host:  instance.Host,
					Label: formatInstanceLabel(instance.Host),
				})
			}

			chosen, pickErr := tui.RunInstancePicker(options)
			if pickErr != nil {
				if errors.Is(pickErr, tui.ErrCancelled) {
					return errors.New("no instance selected")
				}
				return fmt.Errorf("instance picker failed: %w", pickErr)
			}

			for _, instance := range instances {
				if strings.EqualFold(instance.Host, chosen.Host) {
					selected = instance
					break
				}
			}
		} else {
			selected = instances[0]
		}
	}

	if strings.TrimSpace(selected.Host) == "" || strings.TrimSpace(selected.Token) == "" {
		return errors.New("no usable GitLab instance configuration found")
	}

	cfg.Host = selected.Host
	cfg.Token = selected.Token

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

func formatInstanceLabel(host string) string {
	normalized := strings.TrimSpace(host)
	if normalized == "" {
		return host
	}

	u, err := url.Parse(normalized)
	if err != nil {
		return normalized
	}

	if u.Host == "" {
		return normalized
	}

	if strings.TrimSpace(u.Path) == "" {
		return u.Host
	}

	return fmt.Sprintf("%s (%s)", u.Host, strings.TrimSuffix(u.Path, "/"))
}
