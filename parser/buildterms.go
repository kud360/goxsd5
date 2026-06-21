package parser

// Declarations and terms: elements, attributes, particles, model groups,
// wildcards, identity constraints, notations.

import (
	"strings"

	"github.com/kud360/goxsd5/builtin"
	"github.com/kud360/goxsd5/parser/xmltree"
	"github.com/kud360/goxsd5/xsd"
)

// localQName computes the expanded name of a local declaration from its
// form/targetNamespace attributes and the schema-level form default.
func localQName(n *xmltree.Node, doc *schemaDoc, name string, formDefault xsd.Form) xsd.QName {
	if tns, ok := n.Attr("targetNamespace"); ok {
		return xsd.QName{Namespace: tns, Local: name}
	}
	form := formDefault
	if v, ok := n.Attr("form"); ok {
		if strings.TrimSpace(v) == "qualified" {
			form = xsd.FormQualified
		} else {
			form = xsd.FormUnqualified
		}
	}
	if form == xsd.FormQualified {
		return xsd.QName{Namespace: doc.targetNamespace, Local: name}
	}
	return xsd.QName{Local: name}
}

// --- elements ------------------------------------------------------

func (b *builder) buildElementDecl(n *xmltree.Node, doc *schemaDoc, global bool) *xsd.ElementDecl {
	if e, ok := b.elements[n]; ok {
		return e
	}
	e := &xsd.ElementDecl{
		Pos:        n.Pos,
		Global:     global,
		Nillable:   boolAttr(n, "nillable", false),
		Abstract:   boolAttr(n, "abstract", false),
		Block:      derivAttr(n, "block", eltBlockSet, doc.blockDefault),
		Annotation: annotationOf(n, doc),
		Extensions: extensionsOf(n),
	}
	b.elements[n] = e

	name, _ := n.Attr("name")
	if global {
		e.Name = xsd.QName{Namespace: doc.targetNamespace, Local: name}
		e.Final = derivAttr(n, "final", eltFinalSet, doc.finalDefault)
	} else {
		e.Name = localQName(n, doc, name, doc.elementFormDefault)
	}
	if e.Name.Namespace != "" {
		e.Form = xsd.FormQualified
	}

	if v, ok := n.Attr("default"); ok {
		e.Default = &v
	}
	if v, ok := n.Attr("fixed"); ok {
		e.Fixed = &v
	}

	// Substitution groups first: an element with no type takes its first
	// head's type.
	if global {
		b.buildSubstitutionGroups(e, n, doc)
	}

	e.Type = b.elementType(e, n, doc)
	// Type-dependent validation (NOTATION enumeration, value constraints,
	// substitution-group derivation) reads the element's type internals and so
	// is deferred to checkElementDecls, a post-pass that runs once every type
	// is fully built (see the type shell/finish split in buildComplexType).

	b.buildElementChildren(e, n, doc)
	return e
}

// buildSubstitutionGroups resolves a global element's substitutionGroup heads
// onto e.SubstitutionGroups, skipping tokens whose refs are unresolved (those
// are reported by pass 1).
func (b *builder) buildSubstitutionGroups(e *xsd.ElementDecl, n *xmltree.Node, doc *schemaDoc) {
	v, ok := n.Attr("substitutionGroup")
	if !ok {
		return
	}
	for _, tok := range strings.Fields(v) {
		q, err := n.ResolveQName(tok)
		if err != nil {
			continue // reported by pass 1
		}
		d := b.lookupRef(spaceElement, chameleonQName(q, doc), n.Pos, doc)
		if d == nil {
			continue
		}
		e.SubstitutionGroups = append(e.SubstitutionGroups, b.buildElementDecl(d.node, d.doc, true))
	}
}

