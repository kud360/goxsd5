package xsd

import (
	"regexp"
	"strings"
)

// WhiteSpace is the whiteSpace facet value; WSUnset means "not present"
// (treated as preserve when applied).
type WhiteSpace int

const (
	WSUnset WhiteSpace = iota
	WSPreserve
	WSReplace
	WSCollapse
)

func (w WhiteSpace) String() string {
	switch w {
	case WSPreserve:
		return "preserve"
	case WSReplace:
		return "replace"
	case WSCollapse:
		return "collapse"
	}
	return "unset"
}

// Apply normalizes s per the facet (Part 2 §4.3.6).
func (w WhiteSpace) Apply(s string) string {
	switch w {
	case WSReplace:
		return strings.Map(func(r rune) rune {
			if r == '\t' || r == '\n' || r == '\r' {
				return ' '
			}
			return r
		}, s)
	case WSCollapse:
		fields := strings.FieldsFunc(s, func(r rune) bool {
			return r == ' ' || r == '\t' || r == '\n' || r == '\r'
		})
		return strings.Join(fields, " ")
	}
	return s
}

// Pattern is one compiled pattern facet.
type Pattern struct {
	Source string
	Re     *regexp.Regexp
	Pos    Pos
}

// PatternGroup is the set of pattern facets from one derivation step:
// a value must match at least one pattern in each group (OR within a step,
// AND across steps — Part 2 §4.3.4.3).
type PatternGroup []Pattern

// IntFacet is a length/digits-style facet value.
type IntFacet struct {
	Value int
	Fixed bool
	Pos   Pos
}

// Bound is a min/max bounds facet value.
type Bound struct {
	Value   Value
	Lexical string
	Fixed   bool
	Pos     Pos
}

// Enum is one enumeration member.
type Enum struct {
	Value   Value
	Lexical string
	Pos     Pos
}

// ExplicitTimezone is the XSD 1.1 explicitTimezone facet.
type ExplicitTimezone int

const (
	ETZUnset ExplicitTimezone = iota
	ETZOptional
	ETZRequired
	ETZProhibited
)

// Facets is the ordered facet set of a simple type. The struct shape (not a
// slice) fixes evaluation order by construction; see SimpleType.ParseValue
// for the pipeline.
type Facets struct {
	WhiteSpace      WhiteSpace
	WhiteSpaceFixed bool

	PatternGroups []PatternGroup

	Length    *IntFacet
	MinLength *IntFacet
	MaxLength *IntFacet

	MinInclusive *Bound
	MaxInclusive *Bound
	MinExclusive *Bound
	MaxExclusive *Bound

	TotalDigits    *IntFacet
	FractionDigits *IntFacet

	// HasEnumeration distinguishes "no enumeration facet" from an
	// (illegal) empty one.
	HasEnumeration bool
	Enumeration    []Enum

	Assertions []Assertion

	ExplicitTimezone      ExplicitTimezone
	ExplicitTimezoneFixed bool
	ExplicitTimezonePos   Pos
	WhiteSpacePos         Pos
}

