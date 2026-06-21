// Package builtin defines every XSD 1.1 built-in datatype as a package-
// level *xsd.SimpleType, expressed in xsd terms: constraining facets plus
// ParseFunc/CompareFunc, hand-derived from the datatypes hierarchy of
// Part 2 §3.3–3.4 with the HFP fundamental facets of §F. Go's in-package
// var initialization order lets each derived builtin reference its base
// directly.
package builtin

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/kud360/goxsd5/builtin/xsdtype"
	"github.com/kud360/goxsd5/xsd"
	"github.com/kud360/goxsd5/xsdregex"
)

func qn(local string) xsd.QName { return xsd.QName{Namespace: xsd.XSDNS, Local: local} }

// AnyType is xs:anyType, the root of the type hierarchy (a complex type
// with unconstrained content). Its base type is itself, represented as nil.
var AnyType = &xsd.ComplexType{
	Name: qn("anyType"),
	Content: &xsd.ElementContent{
		Mixed: true,
		Particle: &xsd.Particle{
			MinOccurs: 0, MaxOccurs: xsd.UnboundedOccurs,
			Term: &xsd.Wildcard{Mode: xsd.NSConstraintAny, ProcessContents: xsd.ProcessLax},
		},
	},
	AttributeWildcard: &xsd.Wildcard{Mode: xsd.NSConstraintAny, ProcessContents: xsd.ProcessLax},
}

// AnySimpleType is xs:anySimpleType (Part 2 §4.1.6): no constraining
// facets, lexical space = all strings. It carries the identity string parser
// and the default value comparator (xsdtype.CompareValues); since every
// simple type roots here, both resolve through the SimpleType chain for any
// type that does not override them — this is how the core engine reaches the
// built-in value spaces without naming them.
var AnySimpleType = &xsd.SimpleType{
	Name:     qn("anySimpleType"),
	BaseType: AnyType,
	Variety:  xsd.VarietyAtomic,
	Parse:    parseAsString,
	Compare:  xsdtype.CompareValues,
}

// AnyAtomicType is xs:anyAtomicType (XSD 1.1).
var AnyAtomicType = &xsd.SimpleType{
	Name:     qn("anyAtomicType"),
	BaseType: AnySimpleType,
	Variety:  xsd.VarietyAtomic,
}

// ErrorType is xs:error (XSD 1.1 §3.16.7.3): a special simple type with empty
// value and lexical space — no lexical form is ever valid. It is a union with
// no member types and {final} = #all, present in every schema so it can be
// named (notably in conditional type assignment) to force invalidity.
var ErrorType = &xsd.SimpleType{
	Name:     qn("error"),
	BaseType: AnySimpleType,
	Variety:  xsd.VarietyUnion,
	Final:    xsd.AllDerivations,
	Parse: func(lexical string, _ xsd.ValueContext) (xsd.Value, error) {
		return nil, fmt.Errorf("no value is valid against xs:error")
	},
}

func parseAsString(s string, _ xsd.ValueContext) (xsd.Value, error) { return xsdtype.String(s), nil }

// primitive constructs a primitive type: based on anyAtomicType, its own
// primitive ancestor. applic is the primitive's applicable-facet set
// (cos-applicable-facets); it is the authored source of truth that
// xsd.SimpleType.ApplicableFacets reads for this primitive and everything
// derived from it. fund carries the authored fundamental-facet base case (Part 2
// §F.1) that xsd.SimpleType.Fundamentals reads the same way.
func primitive(local string, ws xsd.WhiteSpace, applic xsd.FacetSet, fund *xsd.Fundamentals, parse xsd.ParseFunc) *xsd.SimpleType {
	t := &xsd.SimpleType{
		Name:            qn(local),
		BaseType:        AnyAtomicType,
		Variety:         xsd.VarietyAtomic,
		Parse:           parse,
		Applicable:      applic,
		FundamentalBase: fund,
	}
	t.DeclaredFacets.WhiteSpace = ws
	t.DeclaredFacets.WhiteSpaceFixed = local != "string" && local != "anyURI"
	return t
}

