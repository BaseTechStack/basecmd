package tui

import tea "github.com/charmbracelet/bubbletea"

// Screen is the interface all TUI screens implement.
type Screen interface {
	tea.Model
	// Title returns the screen title for the header.
	Title() string
}

// NavigateMsg tells the root model to push a new screen.
type NavigateMsg struct {
	Screen Screen
}

// NavigateBackMsg tells the root model to pop the current screen.
type NavigateBackMsg struct{}

// ErrorMsg carries an error to display.
type ErrorMsg struct {
	Err error
}

// StatusMsg carries a status line update.
type StatusMsg struct {
	Text string
}

// SuccessMsg carries a success message after an operation completes.
type SuccessMsg struct {
	Text string
}

// WindowSizeMsg is forwarded to child screens when the terminal resizes.
type WindowSizeMsg = tea.WindowSizeMsg
