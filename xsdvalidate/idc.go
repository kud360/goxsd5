package xsdvalidate

import (
	"strings"

	"github.com/kud360/goxsd5/xsd"
)

// Identity-constraint evaluation (cvc-identity-constraint, Part 1 §3.11.4).
//
// The selector/field XPath subset (§3.11.6) is evaluated directly over the
// infoset. One deliberate, documented approximation keeps this independent of
// schema-document state the compiled model does not retain: name tests match by
// LOCAL NAME only. The selector/field expressions resolve prefixes against the
// schema document's in-scope namespaces, which the IdentityConstraint component
// does not carry; local-name matching handles prefixed and unprefixed forms
// uniformly and is sound for the single-namespace cases that dominate the
// corpus. Field values are compared in the value space (via each field node's
// type), falling back to the collapsed lexical form only when a node's type is
// unavailable — so equal-but-differently-written keys (5 vs 5.0) match.

// fieldVal is one resolved field of a key tuple: its value-space value plus the
// type whose comparator governs equality, and the collapsed lexical fallback.
type fieldVal struct {
	v xsd.Value
	t *xsd.SimpleType
	s string
}

// pendingKeyref is one keyref check deferred to the post-walk pass: its evaluated
// tuples and the scope element they were read at (for error positioning).
type pendingKeyref struct {
	ic     *xsd.IdentityConstraint
	tuples [][]fieldVal
	pos    xsd.Pos
}

// checkIdentityConstraints evaluates every identity constraint declared on the
// element decl, scoped at the element node el. Key/unique tables are recorded in
// the document-scoped a.keyTables so a cross-scope keyref can resolve against a
// key declared on another element; keyref checks themselves are deferred to
// flushKeyrefs, run after the full walk when every key table is complete.
func (a *assessor) checkIdentityConstraints(el Element, decl *xsd.ElementDecl) {
	if len(decl.IdentityConstraints) == 0 {
		return
	}
	for _, ic := range decl.IdentityConstraints {
		if ic.Category == xsd.ICKeyref {
			continue
		}
		if a.keyTables == nil {
			a.keyTables = map[*xsd.IdentityConstraint][][]fieldVal{}
		}
		a.keyTables[ic] = a.evalConstraint(el, ic)
	}
	for _, ic := range decl.IdentityConstraints {
		if ic.Category != xsd.ICKeyref || ic.Refer == nil {
			continue
		}
		a.pendingKeyrefs = append(a.pendingKeyrefs, pendingKeyref{
			ic:     ic,
			tuples: a.evalConstraint(el, ic),
			pos:    el.Pos(),
		})
	}
}

// flushKeyrefs checks every deferred keyref against the document-scoped key
// tables, which are complete once the element walk has finished. A keyref whose
// referred key table is still absent — the referred key is scoped at a
// descendant, a schema-design error rather than our check — is left fail-open,
// preserving prior behaviour for that edge case.
func (a *assessor) flushKeyrefs() {
	for _, pk := range a.pendingKeyrefs {
		keyTuples, ok := a.keyTables[pk.ic.Refer]
		if !ok {
			continue
		}
		for _, t := range pk.tuples {
			if !tupleInSet(t, keyTuples) {
				// spec: cvc-identity-constraint — XSD 1.1 Part 1 §3.11.4 (cvc-identity-constraint.4.3)
				a.addf(xsd.SpecCvcIdentity, pk.pos, "keyref %s value has no matching key in %s", pk.ic.Name, pk.ic.Refer.Name)
			}
		}
	}
}

