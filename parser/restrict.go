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

import (
	"github.com/kud360/goxsd5/builtin"
	"github.com/kud360/goxsd5/parser/xmltree"
	"github.com/kud360/goxsd5/xsd"
)

// baseSlot is one particle of a flat base content model: an element
// declaration (with its accepted-name set) or the single wildcard, plus the
// effective occurrence range that base particle contributes.
type baseSlot struct {
	decl     *xsd.ElementDecl // element slot
	wc       *xsd.Wildcard    // wildcard slot
	min, max int
	names    map[xsd.QName]bool
}

// checkParticleRestrict reports cos-particle-restrict violations for ct when ct
// is a complexContent restriction of a complex base whose base content model is
// a flat all/sequence of element/wildcard particles. The restriction content
// model may be the same flat shape (one run) or a choice of flat branches, in
// which case each branch is checked independently: an instance matching the
// choice matches exactly one branch, so every branch must on its own be a valid
// restriction of the base.
func (b *builder) checkParticleRestrict(ct *xsd.ComplexType, accepted func(*xsd.ElementDecl) map[xsd.QName]bool, globalsByName map[xsd.QName]*xsd.ElementDecl) {
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
		return // open-content subset is handled by checkOpenContentRestrict
	}

	bParts, bTop, bOK := flatGroup(bec.Particle)
	if !bOK {
		return // base is not a flat all/sequence: give up safely
	}
	// A base named element unconditionally shadows an overlapping wildcard only
	// in an xs:all group, where particle matching is order-independent. In a
	// sequence the precedence is positional, so an element appearing past the
	// named particle's slot legitimately routes to the wildcard in both base and
	// restriction (saxon wild068) — there the shadow check below must not fire.
	bmg, _ := bec.Particle.Term.(*xsd.ModelGroup)
	baseIsAll := bmg != nil && bmg.Compositor == xsd.CompositorAll

	// With two or more base wildcards a single per-base-particle slot cannot say
	// which one a restriction particle maps to (a restriction wildcard may even
	// straddle several base wildcards), so the bag check below cannot decide it.
	// Hand those off to the multi-wildcard solver, which reasons about the whole
	// packing instead of a one-to-one slot mapping.
	nWC := 0
	for _, bp := range bParts {
		if _, ok := bp.Term.(*xsd.Wildcard); ok {
			nWC++
		}
	}
	if nWC >= 2 {
		b.checkMultiWildcardRestrict(ct, bParts, bTop, rec, accepted)
		return
	}

	// Build a slot per base particle (zero or one base wildcard).
	slots := make([]*baseSlot, 0, len(bParts))
	baseWC := -1
	for _, bp := range bParts {
		s := &baseSlot{min: mulOcc(bTop.MinOccurs, bp.MinOccurs), max: mulOcc(bTop.MaxOccurs, bp.MaxOccurs)}
		switch term := bp.Term.(type) {
		case *xsd.ElementDecl:
			s.decl = term
			s.names = accepted(term)
		case *xsd.Wildcard:
			s.wc = term
			baseWC = len(slots)
		}
		slots = append(slots, s)
	}

	if rParts, rTop, rOK := flatGroup(rec.Particle); rOK {
		b.checkRestrictRun(ct, slots, baseWC, rParts, rTop)
		if baseIsAll {
			b.checkWildcardShadowsNamed(ct, slots, rParts, accepted, globalsByName)
		}
		return
	}
	// A choice restriction: validate every branch on its own. A branch term that
	// is a wildcard, group reference, nested choice, or occurs other than once
	// makes the whole choice unanalyzable, so give up.
	branches, rOK := choiceBranches(rec.Particle)
	if !rOK {
		return
	}
	unitTop := &xsd.Particle{MinOccurs: 1, MaxOccurs: 1}
	for _, br := range branches {
		b.checkRestrictRun(ct, slots, baseWC, br, unitTop)
		if baseIsAll {
			b.checkWildcardShadowsNamed(ct, slots, br, accepted, globalsByName)
		}
	}
}

