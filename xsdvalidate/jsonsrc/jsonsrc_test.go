package jsonsrc_test

// Behaviour tests for the JSON-to-infoset adapter. They build a small schema
// through the real parser entry point and drive JSON instances through
// jsonsrc.Validate / jsonsrc.NewElement, asserting the observable outcome: the
// validity verdict, the cvc-* error ids, and the infoset shape the adapter
// exposes (schema-aware attribute-vs-element classification, member order,
// arrays as repeated children, null as xsi:nil, and scalar-shorthand text).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kud360/goxsd5/parser"
	"github.com/kud360/goxsd5/xsd"
	"github.com/kud360/goxsd5/xsdvalidate"
	"github.com/kud360/goxsd5/xsdvalidate/jsonsrc"
)

func buildValidator(t *testing.T, schema string) *xsdvalidate.Validator {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "schema.xsd")
	if err := os.WriteFile(path, []byte(schema), 0o644); err != nil {
		t.Fatalf("write schema: %v", err)
	}
	schemas, err := parser.Parse(path, nil)
	if err != nil {
		t.Fatalf("schema unexpectedly has errors: %v", err)
	}
	return xsdvalidate.New(schemas, nil)
}

func validate(t *testing.T, v *xsdvalidate.Validator, doc string) *xsdvalidate.Result {
	t.Helper()
	res, _, err := jsonsrc.Validate(v, strings.NewReader(doc), "doc.json")
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	return res
}

func assertValid(t *testing.T, v *xsdvalidate.Validator, doc string) {
	t.Helper()
	res := validate(t, v, doc)
	if !res.Valid() {
		t.Fatalf("want valid, got invalid: %v", res.Err())
	}
}

func assertInvalid(t *testing.T, v *xsdvalidate.Validator, doc string, wantIDs ...string) {
	t.Helper()
	res := validate(t, v, doc)
	if res.Valid() {
		t.Fatalf("want invalid, got valid")
	}
	got := xsd.RefIDs(res.Err())
	for _, want := range wantIDs {
		if !contains(got, want) {
			t.Fatalf("want error id %q among %v\n  errors: %v", want, got, res.Err())
		}
	}
}

func contains(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}

// poSchema mixes attributes and child elements, an array-valued repeating child,
// a nillable element, and a simple-content leaf, exercising every mapping rule.
const poSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="order">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="customer" type="xs:string"/>
        <xs:element name="item" maxOccurs="unbounded">
          <xs:complexType>
            <xs:sequence>
              <xs:element name="name" type="xs:string"/>
              <xs:element name="note" type="xs:string" nillable="true" minOccurs="0"/>
            </xs:sequence>
            <xs:attribute name="sku" type="xs:string" use="required"/>
            <xs:attribute name="qty" type="xs:int"/>
          </xs:complexType>
        </xs:element>
      </xs:sequence>
      <xs:attribute name="id" type="xs:int" use="required"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`

func TestValidOrder(t *testing.T) {
	v := buildValidator(t, poSchema)
	// id is an attribute; item is an array of repeated children; qty stays "3"
	// as an int; customer is scalar-shorthand text content.
	assertValid(t, v, `{"order":{
		"id": 7,
		"customer": "Ada",
		"item": [
			{"sku": "A", "qty": 3, "name": "widget"},
			{"sku": "B", "name": "gadget"}
		]
	}}`)
}

func TestAttributeClassifiedByName(t *testing.T) {
	v := buildValidator(t, poSchema)
	// A missing required attribute must be reported, proving "id" was routed to
	// the attribute use (not treated as a stray child element).
	assertInvalid(t, v, `{"order":{"customer":"Ada","item":[{"sku":"A","name":"x"}]}}`,
		"cvc-complex-type")
}

func TestNumericTokenPreserved(t *testing.T) {
	v := buildValidator(t, `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="v" type="xs:integer"/>
