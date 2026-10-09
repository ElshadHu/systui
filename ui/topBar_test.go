package ui

import "testing"

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
		if got := fitFields(c.width, fields); got != c.want {
			t.Errorf("width %d: want %q, got %q", c.width, c.want, got)
		}
	}
}
