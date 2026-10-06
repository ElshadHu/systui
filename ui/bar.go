package ui

import "strings"

const (
	ansiGreen = "\x1b[32m"
	ansiReset = "\x1b[0m"
)

// renderBar draws with the filled part in green
func renderBar(pct float64, width int) string {
	if width <= 0 {
		return ""
	}
	filled := int(pct/100*float64(width) + 0.5)
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	return "[" +
		ansiGreen + strings.Repeat("|", filled) + ansiReset +
		strings.Repeat(" ", width-filled) +
		"]"
}
