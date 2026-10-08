package metrics

import (
	"errors"

	"github.com/shirou/gopsutil/v4/cpu"
)

// CPUStats holds percentages for one sampling interval
type CPUStats struct {
	Usage   float64
	User    float64
	Nice    float64
	System  float64
	Idle    float64
	Iowait  float64
	Irq     float64
	Softirq float64
	Steal   float64
	Guest   float64
}

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
func (s *CPUSampler) Sample() (CPUStats, error) {
	cur, err := readTotal()
	if err != nil {
		return CPUStats{}, err
	}
	stats := breakdown(s.prev, cur)
	s.prev = cur
	return stats, nil
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

// breakdown returns each counter's share of the time between 2 samples
func breakdown(prev, cur cpu.TimesStat) CPUStats {
	dTotal := total(cur) - total(prev)
	if dTotal <= 0 {
		return CPUStats{}
	}
	share := func(prev, cur float64) float64 {
		return percentLimit((cur - prev) / dTotal * 100)

	}
	s := CPUStats{
		User:    share(prev.User, cur.User),
		Nice:    share(prev.Nice, cur.Nice),
		System:  share(prev.System, cur.System),
		Idle:    share(prev.Idle, cur.Idle),
		Iowait:  share(prev.Iowait, cur.Iowait),
		Irq:     share(prev.Irq, cur.Irq),
		Softirq: share(prev.Softirq, cur.Softirq),
		Steal:   share(prev.Steal, cur.Steal),
		Guest:   share(prev.Guest+prev.GuestNice, cur.Guest+cur.GuestNice),
	}
	s.Usage = percentLimit(100 - s.Idle - s.Iowait)
	return s
}

func percentLimit(pct float64) float64 {
	switch {
	case pct < 0:
		return 0
	case pct > 100:
		return 100
	}
	return pct
}
