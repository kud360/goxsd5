package parser

// Restriction support for complexContent/simpleContent derivations, beyond the
// element-particle subsumption relation (which lives in subsumption.go):
//
//   - the {open content} restriction checks (derivation-ok-restriction §3.4.6.4
//     clause 9, and the cos-ct-extends open-content clause for extensions);
//   - the attribute-use and attribute-wildcard restriction checks (§3.4.6.3);
//   - the type-derivation relations restriction shares with substitution-group
//     and type-alternative validation (validlyDerivedByRestriction /
//     validlySubstitutable, cos-st-derived-ok §3.16.6.3);
//   - the flat-particle support types and helpers (baseSlot, baseRegion,
//     representative-name collection, occurrence arithmetic) the §3.9.6 relation
//     in subsumption.go is built on.

import (
	"sort"

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

// restrictPart is one restriction-run particle reduced to what the packing
// analysis needs: a membership predicate, its effective occurrence range, and
// (for a named element) its declaration for the type/nillability check.
type restrictPart struct {
	accepts    func(xsd.QName) bool
	rmin, rmax int
	pos        xsd.Pos
	decl       *xsd.ElementDecl
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
	// Sort so the representative set — and thus which offending name a diagnostic
	// reports — is deterministic (qset/nsset iterate in random map order).
	sort.Slice(reps, func(i, j int) bool {
		if reps[i].Namespace != reps[j].Namespace {
			return reps[i].Namespace < reps[j].Namespace
		}
		return reps[i].Local < reps[j].Local
	})
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
func (b *builder) checkExtensionOpenContent(ct, bct *xsd.ComplexType, contentNode *xmltree.Node, doc *schemaDoc) {
	if contentNode == nil {
		return // simple-content extension: no element-content open content
	}
	bec, _ := bct.Content.(*xsd.ElementContent)
	bot := effectiveOpenContent(bec)
	if bot == nil || bot.Mode != xsd.OpenContentInterleave {
		// Base has no open content, or it is already suffix: clause 1.4.3.2.2.3.1
		// holds, or there is nothing more open to narrow.
		return
	}
	eotMode := bot.Mode // inherited when no wildcard element governs the extension
	if ocn := firstChild(contentNode, doc, "openContent"); ocn != nil {
		if m := openContentModeAttr(ocn); m != xsd.OpenContentNone {
			eotMode = m // mode=none ⇒ inherit the base (mapping clause 6.1)
		}
	} else if doc.defaultOpenContent != nil {
		if !contentEmpty(ct) || boolAttr(doc.defaultOpenContent, "appliesToEmpty", false) {
			if m := openContentModeAttr(doc.defaultOpenContent); m != xsd.OpenContentNone {
				eotMode = m
			}
		}
	}
	if eotMode == xsd.OpenContentSuffix {
		// spec: cos-ct-extends — XSD 1.1 Part 1 §3.4.6.2 clause 1.4.3.2.2.3
		b.errf(xsd.SpecCosCTExtends, ct.Pos, "the extension %s narrows the base type's interleaved open content to suffix", describeCT(ct))
	}
}

// inheritExtensionOpenContent finalizes an extension's {open content} per the
// §3.4.2.3.3 mapping. The *explicit content type* of an element-only/mixed
// extension already carries the base's {open content} (clause 4.2.3); the
// extension's own <openContent>/<defaultOpenContent> (recorded in ec.OpenContent
// by fillElementOnlyContent as the "wildcard element") then layers on top:
//   - clause 6.1: the wildcard element is absent or mode='none' ⇒ the result is
//     the explicit content type, i.e. the BASE's open content (so an extension
//     can never suppress inherited open content — mode='none' is not removal).
//   - clause 6.2: otherwise the result keeps the wildcard element's mode and its
//     wildcard is the union of the wildcard element's and the base's.
func (b *builder) inheritExtensionOpenContent(ct, bct *xsd.ComplexType) {
	ec, ok := ct.Content.(*xsd.ElementContent)
	if !ok {
		return
	}
	bec, _ := bct.Content.(*xsd.ElementContent)
	baseOC := effectiveOpenContent(bec) // the explicit content type's open content
	own := ec.OpenContent               // the wildcard element (nil if none authored)
	switch {
	case own == nil || own.Mode == xsd.OpenContentNone:
		ec.OpenContent = baseOC // clause 6.1
	case baseOC == nil:
		// clause 6.2 with an absent base open content: the wildcard element stands.
	default:
		// clause 6.2: combine. {mode} is the wildcard element's; {wildcard} is the
		// wildcard union of the wildcard element's and the base's (processContents
		// from the wildcard element — wildcardUnion takes it from its 2nd arg).
		ec.OpenContent = &xsd.OpenContent{
			Mode:     own.Mode,
			Wildcard: wildcardUnion(baseOC.Wildcard, own.Wildcard),
			Pos:      own.Pos,
		}
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
	if ct.IsMixed() {
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

// particleEmptiable reports whether p can match the empty sequence: either the
// occurrence minimum is zero, or (for group terms) the compositor rule holds
// recursively. A nil particle is trivially emptiable.
// spec: Particle Emptiable §3.8.4.1 — XSD 1.1 Part 1.
func particleEmptiable(p *xsd.Particle) bool {
	if p == nil {
		return true
	}
	if p.MinOccurs == 0 {
		return true
	}
	switch term := p.Term.(type) {
	case *xsd.ElementDecl, *xsd.Wildcard:
		return false // a required element or wildcard always needs one occurrence
	case *xsd.ModelGroup:
		switch term.Compositor {
		case xsd.CompositorChoice:
			// emptiable if any particle is emptiable
			for _, c := range term.Particles {
				if particleEmptiable(c) {
					return true
				}
			}
			return false
		default: // all and sequence: emptiable iff every particle is emptiable
			for _, c := range term.Particles {
				if !particleEmptiable(c) {
					return false
				}
			}
			return true
		}
	case *xsd.GroupRef:
		if term.Ref == nil || term.Ref.Group == nil {
			return true // unresolved group: treat as emptiable to avoid false positives
		}
		// A GroupRef particle wraps the group's model group; its own MinOccurs
		// already gated above, so we check the group's compositor directly.
		mg := term.Ref.Group
		switch mg.Compositor {
		case xsd.CompositorChoice:
			for _, c := range mg.Particles {
				if particleEmptiable(c) {
					return true
				}
			}
			return false
		default:
			for _, c := range mg.Particles {
				if !particleEmptiable(c) {
					return false
				}
			}
			return true
		}
	}
	return false
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
	f := st.EffectiveFacets()
	return len(f.PatternGroups) == 0 && !f.HasEnumeration() && len(f.Assertions) == 0 &&
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
		for _, m := range st.BasicMembers() {
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
