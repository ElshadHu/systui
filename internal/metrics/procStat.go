package metrics

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

const (
	linuxClockTicksPerSecond = 100
	statFieldsAfterComm      = 15
	cutimeFieldAfterComm     = 13
	cstimeFieldAfterComm     = 14
)

var (
	errNoPeakRSS   = errors.New("no VmHWM line in status")
	errShortStat   = errors.New("stat line has too few fields")
	errNoCommClose = errors.New("stat line has no closing parenthesis")
)

// parsePeakRSS reads the VmHWM line of /proc/<pid>/status, which is in kB
func parsePeakRSS(status string) (uint64, error) {
	for _, line := range strings.Split(status, "\n") {
		value, ok := strings.CutPrefix(line, "VmHWM:")
		if !ok {
			continue
		}
		kb, err := strconv.ParseUint(strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(value), "kB")), 10, 64)
		if err != nil {
			return 0, err
		}
		return kb << 10, nil
	}
	return 0, errNoPeakRSS
}

// parseChildrenCPUTime sums the cutime and cstime fields of /proc/<pid>/stat
func parseChildrenCPUTime(stat string) (time.Duration, error) {
	commEnd := strings.LastIndexByte(stat, ')')
	if commEnd < 0 {
		return 0, errNoCommClose
	}
	fields := strings.Fields(stat[commEnd+1:])
	if len(fields) < statFieldsAfterComm {
		return 0, errShortStat
	}
	cutime, err := strconv.ParseUint(fields[cutimeFieldAfterComm], 10, 64)
	if err != nil {
		return 0, err
	}
	cstime, err := strconv.ParseUint(fields[cstimeFieldAfterComm], 10, 64)
	if err != nil {
		return 0, err
	}
	return time.Duration(cutime+cstime) * time.Second / linuxClockTicksPerSecond, nil
}
