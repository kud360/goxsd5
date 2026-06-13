package parser

import (
	"slices"
	"strings"
	"testing"

	"github.com/kud360/goxsd5/parser/xmltree"
	"github.com/kud360/goxsd5/xsd"
)

// load parses src as one schema document and runs pass 1 over it.
func load(t *testing.T, src string) (*schemaDoc, *registry, *xsd.ErrorList) {
	t.Helper()
	root, err := xmltree.Parse(strings.NewReader(src), "test.xsd")
	if err != nil {
		t.Fatalf("xmltree.Parse: %v", err)
	}
	errs := &xsd.ErrorList{}
	doc := loadDoc(root, "test.xsd", errs)
	reg := newRegistry()
	if doc != nil {
		registerDoc(reg, doc, errs)
	}
	return doc, reg, errs
}

func wantIDs(t *testing.T, errs *xsd.ErrorList, ids ...string) {
	t.Helper()
	got := xsd.RefIDs(errs.Err())
	for _, id := range ids {
		if !slices.Contains(got, id) {
			t.Errorf("missing expected error id %q; got %v\nerrors: %v", id, got, errs.Err())
		}
	}
}

func wantClean(t *testing.T, errs *xsd.ErrorList) {
	t.Helper()
	if !errs.Empty() {
		t.Errorf("unexpected errors: %v", errs.Err())
	}
}

const xmlnsXS = `xmlns:xs="http://www.w3.org/2001/XMLSchema"`

