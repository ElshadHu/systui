package metrics

import "testing"

func TestPressureFromPSI(t *testing.T) {
	cases := []struct {
		name string
		text string
		want PressureLevel
		ok   bool
	}{
		{"idle", "some avg10=0.00 avg60=0.00 avg300=0.00 total=0\nfull avg10=0.00 avg60=0.00 avg300=0.00 total=0\n", PressureNormal, true},
		{"some stalls", "some avg10=12.50 avg60=3.00 avg300=1.00 total=9\nfull avg10=1.00 avg60=0.00 avg300=0.00 total=1\n", PressureWarning, true},
		{"full stalls", "some avg10=40.00 avg60=3.00 avg300=1.00 total=9\nfull avg10=15.00 avg60=0.00 avg300=0.00 total=1\n", PressureCritical, true},
		{"missing full line", "some avg10=0.00 avg60=0.00 avg300=0.00 total=0\n", PressureNormal, false},
		{"garbage", "nope\n", PressureNormal, false},
	}
	for _, c := range cases {
		got, ok := pressureFromPSI(c.text)
		if got != c.want || ok != c.ok {
			t.Errorf("%s: want %v %v, got %v %v", c.name, c.want, c.ok, got, ok)
		}
	}
}

func TestPressureFromAvailable(t *testing.T) {
	cases := []struct {
		available uint64
		want      PressureLevel
	}{
		{50, PressureNormal},
		{20, PressureNormal},
		{19, PressureWarning},
		{10, PressureWarning},
		{9, PressureCritical},
	}
	for _, c := range cases {
		if got := pressureFromAvailable(c.available, 100); got != c.want {
			t.Errorf("available %d: want %v, got %v", c.available, c.want, got)
		}
	}
	if got := pressureFromAvailable(0, 0); got != PressureNormal {
		t.Errorf("zero total: want normal, got %v", got)
	}
}
