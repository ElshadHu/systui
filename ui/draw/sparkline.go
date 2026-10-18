package draw

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/ElshadHu/systui/internal/apps"
	"github.com/ElshadHu/systui/ui/theme"
)

const (
	SparkWidth          = apps.HistoryLength
	sparkSuffix         = len(" +1023MB")
	SparkGap            = 3
	SparkColumn         = SparkGap + SparkWidth + sparkSuffix
	CPUWarnPercent      = 25
	CPUHighPercent      = 60
	memGrowthBytes      = 100 << 20
	memFullScalePercent = 10
	sparkTopLevel       = 7
)

// Sparkline draws the history of one group for the metric the sort key
// belongs to, padded to SparkColumn
func Sparkline(h apps.History, key apps.SortKey, t theme.Theme, text lipgloss.Style, paint Painter) string {
	var drawn string
	if key.Family() == apps.ByMemory {
		drawn = memorySparkline(h.Memory, t, text, paint)
	} else {
		drawn = cpuSparkline(h.CPU, t, paint)
	}
	lead := paint(text).Render(strings.Repeat(" ", SparkGap))
	return lead + drawn + paint(text).Render(strings.Repeat(" ", max(SparkColumn-SparkGap-lipgloss.Width(drawn), 0)))
}

// cpuSparkline uses a fixed scale that tops out at CPUHighPercent
func cpuSparkline(values []float64, t theme.Theme, paint Painter) string {
	var b strings.Builder
	b.WriteString(sparkMissing(len(values), t, paint))
	for _, v := range values {
		level := int(math.Round(v / CPUHighPercent * sparkTopLevel))
		b.WriteString(paint(sparkCPUStyle(v, t)).Render(sparkBar(level, t)))
	}
	return b.String()
}

// memorySparkline draws each reading as its rise above the lowest reading in the window
func memorySparkline(values []uint64, t theme.Theme, text lipgloss.Style, paint Painter) string {
	style, suffix := text, ""
	if len(values) > 1 && values[len(values)-1] >= values[0]+memGrowthBytes {
		style = t.Warning
		suffix = " +" + growthText(values[len(values)-1]-values[0])
	}
	var b strings.Builder
	b.WriteString(sparkMissing(len(values), t, paint))
	if len(values) > 0 {
		lo := slices.Min(values)
		fullScale := float64(values[len(values)-1]) * memFullScalePercent / 100
		for _, v := range values {
			level := 0
			if fullScale > 0 {
				level = int(math.Round(float64(v-lo) / fullScale * sparkTopLevel))
			}
			b.WriteString(paint(style).Render(sparkBar(level, t)))
		}
	}
	b.WriteString(paint(style).Render(suffix))
	return b.String()
}

func sparkMissing(have int, t theme.Theme, paint Painter) string {
	missing := max(SparkWidth-have, 0)
	return paint(t.Muted).Render(strings.Repeat(t.Glyphs.SparkMissing, missing))
}

func sparkBar(level int, t theme.Theme) string {
	return t.Glyphs.SparkLevels[min(max(level, 0), sparkTopLevel)]
}

func sparkCPUStyle(pct float64, t theme.Theme) lipgloss.Style {
	switch {
	case pct > CPUHighPercent:
		return t.Critical
	case pct >= CPUWarnPercent:
		return t.Warning
	}
	return t.OK
}

// growthText shows a byte change the way the spec does: 140MB, 1.2GB
func growthText(n uint64) string {
	if n >= 1<<30 {
		return fmt.Sprintf("%.1fGB", float64(n)/(1<<30))
	}
	return fmt.Sprintf("%.0fMB", float64(n)/(1<<20))
}