// checkRestrictRun checks one flat run of restriction particles (rParts, each
// scaled by rTop's occurrence) against the base slots, reporting any
// cos-particle-restrict violation. Occurrence accumulation is local to the run,
// so the same slots can be reused for each branch of a choice.
func (b *builder) checkRestrictRun(ct *xsd.ComplexType, slots []*baseSlot, baseWC int, rParts []*xsd.Particle, rTop *xsd.Particle) {
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
			// Find the base element particle whose accepted names include this
			// element (exact name or substitution-group membership).
			slot := -1
			ambiguous := false
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
				if !validlyDerivedByRestriction(term.Type, slots[slot].decl.Type) {
					// spec: cos-particle-restrict — §3.4.6.4 clause 2 /
					// NameAndTypeOK: the restricting element's type must derive
					// from the base's (appealing to cos-st-derived-ok for a
					// union base, §3.16.6.3 clause 2.2.4).
					b.errf(derivationFailureRef(slots[slot].decl.Type), rp.Pos, "element %s in the restriction of %s has a type that is not derived from the base element's type", term.Name, describeCT(ct))
				}
				if !slots[slot].decl.Nillable && term.Nillable {
					b.errf(xsd.SpecCosParticleRestrict, rp.Pos, "element %s in the restriction of %s may not be nillable when the base element is not", term.Name, describeCT(ct))
				}
				if !typeTablesEqual(term, slots[slot].decl) {
					// spec: cos-particle-restrict — §3.4.6.4 / "subsumes" clause 4.6:
					// a restriction element's {type table} must be equivalent to the
					// base element's (you cannot change conditional type assignment
					// when restricting). Equality is conservative — anonymous
					// alternative types compare by zero name — so this only ever
					// misses a violation, never invents one.
					b.errf(xsd.SpecCosParticleRestrict, rp.Pos, "element %s in the restriction of %s has a type table that differs from the base element's", term.Name, describeCT(ct))
				}
				continue
			}
			// No base element matches: the element must be accepted by the base
			// wildcard (NSCompat), else the restriction introduces a name the
			// base disallows.
			if baseWC >= 0 && wildcardAllowsName(slots[baseWC].wc, term.Name) {
				mapTo(baseWC, rp)
				continue
			}
			// spec: cos-particle-restrict — XSD 1.1 Part 1 §3.4.6.4 clause 1
			b.errf(xsd.SpecCosParticleRestrict, rp.Pos, "element %s in the restriction of %s is not allowed by the base content model", term.Name, describeCT(ct))
		case *xsd.Wildcard:
			// A restriction wildcard can only restrict the base wildcard, and
			// only if it is a wildcard subset of it (NSSubset).
			if baseWC < 0 {
				// spec: cos-particle-restrict — §3.4.6.4 clause 1: the base has no
				// wildcard, so the restriction may not introduce one.
				b.errf(xsd.SpecCosParticleRestrict, rp.Pos, "the restriction of %s introduces a wildcard the base content model does not allow", describeCT(ct))
				continue
			}
			if !namespaceConstraintSubset(term, slots[baseWC].wc) {
				// spec: cos-particle-restrict — §3.4.6.4 clause 1 / NSSubset:
				// the restricting wildcard must be a subset of the base's.
				b.errf(xsd.SpecCosParticleRestrict, rp.Pos, "a wildcard in the restriction of %s allows elements the base wildcard does not", describeCT(ct))
			}
			mapTo(baseWC, rp)
		}
	}

	for i, s := range slots {
		if !mapped[i] {
			if s.min > 0 {
				// spec: cos-particle-restrict — §3.4.6.4 clause 1: required base
				// content must remain required in the restriction.
				b.errf(xsd.SpecCosParticleRestrict, rTop.Pos, "the restriction of %s omits required base content (%s)", describeCT(ct), slotName(s.decl))
			}
			continue
		}
		if sumMin[i] < s.min {
			b.errf(xsd.SpecCosParticleRestrict, rTop.Pos, "the restriction of %s allows %s fewer times than the base requires", describeCT(ct), slotName(s.decl))
		}
		if !occLE(sumMax[i], s.max) {
			b.errf(xsd.SpecCosParticleRestrict, rTop.Pos, "the restriction of %s allows %s more times than the base permits", describeCT(ct), slotName(s.decl))
		}
	}
}