// MergeFacets computes the effective facets of a restriction step: declared
// facets overlay the base's effective facets; pattern groups and assertions
// accumulate. Declaring either min (resp. max) bound clears both inherited
// min (resp. max) bounds — the narrowing checks have already ensured the
// declared bound implies the inherited one.
func MergeFacets(base, declared *Facets) Facets {
	eff := *base
	// Slices are shared with the base; that is fine, they are never
	// mutated after construction.
	if declared.WhiteSpace != WSUnset {
		eff.WhiteSpace = declared.WhiteSpace
		eff.WhiteSpaceFixed = declared.WhiteSpaceFixed
		eff.WhiteSpacePos = declared.WhiteSpacePos
	}
	if len(declared.PatternGroups) > 0 {
		eff.PatternGroups = append(append([]PatternGroup{}, base.PatternGroups...), declared.PatternGroups...)
	}
	if declared.Length != nil {
		eff.Length = declared.Length
	}
	if declared.MinLength != nil {
		eff.MinLength = declared.MinLength
	}
	if declared.MaxLength != nil {
		eff.MaxLength = declared.MaxLength
	}
	if declared.MinInclusive != nil || declared.MinExclusive != nil {
		eff.MinInclusive = declared.MinInclusive
		eff.MinExclusive = declared.MinExclusive
	}
	if declared.MaxInclusive != nil || declared.MaxExclusive != nil {
		eff.MaxInclusive = declared.MaxInclusive
		eff.MaxExclusive = declared.MaxExclusive
	}
	if declared.TotalDigits != nil {
		eff.TotalDigits = declared.TotalDigits
	}
	if declared.FractionDigits != nil {
		eff.FractionDigits = declared.FractionDigits
	}
	if declared.HasEnumeration {
		eff.HasEnumeration = true
		eff.Enumeration = declared.Enumeration
	}
	if len(declared.Assertions) > 0 {
		eff.Assertions = append(append([]Assertion{}, base.Assertions...), declared.Assertions...)
	}
	if declared.ExplicitTimezone != ETZUnset {
		eff.ExplicitTimezone = declared.ExplicitTimezone
		eff.ExplicitTimezoneFixed = declared.ExplicitTimezoneFixed
		eff.ExplicitTimezonePos = declared.ExplicitTimezonePos
	}
	return eff
}

// parseFunc resolves the lexical→value mapping: the nearest ancestor (or
// self) that defines one, ultimately the primitive's.
func (t *SimpleType) parseFunc() ParseFunc {
	for st := t; st != nil; {
		if st.Parse != nil {
			return st.Parse
		}
		st, _ = st.BaseType.(*SimpleType)
	}
	return nil
}

// EffectiveCompare returns the comparison function in effect for t (its
// own or the nearest ancestor's override, defaulting to CompareValues).
// The parser uses it to drive ValidateFacetSet/CheckFacetRestriction.
func (t *SimpleType) EffectiveCompare() CompareFunc { return t.compareFunc() }

// compareFunc resolves value comparison the same way, defaulting to
// CompareValues.
func (t *SimpleType) compareFunc() CompareFunc {
	for st := t; st != nil; {
		if st.Compare != nil {
			return st.Compare
		}
		st, _ = st.BaseType.(*SimpleType)
	}
	return CompareValues
}

// ParseValue runs the full lexical→value facet pipeline:
//
//  1. whiteSpace (lexical)
//  2. pattern groups (lexical)
//  3. lexical→value mapping (the space boundary)
//  4. length facets (value: characters / octets / items)
//  5. bounds, totalDigits, fractionDigits (value)
//  6. enumeration (value)
//  7. explicitTimezone (value, XSD 1.1)
//
// Assertions (stage 8) are stored but not evaluated (no XPath engine).
// ctx may be nil except for QName/NOTATION values.
func (t *SimpleType) ParseValue(lexical string, ctx ValueContext) (Value, error) {
	return t.parseValue(lexical, ctx, false)
}

// ParseFacetValue parses a value used as a constraining-facet value (e.g. the
// value of a derived minInclusive/maxExclusive) against t. It is like
// ParseValue but does NOT apply t's own value-range bound facets (min/max
// inclusive/exclusive): a derived bound is permitted to equal the base's
// corresponding bound (e.g. maxExclusive-valid-restriction allows equality),
// and all bound-vs-bound ordering relationships are validated separately by
// CheckFacetRestriction. The lexical space, patterns, whiteSpace, length,
// digit, and enumeration constraints of t still apply.
func (t *SimpleType) ParseFacetValue(lexical string, ctx ValueContext) (Value, error) {
	return t.parseValue(lexical, ctx, true)
}

