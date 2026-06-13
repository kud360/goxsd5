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

	"github.com/kud360/goxsd5/builtin"
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

	// If the <schema> element itself is conditionally excluded (§4.2.2), the
	// document becomes empty: every child (including compositions) is removed
	// and no content is judged. Only targetNamespace/minVersion/maxVersion
	// would survive, none of which we validate further here.
	if variant == "schema" && w.prunedByConditionalInclusion(n) {
		for _, c := range n.Children {
			w.doc.pruned[c] = true
		}
		return
	}

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
		if w.prunedByConditionalInclusion(c) {
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

// prunedByConditionalInclusion implements XSD 1.1 conditional inclusion
// (§4.2.2): an element carrying vc:* attributes is removed unless every
// condition holds — minVersion <= V < maxVersion for this processor's version
// V, every vc:typeAvailable/facetAvailable item is known, and not every
// vc:typeUnavailable/facetUnavailable item is known. Each attribute value
// must also be locally valid (decimal, or list of QName); a malformed value
// is reported (cip) even though it does not by itself stop the element from
// being pruned.
func (w *walker) prunedByConditionalInclusion(n *xmltree.Node) bool {
	pruned := false
	for i := range n.Attrs {
		a := &n.Attrs[i]
		if a.Name.Space != xsd.VCNS {
			continue
		}
		switch a.Name.Local {
		case "minVersion", "maxVersion":
			if _, err := xsd.ParseDecimal(a.Value); err != nil {
				// spec: cip — XSD 1.1 Part 1 §4.2.2: the value must be a valid xs:decimal.
				w.errf(xsd.SpecCIP, a.Pos, "vc:%s value %q is not a valid xs:decimal", a.Name.Local, a.Value)
				continue
			}
			v, _ := strconv.ParseFloat(strings.TrimSpace(a.Value), 64)
			if a.Name.Local == "minVersion" && v > xsdVersion {
				pruned = true
			}
			if a.Name.Local == "maxVersion" && v <= xsdVersion {
				pruned = true
			}
		case "typeAvailable", "typeUnavailable", "facetAvailable", "facetUnavailable":
			isType := strings.HasPrefix(a.Name.Local, "type")
			// An element is pruned when any "Available" item is unknown, or
			// when every "Unavailable" item is known (empty list = vacuously
			// every, so it prunes).
			allKnown, anyUnknown := true, false
			for _, tok := range strings.Fields(a.Value) {
				q, err := n.ResolveQName(tok)
				if err != nil {
					// spec: cip — the value must be a valid list of xs:QName.
					w.errf(xsd.SpecCIP, a.Pos, "vc:%s item %q is not a valid QName", a.Name.Local, tok)
					continue
				}
				known := typeAutomaticallyKnown(q)
				if !isType {
					known = facetKnown(q)
				}
				if !known {
					allKnown, anyUnknown = false, true
				}
			}
			switch a.Name.Local {
			case "typeAvailable", "facetAvailable":
				if anyUnknown {
					pruned = true
				}
			case "typeUnavailable", "facetUnavailable":
				if allKnown {
					pruned = true
				}
			}
		}
	}
	return pruned
}

// typeAutomaticallyKnown reports whether q names a type the processor provides
// automatically: a built-in datatype, or one of the special types. User types
// are not "automatically known" — conditional inclusion runs before assembly.
func typeAutomaticallyKnown(q xsd.QName) bool {
	if q.Namespace != xsd.XSDNS {
		return false
	}
	switch q.Local {
	case "anyType", "anySimpleType", "anyAtomicType", "error":
		return true
	}
	return builtin.Lookup(q.Local) != nil
}

// facetKnown reports whether q names a constraining facet this processor
// supports (by the element name used to apply it). xs:minScale / xs:maxScale
// (precisionDecimal facets) are deliberately absent.
func facetKnown(q xsd.QName) bool {
	if q.Namespace != xsd.XSDNS {
		return false
	}
	switch q.Local {
	case "length", "minLength", "maxLength", "pattern", "enumeration", "whiteSpace",
		"maxInclusive", "maxExclusive", "minInclusive", "minExclusive",
		"totalDigits", "fractionDigits", "assertion", "explicitTimezone":
		return true
	}
	return false
}
