package xsd

import "testing"

func TestParseScaleFacetInt(t *testing.T) {
	cases := []struct {
		in   string
		want int
		err  bool
	}{
		{"0", 0, false},
		{"5", 5, false},
		{"-3", -3, false},
		{"+7", 7, false},
		{"  42  ", 42, false},
		// In-range extremes round-trip exactly.
		{"2147483648", scaleFacetCap, false},
		{"-2147483648", -scaleFacetCap, false},
		// Positive overflow (within int64) saturates to +cap.
		{"9999999999", scaleFacetCap, false},
		// Positive overflow beyond int64 saturates to +cap.
		{"99999999999999999999999999", scaleFacetCap, false},
		// Negative overflow beyond int64 must saturate to -cap, not +cap.
		{"-99999999999999999999999999", -scaleFacetCap, false},
		{"not-an-int", 0, true},
	}
	for _, c := range cases {
		got, err := ParseScaleFacetInt(c.in)
		if c.err {
			if err == nil {
				t.Errorf("ParseScaleFacetInt(%q): want error, got %d", c.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseScaleFacetInt(%q): unexpected error %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseScaleFacetInt(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}