func (t *SimpleType) parseValue(lexical string, ctx ValueContext, skipRange bool) (Value, error) {
	f := &t.Facets

	// Stage 1: whiteSpace.
	norm := f.WhiteSpace.Apply(lexical)

	// Stage 2: patterns — at least one match per group.
	for _, group := range f.PatternGroups {
		matched := false
		for i := range group {
			if group[i].Re.MatchString(norm) {
				matched = true
				break
			}
		}
		if !matched {
			// spec: cvc-pattern-valid — XSD 1.1 Part 2 §4.3.4.4 (xmlschema11-2.md#cvc-pattern-valid)
			return nil, NewError(SpecPatternValid, Pos{}, "value %q does not match pattern facet of %s", norm, t.describe())
		}
	}

	// Stage 3: lexical → value.
	v, err := t.buildValue(norm, lexical, ctx)
	if err != nil {
		return nil, err
	}

	// Stage 4: length facets, in the proper unit for the value kind.
	if f.Length != nil || f.MinLength != nil || f.MaxLength != nil {
		if n, ok := ValueLength(v); ok {
			// spec: cvc-length-valid — XSD 1.1 Part 2 §4.3.1.4 (xmlschema11-2.md#cvc-length-valid)
			if f.Length != nil && n != f.Length.Value {
				return nil, NewError(SpecLengthValid, Pos{}, "length %d of %q != required %d", n, norm, f.Length.Value)
			}
			// spec: cvc-minLength-valid — XSD 1.1 Part 2 §4.3.2.4 (xmlschema11-2.md#cvc-minLength-valid)
			if f.MinLength != nil && n < f.MinLength.Value {
				return nil, NewError(SpecMinLengthValid, Pos{}, "length %d of %q < minLength %d", n, norm, f.MinLength.Value)
			}
			// spec: cvc-maxLength-valid — XSD 1.1 Part 2 §4.3.3.4 (xmlschema11-2.md#cvc-maxLength-valid)
			if f.MaxLength != nil && n > f.MaxLength.Value {
				return nil, NewError(SpecMaxLengthValid, Pos{}, "length %d of %q > maxLength %d", n, norm, f.MaxLength.Value)
			}
		}
	}

	// Stage 5: bounds and digit facets (value space).
	cmp := t.compareFunc()
	check := func(b *Bound, ref SpecRef, want func(Order) bool, what string) error {
		if b == nil {
			return nil
		}
		o, ok := cmp(v, b.Value)
		if !ok || !want(o) {
			return NewError(ref, Pos{}, "value %q violates %s %s", norm, what, b.Lexical)
		}
		return nil
	}
	if !skipRange {
		// spec: cvc-minInclusive-valid — XSD 1.1 Part 2 §4.3.10.4 (xmlschema11-2.md#cvc-minInclusive-valid)
		if err := check(f.MinInclusive, SpecMinInclusiveValid, func(o Order) bool { return o >= OrderEqual }, "minInclusive"); err != nil {
			return nil, err
		}
		// spec: cvc-maxInclusive-valid — XSD 1.1 Part 2 §4.3.7.4 (xmlschema11-2.md#cvc-maxInclusive-valid)
		if err := check(f.MaxInclusive, SpecMaxInclusiveValid, func(o Order) bool { return o <= OrderEqual }, "maxInclusive"); err != nil {
			return nil, err
		}
		// spec: cvc-minExclusive-valid — XSD 1.1 Part 2 §4.3.9.4 (xmlschema11-2.md#cvc-minExclusive-valid)
		if err := check(f.MinExclusive, SpecMinExclusiveValid, func(o Order) bool { return o > OrderEqual }, "minExclusive"); err != nil {
			return nil, err
		}
		// spec: cvc-maxExclusive-valid — XSD 1.1 Part 2 §4.3.8.4 (xmlschema11-2.md#cvc-maxExclusive-valid)
		if err := check(f.MaxExclusive, SpecMaxExclusiveValid, func(o Order) bool { return o < OrderEqual }, "maxExclusive"); err != nil {
			return nil, err
		}
	}
	if f.TotalDigits != nil || f.FractionDigits != nil {
		if d, ok := v.(*Decimal); ok {
			// spec: cvc-totalDigits-valid — XSD 1.1 Part 2 §4.3.11.4 (xmlschema11-2.md#cvc-totalDigits-valid)
			if f.TotalDigits != nil && d.TotalDigits() > f.TotalDigits.Value {
				return nil, NewError(SpecTotalDigitsValid, Pos{}, "value %q has %d digits, totalDigits is %d", norm, d.TotalDigits(), f.TotalDigits.Value)
			}
			// spec: cvc-fractionDigits-valid — XSD 1.1 Part 2 §4.3.12.4 (xmlschema11-2.md#cvc-fractionDigits-valid)
			if f.FractionDigits != nil && d.FractionDigits() > f.FractionDigits.Value {
				return nil, NewError(SpecFractionDigitsValid, Pos{}, "value %q has %d fraction digits, fractionDigits is %d", norm, d.FractionDigits(), f.FractionDigits.Value)
			}
		}
	}

	// Stage 6: enumeration (value-space membership).
	if f.HasEnumeration {
		ok := false
		for i := range f.Enumeration {
			if Equal(v, f.Enumeration[i].Value) {
				ok = true
				break
			}
		}
		if !ok {
			// spec: cvc-enumeration-valid — XSD 1.1 Part 2 §4.3.5.4 (xmlschema11-2.md#cvc-enumeration-valid)
			return nil, NewError(SpecEnumerationValid, Pos{}, "value %q not in enumeration of %s", norm, t.describe())
		}
	}

	// Stage 7: explicitTimezone (XSD 1.1).
	if f.ExplicitTimezone == ETZRequired || f.ExplicitTimezone == ETZProhibited {
		if dt, ok := v.(*DateTime); ok {
			// spec: cvc-explicitTimezone-valid — XSD 1.1 Part 2 §4.3.14.4 (xmlschema11-2.md#cvc-explicitTimezone-valid)
			if f.ExplicitTimezone == ETZRequired && !dt.HasTZ {
				return nil, NewError(SpecExplicitTimezoneValid, Pos{}, "value %q must carry a timezone", norm)
			}
			if f.ExplicitTimezone == ETZProhibited && dt.HasTZ {
				return nil, NewError(SpecExplicitTimezoneValid, Pos{}, "value %q must not carry a timezone", norm)
			}
		}
	}

	// Stage 8: assertions — deliberately not evaluated.
	return v, nil
}

