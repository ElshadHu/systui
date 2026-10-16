//go:build linux

package metrics

import "os"

// ReadPressure asks Pressure Stall Information and falls back to available
// memory on kernels without it
func ReadPressure() (PressureLevel, error) {
	text, err := os.ReadFile("/proc/pressure/memory")
	if err == nil {
		if level, ok := pressureFromPSI(string(text)); ok {
			return level, nil
		}
	}
	return readAvailablePressure()
}
