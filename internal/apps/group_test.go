package apps

import "testing"

func TestKeepOrderHoldsRowsAndAppendsNewGroups(t *testing.T) {
	prev := []Group{
		{ID: "a", Members: []Member{{PID: 1}, {PID: 2}}},
		{ID: "b"},
		{ID: SmallGroupID},
	}
	next := []Group{
		{ID: SmallGroupID},
		{ID: "c"},
		{ID: "b"},
		{ID: "a", Members: []Member{{PID: 3}, {PID: 2}, {PID: 1}}},
	}
	got := KeepOrder(prev, next)
	ids := [4]string{got[0].ID, got[1].ID, got[2].ID, got[3].ID}
	if ids != [4]string{"a", "b", "c", SmallGroupID} {
		t.Fatalf("want a b c small, got %v", ids)
	}
	pids := [3]int32{got[0].Members[0].PID, got[0].Members[1].PID, got[0].Members[2].PID}
	if pids != [3]int32{1, 2, 3} {
		t.Fatalf("members should keep their slots, new ones last, got %v", pids)
	}
}
