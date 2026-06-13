package parser

// Complex type construction (Part 1 §3.4): simpleContent restriction/
// extension, complexContent restriction/extension, and the abbreviated form
// (an implicit restriction of xs:anyType). Particle-level derivation checks
// (cos-ct-restricts / cos-particle-restrict / UPA / EDC) are deferred; see
// NOTES.md.

import (
	"github.com/kud360/goxsd5/builtin"
	"github.com/kud360/goxsd5/parser/xmltree"
	"github.com/kud360/goxsd5/xsd"
)

func (b *builder) buildComplexType(n *xmltree.Node, doc *schemaDoc, name xsd.QName) *xsd.ComplexType {
	if t, ok := b.types[n]; ok {
		if ct, ok := t.(*xsd.ComplexType); ok {
			return ct
		}
	}
	ct := &xsd.ComplexType{
		Name:       name,
		Pos:        n.Pos,
		Abstract:   boolAttr(n, "abstract", false),
		Final:      derivAttr(n, "final", ctFinalSet, doc.finalDefault),
		Block:      derivAttr(n, "block", ctBlockSet, doc.blockDefault),
		Annotation: annotationOf(n, doc),
		Extensions: extensionsOf(n),
	}
	// Memoize the shell before building content so content references back
	// into this type (legal) resolve to the same component instead of
	// looking like cycles.
	b.types[n] = ct
	// Attribute uses are merged with the base's in finishComplexTypes; the
	// content builders below fill in the declared material.
	b.pendingAttrs[ct] = &pendingAttrs{pos: n.Pos, node: n, doc: doc}

	switch {
	case firstChild(n, doc, "simpleContent") != nil:
		b.buildSimpleContent(ct, firstChild(n, doc, "simpleContent"), doc)
	case firstChild(n, doc, "complexContent") != nil:
		b.buildComplexContent(ct, n, firstChild(n, doc, "complexContent"), doc)
	default:
		// Abbreviated form: an implicit restriction of xs:anyType whose
		// content sits directly under <complexType>.
		b.buildElementOnlyContent(ct, n, n, doc, builtin.AnyType, xsd.DeriveRestriction, boolAttr(n, "mixed", false))
	}

	return ct
}

func (b *builder) buildSimpleContent(ct *xsd.ComplexType, sc *xmltree.Node, doc *schemaDoc) {
	r := firstChild(sc, doc, "restriction", "extension")
	if r == nil {
		ct.BaseType = builtin.AnyType
		ct.Content = &xsd.SimpleContent{Type: builtin.AnySimpleType}
		return
	}
	isExtension := r.Name.Local == "extension"
	if isExtension {
		ct.DerivationMethod = xsd.DeriveExtension
	} else {
		ct.DerivationMethod = xsd.DeriveRestriction
	}

	base := xsd.Type(builtin.AnyType)
	if q, ok := qnameAttr(r, doc, "base"); ok {
		base = b.resolveType(q, r.Pos, doc)
	}
	ct.BaseType = base
	b.checkFinalAllows(base, ct.DerivationMethod, r.Pos)

	var contentST *xsd.SimpleType
	switch base := base.(type) {
	case *xsd.SimpleType:
		if !isExtension {
			// spec: src-ct.2.1 — XSD 1.1 Part 1 §3.4.3: a simpleContent
			// restriction's base must be a complex type.
			b.errf(xsd.SpecSrcCT, r.Pos, "simpleContent restriction requires a complex base type, not the simple type %s", base.Name)
			contentST = base
		} else {
			contentST = base
		}
	case *xsd.ComplexType:
		switch c := base.Content.(type) {
		case *xsd.SimpleContent:
			contentST = c.Type
		case *xsd.ElementContent:
			// Only legal for restriction of a mixed type with an emptiable
			// particle and an inline <simpleType> (src-ct.2.2; the
			// emptiable check is deferred).
			if isExtension || !c.Mixed || firstChild(r, doc, "simpleType") == nil {
				// spec: src-ct.2 — XSD 1.1 Part 1 §3.4.3
				b.errf(xsd.SpecSrcCT, r.Pos, "base type %s does not have simple content", base.Name)
			}
			contentST = builtin.AnySimpleType
		default:
			b.errf(xsd.SpecSrcCT, r.Pos, "base type %s does not have simple content", base.Name)
			contentST = builtin.AnySimpleType
		}
	}

	if isExtension {
		ct.Content = &xsd.SimpleContent{Type: contentST}
	} else {
		// The effective simple base is the inline <simpleType> when given,
		// else the base's content type; the declared facets restrict it.
		effBase := contentST
		if inline := firstChild(r, doc, "simpleType"); inline != nil {
			if st, ok := b.buildAnonType(r, doc, contentST).(*xsd.SimpleType); ok {
				effBase = st
			}
		}
		st := &xsd.SimpleType{Pos: r.Pos}
		b.applyRestriction(st, effBase, r, doc)
		ct.Content = &xsd.SimpleContent{Type: st}
	}

	// Attribute uses: own plus the base's (restriction overrides by name),
	// merged in finishComplexTypes.
	p := b.pendingAttrs[ct]
	p.own, p.wc, p.prohibited = b.buildAttrUses(r, doc)
	p.override = !isExtension
	p.wcFallback = true
	p.pos = r.Pos
	ct.Assertions = b.buildAsserts(r, doc)
}

