//go:build !darwin

package proc

import "github.com/shirou/gopsutil/v4/process"

func readStatus(p *process.Process) (string, error) {
	status, err := p.Status()
	if err != nil {
		return "", err
	}
	if len(status) == 0 {
		return "", nil
	}
	return status[0], nil
}
