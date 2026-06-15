package xmltree

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/kud360/goxsd5/xsd"
)

func mustParse(t *testing.T, doc string) *Node {
	t.Helper()
	n, err := Parse(strings.NewReader(doc), "test.xsd")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return n
}

func TestPositions(t *testing.T) {
	doc := "<root>\n  <child a=\"1\"/>\n</root>"
	root := mustParse(t, doc)
	if root.Pos != (xsd.Pos{URI: "test.xsd", Line: 1, Column: 1}) {
		t.Errorf("root pos = %v", root.Pos)
	}
	if len(root.Children) != 1 {
		t.Fatalf("children = %d", len(root.Children))
	}
	child := root.Children[0]
	if child.Pos.Line != 2 || child.Pos.Column != 3 {
		t.Errorf("child pos = %v, want line 2 col 3", child.Pos)
	}
	if root.EndPos.Line != 3 {
		t.Errorf("root end pos = %v, want line 3", root.EndPos)
	}
}

func TestQNameResolutionInAttrValue(t *testing.T) {
	doc := `<s xmlns="http://default" xmlns:tns="http://tns">
  <e type="tns:Foo"/>
  <f type="Bar"/>
</s>`
	root := mustParse(t, doc)
	e, f := root.Children[0], root.Children[1]

	q, err := e.ResolveQName("tns:Foo")
	if err != nil || q != (xsd.QName{Namespace: "http://tns", Local: "Foo"}) {
		t.Errorf("tns:Foo = %v, %v", q, err)
	}
	// Unprefixed resolves to the default namespace.
	q, err = f.ResolveQName("Bar")
	if err != nil || q != (xsd.QName{Namespace: "http://default", Local: "Bar"}) {
		t.Errorf("Bar = %v, %v", q, err)
	}
	if _, err := f.ResolveQName("nope:Bar"); err == nil {
		t.Error("undefined prefix should fail")
	}
	if _, err := f.ResolveQName("a:b:c"); err == nil {
		t.Error("a:b:c should not be a valid QName")
	}
	if _, err := f.ResolveQName(":Bar"); err == nil {
		t.Error(":Bar should not be a valid QName (empty prefix)")
	}
	if _, err := f.ResolveQName("tns:"); err == nil {
		t.Error("tns: should not be a valid QName (empty local part)")
	}
}

func TestNamespaceScoping(t *testing.T) {
	doc := `<a xmlns:p="http://one"><b xmlns:p="http://two"><c/></b><d/></a>`
	root := mustParse(t, doc)
	b := root.Children[0]
	c := b.Children[0]
	d := root.Children[1]
	if uri, _ := c.NS.Lookup("p"); uri != "http://two" {
		t.Errorf("c sees p=%q", uri)
	}
	if uri, _ := d.NS.Lookup("p"); uri != "http://one" {
		t.Errorf("d sees p=%q", uri)
	}
	// Undeclaring the default namespace.
	root2 := mustParse(t, `<a xmlns="http://d"><b xmlns=""><c/></b></a>`)
	c2 := root2.Children[0].Children[0]
	if uri, ok := c2.NS.Lookup(""); ok || uri != "" {
		t.Errorf("default ns should be undeclared, got %q ok=%v", uri, ok)
	}
}

func TestForeignContentAndCDATA(t *testing.T) {
	doc := `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:my="http://my">
  <xs:annotation my:note="kept">
    <xs:documentation><![CDATA[a < b]]> and more</xs:documentation>
    <my:extra><my:deep/></my:extra>
  </xs:annotation>
</xs:schema>`
	root := mustParse(t, doc)
	ann := root.Children[0]
	if v, ok := attrNS(ann, "http://my", "note"); !ok || v != "kept" {
		t.Errorf("foreign attribute lost: %q %v", v, ok)
	}
	docu := ann.Children[0]
	if docu.CharData != "a < b and more" {
		t.Errorf("CDATA text = %q", docu.CharData)
	}
	extra := ann.Children[1]
	if extra.Name.Space != "http://my" || len(extra.Children) != 1 {
		t.Errorf("foreign element not preserved: %+v", extra)
	}
}

func attrNS(n *Node, space, local string) (string, bool) {
	for _, a := range n.Attrs {
		if a.Name.Space == space && a.Name.Local == local {
			return a.Value, true
		}
	}
	return "", false
}

