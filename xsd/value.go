package xsd

import (
	"fmt"
	"math"
)

// Value is a member of some simple type's value space. The set of
// implementations is sealed: every primitive maps into one of the types in
// this file (plus Decimal, DateTime, Duration in their own files).
type Value interface{ isValue() }

// ValueContext supplies what a bare lexical form cannot carry — namespace
// bindings for QName/NOTATION resolution.
type ValueContext interface {
	ResolveQName(prefix, local string) (QName, bool)
}

// ParseFunc maps a (whitespace-normalized) lexical form to a value.
type ParseFunc func(lexical string, ctx ValueContext) (Value, error)

// Order is the result of comparing two values.
type Order int

const (
	OrderLess    Order = -1
	OrderEqual   Order = 0
	OrderGreater Order = 1
)

// CompareFunc compares two values; the bool reports comparability (value
// spaces with partial orders return false for incomparable pairs, and
// equality-only spaces return (OrderEqual, true) only on equality).
type CompareFunc func(a, b Value) (Order, bool)

// String is the value space of xs:string and its kin, xs:anyURI, and any
// other type whose values are just character sequences.
type String string

func (String) isValue() {}

// Boolean is the value space of xs:boolean.
type Boolean bool

func (Boolean) isValue() {}

// Float is the value space of xs:float (single precision).
type Float float32

func (Float) isValue() {}

// Double is the value space of xs:double.
type Double float64

func (Double) isValue() {}

// Bytes is the value space of xs:hexBinary and xs:base64Binary; the length
// facets measure octets.
type Bytes []byte

func (Bytes) isValue() {}

// QNameValue is the value space of xs:QName and xs:NOTATION: the expanded
// name plus the original lexical form (prefix:local) it came from.
type QNameValue struct {
	Name    QName
	Lexical string
}

func (QNameValue) isValue() {}

// ListValue is the value space of list types.
type ListValue []Value

func (ListValue) isValue() {}

// CompareValues compares two values of any sealed kind. It implements the
// default (primitive-ancestor) comparison used when a SimpleType does not
// override Compare. Cross-kind pairs are incomparable.
func CompareValues(a, b Value) (Order, bool) {
	switch av := a.(type) {
	case String:
		if bv, ok := b.(String); ok {
			switch {
			case av < bv:
				return OrderLess, true
			case av > bv:
				return OrderGreater, true
			}
			return OrderEqual, true
		}
	case Boolean:
		if bv, ok := b.(Boolean); ok {
			if av == bv {
				return OrderEqual, true
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
				return OrderEqual, true
			}
			return 0, false
		}
	case QNameValue:
		if bv, ok := b.(QNameValue); ok {
			if av.Name == bv.Name {
				return OrderEqual, true
			}
			return 0, false
		}
	case *DateTime:
		if bv, ok := b.(*DateTime); ok {
			return av.Compare(bv)
		}
	case *Duration:
		if bv, ok := b.(*Duration); ok {
			return av.Compare(bv)
		}
	case ListValue:
		if bv, ok := b.(ListValue); ok {
			if len(av) != len(bv) {
				return 0, false
			}
			for i := range av {
				if o, ok := CompareValues(av[i], bv[i]); !ok || o != OrderEqual {
					return 0, false
				}
			}
			return OrderEqual, true
		}
	}
	return 0, false
}

// compareFloat64 orders IEEE values: NaN equals itself and is incomparable
// with everything else; -0 == +0; infinities order as usual.
func compareFloat64(a, b float64) (Order, bool) {
	an, bn := math.IsNaN(a), math.IsNaN(b)
	if an || bn {
		if an && bn {
			return OrderEqual, true
		}
		return 0, false
	}
	switch {
	case a < b:
		return OrderLess, true
	case a > b:
		return OrderGreater, true
	}
	return OrderEqual, true
}

// Equal reports value equality (identity in the value space).
func Equal(a, b Value) bool {
	o, ok := CompareValues(a, b)
	return ok && o == OrderEqual
}

// ValueLength is the facet-relevant length of a value: characters for
// strings/URIs, octets for binaries, items for lists. The bool reports
// whether length is defined for this value kind.
func ValueLength(v Value) (int, bool) {
	switch v := v.(type) {
	case String:
		n := 0
		for range string(v) {
			n++
		}
		return n, true
	case Bytes:
		return len(v), true
	case ListValue:
		return len(v), true
	case QNameValue:
		// length facets on QName/NOTATION are deprecated and must not
		// constrain anything (Part 2 §4.3.1: no effect).
		return 0, false
	}
	return 0, false
}

// ErrNeedContext is returned when QName/NOTATION lexical→value mapping is
// attempted without a ValueContext.
var ErrNeedContext = fmt.Errorf("QName/NOTATION value requires a namespace resolution context")
