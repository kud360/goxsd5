package parser

// The structural validation table: one entry per xs:* element *as used in a
// particular context* (variants like "element@global" vs "element@local"
// differ in allowed attributes and extra rules). Attribute sets and content
// models are transcribed from the XML Representation Summaries of XSD 1.1
// Part 1 §3–4 and Part 2 §4.3; per-context attribute prohibitions follow
// the schema for schema documents (topLevelElement, localElement, …).

import (
	"maps"
	"strings"

	"github.com/kud360/goxsd5/parser/xmltree"
	"github.com/kud360/goxsd5/xsd"
)

type attrSpec struct {
	required bool
	check    valueCheck
}

type elemSpec struct {
	// ref is the src-* (or equivalent) constraint reported for structural
	// errors on this element: bad attributes, bad children, stray text.
	ref xsd.SpecRef
	// attrs are the permitted no-namespace attributes.
	attrs map[string]attrSpec
	// content is the child content model over XSD-namespace local names.
	content cm
	// children maps a child local name to its table variant; absent keys
	// default to the local name itself.
	children map[string]string
	// allowForeign permits foreign-namespace child elements (the
	// `{any with namespace: ##other}` slot of simple-type restrictions).
	allowForeign bool
	// freeContent skips content validation entirely (appinfo/documentation,
	// whose content is `{any}`).
	freeContent bool
	// extra applies cross-attribute/child rules the table cannot express.
	extra func(w *walker, n *xmltree.Node)
}

// Derivation-set vocabularies per declaring attribute.
const (
	eltBlockSet = xsd.DerivationSet(xsd.DeriveExtension | xsd.DeriveRestriction | xsd.DeriveSubstitution)
	eltFinalSet = xsd.DerivationSet(xsd.DeriveExtension | xsd.DeriveRestriction)
	ctBlockSet  = xsd.DerivationSet(xsd.DeriveExtension | xsd.DeriveRestriction)
	ctFinalSet  = xsd.DerivationSet(xsd.DeriveExtension | xsd.DeriveRestriction)
	// A simple type's {final} may contain any of restriction/extension/list/
	// union (XSD 1.1 permits simpleType/@final="extension" — spec bug 2074 —
	// even though it has no effect unless the type is used as an extension
	// base, which is then blocked).
	stFinalSet  = xsd.DerivationSet(xsd.DeriveList | xsd.DeriveUnion | xsd.DeriveRestriction | xsd.DeriveExtension)
	blockDefSet = eltBlockSet
	finalDefSet = stFinalSet
)

func req(c valueCheck) attrSpec  { return attrSpec{required: true, check: c} }
func optA(c valueCheck) attrSpec { return attrSpec{check: c} }

// Shared content-model fragments.
var (
	annQ      = opt(one("annotation"))
	annOnly   = annQ
	attrDecls = seq(star(names("attribute", "attributeGroup")), opt(one("anyAttribute")))
	particle  = names("group", "all", "choice", "sequence")
	asserts   = star(one("assert"))
	facetsCM  = star(names(
		"minExclusive", "minInclusive", "maxExclusive", "maxInclusive",
		"totalDigits", "fractionDigits", "length", "minLength", "maxLength",
		"enumeration", "whiteSpace", "pattern", "assertion", "explicitTimezone"))
)

// Child-variant maps shared by several entries.
var (
	topLevelChildren = map[string]string{
		"element": "element@global", "attribute": "attribute@global",
		"group": "group@def", "attributeGroup": "attributeGroup@def",
		"simpleType": "simpleType@global", "complexType": "complexType@global",
	}
	particleChildren = map[string]string{
		"element": "element@local", "group": "group@ref",
	}
	ctChildren = map[string]string{
		"group": "group@ref", "all": "all", "choice": "choice", "sequence": "sequence",
		"attribute": "attribute@local", "attributeGroup": "attributeGroup@ref",
	}
	inlineTypeChildren = map[string]string{
		"simpleType": "simpleType@local", "complexType": "complexType@local",
	}
)

