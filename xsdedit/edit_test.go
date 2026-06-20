package xsdedit_test

// Tests for the safe-mutation / validation API, using the real built-in types
// as restriction bases.

import (
	"testing"

	"github.com/kud360/goxsd5/builtin"
	"github.com/kud360/goxsd5/xsd"
	"github.com/kud360/goxsd5/xsdedit"
)

func TestRestrictWithNarrowing(t *testing.T) {
	// string is the base; maxLength 5 is a valid narrowing.
	sub, err := xsdedit.RestrictWith(builtin.String, &xsd.Facets{MaxLength: &xsd.IntFacet{Value: 5}})
	if err != nil {
		t.Fatalf("valid restriction rejected: %v", err)
	}
	if sub.EffectiveFacets().MaxLength == nil || sub.EffectiveFacets().MaxLength.Value != 5 {
		t.Errorf("EffectiveFacets does not reflect maxLength=5")
	}
	if _, err := sub.ParseValue("abc", nil); err != nil {
		t.Errorf("subtype rejected a valid value: %v", err)
	}
	if _, err := sub.ParseValue("abcdef", nil); err == nil {
		t.Errorf("subtype accepted a value over maxLength")
	}
}

func TestRestrictWithRejectsWidening(t *testing.T) {
	// short has maxLength via the integer chain; widen a length-bounded type.
	narrow, err := xsdedit.RestrictWith(builtin.String, &xsd.Facets{MaxLength: &xsd.IntFacet{Value: 3}})
	if err != nil {
		t.Fatalf("setup restriction failed: %v", err)
	}
	// Restricting narrow with maxLength 10 widens it — must be rejected.
	_, err = xsdedit.RestrictWith(narrow, &xsd.Facets{MaxLength: &xsd.IntFacet{Value: 10}})
	if err == nil {
		t.Fatal("widening restriction (maxLength 3 -> 10) was accepted")
	}
	if ids := xsd.RefIDs(err); len(ids) == 0 || ids[0] != "maxLength-valid-restriction" {
		t.Errorf("error id = %v, want maxLength-valid-restriction", ids)
	}
	// The base type must be untouched after the rejected mutation.
	if narrow.EffectiveFacets().MaxLength.Value != 3 {
		t.Errorf("rejected restriction altered the original (maxLength now %d)", narrow.EffectiveFacets().MaxLength.Value)
	}
}

func TestRestrictWithRejectsBadBound(t *testing.T) {
	// minInclusive 5 > maxInclusive 1 — intra-facet inconsistency.
	min, _ := builtin.Integer.ParseValue("5", nil)
	max, _ := builtin.Integer.ParseValue("1", nil)
	_, err := xsdedit.RestrictWith(builtin.Integer, &xsd.Facets{
		MinInclusive: &xsd.Bound{Value: min, Lexical: "5"},
		MaxInclusive: &xsd.Bound{Value: max, Lexical: "1"},
	})
	if err == nil {
		t.Fatal("inconsistent bounds accepted")
	}
}

func TestAddEnumeration(t *testing.T) {
	sub, err := xsdedit.RestrictWith(builtin.Token, &xsd.Facets{})
	if err != nil {
		t.Fatalf("base restriction failed: %v", err)
	}
	if err := xsdedit.AddEnumeration(sub, "red", "green"); err != nil {
		t.Fatalf("AddEnumeration rejected valid values: %v", err)
	}
	// A second call accumulates onto the same enumeration set rather than
	// chaining a restriction that would exclude the first values.
	if err := xsdedit.AddEnumeration(sub, "blue"); err != nil {
		t.Fatalf("second AddEnumeration failed: %v", err)
	}
	if got := len(sub.EffectiveFacets().Enumeration); got != 3 {
		t.Fatalf("enumeration has %d members, want 3", got)
	}
	for _, ok := range []string{"red", "green", "blue"} {
		if _, err := sub.ParseValue(ok, nil); err != nil {
			t.Errorf("enumerated value %q rejected: %v", ok, err)
		}
	}
	if _, err := sub.ParseValue("purple", nil); err == nil {
		t.Error("value outside the enumeration accepted")
	}
}

func TestAddEnumerationRejectsBadValue(t *testing.T) {
	sub, _ := xsdedit.RestrictWith(builtin.Int, &xsd.Facets{})
	if err := xsdedit.AddEnumeration(sub, "notanumber"); err == nil {
		t.Fatal("non-integer enumeration value accepted for an int subtype")
	}
	if len(sub.EffectiveFacets().Enumeration) != 0 {
		t.Error("rejected AddEnumeration left a partial enumeration behind")
	}
}

func TestAddEnumerationRejectsBuiltinMutation(t *testing.T) {
	if err := xsdedit.AddEnumeration(builtin.Token, "x"); err == nil {
		t.Fatal("mutating a built-in type in place was allowed")
	}
	// The shared built-in must be unchanged.
	if builtin.Token.EffectiveFacets().HasEnumeration() {
		t.Error("built-in token gained an enumeration facet")
	}
}

