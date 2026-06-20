package xsdvalidate_test

// ExampleValidate demonstrates the two-step validation flow: compile a schema
// with parser.Parse, build a Validator with xsdvalidate.New, then assess XML
// instance documents.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kud360/goxsd5/parser"
	"github.com/kud360/goxsd5/xsd"
	"github.com/kud360/goxsd5/xsdvalidate"
	"github.com/kud360/goxsd5/xsdvalidate/xmlsrc"
)

func ExampleNew() {
	// Step 1: write a schema to a temp file and parse it.
	dir, err := os.MkdirTemp("", "goxsd5-validate-example-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	const schema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="age" type="xs:positiveInteger"/>
</xs:schema>`

	path := filepath.Join(dir, "schema.xsd")
	if err := os.WriteFile(path, []byte(schema), 0o644); err != nil {
		panic(err)
	}

	schemas, err := parser.Parse(path, nil)
	if err != nil {
		fmt.Println("schema error:", err)
		return
	}

	// Step 2: compile a Validator from the parsed schemas.
	v := xsdvalidate.New(schemas, nil)

	// Step 3: assess a valid instance.
	valid := `<age>42</age>`
	res, err := xmlsrc.Validate(v, strings.NewReader(valid), "valid.xml")
	if err != nil {
		fmt.Println("parse error:", err)
		return
	}
	fmt.Println("valid instance:", res.Valid())

	// Step 4: assess an invalid instance (negative integer fails positiveInteger).
	invalid := `<age>-7</age>`
	res, err = xmlsrc.Validate(v, strings.NewReader(invalid), "invalid.xml")
	if err != nil {
		fmt.Println("parse error:", err)
		return
	}
	fmt.Println("invalid instance:", res.Valid())

	// Report the first error id (cvc-* spec clause).
	if !res.Valid() {
		ids := xsd.RefIDs(res.Err())
		if len(ids) > 0 {
			fmt.Println("first error id:", ids[0])
		}
	}

	// Output:
	// valid instance: true
	// invalid instance: false
	// first error id: cvc-minInclusive-valid
}
