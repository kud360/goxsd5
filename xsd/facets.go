package xsd

import (
	"errors"
	"fmt"
	"math/big"
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

// scaleFacetCap bounds the magnitude of a parsed scale-facet value so an
// absurd authored maxScale/minScale still fits an int; a real scale never
// approaches it.
const scaleFacetCap = 1 << 31

// ParseScaleFacetInt parses the value of a precisionDecimal maxScale/minScale
// facet: an xs:integer, so SIGNED and unbounded (unlike the non-negative
// length/digits facets). The magnitude is saturated at scaleFacetCap to fit an
// int. It is the signed sibling of the parser's nonNegativeInteger helper and
// is the facet-element parser's lexical→int mapping for the scale facets.
func ParseScaleFacetInt(v string) (int, error) {
	i, ok := new(big.Int).SetString(strings.TrimSpace(v), 10)
	if !ok {
		return 0, fmt.Errorf("%q is not an integer", v)
	}
	if !i.IsInt64() || i.Int64() > scaleFacetCap {
		return i.Sign() * scaleFacetCap, nil
	}
	if i.Int64() < -scaleFacetCap {
		return -scaleFacetCap, nil
	}
	return int(i.Int64()), nil
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

func (e ExplicitTimezone) String() string {
	switch e {
	case ETZOptional:
		return "optional"
	case ETZRequired:
		return "required"
	case ETZProhibited:
		return "prohibited"
	default:
		return "unset"
	}
}

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

	// MaxScale and MinScale are the precisionDecimal-specific scale facets
	// (Part 2 precisionDecimal Note §4.2/§4.3). Unlike the digit facets their
	// values are SIGNED (a scale may be negative, e.g. "3.0e2" has scale −1).
	MaxScale *IntFacet
	MinScale *IntFacet

	Enumeration []Enum

	Assertions []Assertion

	ExplicitTimezone      ExplicitTimezone
	ExplicitTimezoneFixed bool
	ExplicitTimezonePos   Pos
	WhiteSpacePos         Pos
}

// HasEnumeration reports whether an enumeration facet constrains the value
// space. A valid enumeration always carries at least one value, so presence is
// equivalent to a non-empty value list; an authored enumeration whose values
// all failed base-type validation is already reported per value and leaves no
// surviving enumeration here.
func (f Facets) HasEnumeration() bool { return len(f.Enumeration) > 0 }

// FacetSet is a set of constraining-facet categories. It expresses facet
// applicability (cos-applicable-facets, Part 2 §4.1.6 / the per-facet
// applicability lists): which facets a given value space admits. A primitive
// declares its set in SimpleType.Applicable; restrictions inherit it and
// list/union varieties have fixed sets (see ApplicableFacets).
type FacetSet uint

const (
	FacetLength FacetSet = 1 << iota
	FacetMinLength
	FacetMaxLength
	FacetPattern
	FacetEnumeration
	FacetWhiteSpace
	FacetBounds // the four min/max inclusive/exclusive facets
	FacetTotalDigits
	FacetFractionDigits
	FacetMaxScale // precisionDecimal §4.2
	FacetMinScale // precisionDecimal §4.3
	FacetAssertion
	FacetExplicitTimezone
)

// Convenience groupings.
const (
	FacetsLength = FacetLength | FacetMinLength | FacetMaxLength
	FacetsCommon = FacetPattern | FacetEnumeration | FacetWhiteSpace | FacetAssertion
	AllFacets    = ^FacetSet(0)
)

// Has reports whether every facet in want is present in s.
func (s FacetSet) Has(want FacetSet) bool { return s&want == want }

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
	if declared.MaxScale != nil {
		eff.MaxScale = declared.MaxScale
	}
	if declared.MinScale != nil {
		eff.MinScale = declared.MinScale
	}
	if declared.HasEnumeration() {
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

// EffectiveFacets computes the type's effective facet set: the facets declared
// on this step overlaid on every ancestor's, in derivation order (the narrowing
// between steps is validated at build time). It is derived on demand from
// DeclaredFacets and the base chain — nothing is cached — so it always reflects
// the current DeclaredFacets. The returned pointer is to a fresh value whose
// facet slices alias the declared ones; treat it as read-only (use the xsdedit
// mutation helpers to change facets).
func (t *SimpleType) EffectiveFacets() *Facets {
	var eff Facets
	if base, ok := t.BaseType.(*SimpleType); ok {
		eff = MergeFacets(base.EffectiveFacets(), &t.DeclaredFacets)
	} else {
		eff = t.DeclaredFacets
	}
	// A list's lexical space is whiteSpace-collapsed and fixed (Part 2 §4.3.6);
	// this holds for the list variety itself and every restriction of one.
	if t.Variety == VarietyList {
		eff.WhiteSpace = WSCollapse
		eff.WhiteSpaceFixed = true
	}
	return &eff
}

// EffectiveCompare returns the comparison function in effect for t (its
// own or the nearest ancestor's override, defaulting to CompareValues).
// The parser uses it to drive ValidateFacetSet/CheckFacetRestriction.
func (t *SimpleType) EffectiveCompare() CompareFunc { return t.compareFunc() }

// compareFunc resolves value comparison the same way. The default comparator
// over the built-in value spaces is wired onto xs:anySimpleType by the
// builtin layer, so the chain walk reaches it for every real type; the
// incomparable fallback applies only to a detached type with no comparator
// anywhere in its chain.
func (t *SimpleType) compareFunc() CompareFunc {
	for st := t; st != nil; {
		if st.Compare != nil {
			return st.Compare
		}
		st, _ = st.BaseType.(*SimpleType)
	}
	return incomparable
}

// incomparable is the fallback CompareFunc: it reports every pair as
// incomparable. Order-based facets against such a value space fail closed.
func incomparable(a, b Value) (Order, bool) { return 0, false }

// ParseValue runs the full lexical→value facet pipeline:
//
//  1. whiteSpace (lexical)
//  2. pattern groups (lexical)
//  3. lexical→value mapping (the space boundary)
//  4. length facets (value: characters / octets / items)
//  5. bounds, totalDigits, fractionDigits, maxScale/minScale (value)
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
	f := t.EffectiveFacets()

	// Stage 1: whiteSpace.
	norm := f.WhiteSpace.Apply(lexical)

	// Stage 2: patterns — at least one match per group. For a union, the pattern
	// is applied to the value as normalized by the validating MEMBER's whiteSpace
	// (the union has none of its own), so the check is deferred to buildValue,
	// which knows which member matched (cvc-pattern-valid; Saxon issue 2247).
	if t.Variety != VarietyUnion && !patternsMatch(f.PatternGroups, norm) {
		// spec: cvc-pattern-valid — XSD 1.1 Part 2 §4.3.4.4 (xmlschema11-2.md#cvc-pattern-valid)
		return nil, NewError(SpecPatternValid, Pos{}, "value %q does not match pattern facet of %s", norm, t.describe())
	}

	// Stage 3: lexical → value.
	v, err := t.buildValue(norm, lexical, ctx)
	if err != nil {
		return nil, err
	}

	// Stages 4-7: facet checks against the typed value, in spec order. Each
	// stage is its own helper; the first violation wins.
	cmp := t.compareFunc()
	if err := t.checkLengthFacets(f, v, norm); err != nil {
		return nil, err
	}
	if err := t.checkRangeFacets(f, v, norm, cmp, skipRange); err != nil {
		return nil, err
	}
	if err := t.checkScaleFacets(f, v, norm); err != nil {
		return nil, err
	}
	if err := t.checkEnumeration(f, v, norm, cmp); err != nil {
		return nil, err
	}
	if err := t.checkTimezoneFacet(f, v, norm); err != nil {
		return nil, err
	}

	// Stage 8: assertions — deliberately not evaluated.
	return v, nil
}

// checkLengthFacets is stage 4: length/minLength/maxLength in the value's
// natural unit (skipped when the value kind has no length).
func (t *SimpleType) checkLengthFacets(f *Facets, v Value, norm string) error {
	if f.Length == nil && f.MinLength == nil && f.MaxLength == nil {
		return nil
	}
	n, ok := ValueLength(v)
	if !ok {
		return nil
	}
	// spec: cvc-length-valid — XSD 1.1 Part 2 §4.3.1.4 (xmlschema11-2.md#cvc-length-valid)
	if f.Length != nil && n != f.Length.Value {
		return NewError(SpecLengthValid, Pos{}, "length %d of %q != required %d", n, norm, f.Length.Value)
	}
	// spec: cvc-minLength-valid — XSD 1.1 Part 2 §4.3.2.4 (xmlschema11-2.md#cvc-minLength-valid)
	if f.MinLength != nil && n < f.MinLength.Value {
		return NewError(SpecMinLengthValid, Pos{}, "length %d of %q < minLength %d", n, norm, f.MinLength.Value)
	}
	// spec: cvc-maxLength-valid — XSD 1.1 Part 2 §4.3.3.4 (xmlschema11-2.md#cvc-maxLength-valid)
	if f.MaxLength != nil && n > f.MaxLength.Value {
		return NewError(SpecMaxLengthValid, Pos{}, "length %d of %q > maxLength %d", n, norm, f.MaxLength.Value)
	}
	return nil
}

// checkRangeFacets is stage 5: inclusive/exclusive bounds and digit facets,
// compared in the value space. skipRange suppresses the bound checks.
func (t *SimpleType) checkRangeFacets(f *Facets, v Value, norm string, cmp CompareFunc, skipRange bool) error {
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
			return err
		}
		// spec: cvc-maxInclusive-valid — XSD 1.1 Part 2 §4.3.7.4 (xmlschema11-2.md#cvc-maxInclusive-valid)
		if err := check(f.MaxInclusive, SpecMaxInclusiveValid, func(o Order) bool { return o <= OrderEqual }, "maxInclusive"); err != nil {
			return err
		}
		// spec: cvc-minExclusive-valid — XSD 1.1 Part 2 §4.3.9.4 (xmlschema11-2.md#cvc-minExclusive-valid)
		if err := check(f.MinExclusive, SpecMinExclusiveValid, func(o Order) bool { return o > OrderEqual }, "minExclusive"); err != nil {
			return err
		}
		// spec: cvc-maxExclusive-valid — XSD 1.1 Part 2 §4.3.8.4 (xmlschema11-2.md#cvc-maxExclusive-valid)
		if err := check(f.MaxExclusive, SpecMaxExclusiveValid, func(o Order) bool { return o < OrderEqual }, "maxExclusive"); err != nil {
			return err
		}
	}
	if f.TotalDigits == nil && f.FractionDigits == nil {
		return nil
	}
	d, ok := v.(DigitCounted)
	if !ok {
		return nil
	}
	// spec: cvc-totalDigits-valid — XSD 1.1 Part 2 §4.3.11.4 (xmlschema11-2.md#cvc-totalDigits-valid)
	if f.TotalDigits != nil && d.TotalDigits() > f.TotalDigits.Value {
		return NewError(SpecTotalDigitsValid, Pos{}, "value %q has %d digits, totalDigits is %d", norm, d.TotalDigits(), f.TotalDigits.Value)
	}
	// spec: cvc-fractionDigits-valid — XSD 1.1 Part 2 §4.3.12.4 (xmlschema11-2.md#cvc-fractionDigits-valid)
	if f.FractionDigits != nil && d.FractionDigits() > f.FractionDigits.Value {
		return NewError(SpecFractionDigitsValid, Pos{}, "value %q has %d fraction digits, fractionDigits is %d", norm, d.FractionDigits(), f.FractionDigits.Value)
	}
	return nil
}

