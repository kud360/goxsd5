package parser

// src-redefine clauses 6.2 / 7.2 (§4.2.4): a no-self-reference <group> or
// <attributeGroup> redefinition must be a *restriction* of the original
// definition it redefines — the new model group accepts a subset of the
// sequences the original does (6.2.2), and the new attribute uses are a subset
// of the original's (7.2.2, via clause 3 of derivation-ok-restriction).
//
// The relation needs both components compiled, so the loader records the
// (new, original) node pairs in redefineRestrict and the builder decides them
// here, after every group, attribute group, and type is built.

import (
	"github.com/kud360/goxsd5/parser/xmltree"
	"github.com/kud360/goxsd5/xsd"
)

// checkRedefineRestrict decides the recorded no-self-reference group/
// attributeGroup redefinitions. A group's model-group subset relation reuses the
// cos-particle-restrict engine (particleRestrictUnified) by viewing the original
// and the redefinition as a base/restriction pair; an attribute group's relation
// reuses the attribute-use subset of derivation-ok-restriction. Both engines are
// false-positive-free: a relation they cannot decide is accepted, exactly as the
// spec permits.
func (b *builder) checkRedefineRestrict(accepted func(*xsd.ElementDecl) map[xsd.QName]bool, globalsByName map[xsd.QName]*xsd.ElementDecl) {
	for _, p := range b.redefineRestrict {
		switch p.kind {
		case "group":
			b.checkRedefineGroupRestrict(p, accepted, globalsByName)
		case "attributeGroup":
			b.checkRedefineAttrGroupRestrict(p)
		}
	}
}

// checkRedefineGroupRestrict enforces src-redefine clause 6.2.2: the redefining
// model group accepts a subset of the sequences the original accepts. It wraps
// each named model group as the content of a synthetic complex type — original
// as base, redefinition as a restriction of it — and runs particleRestrictUnified,
// then re-emits any finding as a single src-redefine violation. The engine gives
// up (reports nothing) on content models outside its tractable flat shape, so an
// undecidable redefinition is accepted, never falsely flagged.
func (b *builder) checkRedefineGroupRestrict(p redefinePair, accepted func(*xsd.ElementDecl) map[xsd.QName]bool, globalsByName map[xsd.QName]*xsd.ElementDecl) {
	newG := b.buildGroup(&decl{name: groupName(p.newNode, p.newDoc), pos: p.pos, node: p.newNode, doc: p.newDoc})
	origG := b.buildGroup(&decl{name: groupName(p.origNode, p.origDoc), pos: p.origNode.Pos, node: p.origNode, doc: p.origDoc})
	if newG == nil || origG == nil || newG.Group == nil || origG.Group == nil {
		return // an empty redefinition restricts any original
	}
	bct := &xsd.ComplexType{Name: origG.Name, Pos: origG.Pos, Content: groupContent(origG.Group)}
	ct := &xsd.ComplexType{
		Name:             newG.Name,
		Pos:              p.pos,
		BaseType:         bct,
		DerivationMethod: xsd.DeriveRestriction,
		Content:          groupContent(newG.Group),
	}
	if vs := b.particleRestrictUnified(ct, accepted, globalsByName); len(vs) > 0 {
		// spec: src-redefine.6.2.2 — XSD 1.1 Part 1 §4.2.4
		b.errf(xsd.SpecSrcRedefine, p.pos, "redefined group %s must be a restriction of the original: %s", newG.Name, vs[0].msg)
	}
}

// groupContent wraps a named group's model group as element content occurring
// exactly once, the shape the particle-restriction engine expects.
func groupContent(mg *xsd.ModelGroup) *xsd.ElementContent {
	return &xsd.ElementContent{Particle: &xsd.Particle{MinOccurs: 1, MaxOccurs: 1, Term: mg, Pos: mg.Pos}}
}

// groupName returns the expanded name of a top-level <group>/<attributeGroup>
// node in its document.
func groupName(n *xmltree.Node, doc *schemaDoc) xsd.QName {
	name, _ := n.Attr("name")
	return xsd.QName{Namespace: doc.targetNamespace, Local: name}
}

// checkRedefineAttrGroupRestrict enforces src-redefine clause 7.2.2: the
// redefining attribute group's {attribute uses} (and {attribute wildcard}) are a
// restriction of the original's per clause 3 of derivation-ok-restriction. Per
// the §4.2.4 note the redefinition's uses are exactly those of its explicit
// <attribute> children — nothing is inherited from the original — so the subset
// test compares the two compiled attribute-group definitions directly: every use
// the redefinition allows must be allowed by the original (matched by name and a
// validly-derived type, or by the original's wildcard), and every use the
// original requires must remain required.
func (b *builder) checkRedefineAttrGroupRestrict(p redefinePair) {
	newG := b.buildAttributeGroup(&decl{name: groupName(p.newNode, p.newDoc), pos: p.pos, node: p.newNode, doc: p.newDoc})
	origG := b.buildAttributeGroup(&decl{name: groupName(p.origNode, p.origDoc), pos: p.origNode.Pos, node: p.origNode, doc: p.origDoc})
	if newG == nil || origG == nil {
		return
	}
	origByName := map[xsd.QName]*xsd.AttributeUse{}
	for _, u := range origG.Uses {
		if u.Decl != nil {
			origByName[u.Decl.Name] = u
		}
	}
	// Every use the redefinition allows must be allowed by the original.
	for _, u := range newG.Uses {
		if u.Decl == nil {
			continue
		}
		base, ok := origByName[u.Decl.Name]
		if !ok {
			if origG.Wildcard != nil && origG.Wildcard.AllowsName(u.Decl.Name) {
				continue // the original's wildcard admits this attribute
			}
			// spec: src-redefine.7.2.2 — XSD 1.1 Part 1 §4.2.4 (derivation-ok-restriction clause 3)
			b.errf(xsd.SpecSrcRedefine, u.Pos, "redefined attribute group %s adds attribute %s the original does not allow", newG.Name, u.Decl.Name)
			continue
		}
		if u.Decl.Type != nil && base.Decl.Type != nil && !validlyDerivedByRestriction(u.Decl.Type, base.Decl.Type) {
			// spec: src-redefine.7.2.2 — XSD 1.1 Part 1 §4.2.4
			b.errf(xsd.SpecSrcRedefine, u.Pos, "attribute %s in redefined attribute group %s has a type not derived from the original's", u.Decl.Name, newG.Name)
		}
	}
	// Every attribute the original requires must remain required in the
	// redefinition; dropping or relaxing it would admit instances the original
	// rejects.
	for _, u := range origG.Uses {
		if u.Decl == nil || !u.Required {
			continue
		}
		nu := findUse(newG.Uses, u.Decl.Name)
		if nu == nil || !nu.Required {
			// spec: src-redefine.7.2.2 — XSD 1.1 Part 1 §4.2.4
			b.errf(xsd.SpecSrcRedefine, p.pos, "redefined attribute group %s no longer requires attribute %s the original requires", newG.Name, u.Decl.Name)
		}
	}
}

// findUse returns the attribute use of the given name in uses, or nil.
func findUse(uses []*xsd.AttributeUse, name xsd.QName) *xsd.AttributeUse {
	for _, u := range uses {
		if u.Decl != nil && u.Decl.Name == name {
			return u
		}
	}
	return nil
}
