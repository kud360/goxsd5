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
	"github.com/kud360/goxsd5/builtin"
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

// particleRestrictUnified decides cos-particle-restrict for ct through one
// region/representative-name relation that replaces the legacy fragments. The
// base content model must be a flat all/sequence of element/wildcard particles
// and the restriction a flat run or a flat choice; any other shape is given up
// (returns no findings) exactly as the fragments did, since deciding it needs
// the full language-inclusion algorithm §3.9.6 permits a processor to skip.
//
// Every case shares the same NameAndTypeOK (type derivation / nillability /
// type table) and xs:all wildcard-shadowing obligations; they differ only in the
// §3.9.6 cardinality rule, selected by how many wildcards the base carries:
//
//   - zero or one base wildcard: NSRecurseCheckCardinality / Recurse — each
//     restriction particle maps to the one base particle it restricts (a named
//     element by name/substitution, otherwise the lone wildcard) and the summed
//     occurrence each base particle receives must lie in its range (slotRun).
//   - two or more base wildcards: a restriction particle may be coverable by
//     several base wildcards or straddle them, so a one-to-one map is
//     impossible; the packing is decided over disjoint name regions instead,
//     giving up the count when the regions overlap (regionRun).
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
		return nil // empty-content restriction: emptiability is checked elsewhere
	}
	if rec.OpenContent != nil || bec.OpenContent != nil {
		return nil // open-content subset is handled by checkOpenContentRestrict
	}
	bParts, bTop, bOK := flatGroup(bec.Particle)
	if !bOK {
		return nil // base is not a flat all/sequence: give up safely
	}
	nBaseWC := 0
	for _, bp := range bParts {
		if _, ok := bp.Term.(*xsd.Wildcard); ok {
			nBaseWC++
		}
	}
	bmg, _ := bec.Particle.Term.(*xsd.ModelGroup)
	baseIsAll := bmg != nil && bmg.Compositor == xsd.CompositorAll

	rep := &rreport{}
	if nBaseWC >= 2 {
		// wildcardAllowsName ignores the dynamic ##defined/##definedSibling
		// sentinels, so a base wildcard carrying one would be treated as accepting
		// more names than it really does and a witness built on it could be unsound.
		for _, bp := range bParts {
			if w, ok := bp.Term.(*xsd.Wildcard); ok && wildcardHasSentinel(w) {
				return nil
			}
		}
		regions := buildBaseRegions(bParts, bTop, accepted)
		forEachRestrictionRun(rec, func(rParts []*xsd.Particle, rTop *xsd.Particle) {
			b.regionRun(rep, ct, regions, bParts, rParts, rTop)
		})
		return rep.vs
	}

	slots, baseWC, baseNamed := buildBaseSlots(bParts, bTop, accepted)
	forEachRestrictionRun(rec, func(rParts []*xsd.Particle, rTop *xsd.Particle) {
		b.slotRun(rep, ct, slots, baseWC, rParts, rTop)
		if baseIsAll {
			b.unifiedShadow(rep, ct, baseNamed, rParts, accepted, globalsByName)
		}
	})
	return rep.vs
}

// forEachRestrictionRun invokes fn for each flat run of a restriction content
// model: once for a flat all/sequence, once per branch of a unit-occurring
// choice of flat branches (an instance matches exactly one branch, so each must
// independently restrict the base). It does nothing for any other shape, which
// is the give-up case.
func forEachRestrictionRun(rec *xsd.ElementContent, fn func(rParts []*xsd.Particle, rTop *xsd.Particle)) {
	if rParts, rTop, ok := flatGroup(rec.Particle); ok {
		fn(rParts, rTop)
		return
	}
	branches, ok := choiceBranches(rec.Particle)
	if !ok {
		return
	}
	unitTop := &xsd.Particle{MinOccurs: 1, MaxOccurs: 1, Pos: rec.Particle.Pos}
	for _, br := range branches {
		fn(br, unitTop)
	}
}

