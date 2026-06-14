package parser

// Unified particle-restriction relation (Track B).
//
// §3.9.6 Particle Valid (Restriction) describes ONE recursive subsumption
// relation (NSRecurseCheckCardinality / NSSubset / Recurse / RecurseUnordered /
// RecurseLax / MapAndSum). The original implementation (restrict.go) grew as a
// set of independent necessary-condition fragments, each a slice of that
// relation. This file collapses them into the single recursion.
//
// The migration is a strangler: checkParticleRestrict (the production entry)
// runs the legacy fragments and emits their findings, while
// particleRestrictUnified computes the same findings through the unified
// recursion. A test hook (restrictDiff) compares the two across the whole
// conformance corpus, and a fuzzer compares them on random particle pairs, so
// any divergence surfaces automatically. The unified relation natively decides
// the cases it has ported and delegates the remainder to the legacy fragments;
// each ported case shrinks the delegated set until the fragments are dead and
// deleted, at which point checkParticleRestrict emits the unified findings
// directly.

import (
	"github.com/kud360/goxsd5/xsd"
)

// restrictViolation is one particle-restriction finding captured as a value, so
// the analysis can be run as a pure function and its output compared (legacy vs
// unified) without touching builder error state.
type restrictViolation struct {
	ref xsd.SpecRef
	pos xsd.Pos
	msg string
}

// restrictDiff, when non-nil, is invoked by checkParticleRestrict for every
// complexContent restriction it analyses, with the violations the legacy
// fragment relation and the unified relation each produce. It exists only so the
// Track-B differential test can assert the two relations agree across the whole
// conformance corpus; production builds leave it nil.
var restrictDiff func(ct *xsd.ComplexType, legacy, unified []restrictViolation)

// checkParticleRestrict reports cos-particle-restrict violations for ct. It is
// the production entry point during the Track-B migration: it emits the legacy
// fragment findings (still authoritative) and, when the differential hook is
// installed, hands both the legacy and unified findings to it for comparison.
func (b *builder) checkParticleRestrict(ct *xsd.ComplexType, accepted func(*xsd.ElementDecl) map[xsd.QName]bool, globalsByName map[xsd.QName]*xsd.ElementDecl) {
	legacy := b.collectLegacyRestrict(ct, accepted, globalsByName)
	if restrictDiff != nil {
		restrictDiff(ct, legacy, b.particleRestrictUnified(ct, accepted, globalsByName))
	}
	for _, v := range legacy {
		b.errs.Addf(v.ref, v.pos, "%s", v.msg)
	}
}

// collectLegacyRestrict runs the legacy fragment analysis against a scratch
// error list and returns its findings as values, leaving the builder's real
// error list untouched.
func (b *builder) collectLegacyRestrict(ct *xsd.ComplexType, accepted func(*xsd.ElementDecl) map[xsd.QName]bool, globalsByName map[xsd.QName]*xsd.ElementDecl) []restrictViolation {
	saved := b.errs
	scratch := &xsd.ErrorList{}
	b.errs = scratch
	b.checkParticleRestrictLegacy(ct, accepted, globalsByName)
	b.errs = saved
	return toViolations(scratch)
}

// toViolations flattens a scratch error list into restrictViolation values.
func toViolations(errs *xsd.ErrorList) []restrictViolation {
	var vs []restrictViolation
	for _, e := range xsd.AllErrors(errs.Err()) {
		if xe, ok := e.(*xsd.Error); ok {
			vs = append(vs, restrictViolation{ref: xe.Ref, pos: xe.Pos, msg: xe.Msg})
		}
	}
	return vs
}

// rreport collects restrictViolation values during a unified-relation run,
// mirroring the builder.errf surface the legacy fragments use so ported code
// reads the same.
type rreport struct {
	vs []restrictViolation
}

func (r *rreport) errf(ref xsd.SpecRef, pos xsd.Pos, format string, args ...any) {
	r.vs = append(r.vs, restrictViolation{ref: ref, pos: pos, msg: xsd.NewError(ref, pos, format, args...).Msg})
}

// particleRestrictUnified decides cos-particle-restrict for ct through the
// unified §3.9.6 recursion. During the migration it natively handles the cases
// it has ported and delegates the rest to the legacy fragments, so its output
// matches checkParticleRestrictLegacy exactly until each case is ported.
func (b *builder) particleRestrictUnified(ct *xsd.ComplexType, accepted func(*xsd.ElementDecl) map[xsd.QName]bool, globalsByName map[xsd.QName]*xsd.ElementDecl) []restrictViolation {
	if ct.DerivationMethod != xsd.DeriveRestriction {
		return nil
	}
	bct, ok := ct.BaseType.(*xsd.ComplexType)
	if !ok {
		return nil
	}
	rec, rok := ct.Content.(*xsd.ElementContent)
	bec, bok := bct.Content.(*xsd.ElementContent)
	if !rok || !bok || rec.Particle == nil || bec.Particle == nil {
		return nil
	}
	if rec.OpenContent != nil || bec.OpenContent != nil {
		return nil
	}
	bParts, bTop, bOK := flatGroup(bec.Particle)
	if !bOK {
		return nil
	}
	// Ported slice: a base content model that is a flat all/sequence of named
	// element particles only (no wildcard), restricted by a flat run or a choice
	// of flat runs that are likewise wildcard-free. Anything involving a wildcard
	// is still decided by the legacy fragments (NSSubset / shadow / multi-wildcard
	// porting is Track B step 2).
	if particlesHaveWildcard(bParts) || !restrictionIsWildcardFree(rec) {
		return b.collectLegacyRestrict(ct, accepted, globalsByName)
	}

	slots := make([]*baseSlot, 0, len(bParts))
	for _, bp := range bParts {
		s := &baseSlot{
			min:  mulOcc(bTop.MinOccurs, bp.MinOccurs),
			max:  mulOcc(bTop.MaxOccurs, bp.MaxOccurs),
			decl: bp.Term.(*xsd.ElementDecl),
		}
		s.names = accepted(s.decl)
		slots = append(slots, s)
	}

	rep := &rreport{}
	if rParts, rTop, rOK := flatGroup(rec.Particle); rOK {
		unifiedRestrictRun(rep, ct, slots, rParts, rTop)
		return rep.vs
	}
	branches, _ := choiceBranches(rec.Particle) // restrictionIsWildcardFree guaranteed flat/choice
	unitTop := &xsd.Particle{MinOccurs: 1, MaxOccurs: 1}
	for _, br := range branches {
		unifiedRestrictRun(rep, ct, slots, br, unitTop)
	}
	return rep.vs
}

