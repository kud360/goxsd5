// Package xsdedit is the safe programmatic mutation / extension API for the
// xsd schema-component model. Each mutator re-runs the same facet-narrowing
// and intra-facet checks the parser enforces (CheckFacetRestriction +
// ValidateFacetSet), so a programmatically built or edited type can never
// bypass a clause the parser would have caught. All in-place mutators are
// copy-on-write: validation runs against a candidate and the target is
// committed only on success — a rejected mutation leaves the original
// untouched.
//
// These are free functions rather than methods because they live outside the
// xsd package; they operate only on xsd's exported surface.
package xsdedit

import (
	"github.com/kud360/goxsd5/xsd"
	"github.com/kud360/goxsd5/xsdregex"
)

// RestrictWith derives a new anonymous simple type from t by restriction with
// the given declared facets, mirroring the parser's applyRestriction: it runs
// the narrowing checks (CheckFacetRestriction) of each declared facet against
// t's effective facets, then validates the merged result (ValidateFacetSet).
// On any violation it returns the error and nil, and t is unchanged. The
// returned type has no name; set Name (and Extensions, to carry foreign
// content) on it as needed.
func RestrictWith(t *xsd.SimpleType, declared *xsd.Facets) (*xsd.SimpleType, error) {
	cmp := t.EffectiveCompare()
	if err := xsd.CheckFacetRestriction(declared, t.EffectiveFacets(), cmp); err != nil {
		return nil, err
	}
	eff := xsd.MergeFacets(t.EffectiveFacets(), declared)
	if err := xsd.ValidateFacetSet(&eff, cmp); err != nil {
		return nil, err
	}
	st := &xsd.SimpleType{
		BaseType:       t,
		Variety:        t.Variety,
		ItemType:       t.ItemType,
		DirectMembers:  t.DirectMembers, // BasicMembers() derives from these
		DeclaredFacets: *declared,
		// The effective facets and the fundamental facets (Part 2 §F) are derived
		// on demand from DeclaredFacets and the base chain by EffectiveFacets()
		// and Fundamentals(); nothing to copy from the base.
	}
	return st, nil
}

// reservedType reports whether t belongs to a spec-reserved namespace — the XSD
// built-ins (XSDNS) or the XSI attribute types (XSINS). Those types are shared,
// immutable singletons owned by the library: mutating one in place would corrupt
// it for every schema in the process, so the in-place mutators refuse it. A
// user's own type, including a future custom primitive, lives in the user's
// namespace and is freely mutable.
func reservedType(t *xsd.SimpleType) bool {
	return t.Name.Namespace == xsd.XSDNS || t.Name.Namespace == xsd.XSINS
}

// AddEnumeration adds enumeration members to t in place. Each lexical must be
// valid against t's base type (enumeration-valid-restriction). The values
// accumulate onto t's own derivation step: calling it repeatedly builds up
// one enumeration set rather than a chain of single-value restrictions. If any
// lexical is invalid the whole call is rejected and t is unchanged.
//
// Built-in types are shared and never mutated in place; derive a subtype with
// RestrictWith and add enumerations to that instead.
func AddEnumeration(t *xsd.SimpleType, lexicals ...string) error {
	if reservedType(t) {
		return xsd.NewError(xsd.SpecEnumerationValidRestriction, t.Pos, "cannot mutate the built-in type %s in place; use RestrictWith to derive a subtype", t.Name)
	}
	base, ok := t.BaseType.(*xsd.SimpleType)
	if !ok {
		return xsd.NewError(xsd.SpecEnumerationValidRestriction, t.Pos, "type %s has no simple base to draw enumeration values from", t.Name)
	}
	add := make([]xsd.Enum, 0, len(lexicals))
	for _, lex := range lexicals {
		v, err := base.ParseValue(lex, nil)
		if err != nil {
			// spec: enumeration-valid-restriction — XSD 1.1 Part 2 §4.3.5.5:
			// every enumeration value must be valid against the base type.
			return xsd.NewError(xsd.SpecEnumerationValidRestriction, t.Pos, "enumeration value %q is not valid against the base type %s: %v", lex, base.TypeName(), err)
		}
		add = append(add, xsd.Enum{Value: v, Lexical: lex})
	}
	// Copy-on-write: build the candidate declared facets, re-derive and
	// validate, and commit only if the result is sound.
	declared := t.DeclaredFacets
	declared.Enumeration = append(append([]xsd.Enum{}, t.DeclaredFacets.Enumeration...), add...)
	eff := xsd.MergeFacets(base.EffectiveFacets(), &declared)
	if err := xsd.ValidateFacetSet(&eff, t.EffectiveCompare()); err != nil {
		return err
	}
	t.DeclaredFacets = declared
	// EffectiveFacets() now derives the enumeration from DeclaredFacets, and
	// {cardinality} becomes finite by virtue of it; Fundamentals() reflects that
	// without a stored field.
	return nil
}