// checkScaleFacets is the precisionDecimal scale stage: maxScale/minScale read
// the value's scale (arithmeticPrecision) through the Scaled capability. A value
// with no scale — the special values NaN/±INF — is exempt and always passes
// (precisionDecimal Note §4.2/§4.3 clause 2).
func (t *SimpleType) checkScaleFacets(f *Facets, v Value, norm string) error {
	if f.MaxScale == nil && f.MinScale == nil {
		return nil
	}
	s, ok := v.(Scaled)
	if !ok {
		return nil
	}
	scale, present := s.Scale()
	if !present {
		return nil
	}
	// spec: cvc-maxScale-valid — XSD 1.1 precisionDecimal Note §4.2 (cvc-maxScale-valid)
	if f.MaxScale != nil && scale > f.MaxScale.Value {
		return NewError(SpecMaxScaleValid, Pos{}, "value %q has scale %d, maxScale is %d", norm, scale, f.MaxScale.Value)
	}
	// spec: cvc-minScale-valid — XSD 1.1 precisionDecimal Note §4.3 (cvc-minScale-valid)
	if f.MinScale != nil && scale < f.MinScale.Value {
		return NewError(SpecMinScaleValid, Pos{}, "value %q has scale %d, minScale is %d", norm, scale, f.MinScale.Value)
	}
	return nil
}

