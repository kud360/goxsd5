package xsd_test

// Tests for the Milestone 8 safe-mutation API. They live in an external test
// package so they can use the real built-in types as restriction bases
// (builtin imports xsd, so an internal test cannot).

import (
	"testing"

	"github.com/kud360/goxsd5/builtin"
	"github.com/kud360/goxsd5/xsd"
)

func TestRestrictWithNarrowing(t *testing.T) {
	// string is the base; maxLength 5 is a valid narrowing.
	sub, err := builtin.String.RestrictWith(&xsd.Facets{MaxLength: &xsd.IntFacet{Value: 5}})
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
	narrow, err := builtin.String.RestrictWith(&xsd.Facets{MaxLength: &xsd.IntFacet{Value: 3}})
	if err != nil {
		t.Fatalf("setup restriction failed: %v", err)
	}
	// Restricting narrow with maxLength 10 widens it — must be rejected.
	_, err = narrow.RestrictWith(&xsd.Facets{MaxLength: &xsd.IntFacet{Value: 10}})
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
	_, err := builtin.Integer.RestrictWith(&xsd.Facets{
		MinInclusive: &xsd.Bound{Value: min, Lexical: "5"},
		MaxInclusive: &xsd.Bound{Value: max, Lexical: "1"},
	})
	if err == nil {
		t.Fatal("inconsistent bounds accepted")
	}
}

func TestAddEnumeration(t *testing.T) {
	sub, err := builtin.Token.RestrictWith(&xsd.Facets{})
	if err != nil {
		t.Fatalf("base restriction failed: %v", err)
	}
	if err := sub.AddEnumeration("red", "green"); err != nil {
		t.Fatalf("AddEnumeration rejected valid values: %v", err)
	}
	// A second call accumulates onto the same enumeration set rather than
	// chaining a restriction that would exclude the first values.
	if err := sub.AddEnumeration("blue"); err != nil {
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
	sub, _ := builtin.Int.RestrictWith(&xsd.Facets{})
	if err := sub.AddEnumeration("notanumber"); err == nil {
		t.Fatal("non-integer enumeration value accepted for an int subtype")
	}
	if len(sub.EffectiveFacets().Enumeration) != 0 {
		t.Error("rejected AddEnumeration left a partial enumeration behind")
	}
}

func TestAddEnumerationRejectsBuiltinMutation(t *testing.T) {
	if err := builtin.Token.AddEnumeration("x"); err == nil {
		t.Fatal("mutating a built-in type in place was allowed")
	}
	// The shared built-in must be unchanged.
	if builtin.Token.EffectiveFacets().HasEnumeration {
		t.Error("built-in token gained an enumeration facet")
	}
}

func TestAddPattern(t *testing.T) {
	sub, _ := builtin.Token.RestrictWith(&xsd.Facets{})
	if err := sub.AddPattern("[a-z]+"); err != nil {
		t.Fatalf("AddPattern rejected a valid pattern: %v", err)
	}
	if _, err := sub.ParseValue("abc", nil); err != nil {
		t.Errorf("pattern-matching value rejected: %v", err)
	}
	if _, err := sub.ParseValue("ABC", nil); err == nil {
		t.Error("value not matching the pattern accepted")
	}
	if err := sub.AddPattern("[unterminated"); err == nil {
		t.Error("malformed pattern accepted")
	}
}

func TestExtensionsSurviveMutateCycle(t *testing.T) {
	// Foreign content set on a derived type must not be disturbed by a later
	// in-place mutation, and a restriction does not inherit the base's.
	base, _ := builtin.Token.RestrictWith(&xsd.Facets{})
	base.Extensions = xsd.Extensions{Attrs: []xsd.ForeignAttr{{Name: xsd.QName{Namespace: "urn:x", Local: "note"}, Value: "keep me"}}}
	if err := base.AddEnumeration("a", "b"); err != nil {
		t.Fatalf("AddEnumeration failed: %v", err)
	}
	if len(base.Extensions.Attrs) != 1 || base.Extensions.Attrs[0].Value != "keep me" {
		t.Error("foreign content lost across a mutate cycle")
	}
	sub, err := base.RestrictWith(&xsd.Facets{})
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
	if err := ct.AddElement(e1, 1, 1); err != nil {
		t.Fatalf("AddElement (first) failed: %v", err)
	}
	ec := ct.Content.(*xsd.ElementContent)
	if ec.Particle == nil || ec.Particle.Term != xsd.Term(e1) {
		t.Fatal("first element not installed as the particle term")
	}
	if err := ct.AddElement(e2, 0, xsd.UnboundedOccurs); err != nil {
		t.Fatalf("AddElement (second) failed: %v", err)
	}
	mg, ok := ec.Particle.Term.(*xsd.ModelGroup)
	if !ok || mg.Compositor != xsd.CompositorSequence || len(mg.Particles) != 2 {
		t.Fatalf("second element did not wrap content in a 2-particle sequence")
	}
	if err := ct.AddElement(e1, 2, 1); err == nil {
		t.Error("invalid occurrence range accepted")
	}
}

func TestAddElementRejectsSimpleContent(t *testing.T) {
	ct := &xsd.ComplexType{
		Name:    xsd.QName{Local: "sc"},
		Content: &xsd.SimpleContent{Type: builtin.String},
	}
	if err := ct.AddElement(&xsd.ElementDecl{Name: xsd.QName{Local: "a"}}, 1, 1); err == nil {
		t.Fatal("AddElement on simple content was allowed")
	}
}
