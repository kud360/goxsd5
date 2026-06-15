package xsd

// SpecRef identifies the spec clause a validation enforces. ID is the stable
// constraint identifier (src-…, cos-…, cvc-…, *-props-correct, facet names);
// Section is the human-readable section number at time of writing (sections
// drift; IDs do not); Anchor points into docs/clean/xmlschema11-<Part>.md.
type SpecRef struct {
	ID      string
	Part    int // 1 = Structures, 2 = Datatypes
	Section string
	Anchor  string
}

func (r SpecRef) IsZero() bool { return r.ID == "" }

// Refs is the registry of every declared SpecRef, keyed by ID. The
// conformance test walks this table and cross-checks CONFORMANCE.md.
var Refs = map[string]SpecRef{}

// ref panics on a duplicate ID: it runs only at package init on the var
// block below, so a duplicate can only be a bad edit to this file.
func ref(part int, id, section, anchor string) SpecRef {
	if _, dup := Refs[id]; dup {
		panic("xsd: duplicate SpecRef ID " + id)
	}
	r := SpecRef{ID: id, Part: part, Section: section, Anchor: anchor}
	Refs[id] = r
	return r
}

// Part 1 — Structures: schema representation constraints (src-*).
var (
	SpecSrcQName          = ref(1, "src-qname", "3.15.3", "src-qname")
	SpecSrcElement        = ref(1, "src-element", "3.3.3", "src-element")
	SpecSrcAttribute      = ref(1, "src-attribute", "3.2.3", "src-attribute")
	SpecSrcCT             = ref(1, "src-ct", "3.4.3", "src-ct")
	SpecSrcSimpleType     = ref(1, "src-simple-type", "3.16.3", "src-simple-type")
	SpecSrcImport         = ref(1, "src-import", "4.2.6.2", "src-import")
	SpecSrcInclude        = ref(1, "src-include", "4.2.3", "src-include")
	SpecSrcRedefine       = ref(1, "src-redefine", "4.2.5", "src-redefine")
	SpecSrcOverride       = ref(1, "src-override", "4.2.4", "src-override")
	SpecSrcResolve        = ref(1, "src-resolve", "3.15.3", "src-resolve")
	SpecSrcList           = ref(1, "src-list-itemType-or-simpleType", "3.16.3", "src-list-itemType-or-simpleType")
	SpecSrcUnion          = ref(1, "src-union-memberTypes-or-simpleTypes", "3.16.3", "src-union-memberTypes-or-simpleTypes")
	SpecSrcRestriction    = ref(1, "src-restriction-base-or-simpleType", "3.16.3", "src-restriction-base-or-simpleType")
	SpecSrcAttributeGroup = ref(1, "src-attribute_group", "3.6.3", "src-attribute_group")
	SpecSrcModelGroup     = ref(1, "src-model_group_defn", "3.7.3", "src-model_group_defn")
	SpecSrcIdentity       = ref(1, "src-identity-constraint", "3.11.3", "src-identity-constraint")
	SpecSrcExpredef       = ref(1, "src-expredef", "4.2.5", "src-expredef")
	SpecSrcSchema         = ref(1, "src-schema", "3.17.3", "src-schema")
	SpecSrcAnnotation     = ref(1, "src-annotation", "3.15.3", "src-annotation")
	SpecSrcWildcard       = ref(1, "src-wildcard", "3.10.3", "src-wildcard")
	SpecSrcDupID          = ref(1, "src-id", "3.17.3", "src-id") // unique xs:ID values within a schema document
	SpecSrcTA             = ref(1, "src-ta", "3.12.3", "src-ta") // type alternative representation OK
	SpecCIP               = ref(1, "cip", "4.2.2", "cip")        // conditional inclusion (vc:* attribute validity)
)