var elemTable = map[string]*elemSpec{
	"schema": {
		ref: xsd.SpecSrcSchema,
		attrs: map[string]attrSpec{
			"attributeFormDefault":  optA(vcEnum("qualified", "unqualified")),
			"elementFormDefault":    optA(vcEnum("qualified", "unqualified")),
			"blockDefault":          optA(vcDerivSet(blockDefSet)),
			"finalDefault":          optA(vcDerivSet(finalDefSet)),
			"defaultAttributes":     optA(vcQName),
			"xpathDefaultNamespace": optA(vcAny),
			"id":                    optA(vcID),
			"targetNamespace":       optA(vcAny),
			"version":               optA(vcAny),
		},
		content: seq(
			star(names("include", "import", "redefine", "override", "annotation")),
			opt(seq(one("defaultOpenContent"), star(one("annotation")))),
			star(seq(
				names("simpleType", "complexType", "group", "attributeGroup",
					"element", "attribute", "notation"),
				star(one("annotation"))))),
		children: topLevelChildren,
		extra: func(w *walker, n *xmltree.Node) {
			// spec: src-schema — XSD 1.1 Part 1 §3.17.3 (src-schema): the
			// empty string is not a legal targetNamespace; use absence.
			if v, ok := n.Attr("targetNamespace"); ok && v == "" {
				w.errf(xsd.SpecSrcSchema, n.Pos, "targetNamespace must not be empty (omit the attribute for no namespace)")
			}
		},
	},

	"include": {
		ref: xsd.SpecSrcInclude,
		attrs: map[string]attrSpec{
			"id": optA(vcID), "schemaLocation": req(vcAny),
		},
		content: annOnly,
	},
	"import": {
		ref: xsd.SpecSrcImport,
		attrs: map[string]attrSpec{
			"id": optA(vcID), "namespace": optA(vcAny), "schemaLocation": optA(vcAny),
		},
		content: annOnly,
		extra: func(w *walker, n *xmltree.Node) {
			ns, hasNS := n.Attr("namespace")
			// spec: src-import.1 — XSD 1.1 Part 1 §4.2.6.2 (src-import)
			if hasNS && ns == w.doc.targetNamespace && ns != "" {
				w.errf(xsd.SpecSrcImport, n.Pos, "import namespace must differ from the importing schema's targetNamespace %q", ns)
			}
			if hasNS && ns == "" {
				w.errf(xsd.SpecSrcImport, n.Pos, "import namespace must not be empty (omit the attribute to import no-namespace components)")
			}
			if !hasNS && w.doc.targetNamespace == "" {
				w.errf(xsd.SpecSrcImport, n.Pos, "import without namespace requires the importing schema to have a targetNamespace")
			}
		},
	},
	"redefine": {
		ref: xsd.SpecSrcRedefine,
		attrs: map[string]attrSpec{
			"id": optA(vcID), "schemaLocation": req(vcAny),
		},
		content:  star(names("annotation", "simpleType", "complexType", "group", "attributeGroup")),
		children: topLevelChildren,
	},
	"override": {
		ref: xsd.SpecSrcOverride,
		attrs: map[string]attrSpec{
			"id": optA(vcID), "schemaLocation": req(vcAny),
		},
		content: star(names("annotation", "simpleType", "complexType", "group",
			"attributeGroup", "element", "attribute", "notation")),
		children: topLevelChildren,
	},

	"annotation": {
		ref:     xsd.SpecSrcAnnotation,
		attrs:   map[string]attrSpec{"id": optA(vcID)},
		content: star(names("appinfo", "documentation")),
	},
	"appinfo": {
		ref:         xsd.SpecSrcAnnotation,
		attrs:       map[string]attrSpec{"source": optA(vcAny)},
		freeContent: true,
	},
	"documentation": {
		ref:         xsd.SpecSrcAnnotation,
		attrs:       map[string]attrSpec{"source": optA(vcAny)},
		freeContent: true,
	},

	// --- simple type definitions -------------------------------------

	"simpleType@global": {
		ref: xsd.SpecSrcSimpleType,
		attrs: map[string]attrSpec{
			"final": optA(vcDerivSet(stFinalSet)), "id": optA(vcID), "name": req(vcNCName),
		},
		content:  seq(annQ, names("restriction", "list", "union")),
		children: map[string]string{"restriction": "restriction@simple"},
	},
	"simpleType@local": {
		ref:      xsd.SpecSrcSimpleType,
		attrs:    map[string]attrSpec{"id": optA(vcID)},
		content:  seq(annQ, names("restriction", "list", "union")),
		children: map[string]string{"restriction": "restriction@simple"},
	},
	"restriction@simple": {
		ref: xsd.SpecSrcRestriction,
		attrs: map[string]attrSpec{
			"base": optA(vcQName), "id": optA(vcID),
		},
		content:      seq(annQ, opt(one("simpleType")), facetsCM),
		children:     inlineTypeChildren,
		allowForeign: true,
		extra: func(w *walker, n *xmltree.Node) {
			// spec: src-restriction-base-or-simpleType — XSD 1.1 Part 1
			// §3.16.3 (src-restriction-base-or-simpleType)
			_, hasBase := n.Attr("base")
			hasInline := countChildren(n, "simpleType") > 0
			if hasBase == hasInline {
				w.errf(xsd.SpecSrcRestriction, n.Pos, "restriction requires either a base attribute or one inline <simpleType>, not %s", bothOrNeither(hasBase))
			}
		},
	},
	"list": {
		ref: xsd.SpecSrcList,
		attrs: map[string]attrSpec{
			"id": optA(vcID), "itemType": optA(vcQName),
		},
		content:  seq(annQ, opt(one("simpleType"))),
		children: inlineTypeChildren,
		extra: func(w *walker, n *xmltree.Node) {
			// spec: src-list-itemType-or-simpleType — XSD 1.1 Part 1 §3.16.3
			_, hasItem := n.Attr("itemType")
			hasInline := countChildren(n, "simpleType") > 0
			if hasItem == hasInline {
				w.errf(xsd.SpecSrcList, n.Pos, "list requires either an itemType attribute or one inline <simpleType>, not %s", bothOrNeither(hasItem))
			}
		},
	},
	"union": {
		ref: xsd.SpecSrcUnion,
		attrs: map[string]attrSpec{
			"id": optA(vcID), "memberTypes": optA(vcQNameList),
		},
		content:  seq(annQ, star(one("simpleType"))),
		children: inlineTypeChildren,
		extra: func(w *walker, n *xmltree.Node) {
			// spec: src-union-memberTypes-or-simpleTypes — XSD 1.1 Part 1 §3.16.3
			mt, _ := n.Attr("memberTypes")
			if len(strings.Fields(mt)) == 0 && countChildren(n, "simpleType") == 0 {
				w.errf(xsd.SpecSrcUnion, n.Pos, "union requires a non-empty memberTypes or at least one inline <simpleType>")
			}
		},
	},

	// --- complex type definitions ------------------------------------

	"complexType@global": {
		ref: xsd.SpecSrcCT,
		attrs: map[string]attrSpec{
			"abstract": optA(vcBool), "block": optA(vcDerivSet(ctBlockSet)),
			"final": optA(vcDerivSet(ctFinalSet)), "id": optA(vcID),
			"mixed": optA(vcBool), "name": req(vcNCName),
			"defaultAttributesApply": optA(vcBool),
		},
		content:  ctContent,
		children: ctChildren,
	},
	"complexType@local": {
		ref: xsd.SpecSrcCT,
		attrs: map[string]attrSpec{
			"id": optA(vcID), "mixed": optA(vcBool),
			"defaultAttributesApply": optA(vcBool),
		},
		content:  ctContent,
		children: ctChildren,
	},
	"simpleContent": {
		ref:     xsd.SpecSrcCT,
		attrs:   map[string]attrSpec{"id": optA(vcID)},
		content: seq(annQ, names("restriction", "extension")),
		children: map[string]string{
			"restriction": "restriction@simpleContent", "extension": "extension@simpleContent",
		},
	},
	"complexContent": {
		ref:     xsd.SpecSrcCT,
		attrs:   map[string]attrSpec{"id": optA(vcID), "mixed": optA(vcBool)},
		content: seq(annQ, names("restriction", "extension")),
		children: map[string]string{
			"restriction": "restriction@complexContent", "extension": "extension@complexContent",
		},
	},
	"restriction@simpleContent": {
		ref: xsd.SpecSrcCT,
		attrs: map[string]attrSpec{
			"base": req(vcQName), "id": optA(vcID),
		},
		content:      seq(annQ, opt(one("simpleType")), facetsCM, attrDecls, asserts),
		children:     mergeChildren(ctChildren, inlineTypeChildren),
		allowForeign: true,
	},
	"extension@simpleContent": {
		ref: xsd.SpecSrcCT,
		attrs: map[string]attrSpec{
			"base": req(vcQName), "id": optA(vcID),
		},
		content:  seq(annQ, attrDecls, asserts),
		children: ctChildren,
	},
	"restriction@complexContent": {
		ref: xsd.SpecSrcCT,
		attrs: map[string]attrSpec{
			"base": req(vcQName), "id": optA(vcID),
		},
		content:  seq(annQ, opt(one("openContent")), opt(particle), attrDecls, asserts),
		children: ctChildren,
	},
	"extension@complexContent": {
		ref: xsd.SpecSrcCT,
		attrs: map[string]attrSpec{
			"base": req(vcQName), "id": optA(vcID),
		},
		content:  seq(annQ, opt(one("openContent")), opt(particle), attrDecls, asserts),
		children: ctChildren,
	},
	"openContent": {
		ref: xsd.SpecSrcCT,
		attrs: map[string]attrSpec{
			"id": optA(vcID), "mode": optA(vcEnum("none", "interleave", "suffix")),
		},
		content: seq(annQ, opt(one("any"))),
		extra: func(w *walker, n *xmltree.Node) {
			// spec: src-ct — XSD 1.1 Part 1 §3.4.2.4: mode="none" admits no
			// wildcard child; other modes require one.
			mode, _ := n.Attr("mode")
			hasAny := countChildren(n, "any") > 0
			if strings.TrimSpace(mode) == "none" && hasAny {
				w.errf(xsd.SpecSrcCT, n.Pos, `openContent with mode="none" must not have an <any> child`)
			}
			if strings.TrimSpace(mode) != "none" && !hasAny {
				w.errf(xsd.SpecSrcCT, n.Pos, "openContent requires an <any> child unless mode is \"none\"")
			}
		},
	},
	"defaultOpenContent": {
		ref: xsd.SpecSrcSchema,
		attrs: map[string]attrSpec{
			"appliesToEmpty": optA(vcBool), "id": optA(vcID),
			"mode": optA(vcEnum("interleave", "suffix")),
		},
		content: seq(annQ, one("any")),
	},
	"assert": {
		ref: xsd.SpecSrcCT,
		attrs: map[string]attrSpec{
			"id": optA(vcID), "test": req(vcXPath), "xpathDefaultNamespace": optA(vcAny),
		},
		content: annOnly,
	},

	// --- element declarations ----------------------------------------

	"element@global": {
		ref: xsd.SpecSrcElement,
		attrs: map[string]attrSpec{
			"abstract": optA(vcBool), "block": optA(vcDerivSet(eltBlockSet)),
			"default": optA(vcAny), "final": optA(vcDerivSet(eltFinalSet)),
			"fixed": optA(vcAny), "id": optA(vcID), "name": req(vcNCName),
			"nillable": optA(vcBool), "substitutionGroup": optA(vcQNameList),
			"type": optA(vcQName),
		},
		content:  elementContent,
		children: mergeChildren(inlineTypeChildren, nil),
		extra:    elementCommonExtra,
	},
	"element@local": {
		ref: xsd.SpecSrcElement,
		attrs: map[string]attrSpec{
			"block": optA(vcDerivSet(eltBlockSet)), "default": optA(vcAny),
			"fixed": optA(vcAny), "form": optA(vcEnum("qualified", "unqualified")),
			"id": optA(vcID), "maxOccurs": optA(vcMaxOccurs), "minOccurs": optA(vcMinOccurs),
			"name": optA(vcNCName), "nillable": optA(vcBool), "ref": optA(vcQName),
			"targetNamespace": optA(vcAny), "type": optA(vcQName),
		},
		content:  elementContent,
		children: mergeChildren(inlineTypeChildren, nil),
		extra: func(w *walker, n *xmltree.Node) {
			elementCommonExtra(w, n)
			_, hasName := n.Attr("name")
			_, hasRef := n.Attr("ref")
			// spec: src-element.2.1 — XSD 1.1 Part 1 §3.3.3 (src-element)
			if hasName == hasRef {
				w.errf(xsd.SpecSrcElement, n.Pos, "element requires either name or ref, not %s", bothOrNeither(hasName))
			}
			if hasRef {
				// spec: src-element.2.2 — only minOccurs/maxOccurs/id may
				// accompany ref; no children other than annotation.
				forbidWithRef(w, n, xsd.SpecSrcElement, "element",
					[]string{"nillable", "default", "fixed", "form", "block", "type", "targetNamespace"},
					[]string{"simpleType", "complexType", "alternative", "unique", "key", "keyref"})
			}
			// spec: src-element.4 — targetNamespace requires form absent.
			if _, hasTNS := n.Attr("targetNamespace"); hasTNS {
				if _, hasForm := n.Attr("form"); hasForm {
					w.errf(xsd.SpecSrcElement, n.Pos, "element targetNamespace and form must not both be present")
				}
			}
			w.checkOccurs(n)
		},
	},
	"alternative": {
		ref: xsd.SpecSrcTA,
		attrs: map[string]attrSpec{
			"id": optA(vcID), "test": optA(vcXPath), "type": optA(vcQName),
			"xpathDefaultNamespace": optA(vcAny),
		},
		content:  seq(annQ, opt(names("simpleType", "complexType"))),
		children: inlineTypeChildren,
		extra: func(w *walker, n *xmltree.Node) {
			// spec: src-ta — XSD 1.1 Part 1 §3.12.3 (src-ta): exactly one of
			// a type attribute, a <simpleType> child, or a <complexType> child.
			count := countChildren(n, "simpleType") + countChildren(n, "complexType")
			if _, hasType := n.Attr("type"); hasType {
				count++
			}
			if count != 1 {
				w.errf(xsd.SpecSrcTA, n.Pos, "alternative requires exactly one of a type attribute or an inline type definition")
			}
		},
	},

	// --- attribute declarations --------------------------------------

	"attribute@global": {
		ref: xsd.SpecSrcAttribute,
		attrs: map[string]attrSpec{
			"default": optA(vcAny), "fixed": optA(vcAny), "id": optA(vcID),
			"name": req(vcNCName), "type": optA(vcQName), "inheritable": optA(vcBool),
		},
		content:  seq(annQ, opt(one("simpleType"))),
		children: inlineTypeChildren,
		extra:    attributeCommonExtra,
	},
	"attribute@local": {
		ref: xsd.SpecSrcAttribute,
		attrs: map[string]attrSpec{
			"default": optA(vcAny), "fixed": optA(vcAny),
			"form": optA(vcEnum("qualified", "unqualified")), "id": optA(vcID),
			"name": optA(vcNCName), "ref": optA(vcQName),
			"targetNamespace": optA(vcAny), "type": optA(vcQName),
			"use":         optA(vcEnum("optional", "prohibited", "required")),
			"inheritable": optA(vcBool),
		},
		content:  seq(annQ, opt(one("simpleType"))),
		children: inlineTypeChildren,
		extra: func(w *walker, n *xmltree.Node) {
			attributeCommonExtra(w, n)
			_, hasName := n.Attr("name")
			_, hasRef := n.Attr("ref")
			// spec: src-attribute.3.1 — XSD 1.1 Part 1 §3.2.3 (src-attribute)
			if hasName == hasRef {
				w.errf(xsd.SpecSrcAttribute, n.Pos, "attribute requires either name or ref, not %s", bothOrNeither(hasName))
			}
			if hasRef {
				// spec: src-attribute.3.2 — only simpleType, form and type
				// are excluded by ref; inheritable may override the
				// declaration's, and targetNamespace already requires name
				// (clause 6.1, checked in attributeCommonExtra).
				forbidWithRef(w, n, xsd.SpecSrcAttribute, "attribute",
					[]string{"form", "type"},
					[]string{"simpleType"})
			}
			// spec: src-attribute.2 — default requires use=optional (or absent).
			if _, hasDefault := n.Attr("default"); hasDefault {
				if use, ok := n.Attr("use"); ok && strings.TrimSpace(use) != "optional" {
					w.errf(xsd.SpecSrcAttribute, n.Pos, `attribute with default requires use="optional"`)
				}
			}
			if _, hasTNS := n.Attr("targetNamespace"); hasTNS {
				if _, hasForm := n.Attr("form"); hasForm {
					w.errf(xsd.SpecSrcAttribute, n.Pos, "attribute targetNamespace and form must not both be present")
				}
			}
		},
	},

	// --- groups, attribute groups, model groups ----------------------

	"group@def": {
		ref: xsd.SpecSrcModelGroup,
		attrs: map[string]attrSpec{
			"id": optA(vcID), "name": req(vcNCName),
		},
		content: seq(annQ, names("all", "choice", "sequence")),
		children: map[string]string{
			"all": "all@def", "choice": "choice@def", "sequence": "sequence@def",
		},
	},
	"group@ref": {
		ref: xsd.SpecSrcModelGroup,
		attrs: map[string]attrSpec{
			"id": optA(vcID), "ref": req(vcQName),
			"maxOccurs": optA(vcMaxOccurs), "minOccurs": optA(vcMinOccurs),
		},
		content: annOnly,
		extra:   func(w *walker, n *xmltree.Node) { w.checkOccurs(n) },
	},
	"attributeGroup@def": {
		ref: xsd.SpecSrcAttributeGroup,
		attrs: map[string]attrSpec{
			"id": optA(vcID), "name": req(vcNCName),
		},
		content: seq(annQ, attrDecls),
		children: map[string]string{
			"attribute": "attribute@local", "attributeGroup": "attributeGroup@ref",
		},
	},
	"attributeGroup@ref": {
		ref: xsd.SpecSrcAttributeGroup,
		attrs: map[string]attrSpec{
			"id": optA(vcID), "ref": req(vcQName),
		},
		content: annOnly,
	},

	"all":          modelGroupSpec(vcZeroOne, vcZeroOne, allContent),
	"choice":       modelGroupSpec(vcMinOccurs, vcMaxOccurs, seqChoiceContent),
	"sequence":     modelGroupSpec(vcMinOccurs, vcMaxOccurs, seqChoiceContent),
	"all@def":      modelGroupSpec(nil, nil, allContent),
	"choice@def":   modelGroupSpec(nil, nil, seqChoiceContent),
	"sequence@def": modelGroupSpec(nil, nil, seqChoiceContent),

	"any": {
		ref: xsd.SpecSrcWildcard,
		attrs: map[string]attrSpec{
			"id": optA(vcID), "maxOccurs": optA(vcMaxOccurs), "minOccurs": optA(vcMinOccurs),
			"namespace": optA(vcWildcardNS), "notNamespace": optA(vcNotNamespace),
			"notQName":        optA(vcNotQName(true)),
			"processContents": optA(vcEnum("lax", "skip", "strict")),
		},
		content: annOnly,
		extra: func(w *walker, n *xmltree.Node) {
			wildcardExtra(w, n)
			w.checkOccurs(n)
		},
	},
	"anyAttribute": {
		ref: xsd.SpecSrcWildcard,
		attrs: map[string]attrSpec{
			"id":        optA(vcID),
			"namespace": optA(vcWildcardNS), "notNamespace": optA(vcNotNamespace),
			"notQName":        optA(vcNotQName(false)),
			"processContents": optA(vcEnum("lax", "skip", "strict")),
		},
		content: annOnly,
		extra:   wildcardExtra,
	},

	// --- identity constraints ----------------------------------------

	"unique": icSpec(false),
	"key":    icSpec(false),
	"keyref": icSpec(true),
	"selector": {
		ref: xsd.SpecSrcIdentity,
		attrs: map[string]attrSpec{
			"id": optA(vcID), "xpath": req(vcXPath), "xpathDefaultNamespace": optA(vcAny),
		},
		content: annOnly,
	},
	"field": {
		ref: xsd.SpecSrcIdentity,
		attrs: map[string]attrSpec{
			"id": optA(vcID), "xpath": req(vcXPath), "xpathDefaultNamespace": optA(vcAny),
		},
		content: annOnly,
	},

	"notation": {
		ref: xsd.SpecNPropsCorrect,
		attrs: map[string]attrSpec{
			"id": optA(vcID), "name": req(vcNCName),
			"public": optA(vcAny), "system": optA(vcAny),
		},
		content: annOnly,
		extra: func(w *walker, n *xmltree.Node) {
			// spec: n-props-correct — XSD 1.1 Part 1 §3.14.6: at least one of
			// public or system identifier.
			_, hasPub := n.Attr("public")
			_, hasSys := n.Attr("system")
			if !hasPub && !hasSys {
				w.errf(xsd.SpecNPropsCorrect, n.Pos, "notation requires a public or system identifier")
			}
		},
	},

	// --- facets (Part 2 §4.3) ----------------------------------------

	"length":           facetSpec(req(vcNonNegInt), true),
	"minLength":        facetSpec(req(vcNonNegInt), true),
	"maxLength":        facetSpec(req(vcNonNegInt), true),
	"totalDigits":      facetSpec(req(vcPosInt), true),
	"fractionDigits":   facetSpec(req(vcNonNegInt), true),
	"whiteSpace":       facetSpec(req(vcEnum("preserve", "replace", "collapse")), true),
	"explicitTimezone": facetSpec(req(vcEnum("optional", "required", "prohibited")), true),
	// Value lexicals of bounds/enumerations are parsed with the base type in
	// pass 2; patterns are compiled there too.
	"pattern":      facetSpec(req(vcAny), false),
	"enumeration":  facetSpec(req(vcAny), false),
	"minExclusive": facetSpec(req(vcAny), true),
	"minInclusive": facetSpec(req(vcAny), true),
	"maxExclusive": facetSpec(req(vcAny), true),
	"maxInclusive": facetSpec(req(vcAny), true),
	"assertion": {
		ref: xsd.SpecSrcSimpleType,
		attrs: map[string]attrSpec{
			"id": optA(vcID), "test": req(vcXPath), "xpathDefaultNamespace": optA(vcAny),
		},
		content: annOnly,
	},
}

