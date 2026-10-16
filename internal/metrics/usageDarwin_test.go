//go:build darwin

package metrics

import (
	"os"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

func TestReadRusageMatchesGopsutilAndKeepsThePeak(t *testing.T) {
	p, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		t.Fatal(err)
	}
	got, err := readRusage(p.Pid)
	if err != nil {
		t.Fatal(err)
	}
	times, err := p.Times()
	if err != nil {
		t.Fatal(err)
	}
	own := time.Duration((times.User + times.System) * float64(time.Second))
	if got.CPUTime < own || got.CPUTime > own+time.Second {
		t.Fatalf("rusage CPU time %v should be on the same scale as gopsutil's %v", got.CPUTime, own)
	}
	if got.Memory == 0 || got.PeakMemory < got.Memory {
		t.Fatalf("peak %d should be at least the footprint %d", got.PeakMemory, got.Memory)
	}
}
