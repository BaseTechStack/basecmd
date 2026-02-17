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

type versionResultMsg struct {
	info operations.VersionInfo
}

// VersionScreen displays CLI version information and checks for updates.
type VersionScreen struct {
	info    *operations.VersionInfo
	spinner spinner.Model
	loading bool
	width   int
	height  int
}

// NewVersionScreen creates a new version info screen.
func NewVersionScreen() *VersionScreen {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.SpinnerStyle
	return &VersionScreen{spinner: s, loading: true}
}

func (v *VersionScreen) Title() string { return "Version" }

func (v *VersionScreen) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, func() tea.Msg {
		info := operations.GetVersionInfo()
		return versionResultMsg{info: info}
	})
}

func (v *VersionScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		return v, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return v, func() tea.Msg { return tui.NavigateBackMsg{} }
		}
	case versionResultMsg:
		v.info = &msg.info
		v.loading = false
		return v, nil
	}

	if v.loading {
		var cmd tea.Cmd
		v.spinner, cmd = v.spinner.Update(msg)
		return v, cmd
	}

	return v, nil
}

func (v *VersionScreen) View() string {
	header := components.Header(v.width)
	var body string

	if v.loading {
		body = fmt.Sprintf("\n  %s Checking version info...\n", v.spinner.View())
	} else if v.info != nil {
		var sb strings.Builder
		sb.WriteString("\n")
		sb.WriteString(styles.TitleStyle.Render("  Version Information"))
		sb.WriteString("\n\n")
		sb.WriteString(fmt.Sprintf("  Version:    %s\n", styles.InfoStyle.Render(v.info.BuildInfo.Version)))
		sb.WriteString(fmt.Sprintf("  Commit:     %s\n", v.info.BuildInfo.CommitHash))
		sb.WriteString(fmt.Sprintf("  Built:      %s\n", v.info.BuildInfo.BuildDate))
		sb.WriteString(fmt.Sprintf("  Go version: %s\n", v.info.BuildInfo.GoVersion))
		sb.WriteString("\n")

		if v.info.UpdateAvail {
			if v.info.IsMajor {
				sb.WriteString(styles.WarningStyle.Render(fmt.Sprintf("  MAJOR VERSION AVAILABLE: %s -> %s", v.info.BuildInfo.Version, v.info.LatestVersion)))
			} else {
				sb.WriteString(styles.InfoStyle.Render(fmt.Sprintf("  Update available: %s -> %s", v.info.BuildInfo.Version, v.info.LatestVersion)))
			}
			sb.WriteString("\n  Run: base upgrade\n")
		} else {
			sb.WriteString(styles.SuccessStyle.Render(fmt.Sprintf("  You're up to date! Using the latest version %s", v.info.BuildInfo.Version)))
			sb.WriteString("\n")
		}

		body = sb.String()
	}

	footer := components.StatusBar(v.width,
		fmt.Sprintf("%s back", styles.KeyStyle.Render("esc")),
		fmt.Sprintf("%s quit", styles.KeyStyle.Render("q")),
	)

	content := lipgloss.JoinVertical(lipgloss.Left, header, body)
	contentHeight := lipgloss.Height(content)
	if v.height > contentHeight+1 {
		content += strings.Repeat("\n", v.height-contentHeight-1)
	}
	return content + footer
}