// restrict constructs a built-in restriction step.
func restrict(local string, base *xsd.SimpleType, mod func(f *xsd.Facets)) *xsd.SimpleType {
	var declared xsd.Facets
	if mod != nil {
		mod(&declared)
	}
	t := &xsd.SimpleType{
		Name:           qn(local),
		BaseType:       base,
		Variety:        base.Variety,
		ItemType:       base.ItemType,
		DeclaredFacets: declared,
	}
	return t
}

// list constructs a built-in list type (whiteSpace collapse, fixed).
func list(local string, item *xsd.SimpleType, mod func(f *xsd.Facets)) *xsd.SimpleType {
	t := &xsd.SimpleType{
		Name:     qn(local),
		BaseType: AnySimpleType,
		Variety:  xsd.VarietyList,
		ItemType: item,
	}
	// whiteSpace=collapse (fixed) is supplied by EffectiveFacets() from the list
	// variety; only the explicitly declared facets are stored here.
	if mod != nil {
		mod(&t.DeclaredFacets)
	}
	return t
}

// pattern panics on a bad source: it runs only at package init on the
// compile-time literals of the built-in type table, where an error cannot
// be returned and can only mean the literal itself is wrong.
func pattern(f *xsd.Facets, sources ...string) {
	var group xsd.PatternGroup
	for _, src := range sources {
		re, err := xsdregex.CompileRegex(src)
		if err != nil {
			panic(fmt.Sprintf("builtin pattern %q: %v", src, err))
		}
		group = append(group, xsd.Pattern{Source: src, Re: re})
	}
	f.PatternGroups = append(f.PatternGroups, group)
}

// mustDecimal panics on a bad literal for the same reason as pattern:
// package-init only, compile-time constant inputs.
func mustDecimal(s string) xsd.Value {
	d, err := xsdtype.ParseDecimal(s)
	if err != nil {
		panic(fmt.Sprintf("builtin decimal bound %q: %v", s, err))
	}
	return d
}

func decBound(s string, fixed bool) *xsd.Bound {
	return &xsd.Bound{Value: mustDecimal(s), Lexical: s, Fixed: fixed}
}

func intFacet(v int, fixed bool) *xsd.IntFacet { return &xsd.IntFacet{Value: v, Fixed: fixed} }

// ---- primitive parse functions ----

func parseBoolean(s string, _ xsd.ValueContext) (xsd.Value, error) {
	switch s {
	case "true", "1":
		return xsdtype.Boolean(true), nil
	case "false", "0":
		return xsdtype.Boolean(false), nil
	}
	return nil, fmt.Errorf("invalid boolean %q", s)
}

func parseDecimalV(s string, _ xsd.ValueContext) (xsd.Value, error) {
	return xsdtype.ParseDecimal(s)
}

func parsePrecisionDecimalV(s string, _ xsd.ValueContext) (xsd.Value, error) {
	return xsdtype.ParsePrecisionDecimal(s)
}

var floatLexical = regexp.MustCompile(`\A[+-]?([0-9]+(\.[0-9]*)?|\.[0-9]+)([eE][+-]?[0-9]+)?\z`)

func parseFloating(s string) (float64, error) {
	switch s {
	case "INF", "+INF": // +INF is XSD 1.1
		return math.Inf(1), nil
	case "-INF":
		return math.Inf(-1), nil
	case "NaN":
		return math.NaN(), nil
	}
	if !floatLexical.MatchString(s) {
		return 0, fmt.Errorf("invalid float/double %q", s)
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		// Out-of-range overflows map to ±Inf per IEEE rounding.
		if ne, ok := err.(*strconv.NumError); ok && ne.Err == strconv.ErrRange {
			return v, nil
		}
		return 0, fmt.Errorf("invalid float/double %q: %w", s, err)
	}
	return v, nil
}

