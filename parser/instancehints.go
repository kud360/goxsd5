package parser

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/kud360/goxsd5/parser/xmltree"
	"github.com/kud360/goxsd5/xsd"
)

// SchemaLocationHints returns the schema-document locations a document's root
// element points to through the namespace-qualified xsi:schemaLocation and
// xsi:noNamespaceSchemaLocation attributes (§4.3.2 / cvc-assess). The value of
// xsi:schemaLocation is a whitespace-separated list of (namespace, location)
// pairs; only the location of each pair is returned. Locations are reported
// verbatim (relative to the instance document's directory) in document order,
// schemaLocation pairs first. Only root-element hints are read; xml:base and
// descendant-element hints (§4.3.2 items 4–5) are not honoured. These are hints
// a processor MAY use to locate additional schema documents for assessment;
// callers resolve and load them as they see fit.
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

// HintAugmentedPaths returns basePaths extended with any schema documents the
// instance at instancePath names via xsi:schemaLocation /
// xsi:noNamespaceSchemaLocation (SchemaLocationHints) that are not already in
// basePaths. Each hint is resolved relative to the instance file's directory and
// path-cleaned with filepath.Join — the same basis the loader dedupes roots on,
// so the dedup decision here matches the loader's. added reports whether any
// hint extended the set. When the instance cannot be opened or parsed as XML the
// call is fail-open: basePaths is returned unchanged with added false.
func HintAugmentedPaths(basePaths []string, instancePath string) (paths []string, added bool) {
	f, err := os.Open(instancePath)
	if err != nil {
		return basePaths, false
	}
	root, perr := xmltree.Parse(f, instancePath)
	f.Close()
	if perr != nil {
		return basePaths, false
	}
	have := map[string]bool{}
	for _, p := range basePaths {
		have[p] = true
	}
	paths = append([]string{}, basePaths...)
	for _, loc := range SchemaLocationHints(root) {
		p := filepath.Join(filepath.Dir(instancePath), filepath.FromSlash(loc))
		if have[p] {
			continue
		}
		have[p] = true
		paths = append(paths, p)
		added = true
	}
	return paths, added
}

// AugmentSchemasFromHints returns the schema set to assess the instance at
// instancePath against: base, extended with any schema documents the instance
// names via xsi:schemaLocation hints (resolved relative to the instance file's
// directory) that basePaths does not already cover. basePaths must be the source
// locations base was parsed from. When hints genuinely extend the set the whole
// set is re-parsed via ParseMultiple; otherwise — and on any read, parse, or
// re-parse error — base is returned unchanged (fail-open).
func AugmentSchemasFromHints(base []*xsd.Schema, basePaths []string, instancePath string, opts *Options) []*xsd.Schema {
	paths, added := HintAugmentedPaths(basePaths, instancePath)
	if !added {
		return base
	}
	augmented, err := ParseMultiple(paths, opts)
	if err != nil {
		return base
	}
	return augmented
}
