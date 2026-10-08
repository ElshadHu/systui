package metrics

import (
	"testing"

	"github.com/shirou/gopsutil/v4/cpu"
)

func TestBreakdown(t *testing.T) {
	prev := cpu.TimesStat{User: 100, System: 50, Idle: 850}
	cur := cpu.TimesStat{User: 160, System: 70, Idle: 870}
	got := breakdown(prev, cur)
	want := CPUStats{Usage: 80, User: 60, System: 20, Idle: 20}
	if got != want {
		t.Fatalf("want %+v, got %+v", want, got)
	}
}

func TestBreakdownGuestDoesNotChangeUsage(t *testing.T) {
	prev := cpu.TimesStat{User: 100, Idle: 100}
	cur := cpu.TimesStat{User: 150, Idle: 150, Guest: 25}
	got := breakdown(prev, cur)
	if got.Usage != 50 || got.Guest != 25 {
		t.Fatalf("want usage 50 guest 25, got %+v", got)
	}
}

func TestTotalIgnoresGuest(t *testing.T) {
	s := cpu.TimesStat{User: 10, Nice: 1, System: 2, Idle: 3,
		Iowait: 4, Irq: 5, Softirq: 6, Steal: 7, Guest: 100, GuestNice: 100}
	if got := total(s); got != 38 {
		t.Fatalf("want 38, got %v", got)
	}
}

func TestBreakdownZeroDelta(t *testing.T) {
	s := cpu.TimesStat{User: 1, Idle: 1}
	if got := breakdown(s, s); got != (CPUStats{}) {
		t.Fatalf("want zero stats, got %+v", got)
	}
}
