package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

type View int

const (
	OverviewView View = iota + 1
	AppsView
	DevView
	AllView
	TreeView
)

var viewNames = map[View]string{
	OverviewView: "Overview",
	AppsView:     "Apps",
	DevView:      "Dev",
	AllView:      "All",
	TreeView:     "Tree",
}

// tabBar lists the views with the current one highlighted, and on the right
// the filter, the sort pause and the cursor position of the Apps list
func (m Model) tabBar(first, last, total int) string {
	t := m.theme
	tabs := make([]string, 0, len(viewNames))
	for v := OverviewView; v <= TreeView; v++ {
		key := t.Accent.Render(fmt.Sprintf("[%d]", v))
		name := t.Muted.Render(viewNames[v])
		if v == m.view {
			name = t.Bright.Render(viewNames[v])
		}
		tabs = append(tabs, key+" "+name)
	}
	left := strings.Join(tabs, "  ")

	var right []string
	if m.view == AppsView {
		if m.apps.filter != "" || m.apps.typing {
			right = append(right, t.Accent.Render("/"+m.apps.filter))
		}
		if m.apps.frozen(m.now) {
			right = append(right, t.Muted.Render(t.Glyphs.Paused+" sort paused"))
		}
		if total > 0 {
			right = append(right, t.Muted.Render(fmt.Sprintf("rows %d–%d of %d", first, last, total)))
		}
	}
	return alignEnds(left, strings.Join(right, "  "), m.width)
}

// alignEnds puts left at the start and right at the end of a width wide line
func alignEnds(left, right string, width int) string {
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if width <= 0 || gap < 1 {
		if right == "" {
			return left
		}
		return left + "  " + right
	}
	return left + strings.Repeat(" ", gap) + right
}
