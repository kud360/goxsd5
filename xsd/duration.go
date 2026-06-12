package xsd

import (
	"fmt"
	"math/big"
	"strings"
)

// Duration is the value space of xs:duration: a months part and an exact
// seconds part (Part 2 §3.3.6 reduces Y/M to months and D/H/M/S to
// seconds). Both parts carry the sign.
type Duration struct {
	Months  int64
	Seconds *big.Rat
}

func (*Duration) isValue() {}

// ParseDuration parses the xs:duration lexical form
// -?PnYnMnDTnHnMnS (each component optional, at least one present,
// fractions only on seconds, T only if a time component follows).
func ParseDuration(s string) (*Duration, error) {
	orig := s
	neg := false
	if strings.HasPrefix(s, "-") {
		neg = true
		s = s[1:]
	}
	if !strings.HasPrefix(s, "P") {
		return nil, fmt.Errorf("invalid duration %q", orig)
	}
	s = s[1:]
	datePart, timePart := s, ""
	if i := strings.IndexByte(s, 'T'); i >= 0 {
		datePart, timePart = s[:i], s[i+1:]
		if timePart == "" {
			return nil, fmt.Errorf("invalid duration %q: T with no time fields", orig)
		}
	}
	if datePart == "" && timePart == "" {
		return nil, fmt.Errorf("invalid duration %q: no fields", orig)
	}

	months := int64(0)
	seconds := new(big.Rat)
	any := false

	// Date fields: Y, M, D in order, integers only.
	rest := datePart
	for _, f := range [3]struct {
		designator byte
		months     int64 // contribution per unit, in months…
		seconds    int64 // …or in seconds
	}{{'Y', 12, 0}, {'M', 1, 0}, {'D', 0, 86400}} {
		if v, ok, r, err := takeDurField(rest, f.designator, false); err != nil {
			return nil, fmt.Errorf("invalid duration %q: %w", orig, err)
		} else if ok {
			rest = r
			any = true
			if f.months != 0 {
				// Mirror parseYear's 2^31 cap so the months total cannot
				// overflow int64.
				if n := v.Num(); !n.IsInt64() || n.Int64() > 1<<31 {
					return nil, fmt.Errorf("invalid duration %q: %c field out of range", orig, f.designator)
				}
				months += v.Num().Int64() * f.months
			}
			if f.seconds != 0 {
				seconds.Add(seconds, new(big.Rat).Mul(v, new(big.Rat).SetInt64(f.seconds)))
			}
		}
	}
	if rest != "" {
		return nil, fmt.Errorf("invalid duration %q", orig)
	}
	rest = timePart
	for _, f := range [3]struct {
		designator byte
		secs       int64
		frac       bool
	}{{'H', 3600, false}, {'M', 60, false}, {'S', 1, true}} {
		if v, ok, r, err := takeDurField(rest, f.designator, f.frac); err != nil {
			return nil, fmt.Errorf("invalid duration %q: %w", orig, err)
		} else if ok {
			rest = r
			any = true
			seconds.Add(seconds, new(big.Rat).Mul(v, new(big.Rat).SetInt64(f.secs)))
		}
	}
	if rest != "" || !any {
		return nil, fmt.Errorf("invalid duration %q", orig)
	}
	if neg {
		months = -months
		seconds.Neg(seconds)
	}
	return &Duration{Months: months, Seconds: seconds}, nil
}

// takeDurField consumes a leading "<digits><designator>" if the next field
// is this designator (fields are ordered, so a non-match leaves rest as-is).
func takeDurField(s string, designator byte, allowFrac bool) (v *big.Rat, ok bool, rest string, err error) {
	if s == "" {
		return nil, false, s, nil
	}
	i := 0
	for i < len(s) && (s[i] >= '0' && s[i] <= '9' || s[i] == '.') {
		i++
	}
	if i == 0 || i >= len(s) || s[i] != designator {
		return nil, false, s, nil
	}
	num := s[:i]
	if dot := strings.IndexByte(num, '.'); dot >= 0 {
		if !allowFrac || dot == 0 || dot == len(num)-1 || strings.IndexByte(num[dot+1:], '.') >= 0 {
			return nil, false, "", fmt.Errorf("bad field %q", num)
		}
	}
	r, okp := new(big.Rat).SetString(num)
	if !okp {
		return nil, false, "", fmt.Errorf("bad field %q", num)
	}
	return r, true, s[i+1:], nil
}

// durationRefs are the four reference dateTimes of Part 2 §3.3.6.2 used to
// define the partial order on durations.
var durationRefs = []*DateTime{
	{Kind: KindDateTime, Year: 1696, Month: 9, Day: 1, HasTZ: true},
	{Kind: KindDateTime, Year: 1697, Month: 2, Day: 1, HasTZ: true},
	{Kind: KindDateTime, Year: 1903, Month: 3, Day: 1, HasTZ: true},
	{Kind: KindDateTime, Year: 1903, Month: 7, Day: 1, HasTZ: true},
}

// Compare implements the duration partial order: d < o iff ref+d < ref+o
// for all four reference dateTimes.
func (d *Duration) Compare(o *Duration) (Order, bool) {
	if d.Months == o.Months {
		return Order(d.Seconds.Cmp(o.Seconds)), true
	}
	if d.Seconds.Cmp(o.Seconds) == 0 {
		switch {
		case d.Months < o.Months:
			return OrderLess, true
		default:
			return OrderGreater, true
		}
	}
	var first Order
	for i, ref := range durationRefs {
		a := ref.AddDuration(d)
		b := ref.AddDuration(o)
		ord, ok := a.Compare(b)
		if !ok {
			// Cannot happen (both operands are timezoned dateTimes of the
			// same kind), but incomparable is the safe answer if it did.
			return 0, false
		}
		if i == 0 {
			first = ord
		} else if ord != first {
			return 0, false
		}
	}
	if first == OrderEqual {
		return 0, false // equal under all refs but different components: incomparable
	}
	return first, true
}
