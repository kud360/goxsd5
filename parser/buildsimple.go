package parser

// Simple type construction (Part 2 §4.1): restriction with the full facet
// pipeline, list, and union. Facet lexicals (bounds, enumerations) are
// parsed with the *base* type, which yields the base-membership constraints
// (enumeration-valid-restriction etc.) as a side effect of parsing.

import (
	"strings"

	"github.com/kud360/goxsd5/builtin"
	"github.com/kud360/goxsd5/parser/xmltree"
	"github.com/kud360/goxsd5/xsd"
)

func (b *builder) buildSimpleType(n *xmltree.Node, doc *schemaDoc, name xsd.QName) *xsd.SimpleType {
	st := &xsd.SimpleType{
		Name:       name,
		Pos:        n.Pos,
		Final:      derivAttr(n, "final", stFinalSet, doc.finalDefault),
		Annotation: annotationOf(n, doc),
		Extensions: extensionsOf(n),
	}
	deriv := firstChild(n, doc, "restriction", "list", "union")
	if deriv == nil {
		// Already reported by the structural pass.
		return builtin.AnySimpleType
	}
	switch deriv.Name.Local {
	case "restriction":
		b.buildSTRestriction(st, deriv, doc)
	case "list":
		b.buildSTList(st, deriv, doc)
	case "union":
		b.buildSTUnion(st, deriv, doc)
	}
	return st
}

// buildSTRestriction fills st from an xs:restriction inside xs:simpleType.
// It is also the facet half of simpleContent restrictions (buildcomplex).
func (b *builder) buildSTRestriction(st *xsd.SimpleType, r *xmltree.Node, doc *schemaDoc) {
	base := builtin.AnySimpleType
	if q, ok := qnameAttr(r, doc, "base"); ok {
		// spec: cos-st-restricts — XSD 1.1 Part 2 §4.1.6: the special types
		// must not be restriction bases in schema documents.
		if q.Namespace == xsd.XSDNS && (q.Local == "anySimpleType" || q.Local == "anyAtomicType") {
			b.errf(xsd.SpecCosSTRestricts, r.Pos, "xs:%s must not be used directly as a restriction base", q.Local)
		} else {
			base = b.resolveSimpleType(q, r.Pos, doc, xsd.SpecCosSTRestricts)
		}
	} else if inline := firstChild(r, doc, "simpleType"); inline != nil {
		base, _ = b.buildAnonType(r, doc, builtin.AnySimpleType).(*xsd.SimpleType)
	}

	b.checkFinalAllows(base, xsd.DeriveRestriction, r.Pos)
	b.applyRestriction(st, base, r, doc)
}

// applyRestriction derives st from base by the facets declared on r. It is
// shared between named/anonymous simple types and the simple content of
// complex types.
func (b *builder) applyRestriction(st *xsd.SimpleType, base *xsd.SimpleType, r *xmltree.Node, doc *schemaDoc) {
	st.BaseType = base
	st.Variety = base.Variety
	st.Primitive = base.Primitive
	st.ItemType = base.ItemType
	st.MemberTypes = base.MemberTypes

	declared := b.buildFacets(r, doc, base)
	cmp := base.EffectiveCompare()
	b.errs.Add(xsd.CheckFacetRestriction(declared, &base.Facets, cmp))
	eff := xsd.MergeFacets(&base.Facets, declared)
	b.errs.Add(xsd.ValidateFacetSet(&eff, cmp))
	st.DeclaredFacets = *declared
	st.Facets = eff

	// Fundamental facets (Part 2 §F): ordering and numericness are those of
	// the base; boundedness and cardinality can only tighten.
	st.Ordered = base.Ordered
	st.Numeric = base.Numeric
	st.Bounded = base.Bounded ||
		((eff.MinInclusive != nil || eff.MinExclusive != nil) &&
			(eff.MaxInclusive != nil || eff.MaxExclusive != nil))
	st.Cardinality = base.Cardinality
	if eff.HasEnumeration || eff.Length != nil || eff.MaxLength != nil || eff.TotalDigits != nil {
		st.Cardinality = xsd.CardinalityFinite
	}
}

