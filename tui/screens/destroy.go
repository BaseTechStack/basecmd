package screens

import (
	"fmt"
	"os"
	"path/filepath"
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

type destroyDoneMsg struct{ err error }
type destroyModulesScannedMsg struct{ modules []string }

// DestroyScreen scans for existing modules and lets the user select which to remove.
type DestroyScreen struct {
	form     *huh.Form
	spinner  spinner.Model
	state    string // "scanning", "select", "running", "done"
	modules  []string
	selected []string
	progress []string
	err      error
	width    int
	height   int
}

// NewDestroyScreen creates the module destroy screen.
func NewDestroyScreen() *DestroyScreen {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.SpinnerStyle
	return &DestroyScreen{spinner: s, state: "scanning"}
}

func (d *DestroyScreen) Title() string { return "Destroy Module" }

func (d *DestroyScreen) Init() tea.Cmd {
	return tea.Batch(d.spinner.Tick, d.scanModules())
}

func (d *DestroyScreen) scanModules() tea.Cmd {
	return func() tea.Msg {
		var modules []string
		appDir := filepath.Join(".", "app")
		entries, err := os.ReadDir(appDir)
		if err != nil {
			return destroyModulesScannedMsg{modules: nil}
		}
		for _, e := range entries {
			if e.IsDir() && e.Name() != "models" && e.Name() != "model" {
				modules = append(modules, e.Name())
			}
		}
		return destroyModulesScannedMsg{modules: modules}
	}
}

func (d *DestroyScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height
		return d, nil
	case tea.KeyMsg:
		if msg.String() == "esc" && d.state != "running" {
			return d, func() tea.Msg { return tui.NavigateBackMsg{} }
		}
	case destroyModulesScannedMsg:
		d.modules = msg.modules
		if len(d.modules) == 0 {
			d.state = "done"
			d.err = fmt.Errorf("no modules found in app/ directory")
			return d, nil
		}
		d.state = "select"
		options := make([]huh.Option[string], len(d.modules))
		for i, m := range d.modules {
			options[i] = huh.NewOption(m, m)
		}
		d.form = huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("Select modules to destroy").
					Description("Use space to select, enter to confirm").
					Options(options...).
					Value(&d.selected),
				huh.NewConfirm().
					Title("Are you sure?").
					Description("This will permanently delete the selected modules").
					Affirmative("Yes, destroy").
					Negative("Cancel"),
			),
		).WithShowHelp(false)
		return d, d.form.Init()
	case destroyDoneMsg:
		d.state = "done"
		d.err = msg.err
		return d, nil
	}

	switch d.state {
	case "scanning":
		var cmd tea.Cmd
		d.spinner, cmd = d.spinner.Update(msg)
		return d, cmd
	case "select":
		form, cmd := d.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			d.form = f
		}
		if d.form.State == huh.StateCompleted {
			if len(d.selected) == 0 {
				return d, func() tea.Msg { return tui.NavigateBackMsg{} }
			}
			d.state = "running"
			return d, tea.Batch(d.spinner.Tick, d.runDestroy())
		}
		if d.form.State == huh.StateAborted {
			return d, func() tea.Msg { return tui.NavigateBackMsg{} }
		}
		return d, cmd
	case "running":
		var cmd tea.Cmd
		d.spinner, cmd = d.spinner.Update(msg)
		return d, cmd
	}

	return d, nil
}

func (d *DestroyScreen) runDestroy() tea.Cmd {
	return func() tea.Msg {
		err := operations.DestroyModules(d.selected, func(s string) {
			d.progress = append(d.progress, s)
		})
		return destroyDoneMsg{err: err}
	}
}

func (d *DestroyScreen) View() string {
	header := components.Header(d.width)
	var body string

	switch d.state {
	case "scanning":
		body = fmt.Sprintf("\n  %s Scanning for modules...\n", d.spinner.View())
	case "select":
		body = "\n" + styles.TitleStyle.Render("  Destroy Modules") + "\n\n"
		body += d.form.View()
	case "running":
		body = "\n" + styles.TitleStyle.Render("  Destroying Modules...") + "\n\n"
		for _, p := range d.progress {
			body += fmt.Sprintf("  %s\n", p)
		}
		body += fmt.Sprintf("\n  %s Working...\n", d.spinner.View())
	case "done":
		body = "\n"
		if d.err != nil {
			body += styles.ErrorStyle.Render(fmt.Sprintf("  Error: %v", d.err)) + "\n"
		} else {
			body += styles.SuccessStyle.Render(fmt.Sprintf("  Successfully destroyed %d module(s)!", len(d.selected))) + "\n"
		}
		for _, p := range d.progress {
			body += fmt.Sprintf("  %s\n", p)
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
