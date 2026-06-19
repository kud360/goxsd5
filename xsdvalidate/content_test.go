package xsdvalidate_test

// Behaviour tests for content-model assessment: element-only / simple / empty
// content, fixed values, xsi:type overrides, and abstractness.

import "testing"

const xsi = `xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"`

const seqSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="root">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="a" type="xs:string"/>
        <xs:element name="b" type="xs:int" minOccurs="0"/>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`

func TestElementContent(t *testing.T) {
	v := buildSchema(t, seqSchema)

	t.Run("valid full", func(t *testing.T) {
		assertValid(t, v, `<root><a>x</a><b>2</b></root>`)
	})
	t.Run("valid minimal", func(t *testing.T) {
		assertValid(t, v, `<root><a>x</a></root>`)
	})
	t.Run("missing required child", func(t *testing.T) {
		assertInvalid(t, v, `<root><b>2</b></root>`, "cvc-particle")
	})
	t.Run("out of order", func(t *testing.T) {
		assertInvalid(t, v, `<root><b>2</b><a>x</a></root>`, "cvc-particle")
	})
	t.Run("unexpected child", func(t *testing.T) {
		assertInvalid(t, v, `<root><a>x</a><z/></root>`, "cvc-particle")
	})
	t.Run("char data in element-only content", func(t *testing.T) {
		// cvc-complex-type.2.3: element-only content admits no non-whitespace text.
		assertInvalid(t, v, `<root>junk<a>x</a></root>`, "cvc-complex-type")
	})
	t.Run("bad child value", func(t *testing.T) {
		assertInvalid(t, v, `<root><a>x</a><b>notint</b></root>`)
	})
}

const fixedSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="f" type="xs:string" fixed="hello"/>
</xs:schema>`

func TestFixedSimpleContent(t *testing.T) {
	v := buildSchema(t, fixedSchema)
	t.Run("match", func(t *testing.T) { assertValid(t, v, `<f>hello</f>`) })
	t.Run("empty takes fixed", func(t *testing.T) { assertValid(t, v, `<f/>`) })
	t.Run("mismatch", func(t *testing.T) {
		// cvc-elt.5.2.2.2: content must equal the fixed value.
		assertInvalid(t, v, `<f>world</f>`, "cvc-elt")
	})
}

const simpleTypeSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="s" type="xs:string"/>
</xs:schema>`

func TestSimpleTypeRejectsChildren(t *testing.T) {
	v := buildSchema(t, simpleTypeSchema)
	t.Run("ok", func(t *testing.T) { assertValid(t, v, `<s>text</s>`) })
	t.Run("element child", func(t *testing.T) {
		// cvc-type.3.1.2: a simple-typed element admits no element children.
		assertInvalid(t, v, `<s><child/></s>`, "cvc-type")
	})
}

const emptySchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="e">
    <xs:complexType>
      <xs:attribute name="a" type="xs:string"/>
    </xs:complexType>
  </xs:element>
</xs:schema>`

func TestEmptyContent(t *testing.T) {
	v := buildSchema(t, emptySchema)
	t.Run("empty", func(t *testing.T) { assertValid(t, v, `<e a="x"/>`) })
	t.Run("with text", func(t *testing.T) {
		// cvc-complex-type.2.1: empty content admits no character data.
		assertInvalid(t, v, `<e>text</e>`, "cvc-complex-type")
	})
	t.Run("with child", func(t *testing.T) {
		// An empty content type whose particle cannot match the child: cvc-particle.
		assertInvalid(t, v, `<e><c/></e>`, "cvc-particle")
	})
}

// xsi:type override and derivation. base/derived are both global complex types;
// `derived` extends `base`, `unrelated` does not derive from it.
const xsiTypeSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           xmlns:t="urn:t" targetNamespace="urn:t">
  <xs:complexType name="base">
    <xs:sequence><xs:element name="a" type="xs:string"/></xs:sequence>
  </xs:complexType>
  <xs:complexType name="derived">
    <xs:complexContent>
      <xs:extension base="t:base">
        <xs:sequence><xs:element name="b" type="xs:string"/></xs:sequence>
      </xs:extension>
    </xs:complexContent>
  </xs:complexType>
  <xs:complexType name="unrelated">
    <xs:sequence><xs:element name="z" type="xs:string"/></xs:sequence>
  </xs:complexType>
  <xs:element name="root" type="t:base"/>
</xs:schema>`

func TestXSIType(t *testing.T) {
	v := buildSchema(t, xsiTypeSchema)
	// root is in urn:t; its children a/b/z are local (unqualified) declarations.
	const hdr = `xmlns:t="urn:t" ` + xsi

	t.Run("derived override valid", func(t *testing.T) {
		assertValid(t, v, `<t:root `+hdr+` xsi:type="t:derived"><a>x</a><b>y</b></t:root>`)
	})
	t.Run("not derived", func(t *testing.T) {
		// cvc-elt.4.3: xsi:type must be validly derived from the declared type.
		assertInvalid(t, v, `<t:root `+hdr+` xsi:type="t:unrelated"><z>x</z></t:root>`, "cvc-elt")
	})
	t.Run("unknown type", func(t *testing.T) {
		assertInvalid(t, v, `<t:root `+hdr+` xsi:type="t:nope"/>`, "cvc-elt")
	})
}

const abstractSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="ae" type="xs:string" abstract="true"/>
  <xs:complexType name="abs" abstract="true">
    <xs:sequence><xs:element name="x" type="xs:string"/></xs:sequence>
  </xs:complexType>
  <xs:element name="ce" type="abs"/>
</xs:schema>`

func TestAbstract(t *testing.T) {
	v := buildSchema(t, abstractSchema)
	t.Run("abstract element", func(t *testing.T) {
		// cvc-elt.2: an abstract element declaration cannot appear in an instance.
		assertInvalid(t, v, `<ae>x</ae>`, "cvc-elt")
	})
	t.Run("abstract type", func(t *testing.T) {
		// cvc-complex-type.1: an abstract type cannot be instantiated.
		assertInvalid(t, v, `<ce><x>y</x></ce>`, "cvc-complex-type")
	})
}

const noRootSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="known" type="xs:string"/>
</xs:schema>`

func TestNoGoverningDeclaration(t *testing.T) {
	v := buildSchema(t, noRootSchema)
	// cvc-elt: there is no global element declaration for the root.
	assertInvalid(t, v, `<unknown/>`, "cvc-elt")
}
