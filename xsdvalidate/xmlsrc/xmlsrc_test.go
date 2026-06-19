package xmlsrc_test

// Behaviour tests for the XML-to-infoset adapter. They assert the observable
// shape the adapter exposes to the assessor — element name, attributes (xsi:*
// kept, xmlns:* dropped), ordered children with the single leading text run,
// namespace-prefix resolution, and unparsed-entity reporting — plus the
// well-formedness error path of Validate.

import (
	"strings"
	"testing"

	"github.com/kud360/goxsd5/parser/xmltree"
	"github.com/kud360/goxsd5/xsd"
	"github.com/kud360/goxsd5/xsdvalidate"
	"github.com/kud360/goxsd5/xsdvalidate/xmlsrc"
)

func parse(t *testing.T, doc string) xsdvalidate.Element {
	t.Helper()
	n, err := xmltree.Parse(strings.NewReader(doc), "doc.xml")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return xmlsrc.NewElement(n)
}

func TestAdapterName(t *testing.T) {
	el := parse(t, `<r xmlns="urn:x"/>`)
	if got := el.Name(); got != (xsd.QName{Namespace: "urn:x", Local: "r"}) {
		t.Fatalf("Name() = %v", got)
	}
}

func TestAdapterAttributes(t *testing.T) {
	el := parse(t, `<r xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" a="1" xsi:nil="true"/>`)
	got := map[string]string{}
	for _, at := range el.Attributes() {
		got[at.Name().Local] = at.Value()
	}
	// The plain attribute and the xsi:* attribute are exposed; the xmlns:xsi
	// namespace declaration is not an attribute.
	if got["a"] != "1" {
		t.Fatalf("attribute a = %q, want 1", got["a"])
	}
	if got["nil"] != "true" {
		t.Fatalf("attribute xsi:nil = %q, want true", got["nil"])
	}
	if _, ok := got["xsi"]; ok {
		t.Fatalf("namespace declaration leaked as an attribute: %v", got)
	}
}

func TestAdapterChildrenOrder(t *testing.T) {
	el := parse(t, `<r>lead<a/><b/></r>`)
	kids := el.Children()
	if len(kids) != 3 {
		t.Fatalf("got %d children, want 3 (one text + two elements)", len(kids))
	}
	// encoding/xml concatenates text into one leading run, reported first.
	txt, ok := kids[0].(xsdvalidate.Text)
	if !ok || txt.Data() != "lead" {
		t.Fatalf("first child = %#v, want text %q", kids[0], "lead")
	}
	a, ok := kids[1].(xsdvalidate.Element)
	if !ok || a.Name().Local != "a" {
		t.Fatalf("second child = %#v, want element a", kids[1])
	}
	b, ok := kids[2].(xsdvalidate.Element)
	if !ok || b.Name().Local != "b" {
		t.Fatalf("third child = %#v, want element b", kids[2])
	}
}

func TestAdapterLookup(t *testing.T) {
	el := parse(t, `<r xmlns:p="urn:p"/>`)
	if uri, ok := el.Lookup("p"); !ok || uri != "urn:p" {
		t.Fatalf(`Lookup("p") = %q,%v; want "urn:p",true`, uri, ok)
	}
	if _, ok := el.Lookup("missing"); ok {
		t.Fatalf("Lookup of an undeclared prefix should fail")
	}
}

func TestAdapterUnparsedEntities(t *testing.T) {
	const doc = `<?xml version="1.0"?>
<!DOCTYPE r [
  <!NOTATION png SYSTEM "image/png">
  <!ENTITY pic SYSTEM "p.png" NDATA png>
]>
<r/>`
	el := parse(t, doc)
	di, ok := el.(xsdvalidate.DocumentInfo)
	if !ok {
		t.Fatal("root element does not implement DocumentInfo")
	}
	if !di.UnparsedEntities()["pic"] {
		t.Fatalf("UnparsedEntities = %v, want it to contain pic", di.UnparsedEntities())
	}
}

func TestValidateMalformedXML(t *testing.T) {
	v := xsdvalidate.New(nil, nil)
	// A non-nil error means the document could not be parsed as XML at all.
	if _, err := xmlsrc.Validate(v, strings.NewReader(`<r><unclosed></r>`), "doc.xml"); err == nil {
		t.Fatal("want a parse error for malformed XML, got nil")
	}
}
