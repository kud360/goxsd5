package xsdtype

import (
	"errors"
	"testing"

	"github.com/kud360/goxsd5/xsd"
)

// These tests exercise the PD-2 facet-engine support (maxScale/minScale
// validation and narrowing, the totalDigits special-value exemption, and the
// NaN identity-vs-order split) by driving the engine through a hand-built
// precisionDecimal-backed SimpleType. The facet *elements* are not yet wired
// into the parser (that is PD-3), so the corpus cases pdecimal006/007/008 are
// not yet reachable; these unit tests stand in until then.

// pdType builds an atomic SimpleType whose lexical→value mapping yields a
// *PrecisionDecimal and whose comparator is the shared CompareValues, with the
// given declared facets. It is the minimal vehicle for running the facet
// pipeline against precisionDecimal values without the (not-yet-registered)
// builtin.
func pdType(f xsd.Facets) *xsd.SimpleType {
	return &xsd.SimpleType{
		Variety: xsd.VarietyAtomic,
		Parse: func(lexical string, _ xsd.ValueContext) (xsd.Value, error) {
			return ParsePrecisionDecimal(lexical)
		},
		Compare:        CompareValues,
		DeclaredFacets: f,
	}
}

func intFacet(v int) *xsd.IntFacet { return &xsd.IntFacet{Value: v} }

func TestPDMaxScaleValid(t *testing.T) {
	st := pdType(xsd.Facets{MaxScale: intFacet(0)})
	cases := []struct {
		in string
		ok bool
	}{
		// scale 0 — allowed.
		{"3", true},
		// scale 2 — exceeds maxScale 0.
		{"3.00", false},
		{"3.0", false},
		// negative scale (3.0e2 has scale −1) — within maxScale 0.
		{"3.0e2", true},
		// special values have no scale: always pass (§4.2 clause 2).
		{"NaN", true},
		{"INF", true},
		{"-INF", true},
	}
	for _, tc := range cases {
		_, err := st.ParseValue(tc.in, nil)
		if (err == nil) != tc.ok {
			t.Errorf("maxScale=0 ParseValue(%q): err=%v, want ok=%v", tc.in, err, tc.ok)
		}
		if err != nil && !tc.ok {
			assertSpec(t, err, "cvc-maxScale-valid")
		}
	}
}

func TestPDMinScaleValid(t *testing.T) {
	st := pdType(xsd.Facets{MinScale: intFacet(2)})
	cases := []struct {
		in string
		ok bool
	}{
		// scale 2 — meets minScale 2.
		{"3.00", true},
		// scale 0 / 1 — below minScale 2.
		{"3", false},
		{"3.0", false},
		// scale 3 — above minScale 2.
		{"3.000", true},
		// specials always pass (§4.3 clause 2).
		{"NaN", true},
		{"INF", true},
		{"-INF", true},
	}
	for _, tc := range cases {
		_, err := st.ParseValue(tc.in, nil)
		if (err == nil) != tc.ok {
			t.Errorf("minScale=2 ParseValue(%q): err=%v, want ok=%v", tc.in, err, tc.ok)
		}
		if err != nil && !tc.ok {
			assertSpec(t, err, "cvc-minScale-valid")
		}
	}
}

// 3 vs 3.00 under maxScale=0 and under minScale=2: the cohort members are
// distinguished by their scale even though they are numerically equal.
func TestPDScaleDistinguishesCohort(t *testing.T) {
	maxZero := pdType(xsd.Facets{MaxScale: intFacet(0)})
	if _, err := maxZero.ParseValue("3", nil); err != nil {
		t.Errorf("maxScale=0 should accept 3 (scale 0): %v", err)
	}
	if _, err := maxZero.ParseValue("3.00", nil); err == nil {
		t.Error("maxScale=0 should reject 3.00 (scale 2)")
	}

	minTwo := pdType(xsd.Facets{MinScale: intFacet(2)})
	if _, err := minTwo.ParseValue("3.00", nil); err != nil {
		t.Errorf("minScale=2 should accept 3.00 (scale 2): %v", err)
	}
	if _, err := minTwo.ParseValue("3", nil); err == nil {
		t.Error("minScale=2 should reject 3 (scale 0)")
	}
}

