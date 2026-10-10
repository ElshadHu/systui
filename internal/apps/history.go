package apps

import "time"

const (
	historyLength = 60
	historyStep   = 10 * time.Second
)

// History is a ring of memory samples, one per historyStep, oldest first in Values
type History struct {
	samples [historyLength]uint64
	count   int
	next    int
	last    time.Time
}

func (h *History) record(now time.Time, memory uint64) {
	if !h.last.IsZero() && now.Sub(h.last) < historyStep {
		return
	}
	h.samples[h.next] = memory
	h.next = (h.next + 1) % historyLength
	h.count = min(h.count+1, historyLength)
	h.last = now
}

func (h *History) Values() []uint64 {
	out := make([]uint64, 0, h.count)
	start := (h.next - h.count + historyLength) % historyLength
	for i := 0; i < h.count; i++ {
		out = append(out, h.samples[(start+i)%historyLength])
	}
	return out
}

// Trends keeps a memory History per group across samples
type Trends struct {
	byGroup map[string]*History
}

func NewTrends() *Trends {
	return &Trends{byGroup: make(map[string]*History)}
}

// Record adds the current memory of every group and forgets groups that are gone
func (t *Trends) Record(groups []Group, now time.Time) {
	seen := make(map[string]bool, len(groups))
	for _, g := range groups {
		seen[g.ID] = true
		h, ok := t.byGroup[g.ID]
		if !ok {
			h = &History{}
			t.byGroup[g.ID] = h
		}
		h.record(now, g.Memory)
	}
	for id := range t.byGroup {
		if !seen[id] {
			delete(t.byGroup, id)
		}
	}
}

// Of returns the recorded memory of one group, oldest first
func (t *Trends) Of(id string) []uint64 {
	h, ok := t.byGroup[id]
	if !ok {
		return nil
	}
	return h.Values()
}
