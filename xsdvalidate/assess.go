package xsdvalidate

import (
	"strings"

	"github.com/kud360/goxsd5/builtin"
	"github.com/kud360/goxsd5/xsd"
	"github.com/kud360/goxsd5/xsdwalk"
)

// assessor holds the per-run state of one Assess call: the immutable validator,
// the result being built, and the document-scoped tables (cvc-id) the engine
// owns because they are not per-datatype.
type assessor struct {
	v   *Validator
	res *Result

	// ids maps each declared xs:ID value to the element that bears it. XSD 1.1
	// scopes ID uniqueness per element: an element may carry several ID
	// attributes with the same value, but two distinct elements may not share an
	// ID value (cvc-id.2, §3.4.4 — relaxed from XSD 1.0's single-ID-per-element).
	ids    map[string]Element
	idrefs []idref // pending IDREF values, resolved after the walk

	// entities is the set of unparsed (NDATA) entity names declared in the
	// document, for xs:ENTITY referential validity; haveEntities records whether
	// the source supplied DTD info at all (else ENTITY checking is fail-open).
	entities     map[string]bool
	haveEntities bool

	// attrType records the simple type assessed for each attribute node, so
	// identity-constraint fields selecting attributes can compare typed values.
	attrType map[Attribute]*xsd.SimpleType

	// skipped records elements matched by a processContents="skip" wildcard:
	// they are not assessed, so they (and their subtrees) are excluded from
	// identity-constraint target/field selection.
	skipped map[Element]bool
}

type idref struct {
	val string
	pos xsd.Pos
}

// nsContext adapts an instance element's in-scope namespaces to the
// xsd.ValueContext that QName/NOTATION value parsing needs.
type nsContext struct{ el Element }

func (c nsContext) ResolveQName(prefix, local string) (xsd.QName, bool) {
	uri, ok := c.el.Lookup(prefix)
	if prefix == "" {
		// An unprefixed QName uses the default namespace if declared, else no
		// namespace; either way it resolves.
		return xsd.QName{Namespace: uri, Local: local}, true
	}
	if !ok {
		return xsd.QName{}, false
	}
	return xsd.QName{Namespace: uri, Local: local}, true
}

func (a *assessor) addf(ref xsd.SpecRef, pos xsd.Pos, format string, args ...any) {
	a.res.errs.Addf(ref, pos, format, args...)
}

// addValueErr re-emits a value-space error (from ParseValue) anchored at pos.
func (a *assessor) addValueErr(err error, pos xsd.Pos) {
	for _, e := range xsd.AllErrors(err) {
		if xe, ok := e.(*xsd.Error); ok {
			if xe.Pos.IsZero() {
				xe.Pos = pos
			}
			a.res.errs.Add(xe)
			continue
		}
		a.res.errs.Add(e)
	}
}

func (a *assessor) assessRoot(root Element) {
	if di, ok := root.(DocumentInfo); ok {
		a.haveEntities = true
		a.entities = di.UnparsedEntities()
	}
	decl := a.v.elements[root.Name()]
	if decl == nil {
		// No governing element declaration. Per cvc-assess-elt, if xsi:type names
		// a type that becomes the governing type definition (clause 8) and the
		// root is strictly assessed against it; otherwise there is nothing to
		// validate the root against and the document is treated as invalid.
		if typeStr, ok := a.xsiValue(root, "type"); ok {
			if t := a.resolveRootXSIType(root, typeStr); t != nil {
				a.res.Types[root] = t
				a.assessType(root, t, nil, nil, nil)
			}
			return
		}
		// spec: cvc-elt — XSD 1.1 Part 1 §3.3.4 (xmlschema11-1.md#cvc-elt)
		a.addf(xsd.SpecCvcElt, root.Pos(), "no global element declaration for root element %s", root.Name())
		return
	}
	a.assessElement(root, decl, nil, nil)
}

