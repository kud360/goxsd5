package parser

// Schema assembly: build every global component of the loaded documents
// into xsd.Schemas, one per target namespace (include merges documents;
// import links separate namespaces). Redefine/override replacement children
// are global components of their namespace and are added alongside.

import (
	"slices"

	"github.com/kud360/goxsd5/builtin"
	"github.com/kud360/goxsd5/parser/xmltree"
	"github.com/kud360/goxsd5/xsd"
)

// buildSchemas runs pass 2 over every loaded document and returns the
// linked schemas in first-encounter namespace order (the root document's
// namespace first).
func buildSchemas(reg *registry, l *loader, errs *xsd.ErrorList) []*xsd.Schema {
	b := newBuilder(reg, errs)
	byTNS := map[string]*xsd.Schema{}
	var schemas []*xsd.Schema
	for _, doc := range l.order {
		s := byTNS[doc.targetNamespace]
		if s == nil {
			s = newSchemaShell(doc)
			byTNS[doc.targetNamespace] = s
			schemas = append(schemas, s)
		}
		for _, comp := range doc.compositions {
			if comp.kind == "import" && !slices.Contains(s.Imports, comp.namespace) {
				s.Imports = append(s.Imports, comp.namespace)
			}
		}
		b.addDocComponents(s, doc)
	}
	for _, rep := range l.reps {
		if s := byTNS[rep.owner.targetNamespace]; s != nil {
			b.addReplacementComponents(s, rep)
		}
	}
	for _, s := range schemas {
		b.checkTypeCycles(s)
	}
	b.finishComplexTypes()
	return schemas
}

// buildSchema runs pass 2 over a single registered document (no
// compositions) and returns the linked schema.
func buildSchema(reg *registry, doc *schemaDoc, errs *xsd.ErrorList) *xsd.Schema {
	b := newBuilder(reg, errs)
	s := newSchemaShell(doc)
	for _, comp := range doc.compositions {
		if comp.kind == "import" {
			s.Imports = append(s.Imports, comp.namespace)
		}
	}
	b.addDocComponents(s, doc)
	b.checkTypeCycles(s)
	b.finishComplexTypes()
	return s
}

// newSchemaShell creates the schema for a namespace from its first
// document's properties. (Per-document properties like the form defaults
// keep applying per document during building; the schema-level fields are
// informational.)
func newSchemaShell(doc *schemaDoc) *xsd.Schema {
	return &xsd.Schema{
		TargetNamespace:        doc.targetNamespace,
		Location:               doc.uri,
		Pos:                    doc.root.Pos,
		Version:                doc.version,
		ElementFormDefault:     doc.elementFormDefault,
		AttributeFormDefault:   doc.attributeFormDefault,
		BlockDefault:           doc.blockDefault,
		FinalDefault:           doc.finalDefault,
		DefaultAttributesGroup: doc.defaultAttributes,
		Types:                  map[xsd.QName]xsd.Type{},
		Elements:               map[xsd.QName]*xsd.ElementDecl{},
		Attributes:             map[xsd.QName]*xsd.AttributeDecl{},
		Groups:                 map[xsd.QName]*xsd.Group{},
		AttributeGroups:        map[xsd.QName]*xsd.AttributeGroup{},
		Notations:              map[xsd.QName]*xsd.Notation{},
		Extensions:             extensionsOf(doc.root),
	}
}

// addDocComponents builds doc's global components into s.
func (b *builder) addDocComponents(s *xsd.Schema, doc *schemaDoc) {
	for _, c := range xsdElems(doc.root, doc) {
		b.addComponent(s, c, doc)
	}
}

// addReplacementComponents builds the redefine/override children of one
// composition into s. Children the loader did not register (unmatched
// override children, duplicates) are skipped by the same registry check
// that guards ordinary globals.
func (b *builder) addReplacementComponents(s *xsd.Schema, rep *replacement) {
	for _, c := range xsdElems(rep.node, rep.owner) {
		b.addComponent(s, c, rep.owner)
	}
}

