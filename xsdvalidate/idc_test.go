package xsdvalidate_test

// Behaviour tests for identity-constraint assessment (idc.go): xs:unique,
// xs:key, xs:keyref — selector/field selection, duplicate detection, missing
// key fields, and keyref resolution against its key. All violations carry the
// cvc-identity-constraint id.

import "testing"

const uniqueSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="cat">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="item" maxOccurs="unbounded">
          <xs:complexType>
            <xs:attribute name="id" type="xs:string"/>
          </xs:complexType>
        </xs:element>
      </xs:sequence>
    </xs:complexType>
    <xs:unique name="uId">
      <xs:selector xpath="item"/>
      <xs:field xpath="@id"/>
    </xs:unique>
  </xs:element>
</xs:schema>`

func TestUnique(t *testing.T) {
	v := buildSchema(t, uniqueSchema)
	t.Run("distinct ok", func(t *testing.T) {
		assertValid(t, v, `<cat><item id="a"/><item id="b"/></cat>`)
	})
	t.Run("missing field ok", func(t *testing.T) {
		// xs:unique: an absent field just omits the tuple, not an error.
		assertValid(t, v, `<cat><item id="a"/><item/></cat>`)
	})
	t.Run("duplicate", func(t *testing.T) {
		assertInvalid(t, v, `<cat><item id="a"/><item id="a"/></cat>`, "cvc-identity-constraint")
	})
}

const keySchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="cat">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="item" maxOccurs="unbounded">
          <xs:complexType>
            <xs:attribute name="id" type="xs:string"/>
          </xs:complexType>
        </xs:element>
        <xs:element name="ref" minOccurs="0" maxOccurs="unbounded">
          <xs:complexType>
            <xs:attribute name="to" type="xs:string"/>
          </xs:complexType>
        </xs:element>
      </xs:sequence>
    </xs:complexType>
    <xs:key name="kId">
      <xs:selector xpath="item"/>
      <xs:field xpath="@id"/>
    </xs:key>
    <xs:keyref name="kRef" refer="kId">
      <xs:selector xpath="ref"/>
      <xs:field xpath="@to"/>
    </xs:keyref>
  </xs:element>
</xs:schema>`

func TestKeyAndKeyref(t *testing.T) {
	v := buildSchema(t, keySchema)
	t.Run("valid", func(t *testing.T) {
		assertValid(t, v, `<cat><item id="a"/><item id="b"/><ref to="a"/></cat>`)
	})
	t.Run("duplicate key", func(t *testing.T) {
		assertInvalid(t, v, `<cat><item id="a"/><item id="a"/></cat>`, "cvc-identity-constraint")
	})
	t.Run("key field missing", func(t *testing.T) {
		// cvc-identity-constraint.4.2.1: a key tuple must have every field.
		assertInvalid(t, v, `<cat><item/></cat>`, "cvc-identity-constraint")
	})
	t.Run("keyref unresolved", func(t *testing.T) {
		// cvc-identity-constraint.4.3: keyref value with no matching key.
		assertInvalid(t, v, `<cat><item id="a"/><ref to="zzz"/></cat>`, "cvc-identity-constraint")
	})
	t.Run("keyref resolves to value not lexical", func(t *testing.T) {
		// Comparison is in the value space; here both are plain strings, so the
		// keyref to an existing key resolves.
		assertValid(t, v, `<cat><item id="a"/><item id="b"/><ref to="b"/></cat>`)
	})
}

// A key whose field is typed numerically: 5 and 5.0 denote the same value, so a
// keyref written 5.0 resolves to a key written 5 (value-space comparison).
const numKeySchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="cat">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="item" maxOccurs="unbounded">
          <xs:complexType>
            <xs:attribute name="id" type="xs:decimal"/>
          </xs:complexType>
        </xs:element>
        <xs:element name="ref" minOccurs="0">
          <xs:complexType>
            <xs:attribute name="to" type="xs:decimal"/>
          </xs:complexType>
        </xs:element>
      </xs:sequence>
    </xs:complexType>
    <xs:key name="kId">
      <xs:selector xpath="item"/>
      <xs:field xpath="@id"/>
    </xs:key>
    <xs:keyref name="kRef" refer="kId">
      <xs:selector xpath="ref"/>
      <xs:field xpath="@to"/>
    </xs:keyref>
  </xs:element>