func TestAddPattern(t *testing.T) {
	sub, _ := xsdedit.RestrictWith(builtin.Token, &xsd.Facets{})
	if err := xsdedit.AddPattern(sub, "[a-z]+"); err != nil {
		t.Fatalf("AddPattern rejected a valid pattern: %v", err)
	}
	if _, err := sub.ParseValue("abc", nil); err != nil {
		t.Errorf("pattern-matching value rejected: %v", err)
	}
	if _, err := sub.ParseValue("ABC", nil); err == nil {
		t.Error("value not matching the pattern accepted")
	}
	if err := xsdedit.AddPattern(sub, "[unterminated"); err == nil {
		t.Error("malformed pattern accepted")
	}
}

func TestExtensionsSurviveMutateCycle(t *testing.T) {
	// Foreign content set on a derived type must not be disturbed by a later
	// in-place mutation, and a restriction does not inherit the base's.
	base, _ := xsdedit.RestrictWith(builtin.Token, &xsd.Facets{})
	base.Extensions = xsd.Extensions{Attrs: []xsd.ForeignAttr{{Name: xsd.QName{Namespace: "urn:x", Local: "note"}, Value: "keep me"}}}
	if err := xsdedit.AddEnumeration(base, "a", "b"); err != nil {
		t.Fatalf("AddEnumeration failed: %v", err)
	}
	if len(base.Extensions.Attrs) != 1 || base.Extensions.Attrs[0].Value != "keep me" {
		t.Error("foreign content lost across a mutate cycle")
	}
	sub, err := xsdedit.RestrictWith(base, &xsd.Facets{})
	if err != nil {
		t.Fatalf("RestrictWith failed: %v", err)
	}
	if len(sub.Extensions.Attrs) != 0 {
		t.Error("restriction subtype inherited the base's foreign content")
	}
}

func TestAddElement(t *testing.T) {
	ct := &xsd.ComplexType{
		Name:    xsd.QName{Local: "c"},
		Content: &xsd.ElementContent{},
	}
	e1 := &xsd.ElementDecl{Name: xsd.QName{Local: "a"}, Type: builtin.String}
	e2 := &xsd.ElementDecl{Name: xsd.QName{Local: "b"}, Type: builtin.String}
	if err := xsdedit.AddElement(ct, e1, 1, 1); err != nil {
		t.Fatalf("AddElement (first) failed: %v", err)
	}
	ec := ct.Content.(*xsd.ElementContent)
	if ec.Particle == nil || ec.Particle.Term != xsd.Term(e1) {
		t.Fatal("first element not installed as the particle term")
	}
	if err := xsdedit.AddElement(ct, e2, 0, xsd.UnboundedOccurs); err != nil {
		t.Fatalf("AddElement (second) failed: %v", err)
	}
	mg, ok := ec.Particle.Term.(*xsd.ModelGroup)
	if !ok || mg.Compositor != xsd.CompositorSequence || len(mg.Particles) != 2 {
		t.Fatalf("second element did not wrap content in a 2-particle sequence")
	}
	if err := xsdedit.AddElement(ct, e1, 2, 1); err == nil {
		t.Error("invalid occurrence range accepted")
	}
}

func TestAddElementRejectsSimpleContent(t *testing.T) {
	ct := &xsd.ComplexType{
		Name:    xsd.QName{Local: "sc"},
		Content: &xsd.SimpleContent{Type: builtin.String},
	}
	if err := xsdedit.AddElement(ct, &xsd.ElementDecl{Name: xsd.QName{Local: "a"}}, 1, 1); err == nil {
		t.Fatal("AddElement on simple content was allowed")
	}
}

func TestValidateGoodTypes(t *testing.T) {
	// Real built-ins and a sound restriction must pass Validate.
	for _, st := range []*xsd.SimpleType{
		builtin.String, builtin.Decimal, builtin.Int, builtin.NMTOKENS, builtin.DateTime,
	} {
		if err := xsdedit.Validate(st); err != nil {
			t.Errorf("Validate(%s) rejected a valid type: %v", st.TypeName(), err)
		}
	}
	sub, err := xsdedit.RestrictWith(builtin.Decimal, &xsd.Facets{TotalDigits: &xsd.IntFacet{Value: 4}})
	if err != nil {
		t.Fatalf("setup restriction failed: %v", err)
	}
	if err := xsdedit.Validate(sub); err != nil {
		t.Errorf("Validate rejected a valid decimal restriction: %v", err)
	}
}

func TestValidateMissingParse(t *testing.T) {
	// Atomic type with no Parse anywhere in the chain.
	st := &xsd.SimpleType{Name: xsd.QName{Local: "noparse"}, Variety: xsd.VarietyAtomic}
	if err := xsdedit.Validate(st); err == nil {
		t.Error("Validate accepted an atomic type with no lexical mapping")
	}
}

