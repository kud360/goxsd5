package xsd

// Value is a member of some simple type's value space. It is deliberately
// open (any): the concrete value implementations live in the builtin layer
// (builtin/xsdtype), and user code may introduce its own. The facet engine
// never names a concrete value type — it discovers what a value can do
// through the capability interfaces below (Lengthed/DigitCounted/
// TimezoneAware) and the per-type CompareFunc, so a custom primitive with
// its own parser and comparator participates in the machinery without any
// change here.
type Value any

// Lengthed is implemented by values the length facets (length/minLength/
// maxLength) can measure: the unit is the value space's own (characters for
// strings, octets for binaries, items for lists). A value that does not
// implement it is exempt from length facets (e.g. QName/NOTATION) — the
// interface's presence is itself the "length is defined" signal. The method
// is Len() int, matching the stdlib convention so the value types satisfy a
// plain interface{ Len() int } too.
type Lengthed interface {
	Len() int
}

// DigitCounted is implemented by values the totalDigits/fractionDigits
// facets apply to (xs:decimal and its derivations).
type DigitCounted interface {
	TotalDigits() int
	FractionDigits() int
}

// Scaled is implemented by values that carry a scale (arithmeticPrecision)
// distinct from their numerical value — xs:precisionDecimal, whose cohort
// members (3, 3.0, 3.00) share a numerical value but differ by scale. The
// bool reports whether a scale is defined: it is absent for the special
// values (±INF/NaN). The fractionDigits/totalDigits facets on precisionDecimal
// read scale through this interface.
type Scaled interface {
	Scale() (int, bool)
}

// TimezoneAware is implemented by values the explicitTimezone facet applies
// to (the date/time value spaces): it reports whether the value carries a
// timezone offset.
type TimezoneAware interface {
	HasTimezone() bool
}

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
// equality-only spaces return (OrderEqual, true) only on equality). The
// concrete default comparator over the built-in value spaces lives in
// builtin/xsdtype (CompareValues); the core engine reaches it through a
// SimpleType's effective CompareFunc, never by naming a value type.
type CompareFunc func(a, b Value) (Order, bool)

// ListValue is the value space of list types. It is structural (intrinsic to
// the list variety) and so lives in the core model, unlike the atomic value
// implementations.
type ListValue []Value

// Len counts items, the length-facet unit for list values.
func (l ListValue) Len() int { return len(l) }

// ValueLength is the facet-relevant length of a value: characters for
// strings/URIs, octets for binaries, items for lists. The bool reports
// whether length is defined for this value kind — it is exactly whether the
// value implements Lengthed.
func ValueLength(v Value) (int, bool) {
	if l, ok := v.(Lengthed); ok {
		return l.Len(), true
	}
	return 0, false
}
