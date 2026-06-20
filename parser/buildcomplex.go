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

// buildComplexType returns the SHELL of a complex type: its header properties
// (name, abstract, final, block) and its derivation (base type and method),
// resolved eagerly so derivation edges are known for cycle detection and the
// topological finish pass. The content model and attribute uses are filled
// later by finishComplexType, base-first — a base's own content may legally
// reach back into a type derived from it, a cycle no ordering of derivation
// edges linearizes, so content construction is decoupled from derivation.
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
	// Memoize the shell before resolving the base so a cyclic derivation (or a
	// content reference, once content is filled) resolves to this same
	// component instead of looking like an unbounded recursion.
	b.types[n] = ct
	ct.BaseType, ct.DerivationMethod = b.resolveCTBase(n, doc)
	// Enqueue for the base-first content/attribute finish pass.
	b.ctFinish[ct] = &ctFinishEntry{n: n, doc: doc}
	b.ctOrder = append(b.ctOrder, ct)
	return ct
}

// resolveCTBase determines a complex type's {base type definition} and
// {derivation method} from its content node and reports the base-related
// structural constraints (final-for-derivation, complexContent simple base).
// It resolves the base to its shell only; the base's content is not needed
// until the finish pass.
func (b *builder) resolveCTBase(n *xmltree.Node, doc *schemaDoc) (xsd.Type, xsd.Derivation) {
	derive := func(sub *xmltree.Node) (xsd.Type, xsd.Derivation, bool) {
		r := firstChild(sub, doc, "restriction", "extension")
		if r == nil {
			return builtin.AnyType, xsd.DeriveRestriction, false
		}
		method := xsd.DeriveRestriction
		if r.Name.Local == "extension" {
			method = xsd.DeriveExtension
		}
		base := xsd.Type(builtin.AnyType)
		if q, ok := qnameAttr(r, doc, "base"); ok {
			base = b.resolveType(q, r.Pos, doc)
		}
		return base, method, true
	}
	if sc := firstChild(n, doc, "simpleContent"); sc != nil {
		base, method, ok := derive(sc)
		if ok {
			b.checkFinalAllows(base, method, firstChild(sc, doc, "restriction", "extension").Pos)
		}
		return base, method
	}
	if cc := firstChild(n, doc, "complexContent"); cc != nil {
		base, method, ok := derive(cc)
		if ok {
			r := firstChild(cc, doc, "restriction", "extension")
			if _, isST := base.(*xsd.SimpleType); isST {
				// spec: src-ct.1 — complexContent requires a complex base type.
				b.errf(xsd.SpecSrcCT, r.Pos, "complexContent requires a complex base type, not the simple type %s", base.TypeName())
				base = builtin.AnyType
			}
			b.checkFinalAllows(base, method, r.Pos)
		}
		return base, method
	}
	// Abbreviated form: an implicit restriction of xs:anyType.
	return builtin.AnyType, xsd.DeriveRestriction
}

// finishComplexTypes fills every complex type's content model and attribute
// uses, base-first. The worklist grows as anonymous types are discovered while
// filling content, so it is drained by index.
func (b *builder) finishComplexTypes() {
	for i := 0; i < len(b.ctOrder); i++ {
		b.finishComplexType(b.ctOrder[i])
	}
	b.runStaticTypeChecks()
}

// finishComplexType fills ct's content model and attribute uses, ensuring its
// base type is finished first so extension-particle assembly and attribute-use
// merging read complete base properties. Filling content references other
// types only through their shells, so this recursion follows derivation edges
// alone and terminates even when a base's content reaches back into ct.
func (b *builder) finishComplexType(ct *xsd.ComplexType) {
	if b.ctDone[ct] {
		return
	}
	b.ctDone[ct] = true
	if bct, ok := ct.BaseType.(*xsd.ComplexType); ok {
		b.finishComplexType(bct)
	}
	ent := b.ctFinish[ct]
	if ent == nil {
		return // a builtin (xs:anyType / xs:error): already complete
	}
	n, doc := ent.n, ent.doc
	am := &attrMaterial{node: n, doc: doc, pos: n.Pos}
	switch {
	case firstChild(n, doc, "simpleContent") != nil:
		b.fillSimpleContent(ct, firstChild(n, doc, "simpleContent"), doc, am)
	case firstChild(n, doc, "complexContent") != nil:
		b.fillComplexContent(ct, n, firstChild(n, doc, "complexContent"), doc, am)
	default:
		// Abbreviated form: an implicit restriction of xs:anyType whose content
		// sits directly under <complexType>.
		b.fillElementOnlyContent(ct, n, n, doc, ct.BaseType, ct.DerivationMethod, boolAttr(n, "mixed", false), am)
	}
	b.mergeComplexType(ct, am)
}

