package parser

// Schema assembly: build every global component of one validated document
// into an xsd.Schema. (M7 extends this over the transitive closure of
// includes/imports; redefine/override children are registered in the scoped
// registry but their cross-document semantics are not wired yet.)

import (
	"github.com/kud360/goxsd5/builtin"
	"github.com/kud360/goxsd5/xsd"
)

// buildSchema runs pass 2 over doc and returns the linked schema.
func buildSchema(reg *registry, doc *schemaDoc, errs *xsd.ErrorList) *xsd.Schema {
	b := newBuilder(reg, errs)
	s := &xsd.Schema{
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
	for _, comp := range doc.compositions {
		if comp.kind == "import" {
			s.Imports = append(s.Imports, comp.namespace)
		}
	}

	r := b.registryFor(doc)
	for _, c := range xsdElems(doc.root, doc) {
		name, _ := c.Attr("name")
		q := xsd.QName{Namespace: doc.targetNamespace, Local: name}
		// Build through the registry declaration so that memoization and
		// cycle marks are shared with reference resolution. A node that is
		// not the registered declaration is a duplicate (already reported);
		// it is skipped rather than built.
		current := func(space space) *decl {
			d := r.lookup(space, q)
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
			s.Notations[q] = b.buildNotation(c, doc)
		case "annotation":
			s.Annotations = append(s.Annotations, buildAnnotation(c, doc))
		}
	}
	b.checkTypeCycles(s)
	return s
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
