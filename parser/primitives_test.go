package parser

// Tests for Options.Primitives: the parse-time hook that swaps the simple-type
// value layer (strict builtin vs. lax builtin/gotype).

import (
	"testing"

	"github.com/kud360/goxsd5/builtin"
	"github.com/kud360/goxsd5/builtin/gotype"
	"github.com/kud360/goxsd5/xsd"
)

const primSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="urn:t" xmlns:t="urn:t">
  <xs:simpleType name="count">
    <xs:restriction base="xs:int"/>
  </xs:simpleType>
</xs:schema>`

func parsePrim(t *testing.T, prims []*xsd.SimpleType) *xsd.SimpleType {
	t.Helper()
	opts := &Options{Resolver: mapResolver{"s.xsd": primSchema}, Primitives: prims}
	schemas, err := Parse("s.xsd", opts)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	st, _ := schemas[0].Types[xsd.QName{Namespace: "urn:t", Local: "count"}].(*xsd.SimpleType)
	if st == nil {
		t.Fatal("type count not found")
	}
	return st
}

// TestPrimitivesDefaultUnchanged: nil Primitives keeps the strict value space —
// a restriction of xs:int does not parse to a Go-native int64.
func TestPrimitivesDefaultUnchanged(t *testing.T) {
	st := parsePrim(t, nil)
	v, err := st.ParseValue("42", nil)
	if err != nil {
		t.Fatalf("ParseValue: %v", err)
	}
	if _, ok := v.(int64); ok {
		t.Error("default (strict) parse produced a Go-native int64")
	}
}

// TestPrimitivesLaxSwitch: gotype.AllBuiltins() makes the same restriction parse
// to a Go-native int64, and the inherited xs:int bounds still apply.
func TestPrimitivesLaxSwitch(t *testing.T) {
	st := parsePrim(t, gotype.AllBuiltins())
	v, err := st.ParseValue("42", nil)
	if err != nil {
		t.Fatalf("ParseValue: %v", err)
	}
	if _, ok := v.(int64); !ok {
		t.Errorf("lax parse produced %T, want int64", v)
	}
	// The xs:int upper bound (2147483647) is inherited and enforced lax.
	if _, err := st.ParseValue("9999999999", nil); err == nil {
		t.Error("lax xs:int restriction accepted a value over the int bound")
	}
}

// TestPrimitivesReplaceMix: Replace lets a single type go lax while the rest
// stay strict.
func TestPrimitivesReplaceMix(t *testing.T) {
	// Without replacing xs:int, the restriction stays strict even though a lax
	// double is mixed in.
	mixed := gotype.Replace(builtin.AllBuiltins(), gotype.Double)
	st := parsePrim(t, mixed)
	v, err := st.ParseValue("42", nil)
	if err != nil {
		t.Fatalf("ParseValue: %v", err)
	}
	if _, ok := v.(int64); ok {
		t.Error("xs:int parsed lax though only double was replaced")
	}
}
