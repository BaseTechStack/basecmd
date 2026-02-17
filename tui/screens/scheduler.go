package screens

import (
	"fmt"
	"strings"

	"github.com/base-go/cmd/operations"
	"github.com/base-go/cmd/tui"
	"github.com/base-go/cmd/tui/components"
	"github.com/base-go/cmd/tui/styles"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type schedulerActionDoneMsg struct {
	result string
	err    error
}

// SchedulerScreen provides a sub-menu for all scheduler operations.
type SchedulerScreen struct {
	list    list.Model
	form    *huh.Form
	spinner spinner.Model
	state   string // "menu", "form", "running", "result"
	action  string
	result  string
	err     error
	// form field values
	moduleName string
	taskName   string
	apiKey     string
	baseURL    string
	width      int
	height     int
}

// NewSchedulerScreen creates the scheduler management screen.
func NewSchedulerScreen() *SchedulerScreen {
	items := []list.Item{
		menuItem{"Generate Task", "Generate a new scheduled task", nil},
		menuItem{"List Tasks", "List all registered tasks", nil},
		menuItem{"Run Task", "Run a specific task immediately", nil},
		menuItem{"Enable Task", "Enable a specific task", nil},
		menuItem{"Disable Task", "Disable a specific task", nil},
		menuItem{"Scheduler Status", "Get scheduler status", nil},
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(styles.Primary).
		BorderLeftForeground(styles.Primary)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(styles.Subtle).
		BorderLeftForeground(styles.Primary)

	l := list.New(items, delegate, 0, 0)
	l.Title = ""
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.SpinnerStyle

	return &SchedulerScreen{
		list:    l,
		spinner: s,
		state:   "menu",
		baseURL: "http://localhost:8100",
	}
}

func (s *SchedulerScreen) Title() string { return "Scheduler" }

func (s *SchedulerScreen) Init() tea.Cmd { return nil }

func (s *SchedulerScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		headerHeight := 8
		footerHeight := 1
		s.list.SetSize(msg.Width, msg.Height-headerHeight-footerHeight)
		return s, nil
	case tea.KeyMsg:
		switch s.state {
		case "menu":
			switch msg.String() {
			case "esc", "q":
				return s, func() tea.Msg { return tui.NavigateBackMsg{} }
			case "enter":
				if item, ok := s.list.SelectedItem().(menuItem); ok {
					s.action = item.title
					return s, s.buildForm()
				}
			}
		case "form":
			if msg.String() == "esc" {
				s.state = "menu"
				return s, nil
			}
		case "result":
			if msg.String() == "esc" || msg.String() == "q" {
				s.state = "menu"
				return s, nil
			}
		}
	case schedulerActionDoneMsg:
		s.state = "result"
		s.result = msg.result
		s.err = msg.err
		return s, nil
	}

	switch s.state {
	case "menu":
		var cmd tea.Cmd
		s.list, cmd = s.list.Update(msg)
		return s, cmd
	case "form":
		form, cmd := s.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			s.form = f
		}
		if s.form.State == huh.StateCompleted {
			s.state = "running"
			return s, tea.Batch(s.spinner.Tick, s.runAction())
		}
		if s.form.State == huh.StateAborted {
			s.state = "menu"
			return s, nil
		}
		return s, cmd
	case "running":
		var cmd tea.Cmd
		s.spinner, cmd = s.spinner.Update(msg)
		return s, cmd
	}

	return s, nil
}

func (s *SchedulerScreen) buildForm() tea.Cmd {
	s.state = "form"

	switch s.action {
	case "Generate Task":
		s.form = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().Title("Module Name").Placeholder("posts").Value(&s.moduleName),
				huh.NewInput().Title("Task Name").Placeholder("cleanup-old-posts").Value(&s.taskName),
			),
		).WithShowHelp(false)

	case "List Tasks", "Scheduler Status":
		s.form = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().Title("API Key").Description("Leave empty for offline mode").Placeholder("your-api-key").Value(&s.apiKey),
				huh.NewInput().Title("Server URL").Placeholder("http://localhost:8100").Value(&s.baseURL),
			),
		).WithShowHelp(false)

	case "Run Task", "Enable Task", "Disable Task":
		s.form = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().Title("Task Name").Placeholder("cleanup-old-posts").Value(&s.taskName),
				huh.NewInput().Title("API Key").Placeholder("your-api-key").Value(&s.apiKey),
				huh.NewInput().Title("Server URL").Placeholder("http://localhost:8100").Value(&s.baseURL),
			),
		).WithShowHelp(false)
	}

	if s.baseURL == "" {
		s.baseURL = "http://localhost:8100"
	}
	return s.form.Init()
}

