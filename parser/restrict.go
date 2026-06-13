package parser

// Particle restriction (cos-particle-restrict / Content type restricts,
// §3.4.6.4 and §3.9.6) for complexContent restrictions.
//
// XSD 1.1 defines a valid restriction semantically: every sequence of element
// information items locally valid against the restriction R must also be valid
// against the base B, and the types R assigns must validly derive from those B
// assigns. Implementing the full relation (language inclusion over particle
// automata with occurrence counting, wildcards, choice and substitution
// groups) is the hardest algorithm in the specification, and §3.9.6 explicitly
// permits a processor to provisionally accept an <all>-group derivation it
// cannot decide by examining the schema alone.
//
// This pass implements the part that is both tractable and free of false
// positives: the per-name occurrence/type "bag" check for restrictions whose
// content model and whose base's content model are a flat all or sequence of
// element particles. For such models the count of each element name an
// instance may carry ranges exactly over [Σ minOccurs, Σ maxOccurs] of the
// particles producing it, so a necessary condition for L(R) ⊆ L(B) is:
//
//   - every element name R can produce is producible by B (directly or, when
//     B uses a substitution-group head, by a member of that group);
//   - the summed occurrence range R assigns each base particle lies within
//     that base particle's range;
//   - every required (minOccurs > 0) base particle is produced by R;
//   - the type R assigns each element validly derives from the base's;
//   - nillability is not widened.
//
// Violating any of these means R is definitely not a valid restriction, so it
// is safe to report. When the model contains a wildcard, choice, nested group,
// group reference, open content, or a name that is ambiguously substitutable,
// the analysis gives up (reports nothing): those cases stay tolerated rather
// than risk a false positive. They are the wildcard-subset / choice cases
// listed in NOTES.md.

import "github.com/kud360/goxsd5/xsd"

