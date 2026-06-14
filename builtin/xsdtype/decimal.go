package xsdtype

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/kud360/goxsd5/xsd"
)

// Decimal is the value space of xs:decimal: an arbitrary-precision decimal
// kept as unscaled×10^-scale, normalized so Unscaled has no trailing zeros
// (scale is minimal, never negative). This representation makes the
// totalDigits/fractionDigits facets direct to evaluate.
type Decimal struct {
	Unscaled *big.Int
	Scale    int
}

// Decimal already exposes TotalDigits()/FractionDigits() below, so it
// satisfies xsd.DigitCounted (the totalDigits/fractionDigits facets).

// ParseDecimal parses an xs:decimal lexical form.
func ParseDecimal(s string) (*Decimal, error) {
	t := s
	neg := false
	if len(t) > 0 && (t[0] == '+' || t[0] == '-') {
		neg = t[0] == '-'
		t = t[1:]
	}
	intPart, fracPart := t, ""
	if i := strings.IndexByte(t, '.'); i >= 0 {
		intPart, fracPart = t[:i], t[i+1:]
	}
	if intPart == "" && fracPart == "" {
		return nil, fmt.Errorf("invalid decimal %q", s)
	}
	for _, part := range [2]string{intPart, fracPart} {
		for i := 0; i < len(part); i++ {
			if part[i] < '0' || part[i] > '9' {
				return nil, fmt.Errorf("invalid decimal %q", s)
			}
		}
	}
	u := new(big.Int)
	if _, ok := u.SetString(intPart+fracPart, 10); !ok {
		// Only possible for the empty string, handled above.
		return nil, fmt.Errorf("invalid decimal %q", s)
	}
	if neg {
		u.Neg(u)
	}
	d := &Decimal{Unscaled: u, Scale: len(fracPart)}
	d.normalize()
	return d, nil
}

var bigTen = big.NewInt(10)

func (d *Decimal) normalize() {
	if d.Unscaled.Sign() == 0 {
		d.Scale = 0
		return
	}
	q, r := new(big.Int), new(big.Int)
	for d.Scale > 0 {
		q.QuoRem(d.Unscaled, bigTen, r)
		if r.Sign() != 0 {
			break
		}
		d.Unscaled.Set(q)
		d.Scale--
	}
}

// Cmp compares two decimals numerically.
func (d *Decimal) Cmp(o *Decimal) xsd.Order {
	a, b := d.Unscaled, o.Unscaled
	if d.Scale < o.Scale {
		a = new(big.Int).Mul(a, pow10(o.Scale-d.Scale))
	} else if o.Scale < d.Scale {
		b = new(big.Int).Mul(b, pow10(d.Scale-o.Scale))
	}
	return xsd.Order(a.Cmp(b))
}

func pow10(n int) *big.Int {
	return new(big.Int).Exp(bigTen, big.NewInt(int64(n)), nil)
}

// TotalDigits is the number of significant decimal digits in the value
// (minimum digits needed; at least 1, sign and leading/trailing zeros not
// counted — the integer 0 has 1 digit).
func (d *Decimal) TotalDigits() int {
	u := new(big.Int).Abs(d.Unscaled)
	s := u.String()
	if s == "0" {
		return 1
	}
	// The normalized form has no trailing fractional zeros; if the value
	// still ends in zeros they are integer-part zeros, which do count.
	n := len(s)
	if n < d.Scale+1 {
		// Pure fraction like 0.005: unscaled "5", scale 3 → 1 digit.
		n = len(strings.TrimLeft(s, "0"))
	}
	return n
}

// FractionDigits is the number of digits to the right of the decimal point
// in the canonical form.
func (d *Decimal) FractionDigits() int { return d.Scale }

// IsInteger reports whether the value is an integer.
func (d *Decimal) IsInteger() bool { return d.Scale == 0 }

// Int64 returns the value as int64 if it is an integer that fits.
func (d *Decimal) Int64() (int64, bool) {
	if d.Scale != 0 || !d.Unscaled.IsInt64() {
		return 0, false
	}
	return d.Unscaled.Int64(), true
}

// String renders the canonical lexical form.
func (d *Decimal) String() string {
	s := new(big.Int).Abs(d.Unscaled).String()
	sign := ""
	if d.Unscaled.Sign() < 0 {
		sign = "-"
	}
	if d.Scale == 0 {
		return sign + s
	}
	for len(s) <= d.Scale {
		s = "0" + s
	}
	return sign + s[:len(s)-d.Scale] + "." + s[len(s)-d.Scale:]
}

// NewDecimalFromInt builds an integer-valued decimal.
func NewDecimalFromInt(v int64) *Decimal {
	return &Decimal{Unscaled: big.NewInt(v), Scale: 0}
}
