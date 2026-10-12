package apps

import "testing"

func TestHistoriesRecordBothMetricsAndForgetGoneGroups(t *testing.T) {
	hs := NewHistories()
	hs.Record([]Group{{ID: "a", CPU: 1, Memory: 10}, {ID: "b", CPU: 9, Memory: 90}})
	hs.Record([]Group{{ID: "a", CPU: 2, Memory: 20}})
	got := hs.Of("a")
	if len(got.CPU) != 2 || got.CPU[0] != 1 || got.CPU[1] != 2 {
		t.Fatalf("want CPU [1 2], got %v", got.CPU)
	}
	if len(got.Memory) != 2 || got.Memory[0] != 10 || got.Memory[1] != 20 {
		t.Fatalf("want memory [10 20], got %v", got.Memory)
	}
	if gone := hs.Of("b"); gone.CPU != nil || gone.Memory != nil {
		t.Fatal("vanished group should be forgotten")
	}
}

func TestHistoryKeepsTheLastReadingsOldestFirst(t *testing.T) {
	var h History
	for i := 0; i < HistoryLength+5; i++ {
		h.record(float64(i), uint64(i))
	}
	if len(h.CPU) != HistoryLength || h.CPU[0] != 5 || h.CPU[HistoryLength-1] != HistoryLength+4 {
		t.Fatalf("want the last %d values oldest first, got %v", HistoryLength, h.CPU)
	}
	if len(h.Memory) != HistoryLength || h.Memory[0] != 5 {
		t.Fatalf("memory should match CPU, got %v", h.Memory)
	}
}
