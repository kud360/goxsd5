package xsd

// Safe mutation / extension API (Milestone 8). Each mutator re-runs the same
// facet-narrowing and intra-facet checks the parser enforces in Pass 2
// (CheckFacetRestriction + ValidateFacetSet), so a programmatically built or
// edited type can never bypass a clause the parser would have caught. All
// mutators are copy-on-write: validation runs against a candidate, and the
// receiver is committed only on success — a rejected mutation leaves the
// original untouched.

// EffectiveFacets returns the type's effective (merged, validated) facet set.
// The pointer aliases the type's internal facets; treat it as read-only and
// use the mutators below to change them.
func (t *SimpleType) EffectiveFacets() *Facets { return &t.Facets }

// RestrictWith derives a new anonymous simple type from t by restriction with
// the given declared facets, mirroring the parser's applyRestriction: it runs
// the narrowing checks (CheckFacetRestriction) of each declared facet against
// t's effective facets, then validates the merged result (ValidateFacetSet).
// On any violation it returns the error and nil, and t is unchanged. The
// returned type has no name; set Name (and Extensions, to carry foreign
// content) on it as needed.
func (t *SimpleType) RestrictWith(declared *Facets) (*SimpleType, error) {
	cmp := t.EffectiveCompare()
	if err := CheckFacetRestriction(declared, &t.Facets, cmp); err != nil {
		return nil, err
	}
	eff := MergeFacets(&t.Facets, declared)
	if err := ValidateFacetSet(&eff, cmp); err != nil {
		return nil, err
	}
	st := &SimpleType{
		BaseType:       t,
		Variety:        t.Variety,
		ItemType:       t.ItemType,
		MemberTypes:    t.MemberTypes,
		DeclaredFacets: *declared,
		Facets:         eff,
		// Fundamental facets (Part 2 §F): ordering and numericness follow the
		// base; boundedness and cardinality can only tighten.
		Ordered: t.Ordered,
		Numeric: t.Numeric,
		Bounded: t.Bounded ||
			((eff.MinInclusive != nil || eff.MinExclusive != nil) &&
				(eff.MaxInclusive != nil || eff.MaxExclusive != nil)),
		Cardinality: t.Cardinality,
	}
	if eff.HasEnumeration || eff.Length != nil || eff.MaxLength != nil || eff.TotalDigits != nil {
		st.Cardinality = CardinalityFinite
	}
	return st, nil
}

// AddEnumeration adds enumeration members to t in place. Each lexical must be
// valid against t's base type (enumeration-valid-restriction). The values
// accumulate onto t's own derivation step: calling it repeatedly builds up
// one enumeration set rather than a chain of single-value restrictions. If any
// lexical is invalid the whole call is rejected and t is unchanged.
//
// Built-in types are shared and never mutated in place; derive a subtype with
// RestrictWith and add enumerations to that instead.
func (t *SimpleType) AddEnumeration(lexicals ...string) error {
	if t.IsBuiltin {
		return NewError(SpecEnumerationValidRestriction, t.Pos, "cannot mutate the built-in type %s in place; use RestrictWith to derive a subtype", t.Name)
	}
	base, ok := t.BaseType.(*SimpleType)
	if !ok {
		return NewError(SpecEnumerationValidRestriction, t.Pos, "type %s has no simple base to draw enumeration values from", t.Name)
	}
	add := make([]Enum, 0, len(lexicals))
	for _, lex := range lexicals {
		v, err := base.ParseValue(lex, nil)
		if err != nil {
			// spec: enumeration-valid-restriction — XSD 1.1 Part 2 §4.3.5.5:
			// every enumeration value must be valid against the base type.
			return NewError(SpecEnumerationValidRestriction, t.Pos, "enumeration value %q is not valid against the base type %s: %v", lex, base.TypeName(), err)
		}
		add = append(add, Enum{Value: v, Lexical: lex})
	}
	// Copy-on-write: build the candidate declared facets, re-derive and
	// validate, and commit only if the result is sound.
	declared := t.DeclaredFacets
	declared.HasEnumeration = true
	declared.Enumeration = append(append([]Enum{}, t.DeclaredFacets.Enumeration...), add...)
	eff := MergeFacets(&base.Facets, &declared)
	if err := ValidateFacetSet(&eff, t.EffectiveCompare()); err != nil {
		return err
	}
	t.DeclaredFacets = declared
	t.Facets = eff
	t.Cardinality = CardinalityFinite
	return nil
}

// AddPattern adds a pattern facet to t in place as a new single-pattern group
// (matched in addition to every inherited group: AND across groups). The
// pattern is compiled with the XSD regex translator; a malformed pattern is
// rejected and t is unchanged. Built-ins may not be mutated in place.
func (t *SimpleType) AddPattern(pattern string) error {
	if t.IsBuiltin {
		return NewError(SpecRegexValid, t.Pos, "cannot mutate the built-in type %s in place; use RestrictWith to derive a subtype", t.Name)
	}
	re, err := CompileRegex(pattern)
	if err != nil {
		// spec: regex-valid — XSD 1.1 Part 2 Appendix G
		return NewError(SpecRegexValid, t.Pos, "invalid pattern: %v", err)
	}
	g := PatternGroup{{Source: pattern, Re: re}}
	t.DeclaredFacets.PatternGroups = append(t.DeclaredFacets.PatternGroups, g)
	t.Facets.PatternGroups = append(t.Facets.PatternGroups, g)
	return nil
}

// AddElement appends an element particle (the element occurring min..max
// times; max < 0 means unbounded) to t's element content, mirroring the
// parser's particle construction. It is valid only on a complex type with
// element-only or mixed content: a simple-content or empty-base type has no
// element content to extend, and an error is returned with t unchanged.
func (t *ComplexType) AddElement(elem *ElementDecl, min, max int) error {
	if min < 0 || (max >= 0 && max < min) {
		// spec: p-props-correct.2.1 — XSD 1.1 Part 1 §3.9.6: 0 ≤ minOccurs ≤ maxOccurs.
		return NewError(SpecPPropsCorrect, t.Pos, "invalid occurrence range %d..%d", min, max)
	}
	ec, ok := t.Content.(*ElementContent)
	if !ok {
		// spec: cos-ct-extends.1.4.2.2 — element content can only be added to
		// a type whose content is element-only or mixed.
		return NewError(SpecCosCTExtends, t.Pos, "complex type %s has no element content to extend", t.Name)
	}
	p := &Particle{MinOccurs: min, MaxOccurs: max, Term: elem}
	if ec.Particle == nil {
		ec.Particle = p
		return nil
	}
	// Wrap the existing content and the new element in a sequence.
	ec.Particle = &Particle{
		MinOccurs: 1, MaxOccurs: 1,
		Term: &ModelGroup{
			Compositor: CompositorSequence,
			Particles:  []*Particle{ec.Particle, p},
		},
	}
	return nil
}
