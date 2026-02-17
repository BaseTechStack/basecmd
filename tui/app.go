package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// AppModel is the root Bubbletea model. It holds a stack of screens.
type AppModel struct {
	screenStack []Screen
	width       int
	height      int
	quitting    bool
}

// NewApp creates the root application model.
// The initial screen (main menu) must be pushed after construction
// so callers can inject it.
func NewApp(initial Screen) AppModel {
	return AppModel{
		screenStack: []Screen{initial},
	}
}

func (m AppModel) current() Screen {
	if len(m.screenStack) == 0 {
		return nil
	}
	return m.screenStack[len(m.screenStack)-1]
}

// Init delegates to the current screen's Init.
func (m AppModel) Init() tea.Cmd {
	if s := m.current(); s != nil {
		return s.Init()
	}
	return nil
}

// Update handles global keys and delegates the rest to the active screen.
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Forward to current screen
		if s := m.current(); s != nil {
			updated, cmd := s.Update(msg)
			m.screenStack[len(m.screenStack)-1] = updated.(Screen)
			return m, cmd
		}
		return m, nil

	case NavigateMsg:
		m.screenStack = append(m.screenStack, msg.Screen)
		// Send window size to the new screen so it can layout.
		return m, tea.Batch(
			m.current().Init(),
			func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			},
		)

	case NavigateBackMsg:
		if len(m.screenStack) > 1 {
			m.screenStack = m.screenStack[:len(m.screenStack)-1]
			// Refresh current screen's size
			return m, func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			}
		}
		// At root — quit
		m.quitting = true
		return m, tea.Quit

	case tea.KeyMsg:
		// ctrl+c always quits
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
	}

	// Delegate to active screen
	if s := m.current(); s != nil {
		updated, cmd := s.Update(msg)
		m.screenStack[len(m.screenStack)-1] = updated.(Screen)
		return m, cmd
	}

	return m, nil
}

// View delegates to the active screen.
func (m AppModel) View() string {
	if m.quitting {
		return ""
	}
	if s := m.current(); s != nil {
		return s.View()
	}
	return ""
}
