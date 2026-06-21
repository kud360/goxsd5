// Package xmltree builds a generic, schema-agnostic XML node tree over
// encoding/xml. It captures what the schema parser needs and encoding/xml
// does not provide: per-node in-scope namespace bindings (so QNames inside
// attribute *values* can be resolved later), line/column positions, and all
// foreign attributes/elements preserved verbatim.
package xmltree

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html/charset"

	"github.com/kud360/goxsd5/xsd"
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
	// UnparsedEntities holds the names of unparsed (NDATA) general entities
	// declared in the document's internal DTD subset. It is populated only on
	// the root node and supports xs:ENTITY/ENTITIES referential validation.
	UnparsedEntities map[string]bool
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

// InScope returns a flattened copy of every prefix→URI binding in scope at this
// context, with inner declarations shadowing outer ones. The "" key is the
// default namespace. Prefixes un-declared by an empty URI are omitted. Returns
// nil when nothing is in scope. The reserved xml/xmlns prefixes are not
// included; resolve them via Lookup.
func (c *NSContext) InScope() map[string]string {
	if c == nil {
		return nil
	}
	var out map[string]string
	// Walk leaf→root collecting prefixes; an already-seen prefix is shadowed by
	// the inner binding, so the first sighting wins.
	seen := map[string]bool{}
	for ctx := c; ctx != nil; ctx = ctx.parent {
		for prefix, uri := range ctx.bindings {
			if seen[prefix] {
				continue
			}
			seen[prefix] = true
			if uri == "" {
				continue // un-declared binding
			}
			if out == nil {
				out = map[string]string{}
			}
			out[prefix] = uri
		}
	}
	return out
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

// MaxDocumentBytes is the default ceiling Parse imposes on a single document.
// It guards against unbounded memory use when reading from untrusted streams.
// Callers needing a different limit (including none) can use ParseLimit.
const MaxDocumentBytes = 1 << 30 // 1 GiB

// Parse reads an XML document and returns its root node. uri is used only
// for positions. Non-UTF-8 encodings are transcoded (BOM and <?xml
// encoding=…?> sniffing). Input larger than MaxDocumentBytes is rejected;
// use ParseLimit for a different ceiling.
func Parse(r io.Reader, uri string) (*Node, error) {
	return ParseLimit(r, uri, MaxDocumentBytes)
}

// ParseLimit is like Parse but reads at most maxBytes from r, returning an
// error if the document is larger. A maxBytes <= 0 means no limit. The document
// is streamed: only the prolog (for its internal DTD subset) and a line-offset
// index are held in memory, not the whole byte content.
func ParseLimit(r io.Reader, uri string, maxBytes int64) (*Node, error) {
	// Transcode to UTF-8 on the fly so byte offsets used for line/column
	// mapping refer to the same bytes the decoder consumes.
	utf8r, err := charset.NewReader(limitReader(r, maxBytes), "")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", uri, err)
	}
	br := bufio.NewReader(utf8r)
	// The transcoder preserves a byte order mark as U+FEFF, which
	// encoding/xml would report as character data before the root element.
	if bom, _ := br.Peek(3); bytes.Equal(bom, []byte("\uFEFF")) {
		_, _ = br.Discard(3)
	}
	// Buffer only the prolog so its internal DTD subset can be scanned for
	// general entity declarations before the decoder reaches the body; the rest
	// of the document streams straight into the decoder.
	prolog, err := readProlog(br)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", uri, err)
	}

	p := &treeParser{uri: uri, lines: []int{0}, unparsed: parseDTDUnparsedEntities(prolog)}
	// lineReader records newline offsets as bytes flow to the decoder, so pos()
	// can map offsets to line/column without retaining the whole document.
	src := &lineReader{r: io.MultiReader(bytes.NewReader(prolog), br), p: p}
	d := xml.NewDecoder(src)
	// encoding/xml does not process the internal DTD subset, so a reference to
	// a custom general entity (&name;) would otherwise be a hard error. Hand the
	// resolved replacement texts to the decoder, which substitutes them wherever
	// the entity is referenced.
	if ents := parseDTDEntities(prolog); ents != nil {
		d.Entity = ents
	}
	return p.parse(d)
}

// limitReader wraps r so that reading more than max bytes yields an error. A
// max <= 0 disables the limit.
func limitReader(r io.Reader, max int64) io.Reader {
	if max <= 0 {
		return r
	}
	return &boundedReader{r: r, max: max}
}

type boundedReader struct {
	r    io.Reader
	max  int64
	read int64
}

func (b *boundedReader) Read(p []byte) (int, error) {
	n, err := b.r.Read(p)
	b.read += int64(n)
	if b.read > b.max {
		return n, fmt.Errorf("document exceeds maximum size of %d bytes", b.max)
	}
	return n, err
}

