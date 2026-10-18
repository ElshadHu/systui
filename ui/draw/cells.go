package draw

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/ElshadHu/systui/ui/theme"
)

// Painter puts a background behind a style, or leaves it alone
type Painter func(lipgloss.Style) lipgloss.Style

// Surface paints the theme's surface color behind header and selected rows
func Surface(t theme.Theme, on bool) Painter {
	bg := t.Colors.Surface
	return func(s lipgloss.Style) lipgloss.Style {
		if !on || bg == nil {
			return s
		}
		return s.Background(bg)
	}
}

// Cell is one column of a row: its style, text, width and alignment
type Cell struct {
	Style lipgloss.Style
	Text  string
	Width int
	Right bool
}

func JoinCells(cells []Cell, gap int, gapStyle lipgloss.Style) string {
	parts := make([]string, len(cells))
	for i, c := range cells {
		text := Truncate(c.Text, c.Width)
		pad := strings.Repeat(" ", max(c.Width-lipgloss.Width(text), 0))
		if c.Right {
			parts[i] = c.Style.Render(pad + text)
		} else {
			parts[i] = c.Style.Render(text + pad)
		}
	}
	return strings.Join(parts, gapStyle.Render(strings.Repeat(" ", gap)))
}

// FillLine pads a line to width so a painted background reaches the edge
func FillLine(line string, width int, style lipgloss.Style) string {
	return line + style.Render(strings.Repeat(" ", max(width-lipgloss.Width(line), 0)))
}
