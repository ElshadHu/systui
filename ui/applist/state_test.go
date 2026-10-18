package applist

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

func newState() State {
	var s State
	s.SetGroups(sampleGroups(), time.Unix(0, 0))
	return s
}

func rowIDs(rows []Row) []RowID {
	ids := make([]RowID, len(rows))
	for i, r := range rows {
		ids[i] = r.ID
	}
	return ids
}

func TestGroupsStartCollapsedAndExpandOneAtATime(t *testing.T) {
	s := newState()
	if len(s.Rows) != 3 || s.Selected != (RowID{Group: "chrome"}) {
		t.Fatalf("want 3 collapsed rows with the first selected, got %v selected %v", rowIDs(s.Rows), s.Selected)
	}
	s.Expand()
	if len(s.Rows) != 3+maxShownMembers+1 || s.Rows[6].Kind != MoreRow || s.Rows[6].Hidden != 2 {
		t.Fatalf("chrome should show 5 members and a more row, got %v", rowIDs(s.Rows))
	}
	s.MoveTo(7)
	s.Expand()
	if s.expanded != "code" || len(s.Rows) != 5 {
		t.Fatalf("expanding VS Code should fold Chrome, got %v", rowIDs(s.Rows))
	}
	s.MoveTo(2)
	s.Collapse()
	if len(s.Rows) != 3 || s.Selected != (RowID{Group: "code"}) {
		t.Fatalf("collapse should fold the group and select its row, got %v selected %v", rowIDs(s.Rows), s.Selected)
	}
}

func TestMoreRowShowsTheRest(t *testing.T) {
	s := newState()
	s.Expand()
	s.MoveTo(6)
	s.Expand()
	if len(s.Rows) != 3+7 || s.showAll != "chrome" {
		t.Fatalf("more row should reveal every member, got %v", rowIDs(s.Rows))
	}
}

func TestExpandAllAndCollapseAll(t *testing.T) {
	s := newState()
	s.ExpandAll()
	if len(s.Rows) != 3+6+2+1 {
		t.Fatalf("expand all should open every group, got %d rows", len(s.Rows))
	}
	s.MoveTo(8)
	s.CollapseAll()
	if len(s.Rows) != 3 || s.Selected != (RowID{Group: "code"}) {
		t.Fatalf("collapse all should keep the cursor on the group it was in, got %v", s.Selected)
	}
}

func TestSelectionFollowsIdentityAcrossRefresh(t *testing.T) {
	s := newState()
	s.Expand()
	s.MoveTo(2)
	if s.Selected != (RowID{Group: "chrome", PID: 11}) {
		t.Fatalf("unexpected selection %v", s.Selected)
	}
	groups := sampleGroups()
	groups[0].Members[1], groups[0].Members[2] = groups[0].Members[2], groups[0].Members[1]
	s.SetGroups(groups, time.Unix(1, 0))
	if s.Selected != (RowID{Group: "chrome", PID: 11}) || s.selectedIndex() != 3 {
		t.Fatalf("cursor should follow PID 11 to its new row, got %v at %d", s.Selected, s.selectedIndex())
	}
}

func TestCursorClimbsToGroupWhenProcessExits(t *testing.T) {
	s := newState()
	s.Expand()
	s.MoveTo(2)
	groups := sampleGroups()
	groups[0].Members = groups[0].Members[:1]
	s.SetGroups(groups, time.Unix(1, 0))
	if s.Selected != (RowID{Group: "chrome"}) {
		t.Fatalf("cursor should move to the group row, got %v", s.Selected)
	}
}

func TestCursorKeepsPositionWhenGroupExits(t *testing.T) {
	s := newState()
	s.MoveTo(1)
	groups := sampleGroups()
	s.SetGroups([]apps.Group{groups[0], groups[2]}, time.Unix(1, 0))
	if s.Selected != (RowID{Group: "loom"}) {
		t.Fatalf("cursor should land on the row now in that position, got %v", s.Selected)
	}
}

func TestSortPauseKeepsRowOrder(t *testing.T) {
	s := newState()
	now := time.Unix(10, 0)
	s.Freeze(now)
	reversed := sampleGroups()
	reversed[0], reversed[2] = reversed[2], reversed[0]
	s.SetGroups(reversed, now.Add(time.Second))
	if s.Rows[0].ID.Group != "chrome" {
		t.Fatalf("order should hold while paused, got %v", rowIDs(s.Rows))
	}
	reversed = sampleGroups()
	reversed[0], reversed[2] = reversed[2], reversed[0]
	s.SetGroups(reversed, now.Add(sortPause))
	if s.Rows[0].ID.Group != "loom" {
		t.Fatalf("order should follow the sample once the pause ends, got %v", rowIDs(s.Rows))
	}
}

func TestFilterMatchesGroupsAndMembers(t *testing.T) {
	s := newState()
	s.SetFilter("gpu")
	if len(s.Rows) != 2 || s.Rows[0].ID.Group != "chrome" || s.Rows[1].ID.PID != 13 {
		t.Fatalf("filter should keep Chrome open on its GPU helper, got %v", rowIDs(s.Rows))
	}
	s.SetFilter("loom")
	if len(s.Rows) != 1 || s.Rows[0].ID.Group != "loom" {
		t.Fatalf("filter should match group names, got %v", rowIDs(s.Rows))
	}
	s.SetFilter("")
	if len(s.Rows) != 3 {
		t.Fatalf("clearing the filter should bring every group back, got %v", rowIDs(s.Rows))
	}
}

func TestViewportSlidesOneRowAndCountsEdges(t *testing.T) {
	s := newState()
	s.ExpandAll()
	shown, above, below := s.Viewport(5)
	if above != 0 || below != 8 || len(shown) != 4 {
		t.Fatalf("top of list: want 4 rows and 8 below, got %d rows, %d above, %d below", len(shown), above, below)
	}
	s.MoveTo(4)
	shown, above, below = s.Viewport(5)
	if above != 2 || below != 7 || len(shown) != 3 || shown[2].ID != s.Selected {
		t.Fatalf("cursor past the edge should slide by one row, got %d rows, %d above, %d below", len(shown), above, below)
	}
	s.MoveTo(11)
	shown, above, below = s.Viewport(5)
	if above != 8 || below != 0 || len(shown) != 4 || shown[3].ID != s.Selected {
		t.Fatalf("bottom of list: got %d rows, %d above, %d below", len(shown), above, below)
	}
	first, last := s.Position(shown)
	if first != 9 || last != 12 {
		t.Fatalf("want rows 9–12, got %d–%d", first, last)
	}
}
