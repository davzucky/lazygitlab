package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestStartupContextModelWithoutLastProjectShowsSingleOption(t *testing.T) {
	t.Parallel()

	m := newStartupContextModel("")
	if len(m.options) != 1 {
		t.Fatalf("option count = %d want %d", len(m.options), 1)
	}
	if m.options[0].action != StartupActionSelectContext {
		t.Fatalf("first action = %v want %v", m.options[0].action, StartupActionSelectContext)
	}
}

func TestStartupContextEnterSelectsContextOption(t *testing.T) {
	t.Parallel()

	m := newStartupContextModel("")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(startupContextModel)

	if cmd == nil {
		t.Fatal("expected quit command")
	}
	if model.choice.Action != StartupActionSelectContext {
		t.Fatalf("action = %v want %v", model.choice.Action, StartupActionSelectContext)
	}
}

func TestStartupContextCanSelectLastProjectOption(t *testing.T) {
	t.Parallel()

	m := newStartupContextModel("group/project")
	if len(m.options) != 2 {
		t.Fatalf("option count = %d want %d", len(m.options), 2)
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	model := updated.(startupContextModel)
	if model.selected != 1 {
		t.Fatalf("selected = %d want %d", model.selected, 1)
	}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(startupContextModel)
	if cmd == nil {
		t.Fatal("expected quit command")
	}
	if model.choice.Action != StartupActionUseLastProject {
		t.Fatalf("action = %v want %v", model.choice.Action, StartupActionUseLastProject)
	}
}

func TestStartupContextEscCancels(t *testing.T) {
	t.Parallel()

	m := newStartupContextModel("group/project")

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model := updated.(startupContextModel)

	if cmd == nil {
		t.Fatal("expected quit command")
	}
	if !model.cancelled {
		t.Fatal("expected model to be cancelled")
	}
}
