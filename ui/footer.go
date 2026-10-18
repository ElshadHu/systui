package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/ElshadHu/systui/ui/draw"
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
	case m.listsApps() && m.apps.Typing:
		hints = []keyHint{{"type", "to filter"}, {"esc", "clear"}, {t.Glyphs.Enter, "done"}}
	case m.listsApps():
		hints = []keyHint{
			{"↑↓", "move"}, {"→←", "expand"}, {"E/C", "all"},
			{"m p c t", "sort"}, {"/", "filter"}, {"1-5", "view"}, {"q", "quit"},
		}
		if m.apps.Filter() != "" {
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
	legend := m.legend()
	if legend == "" || lipgloss.Width(keys)+2+lipgloss.Width(legend) > m.width {
		return keys
	}
	return draw.AlignEnds(keys, legend, m.width)
}

// legend explains the tab's own column in one line, shown when it fits
func (m Model) legend() string {
	switch {
	case m.apps.Typing:
		return ""
	case m.view == LifetimeView:
		return m.theme.Muted.Render("peak and CPU time since process start, group peak since systui start")
	case m.view == AppsView && m.columns().has(sparkField):
		return m.trendLegend()
	}
	return ""
}

// trendLegend explains the sparkline column in one line, shown when it fits
func (m Model) trendLegend() string {
	t := m.theme
	levels := t.Glyphs.SparkLevels
	return t.Accent.Render(levels[0]+levels[2]+levels[5]) + " " +
		t.Muted.Render("trend: last 20 s, one bar per second, newest on the right")
}
