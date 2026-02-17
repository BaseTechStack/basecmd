package components

import (
	"fmt"

	"github.com/base-go/cmd/tui/styles"
	"github.com/base-go/cmd/version"
	"github.com/charmbracelet/lipgloss"
)

const banner = `
 ____
| __ )  __ _ ___  ___
|  _ \ / _` + "`" + ` / __|/ _ \
| |_) | (_| \__ \  __/
|____/ \__,_|___/\___|`

// Header renders the ASCII banner with version.
func Header(width int) string {
	info := version.GetBuildInfo()

	bannerStyled := styles.TitleStyle.Render(banner)
	ver := styles.SubtleStyle.Render(fmt.Sprintf("  v%s", info.Version))

	block := lipgloss.JoinVertical(lipgloss.Left, bannerStyled, ver, "")

	if width > 0 {
		block = lipgloss.PlaceHorizontal(width, lipgloss.Center, block)
	}
	return block
}
