package metrics

import (
	"testing"
	"time"
)

func TestParsePeakRSSReadsTheHighWaterMark(t *testing.T) {
	status := "Name:\tnode\nVmRSS:\t  204800 kB\nVmHWM:\t 2097152 kB\nThreads:\t4\n"
	got, err := parsePeakRSS(status)
	if err != nil || got != 2<<30 {
		t.Fatalf("want 2 GB, got %d, %v", got, err)
	}
	if _, err := parsePeakRSS("Name:\tnode\n"); err == nil {
		t.Fatal("a status without VmHWM should fail")
	}
}

func TestParseChildrenCPUTimeSkipsTheCommandName(t *testing.T) {
	stat := "4242 (Web Content (x)) S 1 4242 4242 0 -1 4194560 100 200 0 0 150 50 3000 1500 20 0 4 0 100 0 0"
	got, err := parseChildrenCPUTime(stat)
	if err != nil || got != 45*time.Second {
		t.Fatalf("want 45s from 4500 ticks, got %v, %v", got, err)
	}
	if _, err := parseChildrenCPUTime("4242 (x) S 1"); err == nil {
		t.Fatal("a short stat line should fail")
	}
}
