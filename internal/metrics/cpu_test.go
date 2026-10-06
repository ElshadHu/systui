package metrics

import (
	"testing"

	"github.com/shirou/gopsutil/v4/cpu"
)

func TestUsagePercent(t *testing.T) {
	prev := cpu.TimesStat{User: 100, System: 50, Idle: 850}
	cur := cpu.TimesStat{User: 160, System: 70, Idle: 870}
	got := usagePercent(prev, cur)
	if got != 80 {
		t.Fatalf("want 80, got %v", got)
	}
}

func TestTotalIgnoresGuest(t *testing.T) {
	s := cpu.TimesStat{User: 10, Nice: 1, System: 2, Idle: 3,
		Iowait: 4, Irq: 5, Softirq: 6, Steal: 7, Guest: 100, GuestNice: 100}
	if got := total(s); got != 38 {
		t.Fatalf("want 38, got %v", got)
	}
}

func TestUsagePercentZeroDelta(t *testing.T) {
	s := cpu.TimesStat{User: 1, Idle: 1}
	if got := usagePercent(s, s); got != 0 {
		t.Fatalf("want 0, got %v", got)
	}
}