func (s *SchedulerScreen) runAction() tea.Cmd {
	return func() tea.Msg {
		var result string
		var err error

		switch s.action {
		case "Generate Task":
			err = operations.SchedulerGenerateTask(s.moduleName, s.taskName, nil)
			if err == nil {
				result = fmt.Sprintf("Task '%s' generated in module '%s'", s.taskName, s.moduleName)
			}

		case "List Tasks":
			if s.apiKey == "" {
				result = "Provide an API key to list tasks from a running server.\n\nUsage: base scheduler list --api-key=your-key"
			} else {
				resp, apiErr := operations.SchedulerListTasks(s.apiKey, s.baseURL)
				if apiErr != nil {
					err = apiErr
				} else {
					result = formatTaskList(resp)
				}
			}

		case "Run Task":
			if s.apiKey == "" {
				result = "Provide an API key to run tasks.\n\nUsage: base scheduler run <task> --api-key=your-key"
			} else {
				_, apiErr := operations.SchedulerRunTask(s.taskName, s.apiKey, s.baseURL)
				if apiErr != nil {
					err = apiErr
				} else {
					result = fmt.Sprintf("Task '%s' executed successfully", s.taskName)
				}
			}

		case "Enable Task":
			if s.apiKey == "" {
				result = "Provide an API key to enable tasks."
			} else {
				_, apiErr := operations.SchedulerEnableTask(s.taskName, s.apiKey, s.baseURL)
				if apiErr != nil {
					err = apiErr
				} else {
					result = fmt.Sprintf("Task '%s' enabled", s.taskName)
				}
			}

		case "Disable Task":
			if s.apiKey == "" {
				result = "Provide an API key to disable tasks."
			} else {
				_, apiErr := operations.SchedulerDisableTask(s.taskName, s.apiKey, s.baseURL)
				if apiErr != nil {
					err = apiErr
				} else {
					result = fmt.Sprintf("Task '%s' disabled", s.taskName)
				}
			}

		case "Scheduler Status":
			if s.apiKey == "" {
				result = "Provide an API key to get scheduler status.\n\nUsage: base scheduler status --api-key=your-key"
			} else {
				resp, apiErr := operations.SchedulerGetStatus(s.apiKey, s.baseURL)
				if apiErr != nil {
					err = apiErr
				} else {
					result = formatSchedulerStatus(resp)
				}
			}
		}

		return schedulerActionDoneMsg{result: result, err: err}
	}
}

func formatTaskList(resp map[string]interface{}) string {
	var sb strings.Builder
	if data, ok := resp["data"]; ok {
		if tasks, ok := data.([]interface{}); ok {
			sb.WriteString(fmt.Sprintf("Found %d tasks:\n\n", len(tasks)))
			for _, taskInterface := range tasks {
				if task, ok := taskInterface.(map[string]interface{}); ok {
					name := operations.GetStringValue(task, "name")
					desc := operations.GetStringValue(task, "description")
					enabled := operations.GetBoolValue(task, "enabled")
					schedule := operations.GetStringValue(task, "schedule")
					status := "Disabled"
					if enabled {
						status = "Enabled"
					}
					sb.WriteString(fmt.Sprintf("  %s [%s]\n", name, status))
					sb.WriteString(fmt.Sprintf("    %s | Schedule: %s\n\n", desc, schedule))
				}
			}
		}
	}
	if sb.Len() == 0 {
		return "No tasks found."
	}
	return sb.String()
}

func formatSchedulerStatus(resp map[string]interface{}) string {
	var sb strings.Builder
	if data, ok := resp["data"]; ok {
		if status, ok := data.(map[string]interface{}); ok {
			running := operations.GetBoolValue(status, "running")
			total := operations.GetIntValue(status, "total_tasks")
			enabled := operations.GetIntValue(status, "enabled_tasks")
			disabled := operations.GetIntValue(status, "disabled_tasks")

			runStr := "Stopped"
			if running {
				runStr = "Running"
			}
			sb.WriteString(fmt.Sprintf("Scheduler: %s\n", runStr))
			sb.WriteString(fmt.Sprintf("Total tasks: %d | Enabled: %d | Disabled: %d\n", total, enabled, disabled))
		}
	}
	if sb.Len() == 0 {
		return "Unable to parse scheduler status."
	}
	return sb.String()
}

func (s *SchedulerScreen) View() string {
	header := components.Header(s.width)
	var body string

	switch s.state {
	case "menu":
		body = "\n" + styles.TitleStyle.Render("  Scheduler") + "\n\n"
		body += s.list.View()

	case "form":
		body = "\n" + styles.TitleStyle.Render(fmt.Sprintf("  Scheduler: %s", s.action)) + "\n\n"
		body += s.form.View()

	case "running":
		body = fmt.Sprintf("\n  %s Running %s...\n", s.spinner.View(), s.action)

	case "result":
		body = "\n" + styles.TitleStyle.Render(fmt.Sprintf("  %s", s.action)) + "\n\n"
		if s.err != nil {
			body += styles.ErrorStyle.Render(fmt.Sprintf("  Error: %v", s.err)) + "\n"
		} else {
			lines := strings.Split(s.result, "\n")
			for _, line := range lines {
				body += fmt.Sprintf("  %s\n", line)
			}
		}
	}

	var hints []string
	hints = append(hints, fmt.Sprintf("%s back", styles.KeyStyle.Render("esc")))
	footer := components.StatusBar(s.width, hints...)

	content := lipgloss.JoinVertical(lipgloss.Left, header, body)
	contentHeight := lipgloss.Height(content)
	if s.height > contentHeight+1 {
		content += strings.Repeat("\n", s.height-contentHeight-1)
	}
	return content + footer
}