// kitchenSink exercises most schema constructs; pass 1 and pass 2 tests
// share it.
var kitchenSink = `<?xml version="1.0"?>
<xs:schema ` + xmlnsXS + ` xmlns:tns="urn:test" targetNamespace="urn:test"
           elementFormDefault="qualified" blockDefault="#all"
           version="1.0" xml:lang="en">
  <xs:annotation>
    <xs:documentation xml:lang="en">A kitchen-sink schema.
      <html xmlns="http://www.w3.org/1999/xhtml"><b>free</b> content</html>
    </xs:documentation>
    <xs:appinfo source="urn:app"><foo:directive xmlns:foo="urn:foo" weight="3"/></xs:appinfo>
  </xs:annotation>
  <xs:import namespace="urn:other" schemaLocation="other.xsd"/>
  <xs:include schemaLocation="more.xsd"/>

  <xs:defaultOpenContent mode="suffix"><xs:any namespace="##other" processContents="lax"/></xs:defaultOpenContent>

  <xs:simpleType name="size">
    <xs:restriction base="xs:token">
      <xs:enumeration value="small"/>
      <xs:enumeration value="large"/>
      <xs:pattern value="[a-z]+"/>
    </xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="sizes"><xs:list itemType="tns:size"/></xs:simpleType>
  <xs:simpleType name="sizeOrInt">
    <xs:union memberTypes="tns:size xs:int"/>
  </xs:simpleType>
  <xs:simpleType name="inlineUnion" final="#all">
    <xs:union>
      <xs:simpleType><xs:restriction base="xs:int"><xs:minInclusive value="0" fixed="true"/></xs:restriction></xs:simpleType>
    </xs:union>
  </xs:simpleType>

  <xs:complexType name="base" abstract="true" mixed="true">
    <xs:openContent mode="interleave"><xs:any namespace="##any" processContents="skip"/></xs:openContent>
    <xs:sequence>
      <xs:element name="a" type="xs:string" minOccurs="0" maxOccurs="unbounded"/>
      <xs:choice minOccurs="0">
        <xs:element ref="tns:doc"/>
        <xs:any notNamespace="urn:test ##local" notQName="##defined" processContents="lax"/>
      </xs:choice>
      <xs:group ref="tns:body" maxOccurs="2"/>
    </xs:sequence>
    <xs:attribute name="id" type="xs:ID" use="required"/>
    <xs:attributeGroup ref="tns:common"/>
    <xs:anyAttribute namespace="##other" notQName="##defined" processContents="lax"/>
    <xs:assert test="@id != ''" xpathDefaultNamespace="##targetNamespace"/>
  </xs:complexType>

  <xs:complexType name="derived">
    <xs:complexContent mixed="true">
      <xs:extension base="tns:base">
        <xs:sequence><xs:element name="extra" type="tns:size" form="unqualified"/></xs:sequence>
        <xs:attribute name="version" type="xs:decimal" default="1.0" use="optional"/>
      </xs:extension>
    </xs:complexContent>
  </xs:complexType>

  <xs:complexType name="measured">
    <xs:simpleContent>
      <xs:restriction base="tns:withUnit">
        <xs:simpleType><xs:restriction base="xs:decimal"><xs:totalDigits value="5"/></xs:restriction></xs:simpleType>
        <xs:minInclusive value="0"/>
        <xs:fractionDigits value="2" fixed="false"/>
        <xs:attribute name="unit" type="xs:NMTOKEN" use="required"/>
      </xs:restriction>
    </xs:simpleContent>
  </xs:complexType>
  <xs:complexType name="withUnit">
    <xs:simpleContent>
      <xs:extension base="xs:decimal">
        <xs:attribute name="unit" type="xs:NMTOKEN"/>
        <xs:anyAttribute namespace="##local"/>
      </xs:extension>
    </xs:simpleContent>
  </xs:complexType>

  <xs:group name="body">
    <xs:choice>
      <xs:element name="p" type="xs:string"/>
      <xs:sequence><xs:element name="q" type="xs:string"/></xs:sequence>
    </xs:choice>
  </xs:group>
  <xs:attributeGroup name="common">
    <xs:attribute name="lang" type="xs:language"/>
    <xs:attribute ref="tns:globalAttr" use="optional"/>
  </xs:attributeGroup>
  <xs:attribute name="globalAttr" type="xs:string" inheritable="true"/>

  <xs:element name="doc" type="tns:derived" nillable="true" block="extension">
    <xs:key name="docKey"><xs:selector xpath=".//tns:a"/><xs:field xpath="@id"/></xs:key>
    <xs:keyref name="docRef" refer="tns:docKey"><xs:selector xpath=".//tns:extra"/><xs:field xpath="@ref"/></xs:keyref>
    <xs:key ref="tns:docKey"/>
  </xs:element>
  <xs:element name="anything" abstract="true">
    <xs:alternative test="@kind='base'" type="tns:base"/>
    <xs:alternative type="xs:anyType"/>
  </xs:element>
  <xs:element name="local-sizes" substitutionGroup="tns:anything">
    <xs:complexType>
      <xs:all>
        <xs:element name="x" type="xs:int" maxOccurs="3"/>
        <xs:any namespace="##targetNamespace" minOccurs="0"/>
      </xs:all>
    </xs:complexType>
  </xs:element>

  <xs:notation name="png" public="image/png" system="viewer.exe"/>
</xs:schema>`

func TestValidKitchenSink(t *testing.T) {
	doc, reg, errs := load(t, kitchenSink)
	wantClean(t, errs)

	// Document-level properties.
	if doc.targetNamespace != "urn:test" {
		t.Errorf("targetNamespace = %q", doc.targetNamespace)
	}
	if doc.elementFormDefault != xsd.FormQualified || doc.attributeFormDefault != xsd.FormUnqualified {
		t.Errorf("form defaults = %v/%v", doc.elementFormDefault, doc.attributeFormDefault)
	}
	if doc.blockDefault != blockDefSet {
		t.Errorf("blockDefault = %v, want #all expansion %v", doc.blockDefault, blockDefSet)
	}
	if doc.finalDefault != 0 {
		t.Errorf("finalDefault = %v, want unset", doc.finalDefault)
	}
	if len(doc.compositions) != 2 || doc.compositions[0].kind != "import" || doc.compositions[1].kind != "include" {
		t.Errorf("compositions = %+v", doc.compositions)
	}
	if doc.defaultOpenContent == nil {
		t.Error("defaultOpenContent not captured")
	}

	// Registry: globals land in their symbol spaces; builtins are seeded.
	for _, probe := range []struct {
		s    space
		name string
	}{
		{spaceType, "size"}, {spaceType, "base"}, {spaceType, "derived"},
		{spaceElement, "doc"}, {spaceAttribute, "globalAttr"},
		{spaceGroup, "body"}, {spaceAttrGroup, "common"}, {spaceNotation, "png"},
		{spaceIC, "docKey"}, {spaceIC, "docRef"},
	} {
		if reg.lookup(probe.s, xsd.QName{Namespace: "urn:test", Local: probe.name}) == nil {
			t.Errorf("registry missing %s %q", probe.s, probe.name)
		}
	}
	if d := reg.lookup(spaceType, xsd.QName{Namespace: xsd.XSDNS, Local: "string"}); d == nil || d.builtin == nil {
		t.Error("builtin xs:string not seeded")
	}
	if d := reg.lookup(spaceType, xsd.QName{Namespace: xsd.XSDNS, Local: "anyType"}); d == nil || d.builtin == nil {
		t.Error("builtin xs:anyType not seeded")
	}
}

