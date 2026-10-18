package draw

import (
	"strings"
	"testing"
)

func TestJoinColumnsAlignsRaggedPanels(t *testing.T) {
	got := JoinColumns("|", []string{"ab", "c"}, []string{"x"})
	assertLines(t, []string{"ab | x", "c  |  "}, got)
}

func TestFitColumnsWrapsWhenNarrow(t *testing.T) {
	a, b, c := []string{"aaaa"}, []string{"bbbb"}, []string{"cccc"}
	got := FitColumns("|", 12, a, b, c)
	assertLines(t, []string{"aaaa | bbbb", "", "cccc"}, got)
}

func TestFitFieldsDropsFromTheRight(t *testing.T) {
	fields := []string{"CPU 15%", "Pressure GREEN", "Swap 0.0 GB", "Compressed 1.2 GB"}
	cases := []struct {
		width int
		want  string
	}{
		{0, "CPU 15%   Pressure GREEN   Swap 0.0 GB   Compressed 1.2 GB"},
		{80, "CPU 15%   Pressure GREEN   Swap 0.0 GB   Compressed 1.2 GB"},
		{40, "CPU 15%   Pressure GREEN   Swap 0.0 GB"},
		{25, "CPU 15%   Pressure GREEN"},
		{5, "CPU 15%"},
	}
	for _, c := range cases {
		if got := FitFields(c.width, "   ", fields); got != c.want {
			t.Errorf("width %d: want %q, got %q", c.width, c.want, got)
		}
	}
}

func assertLines(t *testing.T, want, got []string) {
	t.Helper()
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("want:\n%s\ngot:\n%s", strings.Join(want, "\n"), strings.Join(got, "\n"))
	}
}
