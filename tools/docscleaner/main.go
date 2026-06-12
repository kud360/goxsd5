// docscleaner converts W3C XSD spec HTML files into clean Markdown suitable
// for AI agent consumption. It strips CSS, navigation chrome, and W3C boilerplate
// while preserving headings (with anchor IDs), cross-reference links, code blocks,
// lists, tables, and all normative content.
//
// Usage:
//
//	go run ./tools/docscleaner [--input docs/raw] [--output docs/clean]
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func main() {
	input := flag.String("input", "docs/raw", "directory containing raw HTML files")
	output := flag.String("output", "docs/clean", "directory for cleaned Markdown output")
	flag.Parse()

	if err := os.MkdirAll(*output, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", *output, err)
		os.Exit(1)
	}

	entries, err := os.ReadDir(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", *input, err)
		os.Exit(1)
	}

	ok := true
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".html") {
			continue
		}
		inPath := filepath.Join(*input, e.Name())
		outPath := filepath.Join(*output, strings.TrimSuffix(e.Name(), ".html")+".md")
		if err := convert(inPath, outPath); err != nil {
			fmt.Fprintf(os.Stderr, "error: %s: %v\n", inPath, err)
			ok = false
		} else {
			fmt.Printf("ok: %s -> %s\n", inPath, outPath)
		}
	}
	if !ok {
		os.Exit(1)
	}
}

func convert(inPath, outPath string) error {
	f, err := os.Open(inPath)
	if err != nil {
		return err
	}
	defer f.Close()

	doc, err := html.Parse(f)
	if err != nil {
		return err
	}

	var sb strings.Builder
	if title := findTitle(doc); title != "" {
		sb.WriteString("# " + title + "\n")
	}
	c := newConverter(&sb)
	c.walk(doc)

	return os.WriteFile(outPath, []byte(finalCleanup(sb.String())), 0644)
}

// converter walks an HTML node tree and emits Markdown.
type converter struct {
	out       *strings.Builder
	inPre     bool
	listStack []listState
}

type listState struct {
	kind  string // "ul" or "ol"
	count int
}

func newConverter(out *strings.Builder) *converter {
	return &converter{out: out}
}

// attr returns the value of the named attribute, or "".
func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func hasClass(n *html.Node, cls string) bool {
	for _, c := range strings.Fields(attr(n, "class")) {
		if c == cls {
			return true
		}
	}
	return false
}

// nodeID returns the id or name attribute of a node.
func nodeID(n *html.Node) string {
	if id := attr(n, "id"); id != "" {
		return id
	}
	return attr(n, "name")
}

// skipEntirely returns true for nodes whose subtrees should be dropped.
func skipEntirely(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}
	switch n.DataAtom {
	case atom.Head, atom.Style, atom.Script, atom.Link, atom.Meta, atom.Hr:
		return true
	case atom.Img:
		return true
	}
	// W3C document header block (logo, title, editors, copyright)
	if n.DataAtom == atom.Div && hasClass(n, "head") {
		return true
	}
	// "Status of this Document" boilerplate
	if n.DataAtom == atom.Div && hasClass(n, "sotd") {
		return true
	}
	// navigation arrows between sections
	if n.DataAtom == atom.Span && hasClass(n, "nav") {
		return true
	}
	// span.arrow wraps the "·" middots that set off defined terms; the
	// surrounding link already marks the term, so the middots are pure noise.
	if n.DataAtom == atom.Span && hasClass(n, "arrow") {
		return true
	}
	return false
}

// firstAnchorID returns the id/name of the first <a> descendant, or "".
// Used to preserve a cross-reference target when collapsing a block to code.
func firstAnchorID(n *html.Node) string {
	if n.Type == html.ElementNode && n.DataAtom == atom.A {
		if id := nodeID(n); id != "" {
			return id
		}
	}
	for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
		if id := firstAnchorID(ch); id != "" {
			return id
		}
	}
	return ""
}

// collectText concatenates all descendant text, turning <br> into newlines.
func collectText(n *html.Node, sb *strings.Builder) {
	switch n.Type {
	case html.TextNode:
		sb.WriteString(n.Data)
	case html.ElementNode:
		if n.DataAtom == atom.Br {
			sb.WriteString("\n")
			return
		}
		for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
			collectText(ch, sb)
		}
	}
}

