package insight

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/ElshadHu/systui/internal/metrics"
	"github.com/ElshadHu/systui/internal/units"
)

type Severity int

const (
	Warning Severity = iota
	Critical
)

// Insight is one line worth telling the user about, ranked by Severity
type Insight struct {
	Severity Severity
	Text     string
}

const (
	cpuWarningPercent  = 80
	cpuCriticalPercent = 95
)

// System reports what is wrong at the machine level
func System(cpuPercent float64, pressure metrics.PressureLevel, swapUsed uint64) []Insight {
	var list []Insight
	switch {
	case cpuPercent >= cpuCriticalPercent:
		list = append(list, Insight{Critical, fmt.Sprintf("CPU %.0f%% busy", cpuPercent)})
	case cpuPercent >= cpuWarningPercent:
		list = append(list, Insight{Warning, fmt.Sprintf("CPU %.0f%% busy", cpuPercent)})
	}
	switch pressure {
	case metrics.PressureWarning:
		list = append(list, Insight{Warning, "Memory under pressure"})
	case metrics.PressureCritical:
		list = append(list, Insight{Critical, "Memory pressure critical"})
	}
	if swapUsed > 0 {
		list = append(list, Insight{Warning, "Swap in use, " + units.GB(swapUsed)})
	}
	return list
}

// Top returns the n most severe insights, keeping the given order among equals
func Top(list []Insight, n int) []Insight {
	ranked := slices.Clone(list)
	slices.SortStableFunc(ranked, func(a, b Insight) int {
		return cmp.Compare(b.Severity, a.Severity)
	})
	if len(ranked) > n {
		ranked = ranked[:n]
	}
	return ranked
}
