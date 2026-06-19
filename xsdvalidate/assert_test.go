package xsdvalidate_test

// Behaviour tests for XSD 1.1 assertion evaluation (cvc-assertion): xs:assert on
// a complex type and xs:assertion as a facet of a simple type. The XPath
// evaluator fails open on constructs it cannot evaluate, so these use simple,
// supported predicates.

import "testing"

// xs:assert over complex content: the value of attribute @hi must exceed @lo.
const complexAssertSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="range">
    <xs:complexType>
      <xs:attribute name="lo" type="xs:int"/>
      <xs:attribute name="hi" type="xs:int"/>
      <xs:assert test="@hi &gt; @lo"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`

func TestComplexAssertion(t *testing.T) {
	v := buildSchema(t, complexAssertSchema)
	t.Run("satisfied", func(t *testing.T) {
		assertValid(t, v, `<range lo="1" hi="5"/>`)
	})
	t.Run("violated", func(t *testing.T) {
		assertInvalid(t, v, `<range lo="5" hi="1"/>`, "cvc-assertion")
	})
}

// xs:assertion as a simple-type facet: $value must be longer than two chars.
const simpleAssertSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="word">
    <xs:simpleType>
      <xs:restriction base="xs:string">
        <xs:assertion test="string-length($value) &gt; 2"/>
      </xs:restriction>
    </xs:simpleType>
  </xs:element>
</xs:schema>`

func TestSimpleAssertion(t *testing.T) {
	v := buildSchema(t, simpleAssertSchema)
	t.Run("satisfied", func(t *testing.T) {
		assertValid(t, v, `<word>abcd</word>`)
	})
	t.Run("violated", func(t *testing.T) {
		assertInvalid(t, v, `<word>ab</word>`, "cvc-assertion")
	})
}

// With assertions disabled, the simple-type assertion above is not enforced.
func TestAssertionsDisabled(t *testing.T) {
	v := buildValidator(t, simpleAssertSchema, &assertionsOff)
	assertValid(t, v, `<word>ab</word>`)
}
