package builtin

import (
	"errors"
	"testing"

	"github.com/kud360/goxsd5/xsd"
)

type testCtx map[string]string

func (c testCtx) ResolveQName(prefix, local string) (xsd.QName, bool) {
	uri, ok := c[prefix]
	if !ok {
		return xsd.QName{}, false
	}
	return xsd.QName{Namespace: uri, Local: local}, true
}

func TestParseValueAcceptReject(t *testing.T) {
	cases := []struct {
		typ     *xsd.SimpleType
		valid   []string
		invalid []string
	}{
		{Boolean, []string{"true", "false", "1", "0", " true "}, []string{"TRUE", "yes", "", "2"}},
		{Decimal, []string{"1.5", "-001", "+.5"}, []string{"1e5", "INF", "one", ""}},
		{Float, []string{"1e5", "-INF", "+INF", "NaN", "12.78e-2", "12"}, []string{"INF+", "1e", "e5", "NAN"}},
		{Double, []string{"0", "-0", "1E300"}, []string{"0x1", "1.0ee5"}},
		{Integer, []string{"42", "-0042", "+7"}, []string{"1.0", "1.", "4 2"}},
		{Byte, []string{"127", "-128"}, []string{"128", "-129"}},
		{UnsignedByte, []string{"0", "255"}, []string{"-1", "256"}},
		{PositiveInteger, []string{"1"}, []string{"0", "-1"}},
		{NonPositiveInteger, []string{"0", "-99999999999999999999"}, []string{"1"}},
		{Long, []string{"9223372036854775807", "-9223372036854775808"}, []string{"9223372036854775808"}},
		{UnsignedLong, []string{"18446744073709551615"}, []string{"18446744073709551616"}},
		{HexBinary, []string{"0FB7", "0fb7", ""}, []string{"FB7", "0G"}},
		{Base64Binary, []string{"MTIz", "bW9z dGx5", "bW9zdGx5IGhhcm1sZXNz", ""}, []string{"M===", "====", "MT=z"}},
		{Language, []string{"en", "en-US", "x-klingon"}, []string{"123", "verylonglanguagetag", "en--us"}},
		{Name, []string{"foo", "_foo", "a:b", "a-b"}, []string{"1foo", "-foo", "foo bar", ""}},
		{NCName, []string{"foo", "_foo"}, []string{"a:b", "1foo"}},
		{NMTOKEN, []string{"foo", "1foo", "-x"}, []string{"foo bar", ""}},
		{Token, []string{"any old text"}, nil},
		{Duration, []string{"P1Y", "-PT0.5S"}, []string{"P", "P1H"}},
		{DateTime, []string{"2002-10-10T12:00:00Z"}, []string{"2002-10-10"}},
		{DateTimeStamp, []string{"2002-10-10T12:00:00Z", "2002-10-10T12:00:00+05:00"}, []string{"2002-10-10T12:00:00"}},
		{YearMonthDuration, []string{"P1Y2M", "-P3M"}, []string{"P1D", "PT1M", "P1Y2M3D"}},
		{DayTimeDuration, []string{"P1D", "PT1M", "-PT0.5S"}, []string{"P1Y", "P3M", "P1MT1M"}},
		{NMTOKENS, []string{"a b c", "x"}, []string{""}},
		{GMonthDay, []string{"--02-29", "--12-31+13:00"}, []string{"--02-30", "02-20"}},
		{AnyURI, []string{"http://example.com", "../relative", ""}, nil},
	}
	for _, tc := range cases {
		for _, s := range tc.valid {
			if _, err := tc.typ.ParseValue(s, nil); err != nil {
				t.Errorf("%s.ParseValue(%q): %v", tc.typ.Name.Local, s, err)
			}
		}
		for _, s := range tc.invalid {
			if _, err := tc.typ.ParseValue(s, nil); err == nil {
				t.Errorf("%s.ParseValue(%q) should fail", tc.typ.Name.Local, s)
			}
		}
	}
}

func TestQNameContext(t *testing.T) {
	// Without context: typed error.
	if _, err := QName.ParseValue("tns:foo", nil); !errors.Is(err, xsd.ErrNeedContext) {
		t.Errorf("QName without context: %v", err)
	}
	ctx := testCtx{"tns": "http://tns", "": "http://default"}
	v, err := QName.ParseValue("tns:foo", ctx)
	if err != nil {
		t.Fatalf("QName with context: %v", err)
	}
	qv := v.(xsd.QNameValue)
	if qv.Name != (xsd.QName{Namespace: "http://tns", Local: "foo"}) || qv.Lexical != "tns:foo" {
		t.Errorf("QName value = %+v", qv)
	}
	if _, err := QName.ParseValue("undefined:foo", ctx); err == nil {
		t.Error("undefined prefix should fail")
	}
	if _, err := QName.ParseValue(":foo", ctx); err == nil {
		t.Error("empty prefix should fail")
	}
}

