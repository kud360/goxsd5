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

	"github.com/kud360/goxsd5/xsd"
)

func qn(local string) xsd.QName { return xsd.QName{Namespace: xsd.XSDNS, Local: local} }

// AnyType is xs:anyType, the root of the type hierarchy (a complex type
// with unconstrained content). Its base type is itself, represented as nil.
var AnyType = &xsd.ComplexType{
	Name:  qn("anyType"),
	Mixed: true,
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
// facets, lexical space = all strings.
var AnySimpleType = &xsd.SimpleType{
	Name:      qn("anySimpleType"),
	BaseType:  AnyType,
	Variety:   xsd.VarietyAtomic,
	Parse:     parseAsString,
	IsBuiltin: true,
}

// AnyAtomicType is xs:anyAtomicType (XSD 1.1).
var AnyAtomicType = &xsd.SimpleType{
	Name:      qn("anyAtomicType"),
	BaseType:  AnySimpleType,
	Variety:   xsd.VarietyAtomic,
	IsBuiltin: true,
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
	IsBuiltin: true,
}

func parseAsString(s string, _ xsd.ValueContext) (xsd.Value, error) { return xsd.String(s), nil }

// primitive constructs a primitive type: based on anyAtomicType, its own
// primitive ancestor.
func primitive(local string, ws xsd.WhiteSpace, parse xsd.ParseFunc, ordered xsd.OrderedFacet, numeric bool) *xsd.SimpleType {
	t := &xsd.SimpleType{
		Name:      qn(local),
		BaseType:  AnyAtomicType,
		Variety:   xsd.VarietyAtomic,
		Parse:     parse,
		Ordered:   ordered,
		Numeric:   numeric,
		IsBuiltin: true,
	}
	t.Primitive = t
	t.Facets.WhiteSpace = ws
	t.Facets.WhiteSpaceFixed = local != "string" && local != "anyURI"
	t.DeclaredFacets = t.Facets
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
		Primitive:      base.Primitive,
		ItemType:       base.ItemType,
		Facets:         xsd.MergeFacets(&base.Facets, &declared),
		DeclaredFacets: declared,
		Ordered:        base.Ordered,
		Numeric:        base.Numeric,
		Bounded:        base.Bounded,
		Cardinality:    base.Cardinality,
		IsBuiltin:      true,
	}
	return t
}

// list constructs a built-in list type (whiteSpace collapse, fixed).
func list(local string, item *xsd.SimpleType, mod func(f *xsd.Facets)) *xsd.SimpleType {
	t := &xsd.SimpleType{
		Name:      qn(local),
		BaseType:  AnySimpleType,
		Variety:   xsd.VarietyList,
		ItemType:  item,
		IsBuiltin: true,
	}
	t.Facets.WhiteSpace = xsd.WSCollapse
	t.Facets.WhiteSpaceFixed = true
	if mod != nil {
		var declared xsd.Facets
		mod(&declared)
		t.Facets = xsd.MergeFacets(&t.Facets, &declared)
		t.DeclaredFacets = declared
	}
	return t
}

// pattern panics on a bad source: it runs only at package init on the
// compile-time literals of the built-in type table, where an error cannot
// be returned and can only mean the literal itself is wrong.
func pattern(f *xsd.Facets, sources ...string) {
	var group xsd.PatternGroup
	for _, src := range sources {
		re, err := xsd.CompileRegex(src)
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
	d, err := xsd.ParseDecimal(s)
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
		return xsd.Boolean(true), nil
	case "false", "0":
		return xsd.Boolean(false), nil
	}
	return nil, fmt.Errorf("invalid boolean %q", s)
}

func parseDecimalV(s string, _ xsd.ValueContext) (xsd.Value, error) {
	return xsd.ParseDecimal(s)
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
	return xsd.Double(v), nil
}

func parseFloat(s string, _ xsd.ValueContext) (xsd.Value, error) {
	v, err := parseFloating(s)
	if err != nil {
		return nil, err
	}
	return xsd.Float(float32(v)), nil
}

func parseDurationV(s string, _ xsd.ValueContext) (xsd.Value, error) {
	return xsd.ParseDuration(s)
}

func dtParser(kind xsd.DateTimeKind) xsd.ParseFunc {
	return func(s string, _ xsd.ValueContext) (xsd.Value, error) {
		return xsd.ParseDateTime(kind, s)
	}
}

