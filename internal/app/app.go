package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

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

	interactive := isInteractiveSession(os.Stdin, os.Stdout)

	if cfg.NeedsSetup() {
		if !interactive {
			return errors.New("first-run setup requires an interactive terminal; configure host and token manually")
		}
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

	selected := config.Instance{Host: cfg.Host, Token: cfg.Token}
	if projectPath == "" {
		if !interactive {
			projectPath = strings.TrimSpace(cfg.LastProject)
			if projectPath == "" {
				return errors.New("no project detected in non-interactive mode; pass --project or set last project in config")
			}
			logger.Printf("using last known project: %s", projectPath)
		} else {
			instances, err := config.LoadInstances()
			if err != nil {
				return fmt.Errorf("load configured instances: %w", err)
			}
			if len(instances) == 0 {
				instances = append(instances, selected)
			}

			instanceByHost := make(map[string]config.Instance, len(instances))
			options := make([]tui.InstanceOption, 0, len(instances))
			for _, instance := range instances {
				host := strings.TrimSpace(instance.Host)
				token := strings.TrimSpace(instance.Token)
				if host == "" || token == "" {
					continue
				}
				instanceByHost[strings.ToLower(host)] = instance
				options = append(options, tui.InstanceOption{Host: host, Label: formatInstanceLabel(host)})
			}

			if len(options) == 0 {
				return errors.New("no usable GitLab instance configuration found")
			}

			startupChoice, startupErr := tui.RunStartupContextFlow(tui.StartupContextFlowOptions{
				LastProject: cfg.LastProject,
				Instances:   options,
				LoadProjects: func(instanceOption tui.InstanceOption) ([]tui.StartupProjectOption, error) {
					instance, ok := instanceByHost[strings.ToLower(strings.TrimSpace(instanceOption.Host))]
					if !ok {
						return nil, fmt.Errorf("selected instance %q is unavailable", strings.TrimSpace(instanceOption.Host))
					}

					pickerClient, clientErr := gitlab.NewClient(instance.Token, instance.Host, logger)
					if clientErr != nil {
						return nil, clientErr
					}

					listCtx, cancelList := context.WithTimeout(ctx, 20*time.Second)
					defer cancelList()

					projects, listErr := pickerClient.ListProjects(listCtx, "")
					if listErr != nil {
						return nil, fmt.Errorf("load projects for picker (timeout 20s): %w", listErr)
					}

					projectOptions := make([]tui.StartupProjectOption, 0, len(projects))
					for _, project := range projects {
						if project == nil {
							continue
						}
						path := strings.TrimSpace(project.PathWithNamespace)
						if path == "" {
							continue
						}
						projectOptions = append(projectOptions, tui.StartupProjectOption{Path: path, Label: path})
					}

					return projectOptions, nil
				},
			})
			if startupErr != nil {
				if errors.Is(startupErr, tui.ErrCancelled) {
					return errors.New("no project selected")
				}
				return fmt.Errorf("startup context failed: %w", startupErr)
			}

			switch startupChoice.Action {
			case tui.StartupActionUseLastProject:
				projectPath = strings.TrimSpace(startupChoice.ProjectPath)
				if projectPath == "" {
					return errors.New("no project selected")
				}
				logger.Printf("using last known project: %s", projectPath)
			case tui.StartupActionSelectContext:
				instance, ok := instanceByHost[strings.ToLower(strings.TrimSpace(startupChoice.Instance.Host))]
				if !ok {
					return errors.New("no instance selected")
				}
				selected = instance
				projectPath = strings.TrimSpace(startupChoice.ProjectPath)
				if projectPath == "" {
					return errors.New("no project selected")
				}
			default:
				return errors.New("no project selected")
			}
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

	if strings.TrimSpace(projectPath) == "" {
		return errors.New("no project selected")
	}

	cfg.LastProject = projectPath
	if err := config.Save(cfg); err != nil {
		logger.Printf("failed to persist last project: %v", err)
	}

	provider := NewProvider(client, projectPath)
	if !interactive {
		renderNonInteractiveSummary(os.Stdout, cfg.Host, projectPath, user.Username)
		return nil
	}

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

func isInteractiveSession(stdin *os.File, stdout *os.File) bool {
	if stdin == nil || stdout == nil {
		return false
	}
	return term.IsTerminal(int(stdin.Fd())) && term.IsTerminal(int(stdout.Fd()))
}

func renderNonInteractiveSummary(w io.Writer, host string, projectPath string, username string) {
	fmt.Fprintf(w, "lazygitlab non-interactive summary\n")
	fmt.Fprintf(w, "host: %s\n", strings.TrimSpace(host))
	fmt.Fprintf(w, "project: %s\n", strings.TrimSpace(projectPath))
	fmt.Fprintf(w, "user: %s\n", strings.TrimSpace(username))
	fmt.Fprintf(w, "tip: run in an interactive terminal to open the TUI\n")
}
