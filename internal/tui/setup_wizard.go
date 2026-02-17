package tui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/davzucky/lazygitlab/internal/config"
)

var ErrCancelled = errors.New("cancelled")

type SetupResult struct {
	Host  string
	Token string
}

type setupModel struct {
	hostInput  textinput.Model
	tokenInput textinput.Model
	focus      int
	err        string
	done       bool
	cancelled  bool
	result     SetupResult
}

func RunSetupWizard(currentHost string) (SetupResult, error) {
	host := strings.TrimSpace(currentHost)
	if host == "" {
		host = "https://gitlab.com"
	}

	m := newSetupModel(host)
	p := tea.NewProgram(m)
	out, err := p.Run()
	if err != nil {
		return SetupResult{}, err
	}

	final := out.(setupModel)
	if final.cancelled {
		return SetupResult{}, ErrCancelled
	}

	return final.result, nil
}

func newSetupModel(host string) setupModel {
	hostInput := textinput.New()
	hostInput.Prompt = "Host URL: "
	hostInput.SetValue(host)
	hostInput.CharLimit = 256
	hostInput.Width = 50
	hostInput.Focus()

	tokenInput := textinput.New()
	tokenInput.Prompt = "Token: "
	tokenInput.EchoMode = textinput.EchoPassword
	tokenInput.EchoCharacter = '•'
	tokenInput.CharLimit = 512
	tokenInput.Width = 50

	return setupModel{hostInput: hostInput, tokenInput: tokenInput}
}

func (m setupModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m setupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit
		case "tab", "shift+tab", "up", "down":
			if msg.String() == "shift+tab" || msg.String() == "up" {
				m.focus--
			} else {
				m.focus++
			}
			if m.focus > 1 {
				m.focus = 0
			}
			if m.focus < 0 {
				m.focus = 1
			}

			if m.focus == 0 {
				m.hostInput.Focus()
				m.tokenInput.Blur()
			} else {
				m.tokenInput.Focus()
				m.hostInput.Blur()
			}

			return m, nil
		case "enter":
			if m.focus == 0 {
				m.focus = 1
				m.tokenInput.Focus()
				m.hostInput.Blur()
				return m, nil
			}

			host, err := config.NormalizeHost(m.hostInput.Value())
			if err != nil {
				m.err = fmt.Sprintf("Invalid host: %v", err)
				return m, nil
			}
			if strings.TrimSpace(m.tokenInput.Value()) == "" {
				m.err = "Token is required"
				return m, nil
			}

			m.result = SetupResult{
				Host:  host,
				Token: strings.TrimSpace(m.tokenInput.Value()),
			}
			m.done = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	if m.focus == 0 {
		m.hostInput, cmd = m.hostInput.Update(msg)
	} else {
		m.tokenInput, cmd = m.tokenInput.Update(msg)
	}

	return m, cmd
}

func (m setupModel) View() string {
	s := newStyles()
	header := s.title.Render("LazyGitLab First-Run Setup")
	content := []string{
		header,
		"",
		"No valid configuration was found.",
		"Enter your GitLab host and personal access token.",
		"",
		m.hostInput.View(),
		m.tokenInput.View(),
		"",
		"Tab to switch fields, Enter to submit, Esc to cancel.",
	}

	if m.err != "" {
		content = append(content, "", s.errorText.Render(m.err))
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(resolveAccentColor()).
		Padding(1, 2).
		Width(72)

	return box.Render(strings.Join(content, "\n"))
}