// buildValue is stage 3: variety-aware lexical→value mapping.
func (t *SimpleType) buildValue(norm, raw string, ctx ValueContext) (Value, error) {
	switch t.Variety {
	case VarietyList:
		if t.ItemType == nil {
			return nil, NewError(SpecDatatypeValid, t.Pos, "list type %s has no item type", t.describe())
		}
		var items []string
		if norm != "" {
			items = strings.Split(norm, " ")
		}
		list := make(ListValue, 0, len(items))
		for _, it := range items {
			v, err := t.ItemType.ParseValue(it, ctx)
			if err != nil {
				return nil, err
			}
			list = append(list, v)
		}
		return list, nil
	case VarietyUnion:
		var firstErr error
		for _, m := range t.MemberTypes {
			v, err := m.ParseValue(raw, ctx)
			if err == nil {
				return v, nil
			}
			if firstErr == nil {
				firstErr = err
			}
		}
		if firstErr == nil {
			firstErr = NewError(SpecDatatypeValid, t.Pos, "union type %s has no member types", t.describe())
		}
		// spec: cvc-datatype-valid — XSD 1.1 Part 2 §4.1.4 (xmlschema11-2.md#cvc-datatype-valid)
		return nil, NewError(SpecDatatypeValid, Pos{}, "value %q matches no member of union %s", raw, t.describe())
	default:
		pf := t.parseFunc()
		if pf == nil {
			// No mapping anywhere in the chain (anySimpleType
			// descendants get one from the builtin package).
			return String(norm), nil
		}
		return pf(norm, ctx)
	}
}

func (t *SimpleType) describe() string {
	if !t.Name.IsZero() {
		return t.Name.String()
	}
	return "anonymous type"
}