// resolveRootXSIType resolves an xsi:type on a root element that has no
// governing element declaration; with no declared type there is no derivation
// constraint (cvc-elt.4.3 is vacuous). It returns nil on an unresolvable or
// unknown type, after recording the error.
func (a *assessor) resolveRootXSIType(root Element, lexical string) xsd.Type {
	tn, err := resolveQNameInScope(root, lexical)
	if err != nil {
		a.addf(xsd.SpecCvcElt, root.Pos(), "xsi:type %q is not a resolvable QName", lexical)
		return nil
	}
	t := a.v.typeByName(tn)
	if t == nil {
		a.addf(xsd.SpecCvcElt, root.Pos(), "xsi:type names unknown type %s", tn)
	}
	return t
}

// assessElement implements cvc-elt (Element Locally Valid (Element), §3.3.4).
// inherited carries the XSD 1.1 inheritable attributes contributed by ancestor
// elements, available to conditional-type-assignment tests on this element.
// parent is the element's parent (nil at the validation root); it is the binding
// target for any xs:ID carried by this element's simple content (§3.3.4.5).
func (a *assessor) assessElement(el Element, decl *xsd.ElementDecl, inherited map[xsd.QName]string, parent Element) {
	// cvc-elt.2: the declaration must not be abstract.
	if decl.Abstract {
		a.addf(xsd.SpecCvcElt, el.Pos(), "element %s is declared abstract and cannot appear in an instance", decl.Name)
	}

	gov := a.selectGoverningType(el, decl, decl.Type, inherited)
	nilled := false

	// cvc-elt.3: xsi:nil.
	if nilStr, ok := a.xsiValue(el, "nil"); ok {
		switch {
		case !decl.Nillable:
			a.addf(xsd.SpecCvcElt, el.Pos(), "xsi:nil on element %s whose declaration is not nillable", decl.Name)
		case parseXSIBool(nilStr):
			nilled = true
			// cvc-elt.3.2.2: a nilled element may not carry a fixed value.
			if decl.Fixed != nil {
				a.addf(xsd.SpecCvcElt, el.Pos(), "element %s is nilled but has a fixed value constraint", decl.Name)
			}
		}
	}

	// cvc-elt.4: xsi:type overrides the governing type.
	if typeStr, ok := a.xsiValue(el, "type"); ok {
		gov = a.resolveXSIType(el, decl, gov, typeStr)
	}
	a.res.Types[el] = gov

	if nilled {
		// cvc-elt.3.2.1: a nilled element must have no character or element
		// content. Attributes are still assessed.
		if hasContent(el) {
			a.addf(xsd.SpecCvcElt, el.Pos(), "nilled element %s must be empty", decl.Name)
		}
		if ct, ok := gov.(*xsd.ComplexType); ok {
			a.assessAttributes(el, ct.AttributeUses, ct.AttributeWildcard)
		} else {
			a.assessAttributes(el, nil, nil)
		}
		return
	}

	// cvc-elt.5: the element is locally valid with respect to its type.
	a.assessType(el, gov, decl, childInherited(el, gov, inherited), parent)

	// Identity constraints are scoped at this element and read its subtree.
	a.checkIdentityConstraints(el, decl)
}

// childInherited extends the inherited-attribute set with this element's own
// attributes whose declarations are inheritable (XSD 1.1 §3.4.4). The element's
// own values override any inherited of the same name.
func childInherited(el Element, gov xsd.Type, parent map[xsd.QName]string) map[xsd.QName]string {
	ct, ok := gov.(*xsd.ComplexType)
	if !ok {
		return parent
	}
	var out map[xsd.QName]string
	clone := func() {
		if out == nil {
			out = make(map[xsd.QName]string, len(parent)+2)
			for k, v := range parent {
				out[k] = v
			}
		}
	}
	for _, u := range ct.AttributeUses {
		if u.Decl == nil || !u.Inheritable {
			continue
		}
		for _, at := range el.Attributes() {
			if at.Name() == u.Decl.Name {
				clone()
				out[u.Decl.Name] = at.Value()
			}
		}
	}
	if out == nil {
		return parent
	}
	return out
}