// addComponent builds one top-level component node into s. The component is
// built through its registry declaration so that memoization and cycle
// marks are shared with reference resolution. A node that is not the
// registered declaration is a duplicate or suppressed original (already
// reported or replaced); it is skipped rather than built.
func (b *builder) addComponent(s *xsd.Schema, c *xmltree.Node, doc *schemaDoc) {
	name, _ := c.Attr("name")
	q := xsd.QName{Namespace: doc.targetNamespace, Local: name}
	current := func(space space) *decl {
		d := b.reg.lookup(space, q)
		if d == nil || d.node != c {
			return nil
		}
		return d
	}
	switch c.Name.Local {
	case "simpleType", "complexType":
		if d := current(spaceType); d != nil {
			s.Types[q] = b.buildTypeDecl(d)
		}
	case "element":
		if d := current(spaceElement); d != nil {
			s.Elements[q] = b.buildElementDecl(d.node, d.doc, true)
		}
	case "attribute":
		if d := current(spaceAttribute); d != nil {
			s.Attributes[q] = b.buildAttributeDecl(d.node, d.doc, true)
		}
	case "group":
		if d := current(spaceGroup); d != nil {
			if g := b.buildGroup(d); g != nil {
				s.Groups[q] = g
			}
		}
	case "attributeGroup":
		if d := current(spaceAttrGroup); d != nil {
			if g := b.buildAttributeGroup(d); g != nil {
				s.AttributeGroups[q] = g
			}
		}
	case "notation":
		if d := current(spaceNotation); d != nil {
			s.Notations[q] = b.buildNotation(d.node, d.doc)
		}
	case "annotation":
		s.Annotations = append(s.Annotations, buildAnnotation(c, doc))
	}
}

// finishComplexTypes merges every complex type's base-dependent properties:
// the effective particle of extensions (base particle followed by the
// type's own) and the attribute uses inherited from the base. It runs as a
// post-pass because a base can still be mid-build when a derived type is
// constructed (the base's content may legally reach back into the derived
// type), so the base's particle and attribute uses are only known once
// everything is assembled. Bases are finished before the types derived from
// them; derivation chains are acyclic here (checkTypeCycles broke any cycle).
func (b *builder) finishComplexTypes() {
	done := map[*xsd.ComplexType]bool{}
	var finish func(ct *xsd.ComplexType)
	finish = func(ct *xsd.ComplexType) {
		if done[ct] {
			return
		}
		done[ct] = true
		bct, _ := ct.BaseType.(*xsd.ComplexType)
		if bct != nil {
			finish(bct)
		}
		if bct != nil && ct.DerivationMethod == xsd.DeriveExtension {
			b.finishExtensionParticle(ct, bct)
		}
		p := b.pendingAttrs[ct]
		if p == nil {
			return // builtin (xs:anyType) or already-complete component
		}
		var baseUses []*xsd.AttributeUse
		var baseWC *xsd.Wildcard
		if bct != nil {
			baseUses = bct.AttributeUses
			baseWC = bct.AttributeWildcard
		}
		prohibited := p.prohibited
		if !p.override {
			prohibited = nil
		}
		ct.AttributeUses = b.mergeBaseAttrUses(p.own, baseUses, prohibited, p.override, p.pos)
		ct.AttributeWildcard = p.wc
		if p.wc == nil && p.wcFallback {
			// Wildcard union (cos-aw-union) is deferred; the base's wildcard
			// stands in when the type declares none.
			ct.AttributeWildcard = baseWC
		}
		if ct.DerivationMethod == xsd.DeriveRestriction && bct != nil && p.wc != nil {
			b.checkAttrWildcardRestriction(ct, p.wc, baseWC)
		}
		b.applyDefaultAttributes(ct, p.node, p.doc)
		b.checkAttrUses(ct)
	}
	for _, t := range b.types {
		if ct, ok := t.(*xsd.ComplexType); ok {
			finish(ct)
		}
	}
	// Element Declarations Consistent runs after extension particles are
	// assembled, so an extension that reintroduces a base element name with a
	// different type is caught. subMembers maps each head element to the
	// declarations substitutable for it (one hop); the walk closes it
	// transitively for the "implicitly contains" rule.
	subMembers := map[*xsd.ElementDecl][]*xsd.ElementDecl{}
	for _, e := range b.elements {
		for _, head := range e.SubstitutionGroups {
			subMembers[head] = append(subMembers[head], e)
		}
	}
	// accepted returns the set of expanded names an element particle for e can
	// match: e itself plus every declaration transitively substitutable for it.
	acceptedCache := map[*xsd.ElementDecl]map[xsd.QName]bool{}
	accepted := func(e *xsd.ElementDecl) map[xsd.QName]bool {
		if s, ok := acceptedCache[e]; ok {
			return s
		}
		s := map[xsd.QName]bool{}
		var add func(d *xsd.ElementDecl)
		add = func(d *xsd.ElementDecl) {
			if s[d.Name] {
				return
			}
			s[d.Name] = true
			for _, m := range subMembers[d] {
				add(m)
			}
		}
		add(e)
		acceptedCache[e] = s
		return s
	}
	for _, t := range b.types {
		if ct, ok := t.(*xsd.ComplexType); ok {
			b.checkElementConsistent(ct, subMembers)
			b.checkAllUPA(ct, accepted)
			b.checkSeqChoiceUPA(ct, accepted)
			b.checkParticleRestrict(ct, accepted)
			b.checkOpenContentRestrict(ct)
		}
	}
	b.checkSubstitutionCycles()
}

