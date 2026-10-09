package insight

import (
	"testing"

	"github.com/ElshadHu/systui/internal/metrics"
)

func TestTopRanksCriticalFirstAndCaps(t *testing.T) {
	list := []Insight{
		{Warning, "first warning"},
		{Critical, "critical"},
		{Warning, "second warning"},
	}
	got := Top(list, 2)
	if len(got) != 2 || got[0].Text != "critical" || got[1].Text != "first warning" {
		t.Fatalf("got %+v", got)
	}
	if list[0].Text != "first warning" {
		t.Fatal("Top must not reorder its input")
	}
}

func TestSystemThresholds(t *testing.T) {
	if got := System(79, metrics.PressureNormal, 0); len(got) != 0 {
		t.Fatalf("quiet machine should have no insights, got %+v", got)
	}
	got := System(96, metrics.PressureWarning, 1<<30)
	if len(got) != 3 || got[0].Severity != Critical || got[1].Severity != Warning || got[2].Text != "Swap in use, 1.0 GB" {
		t.Fatalf("got %+v", got)
	}
}