// resolveXSIType handles cvc-elt.4: resolve and validate the xsi:type override.
func (a *assessor) resolveXSIType(el Element, decl *xsd.ElementDecl, declared xsd.Type, lexical string) xsd.Type {
	tn, err := resolveQNameInScope(el, lexical)
	if err != nil {
		// spec: cvc-elt — XSD 1.1 Part 1 §3.3.4 (cvc-elt.4.1)
		a.addf(xsd.SpecCvcElt, el.Pos(), "xsi:type %q is not a resolvable QName", lexical)
		return declared
	}
	t := a.v.typeByName(tn)
	if t == nil {
		a.addf(xsd.SpecCvcElt, el.Pos(), "xsi:type names unknown type %s", tn)
		return declared
	}
	// cvc-elt.4.3: t must be validly derived from the declared type given the
	// blocking set (element {disallowed substitutions} plus the declared type's
	// {prohibited substitutions}). decl may be nil for a wildcard-matched element
	// governed only by a locally declared type.
	var block xsd.DerivationSet
	if decl != nil {
		block = decl.Block
	}
	if ct, ok := declared.(*xsd.ComplexType); ok {
		block |= ct.Block
	}
	if !xsdwalk.DerivationOK(t, declared, block) {
		a.addf(xsd.SpecCvcElt, el.Pos(), "xsi:type %s is not validly derived from the declared type of %s", tn, el.Name())
		return declared
	}
	return t
}

// assessType implements cvc-type (§3.4.4): dispatch on simple vs complex.
func (a *assessor) assessType(el Element, t xsd.Type, decl *xsd.ElementDecl, inherited map[xsd.QName]string, parent Element) {
	switch t := t.(type) {
	case *xsd.SimpleType:
		// cvc-type.3.1.1/.2: a simple-typed element admits no element children
		// and only xsi:* attributes.
		a.assessAttributes(el, nil, nil)
		if hasElementChildren(el) {
			a.addf(xsd.SpecCvcType, el.Pos(), "element %s has a simple type but contains element children", el.Name())
		}
		a.validateSimpleContent(el, t, charContent(el), decl, parent)
	case *xsd.ComplexType:
		a.assessComplexType(el, t, decl, inherited, parent)
	}
}

// assessComplexType implements cvc-complex-type (§3.4.4).
func (a *assessor) assessComplexType(el Element, ct *xsd.ComplexType, decl *xsd.ElementDecl, inherited map[xsd.QName]string, parent Element) {
	// cvc-complex-type.1: the type must not be abstract.
	if ct.Abstract {
		a.addf(xsd.SpecCvcComplexType, el.Pos(), "complex type %s is abstract and cannot be instantiated", typeName(ct))
	}
	// cvc-complex-type.3: attributes.
	a.assessAttributes(el, ct.AttributeUses, ct.AttributeWildcard)

	switch content := ct.Content.(type) {
	case *xsd.SimpleContent:
		// cvc-complex-type.2.2: simple content admits no element children.
		if hasElementChildren(el) {
			a.addf(xsd.SpecCvcComplexType, el.Pos(), "element %s has simple content but contains element children", decl.Name)
		}
		a.validateSimpleContent(el, content.Type, charContent(el), decl, parent)
	case *xsd.ElementContent:
		a.assessElementContent(el, content, inherited)
		// cvc-elt.5.2.2.1: a fixed value on a mixed content type forbids element
		// children and requires the character content to equal the fixed value.
		if decl != nil && decl.Fixed != nil && content.Mixed {
			if hasElementChildren(el) {
				a.addf(xsd.SpecCvcElt, el.Pos(), "element %s has a fixed value but contains element children", decl.Name)
			} else if charContent(el) != *decl.Fixed {
				a.addf(xsd.SpecCvcElt, el.Pos(), "element %s content %q does not match fixed value %q", decl.Name, charContent(el), *decl.Fixed)
			}
		}
	default: // empty content
		// cvc-complex-type.2.1: empty content admits neither element nor
		// non-whitespace character content.
		if hasElementChildren(el) || hasNonWhitespace(charContent(el)) {
			a.addf(xsd.SpecCvcComplexType, el.Pos(), "element %s has empty content type but is not empty", decl.Name)
		}
	}

	// cvc-complex-type.5 / cvc-assertion: xs:assert children.
	a.checkAssertions(el, ct)
}