// findTitle returns the text of the document's <title> element.
func findTitle(n *html.Node) string {
	if n.Type == html.ElementNode && n.DataAtom == atom.Title {
		var sb strings.Builder
		collectText(n, &sb)
		return strings.TrimSpace(sb.String())
	}
	for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
		if t := findTitle(ch); t != "" {
			return t
		}
	}
	return ""
}

func (c *converter) walk(n *html.Node) {
	switch n.Type {
	case html.DocumentNode:
		c.walkChildren(n)
	case html.TextNode:
		c.emitText(n.Data)
	case html.ElementNode:
		if skipEntirely(n) {
			return
		}
		c.element(n)
	}
}

func (c *converter) walkChildren(n *html.Node) {
	for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
		c.walk(ch)
	}
}

func (c *converter) write(s string) { c.out.WriteString(s) }

func (c *converter) emitText(s string) {
	if c.inPre {
		c.write(s)
		return
	}
	// collapse whitespace for normal flow text
	c.write(reWhitespace.ReplaceAllString(s, " "))
}

// captureChildren renders child nodes into a new string builder and returns the result.
func (c *converter) captureChildren(n *html.Node) string {
	var sb strings.Builder
	child := newConverter(&sb)
	child.inPre = c.inPre
	child.listStack = c.listStack
	child.walkChildren(n)
	return sb.String()
}

// emitSyntax renders an "XML Representation Summary" block (the monospace
// pseudo-XML element syntax) as a fenced code block, preserving the element's
// anchor so that [<foo>](#element-foo) cross-references still resolve.
func (c *converter) emitSyntax(n *html.Node) {
	id := firstAnchorID(n)
	var sb strings.Builder
	collectText(n, &sb)
	text := strings.ReplaceAll(sb.String(), " ", " ") // nbsp -> space
	text = strings.Trim(text, "\n")
	c.write("\n\n")
	if id != "" {
		c.write(`<a id="` + id + `"></a>` + "\n\n")
	}
	c.write("```xml\n" + text + "\n```\n\n")
}