// checkParticleRestrict reports cos-particle-restrict violations for ct when ct
// is a complexContent restriction of a complex base and both content models are
// a flat all/sequence of element particles.
func (b *builder) checkParticleRestrict(ct *xsd.ComplexType, accepted func(*xsd.ElementDecl) map[xsd.QName]bool) {
	if ct.DerivationMethod != xsd.DeriveRestriction {
		return
	}
	bct, ok := ct.BaseType.(*xsd.ComplexType)
	if !ok {
		return
	}
	rec, rok := ct.Content.(*xsd.ElementContent)
	bec, bok := bct.Content.(*xsd.ElementContent)
	if !rok || !bok || rec.Particle == nil || bec.Particle == nil {
		return // empty-content restriction: emptiability is checked elsewhere
	}
	if rec.OpenContent != nil || bec.OpenContent != nil {
		return // open-content subset is deferred (NOTES.md)
	}

	rParts, rTop, rOK := flatGroup(rec.Particle)
	bParts, bTop, bOK := flatGroup(bec.Particle)
	if !rOK || !bOK {
		return // a choice / nested group / group ref is present: give up safely
	}

	// Build a slot per base particle: element slots carry the declaration's
	// accepted-name set; the wildcard slot (at most one is handled) carries the
	// wildcard. With two or more base wildcards a restriction particle could map
	// to either, so give up rather than guess.
	type baseSlot struct {
		decl     *xsd.ElementDecl // element slot
		wc       *xsd.Wildcard    // wildcard slot
		min, max int
		names    map[xsd.QName]bool
		sumMin   int
		sumMax   int
		mapped   bool
	}
	slots := make([]*baseSlot, 0, len(bParts))
	var baseWC *baseSlot
	for _, bp := range bParts {
		s := &baseSlot{min: mulOcc(bTop.MinOccurs, bp.MinOccurs), max: mulOcc(bTop.MaxOccurs, bp.MaxOccurs)}
		switch term := bp.Term.(type) {
		case *xsd.ElementDecl:
			s.decl = term
			s.names = accepted(term)
		case *xsd.Wildcard:
			if baseWC != nil {
				return // more than one base wildcard: give up
			}
			s.wc = term
			baseWC = s
		}
		slots = append(slots, s)
	}

	// mapTo accumulates rp's occurrence onto slot s and marks it mapped.
	mapTo := func(s *baseSlot, rp *xsd.Particle) {
		s.mapped = true
		s.sumMin = addOcc(s.sumMin, mulOcc(rTop.MinOccurs, rp.MinOccurs))
		s.sumMax = addOcc(s.sumMax, mulOcc(rTop.MaxOccurs, rp.MaxOccurs))
	}

	for _, rp := range rParts {
		switch term := rp.Term.(type) {
		case *xsd.ElementDecl:
			// Find the base element particle whose accepted names include this
			// element (exact name or substitution-group membership).
			var slot *baseSlot
			ambiguous := false
			for _, s := range slots {
				if s.decl != nil && s.names[term.Name] {
					if slot != nil {
						ambiguous = true
						break
					}
					slot = s
				}
			}
			if ambiguous {
				return // can't decide which base particle it restricts: give up
			}
			if slot != nil {
				mapTo(slot, rp)
				if !validlyDerivedByRestriction(term.Type, slot.decl.Type) {
					// spec: cos-particle-restrict — §3.4.6.4 clause 2 /
					// NameAndTypeOK: the restricting element's type must derive
					// from the base's.
					b.errf(xsd.SpecCosParticleRestrict, rp.Pos, "element %s in the restriction of %s has a type that is not derived from the base element's type", term.Name, describeCT(ct))
				}
				if !slot.decl.Nillable && term.Nillable {
					b.errf(xsd.SpecCosParticleRestrict, rp.Pos, "element %s in the restriction of %s may not be nillable when the base element is not", term.Name, describeCT(ct))
				}
				continue
			}
			// No base element matches: the element must be accepted by the base
			// wildcard (NSCompat), else the restriction introduces a name the
			// base disallows.
			if baseWC != nil && wildcardAllowsName(baseWC.wc, term.Name) {
				mapTo(baseWC, rp)
				continue
			}
			// spec: cos-particle-restrict — XSD 1.1 Part 1 §3.4.6.4 clause 1
			b.errf(xsd.SpecCosParticleRestrict, rp.Pos, "element %s in the restriction of %s is not allowed by the base content model", term.Name, describeCT(ct))
		case *xsd.Wildcard:
			// A restriction wildcard can only restrict the base wildcard, and
			// only if it is a wildcard subset of it (NSSubset).
			if baseWC == nil {
				// spec: cos-particle-restrict — §3.4.6.4 clause 1: the base has no
				// wildcard, so the restriction may not introduce one.
				b.errf(xsd.SpecCosParticleRestrict, rp.Pos, "the restriction of %s introduces a wildcard the base content model does not allow", describeCT(ct))
				continue
			}
			if !namespaceConstraintSubset(term, baseWC.wc) {
				// spec: cos-particle-restrict — §3.4.6.4 clause 1 / NSSubset:
				// the restricting wildcard must be a subset of the base's.
				b.errf(xsd.SpecCosParticleRestrict, rp.Pos, "a wildcard in the restriction of %s allows elements the base wildcard does not", describeCT(ct))
			}
			mapTo(baseWC, rp)
		}
	}

	for _, s := range slots {
		if !s.mapped {
			if s.min > 0 {
				// spec: cos-particle-restrict — §3.4.6.4 clause 1: required base
				// content must remain required in the restriction.
				b.errf(xsd.SpecCosParticleRestrict, rec.Particle.Pos, "the restriction of %s omits required base content (%s)", describeCT(ct), slotName(s.decl))
			}
			continue
		}
		if s.sumMin < s.min {
			b.errf(xsd.SpecCosParticleRestrict, rec.Particle.Pos, "the restriction of %s allows %s fewer times than the base requires", describeCT(ct), slotName(s.decl))
		}
		if !occLE(s.sumMax, s.max) {
			b.errf(xsd.SpecCosParticleRestrict, rec.Particle.Pos, "the restriction of %s allows %s more times than the base permits", describeCT(ct), slotName(s.decl))
		}
	}
}

