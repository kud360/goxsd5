// Package xsdtype holds the concrete value-space implementations of the XSD
// built-in datatypes — the lexical→value parsers, the value types
// (String/Boolean/Decimal/DateTime/…), their comparators, and the XSD regex
// engine. It depends on the core model package (xsd) for the value
// abstractions (Value, Order, the capability interfaces) but the core never
// depends on it: the model stays agnostic to any specific datatype, and a
// custom value space can be provided here or by user code the same way.
package xsdtype

import (
	"fmt"
	"math"

	"github.com/kud360/goxsd5/xsd"
)

// String is the value space of xs:string and its kin, xs:anyURI, and any
// other type whose values are just character sequences.
type String string

// Len counts characters (runes), the length-facet unit for strings.
func (s String) Len() int {
	n := 0
	for range string(s) {
		n++
	}
	return n
}

// Boolean is the value space of xs:boolean.
type Boolean bool

// Float is the value space of xs:float (single precision).
type Float float32

// Double is the value space of xs:double.
type Double float64

// Bytes is the value space of xs:hexBinary and xs:base64Binary; the length
// facets measure octets.
type Bytes []byte

// Len counts octets, the length-facet unit for binary values.
func (b Bytes) Len() int { return len(b) }

// QNameValue is the value space of xs:QName and xs:NOTATION: the expanded
// name plus the original lexical form (prefix:local) it came from. It does
// not implement xsd.Lengthed: length facets on QName/NOTATION are deprecated
// and have no effect (Part 2 §4.3.1).
type QNameValue struct {
	Name    xsd.QName
	Lexical string
}

// CompareValues compares two values of any built-in kind. It is the default
// comparator wired onto xs:anySimpleType, so every built-in (and every type
// derived from one without its own Compare override) resolves to it through
// the SimpleType compare-func chain. Cross-kind pairs are incomparable.
func CompareValues(a, b xsd.Value) (xsd.Order, bool) {
	switch av := a.(type) {
	case String:
		if bv, ok := b.(String); ok {
			switch {
			case av < bv:
				return xsd.OrderLess, true
			case av > bv:
				return xsd.OrderGreater, true
			}
			return xsd.OrderEqual, true
		}
	case Boolean:
		if bv, ok := b.(Boolean); ok {
			if av == bv {
				return xsd.OrderEqual, true
			}
			return 0, false
		}
	case Float:
		if bv, ok := b.(Float); ok {
			return compareFloat64(float64(av), float64(bv))
		}
	case Double:
		if bv, ok := b.(Double); ok {
			return compareFloat64(float64(av), float64(bv))
		}
	case *Decimal:
		if bv, ok := b.(*Decimal); ok {
			return av.Cmp(bv), true
		}
	case Bytes:
		if bv, ok := b.(Bytes); ok {
			if string(av) == string(bv) {
				return xsd.OrderEqual, true
			}
			return 0, false
		}
	case QNameValue:
		if bv, ok := b.(QNameValue); ok {
			if av.Name == bv.Name {
				return xsd.OrderEqual, true
			}
			return 0, false
		}
	case *DateTime:
		if bv, ok := b.(*DateTime); ok {
			o, c := av.Compare(bv)
			return xsd.Order(o), c
		}
	case *Duration:
		if bv, ok := b.(*Duration); ok {
			o, c := av.Compare(bv)
			return xsd.Order(o), c
		}
	case xsd.ListValue:
		if bv, ok := b.(xsd.ListValue); ok {
			if len(av) != len(bv) {
				return 0, false
			}
			for i := range av {
				if o, ok := CompareValues(av[i], bv[i]); !ok || o != xsd.OrderEqual {
					return 0, false
				}
			}
			return xsd.OrderEqual, true
		}
	}
	return 0, false
}

// compareFloat64 orders IEEE values: NaN equals itself and is incomparable
// with everything else; -0 == +0; infinities order as usual.
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

// Equal reports value equality (identity in the value space).
func Equal(a, b xsd.Value) bool {
	o, ok := CompareValues(a, b)
	return ok && o == xsd.OrderEqual
}

// ErrNeedContext is returned when QName/NOTATION lexical→value mapping is
// attempted without a ValueContext.
var ErrNeedContext = fmt.Errorf("QName/NOTATION value requires a namespace resolution context")