// AddPattern adds a pattern facet to t in place as a new single-pattern group
// (matched in addition to every inherited group: AND across groups). The
// pattern is compiled with the XSD regex translator; a malformed pattern is
// rejected and t is unchanged. Built-ins may not be mutated in place.
func AddPattern(t *xsd.SimpleType, pattern string) error {
	if reservedType(t) {
		return xsd.NewError(xsd.SpecRegexValid, t.Pos, "cannot mutate the built-in type %s in place; use RestrictWith to derive a subtype", t.Name)
	}
	re, err := xsdregex.CompileRegex(pattern)
	if err != nil {
		// spec: regex-valid — XSD 1.1 Part 2 Appendix G
		return xsd.NewError(xsd.SpecRegexValid, t.Pos, "invalid pattern: %v", err)
	}
	g := xsd.PatternGroup{{Source: pattern, Re: re}}
	t.DeclaredFacets.PatternGroups = append(t.DeclaredFacets.PatternGroups, g)
	return nil
}

// AddElement appends an element particle (the element occurring min..max
// times; max < 0 means unbounded) to t's element content, mirroring the
// parser's particle construction. It is valid only on a complex type with
// element-only or mixed content: a simple-content or empty-base type has no
// element content to extend, and an error is returned with t unchanged.
func AddElement(t *xsd.ComplexType, elem *xsd.ElementDecl, min, max int) error {
	if min < 0 || (max >= 0 && max < min) {
		// spec: p-props-correct.2.1 — XSD 1.1 Part 1 §3.9.6: 0 ≤ minOccurs ≤ maxOccurs.
		return xsd.NewError(xsd.SpecPPropsCorrect, t.Pos, "invalid occurrence range %d..%d", min, max)
	}
	ec, ok := t.Content.(*xsd.ElementContent)
	if !ok {
		// spec: cos-ct-extends.1.4.2.2 — element content can only be added to
		// a type whose content is element-only or mixed.
		return xsd.NewError(xsd.SpecCosCTExtends, t.Pos, "complex type %s has no element content to extend", t.Name)
	}
	p := &xsd.Particle{MinOccurs: min, MaxOccurs: max, Term: elem}
	if ec.Particle == nil {
		ec.Particle = p
		return nil
	}
	// Wrap the existing content and the new element in a sequence.
	ec.Particle = &xsd.Particle{
		MinOccurs: 1, MaxOccurs: 1,
		Term: &xsd.ModelGroup{
			Compositor: xsd.CompositorSequence,
			Particles:  []*xsd.Particle{ec.Particle, p},
		},
	}
	return nil
}

