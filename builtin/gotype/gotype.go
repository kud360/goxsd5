package gotype

import (
	"fmt"

	"github.com/kud360/goxsd5/xsd"
	"github.com/kud360/goxsd5/xsdedit"
)

func qn(local string) xsd.QName { return xsd.QName{Namespace: xsd.XSDNS, Local: local} }

// Applicable-facet sets per primitive (cos-applicable-facets, Part 2 §4.1.6),
// mirroring builtin's authored sets so a lax type admits exactly the facets its
// strict sibling does.
const (
	facetsStringy = xsd.FacetsCommon | xsd.FacetsLength
	facetsNumeric = xsd.FacetsCommon | xsd.FacetBounds
	facetsDecimal = facetsNumeric | xsd.FacetTotalDigits | xsd.FacetFractionDigits
	// precisionDecimal admits totalDigits + maxScale + minScale but not
	// fractionDigits/length, mirroring builtin.facetsPrecisionDecimal. The lax
	// number value has no scale, so the scale facets simply do not constrain.
	facetsPrecisionDecimal = facetsNumeric | xsd.FacetTotalDigits | xsd.FacetMaxScale | xsd.FacetMinScale
	facetsDateTimey        = facetsNumeric | xsd.FacetExplicitTimezone
	facetsBoolean          = xsd.FacetPattern | xsd.FacetWhiteSpace | xsd.FacetAssertion
)

// mustPrimitive builds a lax primitive through xsdedit.NewPrimitive, panicking
// on the impossible error: the arguments are package-init compile-time literals,
// so a failure can only mean a wrong literal here, never a runtime condition.
func mustPrimitive(local string, base *xsd.SimpleType, parse xsd.ParseFunc, compare xsd.CompareFunc, applic xsd.FacetSet, ws xsd.WhiteSpace) *xsd.SimpleType {
	t, err := xsdedit.NewPrimitive(qn(local), base, parse, compare, applic, ws)
	if err != nil {
		panic(fmt.Sprintf("gotype primitive %q: %v", local, err))
	}
	return t
}

// restrict builds a lax derivation step. Like builtin's helper it is a plain
// struct builder (not RestrictWith): the built-in ladder's narrowing is already
// proven by the strict suite, and re-validating lax bounds whose value space
// saturates (unsignedLong beyond int64) would reject sound derivations. The
// parser validates each derived type lazily as usual. parse, when non-nil,
// overrides the inherited lexical mapping (xs:integer re-skins decimal onto
// int64); compare does the same for the value comparator.
func restrict(local string, base *xsd.SimpleType, parse xsd.ParseFunc, compare xsd.CompareFunc, mod func(f *xsd.Facets)) *xsd.SimpleType {
	var declared xsd.Facets
	if mod != nil {
		mod(&declared)
	}
	return &xsd.SimpleType{
		Name:           qn(local),
		BaseType:       base,
		Variety:        base.Variety,
		ItemType:       base.ItemType,
		Parse:          parse,
		Compare:        compare,
		DeclaredFacets: declared,
	}
}

// list builds a lax list type (whiteSpace=collapse fixed, supplied by the list
// variety), mirroring builtin's helper.
func list(local string, item *xsd.SimpleType, mod func(f *xsd.Facets)) *xsd.SimpleType {
	t := &xsd.SimpleType{
		Name:     qn(local),
		BaseType: anySimpleType,
		Variety:  xsd.VarietyList,
		ItemType: item,
	}
	if mod != nil {
		mod(&t.DeclaredFacets)
	}
	return t
}

func intFacet(v int) *xsd.IntFacet { return &xsd.IntFacet{Value: v} }

// bound builds an order bound from a lexical, parsing it with the lax integer
// parser (the only ladder that carries explicit numeric bounds). A bad literal
// panics: package-init compile-time inputs.
func bound(s string) *xsd.Bound {
	v, err := parseInteger(s, nil)
	if err != nil {
		panic(fmt.Sprintf("gotype bound %q: %v", s, err))
	}
	return &xsd.Bound{Value: v, Lexical: s}
}

// ---- ur-types ----

// anyType is the root of the type hierarchy. gotype cannot import builtin (leaf
// purity), so it carries its own xs:anyType for anySimpleType to root on; the
// parser always seeds the registry's xs:anyType from the shared builtin.AnyType
// regardless, and the spec's "derived from anyType" relation is by expanded
// name, so this stand-in is never observed as a distinct component.
var anyType = &xsd.ComplexType{Name: qn("anyType")}

// anySimpleType carries the lax identity string parser and the lax default
// comparator, so every lax type roots here and resolves both through the chain
// (the xsdtype/builtin arrangement, mirrored). Its base is xs:anyType.
var anySimpleType = &xsd.SimpleType{
	Name:     qn("anySimpleType"),
	BaseType: anyType,
	Variety:  xsd.VarietyAtomic,
	Parse:    parseString,
	Compare:  compareValues,
}

var anyAtomicType = &xsd.SimpleType{
	Name:     qn("anyAtomicType"),
	BaseType: anySimpleType,
	Variety:  xsd.VarietyAtomic,
}

