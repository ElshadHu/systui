package ui

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/ElshadHu/systui/internal/apps"
	"github.com/ElshadHu/systui/internal/units"
	"github.com/shirou/gopsutil/v4/process"
)

const (
	memColumn      = 8
	cpuColumn      = 7
	procsColumn    = 5
	peakColumn     = 8
	cpuTimeColumn  = 8
	sinceColumn    = 10
	columnGap      = 3
	markerColumn   = 1
	minNameColumn  = 18
	maxNameColumn  = 40
	memberIndent   = 5
	cpuWarnPercent = 25
	cpuHighPercent = 60
	memWarnBytes   = 1 << 30
	memHighBytes   = 2 << 30
	noValue        = "-"
)

// field is one column to the right of the name
type field int

const (
	memField field = iota
	cpuField
	procsField
	sparkField
	peakField
	cpuTimeField
	sinceField
)

var (
	appsFields        = []field{memField, cpuField, procsField, sparkField}
	appsDroppable     = []field{sparkField, procsField}
	lifetimeFields    = []field{memField, peakField, cpuField, cpuTimeField, sinceField}
	lifetimeDroppable = []field{sinceField, cpuField, memField}
)

func (f field) width() int {
	switch f {
	case memField:
		return memColumn
	case cpuField:
		return cpuColumn
	case procsField:
		return procsColumn
	case sparkField:
		return sparkColumn
	case peakField:
		return peakColumn
	case cpuTimeField:
		return cpuTimeColumn
	}
	return sinceColumn
}

func (f field) title() string {
	switch f {
	case memField:
		return "MEM"
	case cpuField:
		return "CPU"
	case procsField:
		return "PROCS"
	case peakField:
		return "PEAK"
	case cpuTimeField:
		return "CPU TIME"
	case sinceField:
		return "SINCE"
	}
	return ""
}

func (f field) sortKey() (apps.SortKey, bool) {
	switch f {
	case memField:
		return apps.ByMemory, true
	case cpuField:
		return apps.ByCPU, true
	case peakField:
		return apps.ByPeak, true
	case cpuTimeField:
		return apps.ByCPUTime, true
	}
	return 0, false
}

// layout is the columns for one terminal width: the name takes what the
// fixed columns leave, and droppable columns go first when it is too narrow
type layout struct {
	name   int
	fields []field
}

func fitAppsColumns(width int) layout {
	return fitLayout(width, appsFields, appsDroppable)
}

func fitLifetimeColumns(width int) layout {
	return fitLayout(width, lifetimeFields, lifetimeDroppable)
}

func fitLayout(width int, fields, droppable []field) layout {
	l := layout{fields: slices.Clone(fields)}
	if width <= 0 {
		l.name = 32
		return l
	}
	for _, drop := range droppable {
		if width-markerColumn-l.fixed() >= minNameColumn {
			break
		}
		l.fields = slices.DeleteFunc(l.fields, func(f field) bool { return f == drop })
	}
	l.name = min(max(width-markerColumn-l.fixed(), minNameColumn), maxNameColumn)
	return l
}

func (l layout) fixed() int {
	n := 0
	for _, f := range l.fields {
		n += columnGap + f.width()
	}
	return n
}

func (l layout) has(f field) bool {
	return slices.Contains(l.fields, f)
}

// columns is the layout of the tab being shown
func (m Model) columns() layout {
	if m.view == LifetimeView {
		return fitLifetimeColumns(m.width)
	}
	return fitAppsColumns(m.width)
}

