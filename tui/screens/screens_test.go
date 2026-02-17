package screens

import (
	"testing"

	"github.com/base-go/cmd/tui"
	tea "github.com/charmbracelet/bubbletea"
)

// Verify all screen constructors return valid Screen implementations.
func TestAllScreenConstructors(t *testing.T) {
	screens := []struct {
		name   string
		screen tui.Screen
	}{
		{"MainMenu", NewMainMenu()},
		{"NewProject", NewNewProjectScreen()},
		{"Generate", NewGenerateScreen()},
		{"Start", NewStartScreen()},
		{"Docs", NewDocsScreen()},
		{"Destroy", NewDestroyScreen()},
		{"Scheduler", NewSchedulerScreen()},
		{"Update", NewUpdateScreen()},
		{"Upgrade", NewUpgradeScreen()},
		{"Version", NewVersionScreen()},
	}

	for _, tt := range screens {
		t.Run(tt.name, func(t *testing.T) {
			if tt.screen == nil {
				t.Fatal("constructor returned nil")
			}
			if tt.screen.Title() == "" {
				t.Error("Title() is empty")
			}
		})
	}
}

// Test that screens handle WindowSizeMsg without panicking.
func TestScreensHandleWindowSize(t *testing.T) {
	screens := []struct {
		name   string
		screen tui.Screen
	}{
		{"MainMenu", NewMainMenu()},
		{"NewProject", NewNewProjectScreen()},
		{"Generate", NewGenerateScreen()},
		{"Start", NewStartScreen()},
		{"Docs", NewDocsScreen()},
		{"Scheduler", NewSchedulerScreen()},
		{"Update", NewUpdateScreen()},
		{"Upgrade", NewUpgradeScreen()},
		{"Version", NewVersionScreen()},
	}

	msg := tea.WindowSizeMsg{Width: 80, Height: 24}

	for _, tt := range screens {
		t.Run(tt.name, func(t *testing.T) {
			updated, _ := tt.screen.Update(msg)
			if updated == nil {
				t.Error("Update(WindowSizeMsg) returned nil model")
			}
		})
	}
}

// Test that screens produce non-empty views.
func TestScreensProduceViews(t *testing.T) {
	screens := []struct {
		name   string
		screen tui.Screen
	}{
		{"MainMenu", NewMainMenu()},
		{"Generate", NewGenerateScreen()},
		{"Update", NewUpdateScreen()},
	}

	for _, tt := range screens {
		t.Run(tt.name, func(t *testing.T) {
			// Set size first
			updated, _ := tt.screen.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
			screen := updated.(tui.Screen)
			view := screen.View()
			if view == "" {
				t.Error("View() returned empty string")
			}
		})
	}
}

// Test MainMenu navigation.
func TestMainMenuEscQuits(t *testing.T) {
	menu := NewMainMenu()
	_, cmd := menu.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("pressing q should return a cmd")
		return
	}
	// Execute the command to get the message
	msg := cmd()
	if _, ok := msg.(tui.NavigateBackMsg); !ok {
		t.Errorf("q should send NavigateBackMsg, got %T", msg)
	}
}

// Test GenerateScreen name input flow.
func TestGenerateScreenNameInput(t *testing.T) {
	g := NewGenerateScreen()
	if g.state != "name" {
		t.Fatalf("initial state = %q, want %q", g.state, "name")
	}

	// Type a module name
	for _, r := range "Post" {
		g.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Press enter to move to fields state
	updated, _ := g.Update(tea.KeyMsg{Type: tea.KeyEnter})
	g = updated.(*GenerateScreen)

	if g.state != "fields" {
		t.Errorf("after entering name, state = %q, want %q", g.state, "fields")
	}
	if g.moduleName != "Post" {
		t.Errorf("moduleName = %q, want %q", g.moduleName, "Post")
	}
}

// Test UpdateScreen confirm flow.
func TestUpdateScreenConfirm(t *testing.T) {
	u := NewUpdateScreen()
	if u.state != "confirm" {
		t.Fatalf("initial state = %q, want %q", u.state, "confirm")
	}

	// Press 'n' to cancel
	_, cmd := u.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if cmd == nil {
		t.Error("pressing n should return a cmd")
		return
	}
	msg := cmd()
	if _, ok := msg.(tui.NavigateBackMsg); !ok {
		t.Errorf("n should send NavigateBackMsg, got %T", msg)
	}
}

// Test VersionScreen initial state.
func TestVersionScreenLoading(t *testing.T) {
	v := NewVersionScreen()
	if !v.loading {
		t.Error("VersionScreen should start in loading state")
	}
}