// ---- primitives ----

var (
	String  = mustPrimitive("string", anyAtomicType, parseString, nil, facetsStringy, xsd.WSPreserve)
	Boolean = mustPrimitive("boolean", anyAtomicType, parseBoolean, nil, facetsBoolean, xsd.WSCollapse)
	Decimal = mustPrimitive("decimal", anyAtomicType, parseNumber, nil, facetsDecimal, xsd.WSCollapse)
	Float   = mustPrimitive("float", anyAtomicType, parseNumber, nil, facetsNumeric, xsd.WSCollapse)
	Double  = mustPrimitive("double", anyAtomicType, parseNumber, nil, facetsNumeric, xsd.WSCollapse)
	// PrecisionDecimal: the lax number value carries no scale, so maxScale/minScale
	// are admitted but inert; the type exists to mirror the strict set member-for-member.
	PrecisionDecimal = mustPrimitive("precisionDecimal", anyAtomicType, parseNumber, nil, facetsPrecisionDecimal, xsd.WSCollapse)

	Duration   = mustPrimitive("duration", anyAtomicType, parseDuration, nil, facetsNumeric, xsd.WSCollapse)
	DateTime   = mustPrimitive("dateTime", anyAtomicType, parseTemporal("dateTime"), compareTime, facetsDateTimey, xsd.WSCollapse)
	Time       = mustPrimitive("time", anyAtomicType, parseTemporal("time"), compareTime, facetsDateTimey, xsd.WSCollapse)
	Date       = mustPrimitive("date", anyAtomicType, parseTemporal("date"), compareTime, facetsDateTimey, xsd.WSCollapse)
	GYearMonth = mustPrimitive("gYearMonth", anyAtomicType, parseTemporal("gYearMonth"), compareTime, facetsDateTimey, xsd.WSCollapse)
	GYear      = mustPrimitive("gYear", anyAtomicType, parseTemporal("gYear"), compareTime, facetsDateTimey, xsd.WSCollapse)
	GMonthDay  = mustPrimitive("gMonthDay", anyAtomicType, parseGMonthDay, compareTime, facetsDateTimey, xsd.WSCollapse)
	GDay       = mustPrimitive("gDay", anyAtomicType, parseGDay, compareTime, facetsDateTimey, xsd.WSCollapse)
	GMonth     = mustPrimitive("gMonth", anyAtomicType, parseGMonth, compareTime, facetsDateTimey, xsd.WSCollapse)

	HexBinary    = mustPrimitive("hexBinary", anyAtomicType, parseHexBinary, nil, facetsStringy, xsd.WSCollapse)
	Base64Binary = mustPrimitive("base64Binary", anyAtomicType, parseBase64Binary, nil, facetsStringy, xsd.WSCollapse)
	AnyURI       = mustPrimitive("anyURI", anyAtomicType, parseString, nil, facetsStringy, xsd.WSCollapse)
	QName        = mustPrimitive("QName", anyAtomicType, parseQName, compareQName, facetsStringy, xsd.WSCollapse)
	NOTATION     = mustPrimitive("NOTATION", anyAtomicType, parseQName, compareQName, facetsStringy, xsd.WSCollapse)
)

// ---- derived numeric ladder ----
//
// xs:integer re-skins xs:decimal's value space onto int64 (its own Parse and
// Compare); every integer derivation inherits that through the chain.

var (
	Integer = restrict("integer", Decimal, parseInteger, compareValues, func(f *xsd.Facets) {
		f.FractionDigits = &xsd.IntFacet{Value: 0, Fixed: true}
	})
	NonPositiveInteger = restrict("nonPositiveInteger", Integer, nil, nil, func(f *xsd.Facets) {
		f.MaxInclusive = bound("0")
	})
	NegativeInteger = restrict("negativeInteger", NonPositiveInteger, nil, nil, func(f *xsd.Facets) {
		f.MaxInclusive = bound("-1")
	})
	Long = restrict("long", Integer, nil, nil, func(f *xsd.Facets) {
		f.MinInclusive = bound("-9223372036854775808")
		f.MaxInclusive = bound("9223372036854775807")
	})
	Int = restrict("int", Long, nil, nil, func(f *xsd.Facets) {
		f.MinInclusive = bound("-2147483648")
		f.MaxInclusive = bound("2147483647")
	})
	Short = restrict("short", Int, nil, nil, func(f *xsd.Facets) {
		f.MinInclusive = bound("-32768")
		f.MaxInclusive = bound("32767")
	})
	Byte = restrict("byte", Short, nil, nil, func(f *xsd.Facets) {
		f.MinInclusive = bound("-128")
		f.MaxInclusive = bound("127")
	})
	NonNegativeInteger = restrict("nonNegativeInteger", Integer, nil, nil, func(f *xsd.Facets) {
		f.MinInclusive = bound("0")
	})
	UnsignedLong = restrict("unsignedLong", NonNegativeInteger, nil, nil, func(f *xsd.Facets) {
		f.MaxInclusive = bound("9223372036854775807") // lax: int64-saturated
	})
	UnsignedInt = restrict("unsignedInt", UnsignedLong, nil, nil, func(f *xsd.Facets) {
		f.MaxInclusive = bound("4294967295")
	})
	UnsignedShort = restrict("unsignedShort", UnsignedInt, nil, nil, func(f *xsd.Facets) {
		f.MaxInclusive = bound("65535")
	})
	UnsignedByte = restrict("unsignedByte", UnsignedShort, nil, nil, func(f *xsd.Facets) {
		f.MaxInclusive = bound("255")
	})
	PositiveInteger = restrict("positiveInteger", NonNegativeInteger, nil, nil, func(f *xsd.Facets) {
		f.MinInclusive = bound("1")
	})
)