// checkAllUPA enforces Unique Particle Attribution (cos-nonambig) within
// <all> model groups. Because every particle of an all group matches an
// independent run of children regardless of order, UPA there reduces to a
// pairwise test: no two element particles may accept a common element (by
// name or substitution group), and no two wildcard particles may have
// overlapping namespace constraints. Sequence/choice UPA is handled by
// checkSeqChoiceUPA (a position automaton).
func (b *builder) checkAllUPA(ct *xsd.ComplexType, accepted func(*xsd.ElementDecl) map[xsd.QName]bool) {
	ec, ok := ct.Content.(*xsd.ElementContent)
	if !ok || ec.Particle == nil {
		return
	}
	seen := map[*xsd.ModelGroup]bool{}
	var walk func(mg *xsd.ModelGroup)
	walk = func(mg *xsd.ModelGroup) {
		if mg == nil || seen[mg] {
			return
		}
		seen[mg] = true
		if mg.Compositor == xsd.CompositorAll {
			b.checkAllGroupParticles(ct, mg, accepted)
		}
		for _, p := range mg.Particles {
			switch term := p.Term.(type) {
			case *xsd.ModelGroup:
				walk(term)
			case *xsd.GroupRef:
				if term.Ref != nil {
					walk(term.Ref.Group)
				}
			}
		}
	}
	if mg, ok := ec.Particle.Term.(*xsd.ModelGroup); ok {
		walk(mg)
	} else if gr, ok := ec.Particle.Term.(*xsd.GroupRef); ok && gr.Ref != nil {
		walk(gr.Ref.Group)
	}
}

func (b *builder) checkAllGroupParticles(ct *xsd.ComplexType, all *xsd.ModelGroup, accepted func(*xsd.ElementDecl) map[xsd.QName]bool) {
	type elemP struct {
		decl  *xsd.ElementDecl
		names map[xsd.QName]bool
		pos   xsd.Pos
	}
	var elems []elemP
	var wilds []*xsd.Wildcard
	for _, p := range all.Particles {
		switch term := p.Term.(type) {
		case *xsd.ElementDecl:
			names := accepted(term)
			for _, prev := range elems {
				if namesOverlap(prev.names, names) {
					// spec: cos-nonambig — XSD 1.1 Part 1 §3.8.6.4 (xmlschema11-1.md#cos-nonambig)
					b.errf(xsd.SpecCosNonambig, p.Pos, "elements %s and %s compete in the <all> group of %s (Unique Particle Attribution)", prev.decl.Name, term.Name, describeCT(ct))
				}
			}
			elems = append(elems, elemP{decl: term, names: names, pos: p.Pos})
		case *xsd.Wildcard:
			for _, prev := range wilds {
				if wildcardsOverlap(prev, term) {
					b.errf(xsd.SpecCosNonambig, p.Pos, "two wildcards compete in the <all> group of %s (Unique Particle Attribution)", describeCT(ct))
				}
			}
			wilds = append(wilds, term)
		}
	}
}

