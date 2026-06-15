package xsdvalidate

import "github.com/kud360/goxsd5/xsd"

// Result is the outcome of an assessment: the accumulated validation errors
// (each an *xsd.Error carrying a cvc-* SpecRef) plus a PSVI-lite record of the
// type assigned to each assessed element. It is the interpreter's analog of the
// post-schema-validation infoset; identity-constraint and assertion evaluation
// read from it, and a future codegen back-end can populate the same shape.
type Result struct {
	errs *xsd.ErrorList
	// Types records the governing type assigned to each element node (by
	// pointer identity of the Element), for PSVI consumers.
	Types map[Element]xsd.Type
}

// Valid reports whether the instance is schema-valid (no errors accumulated).
func (r *Result) Valid() bool { return r.errs.Empty() }

// Err returns the accumulated errors joined, or nil.
func (r *Result) Err() error { return r.errs.Err() }

// Errors returns the flat list of accumulated errors.
func (r *Result) Errors() []error { return xsd.AllErrors(r.errs.Err()) }

func newResult() *Result {
	return &Result{errs: &xsd.ErrorList{}, Types: map[Element]xsd.Type{}}
}