func (b *builder) buildSTList(st *xsd.SimpleType, l *xmltree.Node, doc *schemaDoc) {
	item := builtin.AnySimpleType
	if q, ok := qnameAttr(l, doc, "itemType"); ok {
		item = b.resolveSimpleType(q, l.Pos, doc, xsd.SpecCosSTRestricts)
	} else if inline := firstChild(l, doc, "simpleType"); inline != nil {
		item, _ = b.buildAnonType(l, doc, builtin.AnySimpleType).(*xsd.SimpleType)
	}
	// spec: cos-st-restricts.2 — XSD 1.1 Part 2 §4.1.6: a list's item type
	// must be atomic or a union of atomics.
	switch item.Variety {
	case xsd.VarietyList:
		b.errf(xsd.SpecCosSTRestricts, l.Pos, "item type of a list must not itself be a list")
	case xsd.VarietyUnion:
		for _, m := range item.MemberTypes {
			if m.Variety != xsd.VarietyAtomic {
				b.errf(xsd.SpecCosSTRestricts, l.Pos, "item type of a list must be a union of atomic types; member %s is not atomic", m.TypeName())
			}
		}
	}
	if item.Final.Has(xsd.DeriveList) {
		// spec: st-props-correct.3 — final of the item type forbids list.
		b.errf(xsd.SpecSTPropsCorrect, l.Pos, "item type %s forbids derivation by list", item.TypeName())
	}

	st.BaseType = builtin.AnySimpleType
	st.Variety = xsd.VarietyList
	st.ItemType = item
	st.Facets.WhiteSpace = xsd.WSCollapse
	st.Facets.WhiteSpaceFixed = true
	st.Cardinality = xsd.CardinalityCountablyInfinite
}

func (b *builder) buildSTUnion(st *xsd.SimpleType, u *xmltree.Node, doc *schemaDoc) {
	var members []*xsd.SimpleType
	if v, ok := u.Attr("memberTypes"); ok {
		for _, tok := range strings.Fields(v) {
			q, err := u.ResolveQName(tok)
			if err != nil {
				continue // reported by pass 1
			}
			members = append(members, b.resolveSimpleType(chameleonQName(q, doc), u.Pos, doc, xsd.SpecCosSTRestricts))
		}
	}
	for _, c := range xsdElems(u, doc) {
		if c.Name.Local == "simpleType" {
			m, _ := b.memoType(c, func() xsd.Type { return b.buildSimpleType(c, doc, xsd.QName{}) }).(*xsd.SimpleType)
			if m != nil {
				members = append(members, m)
			}
		}
	}

	// Union members of union variety are replaced by their own members
	// (transitively, since members were built the same way).
	var flat []*xsd.SimpleType
	for _, m := range members {
		if m.Final.Has(xsd.DeriveUnion) {
			// spec: st-props-correct.3 — final of a member forbids union.
			b.errf(xsd.SpecSTPropsCorrect, u.Pos, "member type %s forbids derivation by union", m.TypeName())
		}
		if m.Variety == xsd.VarietyUnion {
			flat = append(flat, m.MemberTypes...)
		} else {
			flat = append(flat, m)
		}
	}

	st.BaseType = builtin.AnySimpleType
	st.Variety = xsd.VarietyUnion
	st.MemberTypes = flat
	st.Cardinality = xsd.CardinalityCountablyInfinite
}

// facetMask says which constraining facets a type admits
// (cos-applicable-facets, Part 2 §4.1.6 / the per-facet applicability lists).
type facetMask uint

const (
	fLength facetMask = 1 << iota
	fMinLength
	fMaxLength
	fPattern
	fEnumeration
	fWhiteSpace
	fBounds // the four min/max facets
	fTotalDigits
	fFractionDigits
	fAssertion
	fExplicitTimezone
)

const (
	lengthFacets = fLength | fMinLength | fMaxLength
	commonFacets = fPattern | fEnumeration | fWhiteSpace | fAssertion
	allFacets    = ^facetMask(0)
)

var facetNameMask = map[string]facetMask{
	"length": fLength, "minLength": fMinLength, "maxLength": fMaxLength,
	"pattern": fPattern, "enumeration": fEnumeration, "whiteSpace": fWhiteSpace,
	"minInclusive": fBounds, "maxInclusive": fBounds,
	"minExclusive": fBounds, "maxExclusive": fBounds,
	"totalDigits": fTotalDigits, "fractionDigits": fFractionDigits,
	"assertion": fAssertion, "explicitTimezone": fExplicitTimezone,
}

