package screens

import (
	"fmt"
	"strings"

	"github.com/base-go/cmd/operations"
	"github.com/base-go/cmd/tui"
	"github.com/base-go/cmd/tui/components"
	"github.com/base-go/cmd/tui/styles"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type generateDoneMsg struct{ err error }

// GenerateScreen handles module generation with a dynamic field builder.
type GenerateScreen struct {
	state      string // "name", "fields", "running", "done"
	nameInput  textinput.Model
	fieldInput textinput.Model
	moduleName string
	fields     []string
	spinner    spinner.Model
	progress   []string
	err        error
	width      int
	height     int
}

// NewGenerateScreen creates the module generation screen.
func NewGenerateScreen() *GenerateScreen {
	ni := textinput.New()
	ni.Placeholder = "Post"
	ni.Focus()
	ni.CharLimit = 64

	fi := textinput.New()
	fi.Placeholder = "title:string"
	fi.CharLimit = 128

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = styles.SpinnerStyle

	return &GenerateScreen{
		state:      "name",
		nameInput:  ni,
		fieldInput: fi,
		spinner:    s,
	}
}

func (g *GenerateScreen) Title() string { return "Generate Module" }

func (g *GenerateScreen) Init() tea.Cmd {
	return textinput.Blink
}

func (g *GenerateScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		g.width = msg.Width
		g.height = msg.Height
		return g, nil
	case tea.KeyMsg:
		switch g.state {
		case "name":
			switch msg.String() {
			case "esc":
				return g, func() tea.Msg { return tui.NavigateBackMsg{} }
			case "enter":
				name := strings.TrimSpace(g.nameInput.Value())
				if name != "" {
					g.moduleName = name
					g.state = "fields"
					g.fieldInput.Focus()
					return g, textinput.Blink
				}
			}
		case "fields":
			switch msg.String() {
			case "esc":
				if len(g.fields) == 0 {
					g.state = "name"
					g.nameInput.Focus()
					return g, textinput.Blink
				}
				// Generate with current fields
				g.state = "running"
				return g, tea.Batch(g.spinner.Tick, g.runGenerate())
			case "enter":
				field := strings.TrimSpace(g.fieldInput.Value())
				if field == "" && len(g.fields) > 0 {
					// Empty enter means done adding fields
					g.state = "running"
					return g, tea.Batch(g.spinner.Tick, g.runGenerate())
				}
				if field != "" {
					g.fields = append(g.fields, field)
					g.fieldInput.SetValue("")
					return g, nil
				}
			}
		case "done":
			if msg.String() == "esc" || msg.String() == "q" {
				return g, func() tea.Msg { return tui.NavigateBackMsg{} }
			}
		}
	case generateDoneMsg:
		g.state = "done"
		g.err = msg.err
		return g, nil
	}

	switch g.state {
	case "name":
		var cmd tea.Cmd
		g.nameInput, cmd = g.nameInput.Update(msg)
		return g, cmd
	case "fields":
		var cmd tea.Cmd
		g.fieldInput, cmd = g.fieldInput.Update(msg)
		return g, cmd
	case "running":
		var cmd tea.Cmd
		g.spinner, cmd = g.spinner.Update(msg)
		return g, cmd
	}

	return g, nil
}

func (g *GenerateScreen) runGenerate() tea.Cmd {
	return func() tea.Msg {
		err := operations.GenerateModule(g.moduleName, g.fields, func(s string) {
			g.progress = append(g.progress, s)
		})
		return generateDoneMsg{err: err}
	}
}

func (g *GenerateScreen) View() string {
	header := components.Header(g.width)
	var body string

	switch g.state {
	case "name":
		body = "\n" + styles.TitleStyle.Render("  Generate Module") + "\n\n"
		body += "  Module name:\n"
		body += "  " + g.nameInput.View() + "\n\n"
		body += styles.MutedStyle.Render("  Example: Post, ProductCategory, UserProfile") + "\n"

	case "fields":
		body = "\n" + styles.TitleStyle.Render(fmt.Sprintf("  Generate: %s", g.moduleName)) + "\n\n"
		if len(g.fields) > 0 {
			body += "  Fields:\n"
			for _, f := range g.fields {
				body += fmt.Sprintf("    %s %s\n", styles.SuccessStyle.Render("*"), f)
			}
			body += "\n"
		}
		body += "  Add field (name:type) or press enter to generate:\n"
		body += "  " + g.fieldInput.View() + "\n\n"
		body += styles.MutedStyle.Render("  Types: string, int, bool, float64, text, image, file, belongsTo, hasMany, hasOne, manyToMany") + "\n"
		body += styles.MutedStyle.Render("  Example: title:string, author:belongsTo:User, cover:image") + "\n"

	case "running":
		body = "\n" + styles.TitleStyle.Render(fmt.Sprintf("  Generating %s...", g.moduleName)) + "\n\n"
		for _, p := range g.progress {
			body += fmt.Sprintf("  %s\n", p)
		}
		body += fmt.Sprintf("\n  %s Working...\n", g.spinner.View())

	case "done":
		body = "\n"
		if g.err != nil {
			body += styles.ErrorStyle.Render(fmt.Sprintf("  Error: %v", g.err)) + "\n"
		} else {
			body += styles.SuccessStyle.Render(fmt.Sprintf("  Module '%s' generated successfully!", g.moduleName)) + "\n\n"
			body += "  Generated files:\n"
			body += "    - model, service, controller, module, validator\n"
		}
	}

	var hints []string
	switch g.state {
	case "name":
		hints = []string{
			fmt.Sprintf("%s confirm", styles.KeyStyle.Render("enter")),
			fmt.Sprintf("%s back", styles.KeyStyle.Render("esc")),
		}
	case "fields":
		hints = []string{
			fmt.Sprintf("%s add/generate", styles.KeyStyle.Render("enter")),
			fmt.Sprintf("%s generate", styles.KeyStyle.Render("esc")),
		}
	case "done":
		hints = []string{fmt.Sprintf("%s back", styles.KeyStyle.Render("esc"))}
	}
	footer := components.StatusBar(g.width, hints...)

	content := lipgloss.JoinVertical(lipgloss.Left, header, body)
	contentHeight := lipgloss.Height(content)
	if g.height > contentHeight+1 {
		content += strings.Repeat("\n", g.height-contentHeight-1)
	}
	return content + footer
}
