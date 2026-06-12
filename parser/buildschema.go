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
