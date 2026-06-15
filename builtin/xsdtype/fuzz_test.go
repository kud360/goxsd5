package xsdtype

import (
	"testing"

	"github.com/kud360/goxsd5/xsd"
)

var decimalSeeds = []string{
	"0", "-0", "+0", "1", "-1", "+1", "3.14", "-3.14", ".5", "5.", "0.0",
	"00012.3400", "12345678901234567890", "-0.0000000001",
	"", " ", "1e3", "NaN", "1.2.3", "--1", "1-", "abc", "+-1", ".",
}

var durationSeeds = []string{
	"P1Y", "P1M", "P1D", "PT1H", "PT1M", "PT1S", "P1Y2M3DT4H5M6S",
	"-P1Y", "P0Y", "PT0.5S", "P1YT1S", "P10675199DT2H48M5.4775807S",
	"", "P", "PT", "1Y", "P1S", "PT1Y", "P-1Y", "PYM",
}

var dateTimeSeeds = []string{
	"2001-10-26T21:32:52", "2001-10-26T21:32:52+02:00", "2001-10-26T21:32:52Z",
	"2001-10-26", "21:32:52", "2001-10", "2001", "--10-26", "---26", "--10",
	"-0001-01-01T00:00:00", "9999-12-31T23:59:59.999999999",
	"24:00:00", "0000-01-01", "2001-13-01", "2001-02-30",
	"", "T", "Z", "2001-10-26T", "2001-10-26T25:00:00", "not-a-date",
}

func allKinds() []DateTimeKind {
	return []DateTimeKind{
		KindDateTime, KindDate, KindTime, KindGYearMonth,
		KindGYear, KindGMonthDay, KindGDay, KindGMonth,
	}
}

// FuzzParseDecimal asserts the lexical→value mapping never panics and that any
// value it produces is internally consistent: equal to itself, round-trips
// through its canonical String(), and reports non-negative digit counts.
func FuzzParseDecimal(f *testing.F) {
	for _, s := range decimalSeeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		d, err := ParseDecimal(s)
		if err != nil {
			return
		}
		if d == nil {
			t.Fatalf("ParseDecimal(%q) returned nil value, nil error", s)
		}
		if got := d.Cmp(d); got != xsd.OrderEqual {
			t.Fatalf("ParseDecimal(%q): d.Cmp(d) = %v, want OrderEqual", s, got)
		}
		if d.TotalDigits() < 0 || d.FractionDigits() < 0 {
			t.Fatalf("ParseDecimal(%q): negative digit count total=%d frac=%d", s, d.TotalDigits(), d.FractionDigits())
		}
		// Canonical form must re-parse to an equal value.
		canon := d.String()
		d2, err2 := ParseDecimal(canon)
		if err2 != nil {
			t.Fatalf("ParseDecimal(%q).String() = %q failed to re-parse: %v", s, canon, err2)
		}
		if got := d.Cmp(d2); got != xsd.OrderEqual {
			t.Fatalf("ParseDecimal(%q) round-trip via %q not equal: %v", s, canon, got)
		}
	})
}

// FuzzParseDuration asserts no panic and reflexive comparability of any value.
func FuzzParseDuration(f *testing.F) {
	for _, s := range durationSeeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		d, err := ParseDuration(s)
		if err != nil {
			return
		}
		if d == nil {
			t.Fatalf("ParseDuration(%q) returned nil value, nil error", s)
		}
		if o, ok := d.Compare(d); !ok || o != xsd.OrderEqual {
			t.Fatalf("ParseDuration(%q): d.Compare(d) = (%v,%v), want (OrderEqual,true)", s, o, ok)
		}
	})
}

// FuzzParseDateTime drives all seven date/time primitives. Each parsed value
// must compare equal to itself.
func FuzzParseDateTime(f *testing.F) {
	for _, s := range dateTimeSeeds {
		for _, k := range allKinds() {
			f.Add(int(k), s)
		}
	}
	f.Fuzz(func(t *testing.T, kindIdx int, s string) {
		kinds := allKinds()
		kind := kinds[((kindIdx%len(kinds))+len(kinds))%len(kinds)]
		dt, err := parseByKind(kind, s)
		if err != nil {
			return
		}
		if dt == nil {
			t.Fatalf("ParseDateTime(%v,%q) returned nil value, nil error", kind, s)
		}
		if o, ok := dt.Compare(dt); !ok || o != xsd.OrderEqual {
			t.Fatalf("ParseDateTime(%v,%q): self-compare = (%v,%v), want (OrderEqual,true)", kind, s, o, ok)
		}
	})
}

// FuzzCompareValues checks the order relation's algebra on two independently
// parsed decimals: comparison is total here, must be reflexively equal, and
// antisymmetric (swapping operands negates the order).
func FuzzCompareValues(f *testing.F) {
	for _, a := range decimalSeeds {
		for _, b := range decimalSeeds {
			f.Add(a, b)
		}
	}
	f.Fuzz(func(t *testing.T, sa, sb string) {
		a, ea := ParseDecimal(sa)
		b, eb := ParseDecimal(sb)
		if ea != nil || eb != nil {
			return
		}
		ab, ok1 := CompareValues(a, b)
		ba, ok2 := CompareValues(b, a)
		if !ok1 || !ok2 {
			t.Fatalf("CompareValues on two decimals (%q,%q) returned not-comparable", sa, sb)
		}
		if ab != -ba {
			t.Fatalf("CompareValues not antisymmetric for (%q,%q): a?b=%v b?a=%v", sa, sb, ab, ba)
		}
		if Equal(a, b) != (ab == xsd.OrderEqual) {
			t.Fatalf("Equal/Compare disagree for (%q,%q): Equal=%v order=%v", sa, sb, Equal(a, b), ab)
		}
	})
}