// ctContent is the complexType content model: simpleContent | complexContent
// | the abbreviated form (an implicit restriction of xs:anyType).
var ctContent = seq(annQ, choice(
	one("simpleContent"),
	one("complexContent"),
	seq(opt(one("openContent")), opt(particle), attrDecls, asserts)))

var elementContent = seq(annQ, opt(names("simpleType", "complexType")),
	star(one("alternative")), star(names("unique", "key", "keyref")))

var allContent = seq(annQ, star(names("element", "any", "group")))
var seqChoiceContent = seq(annQ, star(names("element", "group", "choice", "sequence", "any")))

// modelGroupSpec builds the entry for all/choice/sequence. Occurrence
// checkers are nil for the @def variants, where the schema for schemas
// prohibits minOccurs/maxOccurs (the model group of a named group
// definition always occurs exactly once).
func modelGroupSpec(minC, maxC valueCheck, content cm) *elemSpec {
	attrs := map[string]attrSpec{"id": optA(vcID)}
	var extra func(w *walker, n *xmltree.Node)
	if minC != nil {
		attrs["minOccurs"] = optA(minC)
		attrs["maxOccurs"] = optA(maxC)
		extra = func(w *walker, n *xmltree.Node) { w.checkOccurs(n) }
	}
	return &elemSpec{
		ref:      xsd.SpecSrcModelGroup,
		attrs:    attrs,
		content:  content,
		children: particleChildren,
		extra:    extra,
	}
}

