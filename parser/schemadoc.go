// Package parser turns schema documents into the public xsd model. Pass 1
// (this milestone) structurally validates each document tree against the
// XML representations of XSD 1.1 and collects global components into a
// registry of dangling (kind, QName) → node entries; pass 2 resolves and
// builds components on demand.
package parser

import (
	"strings"

	"github.com/kud360/goxsd5/parser/xmltree"
	"github.com/kud360/goxsd5/xsd"
)

// composition is one include/import/redefine/override child of xs:schema,
// recorded for the document loader (M7).
type composition struct {
	kind           string // "include", "import", "redefine", "override"
	namespace      string // import only; "" otherwise
	schemaLocation string // optional for import
	node           *xmltree.Node
}

// schemaDoc is one structurally validated schema document.
type schemaDoc struct {
	root *xmltree.Node
	uri  string

	targetNamespace      string
	version              string
	elementFormDefault   xsd.Form
	attributeFormDefault xsd.Form
	blockDefault         xsd.DerivationSet
	finalDefault         xsd.DerivationSet
	// defaultAttributes is the QName of the schema-wide default attribute
	// group; resolved in pass 2.
	defaultAttributes  xsd.QName
	defaultOpenContent *xmltree.Node
	xpathDefaultNS     string

	compositions []composition

	// pruned marks elements removed by conditional inclusion
	// (vc:minVersion/vc:maxVersion); later passes must not look at them.
	pruned map[*xmltree.Node]bool

	// scoped shadows the global registry with this document's redefine/
	// override replacements; resolution from inside this document consults
	// it first (M7 wires the cross-document semantics).
	scoped *registry
}

// loadDoc structurally validates one parsed schema document and extracts
// its document-level properties. Errors accumulate in errs; the returned
// doc is usable (with defaulted properties) even when invalid, so one parse
// reports as many problems as possible.
func loadDoc(root *xmltree.Node, uri string, errs *xsd.ErrorList) *schemaDoc {
	if root.Name.Space != xsd.XSDNS || root.Name.Local != "schema" {
		// spec: src-schema — XSD 1.1 Part 1 §3.17.3 (src-schema)
		errs.Addf(xsd.SpecSrcSchema, root.Pos, "root element must be <xs:schema>, found {%s}%s", root.Name.Space, root.Name.Local)
		return nil
	}
	doc := &schemaDoc{
		root:   root,
		uri:    uri,
		pruned: map[*xmltree.Node]bool{},
	}
	doc.targetNamespace, _ = root.Attr("targetNamespace")

	w := &walker{errs: errs, doc: doc, ids: map[string]xsd.Pos{}}
	w.validate(root, "schema")

	// Document-level properties. Value errors were already reported by the
	// table checks; here invalid values just fall back to defaults.
	doc.version, _ = root.Attr("version")
	if v, ok := root.Attr("elementFormDefault"); ok && strings.TrimSpace(v) == "qualified" {
		doc.elementFormDefault = xsd.FormQualified
	}
	if v, ok := root.Attr("attributeFormDefault"); ok && strings.TrimSpace(v) == "qualified" {
		doc.attributeFormDefault = xsd.FormQualified
	}
	if v, ok := root.Attr("blockDefault"); ok {
		doc.blockDefault, _ = parseDerivationSet(v, blockDefSet)
	}
	if v, ok := root.Attr("finalDefault"); ok {
		doc.finalDefault, _ = parseDerivationSet(v, finalDefSet)
	}
	if v, ok := root.Attr("defaultAttributes"); ok {
		doc.defaultAttributes, _ = root.ResolveQName(strings.TrimSpace(v))
	}
	if v, ok := root.Attr("xpathDefaultNamespace"); ok {
		doc.xpathDefaultNS = v
	} else {
		doc.xpathDefaultNS = "##local"
	}

	for _, c := range root.Children {
		if doc.pruned[c] || c.Name.Space != xsd.XSDNS {
			continue
		}
		switch c.Name.Local {
		case "include", "redefine", "override":
			loc, _ := c.Attr("schemaLocation")
			doc.compositions = append(doc.compositions, composition{
				kind: c.Name.Local, schemaLocation: loc, node: c,
			})
		case "import":
			ns, _ := c.Attr("namespace")
			loc, _ := c.Attr("schemaLocation")
			doc.compositions = append(doc.compositions, composition{
				kind: "import", namespace: ns, schemaLocation: loc, node: c,
			})
		case "defaultOpenContent":
			doc.defaultOpenContent = c
		}
	}
	return doc
}