// checkWildcardShadowsNamed reports the cos-particle-restrict / NameAndTypeOK
// violation the plain NSSubset wildcard check misses. The caller invokes this
// only for an xs:all base: there matching is order-independent, so a base named
// element particle for N unconditionally takes UPA precedence over an
// overlapping wildcard (XSD 1.1 lets element declarations shadow it) and B
// always types an <N> child by that named declaration. (In a sequence the
// precedence is positional, so an <N> past the named slot routes to the
// wildcard in both base and restriction — wild068 — which is why the caller
// gates on the all compositor.) If this restriction run drops the named
// particle for N — no
// element particle in the run accepts N — yet a wildcard in the run matches N,
// then R instead routes <N> to that wildcard, whose effective type for N must
// be validly derived from B's named type. Otherwise there is a concrete <N>
// (content suiting the wildcard's binding but not the base type) valid in R and
// invalid in B, so reporting the failure never rejects a valid schema. This is
// exactly the standard NameAndTypeOK obligation, applied across the
// element-shadows-wildcard routing the per-particle NSSubset check cannot see.
// spec: cos-particle-restrict — XSD 1.1 Part 1 §3.4.6.4 clause 2 (catches
// saxon wild069: an xs:all restriction whose ##local lax wildcard binds a global
// e:xs:duration where the base's named e is union(date,time)).
func (b *builder) checkWildcardShadowsNamed(ct *xsd.ComplexType, slots []*baseSlot, rParts []*xsd.Particle, accepted func(*xsd.ElementDecl) map[xsd.QName]bool, globalsByName map[xsd.QName]*xsd.ElementDecl) {
	// Names this run still routes to a named particle (each element's own name
	// plus every substitutable member): they keep the named-vs-named type check
	// and are never re-routed to a wildcard. Over-counting here only suppresses
	// reports, so it cannot introduce a false positive.
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
	for _, s := range slots {
		if s.decl == nil {
			continue // only base named particles shadow a wildcard
		}
		n := s.decl.Name
		if rNamed[n] {
			continue // R keeps a named particle for N: the type check covers it
		}
		// UPA makes R route <N> to at most one wildcard; find it.
		for _, rp := range rWild {
			w := rp.Term.(*xsd.Wildcard)
			if !wildcardAllowsName(w, n) {
				continue
			}
			t, unconstrained, hasInstance := b.wildcardBoundType(w, n, globalsByName)
			switch {
			case !hasInstance:
				// A strict wildcard with no global declaration admits no valid
				// <N>, so R has no instance to be unsound about.
			case unconstrained:
				// A skip wildcard (or lax with no global) accepts any content for
				// <N>; only an anyType base named particle is that permissive.
				if s.decl.Type != builtin.AnyType {
					b.errf(xsd.SpecCosParticleRestrict, rp.Pos, "element %s in the restriction of %s is bound by a wildcard that accepts content the base element's type forbids", n, describeCT(ct))
				}
			case !validlyDerivedByRestriction(t, s.decl.Type):
				b.errf(derivationFailureRef(s.decl.Type), rp.Pos, "element %s in the restriction of %s is bound by a wildcard to a type that is not derived from the base element's type", n, describeCT(ct))
			}
			break
		}
	}
}

// wildcardBoundType gives the effective type a wildcard assigns to an element
// named n when it binds it: the matched global declaration's type under
// strict/lax assessment, or an unconstrained binding (skip, or lax with no
// global, which is laxly assessed to no declaration and so accepts any content).
// A strict wildcard with no global declaration admits no valid element of that
// name, reported by hasInstance == false.
func (b *builder) wildcardBoundType(w *xsd.Wildcard, n xsd.QName, globalsByName map[xsd.QName]*xsd.ElementDecl) (t xsd.Type, unconstrained, hasInstance bool) {
	switch w.ProcessContents {
	case xsd.ProcessSkip:
		return nil, true, true
	case xsd.ProcessLax:
		if g := globalsByName[n]; g != nil {
			return g.Type, false, true
		}
		return nil, true, true
	default: // strict
		if g := globalsByName[n]; g != nil {
			return g.Type, false, true
		}
		return nil, false, false
	}
}