// checkEnumeration is stage 6: value-space membership. Membership is by the
// identity relation: a value implementing Identical is matched through it (so
// precisionDecimal NaN enumerates to NaN despite being order-incomparable),
// otherwise the type's effective order comparator's OrderEqual stands in, which
// honors a custom value space's own equality.
func (t *SimpleType) checkEnumeration(f *Facets, v Value, norm string, cmp CompareFunc) error {
	if !f.HasEnumeration() {
		return nil
	}
	for i := range f.Enumeration {
		if valuesIdentical(v, f.Enumeration[i].Value, cmp) {
			return nil
		}
	}
	// spec: cvc-enumeration-valid — XSD 1.1 Part 2 §4.3.5.4 (xmlschema11-2.md#cvc-enumeration-valid)
	return NewError(SpecEnumerationValid, Pos{}, "value %q not in enumeration of %s", norm, t.describe())
}

// valuesIdentical reports value-space identity for enumeration matching. A
// value exposing the Identical capability defines its own identity relation
// (distinct from its order relation); otherwise identity is the order
// comparator reporting OrderEqual.
func valuesIdentical(v, want Value, cmp CompareFunc) bool {
	if id, ok := v.(Identical); ok {
		return id.Identical(want)
	}
	o, c := cmp(v, want)
	return c && o == OrderEqual
}