// assessElementContent matches the children against the content model and
// recurses (cvc-complex-type.2.3/.4, cvc-particle).
func (a *assessor) assessElementContent(el Element, ec *xsd.ElementContent, inherited map[xsd.QName]string) {
	// cvc-complex-type.2.3: element-only content admits only whitespace text.
	if !ec.Mixed && hasNonWhitespace(charContent(el)) {
		a.addf(xsd.SpecCvcComplexType, el.Pos(), "element-only content has character data")
	}
	kids := elementChildren(el)
	names := make([]xsd.QName, len(kids))
	for i, k := range kids {
		names[i] = k.Name()
	}
	terms, ok := a.v.matcher.Match(ec.Particle, names, ec.OpenContent)
	if !ok {
		// spec: cvc-particle — XSD 1.1 Part 1 §3.9.4 (xmlschema11-1.md#cvc-particle)
		a.addf(xsd.SpecCvcParticle, el.Pos(), "content of element does not match its content model")
		// Still recurse into children we can place, for ID collection etc.
		return
	}
	local := localDeclTypes(ec.Particle)
	for i, k := range kids {
		a.assessChild(k, terms[i], inherited, el, local)
	}
}

// localDeclTypes maps each element name appearing in the content model particle
// to its declared {type definition} — the "locally declared type" of §3.9. Used
// to govern a wildcard-matched element whose name collides with a model element
// (XSD 1.1 tighter EDC matching: cvc-assess-elt governing-type clause 7).
func localDeclTypes(p *xsd.Particle) map[xsd.QName]xsd.Type {
	out := map[xsd.QName]xsd.Type{}
	var walk func(*xsd.Particle)
	walk = func(p *xsd.Particle) {
		if p == nil {
			return
		}
		switch t := p.Term.(type) {
		case *xsd.ElementDecl:
			if _, seen := out[t.Name]; !seen {
				out[t.Name] = t.Type
			}
		case *xsd.ModelGroup:
			for _, sub := range t.Particles {
				walk(sub)
			}
		case *xsd.GroupRef:
			if t.Ref != nil && t.Ref.Group != nil {
				for _, sub := range t.Ref.Group.Particles {
					walk(sub)
				}
			}
		}
	}
	walk(p)
	return out
}

// assessChild recurses into one matched child element. parent is the element
// whose content model placed child (child's binding parent for cvc-id); local
// maps model element names to their locally declared types (§3.9).
func (a *assessor) assessChild(child Element, mt xsdwalk.MatchedTerm, inherited map[xsd.QName]string, parent Element, local map[xsd.QName]xsd.Type) {
	if mt.Elem != nil {
		a.assessElement(child, mt.Elem, inherited, parent)
		return
	}
	if mt.Wildcard == nil {
		return
	}
	switch mt.Wildcard.ProcessContents {
	case xsd.ProcessSkip:
		// No assessment of the wildcard-matched subtree; record it so identity
		// constraints do not reach into the unassessed region.
		if a.skipped == nil {
			a.skipped = map[Element]bool{}
		}
		a.skipped[child] = true
	case xsd.ProcessLax:
		if d := a.v.elements[child.Name()]; d != nil {
			a.assessElement(child, d, inherited, parent)
			a.checkDynamicEDC(child, local)
		} else if t, ok := local[child.Name()]; ok {
			// No global declaration, but the name has a locally declared type: it
			// governs the element (cvc-assess-elt clause 7), xsi:type-overridable.
			a.assessLocallyTyped(child, t, inherited, parent)
		}
	case xsd.ProcessStrict:
		d := a.v.elements[child.Name()]
		if d == nil {
			if t, ok := local[child.Name()]; ok {
				a.assessLocallyTyped(child, t, inherited, parent)
				return
			}
			// spec: cvc-wildcard — XSD 1.1 Part 1 §3.10.4 (xmlschema11-1.md#cvc-wildcard)
			a.addf(xsd.SpecCvcWildcard, child.Pos(), "no declaration for element %s matched by a strict wildcard", child.Name())
			return
		}
		a.assessElement(child, d, inherited, parent)
		a.checkDynamicEDC(child, local)
	}
}