func parseDouble(s string, _ xsd.ValueContext) (xsd.Value, error) {
	v, err := parseFloating(s)
	if err != nil {
		return nil, err
	}
	return xsdtype.Double(v), nil
}

func parseFloat(s string, _ xsd.ValueContext) (xsd.Value, error) {
	v, err := parseFloating(s)
	if err != nil {
		return nil, err
	}
	return xsdtype.Float(float32(v)), nil
}

func parseDurationV(s string, _ xsd.ValueContext) (xsd.Value, error) {
	return xsdtype.ParseDuration(s)
}

func parseHexBinary(s string, _ xsd.ValueContext) (xsd.Value, error) {
	if len(s)%2 != 0 {
		return nil, fmt.Errorf("invalid hexBinary %q: odd length", s)
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid hexBinary %q: %w", s, err)
	}
	return xsdtype.Bytes(b), nil
}

// base64Lexical is the lexical space of Part 2 §3.3.5: RFC 2045 alphabet
// with single embedded spaces allowed and strict padding placement.
var base64Lexical = regexp.MustCompile(`\A((([A-Za-z0-9+/] ?){4})*(([A-Za-z0-9+/] ?){3}[A-Za-z0-9+/]|([A-Za-z0-9+/] ?){2}[AEIMQUYcgkosw048] ?=|[A-Za-z0-9+/] ?[AQgw] ?= ?=))?\z`)

func parseBase64Binary(s string, _ xsd.ValueContext) (xsd.Value, error) {
	if !base64Lexical.MatchString(s) {
		return nil, fmt.Errorf("invalid base64Binary %q", s)
	}
	b, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(s, " ", ""))
	if err != nil {
		return nil, fmt.Errorf("invalid base64Binary %q: %w", s, err)
	}
	return xsdtype.Bytes(b), nil
}

func parseAnyURI(s string, _ xsd.ValueContext) (xsd.Value, error) {
	// Part 2 §3.3.17: the lexical space is intentionally lax (any string
	// can be %-escaped into a legal IRI); no structural validation.
	return xsdtype.String(s), nil
}

func parseQName(s string, ctx xsd.ValueContext) (xsd.Value, error) {
	// spec: qname-special — XSD 1.1 Part 2 §3.3.18 (xmlschema11-2.md#QName)
	prefix, local := "", s
	if i := strings.IndexByte(s, ':'); i >= 0 {
		prefix, local = s[:i], s[i+1:]
		if prefix == "" {
			return nil, fmt.Errorf("invalid QName %q", s)
		}
	}
	if !isNCName(local) || (prefix != "" && !isNCName(prefix)) {
		return nil, fmt.Errorf("invalid QName %q", s)
	}
	if ctx == nil {
		return nil, fmt.Errorf("cannot resolve QName %q: %w", s, xsdtype.ErrNeedContext)
	}
	name, ok := ctx.ResolveQName(prefix, local)
	if !ok {
		return nil, fmt.Errorf("undefined namespace prefix %q in QName %q", prefix, s)
	}
	return xsdtype.QNameValue{Name: name, Lexical: s}, nil
}

var ncNamePattern = regexp.MustCompile(`\A[\p{L}_][\p{L}\p{N}\p{M}_.\x{B7}-]*\z`)

func isNCName(s string) bool {
	// Close approximation of the XML NCName production; the parser's
	// xmltree package holds the exact table, but builtin must not depend
	// on parser packages.
	return ncNamePattern.MatchString(s)
}

// ---- primitives ----

