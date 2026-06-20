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

	// Primitives, when non-nil, replaces the built-in simple types the registry
	// is seeded with — the value-space layer the schema's types resolve their
	// lexical→value mapping and comparators through. It is how a caller opts into
	// an alternative value semantics (e.g. builtin/gotype's Go-native, lax set)
	// at parse time. nil means the default strict built-ins (builtin.AllBuiltins),
	// so the default behaviour is unchanged. The slice must supply xs:anySimpleType
	// and xs:anyAtomicType; gotype.AllBuiltins does.
	Primitives []*xsd.SimpleType
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
	return finish(l, optPrimitives(opts), errs)
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
	return finish(l, optPrimitives(opts), errs)
}

// optPrimitives returns the custom primitive set from opts, or nil when none is
// configured (the default strict built-ins).
func optPrimitives(opts *Options) []*xsd.SimpleType {
	if opts == nil {
		return nil
	}
	return opts.Primitives
}

// finish registers and builds everything the loader discovered. primitives, when
// non-nil, seeds the registry with an alternative value-space layer in place of
// the default built-ins.
func finish(l *loader, primitives []*xsd.SimpleType, errs *xsd.ErrorList) ([]*xsd.Schema, error) {
	reg := newRegistry(primitives)
	l.register(reg)
	schemas := buildSchemas(reg, l, errs)
	return schemas, errs.Err()
}
