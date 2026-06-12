package parser

// The component registry of pass 1: global declarations by symbol space and
// expanded name, still dangling (a node plus its owning document). Pass 2
// resolves references through it and builds components on demand. Redefine/
// override children go into a per-document scoped registry chained over the
// global one, so replacements shadow without mutating the global view.

import (
	"github.com/kud360/goxsd5/builtin"
	"github.com/kud360/goxsd5/parser/xmltree"
	"github.com/kud360/goxsd5/xsd"
)

// space is a symbol space per XSD 1.1 Part 1 §2.5: each kind of named
// component has its own. Type definitions (simple and complex) share one.
type space int

const (
	spaceType space = iota
	spaceElement
	spaceAttribute
	spaceGroup
	spaceAttrGroup
	spaceNotation
	spaceIC // identity constraints are named components in XSD 1.1
	numSpaces
)

func (s space) String() string {
	switch s {
	case spaceType:
		return "type"
	case spaceElement:
		return "element"
	case spaceAttribute:
		return "attribute"
	case spaceGroup:
		return "group"
	case spaceAttrGroup:
		return "attribute group"
	case spaceNotation:
		return "notation"
	case spaceIC:
		return "identity constraint"
	}
	return "?"
}

// decl is one global declaration: either a dangling node from a schema
// document, or a prebuilt component (built-in types).
type decl struct {
	name xsd.QName
	pos  xsd.Pos
	node *xmltree.Node // nil for builtins
	doc  *schemaDoc    // nil for builtins
	// builtin is the prebuilt component for the seeded entries.
	builtin xsd.Type
}

// registry maps (space, QName) → declaration. parent chains a scoped
// registry (redefine/override) over the global one.
type registry struct {
	parent *registry
	decls  [numSpaces]map[xsd.QName]*decl
}

// newRegistry returns a registry seeded with the built-in types: xs:anyType
// and every built-in simple type (xs:anySimpleType and xs:anyAtomicType
// included).
func newRegistry() *registry {
	r := &registry{}
	r.add(spaceType, &decl{name: builtin.AnyType.Name, builtin: builtin.AnyType})
	for _, t := range builtin.AllBuiltins() {
		r.add(spaceType, &decl{name: t.Name, builtin: t})
	}
	return r
}

// scope returns an empty registry shadowing r.
func (r *registry) scope() *registry { return &registry{parent: r} }

func (r *registry) add(s space, d *decl) {
	if r.decls[s] == nil {
		r.decls[s] = map[xsd.QName]*decl{}
	}
	r.decls[s][d.name] = d
}

// lookup resolves a name through the scope chain.
func (r *registry) lookup(s space, q xsd.QName) *decl {
	for reg := r; reg != nil; reg = reg.parent {
		if d := reg.decls[s][q]; d != nil {
			return d
		}
	}
	return nil
}

// register enters d unless the same name is already declared in this exact
// registry (not the chain: a scoped redefinition legitimately shadows its
// parent, and builtins may not be re-declared).
func (r *registry) register(s space, d *decl, errs *xsd.ErrorList) {
	if prev := r.decls[s][d.name]; prev != nil {
		// spec: sch-props-correct.2 — XSD 1.1 Part 1 §3.17.6: no two global
		// components of the same kind may share an expanded name.
		err := xsd.NewError(xsd.SpecSchemaPropsCorrect, d.pos, "duplicate %s %s", s, d.name)
		err.OtherPos = prev.pos
		errs.Add(err)
		return
	}
	r.add(s, d)
}

// globalSpaces maps top-level element local names to their symbol space.
var globalSpaces = map[string]space{
	"simpleType":     spaceType,
	"complexType":    spaceType,
	"element":        spaceElement,
	"attribute":      spaceAttribute,
	"group":          spaceGroup,
	"attributeGroup": spaceAttrGroup,
	"notation":       spaceNotation,
}

// registerDoc enters every global declaration of doc into reg, the
// redefine/override children into doc.scoped, and every named identity
// constraint (wherever it appears) into the identity-constraint space.
func registerDoc(reg *registry, doc *schemaDoc, errs *xsd.ErrorList) {
	doc.scoped = reg.scope()
	for _, c := range doc.root.Children {
		if doc.pruned[c] || c.Name.Space != xsd.XSDNS {
			continue
		}
		if s, ok := globalSpaces[c.Name.Local]; ok {
			registerGlobal(reg, s, c, doc, errs)
			continue
		}
		if c.Name.Local == "redefine" || c.Name.Local == "override" {
			for _, rc := range c.Children {
				if doc.pruned[rc] || rc.Name.Space != xsd.XSDNS {
					continue
				}
				if s, ok := globalSpaces[rc.Name.Local]; ok {
					registerGlobal(doc.scoped, s, rc, doc, errs)
				}
			}
		}
	}
	collectICs(reg, doc.root, doc, errs)
}

func registerGlobal(into *registry, s space, n *xmltree.Node, doc *schemaDoc, errs *xsd.ErrorList) {
	name, ok := n.Attr("name")
	if !ok {
		// Already reported as a missing required attribute.
		return
	}
	into.register(s, &decl{
		name: xsd.QName{Namespace: doc.targetNamespace, Local: name},
		pos:  n.Pos,
		node: n,
		doc:  doc,
	}, errs)
}

// collectICs registers named unique/key/keyref definitions. They are
// reachable only inside element declarations, but collecting from the whole
// tree is simpler and equivalent.
func collectICs(reg *registry, n *xmltree.Node, doc *schemaDoc, errs *xsd.ErrorList) {
	if doc.pruned[n] || n.Name.Space != xsd.XSDNS {
		return
	}
	switch n.Name.Local {
	case "unique", "key", "keyref":
		if name, ok := n.Attr("name"); ok {
			reg.register(spaceIC, &decl{
				name: xsd.QName{Namespace: doc.targetNamespace, Local: name},
				pos:  n.Pos,
				node: n,
				doc:  doc,
			}, errs)
		}
		return
	case "appinfo", "documentation":
		return // free content; nothing schema-relevant below
	}
	for _, c := range n.Children {
		collectICs(reg, c, doc, errs)
	}
}
