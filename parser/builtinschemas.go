package parser

// Built-in schema documents for namespaces every processor is expected to
// know. The xml namespace (xml:lang, xml:space, xml:base, xml:id) is the
// canonical case: schemas import it from a well-known w3.org URL that a plain
// FileResolver cannot fetch, yet the import is meant to always succeed because
// the components are conventionally built in. builtinResolver intercepts those
// URLs and serves a bundled copy, delegating everything else to the wrapped
// resolver.

import (
	"io"
	"strings"
)

// xmlNamespaceSchema is the standard schema for the XML namespace
// (http://www.w3.org/XML/1998/namespace), per the W3C-published xml.xsd. It
// declares the four xml: attributes and the specialAttrs group.
const xmlNamespaceSchema = `<?xml version='1.0'?>
<xs:schema targetNamespace="http://www.w3.org/XML/1998/namespace"
           xmlns:xs="http://www.w3.org/2001/XMLSchema"
           xml:lang="en">
  <xs:attribute name="lang">
    <xs:simpleType>
      <xs:union memberTypes="xs:language">
        <xs:simpleType>
          <xs:restriction base="xs:string">
            <xs:enumeration value=""/>
          </xs:restriction>
        </xs:simpleType>
      </xs:union>
    </xs:simpleType>
  </xs:attribute>
  <xs:attribute name="space">
    <xs:simpleType>
      <xs:restriction base="xs:NCName">
        <xs:enumeration value="default"/>
        <xs:enumeration value="preserve"/>
      </xs:restriction>
    </xs:simpleType>
  </xs:attribute>
  <xs:attribute name="base" type="xs:anyURI"/>
  <xs:attribute name="id" type="xs:ID"/>
  <xs:attributeGroup name="specialAttrs">
    <xs:attribute ref="xml:base"/>
    <xs:attribute ref="xml:lang"/>
    <xs:attribute ref="xml:space"/>
    <xs:attribute ref="xml:id"/>
  </xs:attributeGroup>
</xs:schema>`

// builtinLocations maps the well-known schemaLocation URLs (the various
// revisions W3C has published over the years) to bundled document bodies.
var builtinLocations = map[string]string{
	"http://www.w3.org/2001/xml.xsd":     xmlNamespaceSchema,
	"https://www.w3.org/2001/xml.xsd":    xmlNamespaceSchema,
	"http://www.w3.org/2009/01/xml.xsd":  xmlNamespaceSchema,
	"https://www.w3.org/2009/01/xml.xsd": xmlNamespaceSchema,
	"http://www.w3.org/2007/08/xml.xsd":  xmlNamespaceSchema,
	"https://www.w3.org/2007/08/xml.xsd": xmlNamespaceSchema,
}

// builtinResolver serves bundled documents for well-known locations and
// delegates anything else to inner.
type builtinResolver struct{ inner SchemaResolver }

func (r builtinResolver) Resolve(location, base string) (io.ReadCloser, error) {
	if body, ok := builtinLocations[location]; ok {
		return io.NopCloser(strings.NewReader(body)), nil
	}
	if body, ok := builtinLocations[resolveLocation(location, base)]; ok {
		return io.NopCloser(strings.NewReader(body)), nil
	}
	return r.inner.Resolve(location, base)
}
