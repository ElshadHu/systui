package apps

import (
	"testing"
	"time"
)

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

func sorted(groups []Group, key SortKey) []Group {
	Sort(groups, key)
	return groups
}

func TestSortByPeakAndCPUTimeFallBackToTheLiveColumns(t *testing.T) {
	groups := []Group{
		{ID: "steady", Memory: 2 << 30, PeakMemory: 2 << 30, CPU: 1, CPUTime: 3 * time.Hour},
		{ID: "leaky", Memory: 1 << 30, PeakMemory: 4 << 30, CPU: 1, CPUTime: 3 * time.Hour},
		{ID: "burst", Memory: 1 << 30, PeakMemory: 4 << 30, CPU: 40, CPUTime: 2 * time.Minute},
	}
	Sort(groups, ByPeak)
	if groups[0].ID != "burst" || groups[1].ID != "leaky" || groups[2].ID != "steady" {
		t.Fatalf("peak first, then memory, then CPU, got %v %v %v", groups[0].ID, groups[1].ID, groups[2].ID)
	}
	Sort(groups, ByCPUTime)
	if groups[0].ID != "steady" || groups[1].ID != "leaky" || groups[2].ID != "burst" {
		t.Fatalf("CPU time first, then memory, got %v %v %v", groups[0].ID, groups[1].ID, groups[2].ID)
	}
	if ByPeak.Family() != ByMemory || ByCPUTime.Family() != ByCPU || ByCPU.Family() != ByCPU {
		t.Fatal("peak belongs with memory and CPU time with CPU")
	}
}
