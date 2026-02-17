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
	"github.com/charmbracelet/lipgloss"
)

type updateDoneMsg struct{ err error }

// UpdateScreen handles core framework file updates with a confirmation step.
type UpdateScreen struct {
	spinner  spinner.Model
	state    string // "confirm", "running", "done"
	progress []string
	err      error
	width    int
	height   int
}

// NewUpdateScreen creates the core-update screen.
func NewUpdateScreen() *UpdateScreen {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.SpinnerStyle
	return &UpdateScreen{spinner: s, state: "confirm"}
}

func (u *UpdateScreen) Title() string { return "Update Core" }

func (u *UpdateScreen) Init() tea.Cmd { return nil }

func (u *UpdateScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		u.width = msg.Width
		u.height = msg.Height
		return u, nil
	case tea.KeyMsg:
		switch u.state {
		case "confirm":
			switch msg.String() {
			case "y", "Y", "enter":
				u.state = "running"
				return u, tea.Batch(u.spinner.Tick, u.runUpdate())
			case "n", "N", "esc", "q":
				return u, func() tea.Msg { return tui.NavigateBackMsg{} }
			}
		case "done":
			if msg.String() == "esc" || msg.String() == "q" {
				return u, func() tea.Msg { return tui.NavigateBackMsg{} }
			}
		}
	case updateDoneMsg:
		u.state = "done"
		u.err = msg.err
		return u, nil
	}

	if u.state == "running" {
		var cmd tea.Cmd
		u.spinner, cmd = u.spinner.Update(msg)
		return u, cmd
	}

	return u, nil
}

func (u *UpdateScreen) runUpdate() tea.Cmd {
	return func() tea.Msg {
		err := operations.UpdateCore(func(s string) {
			u.progress = append(u.progress, s)
		})
		return updateDoneMsg{err: err}
	}
}

func (u *UpdateScreen) View() string {
	header := components.Header(u.width)
	var body string

	switch u.state {
	case "confirm":
		body = "\n" + styles.TitleStyle.Render("  Update Core") + "\n\n"
		body += "  This will update the core/ directory to the latest version.\n"
		body += "  A backup will be created before updating.\n\n"
		body += fmt.Sprintf("  %s to proceed, %s to cancel\n",
			styles.KeyStyle.Render("y"),
			styles.KeyStyle.Render("n"))

	case "running":
		body = "\n" + styles.TitleStyle.Render("  Updating Core...") + "\n\n"
		for _, p := range u.progress {
			body += fmt.Sprintf("  %s\n", p)
		}
		body += fmt.Sprintf("\n  %s Working...\n", u.spinner.View())

	case "done":
		body = "\n"
		if u.err != nil {
			body += styles.ErrorStyle.Render(fmt.Sprintf("  Error: %v", u.err)) + "\n"
		} else {
			body += styles.SuccessStyle.Render("  Core updated successfully!") + "\n"
		}
		for _, p := range u.progress {
			body += fmt.Sprintf("  %s\n", p)
		}
	}

	var hints []string
	if u.state != "running" {
		hints = append(hints, fmt.Sprintf("%s back", styles.KeyStyle.Render("esc")))
	}
	footer := components.StatusBar(u.width, hints...)

	content := lipgloss.JoinVertical(lipgloss.Left, header, body)
	contentHeight := lipgloss.Height(content)
	if u.height > contentHeight+1 {
		content += strings.Repeat("\n", u.height-contentHeight-1)
	}
	return content + footer
}
