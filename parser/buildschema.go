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
	b.finishComplexTypes()
	return s
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
		b.applyDefaultAttributes(ct, p.node, p.doc)
		b.checkAttrUses(ct)
	}
	for _, t := range b.types {
		if ct, ok := t.(*xsd.ComplexType); ok {
			finish(ct)
		}
	}
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
		if bc.Particle == nil {
			return
		}
		if ec.Particle == nil {
			ec.Particle = bc.Particle
		} else {
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
