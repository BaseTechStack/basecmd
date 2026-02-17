package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// mockScreen is a minimal Screen implementation for testing.
type mockScreen struct {
	title   string
	updated bool
	viewStr string
}

func newMockScreen(title string) *mockScreen {
	return &mockScreen{title: title, viewStr: title + " view"}
}

func (m *mockScreen) Title() string { return m.title }
func (m *mockScreen) Init() tea.Cmd { return nil }
func (m *mockScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.updated = true
	return m, nil
}
func (m *mockScreen) View() string { return m.viewStr }

func TestNewApp(t *testing.T) {
	screen := newMockScreen("main")
	app := NewApp(screen)

	if app.current() == nil {
		t.Fatal("current() should not be nil after NewApp")
	}
	if app.current().Title() != "main" {
		t.Errorf("current().Title() = %q, want %q", app.current().Title(), "main")
	}
}

func TestNavigateMsg(t *testing.T) {
	main := newMockScreen("main")
	app := NewApp(main)

	// Navigate to a new screen
	second := newMockScreen("second")
	updated, _ := app.Update(NavigateMsg{Screen: second})
	app = updated.(AppModel)

	if app.current().Title() != "second" {
		t.Errorf("after navigate, Title() = %q, want %q", app.current().Title(), "second")
	}
	if len(app.screenStack) != 2 {
		t.Errorf("screenStack len = %d, want 2", len(app.screenStack))
	}
}

func TestNavigateBackMsg(t *testing.T) {
	main := newMockScreen("main")
	app := NewApp(main)

	// Push second screen
	second := newMockScreen("second")
	updated, _ := app.Update(NavigateMsg{Screen: second})
	app = updated.(AppModel)

	// Navigate back
	updated, _ = app.Update(NavigateBackMsg{})
	app = updated.(AppModel)

	if app.current().Title() != "main" {
		t.Errorf("after back, Title() = %q, want %q", app.current().Title(), "main")
	}
	if len(app.screenStack) != 1 {
		t.Errorf("screenStack len = %d, want 1", len(app.screenStack))
	}
}

func TestNavigateBackFromRoot(t *testing.T) {
	main := newMockScreen("main")
	app := NewApp(main)

	// Navigate back from root should trigger quit
	updated, cmd := app.Update(NavigateBackMsg{})
	app = updated.(AppModel)

	if !app.quitting {
		t.Error("navigating back from root should set quitting = true")
	}
	if cmd == nil {
		t.Error("navigating back from root should return a Quit cmd")
	}
}

func TestWindowSizeMsg(t *testing.T) {
	main := newMockScreen("main")
	app := NewApp(main)

	updated, _ := app.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	app = updated.(AppModel)

	if app.width != 80 || app.height != 24 {
		t.Errorf("size = %dx%d, want 80x24", app.width, app.height)
	}
}

func TestViewDelegation(t *testing.T) {
	main := newMockScreen("main")
	app := NewApp(main)

	view := app.View()
	if view != "main view" {
		t.Errorf("View() = %q, want %q", view, "main view")
	}
}

func TestViewWhenQuitting(t *testing.T) {
	main := newMockScreen("main")
	app := NewApp(main)
	app.quitting = true

	view := app.View()
	if view != "" {
		t.Errorf("View() when quitting should be empty, got %q", view)
	}
}

func TestCtrlCQuits(t *testing.T) {
	main := newMockScreen("main")
	app := NewApp(main)

	updated, cmd := app.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	app = updated.(AppModel)

	if !app.quitting {
		t.Error("ctrl+c should set quitting = true")
	}
	if cmd == nil {
		t.Error("ctrl+c should return a Quit cmd")
	}
}

func TestDeepNavigation(t *testing.T) {
	s1 := newMockScreen("first")
	app := NewApp(s1)

	s2 := newMockScreen("second")
	s3 := newMockScreen("third")

	updated, _ := app.Update(NavigateMsg{Screen: s2})
	app = updated.(AppModel)
	updated, _ = app.Update(NavigateMsg{Screen: s3})
	app = updated.(AppModel)

	if len(app.screenStack) != 3 {
		t.Fatalf("screenStack len = %d, want 3", len(app.screenStack))
	}
	if app.current().Title() != "third" {
		t.Errorf("current = %q, want %q", app.current().Title(), "third")
	}

	// Pop back twice
	updated, _ = app.Update(NavigateBackMsg{})
	app = updated.(AppModel)
	if app.current().Title() != "second" {
		t.Errorf("after 1st back = %q, want %q", app.current().Title(), "second")
	}

	updated, _ = app.Update(NavigateBackMsg{})
	app = updated.(AppModel)
	if app.current().Title() != "first" {
		t.Errorf("after 2nd back = %q, want %q", app.current().Title(), "first")
	}
}
