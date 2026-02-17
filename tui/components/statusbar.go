package components

import (
	"github.com/base-go/cmd/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

// StatusBar renders the bottom bar with key hints.
func StatusBar(width int, hints ...string) string {
	var parts string
	for i, h := range hints {
		if i > 0 {
			parts += styles.MutedStyle.Render(" · ")
		}
		parts += styles.SubtleStyle.Render(h)
	}

	bar := styles.StatusBarStyle.
		Width(width).
		Render(parts)

	return lipgloss.PlaceHorizontal(width, lipgloss.Left, bar)
}
