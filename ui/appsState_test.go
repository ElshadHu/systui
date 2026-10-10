package ui

import (
	"testing"
	"time"

	"github.com/ElshadHu/systui/internal/apps"
)

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

func newState() appsState {
	var s appsState
	s.setGroups(sampleGroups(), time.Unix(0, 0))
	return s
}

func rowIDs(rows []row) []rowID {
	ids := make([]rowID, len(rows))
	for i, r := range rows {
		ids[i] = r.id
	}
	return ids
}

func TestGroupsStartCollapsedAndExpandOneAtATime(t *testing.T) {
	s := newState()
	if len(s.rows) != 3 || s.selected != (rowID{group: "chrome"}) {
		t.Fatalf("want 3 collapsed rows with the first selected, got %v selected %v", rowIDs(s.rows), s.selected)
	}
	s.expand()
	if len(s.rows) != 3+maxShownMembers+1 || s.rows[6].kind != moreRow || s.rows[6].hidden != 2 {
		t.Fatalf("chrome should show 5 members and a more row, got %v", rowIDs(s.rows))
	}
	s.moveTo(7)
	s.expand()
	if s.expanded != "code" || len(s.rows) != 5 {
		t.Fatalf("expanding VS Code should fold Chrome, got %v", rowIDs(s.rows))
	}
	s.moveTo(2)
	s.collapse()
	if len(s.rows) != 3 || s.selected != (rowID{group: "code"}) {
		t.Fatalf("collapse should fold the group and select its row, got %v selected %v", rowIDs(s.rows), s.selected)
	}
}

func TestMoreRowShowsTheRest(t *testing.T) {
	s := newState()
	s.expand()
	s.moveTo(6)
	s.expand()
	if len(s.rows) != 3+7 || s.showAll != "chrome" {
		t.Fatalf("more row should reveal every member, got %v", rowIDs(s.rows))
	}
}

func TestExpandAllAndCollapseAll(t *testing.T) {
	s := newState()
	s.expandAll()
	if len(s.rows) != 3+6+2+1 {
		t.Fatalf("expand all should open every group, got %d rows", len(s.rows))
	}
	s.moveTo(8)
	s.collapseAll()
	if len(s.rows) != 3 || s.selected != (rowID{group: "code"}) {
		t.Fatalf("collapse all should keep the cursor on the group it was in, got %v", s.selected)
	}
}

func TestSelectionFollowsIdentityAcrossRefresh(t *testing.T) {
	s := newState()
	s.expand()
	s.moveTo(2)
	if s.selected != (rowID{group: "chrome", pid: 11}) {
		t.Fatalf("unexpected selection %v", s.selected)
	}
	groups := sampleGroups()
	groups[0].Members[1], groups[0].Members[2] = groups[0].Members[2], groups[0].Members[1]
	s.setGroups(groups, time.Unix(1, 0))
	if s.selected != (rowID{group: "chrome", pid: 11}) || s.selectedIndex() != 3 {
		t.Fatalf("cursor should follow PID 11 to its new row, got %v at %d", s.selected, s.selectedIndex())
	}
}

func TestCursorClimbsToGroupWhenProcessExits(t *testing.T) {
	s := newState()
	s.expand()
	s.moveTo(2)
	groups := sampleGroups()
	groups[0].Members = groups[0].Members[:1]
	s.setGroups(groups, time.Unix(1, 0))
	if s.selected != (rowID{group: "chrome"}) {
		t.Fatalf("cursor should move to the group row, got %v", s.selected)
	}
}

func TestCursorKeepsPositionWhenGroupExits(t *testing.T) {
	s := newState()
	s.moveTo(1)
	groups := sampleGroups()
	s.setGroups([]apps.Group{groups[0], groups[2]}, time.Unix(1, 0))
	if s.selected != (rowID{group: "loom"}) {
		t.Fatalf("cursor should land on the row now in that position, got %v", s.selected)
	}
}

func TestSortPauseKeepsRowOrder(t *testing.T) {
	s := newState()
	now := time.Unix(10, 0)
	s.freeze(now)
	reversed := sampleGroups()
	reversed[0], reversed[2] = reversed[2], reversed[0]
	s.setGroups(reversed, now.Add(time.Second))
	if s.rows[0].id.group != "chrome" {
		t.Fatalf("order should hold while paused, got %v", rowIDs(s.rows))
	}
	reversed = sampleGroups()
	reversed[0], reversed[2] = reversed[2], reversed[0]
	s.setGroups(reversed, now.Add(sortPause))
	if s.rows[0].id.group != "loom" {
		t.Fatalf("order should follow the sample once the pause ends, got %v", rowIDs(s.rows))
	}
}

func TestFilterMatchesGroupsAndMembers(t *testing.T) {
	s := newState()
	s.setFilter("gpu")
	if len(s.rows) != 2 || s.rows[0].id.group != "chrome" || s.rows[1].id.pid != 13 {
		t.Fatalf("filter should keep Chrome open on its GPU helper, got %v", rowIDs(s.rows))
	}
	s.setFilter("loom")
	if len(s.rows) != 1 || s.rows[0].id.group != "loom" {
		t.Fatalf("filter should match group names, got %v", rowIDs(s.rows))
	}
	s.setFilter("")
	if len(s.rows) != 3 {
		t.Fatalf("clearing the filter should bring every group back, got %v", rowIDs(s.rows))
	}
}

func TestViewportSlidesOneRowAndCountsEdges(t *testing.T) {
	s := newState()
	s.expandAll()
	shown, above, below := s.viewport(5)
	if above != 0 || below != 8 || len(shown) != 4 {
		t.Fatalf("top of list: want 4 rows and 8 below, got %d rows, %d above, %d below", len(shown), above, below)
	}
	s.moveTo(4)
	shown, above, below = s.viewport(5)
	if above != 2 || below != 7 || len(shown) != 3 || shown[2].id != s.selected {
		t.Fatalf("cursor past the edge should slide by one row, got %d rows, %d above, %d below", len(shown), above, below)
	}
	s.moveTo(11)
	shown, above, below = s.viewport(5)
	if above != 8 || below != 0 || len(shown) != 4 || shown[3].id != s.selected {
		t.Fatalf("bottom of list: got %d rows, %d above, %d below", len(shown), above, below)
	}
	first, last := s.position(shown)
	if first != 9 || last != 12 {
		t.Fatalf("want rows 9–12, got %d–%d", first, last)
	}
}