// icSpec builds the entry for unique/key/keyref.
func icSpec(isKeyref bool) *elemSpec {
	attrs := map[string]attrSpec{
		"id": optA(vcID), "name": optA(vcNCName), "ref": optA(vcQName),
	}
	if isKeyref {
		attrs["refer"] = optA(vcQName)
	}
	return &elemSpec{
		ref:     xsd.SpecSrcIdentity,
		attrs:   attrs,
		content: seq(annQ, opt(seq(one("selector"), plus(one("field"))))),
		extra: func(w *walker, n *xmltree.Node) {
			// spec: src-identity-constraint — XSD 1.1 Part 1 §3.11.3
			_, hasName := n.Attr("name")
			_, hasRef := n.Attr("ref")
			_, hasRefer := n.Attr("refer")
			if hasName == hasRef { // clause 1
				w.errf(xsd.SpecSrcIdentity, n.Pos, "identity constraint requires either name or ref, not %s", bothOrNeither(hasName))
			}
			if hasName && countChildren(n, "selector") == 0 { // clause 2
				w.errf(xsd.SpecSrcIdentity, n.Pos, "named identity constraint requires a <selector> child")
			}
			if isKeyref && hasName && !hasRefer { // clause 3
				w.errf(xsd.SpecSrcIdentity, n.Pos, "named keyref requires a refer attribute")
			}
			if hasRef { // clause 4
				if hasRefer {
					w.errf(xsd.SpecSrcIdentity, n.Pos, "identity constraint with ref must not have refer")
				}
				if countChildren(n, "selector") > 0 || countChildren(n, "field") > 0 {
					w.errf(xsd.SpecSrcIdentity, n.Pos, "identity constraint with ref must not have selector/field children")
				}
			}
		},
	}
}

