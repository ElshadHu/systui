//go:build windows

package metrics

// ReadPressure derives the level from available memory, since Windows has no pressure signal
func ReadPressure() (PressureLevel, error) {
	return readAvailablePressure()
}
