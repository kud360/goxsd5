package xsdvalidate_test

// Fuzz coverage for the instance-validation engine — the primary surface for
// processing arbitrary user-supplied XML. The assessor dispatches wildcard
// content models, evaluates identity constraints and XPath assertions, and
// builds the PSVI-lite Result, all over untrusted input; a panic on any path is
// a denial-of-service vector for callers of xmlsrc.Validate. The fuzz target
// drives that whole pipeline through the same public entry points the W3C
// instance conformance suite uses.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kud360/goxsd5/parser"
	"github.com/kud360/goxsd5/xsd"
	"github.com/kud360/goxsd5/xsdvalidate"
	"github.com/kud360/goxsd5/xsdvalidate/xmlsrc"
)

// kitchenSinkSchema exercises a broad slice of the assessor: complex content
// (sequence/choice/all), repetition, attributes and attribute use, simple-type
// derivation (restriction facets, list, union), a wildcard, an identity
// constraint (key/keyref), and an xs:assert — so a fuzzed body reaches the
// content-model, IDC, assertion, and PSVI code paths.
const kitchenSinkSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           xmlns:t="urn:t"
           xmlns="urn:t"
           targetNamespace="urn:t"
           elementFormDefault="qualified">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="item" maxOccurs="unbounded">
          <xs:complexType>
            <xs:sequence>
              <xs:element name="name" type="xs:string"/>
              <xs:element name="qty" type="t:countType" minOccurs="0"/>
              <xs:choice minOccurs="0">
                <xs:element name="tag" type="t:tagList"/>
                <xs:element name="code" type="t:codeUnion"/>
              </xs:choice>
              <xs:any namespace="##other" processContents="lax"
                      minOccurs="0" maxOccurs="unbounded"/>
            </xs:sequence>
            <xs:attribute name="id" type="xs:ID" use="required"/>
            <xs:attribute name="ref" type="xs:IDREF"/>
            <xs:assert test="@id != ''"/>
          </xs:complexType>
        </xs:element>
        <xs:element name="when" type="xs:dateTime" minOccurs="0"/>
      </xs:sequence>
      <xs:attribute name="kind" type="xs:string"/>
    </xs:complexType>
    <xs:key name="itemKey">
      <xs:selector xpath="item"/>
      <xs:field xpath="@id"/>
    </xs:key>
    <xs:keyref name="itemRef" refer="itemKey">
      <xs:selector xpath="item"/>
      <xs:field xpath="@ref"/>
    </xs:keyref>
  </xs:element>

  <xs:simpleType name="countType">
    <xs:restriction base="xs:int">
      <xs:minInclusive value="0"/>
      <xs:maxInclusive value="1000"/>
    </xs:restriction>
  </xs:simpleType>

  <xs:simpleType name="tagList">
    <xs:list itemType="xs:NCName"/>
  </xs:simpleType>

  <xs:simpleType name="codeUnion">
    <xs:union memberTypes="xs:int xs:NCName"/>
  </xs:simpleType>