var primitiveFacetMask = map[string]facetMask{
	"string": commonFacets | lengthFacets, "anyURI": commonFacets | lengthFacets,
	"hexBinary": commonFacets | lengthFacets, "base64Binary": commonFacets | lengthFacets,
	"QName": commonFacets | lengthFacets, "NOTATION": commonFacets | lengthFacets,
	"boolean": fPattern | fWhiteSpace | fAssertion,
	"decimal": commonFacets | fBounds | fTotalDigits | fFractionDigits,
	"float":   commonFacets | fBounds, "double": commonFacets | fBounds,
	"duration":   commonFacets | fBounds,
	"dateTime":   commonFacets | fBounds | fExplicitTimezone,
	"time":       commonFacets | fBounds | fExplicitTimezone,
	"date":       commonFacets | fBounds | fExplicitTimezone,
	"gYearMonth": commonFacets | fBounds | fExplicitTimezone,
	"gYear":      commonFacets | fBounds | fExplicitTimezone,
	"gMonthDay":  commonFacets | fBounds | fExplicitTimezone,
	"gDay":       commonFacets | fBounds | fExplicitTimezone,
	"gMonth":     commonFacets | fBounds | fExplicitTimezone,
}

func applicableFacets(base *xsd.SimpleType) facetMask {
	switch base.Variety {
	case xsd.VarietyList:
		return commonFacets | lengthFacets
	case xsd.VarietyUnion:
		return fPattern | fEnumeration | fAssertion
	}
	if base.Primitive == nil {
		// anySimpleType fallback after an earlier error: don't cascade.
		return allFacets
	}
	if m, ok := primitiveFacetMask[base.Primitive.Name.Local]; ok {
		return m
	}
	return allFacets
}

