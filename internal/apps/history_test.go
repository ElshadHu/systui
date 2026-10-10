package apps

import (
	"testing"
	"time"
)

func TestTrendsRecordEveryStepAndForgetGoneGroups(t *testing.T) {
	tr := NewTrends()
	start := time.Unix(1000, 0)
	tr.Record([]Group{{ID: "a", Memory: 1}, {ID: "b", Memory: 9}}, start)
	tr.Record([]Group{{ID: "a", Memory: 2}}, start.Add(3*time.Second))
	tr.Record([]Group{{ID: "a", Memory: 3}}, start.Add(historyStep))
	got := tr.Of("a")
	if len(got) != 2 || got[0] != 1 || got[1] != 3 {
		t.Fatalf("want [1 3], got %v", got)
	}
	if tr.Of("b") != nil {
		t.Fatal("vanished group should be forgotten")
	}
}

func TestHistoryRingWrapsOldestFirst(t *testing.T) {
	var h History
	now := time.Unix(0, 0)
	for i := 0; i < historyLength+5; i++ {
		h.record(now.Add(time.Duration(i)*historyStep), uint64(i))
	}
	got := h.Values()
	if len(got) != historyLength || got[0] != 5 || got[historyLength-1] != historyLength+4 {
		t.Fatalf("ring should hold the last %d values oldest first, got %v", historyLength, got)
	}
}