// elementType determines an element's type: an anonymous type if present,
// else the @type reference, else the first substitution-group head's type,
// else the ur-type.
func (b *builder) elementType(e *xsd.ElementDecl, n *xmltree.Node, doc *schemaDoc) xsd.Type {
	if t := b.buildAnonType(n, doc, nil); t != nil {
		return t
	}
	if q, ok := qnameAttr(n, doc, "type"); ok {
		return b.resolveType(q, n.Pos, doc)
	}
	if len(e.SubstitutionGroups) > 0 && e.SubstitutionGroups[0].Type != nil {
		return e.SubstitutionGroups[0].Type
	}
	return builtin.AnyType
}

// buildElementChildren appends the element's identity constraints and type
// alternatives from its xsd: child elements.
func (b *builder) buildElementChildren(e *xsd.ElementDecl, n *xmltree.Node, doc *schemaDoc) {
	for _, c := range xsdElems(n, doc) {
		switch c.Name.Local {
		case "unique", "key", "keyref":
			if ic := b.buildIC(c, doc); ic != nil {
				e.IdentityConstraints = append(e.IdentityConstraints, ic)
			}
		case "alternative":
			alt := &xsd.TypeAlternative{Pos: c.Pos}
			alt.Test, _ = c.Attr("test")
			b.checkXPathTest(alt.Test, c, doc, xsd.SpecSrcTA)
			alt.Type = b.buildAnonType(c, doc, nil)
			if alt.Type == nil {
				if q, ok := qnameAttr(c, doc, "type"); ok {
					alt.Type = b.resolveType(q, c.Pos, doc)
				}
			}
			e.TypeAlternatives = append(e.TypeAlternatives, alt)
		}
	}
}

// checkElementDecls runs the per-element validation that reads each element's
// type internals (and so must run after every type is fully built). It walks
// the built element declarations and reports NOTATION-enumeration, value
// constraint, and substitution-group derivation violations. The node key
// supplies positions and the namespace context for value parsing.
func (b *builder) checkElementDecls() {
	for n, e := range b.elements {
		b.checkNotationEnum(e.Type, n.Pos)
		b.checkElementValueConstraint(e, n)
		b.checkSubstitutionGroupExclusions(e, n)
	}
}

// checkElementValueConstraint validates an element's default/fixed value
// against its type. spec: e-props-correct.2 / cos-valid-default — XSD 1.1
// Part 1 §3.3.6.
func (b *builder) checkElementValueConstraint(e *xsd.ElementDecl, n *xmltree.Node) {
	vc := valueConstraint(e.Default, e.Fixed)
	if vc == nil {
		return
	}
	if st := contentSimpleType(e.Type); st != nil {
		// Note: XSD 1.1 dropped the 1.0 rule (old e-props-correct.5)
		// forbidding value constraints on ID-derived element types; it is
		// now only a non-normative "should avoid" note, so we no longer
		// reject it.
		if _, err := st.ParseValue(*vc, nsContext{n}); err != nil {
			b.errf(xsd.SpecCosValidDefault, n.Pos, "default/fixed value %q is not valid for the type of element %s: %v", *vc, e.Name, err)
		}
		return
	}
	ct, ok := e.Type.(*xsd.ComplexType)
	if !ok {
		return
	}
	ec, ok := ct.Content.(*xsd.ElementContent)
	if !ok {
		return
	}
	if !ec.Mixed {
		// spec: cos-valid-default.2.1 — element-only content admits
		// no value constraint.
		b.errf(xsd.SpecCosValidDefault, n.Pos, "element %s has element-only content and must not have a default or fixed value", e.Name)
		return
	}
	if !particleEmptiable(ec.Particle) {
		// spec: cos-valid-default.2.2 — mixed content with a value
		// constraint requires an emptiable particle.
		b.errf(xsd.SpecCosValidDefault, n.Pos, "element %s has mixed content with a non-emptiable particle and must not have a default or fixed value", e.Name)
	}
}