</xs:schema>`)
	// 5.0 is not a valid xs:integer lexical; the adapter must preserve the token
	// text rather than renormalizing it to "5" (which would validate).
	assertInvalid(t, v, `{"v": 5.0}`, "cvc-pattern-valid")
	assertValid(t, v, `{"v": 5}`)
}

func TestNullIsXSINil(t *testing.T) {
	v := buildValidator(t, poSchema)
	// note is nillable: JSON null maps to xsi:nil="true" and validates.
	assertValid(t, v, `{"order":{
		"id": 1, "customer": "Ada",
		"item": [{"sku": "A", "name": "x", "note": null}]
	}}`)
}

func TestNullOnNonNillableFails(t *testing.T) {
	v := buildValidator(t, `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence><xs:element name="leaf" type="xs:string"/></xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`)
	assertInvalid(t, v, `{"root":{"leaf":null}}`, "cvc-elt")
}

func TestMemberOrderPreserved(t *testing.T) {
	// A sequence customer-then-item must fail if the JSON supplies item first,
	// proving object member order reaches the sequence content model.
	v := buildValidator(t, poSchema)
	assertInvalid(t, v, `{"order":{
		"id": 1,
		"item": [{"sku":"A","name":"x"}],
		"customer": "Ada"
	}}`, "cvc-particle")
}

func TestScalarShorthandText(t *testing.T) {
	v := buildValidator(t, `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="size" type="xs:int"/>