// lineReader counts newlines in the bytes read through it, appending the byte
// offset of each line start to p.lines. Because the decoder reads bytes in
// document order, the index always covers any offset pos() is later asked
// about.
type lineReader struct {
	r      io.Reader
	p      *treeParser
	offset int
}

func (lr *lineReader) Read(b []byte) (int, error) {
	n, err := lr.r.Read(b)
	for i := 0; i < n; i++ {
		if b[i] == '\n' {
			lr.p.lines = append(lr.p.lines, lr.offset+i+1)
		}
	}
	lr.offset += n
	return n, err
}

// readProlog consumes everything before the root element's start tag from br
// (XML declaration, comments, processing instructions, and a DOCTYPE with its
// internal subset) and returns those bytes. The root element's '<' is left
// unread in br, so io.MultiReader of the returned bytes and br reconstructs the
// full document for the decoder.
func readProlog(br *bufio.Reader) ([]byte, error) {
	var buf []byte
	for {
		next, err := br.Peek(1)
		if err != nil {
			if err == io.EOF {
				return buf, nil // no root element; let the decoder report it
			}
			return buf, err
		}
		if next[0] != '<' {
			b, _ := br.ReadByte()
			buf = append(buf, b)
			continue
		}
		kind, _ := br.Peek(len("<!DOCTYPE")) // shorter near EOF; HasPrefix still works
		switch {
		case bytes.HasPrefix(kind, []byte("<!--")):
			err = consumeUntil(br, &buf, "-->")
		case bytes.HasPrefix(kind, []byte("<!DOCTYPE")):
			err = consumeDoctype(br, &buf)
		case bytes.HasPrefix(kind, []byte("<?")):
			err = consumeUntil(br, &buf, "?>")
		default:
			return buf, nil // root element start tag
		}
		if err != nil {
			return buf, err
		}
	}
}

// consumeUntil reads from br into *buf up to and including the first occurrence
// of term, returning the read error (e.g. io.ErrUnexpectedEOF) if the stream
// ends first.
func consumeUntil(br *bufio.Reader, buf *[]byte, term string) error {
	for {
		b, err := br.ReadByte()
		if err != nil {
			return err
		}
		*buf = append(*buf, b)
		if bytes.HasSuffix(*buf, []byte(term)) {
			return nil
		}
	}
}

// consumeDoctype reads a `<!DOCTYPE \u2026 >` declaration (including any internal
// subset) from br into *buf. The closing '>' is the first one found outside any
// quoted literal, comment, or `[ \u2026 ]` internal subset.
func consumeDoctype(br *bufio.Reader, buf *[]byte) error {
	depth := 0
	var quote byte
	for {
		b, err := br.ReadByte()
		if err != nil {
			return err
		}
		*buf = append(*buf, b)
		if quote != 0 {
			if b == quote {
				quote = 0
			}
			continue
		}
		switch b {
		case '"', '\'':
			quote = b
		case '[':
			depth++
		case ']':
			depth--
		case '-':
			if depth >= 1 && bytes.HasSuffix(*buf, []byte("<!--")) {
				if err := consumeUntil(br, buf, "-->"); err != nil {
					return err
				}
			}
		case '>':
			if depth <= 0 {
				return nil
			}
		}
	}
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

// parseDTDUnparsedEntities scans an internal DTD subset for unparsed general
// entity declarations — `<!ENTITY name SYSTEM|PUBLIC … NDATA notation>` — and
// returns the set of their names, used for xs:ENTITY/ENTITIES validity. It
// returns nil when there is no internal subset or no unparsed entity.
func parseDTDUnparsedEntities(data []byte) map[string]bool {
	subset, ok := internalSubset(data)
	if !ok {
		return nil
	}
	var names map[string]bool
	s := subset
	for i := 0; i < len(s); {
		switch {
		case strings.HasPrefix(s[i:], "<!--"):
			end := strings.Index(s[i+4:], "-->")
			if end < 0 {
				return names
			}
			i += 4 + end + 3
		case strings.HasPrefix(s[i:], "<!ENTITY"):
			end := entityDeclEnd(s, i)
			decl := s[i:end]
			// A parameter entity (`<!ENTITY % …>`) cannot be an xs:ENTITY referent.
			if name := unparsedEntityName(decl); name != "" {
				if names == nil {
					names = map[string]bool{}
				}
				names[name] = true
			}
			i = end
		case strings.HasPrefix(s[i:], "<!"), strings.HasPrefix(s[i:], "<?"):
			i = skipMarkupDecl(s, i)
		default:
			i++
		}
	}
	return names
}

// unparsedEntityName returns the declared name of an `<!ENTITY …>` declaration
// if it is an unparsed (NDATA) general entity, else "". decl spans the whole
// declaration including the leading "<!ENTITY" and trailing ">".
func unparsedEntityName(decl string) string {
	body := strings.TrimSuffix(strings.TrimPrefix(decl, "<!ENTITY"), ">")
	fields := strings.Fields(body)
	// Shape: name (SYSTEM "uri" | PUBLIC "id" "uri") NDATA notation
	if len(fields) < 2 || fields[0] == "%" {
		return ""
	}
	for _, f := range fields[1:] {
		if f == "NDATA" {
			return fields[0]
		}
	}
	return ""
}

// entityDeclEnd returns the index just past the `>` that closes the `<!ENTITY`
// declaration starting at s[i], skipping over quoted literals so a `>` inside a
// system identifier or replacement text does not terminate it early.
func entityDeclEnd(s string, i int) int {
	j := i + len("<!ENTITY")
	for j < len(s) {
		switch s[j] {
		case '"', '\'':
			quote := s[j]
			j++
			for j < len(s) && s[j] != quote {
				j++
			}
			if j < len(s) {
				j++
			}
		case '>':
			return j + 1
		default:
			j++
		}
	}
	return j
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
	uri      string
	lines    []int           // byte offset of the start of each line, grown by lineReader
	unparsed map[string]bool // unparsed (NDATA) entity names from the internal DTD subset
}

func (p *treeParser) pos(offset int64) xsd.Pos {
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
	b := &docBuilder{p: p}
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
			if err := b.startElement(t, start); err != nil {
				return nil, err
			}
		case xml.EndElement:
			if err := b.endElement(d.InputOffset()); err != nil {
				return nil, err
			}
		case xml.CharData:
			if err := b.charData(t); err != nil {
				return nil, err
			}
		case xml.Comment, xml.ProcInst, xml.Directive:
			// Ignored: not part of the schema infoset we consume.
		}
	}
	return b.finish()
}

