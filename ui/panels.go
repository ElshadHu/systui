package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// padRight extends s with spaces until it occupies width columns
func padRight(s string, width int) string {
	gap := width - lipgloss.Width(s)
	if gap <= 0 {
		return s
	}
	return s + strings.Repeat(" ", gap)
}

// grid arranges cells into rows, filling each column top to bottom
func grid(cells []string, rows, colWidth int) []string {
	lines := make([]string, rows)
	for i, cell := range cells {
		lines[i%rows] += padRight(cell, colWidth)
	}
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " ")
	}
	return lines
}

// fitColumns places panels side by side while they fit in width columns
// and starts a new group of rows when the next panel would be cut off.
// A width of 0 means the terminal size is not known yet, so nothing is cut.
func fitColumns(sep string, width int, panels ...[]string) []string {
	var out []string
	var group [][]string
	used := 0
	for _, p := range panels {
		need := widest(p)
		if len(group) > 0 {
			need += lipgloss.Width(" " + sep + " ")
		}
		if width > 0 && len(group) > 0 && used+need > width {
			out = append(out, joinColumns(sep, group...)...)
			out = append(out, "")
			group = nil
			used = 0
			need = widest(p)
		}
		group = append(group, p)
		used += need
	}
	return append(out, joinColumns(sep, group...)...)
}

// joinColumns places panels next to each other with sep between them
func joinColumns(sep string, panels ...[]string) []string {
	height := tallest(panels)
	widths := make([]int, len(panels))
	for i, p := range panels {
		widths[i] = widest(p)
	}

	out := make([]string, height)
	parts := make([]string, len(panels))
	for row := 0; row < height; row++ {
		for i, p := range panels {
			parts[i] = padRight(lineAt(p, row), widths[i])
		}
		out[row] = strings.Join(parts, " "+sep+" ")
	}
	return out
}

func tallest(panels [][]string) int {
	n := 0
	for _, p := range panels {
		if len(p) > n {
			n = len(p)
		}
	}
	return n
}

func widest(lines []string) int {
	n := 0
	for _, line := range lines {
		if w := lipgloss.Width(line); w > n {
			n = w
		}
	}
	return n
}

func lineAt(lines []string, row int) string {
	if row < len(lines) {
		return lines[row]
	}
	return ""
}
