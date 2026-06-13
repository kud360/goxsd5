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

	rParts, rTop, rOK := flatElementGroup(rec.Particle)
	bParts, bTop, bOK := flatElementGroup(bec.Particle)
	if !rOK || !bOK {
		return // a wildcard / choice / nested group is present: give up safely
	}

	type baseSlot struct {
		decl     *xsd.ElementDecl
		min, max int
		names    map[xsd.QName]bool
		sumMin   int
		sumMax   int
		mapped   bool
	}
	slots := make([]*baseSlot, 0, len(bParts))
	for _, bp := range bParts {
		decl := bp.Term.(*xsd.ElementDecl)
		slots = append(slots, &baseSlot{
			decl:  decl,
			min:   mulOcc(bTop.MinOccurs, bp.MinOccurs),
			max:   mulOcc(bTop.MaxOccurs, bp.MaxOccurs),
			names: accepted(decl),
		})
	}

	for _, rp := range rParts {
		decl := rp.Term.(*xsd.ElementDecl)
		// Find the base particle whose accepted names include this element
		// (an exact name match or a substitution-group membership).
		var slot *baseSlot
		ambiguous := false
		for _, s := range slots {
			if s.names[decl.Name] {
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
		if slot == nil {
			// spec: cos-particle-restrict — XSD 1.1 Part 1 §3.4.6.4 clause 1
			b.errf(xsd.SpecCosParticleRestrict, rp.Pos, "element %s in the restriction of %s is not allowed by the base content model", decl.Name, describeCT(ct))
			continue
		}
		slot.mapped = true
		slot.sumMin = addOcc(slot.sumMin, mulOcc(rTop.MinOccurs, rp.MinOccurs))
		slot.sumMax = addOcc(slot.sumMax, mulOcc(rTop.MaxOccurs, rp.MaxOccurs))

		if !validlyDerivedByRestriction(decl.Type, slot.decl.Type) {
			// spec: cos-particle-restrict — §3.4.6.4 clause 2 / NameAndTypeOK:
			// the restricting element's type must derive from the base's.
			b.errf(xsd.SpecCosParticleRestrict, rp.Pos, "element %s in the restriction of %s has a type that is not derived from the base element's type", decl.Name, describeCT(ct))
		}
		if !slot.decl.Nillable && decl.Nillable {
			b.errf(xsd.SpecCosParticleRestrict, rp.Pos, "element %s in the restriction of %s may not be nillable when the base element is not", decl.Name, describeCT(ct))
		}
	}

	for _, s := range slots {
		if !s.mapped {
			if s.min > 0 {
				// spec: cos-particle-restrict — §3.4.6.4 clause 1: a required
				// base element must remain required in the restriction.
				b.errf(xsd.SpecCosParticleRestrict, rec.Particle.Pos, "the restriction of %s omits the required base element %s", describeCT(ct), s.decl.Name)
			}
			continue
		}
		if s.sumMin < s.min {
			b.errf(xsd.SpecCosParticleRestrict, rec.Particle.Pos, "the restriction of %s allows element %s fewer times than the base requires", describeCT(ct), s.decl.Name)
		}
		if !occLE(s.sumMax, s.max) {
			b.errf(xsd.SpecCosParticleRestrict, rec.Particle.Pos, "the restriction of %s allows element %s more times than the base permits", describeCT(ct), s.decl.Name)
		}
	}
}

// flatElementGroup returns the element particles of p's term when that term is
// an all or sequence model group every one of whose particles is a plain
// element declaration. It reports ok=false (so the caller gives up) for a
// choice compositor, a group reference, or any wildcard or nested group, since
// those need the full subsumption algorithm.
func flatElementGroup(p *xsd.Particle) (parts []*xsd.Particle, top *xsd.Particle, ok bool) {
	mg, isMG := p.Term.(*xsd.ModelGroup)
	if !isMG || mg.Compositor == xsd.CompositorChoice {
		return nil, nil, false
	}
	for _, c := range mg.Particles {
		if _, isElem := c.Term.(*xsd.ElementDecl); !isElem {
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