// baseRegion is one base content-model particle viewed as a region of the
// element-name universe: a predicate that recognises exactly the names that
// particle accepts (an element's own name plus its substitution-group members,
// or a wildcard's namespace constraint), plus the occurrence range that
// particle contributes. In a UPA-valid base group the regions are pairwise
// disjoint, so every element matches at most one base particle and a region's
// count is exactly the number of instance elements that fall in it.
type baseRegion struct {
	accepts  func(xsd.QName) bool
	min, max int
	decl     *xsd.ElementDecl // named region (for type/nillability + diagnostics)
	name     string           // diagnostic label
}

// checkMultiWildcardRestrict reports cos-particle-restrict violations for a
// complexContent restriction whose base content model is a flat all/sequence
// holding two or more wildcards. The one-to-one slot mapping used for the
// single-wildcard case breaks down here (a restriction particle may be coverable
// by several base wildcards, or straddle them), so this reasons about the whole
// packing: it builds the base particles as disjoint name regions and, for each
// flat restriction run, checks (1) the restriction never admits an element no
// base region accepts, and (2) for every base region the restriction-allowed
// count stays within that region's occurrence range. Both are exact witnesses —
// a violation always yields a concrete instance valid against R but not B — so
// no valid restriction is ever rejected.
func (b *builder) checkMultiWildcardRestrict(ct *xsd.ComplexType, bParts []*xsd.Particle, bTop *xsd.Particle, rec *xsd.ElementContent, accepted func(*xsd.ElementDecl) map[xsd.QName]bool) {
	// wildcardAllowsName ignores the dynamic ##defined/##definedSibling
	// sentinels, so a wildcard carrying one would be treated as accepting more
	// names than it really does and a witness built on it could be unsound. Give
	// up if any base wildcard uses them (the restriction side is guarded per run).
	for _, bp := range bParts {
		if w, ok := bp.Term.(*xsd.Wildcard); ok && wildcardHasSentinel(w) {
			return
		}
	}

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

	if rParts, rTop, rOK := flatGroup(rec.Particle); rOK {
		b.checkMultiWildcardRun(ct, regions, bParts, rParts, rTop)
		return
	}
	branches, rOK := choiceBranches(rec.Particle)
	if !rOK {
		return
	}
	unitTop := &xsd.Particle{MinOccurs: 1, MaxOccurs: 1, Pos: rec.Particle.Pos}
	for _, br := range branches {
		b.checkMultiWildcardRun(ct, regions, bParts, br, unitTop)
	}
}

// restrictPart is one restriction-run particle reduced to what the packing
// analysis needs: a membership predicate, its effective occurrence range, and
// (for a named element) its declaration for the type/nillability check.
type restrictPart struct {
	accepts    func(xsd.QName) bool
	rmin, rmax int
	pos        xsd.Pos
	decl       *xsd.ElementDecl
}