func (c *converter) element(n *html.Node) {
	id := nodeID(n)

	// "XML Representation Summary" syntax blocks (div or p) -> code fence.
	if hasClass(n, "element-syntax") || hasClass(n, "element-syntax-1") {
		c.emitSyntax(n)
		return
	}

	switch n.DataAtom {

	// ── Headings ─────────────────────────────────────────────────────────────
	case atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6:
		level := int(n.Data[1] - '0') // n.Data is "h1".."h6"
		inner := strings.TrimSpace(c.captureChildren(n))
		// strip any trailing anchor html that the inner capture may have emitted
		// (anchors inside headings are already embedded as <a id="x"></a>)
		c.write("\n\n")
		c.write(strings.Repeat("#", level))
		c.write(" ")
		c.write(inner)
		if id != "" {
			c.write(" {#" + id + "}")
		}
		c.write("\n\n")

	// ── Paragraphs ───────────────────────────────────────────────────────────
	case atom.P:
		c.write("\n\n")
		c.walkChildren(n)
		c.write("\n\n")

	// ── Links and anchors ────────────────────────────────────────────────────
	case atom.A:
		if c.inPre {
			// inside a code block, emit only the link text — never markup
			c.walkChildren(n)
			return
		}
		href := attr(n, "href")
		inner := strings.TrimSpace(c.captureChildren(n))
		switch {
		case href != "" && inner != "":
			c.write("[" + inner + "](" + href + ")")
		case href != "":
			// empty link – just emit href as plain text
			c.write(href)
		case inner != "":
			// named anchor with visible text
			if id != "" {
				c.write(`<a id="` + id + `">` + inner + `</a>`)
			} else {
				c.write(inner)
			}
		case id != "":
			// empty named anchor – preserve as HTML so cross-references resolve
			c.write(`<a id="` + id + `"></a>`)
		}

	// ── Code ─────────────────────────────────────────────────────────────────
	case atom.Code:
		if c.inPre {
			c.walkChildren(n)
		} else {
			c.write("`")
			c.walkChildren(n)
			c.write("`")
		}

	case atom.Pre:
		c.write("\n\n```\n")
		c.inPre = true
		c.walkChildren(n)
		c.inPre = false
		c.write("\n```\n\n")

	// ── Inline formatting ────────────────────────────────────────────────────
	case atom.Strong, atom.B:
		c.write("**")
		c.walkChildren(n)
		c.write("**")

	case atom.Em, atom.I, atom.Var:
		c.write("*")
		c.walkChildren(n)
		c.write("*")

	case atom.Q:
		c.write("\"")
		c.walkChildren(n)
		c.write("\"")

	case atom.Sup:
		c.write("^")
		c.walkChildren(n)
		c.write("^")

	case atom.Sub:
		c.write("~")
		c.walkChildren(n)
		c.write("~")

	case atom.Acronym, atom.Abbr:
		c.walkChildren(n)
		if title := attr(n, "title"); title != "" {
			c.write(" (" + title + ")")
		}

	case atom.Br:
		c.write("\n")

	// ── Lists ────────────────────────────────────────────────────────────────
	case atom.Ul:
		c.listStack = append(c.listStack, listState{kind: "ul"})
		c.write("\n")
		c.walkChildren(n)
		c.listStack = c.listStack[:len(c.listStack)-1]
		c.write("\n")

	case atom.Ol:
		c.listStack = append(c.listStack, listState{kind: "ol"})
		c.write("\n")
		c.walkChildren(n)
		c.listStack = c.listStack[:len(c.listStack)-1]
		c.write("\n")

	case atom.Li:
		depth := len(c.listStack) - 1
		if depth < 0 {
			depth = 0
		}
		indent := strings.Repeat("  ", depth)
		if depth < len(c.listStack) && c.listStack[depth].kind == "ol" {
			c.listStack[depth].count++
			c.write(fmt.Sprintf("\n%s%d. ", indent, c.listStack[depth].count))
		} else {
			c.write("\n" + indent + "- ")
		}
		c.walkChildren(n)

	// ── Definition lists ─────────────────────────────────────────────────────
	case atom.Dl:
		c.write("\n")
		c.walkChildren(n)
		c.write("\n")

	case atom.Dt:
		c.write("\n**")
		c.walkChildren(n)
		c.write("**")

	case atom.Dd:
		c.write("  \n  ")
		c.walkChildren(n)
		c.write("\n")

	// ── Blockquote ───────────────────────────────────────────────────────────
	case atom.Blockquote:
		inner := strings.TrimSpace(c.captureChildren(n))
		c.write("\n\n")
		for _, line := range strings.Split(inner, "\n") {
			c.write("> " + line + "\n")
		}
		c.write("\n")

	// ── Tables (kept as clean HTML – too structured for MD conversion) ────────
	case atom.Table:
		c.write("\n\n")
		emitCleanTable(c.out, n)
		c.write("\n\n")

	// ── Span ─────────────────────────────────────────────────────────────────
	case atom.Span:
		if hasClass(n, "rfc2119") {
			// RFC 2119 normative keywords (MUST, SHALL, SHOULD, …) – bold
			c.write("**")
			c.walkChildren(n)
			c.write("**")
		} else {
			c.walkChildren(n)
		}

	// ── Div (structural containers) ───────────────────────────────────────────
	case atom.Div:
		cls := attr(n, "class")
		switch {
		case hasClass(n, "exampleOuter") || hasClass(n, "exampleWrapper"):
			c.write("\n\n")
			c.walkChildren(n)
			c.write("\n\n")

		case hasClass(n, "exampleHeader"):
			inner := strings.TrimSpace(c.captureChildren(n))
			c.write("\n*" + inner + "*\n")

		case hasClass(n, "exampleInner"):
			// the <pre> inside will render as a fenced code block
			c.walkChildren(n)

		case hasClass(n, "schemaComp") || hasClass(n, "component") ||
			hasClass(n, "psviDef") || hasClass(n, "reprdef"):
			c.write("\n\n---\n\n")
			c.walkChildren(n)
			c.write("\n\n---\n\n")

		// Property->Representation mapping rows. These layout divs come in
		// (mapProp, mapRepr) pairs separated by mapSep; rendered flat they glue
		// property names to their descriptions. Emit each as a list row.
		case hasClass(n, "reprHead"):
			inner := strings.TrimSpace(c.captureChildren(n))
			c.write("\n\n**" + inner + "**\n")

		case hasClass(n, "mapProp"):
			inner := strings.TrimSpace(c.captureChildren(n))
			c.write("\n- " + inner + ": ")

		case hasClass(n, "mapRepr"):
			inner := strings.TrimSpace(c.captureChildren(n))
			c.write(inner)

		case hasClass(n, "mapSep"):
			// row separator; the next mapProp already starts a fresh line

		// Nested property/value mini-tables (e.g. a Scope or Value Constraint
		// shown inline within a mapping row). Render as an indented sub-list.
		case hasClass(n, "pvProp"):
			inner := strings.TrimSpace(c.captureChildren(n))
			c.write("\n  - " + inner + " = ")

		case hasClass(n, "pvVal"):
			inner := strings.TrimSpace(c.captureChildren(n))
			c.write(inner)

		case hasClass(n, "constraint"):
			// outermost constraint block — emit as blockquote
			inner := strings.TrimSpace(c.captureChildren(n))
			c.write("\n\n")
			for _, line := range strings.Split(inner, "\n") {
				c.write("> " + line + "\n")
			}
			c.write("\n")

		case hasClass(n, "constraintlist"):
			// ordered list of constraint items — recurse normally
			c.write("\n")
			c.walkChildren(n)
			c.write("\n")

		case hasClass(n, "clnumber"):
			// individual constraint list item
			c.write("\n- ")
			c.walkChildren(n)

		case strings.HasPrefix(cls, "div") || cls == "body" || cls == "toc" ||
			cls == "toc1" || cls == "toc2" || cls == "" ||
			strings.Contains(cls, "compBody") || strings.Contains(cls, "compHeader") ||
			strings.Contains(cls, "scHead") || strings.Contains(cls, "tocLine"):
			c.walkChildren(n)

		default:
			c.walkChildren(n)
		}

	default:
		// pass-through for any unrecognized element
		c.walkChildren(n)
	}
}