// The fundamental-facet base case of each primitive is authored below as a
// *xsd.Fundamentals (the fund* presets, Part 2 §F.1); SimpleType.Fundamentals()
// reads it off the primitive.
// Applicable-facet sets per primitive (cos-applicable-facets, Part 2 §4.1.6).
const (
	facetsStringy = xsd.FacetsCommon | xsd.FacetsLength
	facetsNumeric = xsd.FacetsCommon | xsd.FacetBounds
	facetsDecimal = facetsNumeric | xsd.FacetTotalDigits | xsd.FacetFractionDigits
	// precisionDecimal admits totalDigits, maxScale and minScale but NOT
	// fractionDigits nor the length facets (precisionDecimal Note §4, the
	// per-facet applicability lists); cos-applicable-facets rejects the rest.
	facetsPrecisionDecimal = facetsNumeric | xsd.FacetTotalDigits | xsd.FacetMaxScale | xsd.FacetMinScale
	facetsDateTimey        = facetsNumeric | xsd.FacetExplicitTimezone
	facetsBoolean          = xsd.FacetPattern | xsd.FacetWhiteSpace | xsd.FacetAssertion
)

// The authored fundamental-facet base case of each primitive (Part 2 §F.1). Only
// {ordered}/{numeric} vary; {bounded}/{cardinality} are left at their zero value
// (false / CardinalityCountablyInfinite), which is correct — every primitive is
// unbounded and countably infinite. SimpleType.Fundamentals copies these by
// value, so sharing one preset across primitives is safe.
var (
	fundUnordered = &xsd.Fundamentals{Ordered: xsd.OrderedFalse, Numeric: false}  // string/boolean/binary/anyURI/QName/NOTATION
	fundDecimal   = &xsd.Fundamentals{Ordered: xsd.OrderedTotal, Numeric: true}   // decimal and its derivations
	fundFloating  = &xsd.Fundamentals{Ordered: xsd.OrderedPartial, Numeric: true} // float/double
	fundPDecimal  = &xsd.Fundamentals{Ordered: xsd.OrderedPartial, Numeric: true} // precisionDecimal (±INF/NaN make it partially ordered)
	fundTemporal  = &xsd.Fundamentals{Ordered: xsd.OrderedPartial}                // duration + the date/time family
)

var (
	String  = primitive("string", xsd.WSPreserve, facetsStringy, fundUnordered, parseAsString)
	Boolean = primitive("boolean", xsd.WSCollapse, facetsBoolean, fundUnordered, parseBoolean)
	Decimal = primitive("decimal", xsd.WSCollapse, facetsDecimal, fundDecimal, parseDecimalV)
	Float   = primitive("float", xsd.WSCollapse, facetsNumeric, fundFloating, parseFloat)
	Double  = primitive("double", xsd.WSCollapse, facetsNumeric, fundFloating, parseDouble)

	// PrecisionDecimal is xs:precisionDecimal (XSD 1.1's IEEE-754 decimal
	// floating-point datatype; the precisionDecimal Note §3.2.5). The
	// unconstrained primitive carries NO scale facets: its value space admits
	// any scale, so a direct restriction may set minScale/maxScale to any signed
	// value (corpus pdecimal016/019/020.xsd set negative minScale and are valid).
	// minScale/maxScale narrowing is enforced only between derivation steps that
	// both declare the facet (checkScaleRestriction), per the *-valid-restriction
	// rules.
	PrecisionDecimal = primitive("precisionDecimal", xsd.WSCollapse, facetsPrecisionDecimal, fundPDecimal, parsePrecisionDecimalV)

	Duration   = primitive("duration", xsd.WSCollapse, facetsNumeric, fundTemporal, parseDurationV)
	DateTime   = primitive("dateTime", xsd.WSCollapse, facetsDateTimey, fundTemporal, xsdtype.ParseDateTime)
	Time       = primitive("time", xsd.WSCollapse, facetsDateTimey, fundTemporal, xsdtype.ParseTime)
	Date       = primitive("date", xsd.WSCollapse, facetsDateTimey, fundTemporal, xsdtype.ParseDate)
	GYearMonth = primitive("gYearMonth", xsd.WSCollapse, facetsDateTimey, fundTemporal, xsdtype.ParseGYearMonth)
	GYear      = primitive("gYear", xsd.WSCollapse, facetsDateTimey, fundTemporal, xsdtype.ParseGYear)
	GMonthDay  = primitive("gMonthDay", xsd.WSCollapse, facetsDateTimey, fundTemporal, xsdtype.ParseGMonthDay)
	GDay       = primitive("gDay", xsd.WSCollapse, facetsDateTimey, fundTemporal, xsdtype.ParseGDay)
	GMonth     = primitive("gMonth", xsd.WSCollapse, facetsDateTimey, fundTemporal, xsdtype.ParseGMonth)

	HexBinary    = primitive("hexBinary", xsd.WSCollapse, facetsStringy, fundUnordered, parseHexBinary)
	Base64Binary = primitive("base64Binary", xsd.WSCollapse, facetsStringy, fundUnordered, parseBase64Binary)
	AnyURI       = primitive("anyURI", xsd.WSCollapse, facetsStringy, fundUnordered, parseAnyURI)
	QName        = primitive("QName", xsd.WSCollapse, facetsStringy, fundUnordered, parseQName)
	NOTATION     = primitive("NOTATION", xsd.WSCollapse, facetsStringy, fundUnordered, parseQName)
)

