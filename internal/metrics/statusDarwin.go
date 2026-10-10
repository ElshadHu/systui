//go:build darwin

package metrics

import (
	"github.com/shirou/gopsutil/v4/process"
	"golang.org/x/sys/unix"
)

// p_stat values from sys/proc.h
const (
	stateIdle     = 1
	stateRunning  = 2
	stateSleeping = 3
	stateStopped  = 4
	stateZombie   = 5
)

// readStatus asks the kernel directly
func readStatus(p *process.Process) (string, error) {
	k, err := unix.SysctlKinfoProc("kern.proc.pid", int(p.Pid))
	if err != nil {
		return "", err
	}
	switch k.Proc.P_stat {
	case stateIdle:
		return process.Idle, nil
	case stateRunning:
		return process.Running, nil
	case stateSleeping:
		return process.Sleep, nil
	case stateStopped:
		return process.Stop, nil
	case stateZombie:
		return process.Zombie, nil
	}
	return "", nil
}
