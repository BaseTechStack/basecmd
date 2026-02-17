package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/base-go/cmd/headless"
	"github.com/base-go/cmd/tui"
	"github.com/base-go/cmd/tui/screens"
)

func main() {
	if len(os.Args) > 1 {
		// Headless / CI mode — parse args directly
		if err := headless.Run(os.Args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Interactive TUI mode
	app := tui.NewApp(screens.NewMainMenu())
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
