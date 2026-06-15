// Package xmlsrc adapts an XML document (parsed by parser/xmltree) to the
// xsdvalidate abstract infoset, so an XML instance can be assessed against a
// compiled schema set. It is the first of several format sources (JSON/BER are
// future); the validation engine itself never imports an XML package.
package xmlsrc

import (
	"io"

	"github.com/kud360/goxsd5/parser/xmltree"
	"github.com/kud360/goxsd5/xsd"
	"github.com/kud360/goxsd5/xsdvalidate"
)

// element wraps an xmltree.Node as an xsdvalidate.Element.
type element struct{ n *xmltree.Node }

func (e element) Pos() xsd.Pos                { return e.n.Pos }
func (e element) Name() xsd.QName             { return xsd.QName{Namespace: e.n.Name.Space, Local: e.n.Name.Local} }
func (e element) Lookup(p string) (string, bool) { return e.n.NS.Lookup(p) }

func (e element) Attributes() []xsdvalidate.Attribute {
	out := make([]xsdvalidate.Attribute, len(e.n.Attrs))
	for i := range e.n.Attrs {
		out[i] = attribute{e.n.Attrs[i]}
	}
	return out
}

// Children returns the element's ordered children. encoding/xml (via xmltree)
// concatenates all character data into one CharData string rather than
// preserving its interleaving with elements; that loses position among siblings
// but not its presence, which is all the content-model rules need. The single
// text run is reported first, followed by the element children in order.
func (e element) Children() []xsdvalidate.Node {
	var out []xsdvalidate.Node
	if e.n.CharData != "" {
		out = append(out, text{e.n.CharData})
	}
	for _, c := range e.n.Children {
		out = append(out, element{c})
	}
	return out
}

type attribute struct{ a xmltree.Attr }

func (at attribute) Name() xsd.QName { return xsd.QName{Namespace: at.a.Name.Space, Local: at.a.Name.Local} }
func (at attribute) Value() string  { return at.a.Value }
func (at attribute) Pos() xsd.Pos   { return at.a.Pos }

type text struct{ s string }

func (t text) Data() string { return t.s }

// NewElement adapts an xmltree.Node into an xsdvalidate.Element.
func NewElement(n *xmltree.Node) xsdvalidate.Element { return element{n} }

// Validate parses an XML instance from r (uri is used for positions) and
// assesses it with v. A non-nil error means the document could not be parsed as
// XML at all; schema-validity is reported through the returned Result.
func Validate(v *xsdvalidate.Validator, r io.Reader, uri string) (*xsdvalidate.Result, error) {
	node, err := xmltree.Parse(r, uri)
	if err != nil {
		return nil, err
	}
	return v.Assess(NewElement(node)), nil
}
