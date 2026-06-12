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
	if v, ok := n.Attr("substitutionGroup"); ok && global {
		for _, tok := range strings.Fields(v) {
			q, err := n.ResolveQName(tok)
			if err != nil {
				continue // reported by pass 1
			}
			d := b.registryFor(doc).lookup(spaceElement, q)
			if d == nil {
				// spec: src-resolve — XSD 1.1 Part 1 §3.15.3
				b.errf(xsd.SpecSrcResolve, n.Pos, "substitution group head %s is not declared", q)
				continue
			}
			e.SubstitutionGroups = append(e.SubstitutionGroups, b.buildElementDecl(d.node, d.doc, true))
		}
	}

	e.Type = b.buildAnonType(n, doc, nil)
	if e.Type == nil {
		if q, ok := qnameAttr(n, "type"); ok {
			e.Type = b.resolveType(q, n.Pos, doc)
		} else if len(e.SubstitutionGroups) > 0 && e.SubstitutionGroups[0].Type != nil {
			e.Type = e.SubstitutionGroups[0].Type
		} else {
			e.Type = builtin.AnyType
		}
	}
	b.checkNotationEnum(e.Type, n.Pos)

	// Value constraints.
	// spec: e-props-correct.2 / cos-valid-default — XSD 1.1 Part 1 §3.3.6
	if vc := valueConstraint(e.Default, e.Fixed); vc != nil {
		if st := contentSimpleType(e.Type); st != nil {
			if isIDDerived(st) {
				// spec: e-props-correct.5 — no value constraint for
				// ID-derived element types.
				b.errf(xsd.SpecEPropsCorrect, n.Pos, "element %s has an ID-derived type and must not have a default or fixed value", e.Name)
			} else if _, err := st.ParseValue(*vc, nsContext{n}); err != nil {
				b.errf(xsd.SpecCosValidDefault, n.Pos, "default/fixed value %q is not valid for the type of element %s: %v", *vc, e.Name, err)
			}
		} else if ct, ok := e.Type.(*xsd.ComplexType); ok {
			if ec, ok := ct.Content.(*xsd.ElementContent); ok && !ec.Mixed {
				// spec: cos-valid-default.2.2 — element-only content admits
				// no value constraint (mixed-emptiable check deferred).
				b.errf(xsd.SpecCosValidDefault, n.Pos, "element %s has element-only content and must not have a default or fixed value", e.Name)
			}
		}
	}

	// Substitution-group exclusions: if the member's type is derived from a
	// head's type by a method the head's final excludes, membership is
	// invalid. Unreachable chains are left to the deferred derivation-ok
	// checks rather than guessed at.
	for _, head := range e.SubstitutionGroups {
		if head.Type == nil || head.Final == 0 {
			continue
		}
		if methods, ok := derivationMethods(e.Type, head.Type); ok && methods&head.Final != 0 {
			// spec: e-props-correct.4 — XSD 1.1 Part 1 §3.3.6
			b.errf(xsd.SpecEPropsCorrect, n.Pos, "element %s cannot substitute for %s: the head excludes this derivation", e.Name, head.Name)
		}
	}

	for _, c := range xsdElems(n, doc) {
		switch c.Name.Local {
		case "unique", "key", "keyref":
			if ic := b.buildIC(c, doc); ic != nil {
				e.IdentityConstraints = append(e.IdentityConstraints, ic)
			}
		case "alternative":
			alt := &xsd.TypeAlternative{Pos: c.Pos}
			alt.Test, _ = c.Attr("test")
			alt.Type = b.buildAnonType(c, doc, nil)
			if alt.Type == nil {
				if q, ok := qnameAttr(c, "type"); ok {
					alt.Type = b.resolveType(q, c.Pos, doc)
				}
			}
			e.TypeAlternatives = append(e.TypeAlternatives, alt)
		}
	}
	return e
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