func namesOverlap(a, b map[xsd.QName]bool) bool {
	if len(a) > len(b) {
		a, b = b, a
	}
	for n := range a {
		if b[n] {
			return true
		}
	}
	return false
}

// wildcardsOverlap reports whether two wildcards can match a common element.
// It returns false (no competition) when either carries notQName so the check
// never produces a false positive from the disallowed-names subtraction.
func wildcardsOverlap(a, b *xsd.Wildcard) bool {
	if len(a.NotQName) > 0 || len(b.NotQName) > 0 {
		return false
	}
	if a.Mode == xsd.NSConstraintAny || b.Mode == xsd.NSConstraintAny {
		return true
	}
	switch {
	case a.Mode == xsd.NSConstraintNot && b.Mode == xsd.NSConstraintNot:
		// Both are "any namespace except a finite set"; their intersection
		// (the complement of the union) is always non-empty.
		return true
	case a.Mode == xsd.NSConstraintEnumeration && b.Mode == xsd.NSConstraintEnumeration:
		return namespacesIntersect(a.Namespaces, b.Namespaces)
	default:
		// One enumeration, one not: overlap iff the enumeration lists a
		// namespace the "not" set does not exclude.
		enum, not := a, b
		if a.Mode == xsd.NSConstraintNot {
			enum, not = b, a
		}
		for _, ns := range enum.Namespaces {
			if !slices.Contains(not.Namespaces, ns) {
				return true
			}
		}
		return false
	}
}

func namespacesIntersect(a, b []string) bool {
	for _, x := range a {
		if slices.Contains(b, x) {
			return true
		}
	}
	return false
}

// checkSubstitutionCycles enforces e-props-correct.6: it must not be possible
// to return to an element declaration by repeatedly following its
// {substitution group affiliation}. Each element on a cycle is reported once.
func (b *builder) checkSubstitutionCycles() {
	reported := map[*xsd.ElementDecl]bool{}
	for _, e := range b.elements {
		if len(e.SubstitutionGroups) == 0 || reported[e] {
			continue
		}
		// Walk the affiliation graph from e; if e is reachable from itself
		// (a path of length ≥ 1), e sits on a cycle.
		seen := map[*xsd.ElementDecl]bool{}
		var reaches func(cur *xsd.ElementDecl) bool
		reaches = func(cur *xsd.ElementDecl) bool {
			for _, head := range cur.SubstitutionGroups {
				if head == e {
					return true
				}
				if !seen[head] {
					seen[head] = true
					if reaches(head) {
						return true
					}
				}
			}
			return false
		}
		if reaches(e) {
			// spec: e-props-correct.6 — XSD 1.1 Part 1 §3.3.6 (xmlschema11-1.md#e-props-correct)
			b.errf(xsd.SpecEPropsCorrect, e.Pos, "element %s is part of a circular substitution group", e.Name)
			reported[e] = true
		}
	}
}

