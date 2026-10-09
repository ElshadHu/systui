package metrics

import (
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v4/mem"
)

type PressureLevel int

const (
	PressureNormal PressureLevel = iota
	PressureWarning
	PressureCritical
)

const (
	lowAvailablePercent      = 20
	criticalAvailablePercent = 10
	someStallWarningPercent  = 10
	fullStallCriticalPercent = 10
)

// ReadPressure returns the memory pressure level. macOS and Linux ask the kernel,
// other systems derive it from available memory
func ReadPressure() (PressureLevel, error) {
	return readPressure()
}

// pressureFromAvailable is the fallback for kernels without a pressure signal
func pressureFromAvailable(available, total uint64) PressureLevel {
	if total == 0 {
		return PressureNormal
	}
	pct := float64(available) / float64(total) * 100
	switch {
	case pct < criticalAvailablePercent:
		return PressureCritical
	case pct < lowAvailablePercent:
		return PressureWarning
	}
	return PressureNormal
}

func readAvailablePressure() (PressureLevel, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return PressureNormal, err
	}
	return pressureFromAvailable(vm.Available, vm.Total), nil
}

// pressureFromPSI reads the 10 second stall averages
func pressureFromPSI(text string) (PressureLevel, bool) {
	some, someOK := stallAverage(text, "some")
	full, fullOK := stallAverage(text, "full")
	if !someOK || !fullOK {
		return PressureNormal, false
	}
	switch {
	case full >= fullStallCriticalPercent:
		return PressureCritical, true
	case some >= someStallWarningPercent:
		return PressureWarning, true
	}
	return PressureNormal, true
}

func stallAverage(text, kind string) (float64, bool) {
	for _, line := range strings.Split(text, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != kind {
			continue
		}
		value, ok := strings.CutPrefix(fields[1], "avg10=")
		if !ok {
			return 0, false
		}
		avg, err := strconv.ParseFloat(value, 64)
		return avg, err == nil
	}
	return 0, false
}
