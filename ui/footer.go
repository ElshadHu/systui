package ui

import "strings"

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
	return strings.Join(parts, "  ")
}
