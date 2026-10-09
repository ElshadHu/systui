//go:build linux

package metrics

import "os"

func readPressure() (PressureLevel, error) {
	text, err := os.ReadFile("/proc/pressure/memory")
	if err == nil {
		if level, ok := pressureFromPSI(string(text)); ok {
			return level, nil
		}
	}
	return readAvailablePressure()
}
