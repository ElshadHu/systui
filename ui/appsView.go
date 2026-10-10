package ui

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/ElshadHu/systui/internal/apps"
	"github.com/ElshadHu/systui/internal/units"
	"github.com/shirou/gopsutil/v4/process"
)

const (
	memColumn      = 8
	cpuColumn      = 7
	procsColumn    = 5
	flagColumn     = 14
	columnGap      = 3
	minNameColumn  = 18
	memberIndent   = 5
	cpuWarnPercent = 10
	cpuHighPercent = 50
	memWarnBytes   = 1 << 30
	memHighBytes   = 2 << 30
)

// columns is the layout for one terminal width: the name takes what the
// fixed columns leave, and the rightmost columns go first when it is too narrow
type columns struct {
	name  int
	procs bool
	flag  bool
}

func fitAppsColumns(width int) columns {
	c := columns{procs: true, flag: true}
	if width <= 0 {
		c.name = 32
		return c
	}
	fixed := func() int {
		n := columnGap + memColumn + columnGap + cpuColumn
		if c.procs {
			n += columnGap + procsColumn
		}
		if c.flag {
			n += columnGap + flagColumn
		}
		return n
	}
	c.name = width - 1 - fixed()
	if c.name < minNameColumn {
		c.flag = false
		c.name = width - 1 - fixed()
	}
	if c.name < minNameColumn {
		c.procs = false
		c.name = width - 1 - fixed()
	}
	c.name = max(c.name, minNameColumn)
	return c
}

// appsView renders the column header and the visible rows for height lines
func (m Model) appsView(height int) []string {
	if m.procErr != nil {
		return []string{m.theme.Critical.Render(fmt.Sprintf("%s error: %v", m.theme.Glyphs.Critical, m.procErr))}
	}
	cols := fitAppsColumns(m.width)
	lines := []string{m.appsHeader(cols)}
	shown, above, below := m.apps.viewport(height - 1)
	if above > 0 {
		lines = append(lines, m.theme.Muted.Render(fmt.Sprintf(" %s %d more above", m.theme.Glyphs.Above, above)))
	}
	for _, r := range shown {
		lines = append(lines, m.appsRow(r, cols, r.id == m.apps.selected))
	}
	if below > 0 {
		lines = append(lines, m.theme.Muted.Render(fmt.Sprintf(" %s %d more below", m.theme.Glyphs.Below, below)))
	}
	return lines
}

func (m Model) appsHeader(cols columns) string {
	t := m.theme
	paint := m.painter(true)
	name := paint(t.Muted)
	mem, cpu := paint(t.Muted), paint(t.Muted)
	memTitle, cpuTitle := "MEM", "CPU"
	if m.apps.sort == apps.ByMemory {
		mem, memTitle = paint(t.Accent), "MEM "+t.Glyphs.Expanded
	} else {
		cpu, cpuTitle = paint(t.Accent), "CPU "+t.Glyphs.Expanded
	}
	cells := []cell{
		{name, " APP", cols.name, false},
		{mem, memTitle, memColumn, true},
		{cpu, cpuTitle, cpuColumn, true},
	}
	if cols.procs {
		cells = append(cells, cell{paint(t.Muted), "PROCS", procsColumn, true})
	}
	if cols.flag {
		cells = append(cells, cell{paint(t.Muted), "", flagColumn, false})
	}
	return joinCells(cells, paint(t.Muted))
}

// cell is one column of a row: its style, text, width and alignment
type cell struct {
	style lipgloss.Style
	text  string
	width int
	right bool
}

func joinCells(cells []cell, gapStyle lipgloss.Style) string {
	parts := make([]string, len(cells))
	for i, c := range cells {
		text := truncate(c.text, c.width)
		pad := strings.Repeat(" ", max(c.width-lipgloss.Width(text), 0))
		if c.right {
			parts[i] = c.style.Render(pad + text)
		} else {
			parts[i] = c.style.Render(text + pad)
		}
	}
	return strings.Join(parts, gapStyle.Render(strings.Repeat(" ", columnGap)))
}