// checkOpenContentRestrict reports derivation-ok-restriction violations for the
// {open content} of a complexContent restriction (cos-ct-restricts /
// Derivation Valid (Restriction, Complex), §3.4.6.4 clause 9). A restriction
// that closes the content (no open content, or mode none) is always valid, so
// only a restriction that keeps open content is checked: its mode may be no
// more open than the base's (interleave is more open than suffix), its wildcard
// must be a namespace subset of the base's, and its processContents must be
// identical to or stronger than the base's. Each is a necessary condition for
// L(R) ⊆ L(B), so a violation is always a real error.
func (b *builder) checkOpenContentRestrict(ct *xsd.ComplexType) {
	if ct.DerivationMethod != xsd.DeriveRestriction {
		return
	}
	bct, ok := ct.BaseType.(*xsd.ComplexType)
	if !ok || bct.BaseType == nil {
		return // base is xs:anyType: this is the implicit restriction, not a
		// user complexContent restriction, and open content may be added freely.
	}
	rec, rok := ct.Content.(*xsd.ElementContent)
	if !rok {
		return
	}
	roc := effectiveOpenContent(rec)
	if roc == nil {
		return // closing the content is always a valid restriction
	}
	bec, _ := bct.Content.(*xsd.ElementContent)
	boc := effectiveOpenContent(bec)
	if boc == nil {
		// spec: derivation-ok-restriction — §3.4.6.4 clause 9.1: the base has no
		// open content, so the restriction may not introduce any. Skip when the
		// base's own content model contains a wildcard, which can absorb the
		// open-content elements directly so that the restriction is still valid
		// (open022); deciding that exactly needs the full language inclusion, so
		// give up rather than risk a false positive.
		var bp *xsd.Particle
		if bec != nil {
			bp = bec.Particle
		}
		if !particleHasWildcard(bp) {
			b.errf(xsd.SpecDerivationOKRestriction, roc.Pos, "the restriction of %s introduces open content the base type does not allow", describeCT(ct))
		}
		return
	}
	if openContentOpenness(roc.Mode) > openContentOpenness(boc.Mode) && particleMatchesNonEmpty(rec.Particle) {
		// spec: derivation-ok-restriction — §3.4.6.4 clause 9.2: the restriction's
		// open content mode may be no more permissive than the base's (interleave
		// is more open than suffix). The mode only matters when the restriction's
		// own content model can produce an element to interleave around; with an
		// empty content model interleave and suffix coincide (open020).
		b.errf(xsd.SpecDerivationOKRestriction, roc.Pos, "the open content of the restriction of %s is more permissive (interleave) than the base's (suffix)", describeCT(ct))
	}
	if roc.Wildcard != nil && boc.Wildcard != nil {
		if !namespaceConstraintSubset(roc.Wildcard, boc.Wildcard) {
			// spec: derivation-ok-restriction — §3.4.6.4 clause 9.3 / Wildcard
			// Subset: the restriction's open content wildcard must be a subset of
			// the base's.
			b.errf(xsd.SpecDerivationOKRestriction, roc.Pos, "the open content of the restriction of %s allows elements the base's open content does not", describeCT(ct))
		}
		if !processContentsAtLeastAsStrict(roc.Wildcard, boc.Wildcard) {
			// spec: derivation-ok-restriction — §3.4.6.4 clause 9.3: the open
			// content wildcard's processContents must be identical to or stronger
			// than the base's (strict > lax > skip).
			b.errf(xsd.SpecDerivationOKRestriction, roc.Pos, "the open content of the restriction of %s has weaker processContents than the base's", describeCT(ct))
		}
	}
}

// effectiveOpenContent returns ec's open content, treating an absent one or one
// with {mode} none as no open content.
func effectiveOpenContent(ec *xsd.ElementContent) *xsd.OpenContent {
	if ec == nil || ec.OpenContent == nil || ec.OpenContent.Mode == xsd.OpenContentNone {
		return nil
	}
	return ec.OpenContent
}

// particleMatchesNonEmpty reports whether p can match at least one element
// information item: it has a positive maximum occurrence and contains an
// element or wildcard term (directly or nested) that can itself occur. A nil
// particle, or a model group all of whose particles are unreachable, matches
// only the empty sequence.
func particleMatchesNonEmpty(p *xsd.Particle) bool {
	if p == nil || p.MaxOccurs == 0 {
		return false
	}
	switch term := p.Term.(type) {
	case *xsd.ElementDecl, *xsd.Wildcard:
		return true
	case *xsd.ModelGroup:
		for _, c := range term.Particles {
			if particleMatchesNonEmpty(c) {
				return true
			}
		}
	}
	return false
}

// particleHasWildcard reports whether p contains a wildcard term anywhere in
// its model-group tree.
func particleHasWildcard(p *xsd.Particle) bool {
	if p == nil {
		return false
	}
	switch term := p.Term.(type) {
	case *xsd.Wildcard:
		return true
	case *xsd.ModelGroup:
		for _, c := range term.Particles {
			if particleHasWildcard(c) {
				return true
			}
		}
	}
	return false
}

// openContentOpenness ranks open-content modes by permissiveness: interleave
// (matches anywhere) is more open than suffix (only after the content), which
// is more open than none.
func openContentOpenness(m xsd.OpenContentMode) int {
	switch m {
	case xsd.OpenContentInterleave:
		return 2
	case xsd.OpenContentSuffix:
		return 1
	default:
		return 0
	}
}

// processContentsAtLeastAsStrict reports whether wildcard sub's processContents
// is identical to or stronger than super's (strict > lax > skip), per the
// Particle Derivation OK process-contents condition. When super is skip there is
// no constraint.
func processContentsAtLeastAsStrict(sub, super *xsd.Wildcard) bool {
	if super.ProcessContents == xsd.ProcessSkip {
		return true
	}
	return sub.ProcessContents <= super.ProcessContents
}

