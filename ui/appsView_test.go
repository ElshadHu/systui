package ui

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/ElshadHu/systui/ui/theme"
	"github.com/charmbracelet/x/ansi"
)

func TestFitAppsColumnsDropsFromTheRight(t *testing.T) {
	cases := []struct {
		width       int
		procs, flag bool
	}{
		{120, true, true},
		{60, true, false},
		{45, false, false},
	}
	for _, c := range cases {
		got := fitAppsColumns(c.width)
		if got.procs != c.procs || got.flag != c.flag || got.name < minNameColumn {
			t.Errorf("width %d: got %+v", c.width, got)
		}
	}
}

func testModel(t *testing.T) Model {
	th, err := theme.Named("dark")
	if err != nil {
		t.Fatal(err)
	}
	m, err := New(th)
	if err != nil {
		t.Fatal(err)
	}
	m.width, m.height = 80, 20
	m.apps.setGroups(sampleGroups(), time.Unix(0, 0))
	return m
}

func TestRowsNeverExceedTheTerminalWidth(t *testing.T) {
	m := testModel(t)
	m.apps.expand()
	for _, width := range []int{40, 60, 80, 120} {
		m.width = width
		for _, line := range m.appsView(10) {
			if w := ansi.StringWidth(line); w > width {
				t.Errorf("width %d: line is %d wide: %q", width, w, ansi.Strip(line))
			}
		}
	}
}

func press(m Model, keys ...string) Model {
	for _, k := range keys {
		var msg tea.KeyPressMsg
		switch k {
		case "right", "left", "up", "down", "enter", "esc", "space", "backspace":
			msg = tea.KeyPressMsg{Code: map[string]rune{
				"right": tea.KeyRight, "left": tea.KeyLeft, "up": tea.KeyUp, "down": tea.KeyDown,
				"enter": tea.KeyEnter, "esc": tea.KeyEscape, "space": tea.KeySpace, "backspace": tea.KeyBackspace,
			}[k]}
		default:
			msg = tea.KeyPressMsg{Code: rune(k[0]), Text: k}
		}
		next, _ := m.Update(msg)
		m = next.(Model)
	}
	return m
}

func TestKeysDriveTheAppsView(t *testing.T) {
	m := testModel(t)
	m = press(m, "right", "down", "down")
	if m.apps.selected != (rowID{group: "chrome", pid: 11}) {
		t.Fatalf("right then down twice should land on the second member, got %v", m.apps.selected)
	}
	m = press(m, "c")
	if m.apps.sort != 1 || !strings.Contains(ansi.Strip(m.appsHeader(fitAppsColumns(80))), "CPU "+m.theme.Glyphs.Expanded) {
		t.Fatal("c should sort by CPU and mark the column")
	}
	m = press(m, "/", "l", "o", "o", "enter")
	if m.apps.typing || m.apps.filter != "loo" || len(m.apps.rows) != 1 {
		t.Fatalf("typing a filter should narrow the rows, got %q typing=%v rows=%d", m.apps.filter, m.apps.typing, len(m.apps.rows))
	}
	m = press(m, "esc")
	if m.apps.filter != "" || len(m.apps.rows) != 9 {
		t.Fatalf("esc should clear the filter and keep Chrome open, got %d rows", len(m.apps.rows))
	}
	m = press(m, "1")
	if m.view != OverviewView {
		t.Fatal("1 should switch to Overview")
	}
	if !strings.Contains(ansi.Strip(m.View().Content), "1-5 view") {
		t.Fatal("overview footer should list the view keys")
	}
}

func TestTabBarShowsPositionAndPause(t *testing.T) {
	m := testModel(t)
	m = press(m, "down")
	bar := ansi.Strip(m.View().Content)
	if !strings.Contains(bar, "rows 1–3 of 3") || !strings.Contains(bar, "sort paused") {
		t.Fatalf("tab bar should show the position and the pause, got %q", bar)
	}
}
