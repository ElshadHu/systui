package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

type keyHint struct {
	key  string
	word string
}

// footer shows the keys that work right now, letter in accent and word in muted
func (m Model) footer() string {
	t := m.theme
	var hints []keyHint
	switch {
	case m.view == AppsView && m.apps.typing:
		hints = []keyHint{{"type", "to filter"}, {"esc", "clear"}, {t.Glyphs.Enter, "done"}}
	case m.view == AppsView:
		hints = []keyHint{
			{"↑↓", "move"}, {"→←", "expand"}, {"E/C", "all"},
			{"c m", "sort"}, {"/", "filter"}, {"1-5", "view"}, {"q", "quit"},
		}
		if m.apps.filter != "" {
			hints = append(hints[:len(hints)-2], keyHint{"esc", "clear filter"}, hints[len(hints)-2], hints[len(hints)-1])
		}
	default:
		hints = []keyHint{{"1-5", "view"}, {"q", "quit"}}
	}
	parts := make([]string, len(hints))
	for i, h := range hints {
		parts[i] = t.Accent.Render(h.key) + " " + t.Muted.Render(h.word)
	}
	keys := strings.Join(parts, "  ")
	if m.view != AppsView || m.apps.typing || !fitAppsColumns(m.width).spark {
		return keys
	}
	legend := m.trendLegend()
	if lipgloss.Width(keys)+2+lipgloss.Width(legend) > m.width {
		return keys
	}
	return alignEnds(keys, legend, m.width)
}

// trendLegend explains the sparkline column in one line, shown when it fits
func (m Model) trendLegend() string {
	t := m.theme
	levels := t.Glyphs.SparkLevels
	return t.Accent.Render(levels[0]+levels[2]+levels[5]) + " " +
		t.Muted.Render("trend: last 20 s, one bar per second, newest on the right")
}