// buildBaseSlots builds one slot per base particle for the zero/one-wildcard
// case, returning the slots, the index of the lone wildcard slot (or -1), and
// the base's named declarations (for the shadow check).
func buildBaseSlots(bParts []*xsd.Particle, bTop *xsd.Particle, accepted func(*xsd.ElementDecl) map[xsd.QName]bool) (slots []*baseSlot, baseWC int, baseNamed []*xsd.ElementDecl) {
	baseWC = -1
	for _, bp := range bParts {
		s := &baseSlot{min: mulOcc(bTop.MinOccurs, bp.MinOccurs), max: mulOcc(bTop.MaxOccurs, bp.MaxOccurs)}
		switch term := bp.Term.(type) {
		case *xsd.ElementDecl:
			s.decl = term
			s.names = accepted(term)
			baseNamed = append(baseNamed, term)
		case *xsd.Wildcard:
			s.wc = term
			baseWC = len(slots)
		}
		slots = append(slots, s)
	}
	return slots, baseWC, baseNamed
}

// buildBaseRegions views each base particle as a name region (an element's
// accepted names, or a wildcard's namespace constraint) carrying its effective
// occurrence range, for the multi-wildcard packing analysis.
func buildBaseRegions(bParts []*xsd.Particle, bTop *xsd.Particle, accepted func(*xsd.ElementDecl) map[xsd.QName]bool) []baseRegion {
	regions := make([]baseRegion, 0, len(bParts))
	for _, bp := range bParts {
		r := baseRegion{min: mulOcc(bTop.MinOccurs, bp.MinOccurs), max: mulOcc(bTop.MaxOccurs, bp.MaxOccurs)}
		switch term := bp.Term.(type) {
		case *xsd.ElementDecl:
			names := accepted(term)
			r.decl = term
			r.name = "element " + term.Name.String()
			r.accepts = func(q xsd.QName) bool { return names[q] }
		case *xsd.Wildcard:
			w := term
			r.name = "a wildcard"
			r.accepts = func(q xsd.QName) bool { return wildcardAllowsName(w, q) }
		}
		regions = append(regions, r)
	}
	return regions
}

// nameTypeOK reports the NameAndTypeOK obligations (§3.4.6.4 clause 2) a
// restriction element rDecl owes the base element bDecl it restricts: its type
// must validly derive by restriction, it must not widen nillability, and — when
// checkTypeTable is set (the single-particle map; the multi-wildcard packing
// does not track it) — its conditional type table must be unchanged.
func (b *builder) nameTypeOK(rep *rreport, ct *xsd.ComplexType, rDecl, bDecl *xsd.ElementDecl, pos xsd.Pos, checkTypeTable bool) {
	if !validlyDerivedByRestriction(rDecl.Type, bDecl.Type) {
		rep.errf(derivationFailureRef(bDecl.Type), pos, "element %s in the restriction of %s has a type that is not derived from the base element's type", rDecl.Name, describeCT(ct))
	}
	if !bDecl.Nillable && rDecl.Nillable {
		rep.errf(xsd.SpecCosParticleRestrict, pos, "element %s in the restriction of %s may not be nillable when the base element is not", rDecl.Name, describeCT(ct))
	}
	if checkTypeTable && !typeTablesEqual(rDecl, bDecl) {
		rep.errf(xsd.SpecCosParticleRestrict, pos, "element %s in the restriction of %s has a type table that differs from the base element's", rDecl.Name, describeCT(ct))
	}
}

