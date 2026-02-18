package tui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type StartupAction string

const (
	StartupActionSelectContext  StartupAction = "select_context"
	StartupActionUseLastProject StartupAction = "use_last_project"
)

type StartupContextChoice struct {
	Action      StartupAction
	Instance    InstanceOption
	ProjectPath string
}

type StartupProjectOption struct {
	Path  string
	Label string
}

type StartupContextFlowOptions struct {
	LastProject  string
	Instances    []InstanceOption
	LoadProjects func(instance InstanceOption) ([]StartupProjectOption, error)
}

type startupStage int

const (
	startupStageChoice startupStage = iota
	startupStageInstance
	startupStageLoadingProjects
	startupStageProject
)

type startupProjectsLoadedMsg struct {
	requestID int
	projects  []StartupProjectOption
	err       error
}

type startupOption struct {
	action StartupAction
	label  string
	hint   string
}

type startupContextModel struct {
	stage            startupStage
	lastProject      string
	width            int
	height           int
	options          []startupOption
	choiceSelected   int
	instances        []InstanceOption
	instanceSelected int
	selectedInstance InstanceOption
	projects         []StartupProjectOption
	filteredProjects []StartupProjectOption
	projectSelected  int
	searchInput      textinput.Model
	spinner          spinner.Model
	requestSeq       int
	requestID        int
	loadProjects     func(instance InstanceOption) ([]StartupProjectOption, error)
	errMessage       string
	choice           StartupContextChoice
	cancelled        bool
}

func RunStartupContextFlow(opts StartupContextFlowOptions) (StartupContextChoice, error) {
	m := newStartupContextModel(opts)
	p := tea.NewProgram(m)
	out, err := p.Run()
	if err != nil {
		return StartupContextChoice{}, err
	}

	final, ok := out.(startupContextModel)
	if !ok {
		return StartupContextChoice{}, fmt.Errorf("unexpected model type from program: %T", out)
	}
	if final.cancelled {
		return StartupContextChoice{}, ErrCancelled
	}
	if final.choice.Action == "" {
		return StartupContextChoice{}, fmt.Errorf("no startup action selected")
	}

	return final.choice, nil
}

func newStartupContextModel(opts StartupContextFlowOptions) startupContextModel {
	options := []startupOption{{
		action: StartupActionSelectContext,
		label:  "Select GitLab server and project",
		hint:   "Choose an instance, then choose a project",
	}}

	trimmedLastProject := strings.TrimSpace(opts.LastProject)
	if trimmedLastProject != "" {
		options = append(options, startupOption{
			action: StartupActionUseLastProject,
			label:  fmt.Sprintf("Use last project: %s", trimmedLastProject),
			hint:   "Skip pickers and continue immediately",
		})
	}

	search := textinput.New()
	search.Prompt = "Search: "
	search.Placeholder = "type to filter projects"
	search.CharLimit = 120
	search.Width = 52
	search.Focus()

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("39"))

	instances := make([]InstanceOption, 0, len(opts.Instances))
	for _, instance := range opts.Instances {
		if strings.TrimSpace(instance.Host) == "" {
			continue
		}
		instances = append(instances, instance)
	}

	return startupContextModel{
		stage:        startupStageChoice,
		lastProject:  trimmedLastProject,
		options:      options,
		instances:    instances,
		searchInput:  search,
		spinner:      sp,
		requestSeq:   1,
		requestID:    1,
		loadProjects: opts.LoadProjects,
	}
}

func (m startupContextModel) Init() tea.Cmd {
	return nil
}

func (m startupContextModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if m.stage == startupStageLoadingProjects {
			return m, cmd
		}
		return m, nil

	case startupProjectsLoadedMsg:
		if msg.requestID != m.requestID || m.stage != startupStageLoadingProjects {
			return m, nil
		}
		if msg.err != nil {
			m.stage = startupStageInstance
			m.errMessage = fmt.Sprintf("Could not load projects: %v", msg.err)
			return m, nil
		}

		projects := normalizeStartupProjects(msg.projects)
		if len(projects) == 0 {
			m.stage = startupStageInstance
			m.errMessage = "No projects found for this instance"
			return m, nil
		}

		m.stage = startupStageProject
		m.projects = projects
		m.projectSelected = 0
		m.searchInput.SetValue("")
		m.searchInput.Focus()
		m.applyProjectFilter()
		m.errMessage = ""
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.stage {
		case startupStageChoice:
			return m.updateChoice(msg)
		case startupStageInstance:
			return m.updateInstance(msg)
		case startupStageLoadingProjects:
			return m.updateLoading(msg)
		case startupStageProject:
			return m.updateProject(msg)
		}
	}

	return m, nil
}

