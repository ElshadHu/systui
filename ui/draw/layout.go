package draw

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// PadRight extends s with spaces until it occupies width columns
func PadRight(s string, width int) string {
	gap := width - lipgloss.Width(s)
	if gap <= 0 {
		return s
	}
	return s + strings.Repeat(" ", gap)
}

func Truncate(s string, width int) string {
	if width <= 0 || lipgloss.Width(s) <= width {
		return s
	}
	runes := []rune(s)
	for len(runes) > 0 && lipgloss.Width(string(runes))+1 > width {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}

// AlignEnds puts left at the start and right at the end of a width wide line
func AlignEnds(left, right string, width int) string {
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if width <= 0 || gap < 1 {
		if right == "" {
			return left
		}
		return left + "  " + right
	}
	return left + strings.Repeat(" ", gap) + right
}

// FitFields drops fields from the right until the line fits
func FitFields(width int, gap string, fields []string) string {
	for len(fields) > 1 && width > 0 && lipgloss.Width(strings.Join(fields, gap)) > width {
		fields = fields[:len(fields)-1]
	}
	return strings.Join(fields, gap)
}

// Grid arranges cells into rows, filling each column top to bottom
func Grid(cells []string, rows, colWidth int) []string {
	lines := make([]string, rows)
	for i, cell := range cells {
		lines[i%rows] += PadRight(cell, colWidth)
	}
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " ")
	}
	return lines
}

// FitColumns places panels side by side while they fit in width columns
func FitColumns(sep string, width int, panels ...[]string) []string {
	var out []string
	var group [][]string
	used := 0
	for _, p := range panels {
		need := widest(p)
		if len(group) > 0 {
			need += lipgloss.Width(" " + sep + " ")
		}
		if width > 0 && len(group) > 0 && used+need > width {
			out = append(out, JoinColumns(sep, group...)...)
			out = append(out, "")
			group = nil
			used = 0
			need = widest(p)
		}
		group = append(group, p)
		used += need
	}
	return append(out, JoinColumns(sep, group...)...)
}

// JoinColumns places panels next to each other with sep between them
func JoinColumns(sep string, panels ...[]string) []string {
	height := tallest(panels)
	widths := make([]int, len(panels))
	for i, p := range panels {
		widths[i] = widest(p)
	}

	out := make([]string, height)
	parts := make([]string, len(panels))
	for row := 0; row < height; row++ {
		for i, p := range panels {
			parts[i] = PadRight(lineAt(p, row), widths[i])
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