func TestStructuralNegatives(t *testing.T) {
	cases := []struct {
		name string
		body string // wrapped in <xs:schema targetNamespace="urn:t" ...>
		ids  []string
	}{
		{"unknown element", `<xs:frobnicate/>`, []string{"src-schema"}},
		{"include after declarations",
			`<xs:simpleType name="t"><xs:restriction base="xs:int"/></xs:simpleType><xs:include schemaLocation="x.xsd"/>`,
			[]string{"src-schema"}},
		{"text inside sequence",
			`<xs:complexType name="t"><xs:sequence>oops</xs:sequence></xs:complexType>`,
			[]string{"src-model_group_defn"}},
		{"simpleType inside sequence",
			`<xs:complexType name="t"><xs:sequence><xs:simpleType/></xs:sequence></xs:complexType>`,
			[]string{"src-model_group_defn"}},
		{"simpleContent and particle together",
			`<xs:complexType name="t"><xs:simpleContent><xs:extension base="xs:int"/></xs:simpleContent><xs:sequence/></xs:complexType>`,
			[]string{"src-ct"}},
		{"annotation not first",
			`<xs:simpleType name="t"><xs:restriction base="xs:int"/><xs:annotation/></xs:simpleType>`,
			[]string{"src-simple-type"}},
		{"selector after field",
			`<xs:element name="e" type="xs:int"><xs:key name="k"><xs:field xpath="a"/><xs:selector xpath="b"/></xs:key></xs:element>`,
			[]string{"src-identity-constraint"}},
		{"foreign element in schema content", `<bogus xmlns="urn:other"/>`, []string{"src-schema"}},
		{"simpleType without derivation", `<xs:simpleType name="t"/>`, []string{"src-simple-type"}},
		{"two inline types in element",
			`<xs:element name="e"><xs:simpleType><xs:restriction base="xs:int"/></xs:simpleType><xs:complexType/></xs:element>`,
			[]string{"src-element"}},
		{"maxOccurs on an openContent wildcard",
			`<xs:complexType name="t"><xs:openContent mode="interleave"><xs:any namespace="##any" maxOccurs="unbounded"/></xs:openContent><xs:sequence/></xs:complexType>`,
			[]string{"src-ct"}},
		{"minOccurs on a defaultOpenContent wildcard",
			`<xs:defaultOpenContent mode="interleave"><xs:any namespace="##any" minOccurs="0"/></xs:defaultOpenContent>`,
			[]string{"src-schema"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, errs := load(t, `<xs:schema `+xmlnsXS+` xmlns:tns="urn:t" targetNamespace="urn:t">`+tc.body+`</xs:schema>`)
			if errs.Empty() {
				t.Fatal("expected errors, got none")
			}
			wantIDs(t, errs, tc.ids...)
		})
	}
}

