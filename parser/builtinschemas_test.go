package parser

// Built-in declarations every schema is expected to know without an explicit
// import: the xml namespace served by builtinResolver, the four XSI attribute
// declarations (§3.2.7), and the special xs:error type (§3.16.7.3). These are
// unit-tested here so they hold even when the W3C suite is not checked out.

import (
	"testing"

	"github.com/kud360/goxsd5/xsd"
)

func TestBuiltinXMLNamespaceImport(t *testing.T) {
	// Importing the xml namespace from the well-known w3.org URL must resolve
	// through builtinResolver even though no such file exists locally.
	schemas, err := parseMap(t, map[string]string{
		"a.xsd": `<xs:schema ` + xsNS + ` xmlns:xml="http://www.w3.org/XML/1998/namespace" targetNamespace="urn:a">
  <xs:import namespace="http://www.w3.org/XML/1998/namespace" schemaLocation="http://www.w3.org/2001/xml.xsd"/>
  <xs:complexType name="C">
    <xs:attributeGroup ref="xml:specialAttrs"/>
  </xs:complexType>
</xs:schema>`,
	}, "a.xsd")
	wantNoErr(t, err)
	var xmlSchema *xsd.Schema
	for _, s := range schemas {
		if s.TargetNamespace == xsd.XMLNS {
			xmlSchema = s
		}
	}
	if xmlSchema == nil {
		t.Fatal("xml namespace schema was not produced by the builtin resolver")
	}
	if _, ok := xmlSchema.Attributes[xsd.QName{Namespace: xsd.XMLNS, Local: "lang"}]; !ok {
		t.Error("xml:lang attribute declaration missing")
	}
}

func TestBuiltinXSIAttributesNeedNoImport(t *testing.T) {
	// References to xsi:type / xsi:nil resolve without importing the XSI
	// namespace (§3.2.7); the default on xsi:type is checked against QName.
	_, err := parseMap(t, map[string]string{
		"a.xsd": `<xs:schema ` + xsNS + ` xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" targetNamespace="urn:a">
  <xs:complexType name="C">
    <xs:simpleContent>
      <xs:extension base="xs:decimal">
        <xs:attribute ref="xsi:type" default="xs:date"/>
        <xs:attribute ref="xsi:nil"/>
      </xs:extension>
    </xs:simpleContent>
  </xs:complexType>
</xs:schema>`,
	}, "a.xsd")
	wantNoErr(t, err)
}

func TestBuiltinErrorType(t *testing.T) {
	// xs:error resolves as a type without import; no value is valid against it.
	_, err := parseMap(t, map[string]string{
		"a.xsd": `<xs:schema ` + xsNS + ` targetNamespace="urn:a">
  <xs:element name="e" type="xs:error"/>
</xs:schema>`,
	}, "a.xsd")
	wantNoErr(t, err)
}
