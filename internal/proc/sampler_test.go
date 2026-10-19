package proc

import (
	"os"
	"testing"
)

func TestSampleIncludesOwnProcess(t *testing.T) {
	s := NewSampler()
	stats, err := s.Sample()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range stats {
		if int(p.PID) == os.Getpid() {
			if p.Name == "" || p.User == "" || p.Started.IsZero() {
				t.Fatalf("own process row is incomplete: %+v", p)
			}
			return
		}
	}
	t.Fatalf("own PID %d missing from %d rows", os.Getpid(), len(stats))
}

func TestSampleDropsExitedProcesses(t *testing.T) {
	s := NewSampler()
	s.tracked[-1] = nil
	if _, err := s.Sample(); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.tracked[-1]; ok {
		t.Fatal("PID -1 should have been pruned")
	}
}

func TestSampleReadsOwnProcessStatics(t *testing.T) {
	s := NewSampler()
	stats, err := s.Sample()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range stats {
		if int(p.PID) != os.Getpid() {
			continue
		}
		if p.Exe == "" || p.Cmdline == "" || p.PPID == 0 || p.Memory == 0 {
			t.Fatalf("own process statics are incomplete: %+v", p)
		}
		return
	}
	t.Fatalf("own PID %d missing", os.Getpid())
}

func TestSampleReplacesReusedPID(t *testing.T) {
	s := NewSampler()
	pid := int32(os.Getpid())
	s.tracked[pid] = &trackedProcess{created: -1, stats: Stats{Name: "stale"}}
	if _, err := s.Sample(); err != nil {
		t.Fatal(err)
	}
	if s.tracked[pid].stats.Name == "stale" {
		t.Fatal("tracked entry with another create time should be replaced")
	}
}
