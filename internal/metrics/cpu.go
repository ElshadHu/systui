package metrics

import (
	"errors"

	"github.com/shirou/gopsutil/v4/cpu"
)

// CPUSampler remembers the last CPU time counters
type CPUSampler struct {
	prev cpu.TimesStat
}

// NewCPUSampler creates a CPUSampler
func NewCPUSampler() (*CPUSampler, error) {
	cur, err := readTotal()
	if err != nil {
		return nil, err
	}
	return &CPUSampler{prev: cur}, nil
}

// Usage returns total CPU busy percentage since the prev call
func (s *CPUSampler) Usage() (float64, error) {
	cur, err := readTotal()
	if err != nil {
		return 0, err
	}
	pct := usagePercent(s.prev, cur)
	s.prev = cur
	return pct, nil
}

// readTotal returns the aggregated counters for all cores
func readTotal() (cpu.TimesStat, error) {
	stats, err := cpu.Times(false)
	if err != nil {
		return cpu.TimesStat{}, err
	}

	if len(stats) == 0 {
		return cpu.TimesStat{}, errors.New("cpu.Times returned no data")
	}
	return stats[0], nil
}

// total sums the eight CPU time counters
func total(t cpu.TimesStat) float64 {
	return t.User + t.Nice + t.System + t.Idle + t.Iowait + t.Irq + t.Softirq + t.Steal
}

// usagePercent returns the percentage of time the CPU was busy
func usagePercent(prev, cur cpu.TimesStat) float64 {
	dTotal := total(cur) - total(prev)
	if dTotal <= 0 {
		return 0
	}

	dIdle := (cur.Idle + cur.Iowait) - (prev.Idle + prev.Iowait)
	busy := dTotal - dIdle

	pct := busy / dTotal * 100
	switch {
	case pct < 0:
		return 0
	case pct > 100:
		return 100
	}

	return pct
}
