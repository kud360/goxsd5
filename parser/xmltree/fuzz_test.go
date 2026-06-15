package xmltree

import (
	"strings"
	"testing"
)

var xmlSeeds = []string{
	`<a/>`,
	`<a></a>`,
	`<a x="1"/>`,
	`<a xmlns="urn:n"><b/></a>`,
	`<a xmlns:p="urn:n"><p:b p:x="1"/></a>`,
	`<a><![CDATA[<not a tag>]]></a>`,
	`<a><!-- comment --><b/></a>`,
	`<?xml version="1.0"?><a/>`,
	`<a>text<b/>more</a>`,
	"<a>\n  <b/>\n</a>",
	`<a xmlns:p="urn:n" attr="p:val"/>`,
	// malformed / edge inputs
	``, `<`, `<a>`, `</a>`, `<a><b></a>`, `<a x=/>`, `<a x="&bad;"/>`,
	`<a:b/>`, `<a/><b/>`, `<a><a><a><a/></a></a></a>`,
}

// FuzzParse asserts the XML tree builder never panics on arbitrary input and
// that any tree it returns is structurally sound: a returned root is non-nil,
// and every node's reported position is non-negative and walkable.
func FuzzParse(f *testing.F) {
	for _, s := range xmlSeeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, doc string) {
		root, err := Parse(strings.NewReader(doc), "fuzz.xsd")
		if err != nil {
			return
		}
		if root == nil {
			t.Fatalf("Parse(%q) returned nil root, nil error", doc)
		}
		walk(t, doc, root)
	})
}

// walk recursively validates structural invariants of every node.
func walk(t *testing.T, doc string, n *Node) {
	t.Helper()
	if n.Pos.Line < 0 || n.Pos.Column < 0 {
		t.Fatalf("Parse(%q): negative position %v", doc, n.Pos)
	}
	for _, c := range n.Children {
		if c == nil {
			t.Fatalf("Parse(%q): nil child under %v", doc, n.Name)
		}
		walk(t, doc, c)
	}
}

// FuzzIsNCName drives the NCName predicate, which must never panic on any
// UTF-8 (or invalid-UTF-8) string and must reject names containing a colon.
func FuzzIsNCName(f *testing.F) {
	for _, s := range []string{"", "a", "a1", "_x", "a:b", ":", "1a", "a-b", "ünïcode", "a.b"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		if IsNCName(s) && strings.Contains(s, ":") {
			t.Fatalf("IsNCName(%q) = true but contains a colon", s)
		}
	})
}