// facetSpec builds the entry for a constraining facet element.
func facetSpec(value attrSpec, hasFixed bool) *elemSpec {
	attrs := map[string]attrSpec{"id": optA(vcID), "value": value}
	if hasFixed {
		attrs["fixed"] = optA(vcBool)
	}
	return &elemSpec{ref: xsd.SpecSrcSimpleType, attrs: attrs, content: annOnly}
}

func elementCommonExtra(w *walker, n *xmltree.Node) {
	// spec: src-element.1 — XSD 1.1 Part 1 §3.3.3 (src-element)
	if _, hasDefault := n.Attr("default"); hasDefault {
		if _, hasFixed := n.Attr("fixed"); hasFixed {
			w.errf(xsd.SpecSrcElement, n.Pos, "element default and fixed must not both be present")
		}
	}
	// spec: src-element.3 — a type attribute excludes an inline type child.
	if _, hasType := n.Attr("type"); hasType {
		if countChildren(n, "simpleType")+countChildren(n, "complexType") > 0 {
			w.errf(xsd.SpecSrcElement, n.Pos, "element type attribute and an inline type definition must not both be present")
		}
	}
	// spec: src-element.4.1 — targetNamespace requires name.
	if _, hasTNS := n.Attr("targetNamespace"); hasTNS {
		if _, hasName := n.Attr("name"); !hasName {
			w.errf(xsd.SpecSrcElement, n.Pos, "element targetNamespace requires name")
		}
	}
	// spec: src-element.5 — every alternative but the last has a test.
	alts := xsdChildren(n, "alternative")
	for i, alt := range alts {
		if _, hasTest := alt.Attr("test"); !hasTest && i != len(alts)-1 {
			w.errf(xsd.SpecSrcElement, alt.Pos, "alternative without test must be the last alternative")
		}
	}
}

