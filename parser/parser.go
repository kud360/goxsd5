package parser

// The public entry point: discover the transitive closure of schema
// documents, run pass 1 per document, register all globals, run pass 2.

import (
	"fmt"

	"github.com/kud360/goxsd5/xsd"
)

// Options configures Parse.
type Options struct {
	// Resolver locates composed schema documents. nil means FileResolver.
	Resolver SchemaResolver
}

// Parse loads the schema document at location plus the transitive closure
// of its imports, includes, redefines, and overrides, and returns the
// linked schemas, one per target namespace (the root document's namespace
// first). The error is an *xsd.ErrorList aggregate of every schema problem
// found; the returned schemas are still usable for inspection alongside a
// non-nil error.
func Parse(location string, opts *Options) ([]*xsd.Schema, error) {
	var resolver SchemaResolver = FileResolver{}
	if opts != nil && opts.Resolver != nil {
		resolver = opts.Resolver
	}
	errs := &xsd.ErrorList{}
	l := newLoader(resolver, errs)
	if err := l.loadRoot(location); err != nil {
		return nil, fmt.Errorf("cannot load %s: %w", location, err)
	}
	return finish(l, errs)
}

// ParseMultiple loads the schema documents at each location in locations plus
// the transitive closure of their imports, includes, redefines, and overrides,
// and returns the linked schemas as a single merged set. The loader deduplicates
// by resolved URI, so a location appearing in multiple entries is loaded once.
// The error is an *xsd.ErrorList aggregate; the returned schemas are still
// usable alongside a non-nil error.
func ParseMultiple(locations []string, opts *Options) ([]*xsd.Schema, error) {
	var resolver SchemaResolver = FileResolver{}
	if opts != nil && opts.Resolver != nil {
		resolver = opts.Resolver
	}
	errs := &xsd.ErrorList{}
	l := newLoader(resolver, errs)
	for _, loc := range locations {
		if err := l.loadRoot(loc); err != nil {
			return nil, fmt.Errorf("cannot load %s: %w", loc, err)
		}
	}
	return finish(l, errs)
}

// finish registers and builds everything the loader discovered.
func finish(l *loader, errs *xsd.ErrorList) ([]*xsd.Schema, error) {
	reg := newRegistry()
	l.register(reg)
	schemas := buildSchemas(reg, l, errs)
	return schemas, errs.Err()
}
