package parser

import (
	"strings"

	"github.com/kud360/goxsd5/parser/xmltree"
	"github.com/kud360/goxsd5/xsd"
)

// SchemaLocationHints returns the schema-document locations a document's root
// element points to through the namespace-qualified xsi:schemaLocation and
// xsi:noNamespaceSchemaLocation attributes (§4.3.2 / cvc-assess). The value of
// xsi:schemaLocation is a whitespace-separated list of (namespace, location)
// pairs; only the location of each pair is returned. Locations are reported
// verbatim (relative to the instance document) in document order, schemaLocation
// pairs first. These are hints a processor MAY use to locate additional schema
// documents for assessment; callers resolve and load them as they see fit.
func SchemaLocationHints(root *xmltree.Node) []string {
	if root == nil {
		return nil
	}
	var out []string
	for i := range root.Attrs {
		a := &root.Attrs[i]
		if a.Name.Space != xsd.XSINS {
			continue
		}
		switch a.Name.Local {
		case "schemaLocation":
			// (namespace location)+ — keep every second field (the locations).
			fields := strings.Fields(a.Value)
			for j := 1; j < len(fields); j += 2 {
				out = append(out, fields[j])
			}
		case "noNamespaceSchemaLocation":
			if loc := strings.TrimSpace(a.Value); loc != "" {
				out = append(out, loc)
			}
		}
	}
	return out
}