// checkSubstitutionGroupExclusions reports an element whose type is derived
// from a substitution-group head's type by a method the head's {final}
// excludes. Unreachable chains are left to the deferred derivation-ok checks
// rather than guessed at.
func (b *builder) checkSubstitutionGroupExclusions(e *xsd.ElementDecl, n *xmltree.Node) {
	for _, head := range e.SubstitutionGroups {
		if head.Type == nil || e.Type == nil {
			continue
		}
		// spec: e-props-correct.4 — XSD 1.1 Part 1 §3.3.6 — the member's type
		// must be validly derived from the head's type, and the derivation
		// methods used must not be among the head's {substitution group
		// exclusions} ({final}).
		methods, ok := derivationMethods(e.Type, head.Type)
		if !ok {
			b.errf(xsd.SpecEPropsCorrect, n.Pos, "element %s cannot substitute for %s: its type is not validly derived from the head's type", e.Name, head.Name)
		} else if methods&head.Final != 0 {
			b.errf(xsd.SpecEPropsCorrect, n.Pos, "element %s cannot substitute for %s: the head excludes this derivation", e.Name, head.Name)
		}
	}
}

// valueConstraint returns default or fixed, whichever is present.
func valueConstraint(def, fixed *string) *string {
	if def != nil {
		return def
	}
	return fixed
}

// contentSimpleType returns the simple type governing character content:
// the type itself, or a complex type's simple content.
func contentSimpleType(t xsd.Type) *xsd.SimpleType {
	switch t := t.(type) {
	case *xsd.SimpleType:
		return t
	case *xsd.ComplexType:
		if sc, ok := t.Content.(*xsd.SimpleContent); ok {
			return sc.Type
		}
	}
	return nil
}

// derivationMethods walks t's base-type chain to anc, collecting the
// derivation methods used. ok is false when anc is not an ancestor. The
// visited guard terminates on not-yet-broken cyclic chains (the cycle
// itself is reported by checkTypeCycles).
func derivationMethods(t, anc xsd.Type) (xsd.DerivationSet, bool) {
	var methods xsd.DerivationSet
	seen := map[xsd.Type]bool{}
	for cur := t; cur != nil && !seen[cur]; {
		seen[cur] = true
		if cur == anc {
			return methods, true
		}
		switch c := cur.(type) {
		case *xsd.ComplexType:
			methods |= xsd.DerivationSet(c.DerivationMethod)
		case *xsd.SimpleType:
			switch c.Variety {
			case xsd.VarietyList:
				methods |= xsd.DerivationSet(xsd.DeriveList)
			case xsd.VarietyUnion:
				methods |= xsd.DerivationSet(xsd.DeriveUnion)
			default:
				methods |= xsd.DerivationSet(xsd.DeriveRestriction)
			}
		}
		base := cur.Base()
		if base == cur {
			break
		}
		cur = base
	}
	return 0, false
}

// checkNotationEnum enforces that NOTATION-derived types carry an
// enumeration before they are used for validation.
func (b *builder) checkNotationEnum(t xsd.Type, p xsd.Pos) {
	st := contentSimpleType(t)
	if st == nil {
		return
	}
	prim := st.PrimitiveType()
	if prim == nil {
		return
	}
	if prim.Name.Local == "NOTATION" && !st.EffectiveFacets().HasEnumeration() {
		// spec: enumeration-required-notation — XSD 1.1 Part 2 §3.3.19
		b.errf(xsd.SpecEnumNotation, p, "a NOTATION-derived type must constrain its values with an enumeration facet")
	}
}

// --- attributes ----------------------------------------------------

