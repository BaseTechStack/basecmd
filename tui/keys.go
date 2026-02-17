package tui

import "github.com/charmbracelet/bubbles/key"

// GlobalKeyMap defines keybindings available everywhere.
type GlobalKeyMap struct {
	Quit key.Binding
	Back key.Binding
	Help key.Binding
}

// DefaultGlobalKeyMap returns the default global keybindings.
func DefaultGlobalKeyMap() GlobalKeyMap {
	return GlobalKeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}
