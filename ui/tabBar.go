package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/ElshadHu/systui/ui/draw"
)

type View int

const (
	AppsView View = iota + 1
	OverviewView
	LifetimeView
	AllView
	TreeView
)

var viewNames = map[View]string{
	AppsView:     "Apps",
	OverviewView: "Overview",
	LifetimeView: "Lifetime",
	AllView:      "All",
	TreeView:     "Tree",
}

const tabGap = "  "

func tabKey(v View) string {
	return fmt.Sprintf("[%d]", v)
}

// tabWidth is how many cells one tab takes in the tab bar
func tabWidth(v View) int {
	return lipgloss.Width(tabKey(v) + " " + viewNames[v])
}

// tabAt returns the view whose tab is drawn at column x of the tab bar
func tabAt(x int) (View, bool) {
	start := 0
	for v := AppsView; v <= TreeView; v++ {
		end := start + tabWidth(v)
		if x >= start && x < end {
			return v, true
		}
		start = end + len(tabGap)
	}
	return 0, false
}

// tabBar lists the views with the current one highlighted, and on the right
// the filter, the sort pause and the cursor position of the Apps list
func (m Model) tabBar(first, last, total int) string {
	t := m.theme
	tabs := make([]string, 0, len(viewNames))
	for v := AppsView; v <= TreeView; v++ {
		key := t.Accent.Render(tabKey(v))
		name := t.Muted.Render(viewNames[v])
		if v == m.view {
			name = t.Bright.Render(viewNames[v])
		}
		tabs = append(tabs, key+" "+name)
	}
	left := strings.Join(tabs, tabGap)

	var right []string
	if m.view == AppsView {
		if m.apps.Filter() != "" || m.apps.Typing {
			right = append(right, t.Accent.Render("/"+m.apps.Filter()))
		}
		if m.apps.Frozen(m.now) {
			right = append(right, t.Muted.Render(t.Glyphs.Paused+" sort paused"))
		}
		if total > 0 {
			right = append(right, t.Muted.Render(fmt.Sprintf("rows %d–%d of %d", first, last, total)))
		}
	}
	return draw.AlignEnds(left, strings.Join(right, "  "), m.width)
}