func (m startupContextModel) View() string {
	switch m.stage {
	case startupStageInstance:
		return m.renderInstanceView()
	case startupStageLoadingProjects:
		return m.renderLoadingView()
	case startupStageProject:
		return m.renderProjectView()
	default:
		return m.renderChoiceView()
	}
}

func (m startupContextModel) updateChoice(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc", "q":
		m.cancelled = true
		return m, tea.Quit
	case "down", "j", "tab":
		if m.choiceSelected < len(m.options)-1 {
			m.choiceSelected++
		}
		return m, nil
	case "up", "k", "shift+tab":
		if m.choiceSelected > 0 {
			m.choiceSelected--
		}
		return m, nil
	case "enter":
		if len(m.options) == 0 {
			return m, nil
		}

		choice := m.options[m.choiceSelected]
		if choice.action == StartupActionUseLastProject {
			m.choice = StartupContextChoice{Action: StartupActionUseLastProject, ProjectPath: strings.TrimSpace(m.lastProject)}
			return m, tea.Quit
		}

		m.stage = startupStageInstance
		m.errMessage = ""
		return m, nil
	}

	return m, nil
}

func (m startupContextModel) updateInstance(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		m.cancelled = true
		return m, tea.Quit
	case "esc":
		m.stage = startupStageChoice
		m.errMessage = ""
		return m, nil
	case "down", "j", "tab":
		if m.instanceSelected < len(m.instances)-1 {
			m.instanceSelected++
		}
		return m, nil
	case "up", "k", "shift+tab":
		if m.instanceSelected > 0 {
			m.instanceSelected--
		}
		return m, nil
	case "enter":
		if len(m.instances) == 0 {
			m.errMessage = "No configured instances found"
			return m, nil
		}
		if m.loadProjects == nil {
			m.errMessage = "Project loader is unavailable"
			return m, nil
		}
		m.selectedInstance = m.instances[m.instanceSelected]
		m.stage = startupStageLoadingProjects
		m.errMessage = ""
		m.requestSeq++
		m.requestID = m.requestSeq
		return m, tea.Batch(m.spinner.Tick, m.loadProjectsCmd(m.selectedInstance, m.requestID))
	}

	return m, nil
}

func (m startupContextModel) updateLoading(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		m.cancelled = true
		return m, tea.Quit
	case "esc":
		m.stage = startupStageInstance
		m.errMessage = ""
		return m, nil
	}

	return m, nil
}

func (m startupContextModel) updateProject(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyRunes {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		m.applyProjectFilter()
		return m, cmd
	}

	switch msg.String() {
	case "ctrl+c":
		m.cancelled = true
		return m, tea.Quit
	case "esc":
		m.stage = startupStageInstance
		m.errMessage = ""
		return m, nil
	case "down", "tab":
		if m.projectSelected < len(m.filteredProjects)-1 {
			m.projectSelected++
		}
		return m, nil
	case "up", "shift+tab":
		if m.projectSelected > 0 {
			m.projectSelected--
		}
		return m, nil
	case "enter":
		if len(m.filteredProjects) == 0 {
			return m, nil
		}
		project := strings.TrimSpace(m.filteredProjects[m.projectSelected].Path)
		if project == "" {
			return m, nil
		}
		m.choice = StartupContextChoice{
			Action:      StartupActionSelectContext,
			Instance:    m.selectedInstance,
			ProjectPath: project,
		}
		return m, tea.Quit
	}

	return m, nil
}

func (m startupContextModel) loadProjectsCmd(instance InstanceOption, requestID int) tea.Cmd {
	return func() tea.Msg {
		if m.loadProjects == nil {
			return startupProjectsLoadedMsg{requestID: requestID, err: errors.New("project loader is unavailable")}
		}
		projects, err := m.loadProjects(instance)
		return startupProjectsLoadedMsg{requestID: requestID, projects: projects, err: err}
	}
}

func (m *startupContextModel) applyProjectFilter() {
	if len(m.projects) == 0 {
		m.filteredProjects = nil
		m.projectSelected = 0
		return
	}

	query := strings.ToLower(strings.TrimSpace(m.searchInput.Value()))
	if query == "" {
		m.filteredProjects = append([]StartupProjectOption(nil), m.projects...)
	} else {
		filtered := make([]StartupProjectOption, 0, len(m.projects))
		for _, project := range m.projects {
			label := strings.ToLower(strings.TrimSpace(startupProjectLabel(project)))
			path := strings.ToLower(strings.TrimSpace(project.Path))
			if strings.Contains(label, query) || strings.Contains(path, query) {
				filtered = append(filtered, project)
			}
		}
		m.filteredProjects = filtered
	}

	if m.projectSelected >= len(m.filteredProjects) {
		m.projectSelected = 0
	}
	if m.projectSelected < 0 {
		m.projectSelected = 0
	}
}