// ---- derived numeric ladder ----

var (
	Integer = restrict("integer", Decimal, func(f *xsd.Facets) {
		f.FractionDigits = intFacet(0, true)
		pattern(f, `[\-+]?[0-9]+`)
	})
	NonPositiveInteger = restrict("nonPositiveInteger", Integer, func(f *xsd.Facets) {
		f.MaxInclusive = decBound("0", false)
	})
	NegativeInteger = restrict("negativeInteger", NonPositiveInteger, func(f *xsd.Facets) {
		f.MaxInclusive = decBound("-1", false)
	})
	Long = restrict("long", Integer, func(f *xsd.Facets) {
		f.MinInclusive = decBound("-9223372036854775808", false)
		f.MaxInclusive = decBound("9223372036854775807", false)
	})
	Int = restrict("int", Long, func(f *xsd.Facets) {
		f.MinInclusive = decBound("-2147483648", false)
		f.MaxInclusive = decBound("2147483647", false)
	})
	Short = restrict("short", Int, func(f *xsd.Facets) {
		f.MinInclusive = decBound("-32768", false)
		f.MaxInclusive = decBound("32767", false)
	})
	Byte = restrict("byte", Short, func(f *xsd.Facets) {
		f.MinInclusive = decBound("-128", false)
		f.MaxInclusive = decBound("127", false)
	})
	NonNegativeInteger = restrict("nonNegativeInteger", Integer, func(f *xsd.Facets) {
		f.MinInclusive = decBound("0", false)
	})
	UnsignedLong = restrict("unsignedLong", NonNegativeInteger, func(f *xsd.Facets) {
		f.MaxInclusive = decBound("18446744073709551615", false)
	})
	UnsignedInt = restrict("unsignedInt", UnsignedLong, func(f *xsd.Facets) {
		f.MaxInclusive = decBound("4294967295", false)
	})
	UnsignedShort = restrict("unsignedShort", UnsignedInt, func(f *xsd.Facets) {
		f.MaxInclusive = decBound("65535", false)
	})
	UnsignedByte = restrict("unsignedByte", UnsignedShort, func(f *xsd.Facets) {
		f.MaxInclusive = decBound("255", false)
	})
	PositiveInteger = restrict("positiveInteger", NonNegativeInteger, func(f *xsd.Facets) {
		f.MinInclusive = decBound("1", false)
	})
)

// ---- derived string ladder ----

