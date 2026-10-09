//go:build darwin

package metrics

import "golang.org/x/sys/unix"

func readCompressed() (uint64, bool) {
	n, err := unix.SysctlUint64("vm.compressor_bytes_used")
	return n, err == nil
}