func TestValidateMissingComparator(t *testing.T) {
	// Order facet but no comparator resolvable (detached, no Compare in chain).
	st := &xsd.SimpleType{
		Name:    xsd.QName{Local: "nocmp"},
		Variety: xsd.VarietyAtomic,
		Parse:   func(s string, _ xsd.ValueContext) (xsd.Value, error) { return nil, nil },
	}
	st.DeclaredFacets.MinInclusive = &xsd.Bound{Lexical: "0"}
	if err := xsdedit.Validate(st); err == nil {
		t.Error("Validate accepted order facets with no comparator")
	}
}

func TestValidateInapplicableFacet(t *testing.T) {
	// length is not applicable to a restriction of xs:decimal.
	st := &xsd.SimpleType{
		Name:     xsd.QName{Local: "baddec"},
		Variety:  xsd.VarietyAtomic,
		BaseType: builtin.Decimal,
	}
	st.DeclaredFacets.Length = &xsd.IntFacet{Value: 3}
	if err := xsdedit.Validate(st); err == nil {
		t.Error("Validate accepted a length facet on a decimal restriction")
	}
}

func TestValidateListItemList(t *testing.T) {
	st := &xsd.SimpleType{
		Name:     xsd.QName{Local: "listoflist"},
		Variety:  xsd.VarietyList,
		ItemType: builtin.NMTOKENS, // itself a list
	}
	if err := xsdedit.Validate(st); err == nil {
		t.Error("Validate accepted a list whose item type is a list")
	}
}

func upperParse(s string, _ xsd.ValueContext) (xsd.Value, error) { return s, nil }

func strCompare(a, b xsd.Value) (xsd.Order, bool) {
	as, aok := a.(string)
	bs, bok := b.(string)
	if !aok || !bok {
		return 0, false
	}
	switch {
	case as < bs:
		return xsd.OrderLess, true
	case as > bs:
		return xsd.OrderGreater, true
	}
	return xsd.OrderEqual, true
}

func TestNewPrimitiveGood(t *testing.T) {
	q := xsd.QName{Namespace: "urn:custom", Local: "code"}
	prim, err := xsdedit.NewPrimitive(q, builtin.AnyAtomicType, upperParse, strCompare, xsd.FacetsCommon|xsd.FacetsLength, xsd.WSCollapse)
	if err != nil {
		t.Fatalf("NewPrimitive rejected a sound primitive: %v", err)
	}
	if prim.PrimitiveType() != prim {
		t.Error("a custom primitive is not its own PrimitiveType")
	}
	if !prim.ApplicableFacets().Has(xsd.FacetLength) {
		t.Error("custom primitive does not admit its declared applicable facets")
	}
	// A length restriction of it must work end to end through the lax value.
	sub, err := xsdedit.RestrictWith(prim, &xsd.Facets{MaxLength: &xsd.IntFacet{Value: 3}})
	if err != nil {
		t.Fatalf("restriction of custom primitive failed: %v", err)
	}
	if _, err := sub.ParseValue("ab", nil); err != nil {
		t.Errorf("custom primitive subtype rejected a valid value: %v", err)
	}
}

func TestNewPrimitiveRejectsNoApplicable(t *testing.T) {
	q := xsd.QName{Namespace: "urn:custom", Local: "code"}
	_, err := xsdedit.NewPrimitive(q, builtin.AnyAtomicType, upperParse, nil, 0, xsd.WSCollapse)
	if err == nil {
		t.Fatal("NewPrimitive accepted a primitive with no applicable facets")
	}
	if ids := xsd.RefIDs(err); len(ids) == 0 || ids[0] != "cos-applicable-facets" {
		t.Errorf("error id = %v, want cos-applicable-facets", ids)
	}
}

func TestNewPrimitiveRejectsNoParse(t *testing.T) {
	q := xsd.QName{Namespace: "urn:custom", Local: "code"}
	_, err := xsdedit.NewPrimitive(q, builtin.AnyAtomicType, nil, nil, xsd.FacetsCommon, xsd.WSCollapse)
	if err == nil {
		t.Error("NewPrimitive accepted a primitive with no Parse")
	}
}

func TestValidatePrimitiveNeedsOwnParse(t *testing.T) {
	// A type that carries Applicable (so it is a primitive) but borrows its
	// Parse from the base chain is malformed: a primitive defines its own space.
	st := &xsd.SimpleType{
		Name:       xsd.QName{Local: "borrow"},
		Variety:    xsd.VarietyAtomic,
		BaseType:   builtin.String, // resolves a Parse up the chain
		Applicable: xsd.FacetsCommon,
	}
	if err := xsdedit.Validate(st); err == nil {
		t.Error("Validate accepted a primitive with no Parse of its own")
	}
}
