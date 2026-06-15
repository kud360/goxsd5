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

// wildcardIntersect returns the Attribute Wildcard Intersection of W1 and W2
// per XSD 1.1 Part 1 §3.10.6.3 (cos-aw-intersect). Either argument may be
// nil (no wildcard); nil means "all allowed" for intersection purposes. When
// both are nil the result is nil.
func wildcardIntersect(w1, w2 *xsd.Wildcard) *xsd.Wildcard {
	if w1 == nil {
		return w2
	}
	if w2 == nil {
		return w1
	}
	out := &xsd.Wildcard{}
	// {process contents}: use the stronger of the two (strict > lax > skip),
	// i.e. the minimum of the iota values (ProcessStrict=0 < ProcessLax=1 < ProcessSkip=2).
	out.ProcessContents = min(w1.ProcessContents, w2.ProcessContents)
	// {disallowed names}: union of both NotQName sets (a name disallowed by
	// either wildcard is disallowed by the intersection).
	out.NotQName = notQNameUnion(w1.NotQName, w2.NotQName)
	// {namespace constraint}: per cos-aw-intersect clauses 1–5.
	switch {
	case w1.Mode == xsd.NSConstraintAny && w2.Mode == xsd.NSConstraintAny:
		// clause 1: both any → any
		out.Mode = xsd.NSConstraintAny
	case w1.Mode == xsd.NSConstraintAny:
		// clause 2: one is any → take the other
		out.Mode = w2.Mode
		out.Namespaces = w2.Namespaces
	case w2.Mode == xsd.NSConstraintAny:
		// clause 2 (symmetric)
		out.Mode = w1.Mode
		out.Namespaces = w1.Namespaces
	case w1.Mode == xsd.NSConstraintNot && w2.Mode == xsd.NSConstraintNot:
		// clause 3: Not(S1) ∩ Not(S2) = Not(S1 ∪ S2)
		out.Mode = xsd.NSConstraintNot
		out.Namespaces = stringsUnion(w1.Namespaces, w2.Namespaces)
	case w1.Mode == xsd.NSConstraintNot && w2.Mode == xsd.NSConstraintEnumeration:
		// clause 4: Not(S1) ∩ Enum(S2) = Enum(S2 - S1)
		out.Mode = xsd.NSConstraintEnumeration
		out.Namespaces = stringsDifference(w2.Namespaces, w1.Namespaces)
	case w1.Mode == xsd.NSConstraintEnumeration && w2.Mode == xsd.NSConstraintNot:
		// clause 4 (symmetric)
		out.Mode = xsd.NSConstraintEnumeration
		out.Namespaces = stringsDifference(w1.Namespaces, w2.Namespaces)
	default:
		// clause 5: Enum(S1) ∩ Enum(S2) = Enum(S1 ∩ S2)
		out.Mode = xsd.NSConstraintEnumeration
		out.Namespaces = stringsIntersection(w1.Namespaces, w2.Namespaces)
	}
	return out
}