var (
	NormalizedString = restrict("normalizedString", String, func(f *xsd.Facets) {
		f.WhiteSpace = xsd.WSReplace
	})
	Token = restrict("token", NormalizedString, func(f *xsd.Facets) {
		f.WhiteSpace = xsd.WSCollapse
	})
	Language = restrict("language", Token, func(f *xsd.Facets) {
		pattern(f, `[a-zA-Z]{1,8}(-[a-zA-Z0-9]{1,8})*`)
	})
	NMTOKEN = restrict("NMTOKEN", Token, func(f *xsd.Facets) {
		pattern(f, `\c+`)
	})
	Name = restrict("Name", Token, func(f *xsd.Facets) {
		pattern(f, `\i\c*`)
	})
	NCName = restrict("NCName", Name, func(f *xsd.Facets) {
		pattern(f, `[\i-[:]][\c-[:]]*`)
	})
	ID     = restrict("ID", NCName, nil)
	IDREF  = restrict("IDREF", NCName, nil)
	ENTITY = restrict("ENTITY", NCName, nil)
)

// ---- built-in list types ----

var (
	NMTOKENS = list("NMTOKENS", NMTOKEN, func(f *xsd.Facets) {
		f.MinLength = intFacet(1, false)
	})
	IDREFS = list("IDREFS", IDREF, func(f *xsd.Facets) {
		f.MinLength = intFacet(1, false)
	})
	ENTITIES = list("ENTITIES", ENTITY, func(f *xsd.Facets) {
		f.MinLength = intFacet(1, false)
	})
)

// ---- XSD 1.1 additions ----

var (
	DateTimeStamp = restrict("dateTimeStamp", DateTime, func(f *xsd.Facets) {
		f.ExplicitTimezone = xsd.ETZRequired
		f.ExplicitTimezoneFixed = true
	})
	YearMonthDuration = restrict("yearMonthDuration", Duration, func(f *xsd.Facets) {
		pattern(f, `[^DT]*`)
	})
	DayTimeDuration = restrict("dayTimeDuration", Duration, func(f *xsd.Facets) {
		pattern(f, `[^YM]*(T.*)?`)
	})
)

// AllBuiltins returns every built-in simple type (anySimpleType and
// anyAtomicType included), for seeding a registry.
func AllBuiltins() []*xsd.SimpleType {
	return []*xsd.SimpleType{
		AnySimpleType, AnyAtomicType,
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

// xsiSchemaLocationType is the anonymous list-of-anyURI type of the built-in
// xsi:schemaLocation attribute.
var xsiSchemaLocationType = &xsd.SimpleType{
	Name:     xsd.QName{Namespace: xsd.XSINS, Local: "schemaLocationType"},
	BaseType: AnySimpleType,
	Variety:  xsd.VarietyList,
	ItemType: AnyURI,
}

func xsiAttr(local string, t *xsd.SimpleType) *xsd.AttributeDecl {
	return &xsd.AttributeDecl{
		Name:        xsd.QName{Namespace: xsd.XSINS, Local: local},
		Global:      true,
		Form:        xsd.FormQualified,
		Type:        t,
		Inheritable: false,
	}
}

// XSIAttributes are the four built-in attribute declarations in the XSI
// namespace (§3.2.7): xsi:type, xsi:nil, xsi:schemaLocation and
// xsi:noNamespaceSchemaLocation. They are present in every schema and a
// reference to one requires no <import> of the XSI namespace.
var XSIAttributes = []*xsd.AttributeDecl{
	xsiAttr("type", QName),
	xsiAttr("nil", Boolean),
	xsiAttr("schemaLocation", xsiSchemaLocationType),
	xsiAttr("noNamespaceSchemaLocation", AnyURI),
}

// Lookup returns the built-in simple type with the given local name in the
// XSD namespace, or nil.
func Lookup(local string) *xsd.SimpleType {
	return byName[local]
}

var byName = func() map[string]*xsd.SimpleType {
	m := make(map[string]*xsd.SimpleType)
	for _, t := range AllBuiltins() {
		m[t.Name.Local] = t
	}
	return m
}()