// Validate checks the intrinsic, self-contained invariants of a simple type
// definition — the ones that must hold for the facet engine to operate on it
// regardless of the surrounding schema. It is meant for types assembled
// programmatically (a custom primitive or restriction built and registered
// through the Go API), which do not pass through the parser's schema-level
// checks. It deliberately does NOT re-validate cross-schema constraints
// (derivation-ok, UPA, identity constraints); those belong to the parser.
//
// It catches the failure modes the open value model newly admits:
//   - a missing lexical→value mapping on an atomic type (no Parse);
//   - order/enumeration facets with no comparator to evaluate them;
//   - facets that are not applicable to the base value space
//     (cos-applicable-facets), e.g. length on a numeric primitive;
//   - structurally inconsistent list/union varieties.
func Validate(t *xsd.SimpleType) error {
	var errs xsd.ErrorList

	switch t.Variety {
	case xsd.VarietyAtomic:
		// spec: cvc-datatype-valid — an atomic type must map lexical→value.
		if resolveParse(t) == nil {
			errs.Addf(xsd.SpecDatatypeValid, t.Pos, "atomic simple type %s has no lexical mapping (no Parse in its base chain)", describe(t))
		}
	case xsd.VarietyList:
		if t.ItemType == nil {
			errs.Addf(xsd.SpecCosSTRestricts, t.Pos, "list type %s has no item type", describe(t))
		} else if t.ItemType.Variety == xsd.VarietyList {
			// spec: cos-st-restricts.2 — a list's item type may not be a list.
			errs.Addf(xsd.SpecCosSTRestricts, t.Pos, "list type %s has a list item type", describe(t))
		}
	case xsd.VarietyUnion:
		// A union needs members, unless it is an empty union that rejects every
		// value via its own Parse (the xs:error pattern).
		if len(t.DirectMembers) == 0 && resolveParse(t) == nil {
			errs.Addf(xsd.SpecDatatypeValid, t.Pos, "union type %s has no member types", describe(t))
		}
	}

	// Order/enumeration facets need a real comparator.
	f := t.EffectiveFacets()
	if f.HasEnumeration() || f.MinInclusive != nil || f.MaxInclusive != nil ||
		f.MinExclusive != nil || f.MaxExclusive != nil {
		if !hasComparator(t) {
			errs.Addf(xsd.SpecDatatypeValid, t.Pos, "simple type %s declares order/enumeration facets but has no Compare in its base chain", describe(t))
		}
	}

	// Declared facets must be applicable to the base value space.
	if base, ok := t.BaseType.(*xsd.SimpleType); ok {
		applic := base.ApplicableFacets()
		for _, fk := range declaredFacetKinds(&t.DeclaredFacets) {
			if !applic.Has(fk.kind) {
				// spec: cos-applicable-facets — XSD 1.1 Part 1 §4.1.6
				errs.Addf(xsd.SpecCosApplicableFacets, t.Pos, "facet %s is not applicable to a restriction of %s", fk.name, describe(base))
			}
		}
	}

	return errs.Err()
}

// resolveParse walks t's base chain for the inherited lexical→value mapping
// (the same resolution the facet engine uses), over the exported surface.
func resolveParse(t *xsd.SimpleType) xsd.ParseFunc {
	for st := t; st != nil; {
		if st.Parse != nil {
			return st.Parse
		}
		st, _ = st.BaseType.(*xsd.SimpleType)
	}
	return nil
}

// hasComparator reports whether a real value comparator is resolvable in t's
// base chain (as opposed to the incomparable fallback). Order and enumeration
// facets are meaningless without one.
func hasComparator(t *xsd.SimpleType) bool {
	for st := t; st != nil; {
		if st.Compare != nil {
			return true
		}
		st, _ = st.BaseType.(*xsd.SimpleType)
	}
	return false
}

// describe names t for an error message, mirroring the core helper.
func describe(t *xsd.SimpleType) string {
	if !t.Name.IsZero() {
		return t.Name.String()
	}
	return "anonymous type"
}

// declaredFacetKinds returns the applicability category and name of every
// constraining facet present in f, for the cos-applicable-facets check.
func declaredFacetKinds(f *xsd.Facets) []struct {
	kind xsd.FacetSet
	name string
} {
	var out []struct {
		kind xsd.FacetSet
		name string
	}
	add := func(present bool, kind xsd.FacetSet, name string) {
		if present {
			out = append(out, struct {
				kind xsd.FacetSet
				name string
			}{kind, name})
		}
	}
	add(f.Length != nil, xsd.FacetLength, "length")
	add(f.MinLength != nil, xsd.FacetMinLength, "minLength")
	add(f.MaxLength != nil, xsd.FacetMaxLength, "maxLength")
	add(len(f.PatternGroups) > 0, xsd.FacetPattern, "pattern")
	add(f.HasEnumeration(), xsd.FacetEnumeration, "enumeration")
	add(f.WhiteSpace != xsd.WSUnset, xsd.FacetWhiteSpace, "whiteSpace")
	add(f.MinInclusive != nil, xsd.FacetBounds, "minInclusive")
	add(f.MaxInclusive != nil, xsd.FacetBounds, "maxInclusive")
	add(f.MinExclusive != nil, xsd.FacetBounds, "minExclusive")
	add(f.MaxExclusive != nil, xsd.FacetBounds, "maxExclusive")
	add(f.TotalDigits != nil, xsd.FacetTotalDigits, "totalDigits")
	add(f.FractionDigits != nil, xsd.FacetFractionDigits, "fractionDigits")
	add(len(f.Assertions) > 0, xsd.FacetAssertion, "assertion")
	add(f.ExplicitTimezone != xsd.ETZUnset, xsd.FacetExplicitTimezone, "explicitTimezone")
	return out
}
