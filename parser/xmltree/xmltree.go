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
	hasColon := false
	if i := strings.IndexByte(s, ':'); i >= 0 {
		prefix, local, hasColon = s[:i], s[i+1:], true
	}
	// A QName with a colon must carry a non-empty NCName prefix; ":local" and
	// "prefix:" are both malformed.
	if !IsNCName(local) || (hasColon && !IsNCName(prefix)) {
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
	// The transcoder preserves a byte order mark as U+FEFF, which
	// encoding/xml would report as character data before the root element.
	data = bytes.TrimPrefix(data, []byte("\uFEFF"))

	p := &treeParser{uri: uri, data: data}
	d := xml.NewDecoder(bytes.NewReader(data))
	// encoding/xml does not process the internal DTD subset, so a reference to
	// a custom general entity (&name;) would otherwise be a hard error. Pre-scan
	// the subset and hand the resolved replacement texts to the decoder, which
	// substitutes them wherever the entity is referenced.
	if ents := parseDTDEntities(data); ents != nil {
		d.Entity = ents
	}
	return p.parse(d)
}

// parseDTDEntities extracts general entity declarations from an internal DTD
// subset (`<!DOCTYPE root [ … ]>`) and returns a map of entity name to its
// fully-resolved replacement text, suitable for xml.Decoder.Entity. It returns
// nil when the document has no internal subset. The first declaration of a
// given name wins (per XML). Parameter entities (`<!ENTITY % …>`) and external
// entities are ignored; references inside replacement text to char references,
// predefined entities, and other general entities are expanded recursively.
func parseDTDEntities(data []byte) map[string]string {
	subset, ok := internalSubset(data)
	if !ok {
		return nil
	}
	raw := map[string]string{} // name → declared (unexpanded) replacement text
	s := subset
	for i := 0; i < len(s); {
		switch {
		case isSpace(s[i]):
			i++
		case strings.HasPrefix(s[i:], "<!--"):
			end := strings.Index(s[i+4:], "-->")
			if end < 0 {
				return resolveEntities(raw)
			}
			i += 4 + end + 3
		case strings.HasPrefix(s[i:], "<!ENTITY"):
			name, value, next, ok := parseEntityDecl(s, i)
			if !ok {
				return resolveEntities(raw)
			}
			if name != "" { // "" marks a parameter or malformed entity: skip it
				if _, seen := raw[name]; !seen {
					raw[name] = value
				}
			}
			i = next
		case strings.HasPrefix(s[i:], "<!"), strings.HasPrefix(s[i:], "<?"):
			i = skipMarkupDecl(s, i)
		default:
			i++
		}
	}
	return resolveEntities(raw)
}

// internalSubset returns the text between the `[` and the matching top-level
// `]` of a `<!DOCTYPE … [ … ]>` declaration. The matching `]` is found by
// skipping over nested `<! … >`/`<? … ?>` declarations and quoted strings, so
// a `]` inside an entity value (e.g. a character class) does not terminate it.
func internalSubset(data []byte) (string, bool) {
	doc := string(data)
	dt := strings.Index(doc, "<!DOCTYPE")
	if dt < 0 {
		return "", false
	}
	open := strings.IndexByte(doc[dt:], '[')
	gt := strings.IndexByte(doc[dt:], '>')
	if open < 0 || (gt >= 0 && gt < open) {
		return "", false // no internal subset (external-only or none)
	}
	i := dt + open + 1
	start := i
	for i < len(doc) {
		switch {
		case doc[i] == ']':
			return doc[start:i], true
		case strings.HasPrefix(doc[i:], "<!--"):
			end := strings.Index(doc[i+4:], "-->")
			if end < 0 {
				return doc[start:], true
			}
			i += 4 + end + 3
		case strings.HasPrefix(doc[i:], "<!"), strings.HasPrefix(doc[i:], "<?"):
			i = skipMarkupDecl(doc, i)
		default:
			i++
		}
	}
	return doc[start:], true
}

