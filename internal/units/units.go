package units

import "fmt"

// Bytes shows a byte count with B, KB ...
func Bytes(n uint64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%dB", n)
	}
	names := []string{"KB", "MB", "GB", "TB"}
	v := float64(n) / unit
	i := 0
	for v >= unit && i < len(names)-1 {
		v /= unit
		i++
	}
	return fmt.Sprintf("%.2f%s", v, names[i])
}

// GB shows a byte count in gigabytes with one decimal
func GB(n uint64) string {
	return fmt.Sprintf("%.1f GB", float64(n)/(1<<30))
}
