package ui

import (
	"slices"

	"github.com/ElshadHu/systui/internal/apps"
	"github.com/ElshadHu/systui/ui/draw"
)

const (
	memColumn     = 8
	cpuColumn     = 7
	procsColumn   = 5
	peakColumn    = 8
	cpuTimeColumn = 8
	sinceColumn   = 10
	columnGap     = 3
	markerColumn  = 1
	minNameColumn = 18
	maxNameColumn = 40
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
		return draw.SparkColumn
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
