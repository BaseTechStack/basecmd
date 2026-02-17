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

type upgradeCheckMsg struct {
	result *operations.UpgradeResult
	err    error
}
type upgradeDoneMsg struct{ err error }

// UpgradeScreen checks for CLI updates and offers to install them.
type UpgradeScreen struct {
	spinner  spinner.Model
	state    string // "checking", "confirm", "running", "done"
	result   *operations.UpgradeResult
	progress []string
	err      error
	width    int
	height   int
}

// NewUpgradeScreen creates the CLI upgrade screen.
func NewUpgradeScreen() *UpgradeScreen {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.SpinnerStyle
	return &UpgradeScreen{spinner: s, state: "checking"}
}

func (u *UpgradeScreen) Title() string { return "Upgrade CLI" }

func (u *UpgradeScreen) Init() tea.Cmd {
	return tea.Batch(u.spinner.Tick, func() tea.Msg {
		result, err := operations.CheckUpgrade(false, false)
		return upgradeCheckMsg{result: result, err: err}
	})
}

func (u *UpgradeScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				return u, tea.Batch(u.spinner.Tick, u.runUpgrade())
			case "n", "N", "esc", "q":
				return u, func() tea.Msg { return tui.NavigateBackMsg{} }
			}
		case "done":
			if msg.String() == "esc" || msg.String() == "q" {
				return u, func() tea.Msg { return tui.NavigateBackMsg{} }
			}
		case "checking":
			if msg.String() == "esc" {
				return u, func() tea.Msg { return tui.NavigateBackMsg{} }
			}
		}
	case upgradeCheckMsg:
		if msg.err != nil {
			u.state = "done"
			u.err = msg.err
			return u, nil
		}
		u.result = msg.result
		if u.result.IsUpToDate {
			u.state = "done"
			return u, nil
		}
		u.state = "confirm"
		return u, nil
	case upgradeDoneMsg:
		u.state = "done"
		u.err = msg.err
		return u, nil
	}

	if u.state == "checking" || u.state == "running" {
		var cmd tea.Cmd
		u.spinner, cmd = u.spinner.Update(msg)
		return u, cmd
	}

	return u, nil
}

func (u *UpgradeScreen) runUpgrade() tea.Cmd {
	return func() tea.Msg {
		err := operations.PerformUpgrade(u.result.Release, u.result.TargetVersion, func(s string) {
			u.progress = append(u.progress, s)
		})
		return upgradeDoneMsg{err: err}
	}
}

func (u *UpgradeScreen) View() string {
	header := components.Header(u.width)
	var body string

	switch u.state {
	case "checking":
		body = fmt.Sprintf("\n  %s Checking for updates...\n", u.spinner.View())

	case "confirm":
		body = "\n" + styles.TitleStyle.Render("  Upgrade CLI") + "\n\n"
		body += fmt.Sprintf("  Current version: %s\n", u.result.CurrentVersion)
		body += fmt.Sprintf("  Available version: %s\n\n", styles.InfoStyle.Render(u.result.TargetVersion))
		if u.result.IsMajor {
			body += styles.WarningStyle.Render("  This is a MAJOR version upgrade with potential breaking changes!") + "\n\n"
		}
		body += fmt.Sprintf("  %s to upgrade, %s to cancel\n",
			styles.KeyStyle.Render("y"),
			styles.KeyStyle.Render("n"))

	case "running":
		body = "\n" + styles.TitleStyle.Render("  Upgrading CLI...") + "\n\n"
		for _, p := range u.progress {
			body += fmt.Sprintf("  %s\n", p)
		}
		body += fmt.Sprintf("\n  %s Working...\n", u.spinner.View())

	case "done":
		body = "\n"
		if u.err != nil {
			body += styles.ErrorStyle.Render(fmt.Sprintf("  Error: %v", u.err)) + "\n"
		} else if u.result != nil && u.result.IsUpToDate {
			body += styles.SuccessStyle.Render(fmt.Sprintf("  You're already using the latest version (%s)", u.result.CurrentVersion)) + "\n"
		} else {
			body += styles.SuccessStyle.Render("  CLI upgraded successfully!") + "\n"
			for _, p := range u.progress {
				body += fmt.Sprintf("  %s\n", p)
			}
		}
	}

	var hints []string
	if u.state != "running" && u.state != "checking" {
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
