package ui

import (
	"strings"
	"testing"
)

func TestVisibleWidthIgnoresStyles(t *testing.T) {
	if got := visibleWidth(renderBar(50, 10)); got != 12 {
		t.Fatalf("want 12, got %d", got)
	}
}

func TestJoinColumnsAlignsRaggedPanels(t *testing.T) {
	got := joinColumns("|", []string{"ab", "c"}, []string{"x"})
	assertLines(t, []string{"ab | x", "c  |  "}, got)
}

func TestFitColumnsWrapsWhenNarrow(t *testing.T) {
	a, b, c := []string{"aaaa"}, []string{"bbbb"}, []string{"cccc"}
	got := fitColumns("|", 12, a, b, c)
	assertLines(t, []string{"aaaa | bbbb", "", "cccc"}, got)
}

func assertLines(t *testing.T, want, got []string) {
	t.Helper()
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("want:\n%s\ngot:\n%s", strings.Join(want, "\n"), strings.Join(got, "\n"))
	}
}
