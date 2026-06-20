// Command goxsd5 parses an XSD 1.1 schema document (following its imports,
// includes, redefines, and overrides) and prints a per-namespace summary of
// the resulting components. It exits non-zero if the schema has any errors,
// printing each as `uri:line:col: [constraint-id] message`.
//
// Usage:
//
//	goxsd5 [-q] [-validate doc.xml] <schema.xsd>
//
//	-q              quiet: print only errors, suppress the component summary.
//	-validate FILE  also assess the instance document FILE against the schema
//	                (XSD 1.1 schema-validity assessment) and report cvc-* errors.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/kud360/goxsd5/parser"
	"github.com/kud360/goxsd5/parser/xmltree"
	"github.com/kud360/goxsd5/xsd"
	"github.com/kud360/goxsd5/xsdvalidate"
	"github.com/kud360/goxsd5/xsdvalidate/xmlsrc"
)

func main() {
	quiet := flag.Bool("q", false, "print only errors, no component summary")
	validate := flag.String("validate", "", "assess this instance document against the schema")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: goxsd5 [-q] [-validate doc.xml] <schema.xsd>")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	schemas, err := parser.Parse(flag.Arg(0), nil)
	if !*quiet {
		printSummary(os.Stdout, schemas)
	}
	if err != nil {
		errs := xsd.AllErrors(err)
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, e)
		}
		fmt.Fprintf(os.Stderr, "%d schema error(s)\n", len(errs))
		os.Exit(1)
	}

	if *validate != "" {
		os.Exit(assessInstance(schemas, flag.Arg(0), *validate, *quiet))
	}
}

// assessInstance validates one instance document and returns the process exit
// code (0 valid, 1 invalid or unreadable). schemaPath is the explicitly
// supplied schema location; it is used to resolve xsi:schemaLocation hints
// from the instance and to merge any additional schema documents the instance
// references that are not already covered.
func assessInstance(schemas []*xsd.Schema, schemaPath, path string, quiet bool) int {
	schemas = augmentSchemasFromHints(schemas, schemaPath, path)
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer f.Close()
	v := xsdvalidate.New(schemas, nil)
	res, err := xmlsrc.Validate(v, f, path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
		return 1
	}
	if res.Valid() {
		if !quiet {
			fmt.Printf("%s: valid\n", path)
		}
		return 0
	}
	errs := res.Errors()
	for _, e := range errs {
		fmt.Fprintln(os.Stderr, e)
	}
	fmt.Fprintf(os.Stderr, "%s: invalid (%d error(s))\n", path, len(errs))
	return 1
}

// augmentSchemasFromHints reads the instance at instancePath, extracts any
// xsi:schemaLocation / xsi:noNamespaceSchemaLocation hints, resolves them
// relative to the instance file's directory, and re-parses via
// ParseMultiple when the hints name documents beyond schemaPath. When no
// hints extend the set the original schemas are returned unchanged.
func augmentSchemasFromHints(schemas []*xsd.Schema, schemaPath, instancePath string) []*xsd.Schema {
	f, err := os.Open(instancePath)
	if err != nil {
		return schemas
	}
	root, err := xmltree.Parse(f, instancePath)
	f.Close()
	if err != nil {
		return schemas
	}

	absSchema, err := filepath.Abs(schemaPath)
	if err != nil {
		return schemas
	}
	have := map[string]bool{absSchema: true}
	locations := []string{schemaPath}
	added := false
	for _, loc := range parser.SchemaLocationHints(root) {
		abs, err := filepath.Abs(filepath.Join(filepath.Dir(instancePath), loc))
		if err != nil {
			continue
		}
		if have[abs] {
			continue
		}
		have[abs] = true
		locations = append(locations, filepath.Join(filepath.Dir(instancePath), loc))
		added = true
	}
	if !added {
		return schemas
	}
	augmented, err := parser.ParseMultiple(locations, nil)
	if err != nil {
		return schemas
	}
	return augmented
}

// printSummary lists each namespace's component counts and named components.
func printSummary(w *os.File, schemas []*xsd.Schema) {
	for _, s := range schemas {
		ns := s.TargetNamespace
		if ns == "" {
			ns = "(no namespace)"
		}
		fmt.Fprintf(w, "namespace %s  [%s]\n", ns, s.Location)
		fmt.Fprintf(w, "  types: %d  elements: %d  attributes: %d  groups: %d  attributeGroups: %d  notations: %d\n",
			len(s.Types), len(s.Elements), len(s.Attributes), len(s.Groups), len(s.AttributeGroups), len(s.Notations))
		for _, name := range sortedTypeNames(s.Types) {
			fmt.Fprintf(w, "    type %s — %s\n", name, typeKind(s.Types[name]))
		}
		for _, name := range sortedElemNames(s.Elements) {
			fmt.Fprintf(w, "    element %s\n", name)
		}
	}
}

func typeKind(t xsd.Type) string {
	switch t := t.(type) {
	case *xsd.SimpleType:
		return "simple/" + t.Variety.String()
	case *xsd.ComplexType:
		return "complex"
	}
	return "?"
}

func sortedTypeNames(m map[xsd.QName]xsd.Type) []xsd.QName {
	names := make([]xsd.QName, 0, len(m))
	for n := range m {
		names = append(names, n)
	}
	sortQNames(names)
	return names
}

func sortedElemNames(m map[xsd.QName]*xsd.ElementDecl) []xsd.QName {
	names := make([]xsd.QName, 0, len(m))
	for n := range m {
		names = append(names, n)
	}
	sortQNames(names)
	return names
}

func sortQNames(names []xsd.QName) {
	sort.Slice(names, func(i, j int) bool {
		if names[i].Namespace != names[j].Namespace {
			return names[i].Namespace < names[j].Namespace
		}
		return names[i].Local < names[j].Local
	})
}
