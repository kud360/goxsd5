package xsdtype

import (
	"testing"

	"github.com/kud360/goxsd5/xsd"
)

func TestDecimal(t *testing.T) {
	cases := []struct {
		in          string
		canon       string
		total, frac int
		ok          bool
	}{
		{"1.5", "1.5", 2, 1, true},
		{"-.5", "-0.5", 1, 1, true},
		{"1.", "1", 1, 0, true},
		{"+001.100", "1.1", 2, 1, true},
		{"0", "0", 1, 0, true},
		{"100", "100", 3, 0, true},
		{"0.005", "0.005", 1, 3, true},
		{"123.45", "123.45", 5, 2, true},
		{"", "", 0, 0, false},
		{".", "", 0, 0, false},
		{"1e5", "", 0, 0, false},
		{"1 5", "", 0, 0, false},
		{"--1", "", 0, 0, false},
	}
	for _, tc := range cases {
		d, err := ParseDecimal(tc.in)
		if !tc.ok {
			if err == nil {
				t.Errorf("ParseDecimal(%q) should fail", tc.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseDecimal(%q): %v", tc.in, err)
			continue
		}
		if d.String() != tc.canon {
			t.Errorf("ParseDecimal(%q).String() = %q, want %q", tc.in, d.String(), tc.canon)
		}
		if d.TotalDigits() != tc.total {
			t.Errorf("ParseDecimal(%q).TotalDigits() = %d, want %d", tc.in, d.TotalDigits(), tc.total)
		}
		if d.FractionDigits() != tc.frac {
			t.Errorf("ParseDecimal(%q).FractionDigits() = %d, want %d", tc.in, d.FractionDigits(), tc.frac)
		}
	}

	a, _ := ParseDecimal("1.50")
	b, _ := ParseDecimal("1.5")
	if a.Cmp(b) != xsd.OrderEqual {
		t.Error("1.50 != 1.5")
	}
	c, _ := ParseDecimal("-2")
	if c.Cmp(b) != xsd.OrderLess {
		t.Error("-2 should be < 1.5")
	}
}