// buildFacets collects the facets declared on one restriction step. Facet
// values are parsed with the base type so base-membership violations
// surface here.
func (b *builder) buildFacets(r *xmltree.Node, doc *schemaDoc, base *xsd.SimpleType) *xsd.Facets {
	var f xsd.Facets
	applic := applicableFacets(base)
	seen := map[string]xsd.Pos{}
	var enums []xsd.Enum
	var patterns xsd.PatternGroup

	intFacet := func(c *xmltree.Node) *xsd.IntFacet {
		v, _ := c.Attr("value")
		i, err := parseNonNegInt(v)
		if err != nil {
			return nil // reported by pass 1
		}
		return &xsd.IntFacet{Value: i, Fixed: boolAttr(c, "fixed", false), Pos: c.Pos}
	}
	bound := func(c *xmltree.Node, ref xsd.SpecRef) *xsd.Bound {
		lex, ok := c.Attr("value")
		if !ok {
			return nil
		}
		v, err := base.ParseFacetValue(lex, nsContext{c})
		if err != nil {
			b.errf(ref, c.Pos, "%s value %q is not valid against the base type %s: %v", c.Name.Local, lex, base.TypeName(), err)
			return nil
		}
		return &xsd.Bound{Value: v, Lexical: strings.TrimSpace(lex), Fixed: boolAttr(c, "fixed", false), Pos: c.Pos}
	}

	for _, c := range xsdElems(r, doc) {
		local := c.Name.Local
		mask, isFacet := facetNameMask[local]
		if !isFacet {
			continue // annotation, inline simpleType, attribute uses, assert
		}
		if applic&mask == 0 {
			// spec: cos-applicable-facets — XSD 1.1 Part 1 §4.1.6
			b.errf(xsd.SpecCosApplicableFacets, c.Pos, "facet %s is not applicable to a type derived from %s", local, base.TypeName())
			continue
		}
		if local != "pattern" && local != "enumeration" && local != "assertion" {
			if prev, dup := seen[local]; dup {
				// spec: src-single-facet-value — XSD 1.1 Part 2 §4.3
				err := xsd.NewError(xsd.SpecSingleFacetValue, c.Pos, "facet %s appears more than once in one restriction step", local)
				err.OtherPos = prev
				b.errs.Add(err)
				continue
			}
			seen[local] = c.Pos
		}

		switch local {
		case "length":
			f.Length = intFacet(c)
		case "minLength":
			f.MinLength = intFacet(c)
		case "maxLength":
			f.MaxLength = intFacet(c)
		case "totalDigits":
			f.TotalDigits = intFacet(c)
		case "fractionDigits":
			f.FractionDigits = intFacet(c)
		case "minInclusive":
			f.MinInclusive = bound(c, xsd.SpecMinInclValidRestriction)
		case "maxInclusive":
			f.MaxInclusive = bound(c, xsd.SpecMaxInclValidRestriction)
		case "minExclusive":
			f.MinExclusive = bound(c, xsd.SpecMinExclValidRestriction)
		case "maxExclusive":
			f.MaxExclusive = bound(c, xsd.SpecMaxExclValidRestriction)
		case "whiteSpace":
			switch v, _ := c.Attr("value"); strings.TrimSpace(v) {
			case "preserve":
				f.WhiteSpace = xsd.WSPreserve
			case "replace":
				f.WhiteSpace = xsd.WSReplace
			case "collapse":
				f.WhiteSpace = xsd.WSCollapse
			}
			f.WhiteSpaceFixed = boolAttr(c, "fixed", false)
			f.WhiteSpacePos = c.Pos
		case "explicitTimezone":
			switch v, _ := c.Attr("value"); strings.TrimSpace(v) {
			case "optional":
				f.ExplicitTimezone = xsd.ETZOptional
			case "required":
				f.ExplicitTimezone = xsd.ETZRequired
			case "prohibited":
				f.ExplicitTimezone = xsd.ETZProhibited
			}
			f.ExplicitTimezoneFixed = boolAttr(c, "fixed", false)
			f.ExplicitTimezonePos = c.Pos
		case "pattern":
			src, _ := c.Attr("value")
			re, err := xsd.CompileRegex(src)
			if err != nil {
				// spec: regex-valid — XSD 1.1 Part 2 Appendix G
				b.errf(xsd.SpecRegexValid, c.Pos, "invalid pattern: %v", err)
				continue
			}
			patterns = append(patterns, xsd.Pattern{Source: src, Re: re, Pos: c.Pos})
		case "enumeration":
			lex, _ := c.Attr("value")
			v, err := base.ParseValue(lex, nsContext{c})
			if err != nil {
				// spec: enumeration-valid-restriction — XSD 1.1 Part 2
				// §4.3.5.5: every enumeration value must be valid against
				// the base type.
				b.errf(xsd.SpecEnumerationValidRestriction, c.Pos, "enumeration value %q is not valid against the base type %s: %v", lex, base.TypeName(), err)
				continue
			}
			enums = append(enums, xsd.Enum{Value: v, Lexical: lex, Pos: c.Pos})
		case "assertion":
			f.Assertions = append(f.Assertions, b.buildAssertion(c, doc))
		}
	}

	if len(patterns) > 0 {
		f.PatternGroups = []xsd.PatternGroup{patterns}
	}
	if firstChild(r, doc, "enumeration") != nil {
		f.HasEnumeration = true
		f.Enumeration = enums
	}
	return &f
}

// buildAssertion captures an assertion/assert facet without evaluating it.
func (b *builder) buildAssertion(c *xmltree.Node, doc *schemaDoc) xsd.Assertion {
	test, _ := c.Attr("test")
	return xsd.Assertion{
		Test:             test,
		DefaultNamespace: xpathDefaultNS(c, doc),
		Pos:              c.Pos,
		Extensions:       extensionsOf(c),
	}
}

// xpathDefaultNS resolves the xpathDefaultNamespace in effect at n: its own
// attribute, else the schema document's, with the ## keywords expanded.
func xpathDefaultNS(n *xmltree.Node, doc *schemaDoc) string {
	v, ok := n.Attr("xpathDefaultNamespace")
	if !ok {
		v = doc.xpathDefaultNS
	}
	switch v = strings.TrimSpace(v); v {
	case "##local":
		return ""
	case "##targetNamespace":
		return doc.targetNamespace
	case "##defaultNamespace":
		uri, _ := n.NS.Lookup("")
		return uri
	}
	return v
}