// slotName names a base slot for diagnostics: the element name, or "a wildcard"
// for the wildcard slot.
func slotName(decl *xsd.ElementDecl) string {
	if decl == nil {
		return "a wildcard"
	}
	return "element " + decl.Name.String()
}

// checkAttrWildcardRestriction enforces that a complexContent/simpleContent
// restriction's attribute wildcard admits only attributes the base's does
// (derivation-ok-restriction, §3.4.6.3: every attribute valid against the
// restriction must be valid against the base). r is the restriction's declared
// attribute wildcard (non-nil); base is the base type's, which may be absent.
func (b *builder) checkAttrWildcardRestriction(ct *xsd.ComplexType, r, base *xsd.Wildcard) {
	if base == nil {
		// spec: derivation-ok-restriction — §3.4.6.3: the base allows no
		// wildcard attributes, so the restriction may not introduce one.
		b.errf(xsd.SpecDerivationOKRestriction, r.Pos, "the attribute wildcard on the restriction of %s is not allowed: the base type has no attribute wildcard", describeCT(ct))
		return
	}
	if !namespaceConstraintSubset(r, base) {
		// spec: derivation-ok-restriction — §3.4.6.3 / Wildcard Subset
		// (cos-ns-subset): the restricting wildcard must be a subset of the
		// base's.
		b.errf(xsd.SpecDerivationOKRestriction, r.Pos, "the attribute wildcard on the restriction of %s allows attributes the base type's wildcard does not", describeCT(ct))
	}
}

// flatGroup returns the particles of p's term when that term is an all or
// sequence model group every one of whose particles is a plain element
// declaration or a wildcard. It reports ok=false (so the caller gives up) for a
// choice compositor, a group reference, or any nested model group, since those
// need the full subsumption algorithm.
func flatGroup(p *xsd.Particle) (parts []*xsd.Particle, top *xsd.Particle, ok bool) {
	mg, isMG := p.Term.(*xsd.ModelGroup)
	if !isMG || mg.Compositor == xsd.CompositorChoice {
		return nil, nil, false
	}
	for _, c := range mg.Particles {
		switch c.Term.(type) {
		case *xsd.ElementDecl, *xsd.Wildcard:
		default:
			return nil, nil, false
		}
	}
	return mg.Particles, p, true
}

// validlyDerivedByRestriction reports whether the restriction element's type
// validly derives from the base element's type. It is lenient: an
// undeterminable type (nil) or any derivation chain is accepted; only
// definitely-unrelated types fail, so no valid restriction is ever rejected
// here.
//
// Union membership (cos-st-derived-ok clause 2.2.4: a type derived from a
// member of a union is validly derived from the union) is subtle — and the
// member-union flattening done at build time discards intervening unions — so
// any union involvement is accepted rather than risk a false positive. The
// union-substitutability cases (saxon simple01x) therefore stay tolerated; see
// NOTES.md.
func validlyDerivedByRestriction(rType, bType xsd.Type) bool {
	if rType == nil || bType == nil || rType == bType {
		return true
	}
	if involvesUnion(rType) || involvesUnion(bType) {
		return true
	}
	_, derived := derivationMethods(rType, bType)
	return derived
}

// involvesUnion reports whether t or any type in its simple-type base chain has
// union variety.
func involvesUnion(t xsd.Type) bool {
	st, ok := t.(*xsd.SimpleType)
	if !ok {
		return false
	}
	for cur := st; cur != nil; {
		if cur.Variety == xsd.VarietyUnion {
			return true
		}
		base, ok := cur.BaseType.(*xsd.SimpleType)
		if !ok || base == cur {
			break
		}
		cur = base
	}
	return false
}

// mulOcc multiplies two occurrence counts, propagating unbounded (n·0 = 0).
func mulOcc(a, b int) int {
	if a == 0 || b == 0 {
		return 0
	}
	if a == xsd.UnboundedOccurs || b == xsd.UnboundedOccurs {
		return xsd.UnboundedOccurs
	}
	return a * b
}

// addOcc adds two occurrence counts, propagating unbounded.
func addOcc(a, b int) int {
	if a == xsd.UnboundedOccurs || b == xsd.UnboundedOccurs {
		return xsd.UnboundedOccurs
	}
	return a + b
}

// occLE reports whether maxOccurs a is no greater than maxOccurs b, with
// UnboundedOccurs treated as larger than every finite count.
func occLE(a, b int) bool {
	if b == xsd.UnboundedOccurs {
		return true
	}
	if a == xsd.UnboundedOccurs {
		return false
	}
	return a <= b
}