// slotRun is the §3.9.6 NSRecurseCheckCardinality / Recurse case for a base with
// at most one wildcard: every restriction particle maps to the single base
// particle it restricts and the summed occurrence each base particle receives
// must lie within its range, every required base particle must be retained, and
// each named map obeys NameAndTypeOK.
func (b *builder) slotRun(rep *rreport, ct *xsd.ComplexType, slots []*baseSlot, baseWC int, rParts []*xsd.Particle, rTop *xsd.Particle) {
	sumMin := make([]int, len(slots))
	sumMax := make([]int, len(slots))
	mapped := make([]bool, len(slots))
	mapTo := func(i int, rp *xsd.Particle) {
		mapped[i] = true
		sumMin[i] = addOcc(sumMin[i], mulOcc(rTop.MinOccurs, rp.MinOccurs))
		sumMax[i] = addOcc(sumMax[i], mulOcc(rTop.MaxOccurs, rp.MaxOccurs))
	}

	for _, rp := range rParts {
		switch term := rp.Term.(type) {
		case *xsd.ElementDecl:
			slot, ambiguous := -1, false
			for i, s := range slots {
				if s.decl != nil && s.names[term.Name] {
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
			if slot != -1 {
				mapTo(slot, rp)
				b.nameTypeOK(rep, ct, term, slots[slot].decl, rp.Pos, true)
				continue
			}
			// No base element matches: the element must be accepted by the base
			// wildcard (NSCompat), else the restriction introduces a name the base
			// disallows.
			if baseWC >= 0 && wildcardAllowsName(slots[baseWC].wc, term.Name) {
				mapTo(baseWC, rp)
				continue
			}
			// spec: cos-particle-restrict — §3.4.6.4 clause 1.
			rep.errf(xsd.SpecCosParticleRestrict, rp.Pos, "element %s in the restriction of %s is not allowed by the base content model", term.Name, describeCT(ct))
		case *xsd.Wildcard:
			// A restriction wildcard can only restrict the base wildcard, and only
			// if it is a wildcard subset of it (NSSubset).
			if baseWC < 0 {
				rep.errf(xsd.SpecCosParticleRestrict, rp.Pos, "the restriction of %s introduces a wildcard the base content model does not allow", describeCT(ct))
				continue
			}
			if !namespaceConstraintSubset(term, slots[baseWC].wc) {
				rep.errf(xsd.SpecCosParticleRestrict, rp.Pos, "a wildcard in the restriction of %s allows elements the base wildcard does not", describeCT(ct))
			}
			mapTo(baseWC, rp)
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

// regionRun is the §3.9.6 NSRecurseCheckCardinality case generalised to two or
// more base wildcards, where a restriction particle cannot be mapped to a single
// base particle. It reasons over the base name regions: (1) the restriction never
// admits an element no region accepts (NSCompat); (2) when the regions are
// pairwise disjoint, each region's restriction-allowed count stays within range
// (the flow problem for overlapping regions is given up); (3) a named restriction
// element landing in a named region obeys NameAndTypeOK. Both count bounds are
// exact witnesses, so no valid restriction is rejected.
func (b *builder) regionRun(rep *rreport, ct *xsd.ComplexType, regions []baseRegion, bParts, rParts []*xsd.Particle, rTop *xsd.Particle) {
	rps := make([]restrictPart, 0, len(rParts))
	for _, rp := range rParts {
		e := restrictPart{
			rmin: mulOcc(rTop.MinOccurs, rp.MinOccurs),
			rmax: mulOcc(rTop.MaxOccurs, rp.MaxOccurs),
			pos:  rp.Pos,
		}
		switch term := rp.Term.(type) {
		case *xsd.ElementDecl:
			qn := term.Name
			e.decl = term
			e.accepts = func(q xsd.QName) bool { return q == qn }
		case *xsd.Wildcard:
			if wildcardHasSentinel(term) {
				return // a restriction sentinel would make our witnesses unsound
			}
			w := term
			e.accepts = func(q xsd.QName) bool { return wildcardAllowsName(w, q) }
		}
		rps = append(rps, e)
	}

	reps := collectReps(bParts, rParts)

	// (1) NSCompat.
	for _, e := range rps {
		if e.rmax == 0 {
			continue
		}
		for _, n := range reps {
			if e.accepts(n) && !anyRegionAccepts(regions, n) {
				rep.errf(xsd.SpecCosParticleRestrict, e.pos, "the restriction of %s allows an element (%s) that the base content model does not", describeCT(ct), n)
				break
			}
		}
	}

	// The per-region count and type checks below are sound only when the regions
	// are pairwise disjoint: then every instance element matches exactly one base
	// particle. When a base wildcard overlaps the element it sits beside (legal in
	// an xs:all) an element can be absorbed by either, and the real test is a flow
	// problem we give up on rather than risk a false positive.
	if !regionsDisjoint(regions, reps) {
		return
	}

	// (2) cardinality.
	for i := range regions {
		reg := &regions[i]
		minB, maxB := 0, 0
		for _, e := range rps {
			acceptsAny, allInRegion, intersects := false, true, false
			for _, n := range reps {
				if !e.accepts(n) {
					continue
				}
				acceptsAny = true
				if reg.accepts(n) {
					intersects = true
				} else {
					allInRegion = false
				}
			}
			if acceptsAny && allInRegion {
				minB = addOcc(minB, e.rmin)
			}
			if intersects {
				maxB = addOcc(maxB, e.rmax)
			}
		}
		if minB < reg.min {
			rep.errf(xsd.SpecCosParticleRestrict, rTop.Pos, "the restriction of %s may match %s fewer times than the base requires", describeCT(ct), reg.name)
		}
		if !occLE(maxB, reg.max) {
			rep.errf(xsd.SpecCosParticleRestrict, rTop.Pos, "the restriction of %s allows %s more times than the base permits", describeCT(ct), reg.name)
		}
	}

	// (3) NameAndTypeOK for named restriction elements landing in a named region.
	for _, e := range rps {
		if e.decl == nil {
			continue
		}
		for i := range regions {
			reg := &regions[i]
			if reg.decl == nil || reg.decl.Name != e.decl.Name {
				continue
			}
			b.nameTypeOK(rep, ct, e.decl, reg.decl, e.pos, false)
		}
	}
}

// unifiedShadow reports the cos-particle-restrict / NameAndTypeOK violation the
// plain NSCompat/cardinality checks miss when an xs:all restriction reroutes a
// dropped named element to a wildcard. baseNamed are the base's named element
// declarations. See particleRestrictUnified.
func (b *builder) unifiedShadow(rep *rreport, ct *xsd.ComplexType, baseNamed []*xsd.ElementDecl, rParts []*xsd.Particle, accepted func(*xsd.ElementDecl) map[xsd.QName]bool, globalsByName map[xsd.QName]*xsd.ElementDecl) {
	// Names the run still routes to a named particle (each element's own name plus
	// every substitutable member): these keep the named-vs-named type check and
	// are never re-routed. Over-counting here only suppresses reports.
	rNamed := map[xsd.QName]bool{}
	var rWild []*xsd.Particle
	for _, rp := range rParts {
		switch term := rp.Term.(type) {
		case *xsd.ElementDecl:
			for n := range accepted(term) {
				rNamed[n] = true
			}
		case *xsd.Wildcard:
			rWild = append(rWild, rp)
		}
	}
	if len(rWild) == 0 {
		return
	}
	for _, bDecl := range baseNamed {
		n := bDecl.Name
		if rNamed[n] {
			continue // the run keeps a named particle for N: the type check covers it
		}
		for _, rp := range rWild {
			w := rp.Term.(*xsd.Wildcard)
			if !wildcardAllowsName(w, n) {
				continue
			}
			t, unconstrained, hasInstance := b.wildcardBoundType(w, n, globalsByName)
			switch {
			case !hasInstance:
				// A strict wildcard with no global declaration admits no valid <N>.
			case unconstrained:
				if bDecl.Type != builtin.AnyType {
					rep.errf(xsd.SpecCosParticleRestrict, rp.Pos, "element %s in the restriction of %s is bound by a wildcard that accepts content the base element's type forbids", n, describeCT(ct))
				}
			case !validlyDerivedByRestriction(t, bDecl.Type):
				rep.errf(derivationFailureRef(bDecl.Type), rp.Pos, "element %s in the restriction of %s is bound by a wildcard to a type that is not derived from the base element's type", n, describeCT(ct))
			}
			break
		}
	}
}
