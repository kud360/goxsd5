// Package gotype is a Go-native, lax value layer for the XSD built-in
// datatypes — a sibling of builtin/xsdtype that trades xsdtype's bit-exact
// XSD value spaces for the closest ordinary Go type: float64 for the numeric
// family, int64 for xs:integer, time.Time for the date/time family,
// time.Duration for xs:duration, []byte for the binaries, and string for the
// stringy types. The trade is deliberate: a caller who wants values that drop
// straight into Go code, and who accepts the lenience that follows (decimals
// beyond float64 precision round, integers beyond int64 saturate/err, the
// date/time family collapses onto time.Time), opts in at parse time via
// parser.Options.Primitives = gotype.AllBuiltins().
//
// Like xsdtype it is a pure leaf: it depends on the core model package (xsd)
// for the value abstractions and on xsdedit for the NewPrimitive constructor,
// and nothing depends on it. The default parse path is unchanged — gotype is
// reached only when a caller names it.
package gotype

import (
	"math"
	"time"

	"github.com/kud360/goxsd5/xsd"
)

// The lax value spaces. Most lax values are plain Go types (the parser returns
// a bool, float64, int64, time.Time, time.Duration, []byte, or xsd.QName), so a
// consumer can use them directly. Two carry a wrapper only because they need a
// method the facet engine reads: strVal/binVal implement xsd.Lengthed so the
// length facets have a unit (runes / octets). The comparators below give the
// engine the order it needs for bounds and enumeration facets.
type (
	// strVal backs the stringy types (xs:string, xs:anyURI, the token/name
	// ladder); its length unit is runes, matching xsdtype.String.
	strVal string
	// binVal backs xs:hexBinary and xs:base64Binary; its length unit is octets.
	binVal []byte
)

// Len counts runes, the length-facet unit for strings.
func (s strVal) Len() int {
	n := 0
	for range string(s) {
		n++
	}
	return n
}

// Len counts octets, the length-facet unit for binary values.
func (b binVal) Len() int { return len(b) }

// compareValues is the lax default comparator wired onto gotype's
// xs:anySimpleType, reached through the SimpleType compare chain by every lax
// type that does not override it (the temporal and QName primitives do).
// Cross-kind pairs are incomparable, as in xsdtype.CompareValues.
func compareValues(a, b xsd.Value) (xsd.Order, bool) {
	switch av := a.(type) {
	case strVal:
		if bv, ok := b.(strVal); ok {
			return compareOrdered(av, bv), true
		}
	case bool:
		if bv, ok := b.(bool); ok {
			if av == bv {
				return xsd.OrderEqual, true
			}
			return 0, false
		}
	case float64:
		if bv, ok := b.(float64); ok {
			return compareFloat64(av, bv)
		}
	case int64:
		if bv, ok := b.(int64); ok {
			return compareOrdered(av, bv), true
		}
	case binVal:
		if bv, ok := b.(binVal); ok {
			if string(av) == string(bv) {
				return xsd.OrderEqual, true
			}
			return 0, false
		}
	case xsd.ListValue:
		if bv, ok := b.(xsd.ListValue); ok {
			return compareListValues(av, bv)
		}
	}
	return 0, false
}

// compareOrdered orders two values of a Go-ordered kind.
func compareOrdered[T ~string | ~int64](a, b T) xsd.Order {
	switch {
	case a < b:
		return xsd.OrderLess
	case a > b:
		return xsd.OrderGreater
	}
	return xsd.OrderEqual
}

// compareFloat64 orders IEEE values: NaN equals itself and is incomparable with
// everything else; -0 == +0; infinities order as usual (mirrors xsdtype).
func compareFloat64(a, b float64) (xsd.Order, bool) {
	an, bn := math.IsNaN(a), math.IsNaN(b)
	if an || bn {
		if an && bn {
			return xsd.OrderEqual, true
		}
		return 0, false
	}
	switch {
	case a < b:
		return xsd.OrderLess, true
	case a > b:
		return xsd.OrderGreater, true
	}
	return xsd.OrderEqual, true
}

// compareTime is the comparator for the lax date/time family (all on
// time.Time). Wired directly onto the temporal primitives — the default
// compareValues does not name time.Time, since the lax temporal space is the
// only kind that lands there.
func compareTime(a, b xsd.Value) (xsd.Order, bool) {
	av, ok := a.(time.Time)
	if !ok {
		return 0, false
	}
	bv, ok := b.(time.Time)
	if !ok {
		return 0, false
	}
	switch {
	case av.Before(bv):
		return xsd.OrderLess, true
	case av.After(bv):
		return xsd.OrderGreater, true
	}
	return xsd.OrderEqual, true
}

// compareQName is the comparator for the lax xs:QName / xs:NOTATION space
// (xsd.QName, an equality-only space). Wired directly onto those primitives.
func compareQName(a, b xsd.Value) (xsd.Order, bool) {
	av, ok := a.(xsd.QName)
	if !ok {
		return 0, false
	}
	bv, ok := b.(xsd.QName)
	if !ok {
		return 0, false
	}
	if av == bv {
		return xsd.OrderEqual, true
	}
	return 0, false
}

// compareListValues compares two ListValues item by item; equal only when every
// paired item compares equal.
func compareListValues(a, b xsd.ListValue) (xsd.Order, bool) {
	if len(a) != len(b) {
		return 0, false
	}
	for i := range a {
		if o, ok := compareValues(a[i], b[i]); !ok || o != xsd.OrderEqual {
			return 0, false
		}
	}
	return xsd.OrderEqual, true
}
