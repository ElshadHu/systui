package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/ElshadHu/systui/internal/insight"
	"github.com/ElshadHu/systui/internal/metrics"
	"github.com/ElshadHu/systui/internal/units"
	"github.com/ElshadHu/systui/ui/theme"
)

const (
	fieldGap    = "   "
	maxInsights = 2
)

// topBar renders the status line and the insight line
func (m Model) topBar() []string {
	fields := []string{m.pressureField(), m.swapField()}
	if m.mem.HasCompressed {
		fields = append(fields, m.compressedField())
	}
	return []string{fitFields(m.width, fields), m.insightLine()}
}

// fitFields drops fields from the right until the line fits.
func fitFields(width int, fields []string) string {
	for len(fields) > 1 && width > 0 && lipgloss.Width(strings.Join(fields, fieldGap)) > width {
		fields = fields[:len(fields)-1]
	}
	return strings.Join(fields, fieldGap)
}

func (m Model) pressureField() string {
	t := m.theme
	label, style, glyph := pressureLook(m.pressure, t)
	return t.Muted.Render("Pressure") + " " + style.Render(glyph+" "+label)
}

func pressureLook(level metrics.PressureLevel, t theme.Theme) (string, lipgloss.Style, string) {
	switch level {
	case metrics.PressureWarning:
		return "YELLOW", t.Warning, t.Glyphs.Warning
	case metrics.PressureCritical:
		return "RED", t.Critical, t.Glyphs.Critical
	}
	return "GREEN", t.OK, t.Glyphs.OK
}

func (m Model) swapField() string {
	t := m.theme
	value := t.Muted.Render(units.GB(0))
	if m.mem.SwapUsed > 0 {
		value = t.Warning.Render(t.Glyphs.Warning + " " + units.GB(m.mem.SwapUsed))
	}
	return t.Muted.Render("Swap") + " " + value
}

func (m Model) compressedField() string {
	t := m.theme
	return t.Muted.Render("Compressed") + " " + t.Primary.Render(units.GB(m.mem.Compressed))
}

func (m Model) insightLine() string {
	t := m.theme
	if len(m.insights) == 0 {
		return t.Muted.Render("All normal")
	}
	parts := make([]string, 0, len(m.insights))
	for _, in := range m.insights {
		style, glyph := severityLook(in.Severity, t)
		parts = append(parts, style.Render(glyph+" "+in.Text))
	}
	return strings.Join(parts, t.Muted.Render(" "+t.Glyphs.Dot+" "))
}

func severityLook(s insight.Severity, t theme.Theme) (lipgloss.Style, string) {
	if s == insight.Critical {
		return t.Critical, t.Glyphs.Critical
	}
	return t.Warning, t.Glyphs.Warning
}