// checkElementConsistent enforces Element Declarations Consistent
// (cos-element-consistent): within one content model, all element
// declarations sharing an expanded name — directly, through nested groups, or
// implicitly through substitution groups — must have the same top-level type
// definition.
func (b *builder) checkElementConsistent(ct *xsd.ComplexType, subMembers map[*xsd.ElementDecl][]*xsd.ElementDecl) {
	ec, ok := ct.Content.(*xsd.ElementContent)
	if !ok || ec.Particle == nil {
		return
	}
	byName := map[xsd.QName][]*xsd.ElementDecl{}
	seenGroups := map[*xsd.ModelGroup]bool{}
	seenElems := map[*xsd.ElementDecl]bool{}
	// addElem records e and, transitively, every declaration substitutable
	// for it (implicitly contained per the substitution-group rule).
	var addElem func(e *xsd.ElementDecl)
	addElem = func(e *xsd.ElementDecl) {
		if seenElems[e] {
			return
		}
		seenElems[e] = true
		byName[e.Name] = append(byName[e.Name], e)
		for _, m := range subMembers[e] {
			addElem(m)
		}
	}
	var walk func(p *xsd.Particle)
	walk = func(p *xsd.Particle) {
		if p == nil {
			return
		}
		switch term := p.Term.(type) {
		case *xsd.ElementDecl:
			addElem(term)
		case *xsd.ModelGroup:
			if seenGroups[term] {
				return
			}
			seenGroups[term] = true
			for _, c := range term.Particles {
				walk(c)
			}
		case *xsd.GroupRef:
			if term.Ref != nil && term.Ref.Group != nil && !seenGroups[term.Ref.Group] {
				seenGroups[term.Ref.Group] = true
				for _, c := range term.Ref.Group.Particles {
					walk(c)
				}
			}
		}
	}
	walk(ec.Particle)

	for name, decls := range byName {
		if len(decls) < 2 {
			continue
		}
		ref := decls[0].Type
		for _, d := range decls {
			if d.Type == ref {
				continue
			}
			rq, dq := typeNameOf(ref), typeNameOf(d.Type)
			if rq.IsZero() || dq.IsZero() || rq != dq {
				// spec: cos-element-consistent — XSD 1.1 Part 1 §3.8.6
				b.errf(xsd.SpecCosElementConsistent, d.Pos, "element %s appears more than once in the content model of %s with differing types", name, describeCT(ct))
				break
			}
		}
	}
}

// typeNameOf returns a type's name, or the zero QName for anonymous types.
func typeNameOf(t xsd.Type) xsd.QName {
	if t == nil {
		return xsd.QName{}
	}
	return t.TypeName()
}

