// Package xmltree builds a generic, schema-agnostic XML node tree over
// encoding/xml. It captures what the schema parser needs and encoding/xml
// does not provide: per-node in-scope namespace bindings (so QNames inside
// attribute *values* can be resolved later), line/column positions, and all
// foreign attributes/elements preserved verbatim.
package xmltree

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/kud360/goxsd5/xsd"
	"golang.org/x/net/html/charset"
)

// Node is one element in the document tree.
type Node struct {
	Name     xml.Name // Space is the resolved namespace URI
	Attrs    []Attr
	Children []*Node
	// CharData is the concatenation of all character data immediately
	// under this element (including CDATA sections).
	CharData string
	Pos      xsd.Pos
	EndPos   xsd.Pos
	NS       *NSContext
}

// Attr is one attribute. Namespace declarations (xmlns, xmlns:p) are kept
// out of Attrs; they live in the node's NSContext.
type Attr struct {
	Name  xml.Name // Space is the resolved namespace URI ("" for unprefixed)
	Value string
	Pos   xsd.Pos // position of the owning start tag
}

// Attr returns the value of the named no-namespace attribute and whether it
// is present.
func (n *Node) Attr(local string) (string, bool) {
	for i := range n.Attrs {
		if n.Attrs[i].Name.Space == "" && n.Attrs[i].Name.Local == local {
			return n.Attrs[i].Value, true
		}
	}
	return "", false
}

// NSContext is an immutable chain of prefix→URI bindings in scope at a node.
type NSContext struct {
	parent   *NSContext
	bindings map[string]string // prefix → URI; "" key is the default namespace
}

// Lookup resolves a prefix to a namespace URI. The reserved `xml` prefix is
// always bound. The empty prefix resolves to the default namespace, which
// may be unbound ("", false).
func (c *NSContext) Lookup(prefix string) (string, bool) {
	if prefix == "xml" {
		return xsd.XMLNS, true
	}
	if prefix == "xmlns" {
		return xsd.XMLNSNS, true
	}
	for ctx := c; ctx != nil; ctx = ctx.parent {
		if uri, ok := ctx.bindings[prefix]; ok {
			if uri == "" {
				// An empty URI un-declares the binding (xmlns="" or,
				// in XML Namespaces 1.1, xmlns:p="").
				return "", false
			}
			return uri, true
		}
	}
	return "", false
}

// ResolveQName resolves a lexical QName ("tns:Foo" or "Foo") appearing in an
// attribute value, using this node's in-scope namespaces. Unprefixed names
// resolve to the default namespace if one is declared, else to no namespace.
func (n *Node) ResolveQName(s string) (xsd.QName, error) {
	// spec: src-qname — XSD 1.1 Part 1 §3.15.3 (xmlschema11-1.md#src-qname)
	prefix, local := "", s
	if i := strings.IndexByte(s, ':'); i >= 0 {
		prefix, local = s[:i], s[i+1:]
	}
	if !IsNCName(local) || (prefix != "" && !IsNCName(prefix)) {
		return xsd.QName{}, fmt.Errorf("%q is not a valid QName", s)
	}
	if prefix == "" {
		uri, _ := n.NS.Lookup("")
		return xsd.QName{Namespace: uri, Local: local}, nil
	}
	uri, ok := n.NS.Lookup(prefix)
	if !ok {
		return xsd.QName{}, fmt.Errorf("undefined namespace prefix %q in %q", prefix, s)
	}
	return xsd.QName{Namespace: uri, Local: local}, nil
}

// IsNCName reports whether s is a valid XML NCName (a Name with no colon).
func IsNCName(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if r == ':' {
			return false
		}
		if i == 0 {
			if !isNameStartChar(r) {
				return false
			}
		} else if !isNameChar(r) {
			return false
		}
	}
	return true
}

func isNameStartChar(r rune) bool {
	return r == '_' || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') ||
		(r >= 0xC0 && r <= 0xD6) || (r >= 0xD8 && r <= 0xF6) ||
		(r >= 0xF8 && r <= 0x2FF) || (r >= 0x370 && r <= 0x37D) ||
		(r >= 0x37F && r <= 0x1FFF) || (r >= 0x200C && r <= 0x200D) ||
		(r >= 0x2070 && r <= 0x218F) || (r >= 0x2C00 && r <= 0x2FEF) ||
		(r >= 0x3001 && r <= 0xD7FF) || (r >= 0xF900 && r <= 0xFDCF) ||
		(r >= 0xFDF0 && r <= 0xFFFD) || (r >= 0x10000 && r <= 0xEFFFF)
}

func isNameChar(r rune) bool {
	return isNameStartChar(r) || r == '-' || r == '.' || (r >= '0' && r <= '9') ||
		r == 0xB7 || (r >= 0x300 && r <= 0x36F) || (r >= 0x203F && r <= 0x2040)
}