// checkMultiWildcardRun checks one flat run of restriction particles against the
// base regions. bParts is the raw base run, used only to gather representative
// names. See checkMultiWildcardRestrict for the soundness argument.
func (b *builder) checkMultiWildcardRun(ct *xsd.ComplexType, regions []baseRegion, bParts, rParts []*xsd.Particle, rTop *xsd.Particle) {
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

	// (1) Every name a restriction particle can produce must be accepted by some
	// base region; otherwise the restriction admits an element the base forbids.
	// This holds however the base regions overlap, so it always runs.
	for _, e := range rps {
		if e.rmax == 0 {
			continue
		}
		for _, n := range reps {
			if e.accepts(n) && !anyRegionAccepts(regions, n) {
				// spec: cos-particle-restrict — §3.4.6.4 clause 1: the restriction
				// may not admit an element no base particle accepts.
				b.errf(xsd.SpecCosParticleRestrict, e.pos, "the restriction of %s allows an element (%s) that the base content model does not", describeCT(ct), n)
				break
			}
		}
	}

	// The per-region count and type checks below are only sound when the base
	// regions are pairwise disjoint: then every instance element matches exactly
	// one base particle, so base validity reduces to each region count lying in
	// range. When a base wildcard overlaps the base element it sits beside (legal
	// in an <all>, where element and wildcard particles do not compete) an element
	// can be absorbed by either, and the real test is a flow problem we give up on
	// rather than risk a false positive (wild047/wild049 are valid that way).
	if !regionsDisjoint(regions, reps) {
		return
	}

	// (2) Per region, the restriction-allowed count must stay within the base
	// region's range. The minimum count is the sum of rmin over restriction
	// particles trapped wholly inside the region (they cannot escape it); the
	// maximum is the sum of rmax over particles that can reach it. Both bounds
	// are achievable simultaneously by an independent per-particle placement, so
	// they are exact.
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
			// spec: cos-particle-restrict — §3.4.6.4 clause 1: required base content
			// must stay required; the restriction can yield fewer than the base needs.
			b.errf(xsd.SpecCosParticleRestrict, rTop.Pos, "the restriction of %s may match %s fewer times than the base requires", describeCT(ct), reg.name)
		}
		if !occLE(maxB, reg.max) {
			b.errf(xsd.SpecCosParticleRestrict, rTop.Pos, "the restriction of %s allows %s more times than the base permits", describeCT(ct), reg.name)
		}
	}

	// (3) A restriction named element landing in a base named region must derive
	// from it by restriction and may not widen nillability (NameAndTypeOK).
	for _, e := range rps {
		if e.decl == nil {
			continue
		}
		for i := range regions {
			reg := &regions[i]
			if reg.decl == nil || reg.decl.Name != e.decl.Name {
				continue
			}
			if !validlyDerivedByRestriction(e.decl.Type, reg.decl.Type) {
				b.errf(derivationFailureRef(reg.decl.Type), e.pos, "element %s in the restriction of %s has a type that is not derived from the base element's type", e.decl.Name, describeCT(ct))
			}
			if !reg.decl.Nillable && e.decl.Nillable {
				b.errf(xsd.SpecCosParticleRestrict, e.pos, "element %s in the restriction of %s may not be nillable when the base element is not", e.decl.Name, describeCT(ct))
			}
		}
	}
}

// regionsDisjoint reports whether the base regions are pairwise disjoint over
// the representative names: no representative name is accepted by two regions.
// reps are exhaustive of the predicates' behaviour, so this decides true
// disjointness.
func regionsDisjoint(regions []baseRegion, reps []xsd.QName) bool {
	for _, n := range reps {
		hits := 0
		for i := range regions {
			if regions[i].accepts(n) {
				hits++
				if hits > 1 {
					return false
				}
			}
		}
	}
	return true
}

// anyRegionAccepts reports whether some base region accepts the name q.
func anyRegionAccepts(regions []baseRegion, q xsd.QName) bool {
	for i := range regions {
		if regions[i].accepts(q) {
			return true
		}
	}
	return false
}

// wildcardHasSentinel reports whether w's {disallowed names} include a dynamic
// ##defined/##definedSibling keyword.
func wildcardHasSentinel(w *xsd.Wildcard) bool {
	for _, d := range w.NotQName {
		if d.Local == definedKeyword || d.Local == siblingKeyword {
			return true
		}
	}
	return false
}

// collectReps returns a finite set of expanded names that is exhaustive for the
// wildcard/element predicates appearing in bParts and rParts: a wildcard accepts
// a name by its namespace (a finite mentioned set, plus "everything else") minus
// a finite set of disallowed QNames, so one generic name per mentioned namespace
// (a local guaranteed distinct from every mentioned QName), one generic name in
// a never-mentioned namespace, and every explicitly mentioned QName together
// distinguish all behaviours these predicates can have.
func collectReps(bParts, rParts []*xsd.Particle) []xsd.QName {
	qset := map[xsd.QName]bool{}
	nsset := map[string]bool{"": true}
	visit := func(parts []*xsd.Particle) {
		for _, p := range parts {
			switch t := p.Term.(type) {
			case *xsd.ElementDecl:
				qset[t.Name] = true
				nsset[t.Name.Namespace] = true
			case *xsd.Wildcard:
				for _, ns := range t.Namespaces {
					nsset[ns] = true
				}
				for _, d := range t.NotQName {
					if d.Local == definedKeyword || d.Local == siblingKeyword {
						continue
					}
					qset[d] = true
					nsset[d.Namespace] = true
				}
			}
		}
	}
	visit(bParts)
	visit(rParts)

	// A local containing NUL cannot be a real XML name, so it never collides with
	// any mentioned QName in the same namespace.
	const genericLocal = "\x00generic"
	reps := make([]xsd.QName, 0, len(qset)+len(nsset)+1)
	for q := range qset {
		reps = append(reps, q)
	}
	for ns := range nsset {
		reps = append(reps, xsd.QName{Namespace: ns, Local: genericLocal})
	}
	freshNS := "\x00fresh"
	for nsset[freshNS] {
		freshNS += "x"
	}
	reps = append(reps, xsd.QName{Namespace: freshNS, Local: genericLocal})
	return reps
}

