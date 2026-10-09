//go:build !darwin && !linux

package metrics

func readPressure() (PressureLevel, error) {
	return readAvailablePressure()
}
