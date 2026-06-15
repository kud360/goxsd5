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
	// cos-st-derived-ok clause 2.2.4: when b is a union, d is validly derived
	// from b if it is validly derived from a type in b's transitive membership —
	// provided b and every intervening union have empty {facets} (2.2.4.3). This
	// lets xsi:type name a member of a union-typed element's declared type, but
	// not reach through a union that was narrowed by restriction.
	if bs, ok := b.(*xsd.SimpleType); ok && bs.Variety == xsd.VarietyUnion {
		if derivedFromUnionMember(d, bs, blocked) {
			return true
		}
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

// derivedFromUnionMember implements cos-st-derived-ok clause 2.2.4: d is validly
// derived from some type in union u's transitive membership, given u and every
// intervening union have empty {facets} (2.2.4.3). A union with constraining
// facets (it was narrowed by restriction) blocks the traversal through it.
func derivedFromUnionMember(d xsd.Type, u *xsd.SimpleType, blocked xsd.DerivationSet) bool {
	if !unionFacetsEmpty(u) {
		return false
	}
	for _, m := range u.DirectMembers {
		if m.Variety == xsd.VarietyUnion {
			// An intervening union: only traversable if its facets are empty too.
			if derivedFromUnionMember(d, m, blocked) {
				return true
			}
			continue
		}
		if DerivationOK(d, m, blocked) {
			return true
		}
	}
	return false
}

// unionFacetsEmpty reports whether u carries no value-space-narrowing facet
// (pattern, enumeration, or assertion — the constraining facets applicable to
// the union variety); whiteSpace=collapse is the fixed default and does not
// count. This is the 2.2.4.3 "{facets} … is empty" test.
func unionFacetsEmpty(u *xsd.SimpleType) bool {
	f := u.EffectiveFacets()
	if f == nil {
		return true
	}
	return len(f.PatternGroups) == 0 && len(f.Enumeration) == 0 && len(f.Assertions) == 0
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
		// type, with the head element's {final}+{disallowed substitutions} AND the
		// head type's {prohibited substitutions} excluding the methods used
		// (cvc-substitution / §3.3.6.3).
		exclude := head.Final | head.Block | typeBlock(declType(head))
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

// typeBlock returns a type's {prohibited substitutions}; only complex types
// carry one (simple types have no block).
func typeBlock(t xsd.Type) xsd.DerivationSet {
	if ct, ok := t.(*xsd.ComplexType); ok {
		return ct.Block
	}
	return 0
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