func TestAttributeNegatives(t *testing.T) {
	cases := []struct {
		name string
		body string
		ids  []string
	}{
		{"bogus attribute on schema element",
			`<xs:simpleType name="t" frobnicate="1"><xs:restriction base="xs:int"/></xs:simpleType>`,
			[]string{"src-simple-type"}},
		{"xsd-namespace attribute",
			`<xs:element name="e" xs:type="xs:int"/>`,
			[]string{"src-element"}},
		{"ref on global element", `<xs:element ref="tns:other"/>`, []string{"src-element"}},
		{"global element missing name", `<xs:element type="xs:int"/>`, []string{"src-element"}},
		{"global simpleType missing name", `<xs:simpleType><xs:restriction base="xs:int"/></xs:simpleType>`, []string{"src-simple-type"}},
		{"element name and ref",
			`<xs:complexType name="c"><xs:sequence><xs:element name="a" ref="tns:b"/></xs:sequence></xs:complexType>`,
			[]string{"src-element"}},
		{"element ref with type",
			`<xs:complexType name="c"><xs:sequence><xs:element ref="tns:b" type="xs:int"/></xs:sequence></xs:complexType>`,
			[]string{"src-element"}},
		{"element default and fixed",
			`<xs:element name="e" type="xs:int" default="1" fixed="2"/>`,
			[]string{"src-element"}},
		{"attribute default and fixed",
			`<xs:attribute name="a" type="xs:int" default="1" fixed="2"/>`,
			[]string{"src-attribute"}},
		{"attribute default with use required",
			`<xs:complexType name="c"><xs:attribute name="a" type="xs:int" default="1" use="required"/></xs:complexType>`,
			[]string{"src-attribute"}},
		{"attribute named xmlns", `<xs:attribute name="xmlns" type="xs:string"/>`, []string{"no-xmlns"}},
		{"minOccurs greater than maxOccurs",
			`<xs:complexType name="c"><xs:sequence><xs:element name="a" type="xs:int" minOccurs="3" maxOccurs="2"/></xs:sequence></xs:complexType>`,
			[]string{"p-props-correct"}},
		{"negative maxOccurs",
			`<xs:complexType name="c"><xs:sequence><xs:element name="a" type="xs:int" maxOccurs="-1"/></xs:sequence></xs:complexType>`,
			[]string{"src-element"}},
		{"occurs on all in group definition",
			`<xs:group name="g"><xs:all minOccurs="0"><xs:element name="a" type="xs:int"/></xs:all></xs:group>`,
			[]string{"src-model_group_defn"}},
		{"maxOccurs 2 on all",
			`<xs:complexType name="c"><xs:all maxOccurs="2"><xs:element name="a" type="xs:int"/></xs:all></xs:complexType>`,
			[]string{"src-model_group_defn"}},
		{"restriction base and inline simpleType",
			`<xs:simpleType name="t"><xs:restriction base="xs:int"><xs:simpleType><xs:restriction base="xs:int"/></xs:simpleType></xs:restriction></xs:simpleType>`,
			[]string{"src-restriction-base-or-simpleType"}},
		{"list without item type",
			`<xs:simpleType name="t"><xs:list/></xs:simpleType>`,
			[]string{"src-list-itemType-or-simpleType"}},
		{"empty union",
			`<xs:simpleType name="t"><xs:union/></xs:simpleType>`,
			[]string{"src-union-memberTypes-or-simpleTypes"}},
		{"alternative with type attr and inline type",
			`<xs:element name="e" type="xs:anyType"><xs:alternative test="@a" type="xs:int"><xs:simpleType><xs:restriction base="xs:int"/></xs:simpleType></xs:alternative></xs:element>`,
			[]string{"src-ta"}},
		{"default alternative not last",
			`<xs:element name="e" type="xs:anyType"><xs:alternative type="xs:int"/><xs:alternative test="@a" type="xs:string"/></xs:element>`,
			[]string{"src-element"}},
		{"element type attribute with inline type",
			`<xs:element name="e" type="xs:int"><xs:simpleType><xs:restriction base="xs:int"/></xs:simpleType></xs:element>`,
			[]string{"src-element"}},
		{"attribute type attribute with inline simpleType",
			`<xs:attribute name="a" type="xs:int"><xs:simpleType><xs:restriction base="xs:int"/></xs:simpleType></xs:attribute>`,
			[]string{"src-attribute"}},
		{"attribute fixed with use prohibited",
			`<xs:complexType name="c"><xs:attribute name="a" type="xs:int" fixed="1" use="prohibited"/></xs:complexType>`,
			[]string{"src-attribute"}},
		{"attribute ref with inheritable is fine, with form is not",
			`<xs:complexType name="c"><xs:attribute ref="tns:a" inheritable="true" form="qualified"/></xs:complexType>`,
			[]string{"src-attribute"}},
		{"named keyref without refer",
			`<xs:element name="e" type="xs:int"><xs:keyref name="r"><xs:selector xpath="a"/><xs:field xpath="b"/></xs:keyref></xs:element>`,
			[]string{"src-identity-constraint"}},
		{"identity constraint ref plus selector",
			`<xs:element name="e" type="xs:int"><xs:unique ref="tns:k"><xs:selector xpath="a"/><xs:field xpath="b"/></xs:unique></xs:element>`,
			[]string{"src-identity-constraint"}},
		{"notation without identifiers", `<xs:notation name="n"/>`, []string{"n-props-correct"}},
		{"wildcard namespace and notNamespace",
			`<xs:complexType name="c"><xs:sequence><xs:any namespace="##any" notNamespace="##local"/></xs:sequence></xs:complexType>`,
			[]string{"src-wildcard"}},
		{"##any in namespace list",
			`<xs:complexType name="c"><xs:sequence><xs:any namespace="##any urn:x"/></xs:sequence></xs:complexType>`,
			[]string{"src-wildcard"}},
		{"unresolvable prefix in type", `<xs:element name="e" type="undefined:t"/>`, []string{"src-qname"}},
		{"bad whiteSpace value",
			`<xs:simpleType name="t"><xs:restriction base="xs:string"><xs:whiteSpace value="fold"/></xs:restriction></xs:simpleType>`,
			[]string{"src-simple-type"}},
		{"negative minLength",
			`<xs:simpleType name="t"><xs:restriction base="xs:string"><xs:minLength value="-1"/></xs:restriction></xs:simpleType>`,
			[]string{"src-simple-type"}},
		{"zero totalDigits",
			`<xs:simpleType name="t"><xs:restriction base="xs:decimal"><xs:totalDigits value="0"/></xs:restriction></xs:simpleType>`,
			[]string{"src-simple-type"}},
		{"import of own namespace",
			`<xs:import namespace="urn:t" schemaLocation="self.xsd"/>`,
			[]string{"src-import"}},
		{"missing required facet value",
			`<xs:simpleType name="t"><xs:restriction base="xs:string"><xs:maxLength/></xs:restriction></xs:simpleType>`,
			[]string{"src-simple-type"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, errs := load(t, `<xs:schema `+xmlnsXS+` xmlns:tns="urn:t" targetNamespace="urn:t">`+tc.body+`</xs:schema>`)
			if errs.Empty() {
				t.Fatal("expected errors, got none")
			}
			wantIDs(t, errs, tc.ids...)
		})
	}
}

