package parser

import (
	"testing"

	"github.com/kud360/goxsd5/builtin"
)

// schemaSeeds are whole schema documents exercising the build pipeline:
// simple/complex types, derivation, facets, particles, attributes, and a few
// structurally-broken documents that must be reported, not panicked on.
var schemaSeeds = []string{
	`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"/>`,
	`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
	   <xs:element name="e" type="xs:string"/>
	 </xs:schema>`,
	`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
	   <xs:simpleType name="t">
	     <xs:restriction base="xs:int">
	       <xs:minInclusive value="0"/>
	       <xs:maxInclusive value="10"/>
	     </xs:restriction>
	   </xs:simpleType>
	 </xs:schema>`,
	`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
	   <xs:complexType name="ct">
	     <xs:sequence>
	       <xs:element name="a" type="xs:string"/>
	       <xs:element name="b" type="xs:int" minOccurs="0" maxOccurs="unbounded"/>
	     </xs:sequence>
	     <xs:attribute name="id" type="xs:ID"/>
	   </xs:complexType>
	 </xs:schema>`,
	`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
	   <xs:simpleType name="lst"><xs:list itemType="xs:int"/></xs:simpleType>
	   <xs:simpleType name="un"><xs:union memberTypes="xs:int xs:string"/></xs:simpleType>
	 </xs:schema>`,
	// broken inputs that must surface errors, not crashes
	`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:element/></xs:schema>`,
	`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:element name="e" type="xs:nosuch"/></xs:schema>`,
	`<not-a-schema/>`,
	`<xs:schema`,
	``,
}

// FuzzParseSchema feeds arbitrary documents through the full public Parse
// pipeline via an in-memory resolver. The invariant is that no input — however
// malformed — causes a panic (the test harness fails on any panic), and that
// any schemas Parse hands back are individually non-nil and walkable. Note a
// root document that fails to load at all yields a nil slice with a non-nil
// error, which is a legitimate outcome, not a contract violation.
func FuzzParseSchema(f *testing.F) {
	for _, s := range schemaSeeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, doc string) {
		const root = "fuzz.xsd"
		schemas, err := Parse(root, &Options{Resolver: mapResolver{root: doc}})
		// Walking returned schemas (even partially-built ones) must not panic.
		for i, s := range schemas {
			if s == nil {
				t.Fatalf("nil schema at index %d in returned slice for %q (err: %v)", i, doc, err)
			}
		}
	})
}

var valueSeeds = []string{
	"", " ", "abc", "0", "-1", "1.5", "true", "false", "1", "0",
	"2001-10-26T21:32:52", "P1Y", "  spaced  ", "a b c",
	"http://example.com/", "urn:x", "name", "a:b", "9999999999999999999",
	"NaN", "INF", "-INF", "\t\n", "café",
}

// FuzzParseValue exercises lexical-to-value parsing against every built-in
// simple type. A fuzzed index selects the type; the parser must never panic
// on any lexical, and a successfully parsed value must be non-nil.
func FuzzParseValue(f *testing.F) {
	types := builtin.AllBuiltins()
	for _, lex := range valueSeeds {
		for i := range types {
			f.Add(i, lex)
		}
	}
	f.Fuzz(func(t *testing.T, idx int, lexical string) {
		types := builtin.AllBuiltins()
		if len(types) == 0 {
			t.Skip("no builtins")
		}
		st := types[((idx%len(types))+len(types))%len(types)]
		v, err := st.ParseValue(lexical, nil)
		if err != nil {
			return
		}
		if v == nil {
			t.Fatalf("%s.ParseValue(%q, nil) returned nil value, nil error", st.Name, lexical)
		}
	})
}