// Part 1 — Structures: component constraints (*-props-correct, cos-*, …).
var (
	SpecSTPropsCorrect          = ref(1, "st-props-correct", "3.16.6", "st-props-correct")
	SpecCTPropsCorrect          = ref(1, "ct-props-correct", "3.4.6", "ct-props-correct")
	SpecEPropsCorrect           = ref(1, "e-props-correct", "3.3.6", "e-props-correct")
	SpecAPropsCorrect           = ref(1, "a-props-correct", "3.2.6", "a-props-correct")
	SpecAUPropsCorrect          = ref(1, "au-props-correct", "3.5.6", "au-props-correct")
	SpecAGPropsCorrect          = ref(1, "ag-props-correct", "3.6.6", "ag-props-correct")
	SpecMGPropsCorrect          = ref(1, "mg-props-correct", "3.8.6", "mg-props-correct")
	SpecMGDPropsCorrect         = ref(1, "mgd-props-correct", "3.7.6", "mgd-props-correct")
	SpecPPropsCorrect           = ref(1, "p-props-correct", "3.9.6", "p-props-correct")
	SpecWPropsCorrect           = ref(1, "w-props-correct", "3.10.6", "w-props-correct")
	SpecNPropsCorrect           = ref(1, "n-props-correct", "3.14.6", "n-props-correct")
	SpecCosCTExtends            = ref(1, "cos-ct-extends", "3.4.6", "cos-ct-extends")
	SpecCosCTRestricts          = ref(1, "cos-ct-restricts", "3.4.6", "cos-ct-restricts")
	SpecDerivationOKRestriction = ref(1, "derivation-ok-restriction", "3.4.6", "derivation-ok-restriction")
	SpecCosSTDerivedOK          = ref(1, "cos-st-derived-ok", "3.16.6", "cos-st-derived-ok")
	SpecCosCTDerivedOK          = ref(1, "cos-ct-derived-ok", "3.4.6", "cos-ct-derived-ok")
	SpecCosEquivClass           = ref(1, "cos-equiv-class", "3.3.6", "cos-equiv-class")
	SpecCosParticleRestrict     = ref(1, "cos-particle-restrict", "3.9.6", "cos-particle-restrict")
	SpecCosNonambig             = ref(1, "cos-nonambig", "3.8.6", "cos-nonambig")                     // UPA
	SpecCosElementConsistent    = ref(1, "cos-element-consistent", "3.8.6", "cos-element-consistent") // EDC
	SpecCosAllLimited           = ref(1, "cos-all-limited", "3.8.6", "cos-all-limited")
	SpecCosNSSubset             = ref(1, "cos-ns-subset", "3.10.6", "cos-ns-subset")
	SpecCosAWIntersect          = ref(1, "cos-aw-intersect", "3.10.6", "cos-aw-intersect")
	SpecCosAWUnion              = ref(1, "cos-aw-union", "3.10.6", "cos-aw-union")
	SpecCosApplicableFacets     = ref(1, "cos-applicable-facets", "4.1.6", "cos-applicable-facets")
	SpecCosValidDefault         = ref(1, "cos-valid-default", "3.3.6", "cos-valid-default")
	SpecSchemaPropsCorrect      = ref(1, "sch-props-correct", "3.17.6", "sch-props-correct")
	SpecICPropsCorrect          = ref(1, "c-props-correct", "3.11.6", "c-props-correct")
	SpecNoXmlns                 = ref(1, "no-xmlns", "3.2.6.3", "no-xmlns")
	SpecNoXsi                   = ref(1, "no-xsi", "3.2.6.4", "no-xsi")
)

// Part 1 — Structures: instance validation rules (cvc-*). These govern the
// schema-validity assessment of an instance against the component model and are
// enforced by the xsdvalidate engine, not the schema processor.
var (
	SpecCvcElt          = ref(1, "cvc-elt", "3.3.4", "cvc-elt")
	SpecCvcType         = ref(1, "cvc-type", "3.4.4", "cvc-type")
	SpecCvcComplexType  = ref(1, "cvc-complex-type", "3.4.4", "cvc-complex-type")
	SpecCvcAttribute    = ref(1, "cvc-attribute", "3.2.4", "cvc-attribute")
	SpecCvcAU           = ref(1, "cvc-au", "3.5.4", "cvc-au")
	SpecCvcParticle     = ref(1, "cvc-particle", "3.9.4", "cvc-particle")
	SpecCvcWildcard     = ref(1, "cvc-wildcard", "3.10.4", "cvc-wildcard")
	SpecCvcID           = ref(1, "cvc-id", "3.4.4", "cvc-id")
	SpecCvcIdentity     = ref(1, "cvc-identity-constraint", "3.11.4", "cvc-identity-constraint")
	SpecCvcAssertion    = ref(1, "cvc-assertion", "3.13.4", "cvc-assertion")
)

// Part 2 — Datatypes: facet validation rules (cvc-*).
var (
	SpecWhiteSpaceValid       = ref(2, "cvc-whiteSpace", "4.3.6", "whiteSpace")
	SpecPatternValid          = ref(2, "cvc-pattern-valid", "4.3.4.4", "cvc-pattern-valid")
	SpecEnumerationValid      = ref(2, "cvc-enumeration-valid", "4.3.5.4", "cvc-enumeration-valid")
	SpecLengthValid           = ref(2, "cvc-length-valid", "4.3.1.4", "cvc-length-valid")
	SpecMinLengthValid        = ref(2, "cvc-minLength-valid", "4.3.2.4", "cvc-minLength-valid")
	SpecMaxLengthValid        = ref(2, "cvc-maxLength-valid", "4.3.3.4", "cvc-maxLength-valid")
	SpecMinInclusiveValid     = ref(2, "cvc-minInclusive-valid", "4.3.10.4", "cvc-minInclusive-valid")
	SpecMaxInclusiveValid     = ref(2, "cvc-maxInclusive-valid", "4.3.7.4", "cvc-maxInclusive-valid")
	SpecMinExclusiveValid     = ref(2, "cvc-minExclusive-valid", "4.3.9.4", "cvc-minExclusive-valid")
	SpecMaxExclusiveValid     = ref(2, "cvc-maxExclusive-valid", "4.3.8.4", "cvc-maxExclusive-valid")
	SpecTotalDigitsValid      = ref(2, "cvc-totalDigits-valid", "4.3.11.4", "cvc-totalDigits-valid")
	SpecFractionDigitsValid   = ref(2, "cvc-fractionDigits-valid", "4.3.12.4", "cvc-fractionDigits-valid")
	SpecAssertionsValid       = ref(2, "cvc-assertions-valid", "4.3.13.4", "cvc-assertions-valid")
	SpecExplicitTimezoneValid = ref(2, "cvc-explicitTimezone-valid", "4.3.14.4", "cvc-explicitTimezone-valid")
	SpecDatatypeValid         = ref(2, "cvc-datatype-valid", "4.1.4", "cvc-datatype-valid")
)

