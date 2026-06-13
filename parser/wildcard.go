package parser

// Wildcard namespace-constraint relations (Part 1 §3.10.6): the Wildcard
// Subset relation (cos-ns-subset) and the Wildcard allows Expanded Name rule
// (cvc-wildcard-name) it depends on. Used by particle restriction
// (NSSubset/NSCompat) and by the attribute-wildcard subset check that
// complexContent/simpleContent restrictions must satisfy.

import (
	"slices"

	"github.com/kud360/goxsd5/xsd"
)

// definedKeyword and siblingKeyword are how a notQName ##defined /
// ##definedSibling token is stored in a wildcard's {disallowed names}: a QName
// with an empty namespace and the literal token in Local.
const (
	definedKeyword = "##defined"
	siblingKeyword = "##definedSibling"
)

// namespaceConstraintSubset reports whether wildcard sub's namespace
// constraint is a wildcard subset of super's, per Wildcard Subset
// (cos-ns-subset, §3.10.6.2): sub validates a subset of the names super does.
func namespaceConstraintSubset(sub, super *xsd.Wildcard) bool {
	if !namespaceVarietySubset(sub, super) {
		return false
	}
	// In all variety cases the disallowed-name conditions must also hold: every
	// name super disallows, sub must disallow too (else sub would allow a name
	// super does not).
	for _, d := range super.NotQName {
		switch d.Local {
		case definedKeyword:
			if !hasDisallowedKeyword(sub, definedKeyword) {
				return false
			}
		case siblingKeyword:
			if !hasDisallowedKeyword(sub, siblingKeyword) {
				return false
			}
		default:
			if wildcardAllowsName(sub, d) {
				return false
			}
		}
	}
	return true
}

// namespaceVarietySubset implements the {variety}/{namespaces} half of
// cos-ns-subset (clauses 1–4), ignoring {disallowed names}.
func namespaceVarietySubset(sub, super *xsd.Wildcard) bool {
	switch {
	case super.Mode == xsd.NSConstraintAny:
		return true // clause 1
	case sub.Mode == xsd.NSConstraintEnumeration && super.Mode == xsd.NSConstraintEnumeration:
		// clause 2: super's namespaces ⊇ sub's.
		return stringsSubset(sub.Namespaces, super.Namespaces)
	case sub.Mode == xsd.NSConstraintEnumeration && super.Mode == xsd.NSConstraintNot:
		// clause 3: the two namespace sets are disjoint.
		return !namespacesIntersect(sub.Namespaces, super.Namespaces)
	case sub.Mode == xsd.NSConstraintNot && super.Mode == xsd.NSConstraintNot:
		// clause 4: super's namespaces ⊆ sub's.
		return stringsSubset(super.Namespaces, sub.Namespaces)
	default:
		// sub = any with super ≠ any, or sub = not with super = enumeration:
		// sub admits names super does not.
		return false
	}
}

// wildcardAllowsName reports whether w allows the expanded name q, per
// Wildcard allows Expanded Name (cvc-wildcard-name, §3.10.4.2), for an explicit
// QName q (the ##defined/##definedSibling keywords are handled by the caller).
func wildcardAllowsName(w *xsd.Wildcard, q xsd.QName) bool {
	if !namespaceAllowed(w, q.Namespace) {
		return false
	}
	for _, d := range w.NotQName {
		if d.Local != definedKeyword && d.Local != siblingKeyword && d == q {
			return false
		}
	}
	return true
}

// hasDisallowedKeyword reports whether w's notQName carries the given keyword.
func hasDisallowedKeyword(w *xsd.Wildcard, kw string) bool {
	for _, d := range w.NotQName {
		if d.Local == kw && d.Namespace == "" {
			return true
		}
	}
	return false
}

// stringsSubset reports whether every member of a is in b.
func stringsSubset(a, b []string) bool {
	for _, x := range a {
		if !slices.Contains(b, x) {
			return false
		}
	}
	return true
}