</xs:schema>`)
	assertValid(t, v, `{"size": 10}`)
	assertInvalid(t, v, `{"size": "ten"}`, "cvc-pattern-valid")
}

func TestUnknownRootElement(t *testing.T) {
	v := buildValidator(t, poSchema)
	if _, _, err := jsonsrc.Validate(v, strings.NewReader(`{"nope":{}}`), "doc.json"); err == nil {
		t.Fatal("want an error for an unknown root element")
	}
}

func TestMalformedJSON(t *testing.T) {
	v := buildValidator(t, poSchema)
	if _, _, err := jsonsrc.Validate(v, strings.NewReader(`{"order": `), "doc.json"); err == nil {
		t.Fatal("want a decode error for malformed JSON")
	}
}

func TestTrailingContentRejected(t *testing.T) {
	v := buildValidator(t, poSchema)
	if _, _, err := jsonsrc.Validate(v, strings.NewReader(`{"order":{}} {}`), "doc.json"); err == nil {
		t.Fatal("want an error for trailing content after the top-level value")
	}
}

func TestTopLevelMustBeSingleMemberObject(t *testing.T) {
	v := buildValidator(t, poSchema)
	if _, _, err := jsonsrc.Validate(v, strings.NewReader(`{"a":{},"b":{}}`), "doc.json"); err == nil {
		t.Fatal("want an error for a multi-member top-level object")
	}
	if _, _, err := jsonsrc.Validate(v, strings.NewReader(`[1,2]`), "doc.json"); err == nil {
		t.Fatal("want an error for a non-object top-level value")
	}
}

// collisionSchema declares "note" as BOTH an attribute use and a child element,
// exercising the element-wins-plus-warning collision rule.
const collisionSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="rec">
    <xs:complexType>
      <xs:sequence><xs:element name="note" type="xs:string"/></xs:sequence>
      <xs:attribute name="note" type="xs:string"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`

func TestCollisionElementWinsWithWarning(t *testing.T) {
	v := buildValidator(t, collisionSchema)
	res, warnings, err := jsonsrc.Validate(v, strings.NewReader(`{"rec":{"note":"hi"}}`), "doc.json")
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !res.Valid() {
		t.Fatalf("want valid (note routed to the child element), got: %v", res.Err())
	}
	if len(warnings) == 0 {
		t.Fatal("want a collision warning, got none")
	}
	if !strings.Contains(warnings[0].Msg, "note") {
		t.Fatalf("warning %q should name the colliding key", warnings[0].Msg)
	}
}

// xsiTypeSchema has a base type and a derived type in no namespace, plus an
// element of the base type, so a $type directive can select the derived type.
const xsiTypeSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:complexType name="Base">
    <xs:sequence><xs:element name="a" type="xs:string"/></xs:sequence>
  </xs:complexType>
  <xs:complexType name="Derived">
    <xs:complexContent>
      <xs:extension base="Base">
        <xs:sequence><xs:element name="b" type="xs:string"/></xs:sequence>
      </xs:extension>
    </xs:complexContent>
  </xs:complexType>
  <xs:element name="thing" type="Base"/>
</xs:schema>`

func TestTypeDirectiveSelectsDerivedType(t *testing.T) {
	v := buildValidator(t, xsiTypeSchema)
	// $type sets xsi:type to the derived type, which adds child "b"; the document
	// supplies both children and must validate against the derived content model.
	assertValid(t, v, `{"thing":{"$type":"Derived","a":"x","b":"y"}}`)
	// Without the override the base type governs and "b" is not allowed.
	assertInvalid(t, v, `{"thing":{"a":"x","b":"y"}}`, "cvc-particle")
}

func TestTypeDirectiveUnknownTypeFails(t *testing.T) {
	v := buildValidator(t, xsiTypeSchema)
	// $type names a type that does not exist: the engine rejects the xsi:type.
	assertInvalid(t, v, `{"thing":{"$type":"Nonexistent","a":"x"}}`, "cvc-elt")
}

func TestTypeDirectiveNonDerivedTypeFails(t *testing.T) {
	v := buildValidator(t, xsiTypeSchema)
	// xs:int is a real type but not derived from Base, so cvc-elt.4.3 rejects it.
	assertInvalid(t, v, `{"thing":{"$type":"xs:int","a":"x"}}`, "cvc-elt")
}

// nsTypeSchema puts a derived type in a target namespace so an xsi:type QName
// must be prefixed to name it. Derived restricts Base to the same single child,
// so selecting it changes only the type QName resolution, not the content shape:
// this isolates whether a $xmlns prefix binding reaches the engine's Lookup.
const nsTypeSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
    xmlns:tns="urn:t" targetNamespace="urn:t">
  <xs:complexType name="Base">
    <xs:sequence><xs:element name="a" type="xs:string"/></xs:sequence>
  </xs:complexType>
  <xs:complexType name="Derived">
    <xs:complexContent>
      <xs:restriction base="tns:Base">
        <xs:sequence><xs:element name="a" type="xs:string"/></xs:sequence>
      </xs:restriction>
    </xs:complexContent>
  </xs:complexType>
  <xs:element name="thing" type="tns:Base"/>
</xs:schema>`

func TestXMLNSBindingResolvesTypeQName(t *testing.T) {
	v := buildValidator(t, nsTypeSchema)
	// $xmlns binds prefix "p" to the schema's target namespace; the $type QName
	// "p:Derived" then resolves through the engine's in-scope Lookup, proving the
	// $xmlns binding reaches resolveQNameInScope.
	assertValid(t, v, `{"thing":{
		"$xmlns": {"p": "urn:t"},
		"$type": "p:Derived",
		"a": "x"
	}}`)
	// An unbound prefix must fail to resolve, so the override is rejected.
	assertInvalid(t, v, `{"thing":{"$type":"q:Derived","a":"x"}}`, "cvc-elt")
}

func TestSchemaAccessorAmbiguousLocal(t *testing.T) {
	// Two globals share a local name across namespaces: an unprefixed root key
	// is ambiguous and must not resolve.
	v := buildValidator(t, `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
    xmlns:tns="urn:a" targetNamespace="urn:a">
  <xs:element name="dup" type="xs:string"/>
</xs:schema>`)
	sc := v.Schema()
	if d := sc.ElementByLocal("dup"); d == nil {
		t.Fatal("a single-namespace global should resolve by local name")
	}
	if d := sc.ElementByName(xsd.QName{Namespace: "urn:a", Local: "dup"}); d == nil {
		t.Fatal("ElementByName should resolve the namespaced global")
	}
}