// finishExtensionParticle combines an extension's effective particle: the
// base type's particle followed by the extension's own.
func (b *builder) finishExtensionParticle(ct, bct *xsd.ComplexType) {
	ec, ok := ct.Content.(*xsd.ElementContent)
	if !ok {
		return
	}
	switch bc := bct.Content.(type) {
	case *xsd.SimpleContent:
		// Only reachable when the base was mid-build during construction
		// (buildElementOnlyContent handles completed simple-content bases).
		if ec.Particle == nil && !ec.Mixed {
			ct.Content = &xsd.SimpleContent{Type: bc.Type}
		} else {
			// spec: cos-ct-extends.1.4.2 — a complex extension of a
			// simple-content type cannot add element content.
			b.errf(xsd.SpecCosCTExtends, ct.Pos, "cannot extend %s with element content: its content is simple", bct.Name)
		}
	case *xsd.ElementContent:
		// spec: §3.4.2.3.3 (mapping rules) — when the extension contributes no
		// "effective content" (no own particle and not mixed) the derived
		// {content type} is copied wholesale from the base, including its
		// mixed/element-only variety; no consistency check applies.
		if ec.Particle == nil && !ec.Mixed {
			ec.Mixed = bc.Mixed
			ec.Particle = bc.Particle
			return
		}
		// The extension contributes content. When the base {content type} is
		// non-empty, the two must share the same mixed/element-only variety.
		// spec: cos-ct-extends.1.4.3.2.2.1 — XSD 1.1 Part 1 §3.4.6.2
		if (bc.Particle != nil || bc.Mixed) && ec.Mixed != bc.Mixed {
			want := "element-only"
			if bc.Mixed {
				want = "mixed"
			}
			b.errf(xsd.SpecCosCTExtends, ct.Pos, "extension of %s must have %s content to match the base", describeCT(bct), want)
		}
		if bc.Particle == nil {
			return
		}
		if ec.Particle == nil {
			ec.Particle = bc.Particle
		} else if baseAll, ownAll := allGroupTerm(bc.Particle), allGroupTerm(ec.Particle); baseAll != nil && ownAll != nil {
			// spec: §3.4.2.3.3 clause 4.2.3.2 — extending an <all> group with an
			// <all> group yields a single <all> group whose particles are the
			// base's followed by the extension's (NOT a sequence). This keeps
			// the merged particles in one all group so Unique Particle
			// Attribution (checkAllUPA) sees a re-added element as a conflict.
			// spec: cos-ct-extends — XSD 1.1 Part 1 §3.4.6.2 clause 1.4.3.2.2.2
			// invokes cos-particle-extend.3.1: the extended <all> particle's
			// {min occurs} must equal the base's.
			if ec.Particle.MinOccurs != bc.Particle.MinOccurs {
				b.errf(xsd.SpecCosCTExtends, ec.Particle.Pos, "extending the <all> group of %s requires the same minOccurs as the base", describeCT(bct))
			}
			merged := make([]*xsd.Particle, 0, len(baseAll.Particles)+len(ownAll.Particles))
			merged = append(merged, baseAll.Particles...)
			merged = append(merged, ownAll.Particles...)
			ec.Particle = &xsd.Particle{
				MinOccurs: ec.Particle.MinOccurs, MaxOccurs: 1,
				Term: &xsd.ModelGroup{
					Compositor: xsd.CompositorAll,
					Particles:  merged,
					Pos:        ec.Particle.Pos,
				},
				Pos: ec.Particle.Pos,
			}
		} else {
			// spec: cos-all-limited.1 — clause 4.2.3.3 wraps base + extension in
			// a sequence; if either side is an <all> group (and they did not both
			// merge above), that <all> would sit illegally inside a sequence.
			if allGroupTerm(bc.Particle) != nil || allGroupTerm(ec.Particle) != nil {
				b.errf(xsd.SpecCosAllLimited, ec.Particle.Pos, "cannot extend %s: an <all> group may not be combined with a sequence by extension", describeCT(bct))
			}
			ec.Particle = &xsd.Particle{
				MinOccurs: 1, MaxOccurs: 1,
				Term: &xsd.ModelGroup{
					Compositor: xsd.CompositorSequence,
					Particles:  []*xsd.Particle{bc.Particle, ec.Particle},
					Pos:        ec.Particle.Pos,
				},
				Pos: ec.Particle.Pos,
			}
		}
	}
}

// allGroupTerm returns the model group of p's term if it is an <all> group
// (directly or via a group reference), else nil.
func allGroupTerm(p *xsd.Particle) *xsd.ModelGroup {
	if p == nil {
		return nil
	}
	switch t := p.Term.(type) {
	case *xsd.ModelGroup:
		if t.Compositor == xsd.CompositorAll {
			return t
		}
	case *xsd.GroupRef:
		if t.Ref != nil && t.Ref.Group != nil && t.Ref.Group.Compositor == xsd.CompositorAll {
			return t.Ref.Group
		}
	}
	return nil
}

// checkTypeCycles detects cyclic complex-type derivation chains. Cycles are
// reported once and broken (base reset to xs:anyType) so the model stays
// walkable. Simple-type cycles were already caught eagerly during building.
func (b *builder) checkTypeCycles(s *xsd.Schema) {
	for _, t := range s.Types {
		start, ok := t.(*xsd.ComplexType)
		if !ok {
			continue
		}
		seen := map[xsd.Type]bool{}
		for cur := xsd.Type(start); cur != nil; {
			if seen[cur] {
				ct := cur.(*xsd.ComplexType) // simple types cannot be in the cycle
				// spec: ct-props-correct.3 — XSD 1.1 Part 1 §3.4.6: cyclic
				// complex type definitions are disallowed.
				b.errf(xsd.SpecCTPropsCorrect, ct.TypePos(), "complex type %s is part of a cyclic definition", ct.TypeName())
				ct.BaseType = builtin.AnyType
				break
			}
			seen[cur] = true
			next := cur.Base()
			if next == cur {
				break
			}
			cur = next
		}
	}
}