// checkDynamicEDC enforces XSD 1.1's tighter Element Declarations Consistent
// rule for a wildcard-matched element: when its name also has a locally declared
// type in the content model, the element's actual governing type (after any
// xsi:type / substitution, recorded in res.Types) must be validly derived from
// that locally declared type — else the model would type the same name two
// inconsistent ways. saxon Wild062/063/064 et al.
func (a *assessor) checkDynamicEDC(el Element, local map[xsd.QName]xsd.Type) {
	lt, ok := local[el.Name()]
	if !ok || lt == nil {
		return
	}
	gov := a.res.Types[el]
	if gov == nil {
		return
	}
	if !xsdwalk.DerivationOK(gov, lt, 0) {
		a.addf(xsd.SpecCvcParticle, el.Pos(),
			"element %s matched by a wildcard is not consistent with its locally declared type", el.Name())
	}
}

// assessLocallyTyped assesses a wildcard-matched element governed by a locally
// declared type (no governing element declaration). xsi:type may override the
// type if validly derived from it (cvc-elt.4); there is no nillable/fixed/IDC
// context because the governing declaration is absent.
func (a *assessor) assessLocallyTyped(el Element, declared xsd.Type, inherited map[xsd.QName]string, parent Element) {
	gov := declared
	if typeStr, ok := a.xsiValue(el, "type"); ok {
		gov = a.resolveXSIType(el, nil, declared, typeStr)
	}
	a.res.Types[el] = gov
	a.assessType(el, gov, nil, childInherited(el, gov, inherited), parent)
}

// attrWildcardAllows applies the ##defined keyword exclusion (cvc-wildcard
// clause 2.2) on top of the namespace/notQName check for an attribute wildcard:
// a name that resolves to a global attribute declaration is excluded.
// (##definedSibling is not permitted on attribute wildcards, w-props-correct.5.)
func (a *assessor) attrWildcardAllows(w *xsd.Wildcard, name xsd.QName) bool {
	if !xsdwalk.WildcardAllows(w, name) {
		return false
	}
	for _, d := range w.NotQName {
		if d.Namespace == "" && d.Local == "##defined" && a.v.attrs[name] != nil {
			return false
		}
	}
	return true
}

// assessAttributes implements cvc-complex-type.3 / cvc-attribute / cvc-au.
func (a *assessor) assessAttributes(el Element, uses []*xsd.AttributeUse, wildcard *xsd.Wildcard) {
	seen := map[xsd.QName]bool{}
	for _, attr := range el.Attributes() {
		name := attr.Name()
		seen[name] = true
		switch name.Namespace {
		case xsd.XSINS:
			if d := a.v.attrs[name]; d != nil {
				a.validateAttrValue(el, attr, d.Type)
			}
			continue
		case xsd.XMLNSNS:
			continue
		}
		if u := xsdwalk.AttributeUse(uses, name); u != nil {
			a.validateAttrValue(el, attr, u.Decl.Type)
			a.checkAttrFixed(el, attr, u)
			continue
		}
		if wildcard != nil && a.attrWildcardAllows(wildcard, name) {
			d := a.v.attrs[name]
			switch wildcard.ProcessContents {
			case xsd.ProcessStrict:
				if d == nil {
					a.addf(xsd.SpecCvcAttribute, attr.Pos(), "no declaration for attribute %s matched by a strict wildcard", name)
					continue
				}
				a.validateAttrValue(el, attr, d.Type)
			case xsd.ProcessLax:
				if d != nil {
					a.validateAttrValue(el, attr, d.Type)
				}
			}
			continue
		}
		// spec: cvc-attribute — XSD 1.1 Part 1 §3.2.4 (xmlschema11-1.md#cvc-attribute)
		a.addf(xsd.SpecCvcComplexType, attr.Pos(), "attribute %s is not permitted here", name)
	}
	// cvc-complex-type.4: required attribute uses must be present.
	// An absent attribute use with a default {value constraint} contributes that
	// default to the [attribute] (cvc-complex-type.3 / §3.4.4.2); harvest any ID
	// it carries so default-supplied IDs resolve IDREFs (e.g. id_attr default).
	for _, u := range uses {
		if u.Decl == nil || seen[u.Decl.Name] {
			continue
		}
		if u.Required {
			a.addf(xsd.SpecCvcComplexType, el.Pos(), "required attribute %s is missing", u.Decl.Name)
			continue
		}
		def := u.Default
		if def == nil {
			def = u.Decl.Default
		}
		if def != nil {
			a.collectID(u.Decl.Type, *def, el.Pos(), el, el)
		}
	}
}

