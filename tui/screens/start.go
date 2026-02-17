package screens

import (
	"fmt"
	"strings"

	"github.com/base-go/cmd/operations"
	"github.com/base-go/cmd/tui"
	"github.com/base-go/cmd/tui/components"
	"github.com/base-go/cmd/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type serverExitMsg struct{ err error }

// StartScreen prompts for options then hands the terminal to the running server.
type StartScreen struct {
	form     *huh.Form
	state    string // "form", "running", "done"
	withDocs bool
	err      error
	width    int
	height   int
}

// NewStartScreen creates the start-server screen.
func NewStartScreen() *StartScreen {
	s := &StartScreen{state: "form"}
	s.form = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Generate Swagger docs?").
				Description("Generate and serve Swagger documentation").
				Value(&s.withDocs),
		),
	).WithShowHelp(false)
	return s
}

func (s *StartScreen) Title() string { return "Start Server" }

func (s *StartScreen) Init() tea.Cmd {
	return s.form.Init()
}

func (s *StartScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		return s, nil
	case tea.KeyMsg:
		if msg.String() == "esc" && s.state != "running" {
			return s, func() tea.Msg { return tui.NavigateBackMsg{} }
		}
	case serverExitMsg:
		s.state = "done"
		s.err = msg.err
		return s, nil
	}

	switch s.state {
	case "form":
		form, cmd := s.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			s.form = f
		}

		if s.form.State == huh.StateCompleted {
			s.state = "running"
			// Use tea.ExecProcess to hand over the terminal to the server
			cmd, err := operations.StartServer(s.withDocs, func(string) {})
			if err != nil {
				s.state = "done"
				s.err = err
				return s, nil
			}
			return s, tea.ExecProcess(cmd, func(err error) tea.Msg {
				return serverExitMsg{err: err}
			})
		}

		return s, cmd
	}

	return s, nil
}

func (s *StartScreen) View() string {
	header := components.Header(s.width)
	var body string

	switch s.state {
	case "form":
		body = "\n" + styles.TitleStyle.Render("  Start Server") + "\n\n"
		body += s.form.View()

	case "running":
		body = "\n  Server is running... (this screen suspends while the server is active)\n"

	case "done":
		body = "\n"
		if s.err != nil {
			body += styles.ErrorStyle.Render(fmt.Sprintf("  Server exited with error: %v", s.err)) + "\n"
		} else {
			body += styles.SuccessStyle.Render("  Server stopped.") + "\n"
		}
	}

	var hints []string
	if s.state != "running" {
		hints = append(hints, fmt.Sprintf("%s back", styles.KeyStyle.Render("esc")))
	}
	footer := components.StatusBar(s.width, hints...)

	content := lipgloss.JoinVertical(lipgloss.Left, header, body)
	contentHeight := lipgloss.Height(content)
	if s.height > contentHeight+1 {
		content += strings.Repeat("\n", s.height-contentHeight-1)
	}
	return content + footer
}