// docBuilder accumulates the node tree across decoder tokens. text is parallel
// to stack: text[i] collects the character data of stack[i].
type docBuilder struct {
	p     *treeParser
	root  *Node
	stack []*Node
	ns    *NSContext
	text  []*strings.Builder
}

func (b *docBuilder) startElement(t xml.StartElement, start int64) error {
	p := b.p
	if b.root != nil && len(b.stack) == 0 {
		return fmt.Errorf("%s: multiple root elements", p.uri)
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
		b.ns = &NSContext{parent: b.ns, bindings: bindings}
	}
	node.NS = b.ns
	node.Name = t.Name
	// encoding/xml leaves Name.Space as the raw prefix when the prefix is
	// unbound; detect that as a namespace error.
	if err := checkBound(b.ns, t.Name); err != nil {
		return fmt.Errorf("%s: %w", p.uri, err)
	}
	for _, a := range node.Attrs {
		if err := checkBound(b.ns, a.Name); err != nil {
			return fmt.Errorf("%s: %w", p.uri, err)
		}
	}
	if len(b.stack) == 0 {
		b.root = node
	} else {
		parent := b.stack[len(b.stack)-1]
		parent.Children = append(parent.Children, node)
	}
	b.stack = append(b.stack, node)
	b.text = append(b.text, &strings.Builder{})
	return nil
}

func (b *docBuilder) endElement(off int64) error {
	if len(b.stack) == 0 {
		return fmt.Errorf("%s: unexpected end element", b.p.uri)
	}
	node := b.stack[len(b.stack)-1]
	node.CharData = b.text[len(b.text)-1].String()
	node.EndPos = b.p.pos(off)
	b.stack = b.stack[:len(b.stack)-1]
	b.text = b.text[:len(b.text)-1]
	if node.NS != nil && len(b.stack) > 0 {
		b.ns = b.stack[len(b.stack)-1].NS
	} else if len(b.stack) == 0 {
		b.ns = nil
	}
	return nil
}

func (b *docBuilder) charData(t xml.CharData) error {
	if len(b.stack) > 0 {
		b.text[len(b.text)-1].Write(t)
		return nil
	}
	if len(bytes.TrimSpace(t)) > 0 {
		return fmt.Errorf("%s: character data outside root element", b.p.uri)
	}
	return nil
}

func (b *docBuilder) finish() (*Node, error) {
	if len(b.stack) != 0 {
		return nil, fmt.Errorf("%s: unexpected EOF: unclosed element <%s>", b.p.uri, b.stack[len(b.stack)-1].Name.Local)
	}
	if b.root == nil {
		return nil, fmt.Errorf("%s: no root element", b.p.uri)
	}
	b.root.UnparsedEntities = b.p.unparsed
	return b.root, nil
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
