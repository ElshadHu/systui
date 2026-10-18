package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/ElshadHu/systui/internal/apps"
	"github.com/ElshadHu/systui/internal/units"
	"github.com/ElshadHu/systui/ui/applist"
	"github.com/ElshadHu/systui/ui/draw"
	"github.com/shirou/gopsutil/v4/process"
)

const (
	memberIndent = 5
	memWarnBytes = 1 << 30
	memHighBytes = 2 << 30
	noValue      = "-"
)

// appsView renders the column header and the visible rows for height lines
func (m Model) appsView(height int) []string {
	if m.procErr != nil {
		return []string{m.theme.Critical.Render(fmt.Sprintf("%s error: %v", m.theme.Glyphs.Critical, m.procErr))}
	}
	l := m.columns()
	lines := []string{m.appsHeader(l)}
	shown, above, below := m.apps.Viewport(height - 1)
	if above > 0 {
		lines = append(lines, m.theme.Muted.Render(fmt.Sprintf(" %s %d more above", m.theme.Glyphs.Above, above)))
	}
	for _, r := range shown {
		lines = append(lines, m.appsRow(r, l, r.ID == m.apps.Selected))
	}
	if below > 0 {
		lines = append(lines, m.theme.Muted.Render(fmt.Sprintf(" %s %d more below", m.theme.Glyphs.Below, below)))
	}
	return lines
}

func (m Model) appsHeader(l layout) string {
	t := m.theme
	paint := draw.Surface(t, true)
	muted := paint(t.Muted)
	sorted, _ := m.sortedField(l)
	cells := []draw.Cell{{Style: muted, Text: " APP", Width: l.name}}
	for _, f := range l.fields {
		style, title := muted, f.title()
		switch {
		case f == sparkField:
			title = strings.Repeat(" ", draw.SparkGap) + m.trendTitle() + " TREND " + t.Glyphs.Dot + " 1 bar = 1s"
		case f == sorted:
			style, title = paint(t.Accent), title+" "+t.Glyphs.Expanded
		}
		cells = append(cells, draw.Cell{Style: style, Text: title, Width: f.width(), Right: f != sparkField})
	}
	line := muted.Render(strings.Repeat(" ", markerColumn)) + draw.JoinCells(cells, columnGap, muted)
	return draw.FillLine(line, m.width, muted)
}

// sortedField is the column the sort key marks, falling back to the column
// of its family when the layout has no column for the key itself
func (m Model) sortedField(l layout) (field, bool) {
	for _, key := range []apps.SortKey{m.apps.Sort(), m.apps.Sort().Family()} {
		for _, f := range l.fields {
			if k, ok := f.sortKey(); ok && k == key {
				return f, true
			}
		}
	}
	return 0, false
}

func (m Model) trendTitle() string {
	if m.apps.Sort().Family() == apps.ByMemory {
		return memField.title()
	}
	return cpuField.title()
}

func (m Model) appsRow(r applist.Row, l layout, selected bool) string {
	t := m.theme
	paint := draw.Surface(t, selected)
	text := paint(t.Primary)
	if selected {
		text = paint(t.Bright)
	}
	marker := paint(t.Muted).Render(" ")
	if selected {
		marker = paint(t.Accent).Render(t.Glyphs.Accent)
	}

	var cells []draw.Cell
	switch r.Kind {
	case applist.GroupRow:
		cells = m.groupCells(r.Group, l, text, paint)
	case applist.MemberRow:
		cells = m.memberCells(r.Member, l, text, paint)
	case applist.MoreRow:
		cells = m.moreCells(r.Hidden, l, paint)
	}
	return draw.FillLine(marker+draw.JoinCells(cells, columnGap, paint(t.Muted)), m.width, paint(t.Muted))
}

