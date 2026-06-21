package xsdtype

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/kud360/goxsd5/xsd"
)

// pdKind distinguishes the four members of the precisionDecimal value space:
// finite numbers and the three special values (±INF, NaN).
type pdKind uint8

const (
	pdFinite pdKind = iota
	pdPosInf
	pdNegInf
	pdNaN
)

// PrecisionDecimal is the value space of xs:precisionDecimal (XSD 1.1's
// IEEE-754 decimal floating-point datatype). Unlike xs:decimal it is a
// *cohort* model: a finite value carries not only its numerical value but a
// scale (the spec's arithmeticPrecision), so the lexical forms "3", "3.0" and
// "3.00" map to distinct values even though they are numerically equal. It
// also has signed zero (+0 and −0 are distinct values though they compare
// equal) and the special values ±INF and NaN.
//
// A finite value is Neg · Coef × 10^(−Scale); Coef is non-negative and the
// authored coefficient/scale are kept verbatim — never renormalized, because
// renormalizing (as Decimal does) would collapse the cohort. For ±INF/NaN
// only Kind is meaningful. A signed zero is Kind=pdFinite, Coef=0, with Neg
// recording the sign and Scale its authored scale.
type PrecisionDecimal struct {
	Kind pdKind
	Coef *big.Int
	// Sc is the scale (the spec's arithmeticPrecision); the Scale() method
	// exposes it as the xsd.Scaled capability. It can be negative (e.g.
	// "3.0e2" has Sc −1).
	Sc  int
	Neg bool
}

// PrecisionDecimal exposes TotalDigits()/FractionDigits() below, so it
// satisfies xsd.DigitCounted, and Scale(), so it satisfies xsd.Scaled.

// ParsePrecisionDecimal parses an xs:precisionDecimal lexical form per the
// datatype's lexical mapping: the special values INF/+INF/-INF/NaN (exact
// case only — the spec excludes the xs:float-style aliases such as "Infinity"
// and lowercase forms), or a decimal numeral with an optional exponent. The
// scale (arithmeticPrecision) is the number of fractional digits minus the
// exponent, so "3.0e2" parses with scale −1.
func ParsePrecisionDecimal(s string) (*PrecisionDecimal, error) {
	switch s {
	case "INF", "+INF":
		return &PrecisionDecimal{Kind: pdPosInf}, nil
	case "-INF":
		return &PrecisionDecimal{Kind: pdNegInf}, nil
	case "NaN":
		return &PrecisionDecimal{Kind: pdNaN}, nil
	}

	t := s
	neg := false
	if len(t) > 0 && (t[0] == '+' || t[0] == '-') {
		neg = t[0] == '-'
		t = t[1:]
	}

	mantissa, exp := t, 0
	if i := strings.IndexAny(t, "eE"); i >= 0 {
		e, err := strconv.Atoi(t[i+1:])
		if err != nil {
			return nil, fmt.Errorf("invalid precisionDecimal %q", s)
		}
		mantissa, exp = t[:i], e
	}

	intPart, fracPart := mantissa, ""
	if i := strings.IndexByte(mantissa, '.'); i >= 0 {
		intPart, fracPart = mantissa[:i], mantissa[i+1:]
	}
	if intPart == "" && fracPart == "" {
		return nil, fmt.Errorf("invalid precisionDecimal %q", s)
	}
	for _, part := range [2]string{intPart, fracPart} {
		for i := 0; i < len(part); i++ {
			if part[i] < '0' || part[i] > '9' {
				return nil, fmt.Errorf("invalid precisionDecimal %q", s)
			}
		}
	}

	coef := new(big.Int)
	if _, ok := coef.SetString(intPart+fracPart, 10); !ok {
		// Only possible for the empty string, ruled out above.
		return nil, fmt.Errorf("invalid precisionDecimal %q", s)
	}
	return &PrecisionDecimal{
		Kind: pdFinite,
		Coef: coef,
		Sc:   len(fracPart) - exp,
		Neg:  neg,
	}, nil
}

// Scale reports the value's scale (arithmeticPrecision). The bool is false for
// the special values ±INF/NaN, which have no scale.
func (p *PrecisionDecimal) Scale() (int, bool) {
	if p.Kind != pdFinite {
		return 0, false
	}
	return p.Sc, true
}

// FractionDigits is the number of digits to the right of the decimal point in
// the value's scale; for a negative scale (e.g. "3.0e2") it is 0, matching the
// canonical scientific form. Special values report 0.
func (p *PrecisionDecimal) FractionDigits() int {
	if p.Kind != pdFinite || p.Sc < 0 {
		return 0
	}
	return p.Sc
}

// TotalDigits is the number of significant decimal digits in the coefficient
// (leading zeros not counted; the value 0 has 1 digit). Special values report
// 0.
func (p *PrecisionDecimal) TotalDigits() int {
	if p.Kind != pdFinite {
		return 0
	}
	s := strings.TrimLeft(p.Coef.String(), "0")
	if s == "" {
		return 1
	}
	return len(s)
}