// parseEntityDecl parses one `<!ENTITY name "value">` starting at s[i] (which
// begins "<!ENTITY"). It returns the name, the declared replacement text, the
// index just past the closing `>`, and whether parsing succeeded. A parameter
// entity (`<!ENTITY % …>`) or one without a quoted literal yields name "" so
// the caller skips it.
func parseEntityDecl(s string, i int) (name, value string, next int, ok bool) {
	j := i + len("<!ENTITY")
	for j < len(s) && isSpace(s[j]) {
		j++
	}
	if j < len(s) && s[j] == '%' { // parameter entity: not referenced from content
		return "", "", skipMarkupDecl(s, i), true
	}
	nameStart := j
	for j < len(s) && !isSpace(s[j]) && s[j] != '>' {
		j++
	}
	name = s[nameStart:j]
	for j < len(s) && isSpace(s[j]) {
		j++
	}
	if j >= len(s) || (s[j] != '"' && s[j] != '\'') {
		// External (SYSTEM/PUBLIC) or malformed entity: skip without recording.
		return "", "", skipMarkupDecl(s, i), true
	}
	quote := s[j]
	j++
	valStart := j
	for j < len(s) && s[j] != quote {
		j++
	}
	if j >= len(s) {
		return "", "", len(s), false
	}
	value = s[valStart:j]
	return name, value, skipMarkupDecl(s, i), true
}

// skipMarkupDecl returns the index just past the `>` that closes the markup
// declaration starting at s[i], ignoring any `>` that appears inside a quoted
// literal.
func skipMarkupDecl(s string, i int) int {
	var quote byte
	for ; i < len(s); i++ {
		c := s[i]
		if quote != 0 {
			if c == quote {
				quote = 0
			}
			continue
		}
		switch c {
		case '"', '\'':
			quote = c
		case '>':
			return i + 1
		}
	}
	return len(s)
}

// resolveEntities fully expands every declared entity's replacement text and
// returns name → final text, or nil if there are none. Expansion handles
// character references, the five predefined entities, and nested general
// entity references, guarding against reference cycles.
func resolveEntities(raw map[string]string) map[string]string {
	if len(raw) == 0 {
		return nil
	}
	out := make(map[string]string, len(raw))
	for name := range raw {
		out[name] = expandEntityText(raw[name], raw, map[string]bool{name: true})
	}
	return out
}

// expandEntityText expands references within a replacement text. visiting holds
// the entity names currently being expanded, so a cyclic reference stops rather
// than recursing forever.
func expandEntityText(text string, raw map[string]string, visiting map[string]bool) string {
	var b strings.Builder
	for i := 0; i < len(text); {
		if text[i] != '&' {
			b.WriteByte(text[i])
			i++
			continue
		}
		semi := strings.IndexByte(text[i:], ';')
		if semi < 0 {
			b.WriteString(text[i:])
			break
		}
		ref := text[i+1 : i+semi]
		i += semi + 1
		switch {
		case strings.HasPrefix(ref, "#"):
			if r, ok := parseCharRef(ref[1:]); ok {
				b.WriteRune(r)
			}
		case ref == "amp":
			b.WriteByte('&')
		case ref == "lt":
			b.WriteByte('<')
		case ref == "gt":
			b.WriteByte('>')
		case ref == "quot":
			b.WriteByte('"')
		case ref == "apos":
			b.WriteByte('\'')
		default:
			if val, ok := raw[ref]; ok && !visiting[ref] {
				visiting[ref] = true
				b.WriteString(expandEntityText(val, raw, visiting))
				delete(visiting, ref)
			} else {
				// Unknown or cyclic reference: leave it literal so the decoder
				// surfaces the error rather than silently dropping it.
				b.WriteByte('&')
				b.WriteString(ref)
				b.WriteByte(';')
			}
		}
	}
	return b.String()
}

// parseCharRef decodes the body of a character reference (the part between
// "&#" and ";"), in decimal or "x"-prefixed hexadecimal.
func parseCharRef(s string) (rune, bool) {
	base, digits := 10, s
	if len(s) > 0 && (s[0] == 'x' || s[0] == 'X') {
		base, digits = 16, s[1:]
	}
	if digits == "" {
		return 0, false
	}
	var n int64
	for _, c := range digits {
		var d int64
		switch {
		case c >= '0' && c <= '9':
			d = int64(c - '0')
		case base == 16 && c >= 'a' && c <= 'f':
			d = int64(c-'a') + 10
		case base == 16 && c >= 'A' && c <= 'F':
			d = int64(c-'A') + 10
		default:
			return 0, false
		}
		n = n*int64(base) + d
		if n > 0x10FFFF {
			return 0, false
		}
	}
	return rune(n), true
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
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
