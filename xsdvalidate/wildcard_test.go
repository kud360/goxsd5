package xsdvalidate_test

// Behaviour tests for wildcard / open-content matching and the engine-owned
// ID/IDREF referential rules (cvc-id), which are not per-datatype.

import "testing"

// An element-only content model followed by an xs:any wildcard with strict
// processing. A wildcard-matched element with no declaration is a cvc-wildcard
// violation under strict processing; under skip it is accepted unassessed.
const wildcardStrictSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="known" type="xs:string"/>
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="a" type="xs:string"/>
        <xs:any namespace="##any" processContents="strict" minOccurs="0"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`

func TestWildcardStrict(t *testing.T) {
	v := buildSchema(t, wildcardStrictSchema)
	t.Run("declared element accepted", func(t *testing.T) {
		assertValid(t, v, `<root><a>x</a><known>y</known></root>`)
	})
	t.Run("undeclared rejected under strict", func(t *testing.T) {
		// cvc-wildcard: no declaration for the strict-wildcard-matched element.
		assertInvalid(t, v, `<root><a>x</a><nodecl/></root>`, "cvc-wildcard")
	})
}

const wildcardSkipSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="a" type="xs:string"/>
        <xs:any namespace="##any" processContents="skip" minOccurs="0"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`

func TestWildcardSkip(t *testing.T) {
	v := buildSchema(t, wildcardSkipSchema)
	// A skip wildcard accepts any element subtree unassessed.
	assertValid(t, v, `<root><a>x</a><anything><deeper/></anything></root>`)
}

const wildcardLaxSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="num" type="xs:int"/>
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:any namespace="##any" processContents="lax" minOccurs="0" maxOccurs="unbounded"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`

func TestWildcardLax(t *testing.T) {
	v := buildSchema(t, wildcardLaxSchema)
	t.Run("undeclared accepted unassessed", func(t *testing.T) {
		assertValid(t, v, `<root><whatever/></root>`)
	})
	t.Run("declared element is assessed", func(t *testing.T) {
		// lax: a matched element that HAS a declaration is validated against it.
		assertInvalid(t, v, `<root><num>notanint</num></root>`)
	})
	t.Run("declared element valid", func(t *testing.T) {
		assertValid(t, v, `<root><num>7</num></root>`)
	})
}

// Open content: a complex type that interleaves an open-content wildcard with the
// declared particle (XSD 1.1 §3.4.2).
const openContentSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:openContent mode="interleave">
        <xs:any namespace="##any" processContents="skip"/>
      </xs:openContent>
      <xs:sequence>
        <xs:element name="a" type="xs:string"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`

func TestOpenContent(t *testing.T) {
	v := buildSchema(t, openContentSchema)
	t.Run("declared only", func(t *testing.T) {
		assertValid(t, v, `<root><a>x</a></root>`)
	})
	t.Run("open content interleaved", func(t *testing.T) {
		assertValid(t, v, `<root><extra/><a>x</a><more/></root>`)
	})
}

// cvc-id: xs:ID uniqueness and xs:IDREF referential validity.
const idSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="n" maxOccurs="unbounded">
          <xs:complexType>
            <xs:attribute name="id" type="xs:ID"/>
            <xs:attribute name="ref" type="xs:IDREF"/>
          </xs:complexType>
        </xs:element>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`

func TestIDAndIDREF(t *testing.T) {
	v := buildSchema(t, idSchema)
	t.Run("valid", func(t *testing.T) {
		assertValid(t, v, `<root><n id="x"/><n ref="x"/></root>`)
	})
	t.Run("duplicate id", func(t *testing.T) {
		assertInvalid(t, v, `<root><n id="x"/><n id="x"/></root>`, "cvc-id")
	})
	t.Run("dangling idref", func(t *testing.T) {
		assertInvalid(t, v, `<root><n id="x"/><n ref="y"/></root>`, "cvc-id")
	})
}
