package apps

import "time"

const HistoryLength = 20

// History is the last HistoryLength CPU and memory readings of one group,
// and the highest memory seen since the group appeared
type History struct {
	CPU    []float64
	Memory []uint64
	Peak   uint64
	PeakAt time.Time
}

func (h *History) record(cpu float64, memory uint64, now time.Time) {
	h.CPU = keepLast(h.CPU, cpu)
	h.Memory = keepLast(h.Memory, memory)
	if memory > h.Peak {
		h.Peak, h.PeakAt = memory, now
	}
}

func keepLast[T any](list []T, v T) []T {
	if len(list) < HistoryLength {
		return append(list, v)
	}
	copy(list, list[1:])
	list[HistoryLength-1] = v
	return list
}

// Histories keeps a History per group across samples
type Histories struct {
	byGroup map[string]*History
}

func NewHistories() *Histories {
	return &Histories{byGroup: make(map[string]*History)}
}

// Record adds the current CPU and memory of every group, stamps each group
// with its peak, and forgets groups that are gone
func (t *Histories) Record(groups []Group, now time.Time) {
	seen := make(map[string]bool, len(groups))
	for i := range groups {
		g := &groups[i]
		seen[g.ID] = true
		h, ok := t.byGroup[g.ID]
		if !ok {
			h = &History{}
			t.byGroup[g.ID] = h
		}
		h.record(g.CPU, g.Memory, now)
		g.PeakMemory, g.PeakAt = h.Peak, h.PeakAt
	}
	for id := range t.byGroup {
		if !seen[id] {
			delete(t.byGroup, id)
		}
	}
}

// Of returns the readings of one group
func (t *Histories) Of(id string) History {
	h, ok := t.byGroup[id]
	if !ok {
		return History{}
	}
	return *h
}