// totalDigits is exempt for the special values and zero (§4.1 clause 1): NaN,
// ±INF and a zero never violate a totalDigits facet.
func TestPDTotalDigitsSpecialExemption(t *testing.T) {
	st := pdType(xsd.Facets{TotalDigits: intFacet(1)})
	for _, in := range []string{"NaN", "INF", "-INF", "0", "0.0", "-0.0", "0.0000"} {
		if _, err := st.ParseValue(in, nil); err != nil {
			t.Errorf("totalDigits=1 should exempt %q: %v", in, err)
		}
	}
	// A non-zero finite value with too many digits still fails.
	if _, err := st.ParseValue("12", nil); err == nil {
		t.Error("totalDigits=1 should reject 12 (2 digits)")
	}
}

// NaN identity-vs-order: enumeration matches NaN to NaN (identity), while the
// order relation keeps NaN incomparable so an order-sensitive facet never
// admits it.
func TestPDNaNIdentityVsOrder(t *testing.T) {
	nanEnum := pdType(xsd.Facets{
		Enumeration: []xsd.Enum{
			{Value: mustPD(t, "NaN"), Lexical: "NaN"},
			{Value: mustPD(t, "1.0"), Lexical: "1.0"},
		},
	})
	// NaN enumerates to NaN.
	if _, err := nanEnum.ParseValue("NaN", nil); err != nil {
		t.Errorf("NaN should match the NaN enumeration member: %v", err)
	}
	// Cohort members of 1.0 all match the 1.0 member (numerical identity).
	for _, in := range []string{"1.0", "1", "1e0", "10e-1", "1.00000000000"} {
		if _, err := nanEnum.ParseValue(in, nil); err != nil {
			t.Errorf("%q should match the 1.0 enumeration member: %v", in, err)
		}
	}
	// A value not in the set fails.
	if _, err := nanEnum.ParseValue("17.3", nil); err == nil {
		t.Error("17.3 should not match the enumeration")
	}

	// Order relation: NaN is incomparable with itself, so an order-sensitive
	// bound never admits it. maxInclusive=NaN against a NaN value must reject.
	nanBound := pdType(xsd.Facets{
		MaxInclusive: &xsd.Bound{Value: mustPD(t, "NaN"), Lexical: "NaN"},
	})
	if _, err := nanBound.ParseValue("NaN", nil); err == nil {
		t.Error("maxInclusive=NaN must reject NaN (incomparable in the order relation)")
	}
}

// The Identical capability is exposed and decoupled from the order relation.
func TestPDIdenticalCapability(t *testing.T) {
	var v xsd.Value = mustPD(t, "NaN")
	id, ok := v.(xsd.Identical)
	if !ok {
		t.Fatal("*PrecisionDecimal should satisfy xsd.Identical")
	}
	if !id.Identical(mustPD(t, "NaN")) {
		t.Error("NaN should be identical to NaN")
	}
	if id.Identical(mustPD(t, "1.0")) {
		t.Error("NaN should not be identical to 1.0")
	}
	// Cohort members are identical (numerical equality), ±0 too.
	one := xsd.Identical(mustPD(t, "3"))
	if !one.Identical(mustPD(t, "3.00")) {
		t.Error("3 should be identical to 3.00")
	}
	zero := xsd.Identical(mustPD(t, "0"))
	if !zero.Identical(mustPD(t, "-0.0")) {
		t.Error("0 should be identical to -0.0")
	}
	// Cross-kind is never identical.
	dec, _ := ParseDecimal("3")
	if one.Identical(dec) {
		t.Error("precisionDecimal should not be identical to a decimal")
	}
}

