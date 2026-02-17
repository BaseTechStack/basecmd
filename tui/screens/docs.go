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

type docsDoneMsg struct{ err error }

// DocsScreen handles OpenAPI documentation generation.
type DocsScreen struct {
	form      *huh.Form
	spinner   spinner.Model
	state     string // "form", "running", "done"
	outputDir string
	progress  []string
	err       error
	width     int
	height    int
}

// NewDocsScreen creates the documentation generation screen.
func NewDocsScreen() *DocsScreen {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.SpinnerStyle

	screen := &DocsScreen{
		spinner:   s,
		state:     "form",
		outputDir: "docs",
	}

	screen.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Output Directory").
				Description("Directory for generated documentation files").
				Placeholder("docs").
				Value(&screen.outputDir),
		),
	).WithShowHelp(false)

	return screen
}

func (d *DocsScreen) Title() string { return "Generate Docs" }

func (d *DocsScreen) Init() tea.Cmd {
	return d.form.Init()
}

func (d *DocsScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height
		return d, nil
	case tea.KeyMsg:
		if msg.String() == "esc" && d.state != "running" {
			return d, func() tea.Msg { return tui.NavigateBackMsg{} }
		}
	case docsDoneMsg:
		d.state = "done"
		d.err = msg.err
		return d, nil
	}

	switch d.state {
	case "form":
		form, cmd := d.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			d.form = f
		}

		if d.form.State == huh.StateCompleted {
			if d.outputDir == "" {
				d.outputDir = "docs"
			}
			d.state = "running"
			return d, tea.Batch(d.spinner.Tick, d.runDocs())
		}
		return d, cmd

	case "running":
		var cmd tea.Cmd
		d.spinner, cmd = d.spinner.Update(msg)
		return d, cmd
	}

	return d, nil
}

func (d *DocsScreen) runDocs() tea.Cmd {
	return func() tea.Msg {
		err := operations.GenerateDocs(d.outputDir, func(s string) {
			d.progress = append(d.progress, s)
		})
		return docsDoneMsg{err: err}
	}
}

func (d *DocsScreen) View() string {
	header := components.Header(d.width)
	var body string

	switch d.state {
	case "form":
		body = "\n" + styles.TitleStyle.Render("  Generate Documentation") + "\n\n"
		body += d.form.View()

	case "running":
		body = "\n" + styles.TitleStyle.Render("  Generating Documentation...") + "\n\n"
		for _, p := range d.progress {
			body += fmt.Sprintf("  %s\n", p)
		}
		body += fmt.Sprintf("\n  %s Working...\n", d.spinner.View())

	case "done":
		body = "\n"
		if d.err != nil {
			body += styles.ErrorStyle.Render(fmt.Sprintf("  Error: %v", d.err)) + "\n"
		} else {
			body += styles.SuccessStyle.Render("  Documentation generated successfully!") + "\n\n"
			body += "  Files:\n"
			body += fmt.Sprintf("    - %s/swagger.json\n", d.outputDir)
			body += fmt.Sprintf("    - %s/swagger.yaml\n", d.outputDir)
			body += fmt.Sprintf("    - %s/docs.go\n", d.outputDir)
		}
	}

	var hints []string
	if d.state != "running" {
		hints = append(hints, fmt.Sprintf("%s back", styles.KeyStyle.Render("esc")))
	}
	footer := components.StatusBar(d.width, hints...)

	content := lipgloss.JoinVertical(lipgloss.Left, header, body)
	contentHeight := lipgloss.Height(content)
	if d.height > contentHeight+1 {
		content += strings.Repeat("\n", d.height-contentHeight-1)
	}
	return content + footer
}