// unifiedRestrictRun is the §3.9.6 Recurse / RecurseUnordered case for a flat
// run of named element particles against a wildcard-free flat base: every
// restriction element must map (NameAndTypeOK) to a base element particle whose
// accepted names include it, with valid type derivation, no widened nillability,
// and an unchanged type table; the summed occurrence each base particle receives
// must lie within its range; and every required base particle must be retained.
func unifiedRestrictRun(rep *rreport, ct *xsd.ComplexType, slots []*baseSlot, rParts []*xsd.Particle, rTop *xsd.Particle) {
	sumMin := make([]int, len(slots))
	sumMax := make([]int, len(slots))
	mapped := make([]bool, len(slots))

	for _, rp := range rParts {
		term := rp.Term.(*xsd.ElementDecl) // run is wildcard-free
		slot, ambiguous := -1, false
		for i, s := range slots {
			if s.names[term.Name] {
				if slot != -1 {
					ambiguous = true
					break
				}
				slot = i
			}
		}
		if ambiguous {
			return // can't decide which base particle it restricts: give up
		}
		if slot == -1 {
			// spec: cos-particle-restrict — §3.4.6.4 clause 1: the base content
			// model (no wildcard) does not allow this element name.
			rep.errf(xsd.SpecCosParticleRestrict, rp.Pos, "element %s in the restriction of %s is not allowed by the base content model", term.Name, describeCT(ct))
			continue
		}
		mapped[slot] = true
		sumMin[slot] = addOcc(sumMin[slot], mulOcc(rTop.MinOccurs, rp.MinOccurs))
		sumMax[slot] = addOcc(sumMax[slot], mulOcc(rTop.MaxOccurs, rp.MaxOccurs))
		bDecl := slots[slot].decl
		if !validlyDerivedByRestriction(term.Type, bDecl.Type) {
			rep.errf(derivationFailureRef(bDecl.Type), rp.Pos, "element %s in the restriction of %s has a type that is not derived from the base element's type", term.Name, describeCT(ct))
		}
		if !bDecl.Nillable && term.Nillable {
			rep.errf(xsd.SpecCosParticleRestrict, rp.Pos, "element %s in the restriction of %s may not be nillable when the base element is not", term.Name, describeCT(ct))
		}
		if !typeTablesEqual(term, bDecl) {
			rep.errf(xsd.SpecCosParticleRestrict, rp.Pos, "element %s in the restriction of %s has a type table that differs from the base element's", term.Name, describeCT(ct))
		}
	}

	for i, s := range slots {
		if !mapped[i] {
			if s.min > 0 {
				rep.errf(xsd.SpecCosParticleRestrict, rTop.Pos, "the restriction of %s omits required base content (%s)", describeCT(ct), slotName(s.decl))
			}
			continue
		}
		if sumMin[i] < s.min {
			rep.errf(xsd.SpecCosParticleRestrict, rTop.Pos, "the restriction of %s allows %s fewer times than the base requires", describeCT(ct), slotName(s.decl))
		}
		if !occLE(sumMax[i], s.max) {
			rep.errf(xsd.SpecCosParticleRestrict, rTop.Pos, "the restriction of %s allows %s more times than the base permits", describeCT(ct), slotName(s.decl))
		}
	}
}

// particlesHaveWildcard reports whether any of the flat particles is a wildcard.
func particlesHaveWildcard(parts []*xsd.Particle) bool {
	for _, p := range parts {
		if _, ok := p.Term.(*xsd.Wildcard); ok {
			return true
		}
	}
	return false
}

// restrictionIsWildcardFree reports whether rec's content model is a flat run,
// or a choice of flat runs, that the unified native path can analyse: it must be
// shaped like the legacy analyser expects (flat all/sequence or unit-occurring
// choice of flat branches) and contain no wildcard particle anywhere.
func restrictionIsWildcardFree(rec *xsd.ElementContent) bool {
	if rParts, _, ok := flatGroup(rec.Particle); ok {
		return !particlesHaveWildcard(rParts)
	}
	branches, ok := choiceBranches(rec.Particle)
	if !ok {
		return false
	}
	for _, br := range branches {
		if particlesHaveWildcard(br) {
			return false
		}
	}
	return true
}
