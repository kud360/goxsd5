package parser

// Pass 2: build schema components from the registered nodes, on demand with
// memoization. Reference resolution goes through the owning document's
// scoped registry (redefine/override shadowing, then globals, then
// builtins). A node found in-progress is a cyclic definition (the back-edge
// of the implicit topological sort).
//
// Error recovery: every build function records its errors and returns a
// usable fallback component (xs:anySimpleType / xs:anyType / a zero-value
// declaration) so one parse reports as many independent problems as
// possible.

import (
	"strings"

	"github.com/kud360/goxsd5/builtin"
	"github.com/kud360/goxsd5/parser/xmltree"
	"github.com/kud360/goxsd5/xsd"
)

type builder struct {
	reg  *registry
	errs *xsd.ErrorList

	// Memoization, keyed by defining node. Anonymous components occur at
	// exactly one node, so the key covers globals and locals alike.
	types      map[*xmltree.Node]xsd.Type
	elements   map[*xmltree.Node]*xsd.ElementDecl
	attributes map[*xmltree.Node]*xsd.AttributeDecl
	groups     map[*xmltree.Node]*xsd.Group
	attrGroups map[*xmltree.Node]*xsd.AttributeGroup
	notations  map[*xmltree.Node]*xsd.Notation
	ics        map[*xmltree.Node]*xsd.IdentityConstraint

	// building marks nodes whose construction is on the stack.
	building map[*xmltree.Node]bool

	// pendingAttrs holds each complex type's own attribute material until
	// finishComplexTypes merges it with the base's (the base may still be
	// mid-build when the type is constructed).
	pendingAttrs map[*xsd.ComplexType]*pendingAttrs
}

// pendingAttrs is the attribute material declared directly on one complex
// type, recorded during construction and merged with the (by then complete)
// base type in the finishComplexTypes post-pass.
type pendingAttrs struct {
	own         []*xsd.AttributeUse
	wc          *xsd.Wildcard
	prohibited  map[xsd.QName]bool
	override    bool // restriction semantics: own uses shadow base uses
	wcFallback  bool // inherit the base wildcard when none is declared
	pos         xsd.Pos
	node        *xmltree.Node // the complexType element (defaultAttributesApply)
	contentNode *xmltree.Node // the <restriction>/<extension> element, if any
	doc         *schemaDoc
}

func newBuilder(reg *registry, errs *xsd.ErrorList) *builder {
	return &builder{
		reg:          reg,
		errs:         errs,
		types:        map[*xmltree.Node]xsd.Type{},
		elements:     map[*xmltree.Node]*xsd.ElementDecl{},
		attributes:   map[*xmltree.Node]*xsd.AttributeDecl{},
		groups:       map[*xmltree.Node]*xsd.Group{},
		attrGroups:   map[*xmltree.Node]*xsd.AttributeGroup{},
		notations:    map[*xmltree.Node]*xsd.Notation{},
		ics:          map[*xmltree.Node]*xsd.IdentityConstraint{},
		building:     map[*xmltree.Node]bool{},
		pendingAttrs: map[*xsd.ComplexType]*pendingAttrs{},
	}
}

func (b *builder) errf(ref xsd.SpecRef, pos xsd.Pos, format string, args ...any) {
	b.errs.Addf(ref, pos, format, args...)
}

// registryFor returns the registry visible from inside doc: its scoped one
// when present (it chains to the global), else the global registry.
func (b *builder) registryFor(doc *schemaDoc) *registry {
	if doc != nil && doc.scoped != nil {
		return doc.scoped
	}
	return b.reg
}

// lookupRef resolves a component reference from within doc, enforcing that
// the referenced namespace is reachable: the target namespace, the XSD
// namespace (built-ins are predefined everywhere), or an imported one. It
// reports src-resolve and returns nil on failure.
func (b *builder) lookupRef(s space, q xsd.QName, p xsd.Pos, doc *schemaDoc) *decl {
	if doc != nil && q.Namespace != doc.targetNamespace && q.Namespace != xsd.XSDNS && q.Namespace != xsd.XSINS && !doc.importedNS[q.Namespace] {
		// spec: src-resolve.4 — XSD 1.1 Part 1 §3.15.3: a reference may only
		// reach the target namespace or a namespace named by an import.
		b.errf(xsd.SpecSrcResolve, p, "%s %s references namespace %q, which is not imported here", s, q, q.Namespace)
		return nil
	}
	d := b.registryFor(doc).lookup(s, q)
	if d == nil {
		// spec: src-resolve — XSD 1.1 Part 1 §3.15.3 (src-resolve)
		b.errf(xsd.SpecSrcResolve, p, "%s %s is not declared", s, q)
		return nil
	}
	return d
}

// resolveType resolves a type QName from within doc. A failed resolution
// reports src-resolve and returns xs:anyType (never nil).
func (b *builder) resolveType(q xsd.QName, p xsd.Pos, doc *schemaDoc) xsd.Type {
	d := b.lookupRef(spaceType, q, p, doc)
	if d == nil {
		return builtin.AnyType
	}
	if d.builtin != nil {
		return d.builtin
	}
	return b.buildTypeDecl(d)
}

// resolveSimpleType resolves a type QName that must name a simple type.
// ref is the constraint violated when it names a complex type.
func (b *builder) resolveSimpleType(q xsd.QName, p xsd.Pos, doc *schemaDoc, ref xsd.SpecRef) *xsd.SimpleType {
	t := b.resolveType(q, p, doc)
	if st, ok := t.(*xsd.SimpleType); ok {
		return st
	}
	if t != builtin.AnyType { // suppress the double error after src-resolve
		b.errf(ref, p, "%s is a complex type; a simple type is required", q)
	}
	return builtin.AnySimpleType
}

