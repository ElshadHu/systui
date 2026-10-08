package ui

import "strings"

const (
	ansiBoldWhite = "\x1b[1;97m"
	ansiFill      = "\x1b[92m"
	ansiTrack     = "\x1b[90m"
	ansiReset     = "\x1b[0m"
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

	return boldWhite("[") +
		ansiFill + strings.Repeat("|", filled) + ansiReset +
		ansiTrack + strings.Repeat("|", width-filled) + ansiReset +
		boldWhite("]")
}

// boldWhite renders s in bold bright white
func boldWhite(s string) string {
	return ansiBoldWhite + s + ansiReset
}