func truncate(s string, width int) string {
	if width <= 0 || lipgloss.Width(s) <= width {
		return s
	}
	runes := []rune(s)
	for len(runes) > 0 && lipgloss.Width(string(runes))+1 > width {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}

// painter returns a style modifier that puts the surface background behind
// header and selected rows
func (m Model) painter(surface bool) func(lipgloss.Style) lipgloss.Style {
	bg := m.theme.Colors.Surface
	return func(s lipgloss.Style) lipgloss.Style {
		if !surface || bg == nil {
			return s
		}
		return s.Background(bg)
	}
}

func (m Model) appsRow(r row, cols columns, selected bool) string {
	t := m.theme
	paint := m.painter(selected)
	text := paint(t.Primary)
	if selected {
		text = paint(t.Bright)
	}
	marker := paint(t.Muted).Render(" ")
	if selected {
		marker = paint(t.Accent).Render(t.Glyphs.Accent)
	}

	var cells []cell
	switch r.kind {
	case groupRow:
		cells = m.groupCells(r.group, cols, text, paint)
	case memberRow:
		cells = m.memberCells(r.member, cols, text, paint)
	case moreRow:
		cells = m.moreCells(r.hidden, cols, paint)
	}
	return marker + joinCells(cells, paint(t.Muted))
}

func (m Model) groupCells(g *apps.Group, cols columns, text lipgloss.Style, paint func(lipgloss.Style) lipgloss.Style) []cell {
	t := m.theme
	glyph := t.Glyphs.Collapsed
	if m.apps.allExpanded || m.apps.expanded == g.ID || m.apps.filter != "" {
		glyph = t.Glyphs.Expanded
	}
	name := glyph + " " + g.Name
	nameStyle := text
	switch g.Kind {
	case apps.System:
		name = glyph + " " + t.Glyphs.System + " " + g.Name
		nameStyle = paint(t.System)
	case apps.Small:
		nameStyle = paint(t.Muted)
	}
	if g.Kind == apps.App || g.Kind == apps.Tool {
		name += m.stateGlyph(g.Members)
	}
	cells := []cell{
		{nameStyle, name, cols.name, false},
		{paint(m.groupMemoryStyle(g.Memory, text)), memoryText(g.Memory), memColumn, true},
		{paint(m.cpuStyle(g.CPU, text)), cpuText(g.CPU), cpuColumn, true},
	}
	if cols.procs {
		cells = append(cells, cell{paint(t.Muted), strconv.Itoa(len(g.Members)), procsColumn, true})
	}
	if cols.flag {
		style, flag := m.groupFlag(g)
		cells = append(cells, cell{paint(style), flag, flagColumn, false})
	}
	return cells
}

func (m Model) memberCells(member apps.Member, cols columns, text lipgloss.Style, paint func(lipgloss.Style) lipgloss.Style) []cell {
	indent := strings.Repeat(" ", memberIndent)
	cells := []cell{
		{text, indent + member.Label + m.stateGlyph([]apps.Member{member}), cols.name, false},
		{paint(m.memoryStyle(member.Memory, text)), memoryText(member.Memory), memColumn, true},
		{paint(m.cpuStyle(member.CPU, text)), cpuText(member.CPU), cpuColumn, true},
	}
	if cols.procs {
		cells = append(cells, cell{text, "", procsColumn, true})
	}
	if cols.flag {
		cells = append(cells, cell{text, "", flagColumn, false})
	}
	return cells
}

func (m Model) moreCells(hidden int, cols columns, paint func(lipgloss.Style) lipgloss.Style) []cell {
	t := m.theme
	muted := paint(t.Muted)
	indent := strings.Repeat(" ", memberIndent)
	cells := []cell{
		{muted, fmt.Sprintf("%s%s %d more", indent, t.Glyphs.More, hidden), cols.name, false},
		{muted, "", memColumn, true},
		{muted, "", cpuColumn, true},
	}
	if cols.procs {
		cells = append(cells, cell{muted, "", procsColumn, true})
	}
	if cols.flag {
		cells = append(cells, cell{muted, t.Glyphs.Enter + " expand", flagColumn, false})
	}
	return cells
}

// stateGlyph marks a stopped or zombie process, or one owned by someone else
func (m Model) stateGlyph(members []apps.Member) string {
	g := m.theme.Glyphs
	for _, member := range members {
		switch member.Status {
		case process.Stop:
			return " " + g.Stopped
		case process.Zombie:
			return " " + g.Zombie
		}
	}
	if len(members) == 1 && members[0].UID != m.uid {
		return " " + g.Foreign
	}
	return ""
}

func (m Model) groupFlag(g *apps.Group) (lipgloss.Style, string) {
	t := m.theme
	switch {
	case g.CPU >= cpuHighPercent:
		return t.Critical, t.Glyphs.Critical + " runaway CPU"
	case g.CPU >= cpuWarnPercent:
		return t.Warning, t.Glyphs.Warning + " high CPU"
	case g.Kind == apps.Small:
		return t.Muted, t.Glyphs.Enter + " expand"
	}
	return t.Muted, ""
}

func memoryText(n uint64) string {
	if n == 0 {
		return "-"
	}
	return units.Compact(n)
}

func cpuText(pct float64) string {
	return fmt.Sprintf("%.1f%%", pct)
}

func (m Model) memoryStyle(n uint64, text lipgloss.Style) lipgloss.Style {
	if n == 0 {
		return m.theme.Muted
	}
	return text
}

func (m Model) groupMemoryStyle(n uint64, text lipgloss.Style) lipgloss.Style {
	switch {
	case n == 0:
		return m.theme.Muted
	case n >= memHighBytes:
		return m.theme.Critical
	case n >= memWarnBytes:
		return m.theme.Warning
	}
	return text
}

func (m Model) cpuStyle(pct float64, text lipgloss.Style) lipgloss.Style {
	switch {
	case pct >= cpuHighPercent:
		return m.theme.Critical
	case pct >= cpuWarnPercent:
		return m.theme.Warning
	case pct < 0.05:
		return m.theme.Muted
	}
	return text
}
