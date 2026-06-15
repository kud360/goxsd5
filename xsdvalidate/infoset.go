// Package xsdvalidate is a format-agnostic, in-process XSD 1.1 schema-validity
// assessor: given a compiled []*xsd.Schema and an instance document exposed as
// an abstract infoset, it decides validity per the XSD 1.1 Part 1 §3 / Part 2
// validation rules (the cvc-* constraints) and reports each violation with the
// same xsd.SpecRef identity the schema processor uses.
//
// The engine walks the abstract Element/Attribute/Node interfaces and never
// imports an XML (or JSON/BER) package; each instance format ships an adapter
// (xsdvalidate/xmlsrc for XML). Value-space checks (lexical/facet/QName) are
// delegated to the type via xsd's ParseValue; the engine adds only the rules
// that are not per-datatype (structure, identity constraints, ID/IDREF).
package xsdvalidate

import "github.com/kud360/goxsd5/xsd"

// Name is an expanded (namespace-qualified) name in the instance infoset.
type Name = xsd.QName

// Node is one item in an element's ordered [children]: either an Element or a
// run of character data (Text). It is an open marker interface (any concrete
// item qualifies) so format adapters in other packages can supply their own
// node types; the engine type-switches to Element / Text. Keeping it open is
// what makes the infoset format-pluggable.
type Node any

// Element is one element information item.
type Element interface {
	// Name is the element's expanded name.
	Name() Name
	// Attributes returns the non-namespace attributes (xmlns:* excluded; xsi:*
	// included so the engine can read type/nil/schemaLocation).
	Attributes() []Attribute
	// Children returns the ordered element and character children.
	Children() []Node
	// Lookup resolves an in-scope namespace prefix (for QName-valued content);
	// the empty prefix is the default namespace.
	Lookup(prefix string) (uri string, ok bool)
	// Pos is the source position, for diagnostics.
	Pos() xsd.Pos
}

// Attribute is one attribute information item.
type Attribute interface {
	Name() Name
	Value() string
	Pos() xsd.Pos
}

// Text is a run of character data among an element's children.
type Text interface {
	Data() string
}

// DocumentInfo is an optional capability a root Element may implement to expose
// document-level DTD information. The engine reads UnparsedEntities at the
// validation root to enforce xs:ENTITY/ENTITIES referential validity (each
// value must name an unparsed entity declared in the DTD). A root that does not
// implement it leaves ENTITY referential checking disabled (fail-open).
type DocumentInfo interface {
	UnparsedEntities() map[string]bool
}
