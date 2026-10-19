//go:build linux

package proc

import (
	"os"
	"strconv"

	"github.com/shirou/gopsutil/v4/process"
)

// readUsage adds the peak resident size and the reaped children's CPU time
func readUsage(p *process.Process) usage {
	u := portableUsage(p)
	pid := strconv.Itoa(int(p.Pid))
	if status, err := os.ReadFile("/proc/" + pid + "/status"); err == nil {
		if peak, err := parsePeakRSS(string(status)); err == nil {
			u.PeakMemory = peak
		}
	}
	if stat, err := os.ReadFile("/proc/" + pid + "/stat"); err == nil {
		if children, err := parseChildrenCPUTime(string(stat)); err == nil {
			u.CPUTime += children
		}
	}
	return u
}