func (b *builder) buildAttributeDecl(n *xmltree.Node, doc *schemaDoc, global bool) *xsd.AttributeDecl {
	if a, ok := b.attributes[n]; ok {
		return a
	}
	a := &xsd.AttributeDecl{
		Pos:         n.Pos,
		Global:      global,
		Inheritable: boolAttr(n, "inheritable", false),
		Annotation:  annotationOf(n, doc),
		Extensions:  extensionsOf(n),
	}
	b.attributes[n] = a

	name, _ := n.Attr("name")
	if global {
		a.Name = xsd.QName{Namespace: doc.targetNamespace, Local: name}
		if a.Name.Namespace == xsd.XSINS {
			// spec: no-xsi — XSD 1.1 Part 1 §3.2.6.4: only the four built-in
			// xsi: attribute declarations may target the XML Schema instance
			// namespace. A top-level attribute inherits its {target namespace}
			// from the schema, so a schema whose targetNamespace is the xsi
			// namespace declares one illegally. (The explicit-targetNamespace
			// form on a local attribute is caught structurally in pass 1.)
			b.errf(xsd.SpecNoXsi, n.Pos, "attribute %s must not target the XML Schema instance namespace", name)
		}
	} else {
		a.Name = localQName(n, doc, name, doc.attributeFormDefault)
	}
	if a.Name.Namespace != "" {
		a.Form = xsd.FormQualified
	}

	if inline := firstChild(n, doc, "simpleType"); inline != nil {
		a.Type, _ = b.buildAnonType(n, doc, builtin.AnySimpleType).(*xsd.SimpleType)
	} else if q, ok := qnameAttr(n, doc, "type"); ok {
		// spec: a-props-correct.2 — the type of an attribute declaration
		// must be a simple type definition.
		a.Type = b.resolveSimpleType(q, n.Pos, doc, xsd.SpecAPropsCorrect)
	} else {
		a.Type = builtin.AnySimpleType
	}
	b.checkNotationEnum(a.Type, n.Pos)

	if v, ok := n.Attr("default"); ok {
		a.Default = &v
	}
	if v, ok := n.Attr("fixed"); ok {
		a.Fixed = &v
	}
	if vc := valueConstraint(a.Default, a.Fixed); vc != nil {
		// Note: XSD 1.1 dropped the 1.0 rule (old a-props-correct.3) forbidding
		// value constraints on ID-derived attribute types; only the value
		// constraint's validity (a-props-correct.2) is still enforced.
		if _, err := a.Type.ParseValue(*vc, nsContext{n}); err != nil {
			// spec: a-props-correct.2 (value constraint valid)
			b.errf(xsd.SpecAPropsCorrect, n.Pos, "default/fixed value %q is not valid for the type of attribute %s: %v", *vc, a.Name, err)
		}
	}
	return a
}

// buildAttrUses walks the (attribute | attributeGroup)*, anyAttribute?
// children of parent and returns the attribute uses, the local wildcard,
// and the set of prohibited attribute names.
func (b *builder) buildAttrUses(parent *xmltree.Node, doc *schemaDoc) (uses []*xsd.AttributeUse, wc *xsd.Wildcard, prohibited map[xsd.QName]bool) {
	prohibited = map[xsd.QName]bool{}
	for _, c := range xsdElems(parent, doc) {
		switch c.Name.Local {
		case "attribute":
			use, _ := c.Attr("use")
			if strings.TrimSpace(use) == "prohibited" {
				if q := b.attrUseName(c, doc); !q.IsZero() {
					prohibited[q] = true
				}
				continue
			}
			u := b.buildAttrUse(c, doc)
			if u != nil {
				uses = append(uses, u)
			}
		case "attributeGroup":
			q, ok := qnameAttr(c, doc, "ref")
			if !ok {
				continue
			}
			d := b.lookupRef(spaceAttrGroup, q, c.Pos, doc)
			if d == nil {
				continue
			}
			g := b.buildAttributeGroup(d)
			if g == nil {
				continue // cyclic, already reported
			}
			uses = append(uses, g.Uses...)
			// spec: cos-aw-intersect — XSD 1.1 Part 1 §3.10.6.4: the effective
			// attribute wildcard is the intersection of all group wildcards.
			wc = wildcardIntersect(wc, g.Wildcard)
		case "anyAttribute":
			// The type's own anyAttribute is also intersected with any group
			// wildcards already accumulated (§3.8.4.2).
			wc = wildcardIntersect(wc, b.buildWildcard(c, doc))
		}
	}
	return uses, wc, prohibited
}

// attrUseName computes the expanded name a use refers to (its ref target
// or its local declaration's name).
func (b *builder) attrUseName(c *xmltree.Node, doc *schemaDoc) xsd.QName {
	if q, ok := qnameAttr(c, doc, "ref"); ok {
		return q
	}
	if name, ok := c.Attr("name"); ok {
		return localQName(c, doc, name, doc.attributeFormDefault)
	}
	return xsd.QName{}
}

