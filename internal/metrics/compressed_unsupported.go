//go:build !darwin

package metrics

// readCompressed reports no value where the kernel has no memory compressor
func readCompressed() (uint64, bool) {
	return 0, false
}
