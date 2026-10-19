//go:build windows

package proc

import (
	"unsafe"

	"github.com/shirou/gopsutil/v4/process"
	"golang.org/x/sys/windows"
)

type processMemoryCounters struct {
	Size                       uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
}

var getProcessMemoryInfo = windows.NewLazySystemDLL("psapi.dll").NewProc("GetProcessMemoryInfo")

// readUsage adds the peak working set, which gopsutil does not expose.
func readUsage(p *process.Process) usage {
	u := portableUsage(p)
	if peak, err := readPeakWorkingSet(p.Pid); err == nil {
		u.PeakMemory = peak
	}
	return u
}

func readPeakWorkingSet(pid int32) (uint64, error) {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(handle)
	var counters processMemoryCounters
	counters.Size = uint32(unsafe.Sizeof(counters))
	ok, _, err := getProcessMemoryInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&counters)), uintptr(counters.Size))
	if ok == 0 {
		return 0, err
	}
	return uint64(counters.PeakWorkingSetSize), nil
}
