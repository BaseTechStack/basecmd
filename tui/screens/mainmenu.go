package screens

import (
	"fmt"
	"strings"

	"github.com/base-go/cmd/tui"
	"github.com/base-go/cmd/tui/components"
	"github.com/base-go/cmd/tui/styles"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type menuItem struct {
	title, desc string
	screen      func() tui.Screen
}

func (i menuItem) Title() string       { return i.title }
func (i menuItem) Description() string { return i.desc }
func (i menuItem) FilterValue() string { return i.title }

// MainMenuScreen displays the top-level command list.
type MainMenuScreen struct {
	list   list.Model
	width  int
	height int
}

// NewMainMenu creates the main menu screen with all available commands.
func NewMainMenu() *MainMenuScreen {
	items := []list.Item{
		menuItem{"New Project", "Create a new Base project", func() tui.Screen { return NewNewProjectScreen() }},
		menuItem{"Generate Module", "Generate CRUD module with fields", func() tui.Screen { return NewGenerateScreen() }},
		menuItem{"Start Server", "Start the development server", func() tui.Screen { return NewStartScreen() }},
		menuItem{"Generate Docs", "Generate OpenAPI documentation", func() tui.Screen { return NewDocsScreen() }},
		menuItem{"Destroy Module", "Remove existing modules", func() tui.Screen { return NewDestroyScreen() }},
		menuItem{"Scheduler", "Manage scheduled tasks", func() tui.Screen { return NewSchedulerScreen() }},
		menuItem{"Update Core", "Update framework core files", func() tui.Screen { return NewUpdateScreen() }},
		menuItem{"Upgrade CLI", "Upgrade Base CLI to latest", func() tui.Screen { return NewUpgradeScreen() }},
		menuItem{"Version", "Show version information", func() tui.Screen { return NewVersionScreen() }},
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

	return &MainMenuScreen{list: l}
}

func (m *MainMenuScreen) Title() string { return "Main Menu" }

func (m *MainMenuScreen) Init() tea.Cmd { return nil }

func (m *MainMenuScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerHeight := 8 // banner + version + padding
		footerHeight := 1
		m.list.SetSize(msg.Width, msg.Height-headerHeight-footerHeight)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, func() tea.Msg { return tui.NavigateBackMsg{} }
		case "enter":
			if item, ok := m.list.SelectedItem().(menuItem); ok {
				screen := item.screen()
				return m, func() tea.Msg {
					return tui.NavigateMsg{Screen: screen}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *MainMenuScreen) View() string {
	header := components.Header(m.width)
	menu := m.list.View()
	footer := components.StatusBar(m.width,
		fmt.Sprintf("%s select", styles.KeyStyle.Render("enter")),
		fmt.Sprintf("%s quit", styles.KeyStyle.Render("q")),
	)

	content := lipgloss.JoinVertical(lipgloss.Left, header, menu)
	contentHeight := lipgloss.Height(content)
	if m.height > contentHeight+1 {
		content += strings.Repeat("\n", m.height-contentHeight-1)
	}
	content += footer

	return content
}
