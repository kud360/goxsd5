// Package xsdwalk is the shared model algebra over the compiled xsd model:
// the content-model matcher (automaton.go) plus the model queries used to
// assess an instance — type-derivation validity, substitution-group
// acceptance, wildcard namespace matching, and attribute-use lookup.
//
// It depends only on the pure-leaf xsd package. It deliberately carries no
// instance/infoset knowledge: callers (xsdvalidate today, codegen later) pass
// in already-resolved QNames and types. Push (exhaustive, schema-only) and
// pull (instance-guided) drivers both reuse this algebra; the reusable core
// is the algebra, not the driver.
package xsdwalk

import "github.com/kud360/goxsd5/xsd"

// TypeEq reports whether two type definitions are the same component. Types are
// interned in the registry, so pointer identity is the common case; the QName
// fallback covers the built-in pointers reached through different packages.
func TypeEq(a, b xsd.Type) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	an, bn := a.TypeName(), b.TypeName()
	return !an.IsZero() && an == bn
}

// derivationMethod is the single derivation step by which t was derived from
// its base, as a Derivation flag (for blocking checks).
func derivationMethod(t xsd.Type) xsd.Derivation {
	switch t := t.(type) {
	case *xsd.ComplexType:
		return t.DerivationMethod
	case *xsd.SimpleType:
		switch t.Variety {
		case xsd.VarietyList:
			return xsd.DeriveList
		case xsd.VarietyUnion:
			return xsd.DeriveUnion
		default:
			return xsd.DeriveRestriction
		}
	}
	return 0
}

// DerivationOK reports whether the type D is validly derived from B, with no
// step using a method in the blocked set (cos-ct-derived-ok / cos-st-derived-ok,
// Part 1 §3.4.6 / §3.16.6 — the engine behind cvc-elt.4.3's xsi:type check).
// D==B is trivially OK. The walk climbs D's base chain to B; any blocked step
// fails. A simple type may be derived from a complex type's simple content,
// so the chain crosses the Simple/Complex boundary via Base().
func DerivationOK(d, b xsd.Type, blocked xsd.DerivationSet) bool {
	if TypeEq(d, b) {
		return true
	}
	for cur := d; cur != nil; {
		if blocked.Has(derivationMethod(cur)) {
			return false
		}
		base := cur.Base()
		if base == nil {
			return false
		}
		if TypeEq(base, b) {
			return true
		}
		// Guard against a self-referential base (xs:anyType.Base()==nil ends it).
		if TypeEq(base, cur) {
			return false
		}
		cur = base
	}
	return false
}

// IsDerivedFrom reports whether d is the same as, or transitively derived
// from, b by any method (no blocking).
func IsDerivedFrom(d, b xsd.Type) bool {
	for cur := d; cur != nil; cur = cur.Base() {
		if TypeEq(cur, b) {
			return true
		}
		if TypeEq(cur.Base(), cur) {
			break
		}
	}
	return false
}

// SubstitutableFor reports whether the element declaration member may appear
// where head is expected, via substitution-group membership (cos-equiv-class,
// Part 1 §3.3.6). member==head is the trivial case. Otherwise member must be
// transitively in head's substitution group, every link's type derivation must
// be permitted by head's {substitution group exclusions} (Final/block), and
// no intervening head may block substitution.
//
// blocked is head's {disallowed substitutions}; DeriveSubstitution in it blocks
// all substitution. The per-type final exclusions are checked against the
// derivation method relating member's type to the intervening head's type.
func SubstitutableFor(member, head *xsd.ElementDecl, blocked xsd.DerivationSet) bool {
	if member == head {
		return true
	}
	if blocked.Has(xsd.DeriveSubstitution) {
		return false
	}
	return substChain(member, head, map[*xsd.ElementDecl]bool{})
}

// substChain walks member's declared substitution-group heads toward target,
// validating each link's type derivation against the head's {final} exclusions.
func substChain(member, target *xsd.ElementDecl, seen map[*xsd.ElementDecl]bool) bool {
	if seen[member] {
		return false
	}
	seen[member] = true
	for _, head := range member.SubstitutionGroups {
		// The substitution requires member's type validly derived from head's
		// type, with head's {final}+block excluding the methods used.
		exclude := head.Final | head.Block
		if head.Block.Has(xsd.DeriveSubstitution) {
			continue
		}
		if !DerivationOK(declType(member), declType(head), exclude) {
			continue
		}
		if head == target {
			return true
		}
		if substChain(head, target, seen) {
			return true
		}
	}
	return false
}

func declType(e *xsd.ElementDecl) xsd.Type {
	if e == nil {
		return nil
	}
	return e.Type
}

// WildcardAllows reports whether the wildcard w admits an element/attribute of
// the given expanded name (the namespace constraint plus 1.1 notQName, Part 1
// §3.10.4 cvc-wildcard-namespace). The ##defined / ##definedSibling notQName
// keywords are context-dependent and handled by the caller; here only literal
// disallowed names are honored.
func WildcardAllows(w *xsd.Wildcard, q xsd.QName) bool {
	if !namespaceAllowed(w, q.Namespace) {
		return false
	}
	for _, d := range w.NotQName {
		if d.Namespace == "" && (d.Local == "##defined" || d.Local == "##definedSibling") {
			continue
		}
		if d == q {
			return false
		}
	}
	return true
}

func namespaceAllowed(w *xsd.Wildcard, ns string) bool {
	switch w.Mode {
	case xsd.NSConstraintAny:
		return true
	case xsd.NSConstraintEnumeration:
		return contains(w.Namespaces, ns)
	case xsd.NSConstraintNot:
		return !contains(w.Namespaces, ns)
	}
	return true
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// AttributeUse finds the attribute use in uses whose declaration matches name,
// or nil. Prohibited uses are represented as absent in the compiled model, so
// every entry here is a real use.
func AttributeUse(uses []*xsd.AttributeUse, name xsd.QName) *xsd.AttributeUse {
	for _, u := range uses {
		if u.Decl != nil && u.Decl.Name == name {
			return u
		}
	}
	return nil
}
