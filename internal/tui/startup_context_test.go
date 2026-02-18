package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestStartupContextModelWithoutLastProjectShowsSingleOption(t *testing.T) {
	t.Parallel()

	m := newStartupContextModel(StartupContextFlowOptions{})
	if len(m.options) != 1 {
		t.Fatalf("option count = %d want %d", len(m.options), 1)
	}
	if m.options[0].action != StartupActionSelectContext {
		t.Fatalf("first action = %v want %v", m.options[0].action, StartupActionSelectContext)
	}
}

func TestStartupContextEnterMovesToInstanceSelection(t *testing.T) {
	t.Parallel()

	m := newStartupContextModel(StartupContextFlowOptions{
		Instances: []InstanceOption{{Host: "https://gitlab.com/api/v4", Label: "gitlab.com"}},
	})

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(startupContextModel)

	if model.stage != startupStageInstance {
		t.Fatalf("stage = %v want %v", model.stage, startupStageInstance)
	}
}

func TestStartupContextCanSelectLastProjectOption(t *testing.T) {
	t.Parallel()

	m := newStartupContextModel(StartupContextFlowOptions{LastProject: "group/project"})
	if len(m.options) != 2 {
		t.Fatalf("option count = %d want %d", len(m.options), 2)
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	model := updated.(startupContextModel)
	if model.choiceSelected != 1 {
		t.Fatalf("selected = %d want %d", model.choiceSelected, 1)
	}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(startupContextModel)
	if cmd == nil {
		t.Fatal("expected quit command")
	}
	if model.choice.Action != StartupActionUseLastProject {
		t.Fatalf("action = %v want %v", model.choice.Action, StartupActionUseLastProject)
	}
	if model.choice.ProjectPath != "group/project" {
		t.Fatalf("project path = %q want %q", model.choice.ProjectPath, "group/project")
	}
}

func TestStartupContextEscCancelsFromChoice(t *testing.T) {
	t.Parallel()

	m := newStartupContextModel(StartupContextFlowOptions{LastProject: "group/project"})

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model := updated.(startupContextModel)

	if cmd == nil {
		t.Fatal("expected quit command")
	}
	if !model.cancelled {
		t.Fatal("expected model to be cancelled")
	}
}

func TestStartupContextSelectContextLoadsProjectsAndReturnsSelection(t *testing.T) {
	t.Parallel()

	instance := InstanceOption{Host: "https://gitlab.com/api/v4", Label: "gitlab.com"}
	m := newStartupContextModel(StartupContextFlowOptions{
		Instances: []InstanceOption{instance},
		LoadProjects: func(selected InstanceOption) ([]StartupProjectOption, error) {
			if selected.Host != instance.Host {
				t.Fatalf("instance host = %q want %q", selected.Host, instance.Host)
			}
			return []StartupProjectOption{{Path: "group/project-one"}, {Path: "group/project-two"}}, nil
		},
	})

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model := updated.(startupContextModel)
	if model.stage != startupStageInstance {
		t.Fatalf("stage = %v want %v", model.stage, startupStageInstance)
	}

	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(startupContextModel)
	if model.stage != startupStageLoadingProjects {
		t.Fatalf("stage = %v want %v", model.stage, startupStageLoadingProjects)
	}
	if cmd == nil {
		t.Fatal("expected project load command")
	}

	updated, _ = model.Update(startupProjectsLoadedMsg{
		requestID: model.requestID,
		projects:  []StartupProjectOption{{Path: "group/project-one"}, {Path: "group/project-two"}},
	})
	model = updated.(startupContextModel)
	if model.stage != startupStageProject {
		t.Fatalf("stage = %v want %v", model.stage, startupStageProject)
	}

	updated, quitCmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(startupContextModel)
	if quitCmd == nil {
		t.Fatal("expected quit command")
	}
	if model.choice.Action != StartupActionSelectContext {
		t.Fatalf("action = %v want %v", model.choice.Action, StartupActionSelectContext)
	}
	if model.choice.Instance.Host != instance.Host {
		t.Fatalf("instance host = %q want %q", model.choice.Instance.Host, instance.Host)
	}
	if model.choice.ProjectPath != "group/project-one" {
		t.Fatalf("project path = %q want %q", model.choice.ProjectPath, "group/project-one")
	}
}

func TestStartupContextProjectEscReturnsToInstanceStage(t *testing.T) {
	t.Parallel()

	instance := InstanceOption{Host: "https://gitlab.com/api/v4", Label: "gitlab.com"}
	m := newStartupContextModel(StartupContextFlowOptions{Instances: []InstanceOption{instance}})
	m.stage = startupStageProject
	m.selectedInstance = instance
	m.projects = []StartupProjectOption{{Path: "group/project-one"}}
	m.applyProjectFilter()

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model := updated.(startupContextModel)

	if model.stage != startupStageInstance {
		t.Fatalf("stage = %v want %v", model.stage, startupStageInstance)
	}
}
