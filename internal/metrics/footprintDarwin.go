//go:build darwin

package metrics

import (
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

const (
	systemLibrary = "/usr/lib/libSystem.B.dylib"
	rusageInfoV2  = 2
)

type rusageInfo struct {
	UUID                [16]byte
	UserTime            uint64
	SystemTime          uint64
	PkgIdleWakeups      uint64
	InterruptWakeups    uint64
	Pageins             uint64
	WiredSize           uint64
	ResidentSize        uint64
	PhysFootprint       uint64
	ProcStartAbstime    uint64
	ProcExitAbstime     uint64
	ChildUserTime       uint64
	ChildSystemTime     uint64
	ChildPkgIdleWakeups uint64
	ChildInterruptWkups uint64
	ChildPageins        uint64
	ChildElapsedAbstime uint64
	DiskIOBytesRead     uint64
	DiskIOBytesWritten  uint64
}

var (
	procPidRusage func(pid int32, flavor int32, buffer unsafe.Pointer) int32
	loadRusage    = sync.OnceValue(func() error {
		handle, err := purego.Dlopen(systemLibrary, purego.RTLD_LAZY|purego.RTLD_GLOBAL)
		if err != nil {
			return err
		}
		purego.RegisterLibFunc(&procPidRusage, handle, "proc_pid_rusage")
		return nil
	})
)

// readFootprint returns the physical footprint Activity Monitor shows.
func readFootprint(pid int32) (uint64, bool) {
	if loadRusage() != nil {
		return 0, false
	}
	var usage rusageInfo
	if procPidRusage(pid, rusageInfoV2, unsafe.Pointer(&usage)) != 0 {
		return 0, false
	}
	return usage.PhysFootprint, true
}
