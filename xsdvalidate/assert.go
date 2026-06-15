package xsdvalidate

import "github.com/kud360/goxsd5/xsd"

// Assertions (cvc-assertion, Part 1 §3.13.4) and conditional type assignment
// (xs:alternative, §3.12) are evaluated with the partial XPath subset in
// xpath.go. Both fail open: an assertion whose test cannot be evaluated is
// treated as satisfied, and a type table that cannot be fully evaluated falls
// back to the element's declared type — so V4 only ever ratchets coverage up and
// never rejects an instance for an unsupported expression.

// withInherited returns an Element view of el that also exposes the inherited
// attributes (XSD 1.1), so a conditional-type-assignment test can read an
// attribute value contributed by an ancestor. The element's own attributes take
// precedence. When there is nothing to inherit, el is returned unchanged.
func withInherited(el Element, inherited map[xsd.QName]string) Element {
	if len(inherited) == 0 {
		return el
	}
	return inhElement{Element: el, inh: inherited}
}

type inhElement struct {
	Element
	inh map[xsd.QName]string
}

func (e inhElement) Attributes() []Attribute {
	own := e.Element.Attributes()
	have := make(map[xsd.QName]bool, len(own))
	for _, a := range own {
		have[a.Name()] = true
	}
	out := own
	for n, v := range e.inh {
		if !have[n] {
			out = append(out, inhAttr{n, v})
		}
	}
	return out
}

type inhAttr struct {
	n xsd.QName
	v string
}

func (a inhAttr) Name() xsd.QName { return a.n }
func (a inhAttr) Value() string   { return a.v }
func (a inhAttr) Pos() xsd.Pos    { return xsd.Pos{} }

// checkAssertions evaluates a complex type's xs:assert children against the
// element being assessed (the assertion's context node).
func (a *assessor) checkAssertions(el Element, ct *xsd.ComplexType) {
	if a.v.opts.DisableAssertions {
		return
	}
	for _, as := range ct.Assertions {
		result, ok := evalAssertion(el, as.Test)
		if ok && !result {
			// spec: cvc-assertion — XSD 1.1 Part 1 §3.13.4 (xmlschema11-1.md#cvc-assertion)
			a.addf(xsd.SpecCvcAssertion, el.Pos(), "assertion %q is not satisfied", as.Test)
		}
	}
}

// selectGoverningType applies conditional type assignment: the first
// xs:alternative whose test holds selects the type; an alternative with no test
// is the default. If any preceding test cannot be evaluated, the whole table is
// abandoned in favour of the declared type (conservative: identical to the
// pre-CTA behaviour, so no regression).
func (a *assessor) selectGoverningType(el Element, decl *xsd.ElementDecl, declared xsd.Type, inherited map[xsd.QName]string) xsd.Type {
	if a.v.opts.DisableAssertions || len(decl.TypeAlternatives) == 0 {
		return declared
	}
	ctx := withInherited(el, inherited)
	for _, alt := range decl.TypeAlternatives {
		if alt.Test == "" {
			if alt.Type != nil {
				return alt.Type
			}
			return declared
		}
		r, ok := evalAssertion(ctx, alt.Test)
		if !ok {
			return declared
		}
		if r {
			if alt.Type != nil {
				return alt.Type
			}
			return declared
		}
	}
	return declared
}