// buildTypeDecl builds the type for a registry declaration. Simple types
// carry an in-progress mark: their facet construction requires a completed
// base, so a base-chain re-entry must fail eagerly. Complex types memoize a
// shell before building content instead — content references back into an
// unfinished type are legal (only derivation edges form illegal cycles,
// detected by checkTypeCycles after assembly).
func (b *builder) buildTypeDecl(d *decl) xsd.Type {
	if t, ok := b.types[d.node]; ok {
		return t
	}
	switch d.node.Name.Local {
	case "simpleType":
		if b.building[d.node] {
			// spec: st-props-correct.2 — XSD 1.1 Part 2 §4.1.6: a simple
			// type must not be its own ancestor.
			b.errf(xsd.SpecSTPropsCorrect, d.pos, "simple type %s is part of a cyclic definition", d.name)
			return builtin.AnySimpleType
		}
		b.building[d.node] = true
		defer delete(b.building, d.node)
		t := b.buildSimpleType(d.node, d.doc, d.name)
		b.types[d.node] = t
		return t
	case "complexType":
		return b.buildComplexType(d.node, d.doc, d.name)
	}
	// Unreachable: only type nodes are registered in spaceType.
	return builtin.AnyType
}

// buildAnonType builds the inline type definition child of n (already
// validated to be at most one), or returns def when there is none.
func (b *builder) buildAnonType(n *xmltree.Node, doc *schemaDoc, def xsd.Type) xsd.Type {
	for _, c := range xsdElems(n, doc) {
		switch c.Name.Local {
		case "simpleType":
			return b.memoType(c, func() xsd.Type { return b.buildSimpleType(c, doc, xsd.QName{}) })
		case "complexType":
			return b.buildComplexType(c, doc, xsd.QName{}) // self-memoizing
		}
	}
	return def
}

// memoType memoizes an anonymous type construction (inline types can be
// reached only once, but memoizing keeps the invariant simple).
func (b *builder) memoType(n *xmltree.Node, f func() xsd.Type) xsd.Type {
	if t, ok := b.types[n]; ok {
		return t
	}
	b.building[n] = true
	defer delete(b.building, n)
	t := f()
	b.types[n] = t
	return t
}

// xsdElems returns n's XSD-namespace children, skipping any pruned by
// conditional inclusion.
func xsdElems(n *xmltree.Node, doc *schemaDoc) []*xmltree.Node {
	var out []*xmltree.Node
	for _, c := range n.Children {
		if c.Name.Space == xsd.XSDNS && (doc == nil || !doc.pruned[c]) {
			out = append(out, c)
		}
	}
	return out
}

// firstChild returns the first XSD child with one of the given local names.
func firstChild(n *xmltree.Node, doc *schemaDoc, locals ...string) *xmltree.Node {
	for _, c := range xsdElems(n, doc) {
		for _, l := range locals {
			if c.Name.Local == l {
				return c
			}
		}
	}
	return nil
}

// nsContext adapts a node's in-scope namespaces to xsd.ValueContext for
// QName/NOTATION value parsing.
type nsContext struct{ n *xmltree.Node }

func (c nsContext) ResolveQName(prefix, local string) (xsd.QName, bool) {
	if prefix == "" {
		uri, _ := c.n.NS.Lookup("")
		return xsd.QName{Namespace: uri, Local: local}, true
	}
	uri, ok := c.n.NS.Lookup(prefix)
	if !ok {
		return xsd.QName{}, false
	}
	return xsd.QName{Namespace: uri, Local: local}, true
}

// qnameAttr resolves a QName-valued attribute; ok is false when absent or
// unresolvable (already reported by pass 1). Component references inside a
// chameleon-included document are remapped to the absorbed namespace.
func qnameAttr(n *xmltree.Node, doc *schemaDoc, name string) (xsd.QName, bool) {
	v, ok := n.Attr(name)
	if !ok {
		return xsd.QName{}, false
	}
	q, err := n.ResolveQName(strings.TrimSpace(v))
	if err != nil {
		return xsd.QName{}, false
	}
	return chameleonQName(q, doc), true
}

// chameleonQName applies the chameleon-include transformation to a component
// reference: inside an absorbed document, a reference resolving to no
// namespace means the absorbed (= target) namespace.
// spec: src-include.2.2 — XSD 1.1 Part 1 §4.2.3 (src-include)
func chameleonQName(q xsd.QName, doc *schemaDoc) xsd.QName {
	if q.Namespace == "" && doc != nil && doc.chameleonNS != "" {
		q.Namespace = doc.chameleonNS
	}
	return q
}

func boolAttr(n *xmltree.Node, name string, def bool) bool {
	if v, ok := n.Attr(name); ok {
		if parsed, err := parseBool(v); err == nil {
			return parsed
		}
	}
	return def
}

// derivAttr parses a block/final attribute with the given vocabulary,
// defaulting to def (the schema-level *Default value, already restricted
// to the right vocabulary by intersection).
func derivAttr(n *xmltree.Node, name string, allowed xsd.DerivationSet, def xsd.DerivationSet) xsd.DerivationSet {
	v, ok := n.Attr(name)
	if !ok {
		return def & allowed
	}
	set, err := parseDerivationSet(v, allowed)
	if err != nil {
		return def & allowed
	}
	return set
}
