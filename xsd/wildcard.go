package xsd

import "slices"

// Wildcard name admission (cvc-wildcard, Part 1 §3.10.4). This is the single
// canonical implementation; the parser, the content-model matcher (xsdwalk),
// and the instance assessor (xsdvalidate) all route through it. The split
// mirrors the spec: a static part (namespace constraint + literal disallowed
// names) that needs no context, and the two context-dependent notQName
// keywords (##defined / ##definedSibling) that need a registry to resolve.

// WildcardDefined and WildcardDefinedSibling are the notQName keyword locals
// (Part 1 §3.10.2). Stored in a Wildcard's NotQName entries with an empty
// namespace, they disallow names that resolve to a global declaration, resp.
// to a declaration appearing as a sibling in the content model.
const (
	WildcardDefined        = "##defined"
	WildcardDefinedSibling = "##definedSibling"
)

// AllowsNamespace reports whether ns satisfies w's {namespace constraint},
// ignoring {disallowed names} (cvc-wildcard-namespace, §3.10.4.3).
func (w *Wildcard) AllowsNamespace(ns string) bool {
	switch w.Mode {
	case NSConstraintEnumeration:
		return slices.Contains(w.Namespaces, ns)
	case NSConstraintNot:
		return !slices.Contains(w.Namespaces, ns)
	default: // NSConstraintAny
		return true
	}
}

// AllowsName reports whether w admits the expanded name q under the static part
// of cvc-wildcard-name (§3.10.4.2): the namespace constraint plus the literal
// {disallowed names}. The context-dependent ##defined / ##definedSibling
// keywords are NOT applied — a caller with a registry context uses Allows.
func (w *Wildcard) AllowsName(q QName) bool {
	if !w.AllowsNamespace(q.Namespace) {
		return false
	}
	for _, d := range w.NotQName {
		if d.Namespace == "" && (d.Local == WildcardDefined || d.Local == WildcardDefinedSibling) {
			continue
		}
		if d == q {
			return false
		}
	}
	return true
}

// WildcardContext supplies the registry lookups the context-dependent notQName
// keywords need. Defined answers ##defined (q resolves to a global element or
// attribute declaration); DefinedSibling answers ##definedSibling (q matches a
// declaration appearing in the content model being matched). Either may be nil,
// in which case that keyword never excludes — attribute wildcards leave
// DefinedSibling nil (w-props-correct.5 forbids it on attributes).
type WildcardContext struct {
	Defined        func(QName) bool
	DefinedSibling func(QName) bool
}

// Allows reports whether w admits q under the full cvc-wildcard-name rule,
// including the ##defined / ##definedSibling keyword exclusions resolved through
// ctx, on top of the static AllowsName check.
func (w *Wildcard) Allows(q QName, ctx WildcardContext) bool {
	if !w.AllowsName(q) {
		return false
	}
	for _, d := range w.NotQName {
		if d.Namespace != "" {
			continue
		}
		switch d.Local {
		case WildcardDefined:
			if ctx.Defined != nil && ctx.Defined(q) {
				return false
			}
		case WildcardDefinedSibling:
			if ctx.DefinedSibling != nil && ctx.DefinedSibling(q) {
				return false
			}
		}
	}
	return true
}