func TestWhitespaceHandling(t *testing.T) {
	// string preserves whitespace.
	v, _ := String.ParseValue("  a  b ", nil)
	if v != xsd.String("  a  b ") {
		t.Errorf("string ws: %q", v)
	}
	// normalizedString replaces tabs/newlines.
	v, _ = NormalizedString.ParseValue("a\tb\nc", nil)
	if v != xsd.String("a b c") {
		t.Errorf("normalizedString ws: %q", v)
	}
	// token collapses.
	v, _ = Token.ParseValue("  a \t b  ", nil)
	if v != xsd.String("a b") {
		t.Errorf("token ws: %q", v)
	}
}

func TestListParsing(t *testing.T) {
	v, err := NMTOKENS.ParseValue(" alpha  beta ", nil)
	if err != nil {
		t.Fatal(err)
	}
	lv := v.(xsd.ListValue)
	if len(lv) != 2 || lv[0] != xsd.String("alpha") {
		t.Errorf("NMTOKENS = %v", lv)
	}
}

// TestFacetPipelineSpaceCorrectness: hexBinary length facets count octets,
// not characters.
func TestFacetPipelineSpaceCorrectness(t *testing.T) {
	declared := xsd.Facets{Length: &xsd.IntFacet{Value: 2}}
	st := &xsd.SimpleType{
		Name:      xsd.QName{Local: "twoOctets"},
		BaseType:  HexBinary,
		Variety:   xsd.VarietyAtomic,
		Primitive: HexBinary,
		Facets:    xsd.MergeFacets(&HexBinary.Facets, &declared),
	}
	if _, err := st.ParseValue("0FB7", nil); err != nil { // 4 chars = 2 octets
		t.Errorf("0FB7 should be 2 octets: %v", err)
	}
	if _, err := st.ParseValue("0F", nil); err == nil {
		t.Error("0F is 1 octet, should fail length=2")
	}
}

// TestPatternANDAcrossSteps: patterns from different derivation steps must
// all match; multiple patterns within one step are OR'd.
func TestPatternANDAcrossSteps(t *testing.T) {
	mk := func(base *xsd.SimpleType, sources ...string) *xsd.SimpleType {
		var group xsd.PatternGroup
		for _, src := range sources {
			re, err := xsd.CompileRegex(src)
			if err != nil {
				t.Fatal(err)
			}
			group = append(group, xsd.Pattern{Source: src, Re: re})
		}
		declared := xsd.Facets{PatternGroups: []xsd.PatternGroup{group}}
		return &xsd.SimpleType{
			BaseType:  base,
			Variety:   xsd.VarietyAtomic,
			Primitive: base.Primitive,
			Facets:    xsd.MergeFacets(&base.Facets, &declared),
		}
	}
	// Step 1: starts with a, OR starts with b. Step 2: ends with z.
	step1 := mk(Token, "a.*", "b.*")
	step2 := mk(step1, ".*z")
	for s, want := range map[string]bool{
		"az":  true,
		"bzz": true,
		"a":   false, // fails step 2
		"cz":  false, // fails step 1
	} {
		_, err := step2.ParseValue(s, nil)
		if got := err == nil; got != want {
			t.Errorf("step2.ParseValue(%q) ok=%v, want %v (%v)", s, got, want, err)
		}
	}
}

func TestEnumerationValueSpace(t *testing.T) {
	// Enumeration on integers is value-space: "1" and "01" are the same.
	one, _ := Integer.ParseValue("1", nil)
	declared := xsd.Facets{HasEnumeration: true, Enumeration: []xsd.Enum{{Value: one, Lexical: "1"}}}
	st := &xsd.SimpleType{
		BaseType:  Integer,
		Variety:   xsd.VarietyAtomic,
		Primitive: Decimal,
		Facets:    xsd.MergeFacets(&Integer.Facets, &declared),
	}
	if _, err := st.ParseValue("01", nil); err != nil {
		t.Errorf("01 should match enum value 1: %v", err)
	}
	if _, err := st.ParseValue("2", nil); err == nil {
		t.Error("2 should not match enum {1}")
	}
}

func TestAllBuiltinsWellFormed(t *testing.T) {
	seen := map[string]bool{}
	for _, b := range AllBuiltins() {
		if b.Name.Namespace != xsd.XSDNS || b.Name.Local == "" {
			t.Errorf("bad name %v", b.Name)
		}
		if seen[b.Name.Local] {
			t.Errorf("duplicate builtin %s", b.Name.Local)
		}
		seen[b.Name.Local] = true
		if b != AnySimpleType && b.BaseType == nil {
			t.Errorf("%s has no base", b.Name.Local)
		}
		if b.Variety == xsd.VarietyAtomic && b != AnySimpleType && b != AnyAtomicType && b.Primitive == nil {
			t.Errorf("%s has no primitive", b.Name.Local)
		}
	}
	if Lookup("int") != Int || Lookup("nope") != nil {
		t.Error("Lookup misbehaves")
	}
}
