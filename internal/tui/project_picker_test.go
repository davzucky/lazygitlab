package tui

import (
	"fmt"
	"strings"
	"testing"

	gl "gitlab.com/gitlab-org/api/client-go"
)

func TestProjectPickerViewScrollsToSelectedProject(t *testing.T) {
	t.Parallel()

	m := newPickerModel(testProjects(20))
	m.selected = 15

	view := m.View()
	if !strings.Contains(view, "› group/project-16") {
		t.Fatalf("expected selected project to be visible, got view: %q", view)
	}
	if strings.Contains(view, "group/project-01") {
		t.Fatalf("expected earliest project to be outside window, got view: %q", view)
	}
}

func TestProjectPickerViewShowsRangeForLongLists(t *testing.T) {
	t.Parallel()

	m := newPickerModel(testProjects(20))
	m.selected = 15

	view := m.View()
	if !strings.Contains(view, "9-20 of 20") {
		t.Fatalf("expected view range footer for long list, got view: %q", view)
	}
}

func testProjects(count int) []*gl.Project {
	projects := make([]*gl.Project, 0, count)
	for i := 1; i <= count; i++ {
		projects = append(projects, &gl.Project{PathWithNamespace: fmt.Sprintf("group/project-%02d", i)})
	}
	return projects
}