// mergeComplexType assembles ct's base-dependent properties once its content
// and own attribute material are in hand and its base is finished: the
// effective particle of an extension, then the merged attribute uses.
func (b *builder) mergeComplexType(ct *xsd.ComplexType, am *attrMaterial) {
	bct, _ := ct.BaseType.(*xsd.ComplexType)
	if bct != nil {
		// {assertions} is the base type's {assertions} followed by this type's
		// own (XSD 1.1 §3.4.2.3.2/§3.4.2.3.3, both extension and restriction):
		// a derived type must satisfy every assertion in its derivation chain.
		// bct is fully finished here, so bct.Assertions already holds the chain.
		ct.Assertions = append(append([]xsd.Assertion(nil), bct.Assertions...), ct.Assertions...)
	}
	if bct != nil && ct.DerivationMethod == xsd.DeriveExtension {
		b.finishExtensionParticle(ct, bct)
	}
	var baseUses []*xsd.AttributeUse
	var baseWC *xsd.Wildcard
	if bct != nil {
		baseUses = bct.AttributeUses
		baseWC = bct.AttributeWildcard
	}
	prohibited := am.prohibited
	if !am.override {
		prohibited = nil
	}
	ct.AttributeUses = b.mergeBaseAttrUses(am.own, baseUses, prohibited, am.override, am.pos)
	ct.AttributeWildcard = am.wc
	if am.wc == nil && am.wcFallback {
		// With no own wildcard the base's stands in (both restriction and the
		// extension union degenerate to the base wildcard).
		ct.AttributeWildcard = baseWC
	} else if am.wc != nil && baseWC != nil && ct.DerivationMethod == xsd.DeriveExtension {
		// spec: cos-aw-union — XSD 1.1 Part 1 §3.10.6.2 (cos-aw-union): an
		// extension's {attribute wildcard} is the union of the base's wildcard
		// and the extension's own wildcard.
		ct.AttributeWildcard = wildcardUnion(baseWC, am.wc)
	}
	if ct.DerivationMethod == xsd.DeriveRestriction && bct != nil {
		if am.wc != nil {
			b.checkAttrWildcardRestriction(ct, am.wc, baseWC)
		}
		b.checkAttrRestriction(am.own, baseUses)
	}
	if ct.DerivationMethod == xsd.DeriveExtension && bct != nil {
		b.inheritExtensionOpenContent(ct, bct)
		b.checkExtensionOpenContent(ct, bct, am.contentNode, am.doc)
	}
	b.applyDefaultAttributes(ct, am.node, am.doc)
	b.checkAttrUses(ct)
}