// wildcardUnion returns the Attribute Wildcard Union of W1 and W2 per XSD 1.1
// Part 1 §3.10.6.2 (cos-aw-union), used to combine an extension's own attribute
// wildcard (W2) with its base's (W1). A nil argument means "no wildcard"; the
// union with a present wildcard is that wildcard. The result's {process
// contents} is taken from W2 (the extension's local wildcard governs).
func wildcardUnion(w1, w2 *xsd.Wildcard) *xsd.Wildcard {
	if w1 == nil {
		return w2
	}
	if w2 == nil {
		return w1
	}
	out := &xsd.Wildcard{ProcessContents: w2.ProcessContents}
	// {disallowed names} (§3.10.6.3): O1's disallowed names not allowed by O2,
	// plus O2's not allowed by O1, plus a keyword only if both carry it.
	out.NotQName = notQNameUnionDisallowed(w1, w2)
	switch {
	case w1.Mode == xsd.NSConstraintAny || w2.Mode == xsd.NSConstraintAny:
		// clause 2: either any → any
		out.Mode = xsd.NSConstraintAny
	case w1.Mode == xsd.NSConstraintNot && w2.Mode == xsd.NSConstraintNot:
		// clause 4: Not(S1) ∪ Not(S2) = Not(S1 ∩ S2); empty exclusion ⇒ any.
		setUnionNot(out, stringsIntersection(w1.Namespaces, w2.Namespaces))
	case w1.Mode == xsd.NSConstraintNot:
		// clause 5: Not(S1) ∪ Enum(S2) = Not(S1 - S2); empty difference ⇒ any.
		setUnionNot(out, stringsDifference(w1.Namespaces, w2.Namespaces))
	case w2.Mode == xsd.NSConstraintNot:
		// clause 5 (symmetric)
		setUnionNot(out, stringsDifference(w2.Namespaces, w1.Namespaces))
	default:
		// clause 3: Enum(S1) ∪ Enum(S2) = Enum(S1 ∪ S2)
		out.Mode = xsd.NSConstraintEnumeration
		out.Namespaces = stringsUnion(w1.Namespaces, w2.Namespaces)
	}
	return out
}

// setUnionNot sets out to Not(ns), or to Any when ns is empty (Not of nothing
// admits every namespace) — the clause 4.1/5.1 degeneration of cos-aw-union.
func setUnionNot(out *xsd.Wildcard, ns []string) {
	if len(ns) == 0 {
		out.Mode = xsd.NSConstraintAny
		return
	}
	out.Mode = xsd.NSConstraintNot
	out.Namespaces = ns
}

// notQNameUnionDisallowed computes the {disallowed names} of the wildcard union
// of w1 and w2 (cos-aw-union): an explicit QName disallowed by one wildcard
// survives only if the other wildcard does not allow it anyway; a keyword
// (##defined/##definedSibling) survives only if both wildcards carry it.
func notQNameUnionDisallowed(w1, w2 *xsd.Wildcard) []xsd.QName {
	var out []xsd.QName
	add := func(q xsd.QName) {
		if !slices.Contains(out, q) {
			out = append(out, q)
		}
	}
	for _, d := range w1.NotQName {
		if isDisallowedKeyword(d) {
			if hasDisallowedKeyword(w2, d.Local) {
				add(d)
			}
		} else if !wildcardAllowsName(w2, d) {
			add(d)
		}
	}
	for _, d := range w2.NotQName {
		if isDisallowedKeyword(d) {
			continue // keyword case already settled from w1's side
		}
		if !wildcardAllowsName(w1, d) {
			add(d)
		}
	}
	return out
}

// isDisallowedKeyword reports whether q is a ##defined/##definedSibling token
// rather than an explicit disallowed QName.
func isDisallowedKeyword(q xsd.QName) bool {
	return q.Namespace == "" && (q.Local == definedKeyword || q.Local == siblingKeyword)
}

// notQNameUnion returns the union of two NotQName slices, deduplicating.
func notQNameUnion(a, b []xsd.QName) []xsd.QName {
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
	out := make([]xsd.QName, len(a))
	copy(out, a)
	for _, q := range b {
		if !slices.Contains(out, q) {
			out = append(out, q)
		}
	}
	return out
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

// stringsUnion returns the union of two string slices, deduplicating.
func stringsUnion(a, b []string) []string {
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
	out := make([]string, len(a))
	copy(out, a)
	for _, s := range b {
		if !slices.Contains(out, s) {
			out = append(out, s)
		}
	}
	return out
}

// stringsDifference returns the elements of a that are not in b.
func stringsDifference(a, b []string) []string {
	var out []string
	for _, s := range a {
		if !slices.Contains(b, s) {
			out = append(out, s)
		}
	}
	return out
}

// stringsIntersection returns the elements common to both a and b.
func stringsIntersection(a, b []string) []string {
	var out []string
	for _, s := range a {
		if slices.Contains(b, s) {
			out = append(out, s)
		}
	}
	return out
}
