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