// ---- derived string ladder ----

var (
	NormalizedString = restrict("normalizedString", String, nil, nil, func(f *xsd.Facets) {
		f.WhiteSpace = xsd.WSReplace
	})
	Token = restrict("token", NormalizedString, nil, nil, func(f *xsd.Facets) {
		f.WhiteSpace = xsd.WSCollapse
	})
	Language = restrict("language", Token, nil, nil, nil)
	NMTOKEN  = restrict("NMTOKEN", Token, nil, nil, nil)
	Name     = restrict("Name", Token, nil, nil, nil)
	NCName   = restrict("NCName", Name, nil, nil, nil)
	ID       = restrict("ID", NCName, nil, nil, nil)
	IDREF    = restrict("IDREF", NCName, nil, nil, nil)
	ENTITY   = restrict("ENTITY", NCName, nil, nil, nil)
)

// ---- list types ----

var (
	NMTOKENS = list("NMTOKENS", NMTOKEN, func(f *xsd.Facets) { f.MinLength = intFacet(1) })
	IDREFS   = list("IDREFS", IDREF, func(f *xsd.Facets) { f.MinLength = intFacet(1) })
	ENTITIES = list("ENTITIES", ENTITY, func(f *xsd.Facets) { f.MinLength = intFacet(1) })
)

// ---- XSD 1.1 additions ----

var (
	DateTimeStamp = restrict("dateTimeStamp", DateTime, nil, nil, func(f *xsd.Facets) {
		f.ExplicitTimezone = xsd.ETZRequired
		f.ExplicitTimezoneFixed = true
	})
	YearMonthDuration = restrict("yearMonthDuration", Duration, nil, nil, nil)
	DayTimeDuration   = restrict("dayTimeDuration", Duration, nil, nil, nil)
)

// AllBuiltins returns the full lax simple-type set — xs:anySimpleType and
// xs:anyAtomicType, the 19 lax primitives, and their derived/list ladder — for
// seeding a parser registry via parser.Options.Primitives. It mirrors
// builtin.AllBuiltins() one-for-one in membership, differing only in value
// semantics. xs:anyType and xs:error stay structural and are seeded from the
// shared builtin set by the parser regardless.
func AllBuiltins() []*xsd.SimpleType {
	return []*xsd.SimpleType{
		anySimpleType, anyAtomicType,
		String, Boolean, Decimal, Float, Double, PrecisionDecimal,
		Duration, DateTime, Time, Date, GYearMonth, GYear, GMonthDay, GDay, GMonth,
		HexBinary, Base64Binary, AnyURI, QName, NOTATION,
		Integer, NonPositiveInteger, NegativeInteger, Long, Int, Short, Byte,
		NonNegativeInteger, UnsignedLong, UnsignedInt, UnsignedShort, UnsignedByte, PositiveInteger,
		NormalizedString, Token, Language, NMTOKEN, Name, NCName, ID, IDREF, ENTITY,
		NMTOKENS, IDREFS, ENTITIES,
		DateTimeStamp, YearMonthDuration, DayTimeDuration,
	}
}

// Replace returns a copy of base with each type whose expanded name matches one
// in lax replaced by the lax type, leaving every other entry strict. It is the
// way to mix strict and lax value spaces within one registry seed — e.g. opt
// into Go-native doubles while keeping every other type strict:
//
//	prims := gotype.Replace(builtin.AllBuiltins(), gotype.Double, gotype.Float)
//	schemas, err := parser.Parse("schema.xsd", &parser.Options{Primitives: prims})
//
// A lax type whose name is absent from base is appended. base is not mutated.
func Replace(base []*xsd.SimpleType, lax ...*xsd.SimpleType) []*xsd.SimpleType {
	byName := make(map[xsd.QName]*xsd.SimpleType, len(lax))
	for _, t := range lax {
		byName[t.Name] = t
	}
	out := make([]*xsd.SimpleType, 0, len(base)+len(lax))
	seen := make(map[xsd.QName]bool, len(lax))
	for _, t := range base {
		if r, ok := byName[t.Name]; ok {
			out = append(out, r)
			seen[t.Name] = true
			continue
		}
		out = append(out, t)
	}
	for _, t := range lax {
		if !seen[t.Name] {
			out = append(out, t)
		}
	}
	return out
}
