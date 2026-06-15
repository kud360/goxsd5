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

	ids    map[string]bool // declared xs:ID values, for uniqueness (cvc-id.2)
	idrefs []idref         // pending IDREF values, resolved after the walk

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
	decl := a.v.elements[root.Name()]
	if decl == nil {
		// spec: cvc-elt — XSD 1.1 Part 1 §3.3.4 (xmlschema11-1.md#cvc-elt)
		a.addf(xsd.SpecCvcElt, root.Pos(), "no global element declaration for root element %s", root.Name())
		return
	}
	a.assessElement(root, decl)
}

// assessElement implements cvc-elt (Element Locally Valid (Element), §3.3.4).
func (a *assessor) assessElement(el Element, decl *xsd.ElementDecl) {
	// cvc-elt.2: the declaration must not be abstract.
	if decl.Abstract {
		a.addf(xsd.SpecCvcElt, el.Pos(), "element %s is declared abstract and cannot appear in an instance", decl.Name)
	}

	gov := decl.Type
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
	a.assessType(el, gov, decl)

	// Identity constraints are scoped at this element and read its subtree.
	a.checkIdentityConstraints(el, decl)
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
	// {prohibited substitutions}).
	block := decl.Block
	if ct, ok := declared.(*xsd.ComplexType); ok {
		block |= ct.Block
	}
	if !xsdwalk.DerivationOK(t, declared, block) {
		a.addf(xsd.SpecCvcElt, el.Pos(), "xsi:type %s is not validly derived from the declared type of %s", tn, decl.Name)
		return declared
	}
	return t
}

// assessType implements cvc-type (§3.4.4): dispatch on simple vs complex.
func (a *assessor) assessType(el Element, t xsd.Type, decl *xsd.ElementDecl) {
	switch t := t.(type) {
	case *xsd.SimpleType:
		// cvc-type.3.1.1/.2: a simple-typed element admits no element children
		// and only xsi:* attributes.
		a.assessAttributes(el, nil, nil)
		if hasElementChildren(el) {
			a.addf(xsd.SpecCvcType, el.Pos(), "element %s has a simple type but contains element children", decl.Name)
		}
		a.validateSimpleContent(el, t, charContent(el), decl)
	case *xsd.ComplexType:
		a.assessComplexType(el, t, decl)
	}
}

// assessComplexType implements cvc-complex-type (§3.4.4).
func (a *assessor) assessComplexType(el Element, ct *xsd.ComplexType, decl *xsd.ElementDecl) {
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
		a.validateSimpleContent(el, content.Type, charContent(el), decl)
	case *xsd.ElementContent:
		a.assessElementContent(el, content)
	default: // empty content
		// cvc-complex-type.2.1: empty content admits neither element nor
		// non-whitespace character content.
		if hasElementChildren(el) || hasNonWhitespace(charContent(el)) {
			a.addf(xsd.SpecCvcComplexType, el.Pos(), "element %s has empty content type but is not empty", decl.Name)
		}
	}
}

// assessElementContent matches the children against the content model and
// recurses (cvc-complex-type.2.3/.4, cvc-particle).
func (a *assessor) assessElementContent(el Element, ec *xsd.ElementContent) {
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
	for i, k := range kids {
		a.assessChild(k, terms[i])
	}
}

// assessChild recurses into one matched child element.
func (a *assessor) assessChild(child Element, mt xsdwalk.MatchedTerm) {
	if mt.Elem != nil {
		a.assessElement(child, mt.Elem)
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
			a.assessElement(child, d)
		}
	case xsd.ProcessStrict:
		d := a.v.elements[child.Name()]
		if d == nil {
			// spec: cvc-wildcard — XSD 1.1 Part 1 §3.10.4 (xmlschema11-1.md#cvc-wildcard)
			a.addf(xsd.SpecCvcWildcard, child.Pos(), "no declaration for element %s matched by a strict wildcard", child.Name())
			return
		}
		a.assessElement(child, d)
	}
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
		if wildcard != nil && xsdwalk.WildcardAllows(wildcard, name) {
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
	for _, u := range uses {
		if u.Required && u.Decl != nil && !seen[u.Decl.Name] {
			a.addf(xsd.SpecCvcComplexType, el.Pos(), "required attribute %s is missing", u.Decl.Name)
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
	v, err := t.ParseValue(attr.Value(), nsContext{el})
	if err != nil {
		a.addValueErr(err, attr.Pos())
		return
	}
	a.collectID(t, v, attr.Value(), attr.Pos())
}

// validateSimpleContent validates an element's character content against a
// simple type, then enforces a fixed value constraint (cvc-elt.5.2.2).
func (a *assessor) validateSimpleContent(el Element, t *xsd.SimpleType, text string, decl *xsd.ElementDecl) {
	if t == nil {
		return
	}
	v, err := t.ParseValue(text, nsContext{el})
	if err != nil {
		a.addValueErr(err, el.Pos())
		return
	}
	if decl != nil && decl.Fixed != nil {
		if !a.valuesEqual(t, text, *decl.Fixed, el) {
			a.addf(xsd.SpecCvcElt, el.Pos(), "element %s content %q does not match fixed value %q", decl.Name, text, *decl.Fixed)
		}
	}
	a.collectID(t, v, text, el.Pos())
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
func (a *assessor) collectID(t *xsd.SimpleType, v xsd.Value, lexical string, pos xsd.Pos) {
	switch {
	case xsdwalk.IsDerivedFrom(t, builtin.ID):
		if a.ids == nil {
			a.ids = map[string]bool{}
		}
		val := strings.TrimSpace(lexical)
		if a.ids[val] {
			// spec: cvc-id — XSD 1.1 Part 1 §3.4.4 (xmlschema11-1.md#cvc-id)
			a.addf(xsd.SpecCvcID, pos, "duplicate ID value %q", val)
			return
		}
		a.ids[val] = true
	case xsdwalk.IsDerivedFrom(t, builtin.IDREF):
		a.idrefs = append(a.idrefs, idref{val: strings.TrimSpace(lexical), pos: pos})
	case t.Variety == xsd.VarietyList && t.ItemType != nil && xsdwalk.IsDerivedFrom(t.ItemType, builtin.IDREF):
		for _, item := range strings.Fields(lexical) {
			a.idrefs = append(a.idrefs, idref{val: item, pos: pos})
		}
	}
}

// checkIDRefs resolves every collected IDREF against the ID table (cvc-id.3).
func (a *assessor) checkIDRefs() {
	for _, r := range a.idrefs {
		if !a.ids[r.val] {
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