// evalConstraint selects the target node set and builds the tuple list for one
// constraint, reporting field-cardinality, missing-key, and uniqueness errors.
func (a *assessor) evalConstraint(el Element, ic *xsd.IdentityConstraint) [][]fieldVal {
	targets := a.selectTargets(el, ic.Selector)
	var tuples [][]fieldVal
	for _, target := range targets {
		tuple := make([]fieldVal, len(ic.Fields))
		missing, bad := false, false
		for i, f := range ic.Fields {
			nodes := a.selectFieldNodes(target, f)
			switch {
			case len(nodes) > 1:
				// spec: cvc-identity-constraint — §3.11.4 (cvc-identity-constraint.3)
				a.addf(xsd.SpecCvcIdentity, el.Pos(), "field of %s selects more than one node", ic.Name)
				bad = true
			case len(nodes) == 0:
				missing = true
			default:
				tuple[i] = a.fieldValue(nodes[0])
			}
		}
		if bad {
			continue
		}
		if missing {
			if ic.Category == xsd.ICKey {
				// spec: cvc-identity-constraint — §3.11.4 (cvc-identity-constraint.4.2.1)
				a.addf(xsd.SpecCvcIdentity, el.Pos(), "key %s is missing a field value", ic.Name)
			}
			continue
		}
		if ic.Category != xsd.ICKeyref {
			if tupleInSet(tuple, tuples) {
				// spec: cvc-identity-constraint — §3.11.4 (cvc-identity-constraint.4.1/4.2.2)
				a.addf(xsd.SpecCvcIdentity, el.Pos(), "duplicate %s value", ic.Name)
				continue
			}
		}
		tuples = append(tuples, tuple)
	}
	return tuples
}

// fieldValue resolves a field node to its comparable value.
func (a *assessor) fieldValue(n fieldNode) fieldVal {
	var lexical string
	var t *xsd.SimpleType
	if n.at != nil {
		lexical = n.at.Value()
		t = a.attrType[n.at]
	} else {
		lexical = charContent(n.el)
		t = simpleOf(a.res.Types[n.el])
	}
	fv := fieldVal{s: collapse(lexical)}
	if t != nil {
		if v, err := t.ParseValue(lexical, nodeNSContext(n)); err == nil {
			fv.v, fv.t = v, t
		}
	}
	return fv
}

func nodeNSContext(n fieldNode) nsContext {
	if n.el != nil {
		return nsContext{n.el}
	}
	return nsContext{n.owner}
}

// fieldsEqual compares two resolved fields: in the value space when both carry a
// type, by collapsed lexical form otherwise.
// normSingleton unwraps a singleton list value to its single member, so an
// atomic value can compare equal to a one-item list (XSD 1.1 §3.11.4 — "an
// atomic value can be equal to a singleton list").
func normSingleton(fv fieldVal) fieldVal {
	if lv, ok := fv.v.(xsd.ListValue); ok && len(lv) == 1 &&
		fv.t != nil && fv.t.Variety == xsd.VarietyList && fv.t.ItemType != nil {
		return fieldVal{v: lv[0], t: fv.t.ItemType, s: fv.s}
	}
	return fv
}

func fieldsEqual(a, b fieldVal) bool {
	a, b = normSingleton(a), normSingleton(b)
	if a.t != nil && b.t != nil {
		// Values from different value spaces are incomparable, hence unequal —
		// string "3.0" and decimal 3.0 do not conflict. Only fall back to the
		// lexical form when a node's type was unavailable.
		o, ok := a.t.EffectiveCompare()(a.v, b.v)
		return ok && o == xsd.OrderEqual
	}
	return a.s == b.s
}

func tupleEqual(a, b []fieldVal) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !fieldsEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

func tupleInSet(t []fieldVal, set [][]fieldVal) bool {
	for _, x := range set {
		if tupleEqual(t, x) {
			return true
		}
	}
	return false
}

func simpleOf(t xsd.Type) *xsd.SimpleType {
	switch t := t.(type) {
	case *xsd.SimpleType:
		return t
	case *xsd.ComplexType:
		if sc, ok := t.Content.(*xsd.SimpleContent); ok {
			return sc.Type
		}
	}
	return nil
}

// ---- restricted selector/field XPath subset ----

type icStepKind int

const (
	stepSelf     icStepKind = iota // "."
	stepName                       // child element by local name
	stepAny                        // child element wildcard "*" or "p:*"
	stepAttrName                   // "@local"
	stepAttrAny                    // "@*"
)

type icStep struct {
	kind  icStepKind
	local string
}

// icPath is one alternative of a selector/field expression.
type icPath struct {
	descendant bool // ".//" prefix: start from self-or-descendant
	steps      []icStep
}

// fieldNode is a node selected by a field: an element or an attribute. owner is
// the element carrying the attribute (for namespace context).
type fieldNode struct {
	el    Element
	at    Attribute
	owner Element
}

