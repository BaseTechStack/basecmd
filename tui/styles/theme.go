package styles

import "github.com/charmbracelet/lipgloss"

// Color palette
var (
	Primary   = lipgloss.Color("#7C3AED") // violet
	Secondary = lipgloss.Color("#06B6D4") // cyan
	Success   = lipgloss.Color("#10B981") // green
	Warning   = lipgloss.Color("#F59E0B") // amber
	Danger    = lipgloss.Color("#EF4444") // red
	Muted     = lipgloss.Color("#6B7280") // gray
	Text      = lipgloss.Color("#F9FAFB") // white
	Subtle    = lipgloss.Color("#9CA3AF") // light gray
)

// Component styles
var (
	// Title / banner
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary)

	// Highlighted / selected text
	SelectedStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)

	// Normal text
	NormalStyle = lipgloss.NewStyle().
			Foreground(Text)

	// Muted / description text
	MutedStyle = lipgloss.NewStyle().
			Foreground(Muted)

	SubtleStyle = lipgloss.NewStyle().
			Foreground(Subtle)

	// Success message
	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success)

	// Error message
	ErrorStyle = lipgloss.NewStyle().
			Foreground(Danger)

	// Warning message
	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning)

	// Info / secondary
	InfoStyle = lipgloss.NewStyle().
			Foreground(Secondary)

	// Status bar at bottom
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(Subtle).
			Background(lipgloss.Color("#1F2937")).
			Padding(0, 1)

	// Key hint in status bar
	KeyStyle = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)

	// Spinner label
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(Secondary)

	// Box / panel border
	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Muted).
			Padding(1, 2)
)