</xs:schema>`

// instanceSeeds are XML bodies (valid and invalid against kitchenSinkSchema)
// that steer the fuzzer onto each major content-model, IDC, and assertion path.
var instanceSeeds = []string{
	// well-formed and valid
	`<root xmlns="urn:t" kind="order"><item id="a"><name>x</name></item></root>`,
	`<root xmlns="urn:t"><item id="a"><name>x</name><qty>5</qty></item></root>`,
	`<root xmlns="urn:t"><item id="a" ref="b"><name>x</name></item><item id="b"><name>y</name></item></root>`,
	`<root xmlns="urn:t"><item id="a"><name>n</name><tag>one two three</tag></item></root>`,
	`<root xmlns="urn:t"><item id="a"><name>n</name><code>42</code></item></root>`,
	`<root xmlns="urn:t"><item id="a"><name>n</name></item><when>2001-10-26T21:32:52</when></root>`,
	// schema-invalid bodies the assessor must report, not panic on
	`<root xmlns="urn:t"><item><name>x</name></item></root>`,                            // missing required @id
	`<root xmlns="urn:t"><item id="a"><name>x</name><qty>-9</qty></item></root>`,        // facet violation
	`<root xmlns="urn:t"><item id="a" ref="ghost"><name>x</name></item></root>`,         // dangling keyref
	`<root xmlns="urn:t"><item id="a"><name>x</name><qty>nan</qty></item></root>`,       // bad lexical (int)
	`<root xmlns="urn:t"><item id="a"><name>n</name></item><when>garbage</when></root>`, // bad unpatterned lexical (dateTime -> raw error)
	`<root xmlns="urn:t"><item id="a"/><item id="a"><name>y</name></item></root>`,       // duplicate key
	`<wrong xmlns="urn:t"/>`, // unexpected root
	`<root xmlns="urn:t"/>`,  // empty (violates the sequence)
	// not even well-formed XML — Validate returns a parse error, not a Result
	``, `<`, `<root>`, `</root>`, `<root><item id="a"></root>`,
}

// FuzzValidateInstance feeds arbitrary XML bodies through xmlsrc.Validate
// against a fixed kitchen-sink schema. The Validator is compiled once; each
// fuzz call only parses and assesses the instance. Invariants:
//   - no panic anywhere in the parse/assess pipeline (the harness fails on one);
//   - when the body parses as XML, the Result is internally consistent —
//     Valid() iff Err() is nil, and Errors() agrees with Valid();
//   - every reported error is an *xsd.Error carrying a non-empty SpecRef.
//
// Bad lexicals for unpatterned atomic types (e.g. xs:dateTime "garbage",
// xs:boolean, xs:double, or hexBinary/base64Binary/duration/QName at type
// level) yield plain errors from the builtin parse funcs, but SimpleType's
// value-space boundary wraps them in cvc-datatype-valid, so they reach the
// Result as structured errors like every other cvc-* failure.
func FuzzValidateInstance(f *testing.F) {
	v := buildFuzzValidator(f)
	for _, s := range instanceSeeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, body string) {
		res, err := xmlsrc.Validate(v, strings.NewReader(body), "fuzz.xml")
		if err != nil {
			// Not well-formed XML: a nil Result with a non-nil error is the
			// documented outcome, not a contract violation.
			if res != nil {
				t.Fatalf("Validate(%q): non-nil Result alongside parse error %v", body, err)
			}
			return
		}
		if res == nil {
			t.Fatalf("Validate(%q): nil Result and nil error", body)
		}
		if res.Valid() != (res.Err() == nil) {
			t.Fatalf("Validate(%q): Valid()=%v but Err()=%v", body, res.Valid(), res.Err())
		}
		errs := res.Errors()
		if res.Valid() != (len(errs) == 0) {
			t.Fatalf("Validate(%q): Valid()=%v but Errors() has %d entries", body, res.Valid(), len(errs))
		}
		for i, e := range errs {
			// Every instance error must be a structured *xsd.Error carrying a
			// SpecRef — including unpatterned atomic lexicals, now wrapped in
			// cvc-datatype-valid at the value-space boundary.
			var xe *xsd.Error
			if !errors.As(e, &xe) {
				t.Fatalf("Validate(%q): error %d is not an *xsd.Error: %v", body, i, e)
			}
			if xe.Ref.IsZero() {
				t.Fatalf("Validate(%q): error %d is an *xsd.Error with an empty SpecRef: %v", body, i, xe)
			}
		}
	})
}

// buildFuzzValidator compiles kitchenSinkSchema through the real parser entry
// point once, failing the fuzz target if the fixture schema itself is invalid.
func buildFuzzValidator(f *testing.F) *xsdvalidate.Validator {
	f.Helper()
	dir := f.TempDir()
	path := filepath.Join(dir, "kitchen-sink.xsd")
	if err := os.WriteFile(path, []byte(kitchenSinkSchema), 0o644); err != nil {
		f.Fatalf("write schema: %v", err)
	}
	schemas, err := parser.Parse(path, nil)
	if err != nil {
		f.Fatalf("kitchen-sink schema unexpectedly has errors: %v", err)
	}
	return xsdvalidate.New(schemas, nil)
}