func (b *builder) buildAttrUse(c *xmltree.Node, doc *schemaDoc) *xsd.AttributeUse {
	use, _ := c.Attr("use")
	u := &xsd.AttributeUse{Required: strings.TrimSpace(use) == "required", Pos: c.Pos}
	if q, ok := qnameAttr(c, doc, "ref"); ok {
		d := b.lookupRef(spaceAttribute, q, c.Pos, doc)
		if d == nil {
			return nil
		}
		if d.builtinAttr != nil {
			u.Decl = d.builtinAttr
		} else {
			u.Decl = b.buildAttributeDecl(d.node, d.doc, true)
		}
		// {inheritable} of the use: declared on the use if present (it overrides
		// the declaration's), else the declaration's value (XSD 1.1 §3.5.2).
		u.Inheritable = boolAttr(c, "inheritable", u.Decl.Inheritable)
		if v, ok := c.Attr("default"); ok {
			u.Default = &v
		}
		if v, ok := c.Attr("fixed"); ok {
			u.Fixed = &v
		}
		// spec: au-props-correct.2 — XSD 1.1 Part 1 §3.5.6: a use of a
		// declaration with a fixed value must not weaken or change it.
		if u.Decl.Fixed != nil {
			if u.Default != nil {
				b.errf(xsd.SpecAUPropsCorrect, c.Pos, "attribute %s is fixed in its declaration; the use must not declare a default", q)
			}
			if u.Fixed != nil && *u.Fixed != *u.Decl.Fixed {
				b.errf(xsd.SpecAUPropsCorrect, c.Pos, "attribute %s is fixed to %q in its declaration; the use must not change it", q, *u.Decl.Fixed)
			}
		}
		return u
	}
	u.Decl = b.buildAttributeDecl(c, doc, false)
	u.Inheritable = u.Decl.Inheritable
	return u
}

func (b *builder) buildAttributeGroup(d *decl) *xsd.AttributeGroup {
	if g, ok := b.attrGroups[d.node]; ok {
		return g
	}
	if b.building[d.node] {
		// spec: src-attribute_group.3 — circular attribute groups.
		b.errf(xsd.SpecSrcAttributeGroup, d.pos, "attribute group %s is part of a cycle", d.name)
		return nil
	}
	b.building[d.node] = true
	defer delete(b.building, d.node)

	g := &xsd.AttributeGroup{
		Name:       d.name,
		Pos:        d.pos,
		Annotation: annotationOf(d.node, d.doc),
		Extensions: extensionsOf(d.node),
	}
	g.Uses, g.Wildcard, _ = b.buildAttrUses(d.node, d.doc)
	b.attrGroups[d.node] = g
	b.checkAttrGroupDefDups(d)
	return g
}

// checkAttrGroupDefDups enforces ag-props-correct clause 2 at the attribute
// group definition level: no two <attribute> children of the definition share
// an expanded name. A *referenced* group's collisions are already caught after
// merge into a complex type (ct-props-correct.4); this covers a never-
// referenced definition, whose direct attributes are otherwise never compared.
// Nested <attributeGroup ref>s are checked where they resolve, so only the
// direct <attribute> children are considered here.
// spec: ag-props-correct — XSD 1.1 Part 1 §3.6.6 (clause 2)
func (b *builder) checkAttrGroupDefDups(d *decl) {
	seen := map[xsd.QName]bool{}
	for _, c := range xsdElems(d.node, d.doc) {
		if c.Name.Local != "attribute" {
			continue
		}
		if use, _ := c.Attr("use"); strings.TrimSpace(use) == "prohibited" {
			continue
		}
		q := b.attrUseName(c, d.doc)
		if q.IsZero() {
			continue
		}
		if seen[q] {
			b.errf(xsd.SpecAGPropsCorrect, c.Pos, "attribute %s is declared twice in attribute group %s", q, d.name)
			continue
		}
		seen[q] = true
	}
}