// Part 2 — Datatypes: constraints on facet schema components (intra-facet).
var (
	SpecLengthMinMax                   = ref(2, "length-minLength-maxLength", "4.3.1.5", "length-minLength-maxLength")
	SpecMinLELMaxLength                = ref(2, "minLength-less-than-equal-to-maxLength", "4.3.2.5", "minLength-less-than-equal-to-maxLength")
	SpecLengthValidRestriction         = ref(2, "length-valid-restriction", "4.3.1.5", "length-valid-restriction")
	SpecMinLengthValidRestriction      = ref(2, "minLength-valid-restriction", "4.3.2.5", "minLength-valid-restriction")
	SpecMaxLengthValidRestriction      = ref(2, "maxLength-valid-restriction", "4.3.3.5", "maxLength-valid-restriction")
	SpecWhiteSpaceValidRestriction     = ref(2, "whiteSpace-valid-restriction", "4.3.6.5", "whiteSpace-valid-restriction")
	SpecMinInclLEMaxIncl               = ref(2, "minInclusive-less-than-equal-to-maxInclusive", "4.3.10.5", "minInclusive-less-than-equal-to-maxInclusive")
	SpecMinExclLEMaxExcl               = ref(2, "minExclusive-less-than-equal-to-maxExclusive", "4.3.9.5", "minExclusive-less-than-equal-to-maxExclusive")
	SpecMinInclExclusive               = ref(2, "minInclusive-minExclusive", "4.3.10.5", "minInclusive-minExclusive")
	SpecMaxInclExclusive               = ref(2, "maxInclusive-maxExclusive", "4.3.7.5", "maxInclusive-maxExclusive")
	SpecMinExclLTMaxIncl               = ref(2, "minExclusive-less-than-maxInclusive", "4.3.9.5", "minExclusive-less-than-maxInclusive")
	SpecMinInclLTMaxExcl               = ref(2, "minInclusive-less-than-maxExclusive", "4.3.10.5", "minInclusive-less-than-maxExclusive")
	SpecMaxInclValidRestriction        = ref(2, "maxInclusive-valid-restriction", "4.3.7.5", "maxInclusive-valid-restriction")
	SpecMaxExclValidRestriction        = ref(2, "maxExclusive-valid-restriction", "4.3.8.5", "maxExclusive-valid-restriction")
	SpecMinExclValidRestriction        = ref(2, "minExclusive-valid-restriction", "4.3.9.5", "minExclusive-valid-restriction")
	SpecMinInclValidRestriction        = ref(2, "minInclusive-valid-restriction", "4.3.10.5", "minInclusive-valid-restriction")
	SpecTotalDigitsValidRestriction    = ref(2, "totalDigits-valid-restriction", "4.3.11.5", "totalDigits-valid-restriction")
	SpecFracLETotalDigits              = ref(2, "fractionDigits-less-than-equal-to-totalDigits", "4.3.12.5", "fractionDigits-less-than-equal-to-totalDigits")
	SpecFractionDigitsValidRestriction = ref(2, "fractionDigits-valid-restriction", "4.3.12.5", "fractionDigits-valid-restriction")
	SpecEnumerationValidRestriction    = ref(2, "enumeration-valid-restriction", "4.3.5.5", "enumeration-valid-restriction")
	SpecFixedFacetValue                = ref(2, "fixed-facet-value", "4.3", "fixed-facet-value")
	SpecETZValidRestriction            = ref(2, "explicitTimezone-valid-restriction", "4.3.16.5", "explicitTimezone-valid-restriction")
)

// Part 2 — Datatypes: derivation and special types.
var (
	SpecCosSTRestricts    = ref(2, "cos-st-restricts", "4.1.6", "cos-st-restricts")
	SpecSTRestrictsFacets = ref(2, "st-restrict-facets", "3.16.6", "st-restrict-facets")
	SpecQNameSpecial      = ref(2, "qname-special", "3.3.18", "QName")
	SpecNotationSpecial   = ref(2, "notation-special", "3.3.19", "NOTATION")
	SpecEnumNotation      = ref(2, "enumeration-required-notation", "3.3.19", "enumeration-required-notation")
	SpecRegexValid        = ref(2, "regex-valid", "G", "regexs") // Appendix: pattern must be a valid regular expression
	SpecSingleFacetValue  = ref(2, "src-single-facet-value", "4.3", "src-single-facet-value")
	SpecFundamentalFacets = ref(2, "fundamental-facets", "F", "fundamental-facets")
)