// choiceBranches returns the branches of a restriction content model that is a
// choice (occurring exactly once) of flat element runs: each branch is a single
// element/wildcard particle or a once-occurring all/sequence of them. It reports
// ok=false (give up) for a repeating choice or any branch that is itself a
// wildcard, a nested choice, a group reference, or occurs other than once.
func choiceBranches(p *xsd.Particle) (branches [][]*xsd.Particle, ok bool) {
	mg, isMG := p.Term.(*xsd.ModelGroup)
	if !isMG || mg.Compositor != xsd.CompositorChoice || p.MinOccurs != 1 || p.MaxOccurs != 1 {
		return nil, false
	}
	for _, c := range mg.Particles {
		switch c.Term.(type) {
		case *xsd.ElementDecl:
			branches = append(branches, []*xsd.Particle{c})
		case *xsd.ModelGroup:
			// A nested model group branch must be a flat all/sequence occurring
			// exactly once, so each instance taking the branch produces its
			// elements exactly once.
			parts, _, fok := flatGroup(c)
			if !fok || c.MinOccurs != 1 || c.MaxOccurs != 1 {
				return nil, false
			}
			branches = append(branches, parts)
		default:
			return nil, false // wildcard or group-ref branch: give up
		}
	}
	return branches, true
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

// checkExtensionOpenContent enforces the open-content clause of Derivation
// Valid (Extension) (cos-ct-extends §3.4.6.2 clause 1.4.3.2.2.3): when a
// complex type extends a base whose {open content} has {mode} interleave, the
// extension's own {open content} may not narrow it to suffix. (Clause
// 1.4.3.2.2.4's namespace-subset requirement is automatically met: by the
// {content type} mapping §3.4.2.2 clause 6.2 the extension's open-content
// wildcard is the union of its own and the base's, so the base's is always a
// subset.)
//
// The extension's effective {open content} mode (EOT) is determined per the
// {open content} mapping clauses 5-6: an explicit <openContent> child wins;
// otherwise the schema's <defaultOpenContent> applies when the EXPLICIT content
// type is not empty — and for an extension the explicit content type is the
// post-merge content, so a non-empty base makes it non-empty even when the
// extension's own particle is empty (W3C bug 13459, open046); otherwise the
// extension inherits the base's open content unchanged (no narrowing).
func (b *builder) checkExtensionOpenContent(ct, bct *xsd.ComplexType, p *pendingAttrs) {
	bec, _ := bct.Content.(*xsd.ElementContent)
	bot := effectiveOpenContent(bec)
	if bot == nil || bot.Mode != xsd.OpenContentInterleave {
		// Base has no open content, or it is already suffix: clause 1.4.3.2.2.3.1
		// holds, or there is nothing more open to narrow.
		return
	}
	eotMode := bot.Mode // inherited when no wildcard element governs the extension
	if ocn := firstChild(p.contentNode, p.doc, "openContent"); ocn != nil {
		if m := openContentModeAttr(ocn); m != xsd.OpenContentNone {
			eotMode = m // mode=none ⇒ inherit the base (mapping clause 6.1)
		}
	} else if p.doc.defaultOpenContent != nil {
		if !contentEmpty(ct) || boolAttr(p.doc.defaultOpenContent, "appliesToEmpty", false) {
			if m := openContentModeAttr(p.doc.defaultOpenContent); m != xsd.OpenContentNone {
				eotMode = m
			}
		}
	}
	if eotMode == xsd.OpenContentSuffix {
		// spec: cos-ct-extends — XSD 1.1 Part 1 §3.4.6.2 clause 1.4.3.2.2.3
		b.errf(xsd.SpecCosCTExtends, ct.Pos, "the extension %s narrows the base type's interleaved open content to suffix", describeCT(ct))
	}
}

// openContentModeAttr reads the {mode} of an <openContent>/<defaultOpenContent>
// element, defaulting to interleave (mirrors buildOpenContent).
func openContentModeAttr(n *xmltree.Node) xsd.OpenContentMode {
	if v, ok := n.Attr("mode"); ok {
		switch v {
		case "suffix":
			return xsd.OpenContentSuffix
		case "none":
			return xsd.OpenContentNone
		}
	}
	return xsd.OpenContentInterleave
}

// contentEmpty reports whether ct's content type has {variety} empty: no
// character content (not mixed) and no particle that can match an element.
func contentEmpty(ct *xsd.ComplexType) bool {
	if ct.Mixed {
		return false
	}
	ec, ok := ct.Content.(*xsd.ElementContent)
	if !ok {
		return true // EmptyContent (SimpleContent never carries open content)
	}
	return !ec.Mixed && !particleMatchesNonEmpty(ec.Particle)
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
// validly derives from the base element's type (cos-st-derived-ok §3.16.6.3,
// empty blocking set). It is a necessary condition for L(R) ⊆ L(B): an
// undeterminable type (nil) is accepted, and it never rejects a genuinely valid
// derivation, so reporting its failure is always a real error.
//
// When the base type is a union the decision is EXACT via clause 2.2.4 (a type
// is validly derived from a union only if it derives from a type in the union's
// transitive membership AND the union plus every intervening union carry no
// facets). When only the *restriction* type involves a union but the base does
// not, the relation can't be decided cleanly from the flattened model, so it is
// accepted (no false positive).
// derivationFailureRef names the schema constraint that a NameAndTypeOK
// type-derivation failure appeals to: cos-st-derived-ok (§3.16.6.3) when the
// base element's type is a union simple type — the clause 2.2.4 cases — and the
// generic cos-particle-restrict otherwise. The reported id is purely diagnostic;
// the schema is invalid either way.
func derivationFailureRef(bType xsd.Type) xsd.SpecRef {
	if st, ok := bType.(*xsd.SimpleType); ok && st.Variety == xsd.VarietyUnion {
		return xsd.SpecCosSTDerivedOK
	}
	return xsd.SpecCosParticleRestrict
}

func validlyDerivedByRestriction(rType, bType xsd.Type) bool {
	if rType == nil || bType == nil || rType == bType {
		return true
	}
	// Ordinary derivation chain (cos-st-derived-ok clauses 1, 2.2.1, 2.2.2):
	// the base type appears in the restriction type's base chain.
	if _, ok := derivationMethods(rType, bType); ok {
		return true
	}
	// cos-st-derived-ok clause 2.2.4: the base type is a union and the
	// restriction type is validly derived from one of its members. This is the
	// only remaining way to be validly derived from a union base, so a negative
	// result is decisive (catches saxon simple011/014/015).
	if bst, ok := bType.(*xsd.SimpleType); ok && bst.Variety == xsd.VarietyUnion {
		return stDerivedFromUnion(rType, bst)
	}
	// Base is not a union but the restriction type involves one: the flattened
	// model can't decide reliably — give up (accept) rather than risk rejecting
	// a valid schema.
	if involvesUnion(rType) {
		return true
	}
	return false
}

// stDerivedFromUnion decides cos-st-derived-ok clause 2.2.4 (§3.16.6.3): whether
// simple type d is validly derived from the union type u. Per the clause, d must
// be validly derived from some type in u's transitive membership (clause 2.2.4.2)
// while the {facets} of u and of every intervening union are empty (clause
// 2.2.4.3 — a facet anywhere on the path makes the member no longer substitutable
// for the facet-restricted union, the XSD 1.1 correction noted at §G.2).
//
// The walk uses DirectMembers (the un-flattened {member type definitions}); a
// member reachable by an ordinary derivation chain is the endpoint M and its own
// facets are unconstrained, whereas a member union we must pass *through* is an
// intervening union whose facets are required empty by the check at the top of
// the recursive call. The decision is exact, so a false result is never a false
// positive.
func stDerivedFromUnion(d xsd.Type, u *xsd.SimpleType) bool {
	if !unionFacetsEmpty(u) {
		return false // clause 2.2.4.3: u (the base or an intervening union) carries facets
	}
	for _, m := range u.DirectMembers {
		// clause 2.2.4.2: d validly derived from member m (m is the endpoint).
		if _, ok := derivationMethods(d, m); ok {
			return true
		}
		// Otherwise m is an intervening union; recurse with its facets checked.
		if m.Variety == xsd.VarietyUnion && stDerivedFromUnion(d, m) {
			return true
		}
	}
	return false
}

// unionFacetsEmpty reports whether a union type carries no constraining facets,
// the condition cos-st-derived-ok clause 2.2.4.3 imposes on the base and every
// intervening union. (whiteSpace does not apply to unions, so only the
// value-constraining facets are relevant.)
func unionFacetsEmpty(st *xsd.SimpleType) bool {
	f := &st.Facets
	return len(f.PatternGroups) == 0 && !f.HasEnumeration && len(f.Assertions) == 0 &&
		f.Length == nil && f.MinLength == nil && f.MaxLength == nil &&
		f.MinInclusive == nil && f.MaxInclusive == nil &&
		f.MinExclusive == nil && f.MaxExclusive == nil &&
		f.TotalDigits == nil && f.FractionDigits == nil
}

// validlySubstitutable reports whether type s is validly substitutable for type
// t subject to the blocking keywords in block (key-val-sub-type, §3.16.6.3 /
// §3.4.6.5): s must be validly derived from t and none of the derivation steps
// may use a method named in block, unioned — for a complex base — with t's own
// {prohibited substitutions}. A union base is honoured per cos-st-derived-ok
// clause 2.2.4: a type derived from one of the union's (flattened) members is
// validly derived from the union, so it stays substitutable. The relation is a
// necessary condition, so reporting its failure never costs a false positive.
func validlySubstitutable(s, t xsd.Type, block xsd.DerivationSet) bool {
	if s == nil || t == nil || s == t {
		return true
	}
	// xs:anyType is the root of the type hierarchy: every type derives from it
	// (and it imposes no prohibited substitutions), so any type is substitutable
	// for it. A plain complex type's {base type definition} is left implicit
	// (nil) rather than pointing at the xs:anyType singleton, which would
	// otherwise defeat the chain walk below.
	if t == builtin.AnyType {
		return true
	}
	if methods, ok := derivationMethods(s, t); ok {
		eff := block
		if ct, isComplex := t.(*xsd.ComplexType); isComplex {
			eff |= ct.Block // {prohibited substitutions}
		}
		return methods&eff == 0
	}
	// cos-st-derived-ok clause 2.2.4: s may instead be validly derived from a
	// member of a union base. Member-union flattening means the members are
	// tested directly.
	if st, ok := t.(*xsd.SimpleType); ok && st.Variety == xsd.VarietyUnion {
		for _, m := range st.MemberTypes {
			if validlySubstitutable(s, m, block) {
				return true
			}
		}
	}
	return false
}

// checkAttrRestriction enforces the attribute-use half of Derivation Valid
// (Restriction, Complex) (§3.4.6.3 clause 3, via the "subsumes" relation,
// clause 5.3): when a restriction redeclares a base attribute use of the same
// name, the {inheritable} value must not change. (The companion clauses —
// type-derivation 5.1 and value-constraint 5.2 — are checked or deferred
// elsewhere; this adds only the always-sound inheritability equality, a
// necessary condition for a valid restriction.)
func (b *builder) checkAttrRestriction(own, base []*xsd.AttributeUse) {
	byName := map[xsd.QName]*xsd.AttributeUse{}
	for _, u := range base {
		if u.Decl != nil {
			byName[u.Decl.Name] = u
		}
	}
	for _, u := range own {
		if u.Decl == nil {
			continue
		}
		v, ok := byName[u.Decl.Name]
		if !ok {
			continue
		}
		if u.Decl.Inheritable != v.Decl.Inheritable {
			// spec: derivation-ok-restriction — XSD 1.1 Part 1 §3.4.6.3
			b.errf(xsd.SpecDerivationOKRestriction, u.Pos, "attribute %s may not change its inheritability when restricting the base type", u.Decl.Name)
		}
	}
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