// --- particles and model groups -------------------------------------

// occurs reads the occurrence range of a particle node.
func occurs(n *xmltree.Node) (min, max int) {
	min, max = 1, 1
	if v, ok := n.Attr("minOccurs"); ok {
		if i, err := parseNonNegInt(v); err == nil {
			min = i
		}
	}
	if v, ok := n.Attr("maxOccurs"); ok {
		if i, err := parseMaxOccurs(v); err == nil {
			max = i
		}
	}
	return min, max
}

// buildParticle builds the particle for one term node inside a model
// group or complex type. It returns nil for maxOccurs=0 (the mapping says
// such particles correspond to no component) and on unrecoverable errors.
func (b *builder) buildParticle(c *xmltree.Node, doc *schemaDoc) *xsd.Particle {
	min, max := occurs(c)
	if max == 0 {
		return nil
	}
	p := &xsd.Particle{MinOccurs: min, MaxOccurs: max, Pos: c.Pos}
	switch c.Name.Local {
	case "element":
		if q, ok := qnameAttr(c, doc, "ref"); ok {
			d := b.lookupRef(spaceElement, q, c.Pos, doc)
			if d == nil {
				return nil
			}
			p.Term = b.buildElementDecl(d.node, d.doc, true)
		} else {
			p.Term = b.buildElementDecl(c, doc, false)
		}
	case "group":
		q, ok := qnameAttr(c, doc, "ref")
		if !ok {
			return nil
		}
		d := b.lookupRef(spaceGroup, q, c.Pos, doc)
		if d == nil {
			return nil
		}
		g := b.buildGroup(d)
		if g == nil {
			return nil // cyclic, already reported
		}
		p.Term = &xsd.GroupRef{Ref: g, Pos: c.Pos}
	case "all", "choice", "sequence":
		p.Term = b.buildModelGroup(c, doc)
	case "any":
		p.Term = b.buildWildcard(c, doc)
	default:
		return nil
	}
	return p
}

func (b *builder) buildModelGroup(n *xmltree.Node, doc *schemaDoc) *xsd.ModelGroup {
	mg := &xsd.ModelGroup{Pos: n.Pos}
	switch n.Name.Local {
	case "choice":
		mg.Compositor = xsd.CompositorChoice
	case "all":
		mg.Compositor = xsd.CompositorAll
	default:
		mg.Compositor = xsd.CompositorSequence
	}
	for _, c := range xsdElems(n, doc) {
		switch c.Name.Local {
		case "element", "group", "choice", "sequence", "all", "any":
			if p := b.buildParticle(c, doc); p != nil {
				mg.Particles = append(mg.Particles, p)
			}
		}
	}
	if mg.Compositor == xsd.CompositorAll {
		b.checkAllLimited(mg)
	}
	return mg
}

// checkAllLimited enforces the parts of cos-all-limited (§3.8.6.2) that concern
// model groups nested directly inside an <all>: clause 2 (a nested model group
// must itself be <all>) and clause 1.3 (such a nested <all> may appear only as
// the term of a particle that occurs exactly once).
func (b *builder) checkAllLimited(all *xsd.ModelGroup) {
	for _, p := range all.Particles {
		var inner *xsd.ModelGroup
		switch t := p.Term.(type) {
		case *xsd.ModelGroup:
			inner = t
		case *xsd.GroupRef:
			if t.Ref != nil {
				inner = t.Ref.Group
			}
		}
		if inner == nil {
			continue
		}
		// spec: cos-all-limited.2 — XSD 1.1 Part 1 §3.8.6.2 (xmlschema11-1.md#cos-all-limited)
		if inner.Compositor != xsd.CompositorAll {
			b.errf(xsd.SpecCosAllLimited, p.Pos, "a model group nested in <all> must itself be an <all> group")
			continue
		}
		// spec: cos-all-limited.1.3 — a nested <all> must occur exactly once.
		if p.MinOccurs != 1 || p.MaxOccurs != 1 {
			b.errf(xsd.SpecCosAllLimited, p.Pos, "a nested <all> group must have minOccurs = maxOccurs = 1")
		}
	}
}

