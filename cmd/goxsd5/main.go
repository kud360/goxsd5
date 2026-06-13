// Command goxsd5 parses an XSD 1.1 schema document (following its imports,
// includes, redefines, and overrides) and prints a per-namespace summary of
// the resulting components. It exits non-zero if the schema has any errors,
// printing each as `uri:line:col: [constraint-id] message`.
//
// Usage:
//
//	goxsd5 [-q] <schema.xsd>
//
//	-q  quiet: print only errors, suppress the component summary.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/kud360/goxsd5/parser"
	"github.com/kud360/goxsd5/xsd"
)

func main() {
	quiet := flag.Bool("q", false, "print only errors, no component summary")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: goxsd5 [-q] <schema.xsd>")
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
		fmt.Fprintf(os.Stderr, "%d error(s)\n", len(errs))
		os.Exit(1)
	}
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
