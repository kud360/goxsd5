// Package xsd is the public model for XSD 1.1 schemas: components, the
// simple-type value/facet engine, and the structured error types shared by
// the whole module. It has no dependency on the parser.
package xsd

// Well-known namespaces.
const (
	// XSDNS is the XML Schema definition namespace.
	XSDNS = "http://www.w3.org/2001/XMLSchema"
	// XSINS is the XML Schema instance namespace (xsi:type, xsi:nil, …).
	XSINS = "http://www.w3.org/2001/XMLSchema-instance"
	// XMLNS is the namespace bound to the reserved `xml` prefix.
	XMLNS = "http://www.w3.org/XML/1998/namespace"
	// XMLNSNS is the namespace of namespace declarations themselves.
	XMLNSNS = "http://www.w3.org/2000/xmlns/"
	// VCNS is the XSD 1.1 version-control namespace (vc:minVersion, …).
	VCNS = "http://www.w3.org/2007/XMLSchema-versioning"
)

// Pos is a source position within a schema document.
type Pos struct {
	URI    string
	Line   int
	Column int
}

func (p Pos) IsZero() bool { return p == Pos{} }

func (p Pos) String() string {
	if p.IsZero() {
		return "<unknown>"
	}
	return p.URI + ":" + itoa(p.Line) + ":" + itoa(p.Column)
}

// QName is an expanded XML name: namespace URI plus local part.
type QName struct {
	Namespace string
	Local     string
}

func (q QName) IsZero() bool { return q == QName{} }

func (q QName) String() string {
	if q.Namespace == "" {
		return q.Local
	}
	return "{" + q.Namespace + "}" + q.Local
}

// itoa avoids importing strconv in the hot error path for no real reason
// other than keeping this file dependency-free; it is a plain Itoa.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	n := len(buf)
	for i > 0 {
		n--
		buf[n] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		n--
		buf[n] = '-'
	}
	return string(buf[n:])
}