func (b *builder) buildGroup(d *decl) *xsd.Group {
	if g, ok := b.groups[d.node]; ok {
		return g
	}
	// Memoize the shell before building the model group, mirroring complex
	// types: a group's content can legally recurse back to itself through an
	// element declaration's type (the element breaks the chain, so this is no
	// model-group cycle). Returning the shell on re-entry keeps that recursion
	// finite and leaves the structure intact; genuine group-structural cycles
	// are reported separately by checkGroupCycles.
	g := &xsd.Group{
		Name:       d.name,
		Pos:        d.pos,
		Annotation: annotationOf(d.node, d.doc),
		Extensions: extensionsOf(d.node),
	}
	b.groups[d.node] = g
	if mg := firstChild(d.node, d.doc, "all", "choice", "sequence"); mg != nil {
		g.Group = b.buildModelGroup(mg, d.doc)
	}
	return g
}

// --- wildcards -------------------------------------------------------

func (b *builder) buildWildcard(n *xmltree.Node, doc *schemaDoc) *xsd.Wildcard {
	w := &xsd.Wildcard{
		Mode:       xsd.NSConstraintAny,
		Pos:        n.Pos,
		Annotation: annotationOf(n, doc),
		Extensions: extensionsOf(n),
	}
	if v, ok := n.Attr("processContents"); ok {
		switch strings.TrimSpace(v) {
		case "lax":
			w.ProcessContents = xsd.ProcessLax
		case "skip":
			w.ProcessContents = xsd.ProcessSkip
		}
	}

	nsList := func(v string) []string {
		var out []string
		for _, tok := range strings.Fields(v) {
			switch tok {
			case "##targetNamespace":
				out = append(out, doc.targetNamespace)
			case "##local":
				out = append(out, "")
			default:
				out = append(out, tok)
			}
		}
		return out
	}

	if v, ok := n.Attr("namespace"); ok {
		switch strings.TrimSpace(v) {
		case "##any":
			w.Mode = xsd.NSConstraintAny
		case "##other":
			w.Mode = xsd.NSConstraintNot
			// ##other excludes the target namespace and absent.
			w.Namespaces = []string{doc.targetNamespace, ""}
			if doc.targetNamespace == "" {
				w.Namespaces = []string{""}
			}
		default:
			w.Mode = xsd.NSConstraintEnumeration
			w.Namespaces = nsList(v)
		}
	} else if v, ok := n.Attr("notNamespace"); ok {
		w.Mode = xsd.NSConstraintNot
		w.Namespaces = nsList(v)
	}

	if v, ok := n.Attr("notQName"); ok {
		for _, tok := range strings.Fields(v) {
			if tok == "##defined" || tok == "##definedSibling" {
				w.NotQName = append(w.NotQName, xsd.QName{Local: tok})
				continue
			}
			if q, err := n.ResolveQName(tok); err == nil {
				q = chameleonQName(q, doc)
				w.NotQName = append(w.NotQName, q)
				// spec: w-props-correct — XSD 1.1 Part 1 §3.10.6 (xmlschema11-1.md#w-props-correct)
				// Rule 4: a name disallowed via notQName must lie in a namespace
				// the namespace constraint already permits; otherwise it is
				// redundant/contradictory.
				if !w.AllowsNamespace(q.Namespace) {
					b.errf(xsd.SpecWPropsCorrect, n.Pos, "notQName %q names a namespace not allowed by the wildcard's namespace constraint", tok)
				}
			}
		}
	}
	return w
}

// --- identity constraints ---------------------------------------------