func attributeCommonExtra(w *walker, n *xmltree.Node) {
	// spec: src-attribute.1 — XSD 1.1 Part 1 §3.2.3 (src-attribute)
	if _, hasDefault := n.Attr("default"); hasDefault {
		if _, hasFixed := n.Attr("fixed"); hasFixed {
			w.errf(xsd.SpecSrcAttribute, n.Pos, "attribute default and fixed must not both be present")
		}
	}
	// spec: src-attribute.4 — a type attribute excludes an inline simpleType.
	if _, hasType := n.Attr("type"); hasType {
		if countChildren(n, "simpleType") > 0 {
			w.errf(xsd.SpecSrcAttribute, n.Pos, "attribute type attribute and an inline <simpleType> must not both be present")
		}
	}
	// spec: src-attribute.5 — fixed excludes use="prohibited".
	if _, hasFixed := n.Attr("fixed"); hasFixed {
		if use, ok := n.Attr("use"); ok && strings.TrimSpace(use) == "prohibited" {
			w.errf(xsd.SpecSrcAttribute, n.Pos, `attribute with fixed must not have use="prohibited"`)
		}
	}
	// spec: src-attribute.6.1 — targetNamespace requires name.
	if _, hasTNS := n.Attr("targetNamespace"); hasTNS {
		if _, hasName := n.Attr("name"); !hasName {
			w.errf(xsd.SpecSrcAttribute, n.Pos, "attribute targetNamespace requires name")
		}
	}
	// spec: no-xmlns — XSD 1.1 Part 1 §3.2.6.3 (no-xmlns)
	if name, ok := n.Attr("name"); ok && strings.TrimSpace(name) == "xmlns" {
		w.errf(xsd.SpecNoXmlns, n.Pos, `attribute declarations must not be named "xmlns"`)
	}
	// spec: no-xsi — XSD 1.1 Part 1 §3.2.6.4 (no-xsi)
	if tns, ok := n.Attr("targetNamespace"); ok && tns == xsd.XSINS {
		w.errf(xsd.SpecNoXsi, n.Pos, "attribute declarations must not target the XML Schema instance namespace")
	}
}

