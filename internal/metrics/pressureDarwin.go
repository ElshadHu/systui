//go:build darwin

package metrics

import (
	"fmt"

	"golang.org/x/sys/unix"
)

const (
	darwinPressureNormal   = 1
	darwinPressureWarning  = 2
	darwinPressureCritical = 4
)

func readPressure() (PressureLevel, error) {
	level, err := unix.SysctlUint32("kern.memorystatus_vm_pressure_level")
	if err != nil {
		return PressureNormal, err
	}
	switch level {
	case darwinPressureNormal:
		return PressureNormal, nil
	case darwinPressureWarning:
		return PressureWarning, nil
	case darwinPressureCritical:
		return PressureCritical, nil
	}
	return PressureNormal, fmt.Errorf("unknown memory pressure level %d", level)
}