// String renders the canonical lexical form (precisionDecimalCanonicalMap):
// special values map to their literals; a finite value with scale ≥ 0 uses
// decimal-point notation and one with scale < 0 uses scientific notation, in
// both cases preserving the coefficient's digits so the scale is observable.
func (p *PrecisionDecimal) String() string {
	switch p.Kind {
	case pdPosInf:
		return "INF"
	case pdNegInf:
		return "-INF"
	case pdNaN:
		return "NaN"
	}
	sign := ""
	if p.Neg {
		sign = "-"
	}
	if p.Sc < 0 {
		return sign + scientificPrecision(p.Coef, p.Sc)
	}
	return sign + decimalPtPrecision(p.Coef, p.Sc)
}

// decimalPtPrecision renders a non-negative coefficient with scale ≥ 0 in
// decimal-point notation, keeping exactly scale fractional digits (the spec's
// decimalPtPrecision). Coefficient "300" with scale 2 → "3.00".
func decimalPtPrecision(coef *big.Int, scale int) string {
	s := coef.String()
	if scale == 0 {
		return s
	}
	for len(s) <= scale {
		s = "0" + s
	}
	return s[:len(s)-scale] + "." + s[len(s)-scale:]
}

// scientificPrecision renders a non-negative coefficient with scale < 0 in
// scientific notation with one digit before the point (the spec's
// scientificPrecision). The mantissa preserves every coefficient digit and
// the exponent is len(digits)−1−scale. Coefficient "30" with scale −1 →
// "3.0E2".
func scientificPrecision(coef *big.Int, scale int) string {
	digits := coef.String()
	mantissa := digits[:1]
	if len(digits) > 1 {
		mantissa = digits[:1] + "." + digits[1:]
	}
	exp := len(digits) - 1 - scale
	return mantissa + "E" + strconv.Itoa(exp)
}

// ComparePrecisionDecimal orders two precisionDecimal values per the value
// space's order relation: NaN is incomparable with everything (including
// itself), so it yields (0, false); the infinities order −INF < finite < INF;
// signed zeros and cohort members compare by numerical value, so +0 == −0 and
// 3 == 3.00.
func ComparePrecisionDecimal(a, b *PrecisionDecimal) (xsd.Order, bool) {
	if a.Kind == pdNaN || b.Kind == pdNaN {
		return 0, false
	}
	ar, br := pdRank(a), pdRank(b)
	if ar != br {
		if ar < br {
			return xsd.OrderLess, true
		}
		return xsd.OrderGreater, true
	}
	if a.Kind != pdFinite {
		// Same non-finite rank: ±INF equals its like.
		return xsd.OrderEqual, true
	}
	av, bv := pdSignedCoef(a), pdSignedCoef(b)
	// Align to the finer (larger) scale before comparing, so cohort members
	// like 3 and 3.00 compare equal. value = coef × 10^(−scale), so the value
	// with the coarser scale is scaled up by the difference in scales.
	if a.Sc < b.Sc {
		av = new(big.Int).Mul(av, pow10(b.Sc-a.Sc))
	}
	if b.Sc < a.Sc {
		bv = new(big.Int).Mul(bv, pow10(a.Sc-b.Sc))
	}
	return xsd.Order(av.Cmp(bv)), true
}

// Identical implements xsd.Identical: the equality relation the enumeration and
// pattern facets match by, which differs from the order relation only at NaN.
// NaN is incomparable in the order relation (including with itself) yet is a
// single value identical to itself, so enumeration of NaN to NaN succeeds; for
// every other pair identity is order-equality, so the cohort members 3, 3.0,
// 3.00 — and ±0 — are identical. The argument is xsd.Value; a non-precisionDecimal
// is never identical to one.
func (p *PrecisionDecimal) Identical(other xsd.Value) bool {
	q, ok := other.(*PrecisionDecimal)
	if !ok {
		return false
	}
	if p.Kind == pdNaN || q.Kind == pdNaN {
		return p.Kind == pdNaN && q.Kind == pdNaN
	}
	o, c := ComparePrecisionDecimal(p, q)
	return c && o == xsd.OrderEqual
}

// pdRank orders the value-space strata for comparison: −INF below all finite
// values, INF above them. Finite values share a rank and are then compared by
// magnitude.
func pdRank(p *PrecisionDecimal) int {
	switch p.Kind {
	case pdNegInf:
		return -1
	case pdPosInf:
		return 1
	default:
		return 0
	}
}

// pdSignedCoef returns the (signed) coefficient as a fresh big.Int; the value
// is coef × 10^(−scale), aligned to a common scale by the caller.
func pdSignedCoef(p *PrecisionDecimal) *big.Int {
	c := new(big.Int).Set(p.Coef)
	if p.Neg {
		c.Neg(c)
	}
	return c
}
