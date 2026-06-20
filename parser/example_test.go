package parser_test

// ExampleParse demonstrates the primary parse entry point: load an XSD 1.1
// schema from a file, then inspect the resulting Schema components.

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/kud360/goxsd5/parser"
)

func ExampleParse() {
	// Write a small schema to a temp file; parser.Parse requires a file path
	// or URL (use Options.Resolver to serve from memory in your own tests).
	dir, err := os.MkdirTemp("", "goxsd5-example-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	const schema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
           targetNamespace="urn:example"
           xmlns:tns="urn:example">
  <xs:element name="order"  type="xs:string"/>
  <xs:element name="item"   type="xs:string"/>
  <xs:element name="person" type="xs:string"/>
</xs:schema>`

	path := filepath.Join(dir, "schema.xsd")
	if err := os.WriteFile(path, []byte(schema), 0o644); err != nil {
		panic(err)
	}

	schemas, err := parser.Parse(path, nil)
	if err != nil {
		fmt.Println("parse error:", err)
		return
	}

	// schemas[0] is the root document's namespace.
	s := schemas[0]
	fmt.Println("target namespace:", s.TargetNamespace)

	// Collect global element names and sort for deterministic output.
	names := make([]string, 0, len(s.Elements))
	for q := range s.Elements {
		names = append(names, q.Local)
	}
	sort.Strings(names)

	fmt.Println("global elements:", len(names))
	for _, n := range names {
		fmt.Println(" ", n)
	}

	// Output:
	// target namespace: urn:example
	// global elements: 3
	//   item
	//   order
	//   person
}
