package parser

import (
	"strings"

	"github.com/kud360/goxsd5/parser/xmltree"
	"github.com/kud360/goxsd5/xpath"
	"github.com/kud360/goxsd5/xsd"
)

// checkXPathTest validates the {test} XPath expression of an <alternative>
// (conditional type assignment, §3.12). The XSD 1.1 "Type Alternative
// Representation OK" constraint (src-ta, §3.12.3) clause 1 requires the
// expression to contain no static errors as defined in XPath 2.0. Two classes
// of static error are detected here without evaluating the expression:
//
//   - a syntactically malformed expression (it does not parse), and
//   - a cast/castable/treat/instance-of target that names a complex type
//     rather than an atomic type (XPST0051/XPST0080).
//
// The expression is parsed but never evaluated, so a reported error is always a
// genuine static error: this is a necessary condition, never a false positive.
// Anything the check cannot resolve (an unbound prefix, an unknown type) is left
// alone, in keeping with the builder's give-up-rather-than-guess discipline.
func (b *builder) checkXPathTest(test string, n *xmltree.Node, doc *schemaDoc, ref xsd.SpecRef) {
	if strings.TrimSpace(test) == "" {
		return
	}
	expr, err := xpath.Parse(test)
	if err != nil {
		// spec: src-ta — XSD 1.1 Part 1 §3.12.3 (clause 1: no XPath static errors)
		b.errf(ref, n.Pos, "the test expression is not a valid XPath 2.0 expression: %v", err)
		return
	}
	for _, r := range expr.TypeRefs {
		ns, ok := b.resolveXPathPrefix(r.Prefix, n, doc)
		if !ok {
			continue // unbound prefix: not our error to report
		}
		d := b.registryFor(doc).lookup(spaceType, xsd.QName{Namespace: ns, Local: r.Local})
		if d == nil || d.builtin != nil || d.node == nil {
			continue // unknown or built-in type: give up quietly
		}
		if d.node.Name.Local == "complexType" {
			// spec: src-ta — XSD 1.1 Part 1 §3.12.3 (clause 1; XPST0051: a cast
			// or sequence-type target must denote an atomic type)
			b.errf(ref, n.Pos, "the %s target %q in the test expression names a complex type, but an atomic type is required", r.Kind, typeRefName(r))
			return
		}
	}
}

// resolveXPathPrefix maps an XPath namespace prefix to a namespace URI in the
// static context of node n: a bound prefix resolves through the node's in-scope
// namespaces, while the unprefixed (empty) case uses the in-effect
// xpathDefaultNamespace (the default namespace for unprefixed element and type
// names in XPath, per §3.12).
func (b *builder) resolveXPathPrefix(prefix string, n *xmltree.Node, doc *schemaDoc) (string, bool) {
	if prefix == "" {
		return xpathDefaultNS(n, doc), true
	}
	return n.NS.Lookup(prefix)
}

func typeRefName(r xpath.TypeRef) string {
	if r.Prefix == "" {
		return r.Local
	}
	return r.Prefix + ":" + r.Local
}