// emitCleanTable renders a table subtree as HTML without class/style attributes.
func emitCleanTable(sb *strings.Builder, n *html.Node) {
	switch n.Type {
	case html.TextNode:
		sb.WriteString(n.Data)
	case html.ElementNode:
		switch n.DataAtom {
		case atom.Table, atom.Thead, atom.Tbody, atom.Tfoot, atom.Tr, atom.Caption:
			sb.WriteString("<" + n.Data + ">")
			for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
				emitCleanTable(sb, ch)
			}
			sb.WriteString("</" + n.Data + ">")

		case atom.Th, atom.Td:
			tag := n.Data
			var attrs string
			if v := attr(n, "colspan"); v != "" && v != "1" {
				attrs += fmt.Sprintf(` colspan="%s"`, v)
			}
			if v := attr(n, "rowspan"); v != "" && v != "1" {
				attrs += fmt.Sprintf(` rowspan="%s"`, v)
			}
			sb.WriteString("<" + tag + attrs + ">")
			for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
				emitCleanTable(sb, ch)
			}
			sb.WriteString("</" + tag + ">")

		case atom.A:
			href := attr(n, "href")
			id := nodeID(n)
			var inner strings.Builder
			for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
				emitCleanTable(&inner, ch)
			}
			text := inner.String()
			if href != "" {
				sb.WriteString(`<a href="` + href + `">` + text + `</a>`)
			} else if id != "" {
				sb.WriteString(`<a id="` + id + `">` + text + `</a>`)
			} else {
				sb.WriteString(text)
			}

		case atom.Code, atom.Pre:
			sb.WriteString("<" + n.Data + ">")
			for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
				emitCleanTable(sb, ch)
			}
			sb.WriteString("</" + n.Data + ">")

		default:
			for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
				emitCleanTable(sb, ch)
			}
		}
	}
}

var (
	reWhitespace = regexp.MustCompile(`\s+`)
	reBlankLines = regexp.MustCompile(`\n{3,}`)
	reTrailSpace = regexp.MustCompile(` +\n`)
)

func finalCleanup(s string) string {
	s = reTrailSpace.ReplaceAllString(s, "\n")
	s = reBlankLines.ReplaceAllString(s, "\n\n")
	s = strings.TrimSpace(s)
	return s + "\n"
}