func (b *builder) buildComplexContent(ct *xsd.ComplexType, n, cc *xmltree.Node, doc *schemaDoc) {
	mixed := boolAttr(n, "mixed", false)
	if v, ok := cc.Attr("mixed"); ok {
		ccMixed, err := parseBool(v)
		if err == nil {
			if _, onCT := n.Attr("mixed"); onCT && ccMixed != mixed {
				// spec: src-ct — mixed on <complexType> and <complexContent>
				// must agree when both are present.
				b.errf(xsd.SpecSrcCT, cc.Pos, "mixed is declared inconsistently on <complexType> and <complexContent>")
			}
			mixed = ccMixed
		}
	}

	r := firstChild(cc, doc, "restriction", "extension")
	if r == nil {
		b.buildElementOnlyContent(ct, n, n, doc, builtin.AnyType, xsd.DeriveRestriction, mixed)
		return
	}
	method := xsd.DeriveRestriction
	if r.Name.Local == "extension" {
		method = xsd.DeriveExtension
	}

	base := xsd.Type(builtin.AnyType)
	if q, ok := qnameAttr(r, doc, "base"); ok {
		base = b.resolveType(q, r.Pos, doc)
	}
	if _, isST := base.(*xsd.SimpleType); isST {
		// spec: src-ct.1 — complexContent requires a complex base type.
		b.errf(xsd.SpecSrcCT, r.Pos, "complexContent requires a complex base type, not the simple type %s", base.TypeName())
		base = builtin.AnyType
	}
	b.checkFinalAllows(base, method, r.Pos)
	b.buildElementOnlyContent(ct, n, r, doc, base, method, mixed)
}

// checkFinalAllows reports an error when base's {final} set blocks deriving a
// new type from it by the given method (extension or restriction).
// spec: cos-ct-extends.1.1 / derivation-ok-restriction.1 — XSD 1.1 Part 1
// §3.4.6: B.{final} must not contain the derivation method being used.
func (b *builder) checkFinalAllows(base xsd.Type, method xsd.Derivation, pos xsd.Pos) {
	var final xsd.DerivationSet
	switch t := base.(type) {
	case *xsd.SimpleType:
		final = t.Final
	case *xsd.ComplexType:
		final = t.Final
	default:
		return
	}
	if !final.Has(method) {
		return
	}
	ref, verb := xsd.SpecDerivationOKRestriction, "restriction"
	if method == xsd.DeriveExtension {
		ref, verb = xsd.SpecCosCTExtends, "extension"
	}
	name := "the base type"
	if q := base.TypeName(); !q.IsZero() {
		name = q.String()
	}
	b.errf(ref, pos, "%s is final for %s; it cannot be the base of a %s", name, verb, verb)
}

// buildElementOnlyContent fills ct with element (or empty) content read
// from the children of content (= the restriction/extension element, or the
// complexType itself in the abbreviated form).
func (b *builder) buildElementOnlyContent(ct *xsd.ComplexType, n, content *xmltree.Node, doc *schemaDoc, base xsd.Type, method xsd.Derivation, mixed bool) {
	ct.BaseType = base
	ct.DerivationMethod = method
	ct.Mixed = mixed

	var particle *xsd.Particle
	if pn := firstChild(content, doc, "group", "all", "choice", "sequence"); pn != nil {
		particle = b.buildParticle(pn, doc)
	}

	bct, _ := base.(*xsd.ComplexType)
	if method == xsd.DeriveExtension && bct != nil {
		if sc, isSimple := bct.Content.(*xsd.SimpleContent); isSimple {
			if particle == nil && !mixed {
				// Extending a simple-content type without adding element
				// content keeps the simple content.
				ct.Content = &xsd.SimpleContent{Type: sc.Type}
				b.deferAttrs(ct, content, doc, false, n.Pos)
				ct.Assertions = b.buildAsserts(content, doc)
				return
			}
			// spec: cos-ct-extends.1.4.2 — a complex extension of a
			// simple-content type cannot add element content.
			b.errf(xsd.SpecCosCTExtends, content.Pos, "cannot extend %s with element content: its content is simple", bct.Name)
		}
		// The effective particle (base particle followed by the extension's)
		// is combined by finishExtensions: the base may still be mid-build
		// here when its own content reaches back into this type.
	}

	ec := &xsd.ElementContent{Mixed: mixed, Particle: particle}
	if ocn := firstChild(content, doc, "openContent"); ocn != nil {
		ec.OpenContent = b.buildOpenContent(ocn, doc)
	} else if doc.defaultOpenContent != nil {
		if particle != nil || boolAttr(doc.defaultOpenContent, "appliesToEmpty", false) {
			ec.OpenContent = b.buildOpenContent(doc.defaultOpenContent, doc)
		}
	}
	ct.Content = ec

	b.deferAttrs(ct, content, doc, method != xsd.DeriveExtension, content.Pos)
	ct.Assertions = b.buildAsserts(content, doc)
}