func TestSchemaLevelNegatives(t *testing.T) {
	t.Run("empty targetNamespace", func(t *testing.T) {
		_, _, errs := load(t, `<xs:schema `+xmlnsXS+` targetNamespace=""/>`)
		wantIDs(t, errs, "src-schema")
	})
	t.Run("import without namespace into no-namespace schema", func(t *testing.T) {
		_, _, errs := load(t, `<xs:schema `+xmlnsXS+`><xs:import schemaLocation="x.xsd"/></xs:schema>`)
		wantIDs(t, errs, "src-import")
	})
	t.Run("wrong root element", func(t *testing.T) {
		_, _, errs := load(t, `<xs:element `+xmlnsXS+` name="e"/>`)
		wantIDs(t, errs, "src-schema")
	})
	t.Run("duplicate ids", func(t *testing.T) {
		_, _, errs := load(t, `<xs:schema `+xmlnsXS+` id="x">
  <xs:simpleType name="t" id="x"><xs:restriction base="xs:int"/></xs:simpleType>
</xs:schema>`)
		wantIDs(t, errs, "src-id")
	})
	t.Run("duplicate global type", func(t *testing.T) {
		_, _, errs := load(t, `<xs:schema `+xmlnsXS+` xmlns:tns="urn:t" targetNamespace="urn:t">
  <xs:simpleType name="t"><xs:restriction base="xs:int"/></xs:simpleType>
  <xs:complexType name="t"/>
</xs:schema>`)
		wantIDs(t, errs, "sch-props-correct")
	})
	t.Run("duplicate across element and type spaces is fine", func(t *testing.T) {
		_, _, errs := load(t, `<xs:schema `+xmlnsXS+` xmlns:tns="urn:t" targetNamespace="urn:t">
  <xs:simpleType name="t"><xs:restriction base="xs:int"/></xs:simpleType>
  <xs:element name="t" type="tns:t"/>
</xs:schema>`)
		wantClean(t, errs)
	})
	t.Run("redeclaring a builtin type", func(t *testing.T) {
		_, _, errs := load(t, `<xs:schema `+xmlnsXS+` targetNamespace="http://www.w3.org/2001/XMLSchema">
  <xs:simpleType name="string"><xs:restriction base="xs:int"/></xs:simpleType>
</xs:schema>`)
		wantIDs(t, errs, "sch-props-correct")
	})
}

