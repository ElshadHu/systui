package ui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/ElshadHu/systui/internal/apps"
	"github.com/ElshadHu/systui/ui/draw"
	"github.com/charmbracelet/x/ansi"
)

func plainSparkline(m Model, id string) string {
	return ansi.Strip(m.sparkline(id, m.theme.Primary, draw.Surface(m.theme, false)))
}

func TestCPUSparklineFillsFromTheRightAndTopsOutAtHigh(t *testing.T) {
	m := testModel(t)
	m.apps.SetSort(apps.ByCPU)
	for _, cpu := range []float64{0, 30, 60, 250} {
		m.history.Record([]apps.Group{{ID: "a", CPU: cpu}}, m.now)
	}
	got := plainSparkline(m, "a")
	levels := m.theme.Glyphs.SparkLevels
	want := strings.Repeat(" ", draw.SparkGap) + strings.Repeat(m.theme.Glyphs.SparkMissing, 16) + levels[0] + levels[4] + levels[7] + levels[7]
	if !strings.HasPrefix(got, want) {
		t.Fatalf("want %q, got %q", want, got)
	}
	if lipgloss.Width(got) != draw.SparkColumn {
		t.Fatalf("sparkline should be padded to %d, got %d", draw.SparkColumn, lipgloss.Width(got))
	}
}

func TestMemorySparklineScalesToTheRowAndShowsGrowth(t *testing.T) {
	m := testModel(t)
	m.apps.SetSort(apps.ByMemory)
	m.history.Record([]apps.Group{{ID: "flat", Memory: 1 << 30}, {ID: "wobble", Memory: 100 << 20}, {ID: "grow", Memory: 1 << 30}}, m.now)
	m.history.Record([]apps.Group{{ID: "flat", Memory: 1<<30 + 2<<20}, {ID: "wobble", Memory: 102 << 20}, {ID: "grow", Memory: 1<<30 + 140<<20}}, m.now)
	levels := m.theme.Glyphs.SparkLevels

	flat := plainSparkline(m, "flat")
	if !strings.HasSuffix(strings.TrimRight(flat, " "), levels[0]+levels[0]) {
		t.Fatalf("2 MB on a 1 GB row is noise and should draw flat, got %q", flat)
	}
	wobble := plainSparkline(m, "wobble")
	if !strings.HasSuffix(strings.TrimRight(wobble, " "), levels[0]+levels[1]) {
		t.Fatalf("2 MB on a 100 MB row is a small rise, got %q", wobble)
	}
	grow := plainSparkline(m, "grow")
	if !strings.HasSuffix(strings.TrimRight(grow, " "), levels[0]+levels[7]+" +140MB") {
		t.Fatalf("140 MB growth should fill the scale and say so, got %q", grow)
	}
}

func TestSmallBucketHasNoTrend(t *testing.T) {
	m := testModel(t)
	small := &apps.Group{ID: apps.SmallGroupID, Name: "3 small processes", Kind: apps.Small, Memory: 1 << 20}
	m.history.Record([]apps.Group{*small}, m.now)
	cells := m.groupCells(small, fitAppsColumns(120), m.theme.Primary, draw.Surface(m.theme, false))
	if trend := ansi.Strip(cells[len(cells)-1].Text); strings.TrimSpace(trend) != "" {
		t.Fatalf("small bucket should have a blank trend, got %q", trend)
	}
}

func TestSparklineHeaderNamesTheSortedMetric(t *testing.T) {
	m := testModel(t)
	m.apps.SetSort(apps.ByMemory)
	if h := ansi.Strip(m.appsHeader(fitAppsColumns(120))); !strings.Contains(h, "MEM TREND · 1 bar = 1s") {
		t.Fatalf("want MEM · last 20s in header, got %q", h)
	}
	m.apps.SetSort(apps.ByCPU)
	if h := ansi.Strip(m.appsHeader(fitAppsColumns(120))); !strings.Contains(h, "CPU TREND · 1 bar = 1s") {
		t.Fatalf("want CPU · last 20s in header, got %q", h)
	}
}

func TestSparklineKeepsRowWidthWhenSelected(t *testing.T) {
	m := testModel(t)
	m.history.Record(sampleGroups(), m.now)
	cols := fitAppsColumns(120)
	r := m.apps.Rows[0]
	plain, selected := m.appsRow(r, cols, false), m.appsRow(r, cols, true)
	if lipgloss.Width(plain) != lipgloss.Width(selected) {
		t.Fatalf("selected row width %d differs from plain %d", lipgloss.Width(selected), lipgloss.Width(plain))
	}
}
