package xsdvalidate

import (
	"github.com/kud360/goxsd5/builtin"
	"github.com/kud360/goxsd5/xsd"
	"github.com/kud360/goxsd5/xsdwalk"
)

// Options configures a Validator. The zero value is the default.
type Options struct {
	// AssessAssertions enables xs:assert / xs:alternative evaluation (V4).
	// When false, assertions and conditional type assignment are skipped
	// (the element keeps its declared type). Defaults to enabled.
	DisableAssertions bool
}

// Validator is an immutable, concurrency-safe assessor compiled from a schema
// set. Build it once with New; call Assess for each instance.
type Validator struct {
	schemas []*xsd.Schema
	opts    Options

	elements map[xsd.QName]*xsd.ElementDecl
	types    map[xsd.QName]xsd.Type
	attrs    map[xsd.QName]*xsd.AttributeDecl

	matcher *xsdwalk.Matcher
}

// New compiles a Validator over the given schema set. opts may be nil.
func New(schemas []*xsd.Schema, opts *Options) *Validator {
	v := &Validator{
		schemas:  schemas,
		elements: map[xsd.QName]*xsd.ElementDecl{},
		types:    map[xsd.QName]xsd.Type{},
		attrs:    map[xsd.QName]*xsd.AttributeDecl{},
	}
	if opts != nil {
		v.opts = *opts
	}
	// Seed built-in datatypes and xs:anyType so xsi:type can name them even
	// when a schema's per-namespace component table omits the built-ins.
	for _, t := range builtin.AllBuiltins() {
		v.types[t.Name] = t
	}
	v.types[builtin.AnyType.Name] = builtin.AnyType
	for _, a := range builtin.XSIAttributes {
		v.attrs[a.Name] = a
	}
	for _, s := range schemas {
		for n, e := range s.Elements {
			v.elements[n] = e
		}
		for n, t := range s.Types {
			v.types[n] = t
		}
		for n, a := range s.Attributes {
			v.attrs[n] = a
		}
	}
	v.matcher = &xsdwalk.Matcher{LookupGlobal: v.globalElement}
	return v
}

func (v *Validator) globalElement(name xsd.QName) *xsd.ElementDecl { return v.elements[name] }

func (v *Validator) typeByName(name xsd.QName) xsd.Type { return v.types[name] }

// Assess validates root against the schema set and returns the result.
func (v *Validator) Assess(root Element) *Result {
	a := &assessor{v: v, res: newResult()}
	a.assessRoot(root)
	a.flushKeyrefs()
	a.checkIDRefs()
	return a.res
}
