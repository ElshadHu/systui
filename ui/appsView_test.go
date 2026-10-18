package ui

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/ElshadHu/systui/internal/apps"
	"github.com/ElshadHu/systui/ui/applist"
	"github.com/ElshadHu/systui/ui/theme"
	"github.com/charmbracelet/x/ansi"
)

func TestFitAppsColumnsDropsFromTheRight(t *testing.T) {
	cases := []struct {
		width        int
		procs, spark bool
	}{
		{120, true, true},
		{60, true, false},
		{45, false, false},
	}
	for _, c := range cases {
		got := fitAppsColumns(c.width)
		if got.has(procsField) != c.procs || got.has(sparkField) != c.spark || got.name < minNameColumn {
			t.Errorf("width %d: got %+v", c.width, got)
		}
	}
}

func sampleGroups() []apps.Group {
	return []apps.Group{
		{ID: "chrome", Name: "Google Chrome", Kind: apps.App, Memory: 4 << 30, Members: []apps.Member{
			{PID: 10, Label: "Google Chrome", Main: true}, {PID: 11, Label: "Renderer"}, {PID: 12, Label: "Renderer"},
			{PID: 13, Label: "GPU Process"}, {PID: 14, Label: "Network Service"}, {PID: 15, Label: "Utility"}, {PID: 16, Label: "Utility"},
		}},
		{ID: "code", Name: "VS Code", Kind: apps.App, Memory: 2 << 30, Members: []apps.Member{{PID: 20, Label: "Electron", Main: true}, {PID: 21, Label: "Plugin"}}},
		{ID: "loom", Name: "Loom", Kind: apps.App, Memory: 1 << 30, Members: []apps.Member{{PID: 30, Label: "Loom", Main: true}}},
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
	m.apps.SetGroups(sampleGroups(), time.Unix(0, 0))
	return m
}

func TestRowsNeverExceedTheTerminalWidth(t *testing.T) {
	m := testModel(t)
	m.apps.Expand()
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
	if m.apps.Selected != (applist.RowID{Group: "chrome", PID: 11}) {
		t.Fatalf("right then down twice should land on the second member, got %v", m.apps.Selected)
	}
	m = press(m, "c")
	if m.apps.Sort() != 1 || !strings.Contains(ansi.Strip(m.appsHeader(fitAppsColumns(80))), "CPU "+m.theme.Glyphs.Expanded) {
		t.Fatal("c should sort by CPU and mark the column")
	}
	m = press(m, "/", "l", "o", "o", "enter")
	if m.apps.Typing || m.apps.Filter() != "loo" || len(m.apps.Rows) != 1 {
		t.Fatalf("typing a filter should narrow the rows, got %q typing=%v rows=%d", m.apps.Filter(), m.apps.Typing, len(m.apps.Rows))
	}
	m = press(m, "esc")
	if m.apps.Filter() != "" || len(m.apps.Rows) != 9 {
		t.Fatalf("esc should clear the filter and keep Chrome open, got %d rows", len(m.apps.Rows))
	}
	m = press(m, "2")
	if m.view != OverviewView {
		t.Fatal("2 should switch to Overview")
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

func click(m Model, x, y int) Model {
	next, _ := m.Update(tea.MouseClickMsg{X: x, Y: y, Button: tea.MouseLeft})
	return next.(Model)
}

func wheel(m Model, button tea.MouseButton) Model {
	next, _ := m.Update(tea.MouseWheelMsg{Button: button})
	return next.(Model)
}

func TestMouseDrivesTheAppsView(t *testing.T) {
	m := testModel(t)
	tabLine, headerLine, firstRow := m.tabBarLine(), m.tabBarLine()+1, m.tabBarLine()+2
	cols := fitAppsColumns(m.width)

	m = click(m, firstRow+1, firstRow+1)
	if m.apps.Selected != (applist.RowID{Group: "code"}) {
		t.Fatalf("clicking the second row should select VS Code, got %v", m.apps.Selected)
	}
	m = click(m, 0, firstRow+1)
	if !m.apps.Open("code") || len(m.apps.Rows) != 5 {
		t.Fatalf("a second click on a group should open it, got rows=%d", len(m.apps.Rows))
	}
	m = click(m, 0, firstRow+1)
	if m.apps.Open("code") || len(m.apps.Rows) != 3 {
		t.Fatalf("a third click should close it again, got rows=%d", len(m.apps.Rows))
	}

	m = click(m, 0, firstRow)
	m = click(m, 0, firstRow)
	m = click(m, 0, firstRow+6)
	if len(m.apps.Rows) != 10 || m.apps.Rows[len(m.apps.Rows)-1].Kind == applist.MoreRow {
		t.Fatalf("one click on the more row should show every member, got rows=%d", len(m.apps.Rows))
	}

	m = wheel(m, tea.MouseWheelDown)
	m = wheel(m, tea.MouseWheelDown)
	if m.apps.Selected != (applist.RowID{Group: "chrome", PID: 11}) {
		t.Fatalf("two wheel notches down should move two rows, got %v", m.apps.Selected)
	}
	m = wheel(m, tea.MouseWheelUp)
	if m.apps.Selected != (applist.RowID{Group: "chrome", PID: 10}) || !m.apps.Frozen(m.now) {
		t.Fatalf("wheel up should move back one row and pause the sort, got %v", m.apps.Selected)
	}

	m = click(m, markerColumn+cols.name+columnGap+memColumn+columnGap, headerLine)
	if m.apps.Sort() != 1 {
		t.Fatal("clicking the CPU header should sort by CPU")
	}
	m = click(m, markerColumn+cols.name+columnGap, headerLine)
	if m.apps.Sort() != 0 {
		t.Fatal("clicking the MEM header should sort by memory")
	}
	m = click(m, 0, headerLine)
	if m.apps.Sort() != 0 {
		t.Fatal("clicking the APP header should change nothing")
	}

	m = click(m, tabWidth(AppsView)+len(tabGap), tabLine)
	if m.view != OverviewView {
		t.Fatal("clicking the second tab should switch to Overview")
	}
	m = click(m, 0, tabLine)
	if m.view != AppsView {
		t.Fatal("clicking the first tab should switch back to Apps")
	}
	m = click(m, m.width-1, tabLine)
	if m.view != AppsView {
		t.Fatal("clicking past the tabs should change nothing")
	}
}

func TestRowAtSkipsTheAboveMarker(t *testing.T) {
	m := testModel(t)
	m.height = 8
	m.apps.ExpandAll()
	m.apps.MoveTo(len(m.apps.Rows) - 1)
	shown, above, _ := m.apps.Viewport(m.listHeight())
	if above == 0 {
		t.Fatal("the cursor at the end should leave rows above the viewport")
	}
	first := m.tabBarLine() + 3
	if r, ok := m.rowAt(first); !ok || r.ID != shown[0].ID {
		t.Fatalf("the line under the marker should be the first shown row, got %v", r.ID)
	}
	if _, ok := m.rowAt(first - 1); ok {
		t.Fatal("the marker line is not a row")
	}
}
