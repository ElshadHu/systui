package apps

const HistoryLength = 20

// History is the last HistoryLength CPU and memory readings of one group
type History struct {
	CPU    []float64
	Memory []uint64
}

func (h *History) record(cpu float64, memory uint64) {
	h.CPU = keepLast(h.CPU, cpu)
	h.Memory = keepLast(h.Memory, memory)
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

// Record adds the current CPU and memory of every group and forgets groups that are gone
func (t *Histories) Record(groups []Group) {
	seen := make(map[string]bool, len(groups))
	for _, g := range groups {
		seen[g.ID] = true
		h, ok := t.byGroup[g.ID]
		if !ok {
			h = &History{}
			t.byGroup[g.ID] = h
		}
		h.record(g.CPU, g.Memory)
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
