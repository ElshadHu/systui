//go:build !darwin

package metrics

// readFootprint reports no value where the kernel has no footprint counter
func readFootprint(pid int32) (uint64, bool) {
	return 0, false
}
