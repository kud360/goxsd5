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

// upwardScopeSchema declares the keyref on the root <cat> while the referred key
// is declared on a DESCENDANT <reg> element. Per XSD 1.1 §3.11.4/§3.11.5 the
// descendant's node table propagates upward into <cat>'s subtree, so the keyref
// at <cat> resolves against the key sourced in <reg> — the legitimate
// upward-propagation "cross-scope" case. (The inverted direction — key on the
// ancestor, keyref on a descendant — is illegal under clause 4.3 and must FAIL.)
const upwardScopeSchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="cat">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="reg">
          <xs:complexType>
            <xs:sequence>
              <xs:element name="item" maxOccurs="unbounded">
                <xs:complexType>
                  <xs:attribute name="id" type="xs:string"/>
                </xs:complexType>
              </xs:element>
            </xs:sequence>
          </xs:complexType>
          <xs:key name="kId">
            <xs:selector xpath="item"/>
            <xs:field xpath="@id"/>
          </xs:key>
        </xs:element>
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
</xs:schema>`

func TestUpwardScopeKeyref(t *testing.T) {
	v := buildSchema(t, upwardScopeSchema)
	t.Run("valid: keyref resolves to key in descendant subtree", func(t *testing.T) {
		// The <reg>/<item> key table propagates up to <cat>, where the keyref reads it.
		assertValid(t, v, `<cat><reg><item id="a"/><item id="b"/></reg><ref to="a"/></cat>`)
	})
	t.Run("invalid: keyref value with no matching key in scope", func(t *testing.T) {
		// cvc-identity-constraint.4.3: no in-scope key tuple matches "zzz".
		assertInvalid(t, v, `<cat><reg><item id="a"/></reg><ref to="zzz"/></cat>`, "cvc-identity-constraint")
	})
}

// invertedScopeSchema is the ILLEGAL inverted direction: the key is declared on
// the ancestor <cat> and the keyref on a descendant <bag>. The ancestor's node
// table does NOT reach the descendant (it propagates upward, not downward), so
// per cvc-identity-constraint clause 4.3 the keyref finds no in-scope key and
// must FAIL even when a same-valued key exists on the ancestor.
const invertedScopeSchema = `<?xml version="1.0"?>
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
        <!-- keyref refers to kId on the ancestor <cat>; the parser resolves the
             refer QName by NCName across the schema, so the binding exists. -->
      </xs:sequence>
    </xs:complexType>
    <xs:key name="kId">
      <xs:selector xpath="item"/>
      <xs:field xpath="@id"/>
    </xs:key>
  </xs:element>
</xs:schema>`

func TestInvertedScopeKeyrefFails(t *testing.T) {
	v := buildSchema(t, invertedScopeSchema)
	// The ancestor key's node table does not propagate down to the <bag> keyref,
	// so even though item id="a" exists on the ancestor, the keyref to="a" finds
	// no key within its own subtree → clause 4.3 no-match.
	assertInvalid(t, v, `<cat><item id="a"/><bag><ref to="a"/></bag></cat>`, "cvc-identity-constraint")
}

// repeatedSiblingScope declares a key+keyref on a repeating <section>; each
// section instance has its own subtree-scoped node table. A keyref in one section
// must resolve only against the key in its OWN section, not a sibling's. This is
// the case the per-(constraint, instance) scoping fixes: the global-slot model
// would let a ref match any section's key.
const repeatedSiblingScope = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="doc">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="section" maxOccurs="unbounded">
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
          <xs:key name="sId">
            <xs:selector xpath="item"/>
            <xs:field xpath="@id"/>
          </xs:key>
          <xs:keyref name="sRef" refer="sId">
            <xs:selector xpath="ref"/>
            <xs:field xpath="@to"/>
          </xs:keyref>
        </xs:element>
      </xs:sequence>
    </xs:complexType>
  </xs:element>
</xs:schema>`

func TestRepeatedSiblingScopeKeyref(t *testing.T) {
	v := buildSchema(t, repeatedSiblingScope)
	t.Run("valid: each ref matches a key in its own section", func(t *testing.T) {
		assertValid(t, v, `<doc>`+
			`<section><item id="a"/><ref to="a"/></section>`+
			`<section><item id="b"/><ref to="b"/></section>`+
			`</doc>`)
	})
	t.Run("invalid: ref matches only a sibling section's key", func(t *testing.T) {
		// Section 2's ref to="a" matches only section 1's key, which is out of
		// scope; section 2 has no item id="a" → clause 4.3 no-match.
		assertInvalid(t, v, `<doc>`+
			`<section><item id="a"/></section>`+
			`<section><item id="b"/><ref to="a"/></section>`+
			`</doc>`, "cvc-identity-constraint")
	})
}

// nsKeySchema declares a key on the root <cat> whose selector is
// namespace-qualified (tns:item). The content model admits both a tns:item and,
// via a lax wildcard, an other-namespace <item> with the SAME local name. The
// wildcard uses processContents="lax", so the other-namespace <item> IS
// assessed (it is NOT recorded in a.skipped and so is NOT excluded by
// applyElementSteps' skip filter) — the namespace comparison in nameMatches is
// the only thing that can keep it out of the key. A local-name-only matcher
// would treat both as targets and clash on equal ids; the namespace-qualified
// selector must match only the tns:item, so two items of the same id in
// different namespaces do NOT collide.
const nsKeySchema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           xmlns:tns="urn:tns" targetNamespace="urn:tns"
           elementFormDefault="qualified">
  <xs:element name="cat">
    <xs:complexType>
      <xs:sequence>
        <xs:element name="item" maxOccurs="unbounded">
          <xs:complexType>
            <xs:attribute name="id" type="xs:string"/>
          </xs:complexType>
        </xs:element>
        <xs:any namespace="##other" processContents="lax"
                minOccurs="0" maxOccurs="unbounded"/>
      </xs:sequence>
    </xs:complexType>
    <xs:key name="kId">
      <xs:selector xpath="tns:item"/>
      <xs:field xpath="@id"/>
    </xs:key>
  </xs:element>
</xs:schema>`

func TestIDCNamespaceQualifiedSelector(t *testing.T) {
	v := buildSchema(t, nsKeySchema)
	t.Run("same id in another namespace does not clash", func(t *testing.T) {
		// Only the tns:item is a key target; the other-namespace <item id="a">
		// is outside the namespace-qualified selector, so no duplicate key.
		assertValid(t, v, `<tns:cat xmlns:tns="urn:tns" xmlns:o="urn:other">`+
			`<tns:item id="a"/><o:item id="a"/></tns:cat>`)
	})
	t.Run("duplicate within the qualified namespace still clashes", func(t *testing.T) {
		assertInvalid(t, v, `<tns:cat xmlns:tns="urn:tns">`+
			`<tns:item id="a"/><tns:item id="a"/></tns:cat>`, "cvc-identity-constraint")
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
