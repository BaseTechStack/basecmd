package screens

import (
	"fmt"
	"strings"

	"github.com/base-go/cmd/operations"
	"github.com/base-go/cmd/tui"
	"github.com/base-go/cmd/tui/components"
	"github.com/base-go/cmd/tui/styles"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type newProjectDoneMsg struct{ err error }

// NewProjectScreen is a multi-step wizard for creating a new Base project.
type NewProjectScreen struct {
	form     *huh.Form
	spinner  spinner.Model
	state    string // "form", "running", "done"
	name     string
	progress []string
	err      error
	width    int
	height   int
}

// NewNewProjectScreen creates the new-project wizard screen.
func NewNewProjectScreen() *NewProjectScreen {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.SpinnerStyle

	screen := &NewProjectScreen{
		spinner: s,
		state:   "form",
	}

	screen.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project Name").
				Description("Enter a name for your new Base project").
				Placeholder("my-project").
				Value(&screen.name),
		),
	).WithShowHelp(false).WithShowErrors(true)

	return screen
}

func (n *NewProjectScreen) Title() string { return "New Project" }

func (n *NewProjectScreen) Init() tea.Cmd {
	return n.form.Init()
}

func (n *NewProjectScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		n.width = msg.Width
		n.height = msg.Height
		return n, nil
	case tea.KeyMsg:
		if msg.String() == "esc" {
			if n.state == "done" || n.state == "form" {
				return n, func() tea.Msg { return tui.NavigateBackMsg{} }
			}
		}
	case newProjectDoneMsg:
		n.state = "done"
		n.err = msg.err
		return n, nil
	}

	switch n.state {
	case "form":
		form, cmd := n.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			n.form = f
		}

		if n.form.State == huh.StateCompleted {
			n.state = "running"
			return n, tea.Batch(n.spinner.Tick, n.runCreate())
		}

		return n, cmd

	case "running":
		var cmd tea.Cmd
		n.spinner, cmd = n.spinner.Update(msg)
		return n, cmd
	}

	return n, nil
}

func (n *NewProjectScreen) runCreate() tea.Cmd {
	return func() tea.Msg {
		err := operations.NewProject(n.name, func(s string) {
			n.progress = append(n.progress, s)
		})
		return newProjectDoneMsg{err: err}
	}
}

func (n *NewProjectScreen) View() string {
	header := components.Header(n.width)
	var body string

	switch n.state {
	case "form":
		body = "\n" + styles.TitleStyle.Render("  Create New Project") + "\n\n"
		body += n.form.View()

	case "running":
		body = "\n" + styles.TitleStyle.Render("  Creating Project...") + "\n\n"
		for _, p := range n.progress {
			body += fmt.Sprintf("  %s\n", p)
		}
		body += fmt.Sprintf("\n  %s Working...\n", n.spinner.View())

	case "done":
		body = "\n"
		if n.err != nil {
			body += styles.ErrorStyle.Render(fmt.Sprintf("  Error: %v", n.err)) + "\n"
		} else {
			body += styles.SuccessStyle.Render("  Project created successfully!") + "\n\n"
			body += "  Next steps:\n"
			body += fmt.Sprintf("    cd %s && base start\n", n.name)
		}
	}

	var hints []string
	if n.state == "done" || n.state == "form" {
		hints = append(hints, fmt.Sprintf("%s back", styles.KeyStyle.Render("esc")))
	}
	footer := components.StatusBar(n.width, hints...)

	content := lipgloss.JoinVertical(lipgloss.Left, header, body)
	contentHeight := lipgloss.Height(content)
	if n.height > contentHeight+1 {
		content += strings.Repeat("\n", n.height-contentHeight-1)
	}
	return content + footer
}