// Parse reads an XML document and returns its root node. uri is used only
// for positions. Non-UTF-8 encodings are transcoded (BOM and <?xml
// encoding=…?> sniffing).
func Parse(r io.Reader, uri string) (*Node, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", uri, err)
	}
	// Transcode to UTF-8 up front so byte offsets used for line/column
	// mapping refer to the same bytes the decoder consumes.
	utf8r, err := charset.NewReader(bytes.NewReader(raw), "")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", uri, err)
	}
	data, err := io.ReadAll(utf8r)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", uri, err)
	}

	p := &treeParser{uri: uri, data: data}
	d := xml.NewDecoder(bytes.NewReader(data))
	return p.parse(d)
}

type treeParser struct {
	uri   string
	data  []byte
	lines []int // byte offset of the start of each line, built lazily
}

func (p *treeParser) pos(offset int64) xsd.Pos {
	if p.lines == nil {
		p.lines = []int{0}
		for i, b := range p.data {
			if b == '\n' {
				p.lines = append(p.lines, i+1)
			}
		}
	}
	// Binary search for the line containing offset.
	lo, hi := 0, len(p.lines)-1
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if int64(p.lines[mid]) <= offset {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return xsd.Pos{URI: p.uri, Line: lo + 1, Column: int(offset) - p.lines[lo] + 1}
}

func (p *treeParser) parse(d *xml.Decoder) (*Node, error) {
	var root *Node
	var stack []*Node
	var ns *NSContext
	var text []*strings.Builder // parallel to stack

	for {
		start := d.InputOffset()
		tok, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p.uri, err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if root != nil && len(stack) == 0 {
				return nil, fmt.Errorf("%s: multiple root elements", p.uri)
			}
			node := &Node{Pos: p.pos(start)}
			// Split namespace declarations from real attributes.
			var bindings map[string]string
			for _, a := range t.Attr {
				switch {
				case a.Name.Space == "xmlns":
					if bindings == nil {
						bindings = map[string]string{}
					}
					bindings[a.Name.Local] = a.Value
				case a.Name.Space == "" && a.Name.Local == "xmlns":
					if bindings == nil {
						bindings = map[string]string{}
					}
					bindings[""] = a.Value
				default:
					node.Attrs = append(node.Attrs, Attr{Name: a.Name, Value: a.Value, Pos: node.Pos})
				}
			}
			if bindings != nil {
				ns = &NSContext{parent: ns, bindings: bindings}
			}
			node.NS = ns
			node.Name = t.Name
			// encoding/xml leaves Name.Space as the raw prefix when the
			// prefix is unbound; detect that as a namespace error.
			if err := checkBound(ns, t.Name); err != nil {
				return nil, fmt.Errorf("%s: %w", p.uri, err)
			}
			for _, a := range node.Attrs {
				if err := checkBound(ns, a.Name); err != nil {
					return nil, fmt.Errorf("%s: %w", p.uri, err)
				}
			}
			if len(stack) == 0 {
				root = node
			} else {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, node)
			}
			stack = append(stack, node)
			text = append(text, &strings.Builder{})

		case xml.EndElement:
			if len(stack) == 0 {
				return nil, fmt.Errorf("%s: unexpected end element", p.uri)
			}
			node := stack[len(stack)-1]
			node.CharData = text[len(text)-1].String()
			node.EndPos = p.pos(d.InputOffset())
			stack = stack[:len(stack)-1]
			text = text[:len(text)-1]
			if node.NS != nil && len(stack) > 0 {
				ns = stack[len(stack)-1].NS
			} else if len(stack) == 0 {
				ns = nil
			}

		case xml.CharData:
			if len(stack) > 0 {
				text[len(text)-1].Write(t)
			} else if len(bytes.TrimSpace(t)) > 0 {
				return nil, fmt.Errorf("%s: character data outside root element", p.uri)
			}

		case xml.Comment, xml.ProcInst, xml.Directive:
			// Ignored: not part of the schema infoset we consume.
		}
	}
	if len(stack) != 0 {
		return nil, fmt.Errorf("%s: unexpected EOF: unclosed element <%s>", p.uri, stack[len(stack)-1].Name.Local)
	}
	if root == nil {
		return nil, fmt.Errorf("%s: no root element", p.uri)
	}
	return root, nil
}

// checkBound detects names whose Space is a raw, unbound prefix rather than
// a resolved namespace URI. encoding/xml does not error on unbound prefixes;
// it passes the prefix through as Space. A genuine namespace URI virtually
// always contains ':' or '/' (NCNames cannot); if Space looks like a prefix
// and is neither bound in scope nor declared as a URI, the prefix is unbound.
func checkBound(ns *NSContext, name xml.Name) error {
	if name.Space == "" || name.Space == "xml" || strings.ContainsAny(name.Space, ":/") {
		return nil
	}
	if _, ok := ns.Lookup(name.Space); ok {
		return nil
	}
	for c := ns; c != nil; c = c.parent {
		for _, u := range c.bindings {
			if u == name.Space {
				return nil
			}
		}
	}
	return fmt.Errorf("undefined namespace prefix %q in name %q", name.Space, name.Local)
}