func TestConditionalInclusion(t *testing.T) {
	t.Run("future-version garbage is pruned", func(t *testing.T) {
		doc, reg, errs := load(t, `<xs:schema `+xmlnsXS+` xmlns:vc="http://www.w3.org/2007/XMLSchema-versioning" xmlns:tns="urn:t" targetNamespace="urn:t">
  <xs:simpleType name="t" vc:minVersion="1.2"><xs:thisDoesNotExist/></xs:simpleType>
  <xs:simpleType name="t" vc:maxVersion="1.1"><xs:alsoNot/></xs:simpleType>
  <xs:simpleType name="t" vc:minVersion="1.0" vc:maxVersion="1.2"><xs:restriction base="xs:int"/></xs:simpleType>
</xs:schema>`)
		wantClean(t, errs)
		// Only the applicable declaration registered; no duplicate errors.
		d := reg.lookup(spaceType, xsd.QName{Namespace: "urn:t", Local: "t"})
		if d == nil {
			t.Fatal("applicable simpleType not registered")
		}
		if len(doc.pruned) != 2 {
			t.Errorf("pruned %d nodes, want 2", len(doc.pruned))
		}
	})
	t.Run("minVersion equal to processor version is kept", func(t *testing.T) {
		_, _, errs := load(t, `<xs:schema `+xmlnsXS+` xmlns:vc="http://www.w3.org/2007/XMLSchema-versioning">
  <xs:element name="e" type="xs:int" vc:minVersion="1.1"/>
</xs:schema>`)
		wantClean(t, errs)
	})
	t.Run("typeAvailable/typeUnavailable select a branch", func(t *testing.T) {
		// xs:error is available, so the typeAvailable branch is kept and the
		// typeUnavailable branch is pruned: no duplicate attribute.
		_, _, errs := load(t, `<xs:schema `+xmlnsXS+` xmlns:vc="http://www.w3.org/2007/XMLSchema-versioning">
  <xs:complexType name="c">
    <xs:attribute name="a" type="xs:error" vc:typeAvailable="xs:error"/>
    <xs:attribute name="a" vc:typeUnavailable="xs:error"/>
  </xs:complexType>
</xs:schema>`)
		wantClean(t, errs)
	})
	t.Run("unavailable type prunes the element", func(t *testing.T) {
		// xs:precisionDecimal is not supported, so the element using it is
		// removed and its broken content is never judged.
		_, _, errs := load(t, `<xs:schema `+xmlnsXS+` xmlns:vc="http://www.w3.org/2007/XMLSchema-versioning">
  <xs:element name="e" type="xs:precisionDecimal" vc:typeAvailable="xs:precisionDecimal"/>
</xs:schema>`)
		wantClean(t, errs)
	})
	t.Run("invalid vc:minVersion is an error", func(t *testing.T) {
		_, _, errs := load(t, `<xs:schema `+xmlnsXS+` xmlns:vc="http://www.w3.org/2007/XMLSchema-versioning">
  <xs:element name="e" type="xs:int" vc:minVersion="1.1.3"/>
</xs:schema>`)
		wantIDs(t, errs, "cip")
	})
	t.Run("invalid QName in vc:typeUnavailable is an error", func(t *testing.T) {
		_, _, errs := load(t, `<xs:schema `+xmlnsXS+` xmlns:vc="http://www.w3.org/2007/XMLSchema-versioning">
  <xs:element name="e" type="xs:int" vc:typeUnavailable="xs:integer 23"/>
</xs:schema>`)
		wantIDs(t, errs, "cip")
	})
	t.Run("a conditionally-ignored schema element is emptied", func(t *testing.T) {
		// vc:maxVersion <= 1.1 ignores the whole <schema>, so the bogus child
		// is removed rather than rejected.
		_, _, errs := load(t, `<xs:schema `+xmlnsXS+` xmlns:vc="http://www.w3.org/2007/XMLSchema-versioning" vc:maxVersion="0.9">
  <xs:somethingInvalid/>
</xs:schema>`)
		wantClean(t, errs)
	})
}