func wildcardExtra(w *walker, n *xmltree.Node) {
	// spec: src-wildcard — XSD 1.1 Part 1 §3.10.3: namespace and
	// notNamespace must not both be present.
	_, hasNS := n.Attr("namespace")
	_, hasNotNS := n.Attr("notNamespace")
	if hasNS && hasNotNS {
		w.errf(xsd.SpecSrcWildcard, n.Pos, "wildcard namespace and notNamespace must not both be present")
	}
}

// forbidWithRef reports every attribute and child that must be absent when a
// declaration is a reference.
func forbidWithRef(w *walker, n *xmltree.Node, ref xsd.SpecRef, what string, attrs, children []string) {
	for _, a := range attrs {
		if _, ok := n.Attr(a); ok {
			w.errf(ref, n.Pos, "%s with ref must not have %s", what, a)
		}
	}
	for _, c := range children {
		if countChildren(n, c) > 0 {
			w.errf(ref, n.Pos, "%s with ref must not have <%s> children", what, c)
		}
	}
}

func countChildren(n *xmltree.Node, local string) int {
	count := 0
	for _, c := range n.Children {
		if c.Name.Space == xsd.XSDNS && c.Name.Local == local {
			count++
		}
	}
	return count
}

func xsdChildren(n *xmltree.Node, local string) []*xmltree.Node {
	var out []*xmltree.Node
	for _, c := range n.Children {
		if c.Name.Space == xsd.XSDNS && c.Name.Local == local {
			out = append(out, c)
		}
	}
	return out
}

func bothOrNeither(both bool) string {
	if both {
		return "both"
	}
	return "neither"
}

// mergeChildren unions two child-variant maps (later wins, never conflicts
// in practice).
func mergeChildren(ms ...map[string]string) map[string]string {
	out := map[string]string{}
	for _, m := range ms {
		maps.Copy(out, m)
	}
	return out
}