func (b *builder) buildIC(n *xmltree.Node, doc *schemaDoc) *xsd.IdentityConstraint {
	if ic, ok := b.ics[n]; ok {
		return ic
	}
	if b.building[n] {
		// spec: c-props-correct — a ref chain must terminate.
		b.errf(xsd.SpecICPropsCorrect, n.Pos, "identity constraint reference cycle")
		return nil
	}
	b.building[n] = true
	defer delete(b.building, n)

	category := map[string]xsd.ICCategory{
		"unique": xsd.ICUnique, "key": xsd.ICKey, "keyref": xsd.ICKeyref,
	}[n.Name.Local]

	if q, ok := qnameAttr(n, doc, "ref"); ok {
		d := b.lookupRef(spaceIC, q, n.Pos, doc)
		if d == nil {
			return nil
		}
		target := b.buildIC(d.node, d.doc)
		if target != nil && target.Category != category {
			// spec: src-identity-constraint.5 — the referenced constraint's
			// category must match the referring element.
			b.errf(xsd.SpecSrcIdentity, n.Pos, "identity constraint %s is a %s, not a %s", q, icName(target.Category), n.Name.Local)
		}
		b.ics[n] = target
		return target
	}

	name, _ := n.Attr("name")
	ic := &xsd.IdentityConstraint{
		Name:       xsd.QName{Namespace: doc.targetNamespace, Local: name},
		Pos:        n.Pos,
		Category:   category,
		Annotation: annotationOf(n, doc),
		Extensions: extensionsOf(n),
		// In-scope namespaces at the declaration resolve prefixed name tests in
		// the selector/field XPath subset (§3.11.6). Per §3.13.2 the {namespace
		// bindings} formally come from the host element — <selector> for the
		// selector, each <field> for its field — but we capture them once from
		// the <key>/<unique>/<keyref> node n. That set equals each child's
		// in-scope namespaces unless a prefix is re-declared on the <selector>
		// or <field> element itself, a pathological case absent from the corpus;
		// the IC-node set is otherwise a correct superset. Documented limitation.
		NamespaceBindings: n.NS.InScope(),
	}
	b.ics[n] = ic
	if sel := firstChild(n, doc, "selector"); sel != nil {
		ic.Selector, _ = sel.Attr("xpath")
	}
	for _, c := range xsdElems(n, doc) {
		if c.Name.Local == "field" {
			xp, _ := c.Attr("xpath")
			ic.Fields = append(ic.Fields, xp)
		}
	}
	if category == xsd.ICKeyref {
		if q, ok := qnameAttr(n, doc, "refer"); ok {
			d := b.lookupRef(spaceIC, q, n.Pos, doc)
			if d == nil {
				// reported by lookupRef
			} else if ref := b.buildIC(d.node, d.doc); ref != nil {
				// spec: c-props-correct.2 — XSD 1.1 Part 1 §3.11.6: the
				// referenced constraint is a key or unique with the same
				// number of fields.
				if ref.Category == xsd.ICKeyref {
					b.errf(xsd.SpecICPropsCorrect, n.Pos, "keyref %s must refer to a key or unique, not another keyref", ic.Name)
				} else if len(ref.Fields) != len(ic.Fields) {
					b.errf(xsd.SpecICPropsCorrect, n.Pos, "keyref %s has %d fields but its referenced key %s has %d", ic.Name, len(ic.Fields), ref.Name, len(ref.Fields))
				}
				ic.Refer = ref
			}
		}
	}
	return ic
}

func icName(c xsd.ICCategory) string {
	switch c {
	case xsd.ICKey:
		return "key"
	case xsd.ICKeyref:
		return "keyref"
	}
	return "unique"
}

// --- notations ---------------------------------------------------------

func (b *builder) buildNotation(n *xmltree.Node, doc *schemaDoc) *xsd.Notation {
	if no, ok := b.notations[n]; ok {
		return no
	}
	name, _ := n.Attr("name")
	no := &xsd.Notation{
		Name:       xsd.QName{Namespace: doc.targetNamespace, Local: name},
		Pos:        n.Pos,
		Annotation: annotationOf(n, doc),
		Extensions: extensionsOf(n),
	}
	no.System, _ = n.Attr("system")
	no.Public, _ = n.Attr("public")
	b.notations[n] = no
	return no
}
