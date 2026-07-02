package jsonsrc

import (
	"fmt"

	"github.com/kud360/goxsd5/xsd"
	"github.com/kud360/goxsd5/xsdvalidate"
)

// buildRoot maps the top-level JSON value into the root element. The top-level
// must be a single-member object whose key names the root element declaration;
// the member value is the root element's content.
func (b *builder) buildRoot(v jvalue) (*element, error) {
	obj, ok := v.(*jobject)
	if !ok {
		return nil, fmt.Errorf("jsonsrc: top-level JSON value must be an object naming the root element")
	}
	members := structuralMembers(obj)
	if len(members) != 1 {
		return nil, fmt.Errorf("jsonsrc: top-level object must have exactly one member (the root element), got %d", len(members))
	}
	m := members[0]
	decl := b.resolveGlobal(m.key)
	if decl == nil {
		return nil, fmt.Errorf("jsonsrc: no global element declaration named %q", m.key)
	}
	return b.buildElement(decl.Name, decl, m.value, nsScope{}), nil
}

// resolveGlobal finds a global element declaration for an unprefixed root key.
// A no-namespace declaration is preferred (exact match); otherwise the sole
// global with that local name across namespaces is used, so a schema with a
// single target namespace resolves without the author restating it.
func (b *builder) resolveGlobal(local string) *xsd.ElementDecl {
	if d := b.schema.ElementByName(xsd.QName{Local: local}); d != nil {
		return d
	}
	return b.schema.ElementByLocal(local)
}

// buildElement constructs an element named name (governed by decl, which may be
// nil for a wildcard/xsi:type-only child), mapping content from v within the
// namespace scope parent.
func (b *builder) buildElement(name xsd.QName, decl *xsd.ElementDecl, v jvalue, parent nsScope) *element {
	el := &element{name: name, ns: nsScope{parent: &parent, def: name.Namespace, hasDef: true}}
	switch val := v.(type) {
	case *jnull:
		el.pos = val.pos
		el.attrs = append(el.attrs, attribute{
			name:  xsd.QName{Namespace: xsd.XSINS, Local: "nil"},
			value: "true",
			pos:   val.pos,
		})
		return el
	case *jscalar:
		el.pos = val.pos
		el.kids = append(el.kids, text{val.lexical})
		return el
	case *jobject:
		el.pos = val.pos
		b.fillObject(el, decl, val)
		return el
	case *jarray:
		// An array directly as an element value is only meaningful as repeated
		// children under a member key; as a bare element value it has no
		// schema-driven mapping, so treat it as empty content.
		el.pos = val.pos
		return el
	}
	return el
}

// fillObject populates el's attributes and children from a JSON object,
// classifying each structural member against decl's complex type.
func (b *builder) fillObject(el *element, decl *xsd.ElementDecl, obj *jobject) {
	b.applyDirectives(el, obj)
	ct := complexTypeOf(decl)
	cls := classify(ct)
	for _, m := range structuralMembers(obj) {
		b.mapMember(el, cls, m)
	}
}

// mapMember maps one structural object member to an attribute or child
// element(s) of el.
func (b *builder) mapMember(el *element, cls classifier, m jmember) {
	if cls.collides(m.key) {
		b.warnf(m.pos, "member %q matches both an attribute use and a child element; element wins", m.key)
	}
	if childDecl, ok := cls.children[m.key]; ok {
		b.appendChildren(el, childDecl, m.value)
		return
	}
	if attrDecl, ok := cls.attrs[m.key]; ok {
		b.appendAttribute(el, attrDecl, m)
		return
	}
	// Unknown key: surface it as a child element in the element's target
	// namespace so the engine reports the content-model violation (rather than
	// silently dropping it).
	b.appendChildren(el, &xsd.ElementDecl{Name: xsd.QName{Namespace: el.name.Namespace, Local: m.key}}, m.value)
}

// appendAttribute adds one attribute; a null-valued attribute member is omitted
// (JSON null on an attribute means "not supplied").
func (b *builder) appendAttribute(el *element, decl *xsd.AttributeDecl, m jmember) {
	sc, ok := m.value.(*jscalar)
	if !ok {
		// A non-scalar value for an attribute key (object/array/null) is not a
		// valid attribute value; drop it so the engine reports the missing
		// required attribute if any.
		return
	}
	el.attrs = append(el.attrs, attribute{
		name:  xsd.QName{Namespace: decl.Name.Namespace, Local: decl.Name.Local},
		value: sc.lexical,
		pos:   sc.pos,
	})
}

// appendChildren adds one child element (scalar/object/null) or, for an array
// value, one child per item (repeated children).
func (b *builder) appendChildren(el *element, decl *xsd.ElementDecl, v jvalue) {
	if arr, ok := v.(*jarray); ok {
		for _, item := range arr.items {
			el.kids = append(el.kids, b.buildElement(decl.Name, decl, item, el.ns))
		}
		return
	}
	el.kids = append(el.kids, b.buildElement(decl.Name, decl, v, el.ns))
}

// applyDirectives reads the reserved $type and $xmlns keys off obj and applies
// them to el (xsi:type attribute and namespace bindings respectively).
func (b *builder) applyDirectives(el *element, obj *jobject) {
	for _, m := range obj.members {
		switch m.key {
		case keyType:
			if sc, ok := m.value.(*jscalar); ok {
				el.attrs = append(el.attrs, attribute{
					name:  xsd.QName{Namespace: xsd.XSINS, Local: "type"},
					value: sc.lexical,
					pos:   sc.pos,
				})
			}
		case keyXMLNS:
			b.applyXMLNS(el, m.value)
		}
	}
}

// applyXMLNS seeds el's namespace scope from a $xmlns object of prefix->uri.
func (b *builder) applyXMLNS(el *element, v jvalue) {
	obj, ok := v.(*jobject)
	if !ok {
		return
	}
	if el.ns.prefixes == nil {
		el.ns.prefixes = map[string]string{}
	}
	for _, m := range obj.members {
		if sc, ok := m.value.(*jscalar); ok {
			el.ns.prefixes[m.key] = sc.lexical
		}
	}
}

// structuralMembers returns obj's members with reserved directive keys removed,
// in source order.
func structuralMembers(obj *jobject) []jmember {
	out := make([]jmember, 0, len(obj.members))
	for _, m := range obj.members {
		if m.key == keyType || m.key == keyXMLNS {
			continue
		}
		out = append(out, m)
	}
	return out
}

// complexTypeOf returns decl's complex type, or nil when decl is nil or has
// simple/absent content (no members to classify against).
func complexTypeOf(decl *xsd.ElementDecl) *xsd.ComplexType {
	if decl == nil {
		return nil
	}
	ct, _ := decl.Type.(*xsd.ComplexType)
	return ct
}

var _ xsdvalidate.Element = (*element)(nil)
