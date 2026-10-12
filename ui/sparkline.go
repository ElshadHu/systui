package ui

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/ElshadHu/systui/internal/apps"
)

const (
	sparkWidth          = apps.HistoryLength
	sparkSuffix         = len(" +1023MB")
	sparkGap            = 3
	sparkColumn         = sparkGap + sparkWidth + sparkSuffix
	memGrowthBytes      = 100 << 20
	memFullScalePercent = 10
	sparkTopLevel       = 7
)

// sparkline draws the history of one group for the metric
func (m Model) sparkline(id string, text lipgloss.Style, paint func(lipgloss.Style) lipgloss.Style) string {
	h := m.history.Of(id)
	var drawn string
	if m.apps.sort == apps.ByMemory {
		drawn = m.memorySparkline(h.Memory, text, paint)
	} else {
		drawn = m.cpuSparkline(h.CPU, paint)
	}
	lead := paint(text).Render(strings.Repeat(" ", sparkGap))
	return lead + drawn + paint(text).Render(strings.Repeat(" ", max(sparkColumn-sparkGap-lipgloss.Width(drawn), 0)))
}

// cpuSparkline uses a fixed scale that tops out at cpuHighPercent
func (m Model) cpuSparkline(values []float64, paint func(lipgloss.Style) lipgloss.Style) string {
	var b strings.Builder
	b.WriteString(m.sparkMissing(len(values), paint))
	for _, v := range values {
		level := int(math.Round(v / cpuHighPercent * sparkTopLevel))
		b.WriteString(paint(m.sparkCPUStyle(v)).Render(m.sparkBar(level)))
	}
	return b.String()
}

// memorySparkline draws each reading as its rise above the lowest reading in the window
func (m Model) memorySparkline(values []uint64, text lipgloss.Style, paint func(lipgloss.Style) lipgloss.Style) string {
	style, suffix := text, ""
	if len(values) > 1 && values[len(values)-1] >= values[0]+memGrowthBytes {
		style = m.theme.Warning
		suffix = " +" + growthText(values[len(values)-1]-values[0])
	}
	var b strings.Builder
	b.WriteString(m.sparkMissing(len(values), paint))
	if len(values) > 0 {
		lo := slices.Min(values)
		fullScale := float64(values[len(values)-1]) * memFullScalePercent / 100
		for _, v := range values {
			level := 0
			if fullScale > 0 {
				level = int(math.Round(float64(v-lo) / fullScale * sparkTopLevel))
			}
			b.WriteString(paint(style).Render(m.sparkBar(level)))
		}
	}
	b.WriteString(paint(style).Render(suffix))
	return b.String()
}

func (m Model) sparkMissing(have int, paint func(lipgloss.Style) lipgloss.Style) string {
	missing := max(sparkWidth-have, 0)
	return paint(m.theme.Muted).Render(strings.Repeat(m.theme.Glyphs.SparkMissing, missing))
}

func (m Model) sparkBar(level int) string {
	return m.theme.Glyphs.SparkLevels[min(max(level, 0), sparkTopLevel)]
}

func (m Model) sparkCPUStyle(pct float64) lipgloss.Style {
	switch {
	case pct > cpuHighPercent:
		return m.theme.Critical
	case pct >= cpuWarnPercent:
		return m.theme.Warning
	}
	return m.theme.OK
}

// growthText shows a byte change the way the spec does: 140MB, 1.2GB
func growthText(n uint64) string {
	if n >= 1<<30 {
		return fmt.Sprintf("%.1fGB", float64(n)/(1<<30))
	}
	return fmt.Sprintf("%.0fMB", float64(n)/(1<<20))
}