func TestLocalDeclTargetNamespace(t *testing.T) {
	// A local element/attribute with a differing targetNamespace is valid only
	// inside a complexType restriction whose base is not xs:anyType.
	t.Run("local element targetNamespace without a restriction is invalid", func(t *testing.T) {
		_, _, errs := load(t, `<xs:schema `+xmlnsXS+` targetNamespace="urn:a">
  <xs:complexType name="c"><xs:sequence>
    <xs:element name="e" type="xs:int" targetNamespace="urn:b"/>
  </xs:sequence></xs:complexType>
</xs:schema>`)
		wantIDs(t, errs, "src-element")
	})
	t.Run("local attribute targetNamespace without a restriction is invalid", func(t *testing.T) {
		_, _, errs := load(t, `<xs:schema `+xmlnsXS+` targetNamespace="urn:a">
  <xs:complexType name="c">
    <xs:attribute name="s" type="xs:int" targetNamespace="urn:b"/>
  </xs:complexType>
</xs:schema>`)
		wantIDs(t, errs, "src-attribute")
	})
	t.Run("local element targetNamespace equal to the schema's is fine", func(t *testing.T) {
		_, _, errs := load(t, `<xs:schema `+xmlnsXS+` targetNamespace="urn:a">
  <xs:complexType name="c"><xs:sequence>
    <xs:element name="e" type="xs:int" targetNamespace="urn:a"/>
  </xs:sequence></xs:complexType>
</xs:schema>`)
		wantClean(t, errs)
	})
}

func TestRedefineOverrideRegistration(t *testing.T) {
	// Pass 1 alone records compositions; the loader registers redefine/
	// override children once their targets are loaded (parser_test.go covers
	// the cross-document semantics).
	doc, reg, errs := load(t, `<xs:schema `+xmlnsXS+` xmlns:tns="urn:t" targetNamespace="urn:t">
  <xs:redefine schemaLocation="base.xsd">
    <xs:simpleType name="t"><xs:restriction base="tns:t"><xs:maxLength value="3"/></xs:restriction></xs:simpleType>
  </xs:redefine>
  <xs:override schemaLocation="base.xsd">
    <xs:element name="e" type="xs:string"/>
    <xs:notation name="n" public="p"/>
  </xs:override>
  <xs:simpleType name="t2"><xs:restriction base="xs:int"/></xs:simpleType>
</xs:schema>`)
	wantClean(t, errs)
	if reg.lookup(spaceType, xsd.QName{Namespace: "urn:t", Local: "t"}) != nil {
		t.Error("redefine child registered without the loader")
	}
	if reg.lookup(spaceType, xsd.QName{Namespace: "urn:t", Local: "t2"}) == nil {
		t.Error("ordinary global missing from the registry")
	}
	if len(doc.compositions) != 2 {
		t.Errorf("compositions = %+v, want redefine+override", doc.compositions)
	}
}

