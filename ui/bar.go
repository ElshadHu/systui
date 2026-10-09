package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/ElshadHu/systui/ui/theme"
)

const (
	barWarningPercent  = 70
	barCriticalPercent = 90
)

// renderBar draws pct as smooth blocks, the whole fill in one level color
func renderBar(pct float64, width int, t theme.Theme) string {
	if width <= 0 {
		return ""
	}
	g := t.Glyphs
	eighths := min(max(int(pct/100*float64(width*8)+0.5), 0), width*8)
	full, partial := eighths/8, eighths%8
	fill := levelStyle(pct, t)
	track := width - full

	var b strings.Builder
	b.WriteString(t.Border.Render(g.BarLeft))
	b.WriteString(fill.Render(strings.Repeat(g.BarEighths[7], full)))
	if partial > 0 {
		b.WriteString(partialStyle(fill, t).Render(g.BarEighths[partial-1]))
		track--
	}
	b.WriteString(t.Border.Render(strings.Repeat(g.BarTrack, track)))
	b.WriteString(t.Border.Render(g.BarRight))
	return b.String()
}

func levelStyle(pct float64, t theme.Theme) lipgloss.Style {
	switch {
	case pct >= barCriticalPercent:
		return t.Critical
	case pct >= barWarningPercent:
		return t.Warning
	}
	return t.OK
}

// partialStyle paints the unfilled part of a partial block in the track color
func partialStyle(fill lipgloss.Style, t theme.Theme) lipgloss.Style {
	if t.Colors.Border == nil {
		return fill
	}
	return fill.Background(t.Colors.Border)
}