func (m Model) groupCells(g *apps.Group, l layout, text lipgloss.Style, paint draw.Painter) []draw.Cell {
	t := m.theme
	glyph := t.Glyphs.Collapsed
	if m.apps.Open(g.ID) {
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
	cells := []draw.Cell{{Style: nameStyle, Text: name, Width: l.name}}
	for _, f := range l.fields {
		cells = append(cells, m.groupCell(f, g, text, paint))
	}
	return cells
}

func (m Model) groupCell(f field, g *apps.Group, text lipgloss.Style, paint draw.Painter) draw.Cell {
	peak, cpuTime, started := g.PeakMemory, g.CPUTime, g.Started
	if g.Kind == apps.Small {
		peak, cpuTime, started = 0, 0, time.Time{}
	}
	switch f {
	case memField:
		return draw.Cell{Style: paint(m.liveStyle(m.groupMemoryStyle(g.Memory, text))), Text: memoryText(g.Memory), Width: memColumn, Right: true}
	case cpuField:
		return draw.Cell{Style: paint(m.liveStyle(m.cpuStyle(g.CPU, text))), Text: cpuText(g.CPU), Width: cpuColumn, Right: true}
	case procsField:
		return draw.Cell{Style: paint(m.theme.Muted), Text: strconv.Itoa(len(g.Members)), Width: procsColumn, Right: true}
	case sparkField:
		trend := ""
		if g.Kind != apps.Small {
			trend = m.sparkline(g.ID, text, paint)
		}
		return draw.Cell{Style: text, Text: trend, Width: draw.SparkColumn}
	}
	return m.lifetimeCell(f, peak, cpuTime, started, text, paint)
}

// sparkline draws the history of one group for the sorted metric
func (m Model) sparkline(id string, text lipgloss.Style, paint draw.Painter) string {
	return draw.Sparkline(m.history.Of(id), m.apps.Sort(), m.theme, text, paint)
}

func (m Model) memberCells(member apps.Member, l layout, text lipgloss.Style, paint draw.Painter) []draw.Cell {
	indent := strings.Repeat(" ", memberIndent)
	cells := []draw.Cell{{Style: text, Text: indent + member.Label + m.stateGlyph([]apps.Member{member}), Width: l.name}}
	for _, f := range l.fields {
		cells = append(cells, m.memberCell(f, member, text, paint))
	}
	return cells
}

func (m Model) memberCell(f field, member apps.Member, text lipgloss.Style, paint draw.Painter) draw.Cell {
	switch f {
	case memField:
		return draw.Cell{Style: paint(m.liveStyle(m.memoryStyle(member.Memory, text))), Text: memoryText(member.Memory), Width: memColumn, Right: true}
	case cpuField:
		return draw.Cell{Style: paint(m.liveStyle(m.cpuStyle(member.CPU, text))), Text: cpuText(member.CPU), Width: cpuColumn, Right: true}
	case procsField, sparkField:
		return draw.Cell{Style: text, Width: f.width(), Right: f != sparkField}
	}
	return m.lifetimeCell(f, member.PeakMemory, member.CPUTime, time.UnixMilli(member.Started), text, paint)
}

func (m Model) lifetimeCell(f field, peak uint64, cpuTime time.Duration, started time.Time, text lipgloss.Style, paint draw.Painter) draw.Cell {
	switch f {
	case peakField:
		return draw.Cell{Style: paint(m.memoryStyle(peak, text)), Text: memoryText(peak), Width: peakColumn, Right: true}
	case cpuTimeField:
		if cpuTime == 0 {
			return draw.Cell{Style: paint(m.theme.Muted), Text: noValue, Width: cpuTimeColumn, Right: true}
		}
		return draw.Cell{Style: text, Text: units.Duration(cpuTime), Width: cpuTimeColumn, Right: true}
	}
	since := sinceText(started, m.now)
	if since == noValue {
		return draw.Cell{Style: paint(m.theme.Muted), Text: since, Width: sinceColumn, Right: true}
	}
	return draw.Cell{Style: text, Text: since, Width: sinceColumn, Right: true}
}

func (m Model) moreCells(hidden int, l layout, paint draw.Painter) []draw.Cell {
	t := m.theme
	muted := paint(t.Muted)
	indent := strings.Repeat(" ", memberIndent)
	cells := []draw.Cell{{Style: muted, Text: fmt.Sprintf("%s%s %d more", indent, t.Glyphs.More, hidden), Width: l.name}}
	for i, f := range l.fields {
		hint := ""
		if i == len(l.fields)-1 {
			hint = t.Glyphs.Enter + " expand"
			if f == sparkField {
				hint = strings.Repeat(" ", draw.SparkGap) + hint
			}
		}
		cells = append(cells, draw.Cell{Style: muted, Text: hint, Width: f.width(), Right: f != sparkField})
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
	case pct >= draw.CPUHighPercent:
		return m.theme.Critical
	case pct >= draw.CPUWarnPercent:
		return m.theme.Warning
	case pct < 0.05:
		return m.theme.Muted
	}
	return text
}