func parseHexBinary(s string, _ xsd.ValueContext) (xsd.Value, error) {
	if len(s)%2 != 0 {
		return nil, fmt.Errorf("invalid hexBinary %q: odd length", s)
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("invalid hexBinary %q: %w", s, err)
	}
	return xsd.Bytes(b), nil
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
	return xsd.Bytes(b), nil
}

func parseAnyURI(s string, _ xsd.ValueContext) (xsd.Value, error) {
	// Part 2 §3.3.17: the lexical space is intentionally lax (any string
	// can be %-escaped into a legal IRI); no structural validation.
	return xsd.String(s), nil
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
		return nil, fmt.Errorf("cannot resolve QName %q: %w", s, xsd.ErrNeedContext)
	}
	name, ok := ctx.ResolveQName(prefix, local)
	if !ok {
		return nil, fmt.Errorf("undefined namespace prefix %q in QName %q", prefix, s)
	}
	return xsd.QNameValue{Name: name, Lexical: s}, nil
}

var ncNamePattern = regexp.MustCompile(`\A[\p{L}_][\p{L}\p{N}\p{M}_.\x{B7}-]*\z`)

func isNCName(s string) bool {
	// Close approximation of the XML NCName production; the parser's
	// xmltree package holds the exact table, but builtin must not depend
	// on parser packages.
	return ncNamePattern.MatchString(s)
}

// ---- primitives ----

var (
	String  = primitive("string", xsd.WSPreserve, parseAsString, xsd.OrderedFalse, false)
	Boolean = primitive("boolean", xsd.WSCollapse, parseBoolean, xsd.OrderedFalse, false)
	Decimal = primitive("decimal", xsd.WSCollapse, parseDecimalV, xsd.OrderedTotal, true)
	Float   = primitive("float", xsd.WSCollapse, parseFloat, xsd.OrderedPartial, true)
	Double  = primitive("double", xsd.WSCollapse, parseDouble, xsd.OrderedPartial, true)

	Duration   = primitive("duration", xsd.WSCollapse, parseDurationV, xsd.OrderedPartial, false)
	DateTime   = primitive("dateTime", xsd.WSCollapse, dtParser(xsd.KindDateTime), xsd.OrderedPartial, false)
	Time       = primitive("time", xsd.WSCollapse, dtParser(xsd.KindTime), xsd.OrderedPartial, false)
	Date       = primitive("date", xsd.WSCollapse, dtParser(xsd.KindDate), xsd.OrderedPartial, false)
	GYearMonth = primitive("gYearMonth", xsd.WSCollapse, dtParser(xsd.KindGYearMonth), xsd.OrderedPartial, false)
	GYear      = primitive("gYear", xsd.WSCollapse, dtParser(xsd.KindGYear), xsd.OrderedPartial, false)
	GMonthDay  = primitive("gMonthDay", xsd.WSCollapse, dtParser(xsd.KindGMonthDay), xsd.OrderedPartial, false)
	GDay       = primitive("gDay", xsd.WSCollapse, dtParser(xsd.KindGDay), xsd.OrderedPartial, false)
	GMonth     = primitive("gMonth", xsd.WSCollapse, dtParser(xsd.KindGMonth), xsd.OrderedPartial, false)

	HexBinary    = primitive("hexBinary", xsd.WSCollapse, parseHexBinary, xsd.OrderedFalse, false)
	Base64Binary = primitive("base64Binary", xsd.WSCollapse, parseBase64Binary, xsd.OrderedFalse, false)
	AnyURI       = primitive("anyURI", xsd.WSCollapse, parseAnyURI, xsd.OrderedFalse, false)
	QName        = primitive("QName", xsd.WSCollapse, parseQName, xsd.OrderedFalse, false)
	NOTATION     = primitive("NOTATION", xsd.WSCollapse, parseQName, xsd.OrderedFalse, false)
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
		String, Boolean, Decimal, Float, Double,
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
	Name:      xsd.QName{Namespace: xsd.XSINS, Local: "schemaLocationType"},
	BaseType:  AnySimpleType,
	Variety:   xsd.VarietyList,
	ItemType:  AnyURI,
	IsBuiltin: true,
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