func parsePaths(expr string) []icPath {
	var out []icPath
	for _, alt := range strings.Split(expr, "|") {
		out = append(out, parsePath(alt))
	}
	return out
}

func parsePath(alt string) icPath {
	s := strings.TrimSpace(alt)
	p := icPath{}
	switch {
	case strings.HasPrefix(s, ".//"):
		p.descendant = true
		s = s[3:]
	case strings.HasPrefix(s, "//"):
		p.descendant = true
		s = s[2:]
	}
	for _, part := range strings.Split(s, "/") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		var st icStep
		switch {
		case part == ".":
			st = icStep{kind: stepSelf}
		case part == "*":
			st = icStep{kind: stepAny}
		case strings.HasPrefix(part, "@"):
			a := part[1:]
			if a == "*" {
				st = icStep{kind: stepAttrAny}
			} else {
				st = icStep{kind: stepAttrName, local: localPart(a)}
			}
		case strings.HasSuffix(part, ":*"):
			st = icStep{kind: stepAny}
		default:
			st = icStep{kind: stepName, local: localPart(part)}
		}
		p.steps = append(p.steps, st)
	}
	return p
}

func localPart(qname string) string {
	if i := strings.LastIndexByte(qname, ':'); i >= 0 {
		return qname[i+1:]
	}
	return qname
}

// selectTargets evaluates a selector relative to scope, yielding the target
// element set (deduplicated). Elements inside processContents="skip" regions
// are not assessed and so are excluded from identity-constraint scope.
func (a *assessor) selectTargets(scope Element, expr string) []Element {
	var out []Element
	seen := map[Element]bool{}
	for _, p := range parsePaths(expr) {
		base := []Element{scope}
		if p.descendant {
			base = a.selfAndDescendants(scope)
		}
		for _, el := range a.applyElementSteps(base, p.steps) {
			if !seen[el] {
				seen[el] = true
				out = append(out, el)
			}
		}
	}
	return out
}

// selectFieldNodes evaluates a field relative to a target, returning the
// selected element or attribute nodes.
func (a *assessor) selectFieldNodes(target Element, expr string) []fieldNode {
	var out []fieldNode
	seen := map[fieldNode]bool{}
	add := func(n fieldNode) {
		if !seen[n] {
			seen[n] = true
			out = append(out, n)
		}
	}
	for _, p := range parsePaths(expr) {
		base := []Element{target}
		if p.descendant {
			base = a.selfAndDescendants(target)
		}
		if n := len(p.steps); n > 0 && (p.steps[n-1].kind == stepAttrName || p.steps[n-1].kind == stepAttrAny) {
			elems := a.applyElementSteps(base, p.steps[:n-1])
			last := p.steps[n-1]
			for _, el := range elems {
				for _, at := range el.Attributes() {
					if at.Name().Namespace == xsd.XMLNSNS {
						continue
					}
					if last.kind == stepAttrAny || at.Name().Local == last.local {
						add(fieldNode{at: at, owner: el})
					}
				}
			}
		} else {
			for _, el := range a.applyElementSteps(base, p.steps) {
				add(fieldNode{el: el})
			}
		}
	}
	return out
}

func (a *assessor) applyElementSteps(set []Element, steps []icStep) []Element {
	cur := set
	for _, st := range steps {
		switch st.kind {
		case stepSelf:
			// unchanged
		case stepName:
			var next []Element
			for _, el := range cur {
				for _, c := range elementChildren(el) {
					if !a.skipped[c] && c.Name().Local == st.local {
						next = append(next, c)
					}
				}
			}
			cur = next
		case stepAny:
			var next []Element
			for _, el := range cur {
				for _, c := range elementChildren(el) {
					if !a.skipped[c] {
						next = append(next, c)
					}
				}
			}
			cur = next
		default:
			cur = nil
		}
	}
	return cur
}

func (a *assessor) selfAndDescendants(el Element) []Element {
	var out []Element
	var walk func(e Element)
	walk = func(e Element) {
		if a.skipped[e] {
			return
		}
		out = append(out, e)
		for _, c := range elementChildren(e) {
			walk(c)
		}
	}
	walk(el)
	return out
}

func collapse(s string) string { return strings.Join(strings.Fields(s), " ") }