// deferAttrs records the attribute material declared on content for the
// finishComplexTypes merge. Extensions unite with the base's uses and fall
// back to its wildcard (full wildcard union, cos-aw-union, is deferred);
// restrictions override by name and keep only their own wildcard.
func (b *builder) deferAttrs(ct *xsd.ComplexType, content *xmltree.Node, doc *schemaDoc, override bool, pos xsd.Pos) {
	p := b.pendingAttrs[ct]
	p.own, p.wc, p.prohibited = b.buildAttrUses(content, doc)
	p.override = override
	p.wcFallback = !override
	p.pos = pos
}

// mergeBaseAttrUses combines declared uses with inherited ones. For
// restrictions (override=true) a declared use replaces the inherited use of
// the same name and prohibited names are dropped; for extensions the sets
// are united.
func (b *builder) mergeBaseAttrUses(own, base []*xsd.AttributeUse, prohibited map[xsd.QName]bool, override bool, p xsd.Pos) []*xsd.AttributeUse {
	byName := map[xsd.QName]*xsd.AttributeUse{}
	out := make([]*xsd.AttributeUse, 0, len(own)+len(base))
	for _, u := range own {
		if u.Decl == nil {
			continue
		}
		out = append(out, u)
		byName[u.Decl.Name] = u
	}
	for _, u := range base {
		if u.Decl == nil {
			continue
		}
		if prohibited != nil && prohibited[u.Decl.Name] {
			continue
		}
		if _, shadowed := byName[u.Decl.Name]; shadowed {
			if !override {
				// spec: ct-props-correct.4 — two attribute uses with the
				// same expanded name.
				b.errf(xsd.SpecCTPropsCorrect, p, "attribute %s is declared twice (extension conflicts with the base type)", u.Decl.Name)
			}
			continue
		}
		out = append(out, u)
		byName[u.Decl.Name] = u
	}
	return out
}

// applyDefaultAttributes appends the schema's defaultAttributes group
// unless the type opts out.
func (b *builder) applyDefaultAttributes(ct *xsd.ComplexType, n *xmltree.Node, doc *schemaDoc) {
	if doc.defaultAttributes.IsZero() || !boolAttr(n, "defaultAttributesApply", true) {
		return
	}
	// Resolution failure reports once per type use; the schema-level
	// reference has no other resolution point.
	d := b.lookupRef(spaceAttrGroup, doc.defaultAttributes, n.Pos, doc)
	if d == nil {
		return
	}
	g := b.buildAttributeGroup(d)
	if g == nil {
		return
	}
	ct.AttributeUses = b.mergeBaseAttrUses(ct.AttributeUses, g.Uses, nil, true, n.Pos)
	if ct.AttributeWildcard == nil {
		ct.AttributeWildcard = g.Wildcard
	}
}

// checkAttrUses runs the post-merge per-type attribute constraints.
func (b *builder) checkAttrUses(ct *xsd.ComplexType) {
	// Note: XSD 1.1 dropped the 1.0 rule (old ct-props-correct.5) limiting a
	// complex type to a single ID-derived attribute; only the duplicate-name
	// check (ct-props-correct.4) remains.
	seen := map[xsd.QName]bool{}
	for _, u := range ct.AttributeUses {
		if u.Decl == nil {
			continue
		}
		if seen[u.Decl.Name] {
			// spec: ct-props-correct.4 — XSD 1.1 Part 1 §3.4.6
			b.errf(xsd.SpecCTPropsCorrect, u.Pos, "attribute %s is declared twice on %s", u.Decl.Name, describeCT(ct))
			continue
		}
		seen[u.Decl.Name] = true
	}
}

func describeCT(ct *xsd.ComplexType) string {
	if ct.Name.IsZero() {
		return "anonymous complex type"
	}
	return ct.Name.String()
}

func (b *builder) buildAsserts(content *xmltree.Node, doc *schemaDoc) []xsd.Assertion {
	var out []xsd.Assertion
	for _, c := range xsdElems(content, doc) {
		if c.Name.Local == "assert" {
			out = append(out, b.buildAssertion(c, doc))
		}
	}
	return out
}

func (b *builder) buildOpenContent(n *xmltree.Node, doc *schemaDoc) *xsd.OpenContent {
	oc := &xsd.OpenContent{Pos: n.Pos, Mode: xsd.OpenContentInterleave}
	if v, ok := n.Attr("mode"); ok {
		switch v {
		case "suffix":
			oc.Mode = xsd.OpenContentSuffix
		case "none":
			oc.Mode = xsd.OpenContentNone
		}
	}
	if any := firstChild(n, doc, "any"); any != nil {
		oc.Wildcard = b.buildWildcard(any, doc)
	}
	return oc
}
