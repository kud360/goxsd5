package xsdtype

import (
	"testing"

	"github.com/kud360/goxsd5/xsd"
)

func TestPrecisionDecimalParse(t *testing.T) {
	cases := []struct {
		in          string
		canon       string
		total, frac int
		scale       int
		ok          bool
	}{
		// §3.2 worked examples: the cohort 3 / 3.0 / 3.00 stay distinct.
		{"3", "3", 1, 0, 0, true},
		{"3.0", "3.0", 2, 1, 1, true},
		{"3.00", "3.00", 3, 2, 2, true},
		// Scientific notation drives a negative scale (3.0e2 → scale −1).
		{"3.0e2", "3.0E2", 2, 0, -1, true},
		{"3.0E2", "3.0E2", 2, 0, -1, true},
		{"300", "300", 3, 0, 0, true},
		{"30e1", "3.0E2", 2, 0, -1, true},
		// Sign and fractional forms.
		{"-1.5", "-1.5", 2, 1, 1, true},
		{"+001.100", "1.100", 4, 3, 3, true},
		{"0.005", "0.005", 1, 3, 3, true},
		{"1.5e-2", "0.015", 2, 3, 3, true},
		// Plain zero.
		{"0", "0", 1, 0, 0, true},
		// Malformed.
		{"", "", 0, 0, 0, false},
		{".", "", 0, 0, 0, false},
		{"1 5", "", 0, 0, 0, false},
		{"--1", "", 0, 0, 0, false},
		{"1e", "", 0, 0, 0, false},
		{"0x1", "", 0, 0, 0, false},
	}
	for _, tc := range cases {
		p, err := ParsePrecisionDecimal(tc.in)
		if !tc.ok {
			if err == nil {
				t.Errorf("ParsePrecisionDecimal(%q) should fail", tc.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParsePrecisionDecimal(%q): %v", tc.in, err)
			continue
		}
		if got := p.String(); got != tc.canon {
			t.Errorf("ParsePrecisionDecimal(%q).String() = %q, want %q", tc.in, got, tc.canon)
		}
		if got := p.TotalDigits(); got != tc.total {
			t.Errorf("ParsePrecisionDecimal(%q).TotalDigits() = %d, want %d", tc.in, got, tc.total)
		}
		if got := p.FractionDigits(); got != tc.frac {
			t.Errorf("ParsePrecisionDecimal(%q).FractionDigits() = %d, want %d", tc.in, got, tc.frac)
		}
		sc, ok := p.Scale()
		if !ok {
			t.Errorf("ParsePrecisionDecimal(%q).Scale() absent, want %d", tc.in, tc.scale)
		}
		if sc != tc.scale {
			t.Errorf("ParsePrecisionDecimal(%q).Scale() = %d, want %d", tc.in, sc, tc.scale)
		}
	}
}

// Cohort members are distinct values: equal numerically but with different
// canonical forms and scales.
func TestPrecisionDecimalCohortDistinct(t *testing.T) {
	a := mustPD(t, "3")
	b := mustPD(t, "3.00")
	if a.String() == b.String() {
		t.Fatalf("3 and 3.00 should have distinct canonical forms, both %q", a.String())
	}
	if as, _ := a.Scale(); as != 0 {
		t.Errorf("3 scale = %d, want 0", as)
	}
	if bs, _ := b.Scale(); bs != 2 {
		t.Errorf("3.00 scale = %d, want 2", bs)
	}
	// They are numerically equal, so they compare equal.
	if o, ok := ComparePrecisionDecimal(a, b); !ok || o != xsd.OrderEqual {
		t.Errorf("compare(3, 3.00) = (%v, %v), want (equal, true)", o, ok)
	}
}

func TestPrecisionDecimalSpecialValues(t *testing.T) {
	cases := []struct {
		in    string
		canon string
		ok    bool
	}{
		{"INF", "INF", true},
		{"+INF", "INF", true},
		{"-INF", "-INF", true},
		{"NaN", "NaN", true},
		// Exact case only — float-style aliases are rejected.
		{"Infinity", "", false},
		{"-Infinity", "", false},
		{"inf", "", false},
		{"Inf", "", false},
		{"nan", "", false},
		{"NAN", "", false},
		{"+NaN", "", false},
	}
	for _, tc := range cases {
		p, err := ParsePrecisionDecimal(tc.in)
		if !tc.ok {
			if err == nil {
				t.Errorf("ParsePrecisionDecimal(%q) should fail", tc.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParsePrecisionDecimal(%q): %v", tc.in, err)
			continue
		}
		if got := p.String(); got != tc.canon {
			t.Errorf("ParsePrecisionDecimal(%q).String() = %q, want %q", tc.in, got, tc.canon)
		}
		// Special values have no scale and no digit counts.
		if _, ok := p.Scale(); ok {
			t.Errorf("%q should have no scale", tc.in)
		}
	}
}

// §3.1 order relation: ±0 equal, −INF < finite < INF, INF==INF, −INF==−INF,
// and NaN incomparable with everything including itself.
func TestPrecisionDecimalOrder(t *testing.T) {
	posZero := mustPD(t, "0")
	negZero := mustPD(t, "-0")
	one := mustPD(t, "1")
	negOne := mustPD(t, "-1")
	posInf := mustPD(t, "INF")
	negInf := mustPD(t, "-INF")
	nan := mustPD(t, "NaN")

	expectOrder(t, "+0 vs -0", posZero, negZero, xsd.OrderEqual, true)
	expectOrder(t, "INF vs INF", posInf, posInf, xsd.OrderEqual, true)
	expectOrder(t, "-INF vs -INF", negInf, negInf, xsd.OrderEqual, true)
	expectOrder(t, "-INF vs 1", negInf, one, xsd.OrderLess, true)
	expectOrder(t, "1 vs INF", one, posInf, xsd.OrderLess, true)
	expectOrder(t, "INF vs -INF", posInf, negInf, xsd.OrderGreater, true)
	expectOrder(t, "-1 vs 1", negOne, one, xsd.OrderLess, true)
	expectOrder(t, "1 vs -1", one, negOne, xsd.OrderGreater, true)

	// NaN is incomparable with itself and with every other value.
	for _, other := range []*PrecisionDecimal{nan, one, posInf, negInf, posZero} {
		if o, ok := ComparePrecisionDecimal(nan, other); ok {
			t.Errorf("compare(NaN, %s) = (%v, true), want incomparable", other, o)
		}
		if o, ok := ComparePrecisionDecimal(other, nan); ok {
			t.Errorf("compare(%s, NaN) = (%v, true), want incomparable", other, o)
		}
	}
}

// The shared CompareValues entrypoint dispatches precisionDecimal pairs and
// treats a precisionDecimal/decimal cross-pair as incomparable.
func TestPrecisionDecimalCompareValues(t *testing.T) {
	a := xsd.Value(mustPD(t, "3.00"))
	b := xsd.Value(mustPD(t, "3"))
	if o, ok := CompareValues(a, b); !ok || o != xsd.OrderEqual {
		t.Errorf("CompareValues(3.00, 3) = (%v, %v), want (equal, true)", o, ok)
	}
	dec, _ := ParseDecimal("3")
	if _, ok := CompareValues(a, xsd.Value(dec)); ok {
		t.Error("precisionDecimal vs decimal should be incomparable")
	}
}

// The value type satisfies the capability interfaces the facet engine reaches
// for: DigitCounted (totalDigits/fractionDigits) and the new Scaled.
func TestPrecisionDecimalCapabilities(t *testing.T) {
	var v xsd.Value = mustPD(t, "1.5")
	if _, ok := v.(xsd.DigitCounted); !ok {
		t.Error("*PrecisionDecimal should satisfy xsd.DigitCounted")
	}
	s, ok := v.(xsd.Scaled)
	if !ok {
		t.Fatal("*PrecisionDecimal should satisfy xsd.Scaled")
	}
	if sc, ok := s.Scale(); !ok || sc != 1 {
		t.Errorf("Scale() = (%d, %v), want (1, true)", sc, ok)
	}
}

func mustPD(t *testing.T, s string) *PrecisionDecimal {
	t.Helper()
	p, err := ParsePrecisionDecimal(s)
	if err != nil {
		t.Fatalf("ParsePrecisionDecimal(%q): %v", s, err)
	}
	return p
}

func expectOrder(t *testing.T, label string, a, b *PrecisionDecimal, want xsd.Order, wantOK bool) {
	t.Helper()
	o, ok := ComparePrecisionDecimal(a, b)
	if ok != wantOK || o != want {
		t.Errorf("%s: compare = (%v, %v), want (%v, %v)", label, o, ok, want, wantOK)
	}
}
