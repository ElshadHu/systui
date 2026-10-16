//go:build darwin

package metrics

import (
	"errors"
	"sync"
	"time"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/shirou/gopsutil/v4/process"
)

const (
	systemLibrary = "/usr/lib/libSystem.B.dylib"
	rusageInfoV4  = 4
)

type rusageInfo struct {
	UUID                      [16]byte
	UserTime                  uint64
	SystemTime                uint64
	PkgIdleWakeups            uint64
	InterruptWakeups          uint64
	Pageins                   uint64
	WiredSize                 uint64
	ResidentSize              uint64
	PhysFootprint             uint64
	ProcStartAbstime          uint64
	ProcExitAbstime           uint64
	ChildUserTime             uint64
	ChildSystemTime           uint64
	ChildPkgIdleWakeups       uint64
	ChildInterruptWkups       uint64
	ChildPageins              uint64
	ChildElapsedAbstime       uint64
	DiskIOBytesRead           uint64
	DiskIOBytesWritten        uint64
	CPUTimeQosDefault         uint64
	CPUTimeQosMaintenance     uint64
	CPUTimeQosBackground      uint64
	CPUTimeQosUtility         uint64
	CPUTimeQosLegacy          uint64
	CPUTimeQosUserInitiated   uint64
	CPUTimeQosUserInteractive uint64
	BilledSystemTime          uint64
	ServicedSystemTime        uint64
	LogicalWrites             uint64
	LifetimeMaxPhysFootprint  uint64
	Instructions              uint64
	Cycles                    uint64
	BilledEnergy              uint64
	ServicedEnergy            uint64
	IntervalMaxPhysFootprint  uint64
	RunnableTime              uint64
}

type machTimebase struct {
	Numer uint32
	Denom uint32
}

var (
	procPidRusage    func(pid int32, flavor int32, buffer unsafe.Pointer) int32
	machTimebaseInfo func(info *machTimebase) int32
	errRusage        = errors.New("proc_pid_rusage failed")

	loadRusage = sync.OnceValues(func() (machTimebase, error) {
		handle, err := purego.Dlopen(systemLibrary, purego.RTLD_LAZY|purego.RTLD_GLOBAL)
		if err != nil {
			return machTimebase{}, err
		}
		purego.RegisterLibFunc(&procPidRusage, handle, "proc_pid_rusage")
		purego.RegisterLibFunc(&machTimebaseInfo, handle, "mach_timebase_info")
		var timebase machTimebase
		if machTimebaseInfo(&timebase) != 0 || timebase.Denom == 0 {
			return machTimebase{}, errors.New("mach_timebase_info failed")
		}
		return timebase, nil
	})
)

// readUsage takes the footprint, its lifetime peak and the CPU time
func readUsage(p *process.Process) usage {
	u, err := readRusage(p.Pid)
	if err != nil {
		return portableUsage(p)
	}
	return u
}

func readRusage(pid int32) (usage, error) {
	timebase, err := loadRusage()
	if err != nil {
		return usage{}, err
	}
	var info rusageInfo
	if procPidRusage(pid, rusageInfoV4, unsafe.Pointer(&info)) != 0 {
		return usage{}, errRusage
	}
	ticks := info.UserTime + info.SystemTime + info.ChildUserTime + info.ChildSystemTime
	return usage{
		Memory: info.PhysFootprint,
		Lifetime: Lifetime{
			PeakMemory: info.LifetimeMaxPhysFootprint,
			CPUTime:    time.Duration(ticks * uint64(timebase.Numer) / uint64(timebase.Denom)),
		},
	}, nil
}