</xs:schema>`

// A unique constraint whose selector uses the .//descendant axis, reaching items
// nested at any depth (exercises the self-or-descendant selection path).
const descendantSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="cat">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="group" maxOccurs="unbounded">
          <xs:complexType>
            <xs:sequence>
              <xs:element name="item" maxOccurs="unbounded">
                <xs:complexType>
                  <xs:attribute name="id" type="xs:string"/>
                </xs:complexType>
              </xs:element>
            </xs:sequence>
          </xs:complexType>
        </xs:element>
      </xs:sequence>
    </xs:complexType>
    <xs:unique name="uId">
      <xs:selector xpath=".//item"/>
      <xs:field xpath="@id"/>
    </xs:unique>
  </xs:element>
</xs:schema>`

func TestDescendantSelector(t *testing.T) {
	v := buildSchema(t, descendantSchema)
	t.Run("distinct across groups", func(t *testing.T) {
		assertValid(t, v, `<cat><group><item id="a"/></group><group><item id="b"/></group></cat>`)
	})
	t.Run("duplicate across groups", func(t *testing.T) {
		// The .//item selector reaches both groups, so the clash is detected.
		assertInvalid(t, v, `<cat><group><item id="a"/></group><group><item id="a"/></group></cat>`, "cvc-identity-constraint")
	})
}

// crossScopeSchema declares the key on the root element <cat> (its table is built
// over the whole cat subtree), while the keyref is declared on a nested <bag>
// element. The keyref's refer target therefore lives on an ancestor element —
// the cross-scope case (XSD 1.1 §3.11.2): the keyref must resolve against the
// root's key table, which only exists once the root has been assessed.
const crossScopeSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="cat">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="item" maxOccurs="unbounded">
          <xs:complexType>
            <xs:attribute name="id" type="xs:string"/>
          </xs:complexType>
        </xs:element>
        <xs:element name="bag">
          <xs:complexType>
            <xs:sequence>
              <xs:element name="ref" minOccurs="0" maxOccurs="unbounded">
                <xs:complexType>
                  <xs:attribute name="to" type="xs:string"/>
                </xs:complexType>
              </xs:element>
            </xs:sequence>
          </xs:complexType>
          <xs:keyref name="kRef" refer="kId">
            <xs:selector xpath="ref"/>
            <xs:field xpath="@to"/>
          </xs:keyref>
        </xs:element>
      </xs:sequence>
    </xs:complexType>
    <xs:key name="kId">
      <xs:selector xpath="item"/>
      <xs:field xpath="@id"/>
    </xs:key>
  </xs:element>
</xs:schema>`

func TestCrossScopeKeyref(t *testing.T) {
	v := buildSchema(t, crossScopeSchema)
	t.Run("valid: keyref resolves to ancestor key", func(t *testing.T) {
		assertValid(t, v, `<cat><item id="a"/><item id="b"/><bag><ref to="a"/></bag></cat>`)
	})
	t.Run("invalid: keyref has no matching ancestor key", func(t *testing.T) {
		// cvc-identity-constraint.4.3: the cross-scope keyref is now actively
		// checked rather than fail-open, so the unresolved value is reported.
		assertInvalid(t, v, `<cat><item id="a"/><bag><ref to="zzz"/></bag></cat>`, "cvc-identity-constraint")
	})
}

func TestKeyValueSpaceComparison(t *testing.T) {
	v := buildSchema(t, numKeySchema)
	t.Run("5 and 5.0 are the same key", func(t *testing.T) {
		// Duplicate detected in the value space.
		assertInvalid(t, v, `<cat><item id="5"/><item id="5.0"/></cat>`, "cvc-identity-constraint")
	})
	t.Run("keyref 5.0 resolves key 5", func(t *testing.T) {
		assertValid(t, v, `<cat><item id="5"/><ref to="5.0"/></cat>`)
	})
}