func (a *assessor) checkAttrFixed(el Element, attr Attribute, u *xsd.AttributeUse) {
	fixed := u.Fixed
	if fixed == nil {
		fixed = u.Decl.Fixed
	}
	if fixed == nil {
		return
	}
	if !a.valuesEqual(u.Decl.Type, attr.Value(), *fixed, el) {
		// spec: cvc-au — XSD 1.1 Part 1 §3.5.4 (xmlschema11-1.md#cvc-au)
		a.addf(xsd.SpecCvcAU, attr.Pos(), "attribute %s value %q does not match fixed value %q", attr.Name(), attr.Value(), *fixed)
	}
}

func (a *assessor) validateAttrValue(el Element, attr Attribute, t *xsd.SimpleType) {
	if t == nil {
		return
	}
	if a.attrType == nil {
		a.attrType = map[Attribute]*xsd.SimpleType{}
	}
	a.attrType[attr] = t
	if _, err := t.ParseValue(attr.Value(), nsContext{el}); err != nil {
		a.addValueErr(err, attr.Pos())
		return
	}
	a.collectID(t, attr.Value(), attr.Pos(), el, el)
}

// validateSimpleContent validates an element's character content against a
// simple type, then enforces a fixed value constraint (cvc-elt.5.2.2).
func (a *assessor) validateSimpleContent(el Element, t *xsd.SimpleType, text string, decl *xsd.ElementDecl, parent Element) {
	if t == nil {
		return
	}
	// cvc-elt.5.1.1: an ·empty· element with a {value constraint} (default OR
	// fixed) takes that value as its schema-normalized value (validated and
	// ID-harvested). The fixed-equality check below is then trivially satisfied.
	if text == "" && decl != nil && !hasElementChildren(el) {
		if decl.Default != nil {
			text = *decl.Default
		} else if decl.Fixed != nil {
			text = *decl.Fixed
		}
	}
	if _, err := t.ParseValue(text, nsContext{el}); err != nil {
		a.addValueErr(err, el.Pos())
		return
	}
	if decl != nil && decl.Fixed != nil {
		if !a.valuesEqual(t, text, *decl.Fixed, el) {
			a.addf(xsd.SpecCvcElt, el.Pos(), "element %s content %q does not match fixed value %q", decl.Name, text, *decl.Fixed)
		}
	}
	// An xs:ID in element simple content binds the value to the element's PARENT
	// (§3.3.4.5); attributes bind to their own element (see validateAttrValue).
	a.collectID(t, text, el.Pos(), el, parent)
}

// valuesEqual reports whether two lexical forms denote the same value in t.
func (a *assessor) valuesEqual(t *xsd.SimpleType, lexA, lexB string, el Element) bool {
	va, errA := t.ParseValue(lexA, nsContext{el})
	vb, errB := t.ParseValue(lexB, nsContext{el})
	if errA != nil || errB != nil {
		return false
	}
	o, ok := t.EffectiveCompare()(va, vb)
	return ok && o == xsd.OrderEqual
}

