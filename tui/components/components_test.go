package components

import (
	"strings"
	"testing"
)

func TestHeaderContainsBanner(t *testing.T) {
	h := Header(80)
	// The banner ASCII art contains "____" as part of the top line and bottom line.
	// We check for this distinctive pattern rather than "Base" because lipgloss
	// ANSI styling may break up the word across escape sequences.
	if !strings.Contains(h, "____") {
		t.Error("Header should contain '____' from the ASCII banner art")
	}
}

func TestHeaderZeroWidth(t *testing.T) {
	h := Header(0)
	if h == "" {
		t.Error("Header(0) should still produce output")
	}
}

func TestStatusBarRendersHints(t *testing.T) {
	bar := StatusBar(80, "enter select", "q quit")
	if bar == "" {
		t.Error("StatusBar should produce output")
	}
}

func TestStatusBarNoHints(t *testing.T) {
	bar := StatusBar(80)
	if bar == "" {
		t.Error("StatusBar with no hints should still produce output")
	}
}