// appsView renders the column header and the visible rows for height lines
func (m Model) appsView(height int) []string {
	if m.procErr != nil {
		return []string{m.theme.Critical.Render(fmt.Sprintf("%s error: %v", m.theme.Glyphs.Critical, m.procErr))}
	}
	l := m.columns()
	lines := []string{m.appsHeader(l)}
	shown, above, below := m.apps.viewport(height - 1)
	if above > 0 {
		lines = append(lines, m.theme.Muted.Render(fmt.Sprintf(" %s %d more above", m.theme.Glyphs.Above, above)))
	}
	for _, r := range shown {
		lines = append(lines, m.appsRow(r, l, r.id == m.apps.selected))
	}
	if below > 0 {
		lines = append(lines, m.theme.Muted.Render(fmt.Sprintf(" %s %d more below", m.theme.Glyphs.Below, below)))
	}
	return lines
}

func (m Model) appsHeader(l layout) string {
	t := m.theme
	paint := m.painter(true)
	muted := paint(t.Muted)
	sorted, _ := m.sortedField(l)
	cells := []cell{{muted, " APP", l.name, false}}
	for _, f := range l.fields {
		style, title := muted, f.title()
		switch {
		case f == sparkField:
			title = strings.Repeat(" ", sparkGap) + m.trendTitle() + " TREND " + t.Glyphs.Dot + " 1 bar = 1s"
		case f == sorted:
			style, title = paint(t.Accent), title+" "+t.Glyphs.Expanded
		}
		cells = append(cells, cell{style, title, f.width(), f != sparkField})
	}
	return m.fillLine(muted.Render(strings.Repeat(" ", markerColumn))+joinCells(cells, muted), muted)
}

// sortedField is the column the sort key marks, falling back to the column
// of its family when the layout has no column for the key itself
func (m Model) sortedField(l layout) (field, bool) {
	for _, key := range []apps.SortKey{m.apps.sort, m.apps.sort.Family()} {
		for _, f := range l.fields {
			if k, ok := f.sortKey(); ok && k == key {
				return f, true
			}
		}
	}
	return 0, false
}

func (m Model) trendTitle() string {
	if m.apps.sort.Family() == apps.ByMemory {
		return memField.title()
	}
	return cpuField.title()
}

// headerSortAt returns the sort key for the header cell at column x
func headerSortAt(l layout, x int) (apps.SortKey, bool) {
	start := markerColumn + l.name + columnGap
	for _, f := range l.fields {
		if x >= start && x < start+f.width() {
			return f.sortKey()
		}
		start += f.width() + columnGap
	}
	return 0, false
}

// fillLine pads a line to the terminal width so a painted background reaches the edge
func (m Model) fillLine(line string, style lipgloss.Style) string {
	return line + style.Render(strings.Repeat(" ", max(m.width-lipgloss.Width(line), 0)))
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

// painter puts the surface background behind header and selected rows
type painter func(lipgloss.Style) lipgloss.Style

func (m Model) painter(surface bool) painter {
	bg := m.theme.Colors.Surface
	return func(s lipgloss.Style) lipgloss.Style {
		if !surface || bg == nil {
			return s
		}
		return s.Background(bg)
	}
}

func (m Model) appsRow(r row, l layout, selected bool) string {
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
		cells = m.groupCells(r.group, l, text, paint)
	case memberRow:
		cells = m.memberCells(r.member, l, text, paint)
	case moreRow:
		cells = m.moreCells(r.hidden, l, paint)
	}
	return m.fillLine(marker+joinCells(cells, paint(t.Muted)), paint(t.Muted))
}

func (m Model) groupCells(g *apps.Group, l layout, text lipgloss.Style, paint painter) []cell {
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
	cells := []cell{{nameStyle, name, l.name, false}}
	for _, f := range l.fields {
		cells = append(cells, m.groupCell(f, g, text, paint))
	}
	return cells
}