// isIDDerived reports whether st derives from xs:ID by restriction.
func isIDDerived(st *xsd.SimpleType) bool {
	for t := st; t != nil; t, _ = t.BaseType.(*xsd.SimpleType) {
		if t.Name == (xsd.QName{Namespace: xsd.XSDNS, Local: "ID"}) {
			return true
		}
	}
	return false
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
	if st == nil || st.Primitive == nil {
		return
	}
	if st.Primitive.Name.Local == "NOTATION" && !st.Facets.HasEnumeration {
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
	} else {
		a.Name = localQName(n, doc, name, doc.attributeFormDefault)
	}
	if a.Name.Namespace != "" {
		a.Form = xsd.FormQualified
	}

	if inline := firstChild(n, doc, "simpleType"); inline != nil {
		a.Type, _ = b.buildAnonType(n, doc, builtin.AnySimpleType).(*xsd.SimpleType)
	} else if q, ok := qnameAttr(n, "type"); ok {
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
		if isIDDerived(a.Type) {
			// spec: a-props-correct.3 — XSD 1.1 Part 1 §3.2.6
			b.errf(xsd.SpecAPropsCorrect, n.Pos, "attribute %s has an ID-derived type and must not have a default or fixed value", a.Name)
		} else if _, err := a.Type.ParseValue(*vc, nsContext{n}); err != nil {
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
			q, ok := qnameAttr(c, "ref")
			if !ok {
				continue
			}
			d := b.registryFor(doc).lookup(spaceAttrGroup, q)
			if d == nil {
				// spec: src-resolve — XSD 1.1 Part 1 §3.15.3
				b.errf(xsd.SpecSrcResolve, c.Pos, "attribute group %s is not declared", q)
				continue
			}
			g := b.buildAttributeGroup(d)
			if g == nil {
				continue // cyclic, already reported
			}
			uses = append(uses, g.Uses...)
			if wc == nil {
				// Full wildcard intersection (cos-aw-intersect) is deferred;
				// the first wildcard encountered stands in for it.
				wc = g.Wildcard
			}
		case "anyAttribute":
			wc = b.buildWildcard(c, doc)
		}
	}
	return uses, wc, prohibited
}

// attrUseName computes the expanded name a use refers to (its ref target
// or its local declaration's name).
func (b *builder) attrUseName(c *xmltree.Node, doc *schemaDoc) xsd.QName {
	if q, ok := qnameAttr(c, "ref"); ok {
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
	if q, ok := qnameAttr(c, "ref"); ok {
		d := b.registryFor(doc).lookup(spaceAttribute, q)
		if d == nil {
			// spec: src-resolve — XSD 1.1 Part 1 §3.15.3
			b.errf(xsd.SpecSrcResolve, c.Pos, "attribute %s is not declared", q)
			return nil
		}
		u.Decl = b.buildAttributeDecl(d.node, d.doc, true)
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
	return g
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
		if q, ok := qnameAttr(c, "ref"); ok {
			d := b.registryFor(doc).lookup(spaceElement, q)
			if d == nil {
				// spec: src-resolve — XSD 1.1 Part 1 §3.15.3
				b.errf(xsd.SpecSrcResolve, c.Pos, "element %s is not declared", q)
				return nil
			}
			p.Term = b.buildElementDecl(d.node, d.doc, true)
		} else {
			p.Term = b.buildElementDecl(c, doc, false)
		}
	case "group":
		q, ok := qnameAttr(c, "ref")
		if !ok {
			return nil
		}
		d := b.registryFor(doc).lookup(spaceGroup, q)
		if d == nil {
			// spec: src-resolve — XSD 1.1 Part 1 §3.15.3
			b.errf(xsd.SpecSrcResolve, c.Pos, "model group %s is not declared", q)
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
	return mg
}

func (b *builder) buildGroup(d *decl) *xsd.Group {
	if g, ok := b.groups[d.node]; ok {
		return g
	}
	if b.building[d.node] {
		// spec: mg-props-correct.2 — circular model groups are disallowed.
		b.errf(xsd.SpecMGPropsCorrect, d.pos, "model group %s is part of a cycle", d.name)
		return nil
	}
	b.building[d.node] = true
	defer delete(b.building, d.node)

	g := &xsd.Group{
		Name:       d.name,
		Pos:        d.pos,
		Annotation: annotationOf(d.node, d.doc),
		Extensions: extensionsOf(d.node),
	}
	if mg := firstChild(d.node, d.doc, "all", "choice", "sequence"); mg != nil {
		g.Group = b.buildModelGroup(mg, d.doc)
	}
	b.groups[d.node] = g
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
				w.NotQName = append(w.NotQName, q)
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

	if q, ok := qnameAttr(n, "ref"); ok {
		d := b.registryFor(doc).lookup(spaceIC, q)
		if d == nil {
			// spec: src-resolve — XSD 1.1 Part 1 §3.15.3
			b.errf(xsd.SpecSrcResolve, n.Pos, "identity constraint %s is not declared", q)
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
		if q, ok := qnameAttr(n, "refer"); ok {
			d := b.registryFor(doc).lookup(spaceIC, q)
			if d == nil {
				// spec: src-resolve — XSD 1.1 Part 1 §3.15.3
				b.errf(xsd.SpecSrcResolve, n.Pos, "referenced key %s is not declared", q)
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
