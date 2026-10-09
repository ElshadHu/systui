package ui

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/ElshadHu/systui/ui/theme"
	"github.com/charmbracelet/x/ansi"
)

func TestRenderBarRoundsToEighths(t *testing.T) {
	th, _ := theme.Named("dark")
	cases := []struct {
		pct  float64
		want string
	}{
		{0, "▕░░░░░░░░░░▏"},
		{15, "▕█▌░░░░░░░░▏"},
		{100, "▕██████████▏"},
		{140, "▕██████████▏"},
	}
	for _, c := range cases {
		got := ansi.Strip(renderBar(c.pct, 10, th))
		if got != c.want {
			t.Errorf("%v%%: want %s, got %s", c.pct, c.want, got)
		}
		if w := lipgloss.Width(renderBar(c.pct, 10, th)); w != 12 {
			t.Errorf("%v%%: want width 12, got %d", c.pct, w)
		}
	}
}
