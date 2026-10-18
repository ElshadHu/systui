package ui

import (
	"fmt"

	"github.com/ElshadHu/systui/internal/units"
	"github.com/ElshadHu/systui/ui/draw"
)

const (
	defaultBarWidth = 20
	cpuColWidth     = 16
	memColWidth     = 18
)

func (m Model) overviewView() []string {
	return draw.FitColumns(m.theme.Glyphs.Separator, m.width, m.usagePanel(), m.cpuPanel(), m.memPanel())
}

// usagePanel renders the CPU and MEM bars
func (m Model) usagePanel() []string {
	return []string{
		m.theme.Bright.Render("% Usage"),
		m.usageRow("CPU", m.cpu.Usage),
		m.usageRow("MEM", m.mem.UsedPercent),
	}
}

func (m Model) usageRow(label string, pct float64) string {
	return m.theme.Bright.Render(label+":") + " " +
		draw.Bar(pct, m.barWidth, m.theme) + " " +
		m.theme.Bright.Render(fmt.Sprintf("%5.1f%%", pct))
}

// cpuPanel renders the per-counter CPU breakdown
func (m Model) cpuPanel() []string {
	cells := []string{
		m.pct("user", m.cpu.User), m.pct("sys", m.cpu.System), m.pct("idle", m.cpu.Idle),
		m.pct("nice", m.cpu.Nice), m.pct("iowait", m.cpu.Iowait), m.pct("irq", m.cpu.Irq),
		m.pct("softirq", m.cpu.Softirq), m.pct("steal", m.cpu.Steal), m.pct("guest", m.cpu.Guest),
	}
	return append([]string{m.theme.Bright.Render("CPU")}, draw.Grid(cells, 3, cpuColWidth)...)
}

func (m Model) memPanel() []string {
	cells := []string{
		m.kv("total", units.Bytes(m.mem.Total)),
		m.kv("used", units.Bytes(m.mem.Used)),
		m.kv("free", units.Bytes(m.mem.Free)),
		m.kv("active", units.Bytes(m.mem.Active)),
		m.kv("buffers", units.Bytes(m.mem.Buffers)),
		m.kv("cached", units.Bytes(m.mem.Cached)),
	}
	return append([]string{m.theme.Bright.Render("MEM")}, draw.Grid(cells, 3, memColWidth)...)
}

func (m Model) pct(label string, v float64) string {
	return m.theme.Bright.Render(fmt.Sprintf("%s: %.1f%%", label, v))
}

func (m Model) kv(label, v string) string {
	return m.theme.Bright.Render(label + ": " + v)
}
