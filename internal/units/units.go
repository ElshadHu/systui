package units

import (
	"fmt"
	"time"
)

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

// Compact shows a byte count the way Activity Monitor does: 4.1 GB, 820 MB, 12 KB
func Compact(n uint64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(n)/(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.0f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.0f KB", float64(n)/(1<<10))
	}
	return fmt.Sprintf("%d B", n)
}

// Duration shows CPU time the way a person says it: 41s, 58m, 6h 12m, 2d 5h
func Duration(d time.Duration) string {
	const day = 24 * time.Hour
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < day:
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
	return fmt.Sprintf("%dd %dh", int(d/day), int(d.Hours())%24)
}