func (b *builder) fillSimpleContent(ct *xsd.ComplexType, sc *xmltree.Node, doc *schemaDoc, am *attrMaterial) {
	r := firstChild(sc, doc, "restriction", "extension")
	if r == nil {
		ct.Content = &xsd.SimpleContent{Type: builtin.AnySimpleType}
		return
	}
	isExtension := r.Name.Local == "extension"
	base := ct.BaseType

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
			// particle and an inline <simpleType> (src-ct.2.2).
			if isExtension || !c.Mixed || firstChild(r, doc, "simpleType") == nil {
				// spec: src-ct.2 — XSD 1.1 Part 1 §3.4.3
				b.errf(xsd.SpecSrcCT, r.Pos, "base type %s does not have simple content", base.Name)
			} else if !particleEmptiable(c.Particle) {
				// §3.4.2.2 clause 2: simpleContent restriction of a mixed base requires an
				// emptiable particle (§3.9.6.3); a non-emptiable particle produces a type
				// that cannot satisfy derivation-ok-restriction §3.4.6.3 clause 2.2.2.2.
				b.errf(xsd.SpecSrcCT, r.Pos, "base type %s has mixed content with a non-emptiable particle; simpleContent restriction requires an emptiable particle", base.Name)
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
	// merged by mergeComplexType once the base is finished.
	am.own, am.wc, am.prohibited = b.buildAttrUses(r, doc)
	am.override = !isExtension
	am.wcFallback = true
	am.pos = r.Pos
	ct.Assertions = b.buildAsserts(r, doc)
}

// fillComplexContent fills ct's element-only content from its <complexContent>
// child. The base type and derivation method were resolved into ct by
// resolveCTBase; this pass reads only the content (particle, attributes, open
// content).
func (b *builder) fillComplexContent(ct *xsd.ComplexType, n, cc *xmltree.Node, doc *schemaDoc, am *attrMaterial) {
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

	content := n
	if r := firstChild(cc, doc, "restriction", "extension"); r != nil {
		content = r
	}
	b.fillElementOnlyContent(ct, n, content, doc, ct.BaseType, ct.DerivationMethod, mixed, am)
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

// fillElementOnlyContent fills ct with element (or empty) content read from
// the children of content (= the restriction/extension element, or the
// complexType itself in the abbreviated form). The base is already finished,
// so a simple-content base extension is resolved here directly.
func (b *builder) fillElementOnlyContent(ct *xsd.ComplexType, n, content *xmltree.Node, doc *schemaDoc, base xsd.Type, method xsd.Derivation, mixed bool, am *attrMaterial) {
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
				b.collectAttrs(am, content, doc, false, n.Pos)
				ct.Assertions = b.buildAsserts(content, doc)
				return
			}
			// spec: cos-ct-extends.1.4.2 — a complex extension of a
			// simple-content type cannot add element content.
			b.errf(xsd.SpecCosCTExtends, content.Pos, "cannot extend %s with element content: its content is simple", bct.Name)
		}
		// Otherwise the effective particle (base particle followed by the
		// extension's) is assembled by finishExtensionParticle from the merge.
	}

	ec := &xsd.ElementContent{Mixed: mixed, Particle: particle}
	defDoc := b.defaultsDoc(n, doc)
	if ocn := firstChild(content, doc, "openContent"); ocn != nil {
		ec.OpenContent = b.buildOpenContent(ocn, doc)
	} else if defDoc.defaultOpenContent != nil {
		// Apply defaultOpenContent when the content type is non-empty (i.e. the
		// particle can match at least one element, or the type is mixed — mixed
		// always has a non-empty content type), OR when appliesToEmpty=true.
		// A bare <xs:sequence/> with no children is an empty content type, so
		// particleMatchesNonEmpty correctly returns false for it. §3.11.4.2.
		// A type declared inside <xs:override> takes the overridden document's
		// defaultOpenContent (defDoc), not the overriding schema's (saxon open043).
		nonEmpty := mixed || particleMatchesNonEmpty(particle)
		if nonEmpty || boolAttr(defDoc.defaultOpenContent, "appliesToEmpty", false) {
			ec.OpenContent = b.buildOpenContent(defDoc.defaultOpenContent, defDoc)
		}
	}
	ct.Content = ec

	b.collectAttrs(am, content, doc, method != xsd.DeriveExtension, content.Pos)
	ct.Assertions = b.buildAsserts(content, doc)
}

// collectAttrs records the attribute material declared on content into am for
// the mergeComplexType pass. Extensions unite with the base's uses and fall
// back to its wildcard (wildcard union per cos-aw-union is applied in
// mergeComplexType); restrictions override by name and keep only their own
// wildcard.
func (b *builder) collectAttrs(am *attrMaterial, content *xmltree.Node, doc *schemaDoc, override bool, pos xsd.Pos) {
	am.own, am.wc, am.prohibited = b.buildAttrUses(content, doc)
	am.override = override
	am.wcFallback = !override
	am.pos = pos
	am.contentNode = content
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

// defaultsDoc returns the document whose schema-level defaults (defaultAttributes
// / defaultOpenContent) govern the component defined at node n: the overridden
// (target) document when n is declared inside an <xs:override> (§4.2.4 places it
// there), otherwise doc itself.
func (b *builder) defaultsDoc(n *xmltree.Node, doc *schemaDoc) *schemaDoc {
	if target := b.overrideTarget[n]; target != nil {
		return target
	}
	return doc
}

// applyDefaultAttributes appends the schema's defaultAttributes group
// unless the type opts out.
func (b *builder) applyDefaultAttributes(ct *xsd.ComplexType, n *xmltree.Node, doc *schemaDoc) {
	// A complex type declared inside <xs:override> belongs to the overridden
	// document (§4.2.4), so that document's defaultAttributes apply — not the
	// overriding schema's. saxon open045 / ibm s3_4_2_4ii08.
	doc = b.defaultsDoc(n, doc)
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
	// The default attribute group is added like an extra <attributeGroup ref>;
	// a name collision with an existing use is a duplicate (ct-props-correct.4),
	// not a silent override, so append and let checkAttrUses flag it.
	ct.AttributeUses = append(ct.AttributeUses, g.Uses...)
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