// collectID records xs:ID values (for uniqueness) and xs:IDREF references (for
// later resolution) found in a value, implementing the gathering half of cvc-id.
// ID/IDREF binding follows the actual validating type, so list items are
// harvested per item type and union values per the member type that validates
// them (xs:IDREFS, list-of-ID, union-of-ID all reduce to atomic ID/IDREF).
// ctx is the element whose value is being harvested (for namespace-scoped value
// parsing of union members); owner is the element to which a discovered xs:ID
// binds (the value's element for attributes, the PARENT element for simple
// content — nil at the validation root, which makes the binding empty).
func (a *assessor) collectID(t *xsd.SimpleType, lexical string, pos xsd.Pos, ctx, owner Element) {
	if t == nil {
		return
	}
	switch t.Variety {
	case xsd.VarietyList:
		if t.ItemType != nil {
			for _, item := range strings.Fields(lexical) {
				a.collectID(t.ItemType, item, pos, ctx, owner)
			}
		}
		return
	case xsd.VarietyUnion:
		// The actual {member type definition} is the first member that validates
		// the value — harvest ID/IDREF as that member would (cvc-id over §3.16.4).
		for _, m := range t.BasicMembers() {
			if _, err := m.ParseValue(lexical, nsContext{ctx}); err == nil {
				a.collectID(m, lexical, pos, ctx, owner)
				return
			}
		}
		return
	}
	switch {
	case xsdwalk.IsDerivedFrom(t, builtin.ID):
		val := strings.TrimSpace(lexical)
		if owner == nil {
			// cvc-id.1: an xs:ID whose [binding] is empty — e.g. element-content
			// ID on the validation root, whose parent is out of scope — is invalid.
			a.addf(xsd.SpecCvcID, pos, "ID value %q binds to no element in scope", val)
			return
		}
		if a.ids == nil {
			a.ids = map[string]Element{}
		}
		if prev, dup := a.ids[val]; dup {
			// cvc-id.2: the same value on the same element (e.g. two ID attributes,
			// or repeated list items) is allowed; only a clash between distinct
			// elements is an error.
			if prev != owner {
				a.addf(xsd.SpecCvcID, pos, "duplicate ID value %q", val)
			}
			return
		}
		a.ids[val] = owner
	case xsdwalk.IsDerivedFrom(t, builtin.IDREF):
		a.idrefs = append(a.idrefs, idref{val: strings.TrimSpace(lexical), pos: pos})
	case xsdwalk.IsDerivedFrom(t, builtin.ENTITY):
		// xs:ENTITY value space is restricted to the names of unparsed entities
		// declared in the DTD (Part 2 §3.4.13; enforced at the structures level).
		if a.haveEntities {
			val := strings.TrimSpace(lexical)
			if !a.entities[val] {
				a.addf(xsd.SpecDatatypeValid, pos, "ENTITY %q matches no declared unparsed entity", val)
			}
		}
	}
}

// checkIDRefs resolves every collected IDREF against the ID table (cvc-id.3).
func (a *assessor) checkIDRefs() {
	for _, r := range a.idrefs {
		if _, ok := a.ids[r.val]; !ok {
			a.addf(xsd.SpecCvcID, r.pos, "IDREF %q has no matching ID", r.val)
		}
	}
}

// ---- infoset helpers ----

func (a *assessor) xsiValue(el Element, local string) (string, bool) {
	for _, attr := range el.Attributes() {
		n := attr.Name()
		if n.Namespace == xsd.XSINS && n.Local == local {
			return attr.Value(), true
		}
	}
	return "", false
}

func parseXSIBool(s string) bool {
	switch strings.TrimSpace(s) {
	case "true", "1":
		return true
	}
	return false
}

// resolveQNameInScope resolves a lexical QName ("p:Local" or "Local") using el's
// in-scope namespaces.
func resolveQNameInScope(el Element, s string) (xsd.QName, error) {
	prefix, local := "", strings.TrimSpace(s)
	if i := strings.IndexByte(local, ':'); i >= 0 {
		prefix, local = local[:i], local[i+1:]
	}
	uri, ok := el.Lookup(prefix)
	if prefix != "" && !ok {
		return xsd.QName{}, errUndefinedPrefix
	}
	return xsd.QName{Namespace: uri, Local: local}, nil
}

var errUndefinedPrefix = &simpleErr{"undefined namespace prefix"}

type simpleErr struct{ s string }

func (e *simpleErr) Error() string { return e.s }

func elementChildren(el Element) []Element {
	var out []Element
	for _, n := range el.Children() {
		if e, ok := n.(Element); ok {
			out = append(out, e)
		}
	}
	return out
}

func hasElementChildren(el Element) bool {
	for _, n := range el.Children() {
		if _, ok := n.(Element); ok {
			return true
		}
	}
	return false
}

func charContent(el Element) string {
	var b strings.Builder
	for _, n := range el.Children() {
		if t, ok := n.(Text); ok {
			b.WriteString(t.Data())
		}
	}
	return b.String()
}

func hasContent(el Element) bool {
	return hasElementChildren(el) || hasNonWhitespace(charContent(el))
}

func hasNonWhitespace(s string) bool {
	return strings.TrimSpace(s) != ""
}

func typeName(t xsd.Type) xsd.QName {
	if t == nil {
		return xsd.QName{}
	}
	return t.TypeName()
}