// checkTimezoneFacet is stage 7: explicitTimezone (XSD 1.1).
func (t *SimpleType) checkTimezoneFacet(f *Facets, v Value, norm string) error {
	if f.ExplicitTimezone != ETZRequired && f.ExplicitTimezone != ETZProhibited {
		return nil
	}
	tz, ok := v.(TimezoneAware)
	if !ok {
		return nil
	}
	// spec: cvc-explicitTimezone-valid — XSD 1.1 Part 2 §4.3.14.4 (xmlschema11-2.md#cvc-explicitTimezone-valid)
	if f.ExplicitTimezone == ETZRequired && !tz.HasTimezone() {
		return NewError(SpecExplicitTimezoneValid, Pos{}, "value %q must carry a timezone", norm)
	}
	if f.ExplicitTimezone == ETZProhibited && tz.HasTimezone() {
		return NewError(SpecExplicitTimezoneValid, Pos{}, "value %q must not carry a timezone", norm)
	}
	return nil
}

// patternsMatch reports whether s satisfies every pattern group (each group
// requires at least one of its alternatives to match — a conjunction of
// per-derivation-step disjunctions).
func patternsMatch(groups []PatternGroup, s string) bool {
	for _, group := range groups {
		matched := false
		for i := range group {
			if group[i].Re.MatchString(s) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return true
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
		// Validate against the DIRECT member types, not the flattened basic
		// members: a member that is itself a union (possibly a restriction adding
		// facets) must validate the value through its own facets, so an
		// intervening restriction-of-union's pattern/enumeration is enforced
		// (cvc-datatype-valid; e.g. union(restriction(union(string), pattern))).
		// The union's own pattern facets are applied here, against the value as
		// normalized by the matching member's whiteSpace.
		patterns := t.EffectiveFacets().PatternGroups
		for _, m := range t.DirectMembers {
			v, err := m.ParseValue(raw, ctx)
			if err != nil {
				continue
			}
			if !patternsMatch(patterns, m.EffectiveFacets().WhiteSpace.Apply(raw)) {
				continue // member validates but the union pattern rejects its value
			}
			return v, nil
		}
		// spec: cvc-datatype-valid — XSD 1.1 Part 2 §4.1.4 (xmlschema11-2.md#cvc-datatype-valid)
		return nil, NewError(SpecDatatypeValid, Pos{}, "value %q matches no member of union %s", raw, t.describe())
	default:
		pf := t.parseFunc()
		if pf == nil {
			// No lexical→value mapping anywhere in the chain. Every real
			// atomic type resolves one (xs:anySimpleType carries an
			// identity parser installed by the builtin layer), so this is
			// only reached by a malformed, detached type.
			return nil, NewError(SpecDatatypeValid, t.Pos, "atomic type %s has no lexical mapping", t.describe())
		}
		v, err := pf(norm, ctx)
		if err != nil {
			// The builtin lexical→value parsers return plain errors; attach the
			// governing clause here at the value-space boundary so every atomic
			// datatype failure carries cvc-datatype-valid like the rest of the
			// facet checks. An error already carrying a SpecRef is left intact.
			// spec: cvc-datatype-valid — XSD 1.1 Part 2 §4.1.4 (xmlschema11-2.md#cvc-datatype-valid)
			var xe *Error
			if !errors.As(err, &xe) {
				return nil, WrapError(SpecDatatypeValid, err)
			}
			return nil, err
		}
		return v, nil
	}
}

func (t *SimpleType) describe() string {
	if !t.Name.IsZero() {
		return t.Name.String()
	}
	return "anonymous type"
}
