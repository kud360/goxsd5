package parser

// The table-driven structural validator: one recursive walk per schema
// document that checks every xs:* element against elemTable (attributes,
// attribute values, child content model, stray text), enforces per-document
// xs:ID uniqueness (src-id), and applies XSD 1.1 conditional inclusion
// (vc:minVersion/vc:maxVersion pruning) before any content is judged.

import (
	"errors"
	"strconv"
	"strings"

	"github.com/kud360/goxsd5/parser/xmltree"
	"github.com/kud360/goxsd5/xsd"
)

// xsdVersion is the schema language version this processor implements,
// compared against vc:minVersion/vc:maxVersion.
const xsdVersion = 1.1

type walker struct {
	errs *xsd.ErrorList
	doc  *schemaDoc
	// ids maps each xs:ID value seen in this document to its first position.
	ids map[string]xsd.Pos
}

func (w *walker) errf(ref xsd.SpecRef, pos xsd.Pos, format string, args ...any) {
	w.errs.Addf(ref, pos, format, args...)
}

// validate checks n against the table entry for variant and recurses.
func (w *walker) validate(n *xmltree.Node, variant string) {
	spec := elemTable[variant]
	if spec == nil {
		// Defensive: every variant the walker can reach is in the table.
		panic("parser: no table entry for variant " + variant)
	}
	w.checkAttrs(n, spec)
	if spec.extra != nil {
		spec.extra(w, n)
	}
	if spec.freeContent {
		// appinfo/documentation: content is {any}, nothing to check and
		// nothing to recurse into.
		return
	}
	if strings.TrimSpace(n.CharData) != "" {
		w.errf(spec.ref, n.Pos, "<xs:%s> must not contain character data", n.Name.Local)
	}

	// Conditional inclusion first: pruned children are invisible to both
	// the content model and recursion.
	// XSD 1.1 Part 1 §4.2.2 (conditional inclusion, src-cip); no error is
	// raised — matching nodes are pruned, not rejected.
	var kept []*xmltree.Node
	for _, c := range n.Children {
		if w.prunedByVersion(c) {
			w.doc.pruned[c] = true
			continue
		}
		kept = append(kept, c)
	}

	var toks []string
	var tokNodes []*xmltree.Node
	for _, c := range kept {
		if c.Name.Space != xsd.XSDNS {
			if !spec.allowForeign {
				w.errf(spec.ref, c.Pos, "element {%s}%s is not allowed inside <xs:%s>", c.Name.Space, c.Name.Local, n.Name.Local)
			}
			continue
		}
		toks = append(toks, c.Name.Local)
		tokNodes = append(tokNodes, c)
	}

	if ok, failIdx, expected := matchContent(spec.content, toks); !ok {
		want := strings.Join(expected, ", ")
		if failIdx < len(toks) {
			w.errf(spec.ref, tokNodes[failIdx].Pos, "element <xs:%s> not allowed here inside <xs:%s> (expected %s)", toks[failIdx], n.Name.Local, want)
		} else {
			w.errf(spec.ref, n.Pos, "incomplete content in <xs:%s> (expected %s)", n.Name.Local, want)
		}
	}

	for _, c := range tokNodes {
		childVariant := c.Name.Local
		if v, ok := spec.children[c.Name.Local]; ok {
			childVariant = v
		}
		if _, known := elemTable[childVariant]; known {
			w.validate(c, childVariant)
		}
		// Unknown children were already reported by the content model.
	}
}

func (w *walker) checkAttrs(n *xmltree.Node, spec *elemSpec) {
	seen := map[string]bool{}
	for i := range n.Attrs {
		a := &n.Attrs[i]
		switch a.Name.Space {
		case "":
			as, ok := spec.attrs[a.Name.Local]
			if !ok {
				w.errf(spec.ref, a.Pos, "attribute %q is not allowed on <xs:%s>", a.Name.Local, n.Name.Local)
				continue
			}
			seen[a.Name.Local] = true
			if err := as.check(a.Value, n); err != nil {
				ref := spec.ref
				var re *refError
				if errors.As(err, &re) {
					ref = re.ref
				}
				w.errf(ref, a.Pos, "invalid %s on <xs:%s>: %v", a.Name.Local, n.Name.Local, err)
			} else if a.Name.Local == "id" {
				w.recordID(a.Value, a.Pos)
			}
		case xsd.XSDNS:
			w.errf(spec.ref, a.Pos, "attribute %s is not allowed: schema elements take no XSD-namespace attributes", a.Name.Local)
		default:
			// Foreign-namespace attributes (including vc:* and xml:*) are
			// always allowed; pass 2 captures them as Extensions.
		}
	}
	for name, as := range spec.attrs {
		if as.required && !seen[name] {
			w.errf(spec.ref, n.Pos, "<xs:%s> requires attribute %q", n.Name.Local, name)
		}
	}
}

// recordID enforces per-document uniqueness of xs:ID-typed id attributes.
func (w *walker) recordID(v string, pos xsd.Pos) {
	v = strings.TrimSpace(v)
	if prev, dup := w.ids[v]; dup {
		// spec: src-id — XSD 1.1 Part 1 §3.17.3: attributes of type ID must
		// have unique values within a schema document.
		err := xsd.NewError(xsd.SpecSrcDupID, pos, "duplicate id %q in schema document", v)
		err.OtherPos = prev
		w.errs.Add(err)
		return
	}
	w.ids[v] = pos
}

// checkOccurs validates minOccurs <= maxOccurs on a particle-bearing element.
func (w *walker) checkOccurs(n *xmltree.Node) {
	min, max := 1, 1
	if v, ok := n.Attr("minOccurs"); ok {
		if i, err := parseNonNegInt(v); err == nil {
			min = i
		}
	}
	if v, ok := n.Attr("maxOccurs"); ok {
		if i, err := parseMaxOccurs(v); err == nil {
			max = i
		}
	}
	if max != xsd.UnboundedOccurs && min > max {
		// spec: p-props-correct.2.1 — XSD 1.1 Part 1 §3.9.6 (p-props-correct)
		w.errf(xsd.SpecPPropsCorrect, n.Pos, "minOccurs (%d) must not exceed maxOccurs (%d)", min, max)
	}
}

// prunedByVersion implements vc:minVersion/vc:maxVersion conditional
// inclusion: an element is retained iff minVersion <= V < maxVersion for
// this processor's version V. Unparseable version numbers are ignored, and
// the other vc:* conditions (typeAvailable, facetAvailable, …) are not
// evaluated, so their elements are retained — the lenient default.
func (w *walker) prunedByVersion(n *xmltree.Node) bool {
	for i := range n.Attrs {
		a := &n.Attrs[i]
		if a.Name.Space != xsd.VCNS {
			continue
		}
		v, err := strconv.ParseFloat(strings.TrimSpace(a.Value), 64)
		if err != nil {
			continue
		}
		switch a.Name.Local {
		case "minVersion":
			if v > xsdVersion {
				return true
			}
		case "maxVersion":
			if v <= xsdVersion {
				return true
			}
		}
	}
	return false
}