func TestMalformed(t *testing.T) {
	for _, doc := range []string{
		``,                          // no root
		`<a><b></a>`,                // mismatched tags
		`<a/><b/>`,                  // multiple roots
		`<a>`,                       // unclosed
		`<a><p:b xmlns:q="u"/></a>`, // unbound element prefix
		`<a p:x="1"/>`,              // unbound attribute prefix
		`text<a/>`,                  // chardata before root
	} {
		if _, err := Parse(strings.NewReader(doc), "bad.xml"); err == nil {
			t.Errorf("expected error for %q", doc)
		}
	}
}

func TestDTDEntities(t *testing.T) {
	// The prolog (and only the prolog) is buffered to extract internal-subset
	// entity declarations; the entities must be substituted in the body.
	root := mustParse(t, `<!DOCTYPE root [ <!ENTITY foo "bar"> ]>`+"\n<root>&foo;</root>")
	if root.CharData != "bar" {
		t.Errorf("CharData = %q, want %q", root.CharData, "bar")
	}
}

func TestDTDEntityTrickyValue(t *testing.T) {
	// Exercises consumeDoctype: ']' and '>' inside a quoted entity value must
	// not end the internal subset or the DOCTYPE, and a comment containing the
	// same characters must be skipped wholesale.
	doc := `<!DOCTYPE root [ <!ENTITY x "a]b>c"> <!-- ] and > inside --> ]>` +
		"\n<root>&x;</root>"
	root := mustParse(t, doc)
	if root.CharData != "a]b>c" {
		t.Errorf("CharData = %q, want %q", root.CharData, "a]b>c")
	}
}

func TestPrologMiscBeforeRoot(t *testing.T) {
	doc := "<?xml version=\"1.0\"?>\n<!-- comment -->\n<?pi data?>\n<root/>"
	root := mustParse(t, doc)
	if root.Name.Local != "root" {
		t.Fatalf("root = %q", root.Name.Local)
	}
	if root.Pos.Line != 4 {
		t.Errorf("root pos = %v, want line 4", root.Pos)
	}
}

func TestBOMStripped(t *testing.T) {
	root := mustParse(t, "\uFEFF<root>hi</root>")
	if root.Name.Local != "root" || root.CharData != "hi" {
		t.Errorf("root = %q chardata = %q", root.Name.Local, root.CharData)
	}
}

func TestUTF16Transcoded(t *testing.T) {
	// Encode a UTF-8 document as UTF-16LE with a BOM; charset detection must
	// transcode it and the parser must read it as if it were UTF-8.
	const src = `<root a="é">über</root>`
	units := utf16.Encode([]rune("\uFEFF" + src))
	var buf bytes.Buffer
	for _, u := range units {
		buf.WriteByte(byte(u))
		buf.WriteByte(byte(u >> 8))
	}
	root, err := Parse(&buf, "u16.xml")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if v, _ := root.Attr("a"); v != "é" || root.CharData != "über" {
		t.Errorf("a = %q chardata = %q", v, root.CharData)
	}
}

func TestParseLimit(t *testing.T) {
	doc := "<root>" + strings.Repeat("x", 1000) + "</root>"
	if _, err := ParseLimit(strings.NewReader(doc), "big.xml", 100); err == nil {
		t.Fatal("expected size-limit error")
	} else if !strings.Contains(err.Error(), "exceeds maximum size") {
		t.Errorf("error = %v, want size-limit error", err)
	}
	// Exactly at the limit is fine; 0 disables the limit.
	if _, err := ParseLimit(strings.NewReader(doc), "ok.xml", int64(len(doc))); err != nil {
		t.Errorf("at-limit Parse: %v", err)
	}
	if _, err := ParseLimit(strings.NewReader(doc), "ok.xml", 0); err != nil {
		t.Errorf("unlimited Parse: %v", err)
	}
}

func TestXMLPrefixAlwaysBound(t *testing.T) {
	root := mustParse(t, `<a xml:lang="en"/>`)
	if v, ok := attrNS(root, xsd.XMLNS, "lang"); !ok || v != "en" {
		t.Errorf("xml:lang = %q %v", v, ok)
	}
}