func (m Model) groupCell(f field, g *apps.Group, text lipgloss.Style, paint painter) cell {
	peak, cpuTime, started := g.PeakMemory, g.CPUTime, g.Started
	if g.Kind == apps.Small {
		peak, cpuTime, started = 0, 0, time.Time{}
	}
	switch f {
	case memField:
		return cell{paint(m.liveStyle(m.groupMemoryStyle(g.Memory, text))), memoryText(g.Memory), memColumn, true}
	case cpuField:
		return cell{paint(m.liveStyle(m.cpuStyle(g.CPU, text))), cpuText(g.CPU), cpuColumn, true}
	case procsField:
		return cell{paint(m.theme.Muted), strconv.Itoa(len(g.Members)), procsColumn, true}
	case sparkField:
		trend := ""
		if g.Kind != apps.Small {
			trend = m.sparkline(g.ID, text, paint)
		}
		return cell{text, trend, sparkColumn, false}
	}
	return m.lifetimeCell(f, peak, cpuTime, started, text, paint)
}

func (m Model) memberCells(member apps.Member, l layout, text lipgloss.Style, paint painter) []cell {
	indent := strings.Repeat(" ", memberIndent)
	cells := []cell{{text, indent + member.Label + m.stateGlyph([]apps.Member{member}), l.name, false}}
	for _, f := range l.fields {
		cells = append(cells, m.memberCell(f, member, text, paint))
	}
	return cells
}

func (m Model) memberCell(f field, member apps.Member, text lipgloss.Style, paint painter) cell {
	switch f {
	case memField:
		return cell{paint(m.liveStyle(m.memoryStyle(member.Memory, text))), memoryText(member.Memory), memColumn, true}
	case cpuField:
		return cell{paint(m.liveStyle(m.cpuStyle(member.CPU, text))), cpuText(member.CPU), cpuColumn, true}
	case procsField, sparkField:
		return cell{text, "", f.width(), f != sparkField}
	}
	return m.lifetimeCell(f, member.PeakMemory, member.CPUTime, time.UnixMilli(member.Started), text, paint)
}

func (m Model) lifetimeCell(f field, peak uint64, cpuTime time.Duration, started time.Time, text lipgloss.Style, paint painter) cell {
	switch f {
	case peakField:
		return cell{paint(m.memoryStyle(peak, text)), memoryText(peak), peakColumn, true}
	case cpuTimeField:
		if cpuTime == 0 {
			return cell{paint(m.theme.Muted), noValue, cpuTimeColumn, true}
		}
		return cell{text, units.Duration(cpuTime), cpuTimeColumn, true}
	}
	since := sinceText(started, m.now)
	if since == noValue {
		return cell{paint(m.theme.Muted), since, sinceColumn, true}
	}
	return cell{text, since, sinceColumn, true}
}

func (m Model) moreCells(hidden int, l layout, paint painter) []cell {
	t := m.theme
	muted := paint(t.Muted)
	indent := strings.Repeat(" ", memberIndent)
	cells := []cell{{muted, fmt.Sprintf("%s%s %d more", indent, t.Glyphs.More, hidden), l.name, false}}
	for i, f := range l.fields {
		hint := ""
		if i == len(l.fields)-1 {
			hint = t.Glyphs.Enter + " expand"
			if f == sparkField {
				hint = strings.Repeat(" ", sparkGap) + hint
			}
		}
		cells = append(cells, cell{muted, hint, f.width(), f != sparkField})
	}
	return cells
}

// liveStyle mutes the MEM and CPU columns in the Lifetime tab so the eye
// lands on what the tab is for
func (m Model) liveStyle(style lipgloss.Style) lipgloss.Style {
	if m.view == LifetimeView {
		return m.theme.Muted
	}
	return style
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

func memoryText(n uint64) string {
	if n == 0 {
		return noValue
	}
	return units.Compact(n)
}

func cpuText(pct float64) string {
	return fmt.Sprintf("%.1f%%", pct)
}

// sinceText shows a start as a clock time today, yesterday, or a date
func sinceText(started, now time.Time) string {
	if !started.After(time.Unix(0, 0)) {
		return noValue
	}
	if sameDay(started, now) {
		return started.Format("15:04")
	}
	if sameDay(started, now.AddDate(0, 0, -1)) {
		return "yest " + started.Format("15:04")
	}
	return started.Format("Jan 2")
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
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