func normalizeStartupProjects(projects []StartupProjectOption) []StartupProjectOption {
	if len(projects) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(projects))
	normalized := make([]StartupProjectOption, 0, len(projects))
	for _, project := range projects {
		path := strings.TrimSpace(project.Path)
		if path == "" {
			continue
		}
		if _, exists := seen[path]; exists {
			continue
		}
		seen[path] = struct{}{}
		normalized = append(normalized, StartupProjectOption{Path: path, Label: strings.TrimSpace(project.Label)})
	}

	return normalized
}

func startupProjectLabel(project StartupProjectOption) string {
	label := strings.TrimSpace(project.Label)
	if label != "" {
		return label
	}
	return strings.TrimSpace(project.Path)
}

func (m startupContextModel) renderChoiceView() string {
	rows := []string{
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("Startup Context"),
		"",
		"No GitLab repository context was detected for the current directory.",
		"Choose how to continue:",
		"",
	}

	for i, option := range m.options {
		prefix := "  "
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
		if i == m.choiceSelected {
			prefix = "> "
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		}

		rows = append(rows, style.Render(prefix+option.label))
		if strings.TrimSpace(option.hint) != "" {
			rows = append(rows, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("  "+option.hint))
		}
		rows = append(rows, "")
	}

	if strings.TrimSpace(m.errMessage) != "" {
		rows = append(rows, lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.errMessage), "")
	}

	rows = append(rows, "Enter to select, j/k or arrows to move, Tab/Shift+Tab to cycle, q or Esc to cancel")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Width(84)

	return m.renderCenteredBox(box, rows)
}

func (m startupContextModel) renderInstanceView() string {
	rows := []string{
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("Startup Context / GitLab Server"),
		"",
		"Select a GitLab instance:",
		"",
	}

	if len(m.instances) == 0 {
		rows = append(rows, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("No configured instances found"))
	} else {
		for i, instance := range m.instances {
			prefix := "  "
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
			if i == m.instanceSelected {
				prefix = "> "
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
			}

			label := instance.Label
			if strings.TrimSpace(label) == "" {
				label = instance.Host
			}
			rows = append(rows, style.Render(prefix+label))
		}
	}

	if strings.TrimSpace(m.errMessage) != "" {
		rows = append(rows, "", lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.errMessage))
	}

	rows = append(rows, "", "Enter select | j/k move | Esc back | q cancel")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Width(84)

	return m.renderCenteredBox(box, rows)
}

func (m startupContextModel) renderLoadingView() string {
	instanceLabel := strings.TrimSpace(m.selectedInstance.Label)
	if instanceLabel == "" {
		instanceLabel = strings.TrimSpace(m.selectedInstance.Host)
	}

	rows := []string{
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("Startup Context / Projects"),
		"",
		fmt.Sprintf("%s Loading projects from %s...", m.spinner.View(), instanceLabel),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Esc back to server selection"),
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Width(84)

	return m.renderCenteredBox(box, rows)
}

func (m startupContextModel) renderProjectView() string {
	instanceLabel := strings.TrimSpace(m.selectedInstance.Label)
	if instanceLabel == "" {
		instanceLabel = strings.TrimSpace(m.selectedInstance.Host)
	}

	rows := []string{
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Render("Startup Context / Project"),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Instance: " + instanceLabel),
		m.searchInput.View(),
		"",
	}

	if len(m.filteredProjects) == 0 {
		rows = append(rows, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("No matching projects"))
	} else {
		const windowSize = 12
		start, end := visibleRange(len(m.filteredProjects), m.projectSelected, windowSize)
		for i := start; i < end; i++ {
			project := startupProjectLabel(m.filteredProjects[i])
			prefix := "  "
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
			if i == m.projectSelected {
				prefix = "› "
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
			}
			rows = append(rows, style.Render(prefix+project))
		}

		if len(m.filteredProjects) > windowSize {
			rows = append(rows, "", lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(fmt.Sprintf("%d-%d of %d", start+1, end, len(m.filteredProjects))))
		}
	}

	rows = append(rows, "", "Type to filter | Enter select | arrows/tab move | Esc back | Ctrl+C cancel")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Width(96)

	return m.renderCenteredBox(box, rows)
}

func (m startupContextModel) renderCenteredBox(boxStyle lipgloss.Style, rows []string) string {
	content := boxStyle.Render(strings.Join(rows, "\n"))
	if m.width <= 0 || m.height <= 0 {
		return content
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