func TestForeignContent(t *testing.T) {
	t.Run("foreign attributes allowed everywhere", func(t *testing.T) {
		_, _, errs := load(t, `<xs:schema `+xmlnsXS+` xmlns:meta="urn:meta" meta:owner="kud">
  <xs:element name="e" type="xs:int" meta:since="2024"/>
</xs:schema>`)
		wantClean(t, errs)
	})
	t.Run("foreign elements allowed among facets", func(t *testing.T) {
		_, _, errs := load(t, `<xs:schema `+xmlnsXS+` xmlns:m="urn:m">
  <xs:simpleType name="t">
    <xs:restriction base="xs:int"><xs:minInclusive value="1"/><m:hint precision="high"/><xs:maxInclusive value="9"/></xs:restriction>
  </xs:simpleType>
</xs:schema>`)
		wantClean(t, errs)
	})
	t.Run("foreign element rejected in complexType", func(t *testing.T) {
		_, _, errs := load(t, `<xs:schema `+xmlnsXS+` xmlns:m="urn:m">
  <xs:complexType name="t"><m:hint/></xs:complexType>
</xs:schema>`)
		wantIDs(t, errs, "src-ct")
	})
}

func TestContentMatcher(t *testing.T) {
	g := seq(opt(one("a")), star(names("b", "c")), one("d"))
	for _, tc := range []struct {
		toks []string
		ok   bool
	}{
		{[]string{"d"}, true},
		{[]string{"a", "d"}, true},
		{[]string{"a", "b", "c", "b", "d"}, true},
		{[]string{"b", "d"}, true},
		{[]string{"a", "a", "d"}, false},
		{[]string{"a", "b"}, false},
		{[]string{}, false},
		{[]string{"d", "d"}, false},
	} {
		ok, _, _ := matchContent(g, tc.toks)
		if ok != tc.ok {
			t.Errorf("match(%v) = %v, want %v", tc.toks, ok, tc.ok)
		}
	}
	// Failure positions point at the offending token.
	_, idx, expected := matchContent(g, []string{"a", "x", "d"})
	if idx != 1 || len(expected) == 0 {
		t.Errorf("failIdx = %d (expected %v), want 1", idx, expected)
	}
}

// TestBuildSmoke runs pass 2 over the kitchen sink: it must produce a
// linked schema without errors. Focused builder coverage lives in
// build_test.go.
func TestBuildSmoke(t *testing.T) {
	doc, reg, errs := load(t, kitchenSink)
	s := buildSchema(reg, doc, errs)
	wantClean(t, errs)
	if s == nil {
		t.Fatal("no schema built")
	}
	derived, ok := s.Types[xsd.QName{Namespace: "urn:test", Local: "derived"}].(*xsd.ComplexType)
	if !ok {
		t.Fatal("derived type missing")
	}
	base, ok := derived.BaseType.(*xsd.ComplexType)
	if !ok || base.Name.Local != "base" {
		t.Fatalf("derived.BaseType = %v, want tns:base", derived.BaseType)
	}
	if derived.DerivationMethod != xsd.DeriveExtension {
		t.Error("derived should be an extension")
	}
	doc2 := s.Elements[xsd.QName{Namespace: "urn:test", Local: "doc"}]
	if doc2 == nil || doc2.Type != derived {
		t.Fatalf("element doc not linked to tns:derived")
	}
	if len(doc2.IdentityConstraints) != 3 {
		t.Errorf("doc has %d identity constraints, want 3", len(doc2.IdentityConstraints))
	}
	size, ok := s.Types[xsd.QName{Namespace: "urn:test", Local: "size"}].(*xsd.SimpleType)
	if !ok || !size.Facets.HasEnumeration || len(size.Facets.Enumeration) != 2 {
		t.Fatalf("size type facets not built: %+v", size)
	}
	if size.Primitive == nil || size.Primitive.Name.Local != "string" {
		t.Errorf("size primitive = %v, want string", size.Primitive)
	}
}
