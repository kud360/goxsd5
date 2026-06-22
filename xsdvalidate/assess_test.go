package xsdvalidate_test

// Behaviour-first unit tests for the xsdvalidate assessor. Each case is a small,
// hand-written schema + instance pair; the test asserts the observable outcome —
// the validity verdict and, for invalid documents, the distinct cvc-* error ids.
// The schema is built through the real public path (parser.Parse over a temp
// file) and the instance through xmlsrc.Validate, the same entry points the W3C
// instance conformance suite drives, so the tests survive any refactor that
// preserves behaviour.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kud360/goxsd5/parser"
	"github.com/kud360/goxsd5/xsd"
	"github.com/kud360/goxsd5/xsdvalidate"
	"github.com/kud360/goxsd5/xsdvalidate/xmlsrc"
)

// assertionsOff disables XSD 1.1 assertion evaluation (used by buildValidator).
var assertionsOff = xsdvalidate.Options{DisableAssertions: true}

// buildSchema writes the schema text to a temp file and parses it through the
// real parser entry point, failing the test if the schema itself has errors.
func buildSchema(t *testing.T, schema string) *xsdvalidate.Validator {
	return buildValidator(t, schema, nil)
}

// buildValidator is buildSchema with explicit validator options.
func buildValidator(t *testing.T, schema string, opts *xsdvalidate.Options) *xsdvalidate.Validator {
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
	return xsdvalidate.New(schemas, opts)
}

// assess validates the instance text against v and returns the result.
func assess(t *testing.T, v *xsdvalidate.Validator, instance string) *xsdvalidate.Result {
	t.Helper()
	res, err := xmlsrc.Validate(v, strings.NewReader(instance), "instance.xml")
	if err != nil {
		t.Fatalf("instance is not well-formed XML: %v", err)
	}
	return res
}

// assertValid fails unless the instance assesses as schema-valid.
func assertValid(t *testing.T, v *xsdvalidate.Validator, instance string) {
	t.Helper()
	res := assess(t, v, instance)
	if !res.Valid() {
		t.Fatalf("want valid, got invalid: %v", res.Err())
	}
}

// assertInvalid fails unless the instance assesses as invalid AND every id in
// wantIDs appears among the reported cvc-* error ids.
func assertInvalid(t *testing.T, v *xsdvalidate.Validator, instance string, wantIDs ...string) {
	t.Helper()
	res := assess(t, v, instance)
	if res.Valid() {
		t.Fatalf("want invalid, got valid")
	}
	if len(res.Errors()) == 0 {
		t.Fatalf("invalid result reported no errors")
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

// A schema with one element of a simple type and a couple of attributes, used by
// the attribute-assessment cases.
const attrSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="item">
    <xs:complexType>
      <xs:attribute name="id" type="xs:int" use="required"/>
      <xs:attribute name="kind" type="xs:string"/>
      <xs:attribute name="ver" type="xs:string" fixed="v1"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`

func TestAttributeAssessment(t *testing.T) {
	v := buildSchema(t, attrSchema)

	t.Run("valid", func(t *testing.T) {
		assertValid(t, v, `<item id="3" kind="x" ver="v1"/>`)
	})
	t.Run("optional absent", func(t *testing.T) {
		assertValid(t, v, `<item id="3"/>`)
	})
	t.Run("bad value", func(t *testing.T) {
		// id is xs:int; "abc" is not a valid lexical int (fails the type's facet).
		assertInvalid(t, v, `<item id="abc"/>`, "cvc-pattern-valid")
	})
	t.Run("required missing", func(t *testing.T) {
		assertInvalid(t, v, `<item kind="x"/>`, "cvc-complex-type")
	})
	t.Run("undeclared attribute", func(t *testing.T) {
		assertInvalid(t, v, `<item id="3" bogus="y"/>`, "cvc-complex-type")
	})
	t.Run("fixed mismatch", func(t *testing.T) {
		assertInvalid(t, v, `<item id="3" ver="v2"/>`, "cvc-au")
	})
	t.Run("fixed match", func(t *testing.T) {
		assertValid(t, v, `<item id="3" ver="v1"/>`)
	})
}

const nillableSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="n" type="xs:int" nillable="true"/>
  <xs:element name="p" type="xs:int"/>
</xs:schema>`

func TestNillability(t *testing.T) {
	v := buildSchema(t, nillableSchema)
	const xsi = `xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"`

	t.Run("nilled empty", func(t *testing.T) {
		assertValid(t, v, `<n `+xsi+` xsi:nil="true"/>`)
	})
	t.Run("nilled with content", func(t *testing.T) {
		// cvc-elt.3.2.1: a nilled element must be empty.
		assertInvalid(t, v, `<n `+xsi+` xsi:nil="true">5</n>`, "cvc-elt")
	})
	t.Run("nil on non-nillable", func(t *testing.T) {
		// cvc-elt.3.1: xsi:nil on a declaration that is not nillable.
		assertInvalid(t, v, `<p `+xsi+` xsi:nil="true"/>`, "cvc-elt")
	})
	t.Run("nil false carries content", func(t *testing.T) {
		assertValid(t, v, `<n `+xsi+` xsi:nil="false">5</n>`)
	})
}

const atomicLexicalSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="when" type="xs:dateTime"/>
  <xs:element name="flag" type="xs:boolean"/>
  <xs:element name="num" type="xs:double"/>
  <xs:element name="blob" type="xs:hexBinary"/>
  <xs:element name="span" type="xs:duration"/>
</xs:schema>`

// A bad lexical for an unpatterned atomic type fails at the value-space mapping,
// not at any facet. The resulting error must still cite cvc-datatype-valid like
// every other cvc-* failure, rather than surfacing as a raw, id-less error.
func TestAtomicLexicalCarriesDatatypeValid(t *testing.T) {
	v := buildSchema(t, atomicLexicalSchema)

	t.Run("dateTime", func(t *testing.T) {
		assertInvalid(t, v, `<when>garbage</when>`, "cvc-datatype-valid")
	})
	t.Run("boolean", func(t *testing.T) {
		assertInvalid(t, v, `<flag>maybe</flag>`, "cvc-datatype-valid")
	})
	t.Run("double", func(t *testing.T) {
		assertInvalid(t, v, `<num>xyz</num>`, "cvc-datatype-valid")
	})
	t.Run("hexBinary", func(t *testing.T) {
		assertInvalid(t, v, `<blob>zz</blob>`, "cvc-datatype-valid")
	})
	t.Run("duration", func(t *testing.T) {
		assertInvalid(t, v, `<span>P</span>`, "cvc-datatype-valid")
	})
}
