package proc

import (
	"os"

	"github.com/shirou/gopsutil/v4/process"
)

func Kill(pid int32) error {
	return signal(pid, (*process.Process).Kill)
}

func Terminate(pid int32) error {
	return signal(pid, (*process.Process).Terminate)
}

func Suspend(pid int32) error {
	return signal(pid, (*process.Process).Suspend)
}

func Resume(pid int32) error {
	return signal(pid, (*process.Process).Resume)
}

func signal(pid int32, send func(*process.Process) error) error {
	if int(pid) == os.Getpid() {
		return ErrOwnProcess
	}
	p, err := process.NewProcess(pid)
	if err != nil {
		return err
	}
	return send(p)
}