// Restriction narrowing for the scale facets: a declared maxScale may not
// exceed the base maxScale, a declared minScale may not fall below the base
// minScale, the effective minScale may not exceed the effective maxScale, and
// a fixed base scale facet may not change.
func TestPDScaleRestriction(t *testing.T) {
	cases := []struct {
		name     string
		declared xsd.Facets
		base     xsd.Facets
		wantID   string // "" = no error
	}{
		{
			name:     "maxScale narrows down — ok",
			declared: xsd.Facets{MaxScale: intFacet(2)},
			base:     xsd.Facets{MaxScale: intFacet(4)},
		},
		{
			name:     "maxScale widens — error",
			declared: xsd.Facets{MaxScale: intFacet(5)},
			base:     xsd.Facets{MaxScale: intFacet(4)},
			wantID:   "maxScale-valid-restriction",
		},
		{
			name:     "minScale narrows up — ok",
			declared: xsd.Facets{MinScale: intFacet(3)},
			base:     xsd.Facets{MinScale: intFacet(1)},
		},
		{
			name:     "minScale widens down — error",
			declared: xsd.Facets{MinScale: intFacet(0)},
			base:     xsd.Facets{MinScale: intFacet(1)},
			wantID:   "minScale-valid-restriction",
		},
		{
			name:     "effective minScale > maxScale — error",
			declared: xsd.Facets{MinScale: intFacet(3)},
			base:     xsd.Facets{MaxScale: intFacet(2)},
			wantID:   "minScale-totalDigits",
		},
		{
			name:     "minScale == maxScale — ok",
			declared: xsd.Facets{MinScale: intFacet(2), MaxScale: intFacet(2)},
			base:     xsd.Facets{},
		},
		{
			name:     "negative scales narrow — ok",
			declared: xsd.Facets{MinScale: intFacet(-2), MaxScale: intFacet(-1)},
			base:     xsd.Facets{MinScale: intFacet(-3), MaxScale: intFacet(0)},
		},
		{
			name:     "fixed base maxScale changed — error",
			declared: xsd.Facets{MaxScale: intFacet(1)},
			base:     xsd.Facets{MaxScale: &xsd.IntFacet{Value: 2, Fixed: true}},
			wantID:   "fixed-facet-value",
		},
	}
	for _, tc := range cases {
		err := xsd.CheckFacetRestriction(&tc.declared, &tc.base, CompareValues)
		if tc.wantID == "" {
			if err != nil {
				t.Errorf("%s: unexpected error %v", tc.name, err)
			}
			continue
		}
		if err == nil {
			t.Errorf("%s: want error %q, got nil", tc.name, tc.wantID)
			continue
		}
		if !hasSpecID(err, tc.wantID) {
			t.Errorf("%s: error %v does not carry id %q", tc.name, err, tc.wantID)
		}
	}
}

// minScale greater than totalDigits is explicitly NOT a restriction error
// (§4.3): a base totalDigits below the declared minScale must be accepted.
func TestPDMinScaleVsTotalDigitsNotAnError(t *testing.T) {
	declared := xsd.Facets{MinScale: intFacet(5)}
	base := xsd.Facets{TotalDigits: intFacet(2)}
	if err := xsd.CheckFacetRestriction(&declared, &base, CompareValues); err != nil {
		t.Errorf("minScale > totalDigits must not be an error, got %v", err)
	}
}

// hasSpecID reports whether err (an aggregate or single xsd.Error) carries a
// component error with the given spec id.
func hasSpecID(err error, id string) bool {
	for _, got := range xsd.RefIDs(err) {
		if got == id {
			return true
		}
	}
	return false
}

func assertSpec(t *testing.T, err error, id string) {
	t.Helper()
	var se *xsd.Error
	if !errors.As(err, &se) {
		t.Errorf("error %v is not an *xsd.Error", err)
		return
	}
	if se.Ref.ID != id {
		t.Errorf("error id = %q, want %q", se.Ref.ID, id)
	}
}
