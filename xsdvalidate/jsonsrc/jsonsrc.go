// Package jsonsrc adapts a JSON document to the xsdvalidate abstract infoset, so
// a JSON instance can be assessed against a compiled schema set. It is the JSON
// analogue of xsdvalidate/xmlsrc (the XML source); the validation engine itself
// never imports a JSON package.
//
// JSON has no attribute-vs-element distinction and this adapter deliberately
// avoids an @-prefix convention. It is instead schema-aware: for a JSON object
// standing for an element of a known complex type, each member key is matched by
// local name against that type, resolving to either a named attribute use or a
// named child element declaration (child elements reachable through the content
// model, group references resolved). The engine then assesses the resulting
// infoset unchanged.
//
// Mapping rules (see issue #39 for the resolved design):
//   - Top-level JSON is a single-member object whose key names the root element.
//   - A member with a JSON scalar value is the element's simple character
//     content ({"size":10} -> <size>10</size>); numeric token text is preserved
//     (5.0 stays "5.0") so value-space checks see the authored lexical form.
//   - A member with a JSON object value carries attributes and child elements
//     only; such an element is assumed to have no character content.
//   - A member with a JSON array value is repeated children under that key.
//   - A member with JSON null is xsi:nil (valid only where the decl is
//     nillable); the adapter synthesizes an xsi:nil="true" attribute the engine
//     already understands.
//   - Reserved keys: "$type" sets xsi:type; "$xmlns" is an object of
//     prefix->namespace bindings for QName-valued content (inherited by
//     descendants). The default namespace is the matched declaration's own
//     target namespace, so unprefixed QName content resolves.
//
// Object member order is preserved (encoding/json Decoder.Token streaming), so
// sequence content models see children in document order.
package jsonsrc

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/kud360/goxsd5/xsd"
	"github.com/kud360/goxsd5/xsdvalidate"
	"github.com/kud360/goxsd5/xsdwalk"
)

// Reserved member keys. They are structural directives, never element or
// attribute names.
const (
	keyType  = "$type"  // sets xsi:type on the enclosing element
	keyXMLNS = "$xmlns" // { prefix: namespace-uri } bindings for QName content
)

// element is a JSON-backed xsdvalidate.Element.
type element struct {
	name  xsd.QName
	attrs []attribute
	kids  []xsdvalidate.Node
	ns    nsScope
	pos   xsd.Pos
}

func (e *element) Name() xsd.QName                { return e.name }
func (e *element) Pos() xsd.Pos                   { return e.pos }
func (e *element) Lookup(p string) (string, bool) { return e.ns.lookup(p) }

func (e *element) Attributes() []xsdvalidate.Attribute {
	out := make([]xsdvalidate.Attribute, len(e.attrs))
	for i := range e.attrs {
		out[i] = e.attrs[i]
	}
	return out
}

func (e *element) Children() []xsdvalidate.Node { return e.kids }

type attribute struct {
	name  xsd.QName
	value string
	pos   xsd.Pos
}

func (a attribute) Name() xsd.QName { return a.name }
func (a attribute) Value() string   { return a.value }
func (a attribute) Pos() xsd.Pos    { return a.pos }

type text struct{ s string }

func (t text) Data() string { return t.s }

// nsScope is an immutable, chained prefix->namespace environment. The empty
// prefix resolves to the enclosing element's target namespace; other prefixes
// come from $xmlns bindings, with inner scopes shadowing outer ones.
type nsScope struct {
	parent   *nsScope
	prefixes map[string]string
	def      string // default-namespace override for this scope
	hasDef   bool
}

// lookup resolves a prefix to a namespace URI. The empty prefix is answered by
// this scope's own default (the matched declaration's target namespace) before
// any $xmlns binding is consulted, so a $xmlns "" (empty-prefix) default-
// namespace override is deliberately shadowed: an unprefixed QName in an
// element's content always resolves to that element's target namespace.
func (s nsScope) lookup(prefix string) (string, bool) {
	if prefix == "" && s.hasDef {
		return s.def, true
	}
	if uri, ok := s.prefixes[prefix]; ok {
		return uri, true
	}
	if s.parent != nil {
		return s.parent.lookup(prefix)
	}
	return "", false
}

// Warning is a non-fatal diagnostic from building the infoset (e.g. a member key
// that collides with both an attribute use and a child element). It does not
// affect the validation verdict.
type Warning struct {
	Pos xsd.Pos
	Msg string
}

func (w Warning) String() string { return w.Pos.String() + ": " + w.Msg }

// builder decodes a JSON document into the infoset against a schema view,
// accumulating any warnings.
type builder struct {
	schema   xsdvalidate.Schema
	warnings []Warning
}

func (b *builder) warnf(pos xsd.Pos, format string, args ...any) {
	b.warnings = append(b.warnings, Warning{Pos: pos, Msg: fmt.Sprintf(format, args...)})
}

// classifier splits a complex type's members into attribute uses and child
// element declarations, keyed by local name. Element-wins on a local-name
// collision (a warning is emitted at build time when such a key is used).
type classifier struct {
	attrs    map[string]*xsd.AttributeDecl
	children map[string]*xsd.ElementDecl
}

func classify(ct *xsd.ComplexType) classifier {
	c := classifier{
		attrs:    map[string]*xsd.AttributeDecl{},
		children: map[string]*xsd.ElementDecl{},
	}
	if ct == nil {
		return c
	}
	for _, u := range ct.AttributeUses {
		if u.Decl == nil {
			continue
		}
		c.attrs[u.Decl.Name.Local] = u.Decl
	}
	for _, e := range xsdwalk.Elements(ct) {
		c.children[e.Name.Local] = e
	}
	return c
}

// collides reports whether local is declared as both an attribute use and a
// child element.
func (c classifier) collides(local string) bool {
	_, a := c.attrs[local]
	_, e := c.children[local]
	return a && e
}

// NewElement decodes a JSON instance from r into an xsdvalidate.Element,
// resolving keys against schema. It returns the root element, any non-fatal
// build warnings (e.g. attribute/element key collisions), and an error only
// when the input is not well-formed JSON or does not name a known root element.
// uri is used for diagnostic positions.
func NewElement(schema xsdvalidate.Schema, r io.Reader, uri string) (xsdvalidate.Element, []Warning, error) {
	b := &builder{schema: schema}
	dec := json.NewDecoder(r)
	v, err := decode(dec, uri)
	if err != nil {
		return nil, nil, err
	}
	root, err := b.buildRoot(v)
	if err != nil {
		return nil, b.warnings, err
	}
	return root, b.warnings, nil
}

// Validate decodes a JSON instance from r and assesses it with v. A non-nil
// error means the document could not be decoded as a valid JSON instance at all
// (malformed JSON or an unknown root element); schema-validity is reported
// through the returned Result. Non-fatal build warnings are returned alongside.
func Validate(v *xsdvalidate.Validator, r io.Reader, uri string) (*xsdvalidate.Result, []Warning, error) {
	root, warnings, err := NewElement(v.Schema(), r, uri)
	if err != nil {
		return nil, warnings, err
	}
	return v.Assess(root), warnings, nil
}
