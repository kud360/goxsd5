package xpath

// A small XPath-2.0-subset evaluator over an abstract instance tree, the
// "evaluator half" of this package. It walks the AST built by the shared parser
// (parse.go) — the same parser whose syntax checks and TypeRefs drive the
// schema-time validation in Parse — to evaluate xs:assert and xs:alternative/@test
// at validation time. It is intentionally partial: a construct outside the
// evaluable subset (flagged by the parser, see parseTree) or any unsupported
// operation returns an error so callers can fail open — treat an assertion as
// satisfied and a type-table alternative as unmatched — and never reject an
// instance for syntax the evaluator does not understand.
//
// The evaluator is format-agnostic: it walks the Node / NodeAttr interfaces, so
// any infoset (XML today, JSON/BER later) can be assessed without this package
// depending on it. Name tests match by local name; the static namespace context
// of the expression is not threaded in (callers that need built-in casts supply
// EvalContext.Castable).
//
// Tree navigation (parent, sibling and the other reverse axes) is NOT a property
// of the Node interface — the infoset is walked downward only. Instead the
// evaluator threads a positioned view (*nodeCtx: a node plus the parent it was
// reached from and its index among siblings) as it descends. In a tree each
// element has exactly one parent and downward navigation always reaches it via
// that parent, so the synthesized parent matches the real one. This also
// enforces the XSD assertion "stay in subtree" rule for free: the context
// element is the parentless root, so following-sibling/parent/ancestor of it are
// empty and an absolute path (rooted at the absent document node) selects
// nothing.

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kud360/goxsd5/xsdregex"
	"github.com/kud360/goxsd5/xsdtemporal"
)

// Name is an expanded (namespace-qualified) name in the instance tree. Its
// field layout matches xsd.QName so callers can convert directly.
type Name struct {
	Namespace string
	Local     string
}

// Node is one element item of the instance tree the evaluator walks.
type Node interface {
	NodeName() Name
	NodeAttrs() []NodeAttr
	NodeChildren() []Node // element children, in order
	StringValue() string  // the element's string value (deep character data)
}

// NodeAttr is one attribute item.
type NodeAttr interface {
	AttrName() Name
	AttrValue() string
}

// AtomKind tags how a bound atomic value participates in comparison: it lets a
// schema-typed $value compare with the right semantics rather than the
// "numeric if both look numeric, else string" default that untyped node and
// literal values use.
type AtomKind int

const (
	// KindUntyped is the default for node/attribute values and string literals:
	// a comparison coerces to number when both sides parse as numbers, else
	// compares as strings (XPath untypedAtomic-ish behaviour).
	KindUntyped AtomKind = iota
	// KindNumber forces numeric comparison (xs:decimal/double/float and derived).
	KindNumber
	// KindString forces string comparison and never coerces to a number, even
	// when the lexical form looks numeric (xs:string and other non-numeric
	// types, including dates, whose ISO lexical compares chronologically).
	KindString
)

// TypedAtom is one schema-typed atomic value bound to an XPath variable.
type TypedAtom struct {
	Lexical string
	Kind    AtomKind
}

// EvalContext supplies the callbacks the evaluator cannot resolve on its own.
type EvalContext struct {
	// Castable reports whether value is castable to the built-in datatype with
	// the given local name (for `castable as`, `instance of`, and constructor
	// functions like xs:integer(...)). A nil Castable makes those constructs
	// evaluation errors (fail open).
	Castable func(typeLocal, value string) bool

	// Vars binds XPath variables ("$name") to a sequence of atomic string
	// values. A reference to an unbound variable is an evaluation error (fail
	// open). Used by simple-type assertions to bind $value to the value being
	// validated. Runtime bindings introduced by for/some/every shadow these.
	Vars map[string][]string

	// TypedVars binds variables to schema-typed atomic values (checked before
	// Vars). Used to bind $value with its type's comparison semantics.
	TypedVars map[string][]TypedAtom

	// NoContextItem marks an evaluation whose dynamic context has no context
	// item — the case for simple-type xs:assertion facets, where only $value is
	// defined. Referencing the context item (".", a relative path, position(),
	// last()) is then a dynamic error, which makes the assertion unsatisfied
	// rather than failing open.
	NoContextItem bool
}

// item is one XPath item: float64, string, bool, *nodeCtx (a positioned element
// node), or *attrItem (a positioned attribute).
type item any

type seq []item

// nodeCtx is a node together with the position from which it was reached: the
// parent it descends from (nil for the parentless root) and its index among the
// parent's element children. This gives the reverse and sibling axes without
// the Node interface exposing any upward links.
type nodeCtx struct {
	node   Node
	parent *nodeCtx
	index  int // position among parent.node.NodeChildren(); 0 for the root
}

// attrItem is an attribute together with its owner element node.
type attrItem struct {
	attr  NodeAttr
	owner *nodeCtx
}

// typedItem is a schema-typed atomic value (from EvalContext.TypedVars). Its
// kind drives comparison: KindString never coerces to a number even when the
// lexical looks numeric, so an xs:string $value compares as a string.
type typedItem struct {
	lex  string
	kind AtomKind
}

// errUnsupported marks a construct outside the supported subset.
var errUnsupported = fmt.Errorf("xpath: unsupported construct")

// errDynamic marks a genuine XPath dynamic (run-time) error — a failed
// constructor cast, or a reference to an absent context item. Unlike
// errUnsupported it is not "we can't evaluate this": the expression is
// understood and the spec says it raises an error, which for an assertion means
// the assertion is not satisfied. EvalBool reports it as a definite false.
var errDynamic = fmt.Errorf("xpath: dynamic error")

// EvalBool evaluates expr against the context node and returns its effective
// boolean value. ok is false when the expression is outside the supported
// subset, so callers can fail open. A dynamic error evaluates to a definite
// false (ok true): the spec treats an assertion that raises one as unsatisfied.
func EvalBool(expr string, node Node, ec EvalContext) (result, ok bool) {
	v, err := evalExpr(expr, node, ec)
	if err == errDynamic {
		return false, true
	}
	if err != nil {
		return false, false
	}
	return effectiveBool(v), true
}

func evalExpr(expr string, node Node, ec EvalContext) (seq, error) {
	root, _, unsupported, perr := parseTree(expr)
	// A static syntax error or an out-of-subset construct both mean "we cannot
	// evaluate this": fail open (the schema's XPaths were already syntax-checked
	// by Parse at build time, so this is the subset boundary in practice).
	if perr != nil || unsupported {
		return nil, errUnsupported
	}
	ev := &evaluator{ec: ec, vars: map[string]seq{}, ctxAbsent: ec.NoContextItem}
	rootCtx := &nodeCtx{node: node}
	return ev.eval(root, &focus{item: rootCtx, pos: 1, size: 1})
}

func effectiveBool(s seq) bool {
	if len(s) == 0 {
		return false
	}
	if len(s) == 1 {
		switch v := s[0].(type) {
		case bool:
			return v
		case float64:
			return v != 0 && !isNaN(v)
		case string:
			return v != ""
		}
	}
	return true // non-empty node sequence
}

func isNaN(f float64) bool { return f != f }

// stringValue is the value of a string literal token: the strict lexer keeps the
// surrounding quotes and the doubled-quote escapes (REC-xpath20 [74]/[75]), so
// strip the delimiters and collapse each doubled quote to a single character.
func stringValue(lit string) string {
	if len(lit) < 2 {
		return ""
	}
	quote := lit[0]
	body := lit[1 : len(lit)-1]
	if strings.IndexByte(body, quote) < 0 {
		return body
	}
	return strings.ReplaceAll(body, string([]byte{quote, quote}), string(quote))
}

// ---- AST ----

type exprNode any

type binary struct {
	op   string
	l, r exprNode
}
type unary struct {
	op string
	x  exprNode
}
type litNum struct{ v float64 }
type litStr struct{ v string }
type ifExpr struct{ cond, then, els exprNode }
type call struct {
	name string
	args []exprNode
}
type typeOp struct {
	x    exprNode
	kind string // "castable" or "instance"
	typ  string
}
type seqExpr struct{ items []exprNode }  // (e1, e2, ...) sequence construction
type rangeExpr struct{ lo, hi exprNode } // e1 to e2
type varRef struct{ name string }        // $name variable reference

type filterExpr struct { // a primary followed by predicates: primary[p1][p2]
	x     exprNode
	preds []exprNode
}

// pathExpr is a (possibly absolute) path of location steps. start, when set, is
// a primary expression the path navigates from (e.g. $w/following-sibling::*);
// otherwise the path starts from the context node.
type pathExpr struct {
	fromRoot bool
	start    exprNode
	steps    []pathStep
}

type axisKind int

const (
	axChild axisKind = iota
	axDescendant
	axDescendantOrSelf
	axAttribute
	axSelf
	axParent
	axAncestor
	axAncestorOrSelf
	axFollowingSibling
	axPrecedingSibling
	axFollowing
	axPreceding
)

var axisNames = map[string]axisKind{
	"child":              axChild,
	"descendant":         axDescendant,
	"descendant-or-self": axDescendantOrSelf,
	"attribute":          axAttribute,
	"self":               axSelf,
	"parent":             axParent,
	"ancestor":           axAncestor,
	"ancestor-or-self":   axAncestorOrSelf,
	"following-sibling":  axFollowingSibling,
	"preceding-sibling":  axPrecedingSibling,
	"following":          axFollowing,
	"preceding":          axPreceding,
}

type testKind int

const (
	tnName testKind = iota // a specific local name
	tnStar                 // "*": any element (or any attribute on the attribute axis)
	tnNode                 // node(): any node
	tnText                 // text(): a text node (not modelled; matches nothing)
)

type nodeTest struct {
	kind testKind
	name string
}

type pathStep struct {
	axis  axisKind
	test  nodeTest
	preds []exprNode
}

type binding struct {
	name string
	seq  exprNode
}
type quantified struct {
	every bool // "every" vs "some"
	binds []binding
	body  exprNode
}
type forExpr struct {
	binds []binding
	body  exprNode
}

// ---- evaluator ----

// focus is the XPath dynamic focus: the context item plus its position and size
// within the sequence currently being processed (for position()/last()).
type focus struct {
	item item
	pos  int
	size int
}

func (f *focus) node() (*nodeCtx, bool) {
	if f == nil {
		return nil, false
	}
	switch v := f.item.(type) {
	case *nodeCtx:
		return v, true
	case *attrItem:
		return v.owner, true
	}
	return nil, false
}

type evaluator struct {
	ec   EvalContext
	vars map[string]seq
	// ctxAbsent mirrors EvalContext.NoContextItem: when set, referencing the
	// context item is a dynamic error (simple-type assertion semantics).
	ctxAbsent bool
}

func (e *evaluator) eval(node exprNode, f *focus) (seq, error) {
	switch n := node.(type) {
	case *litNum:
		return seq{n.v}, nil
	case *litStr:
		return seq{n.v}, nil
	case *unary:
		v, err := e.eval(n.x, f)
		if err != nil {
			return nil, err
		}
		fl, err := toNumber(v)
		if err != nil {
			return nil, err
		}
		if n.op == "-" {
			fl = -fl
		}
		return seq{fl}, nil
	case *binary:
		return e.evalBinary(n, f)
	case *ifExpr:
		c, err := e.eval(n.cond, f)
		if err != nil {
			return nil, err
		}
		if effectiveBool(c) {
			return e.eval(n.then, f)
		}
		return e.eval(n.els, f)
	case *call:
		return e.evalCall(n, f)
	case *pathExpr:
		return e.evalPath(n, f)
	case *filterExpr:
		return e.evalFilter(n, f)
	case *typeOp:
		return e.evalTypeOp(n, f)
	case *varRef:
		return e.evalVar(n)
	case *seqExpr:
		var out seq
		for _, it := range n.items {
			v, err := e.eval(it, f)
			if err != nil {
				return nil, err
			}
			out = append(out, v...)
		}
		return out, nil
	case *rangeExpr:
		return e.evalRange(n, f)
	case *quantified:
		return e.evalQuantified(n, f)
	case *forExpr:
		return e.evalFor(n, f)
	}
	return nil, errUnsupported
}

func (e *evaluator) evalVar(n *varRef) (seq, error) {
	if v, ok := e.vars[n.name]; ok {
		return v, nil
	}
	if vals, ok := e.ec.TypedVars[n.name]; ok {
		out := make(seq, len(vals))
		for i, v := range vals {
			out[i] = typedItem{lex: v.Lexical, kind: v.Kind}
		}
		return out, nil
	}
	vals, ok := e.ec.Vars[n.name]
	if !ok {
		return nil, errUnsupported
	}
	out := make(seq, len(vals))
	for i, v := range vals {
		out[i] = v
	}
	return out, nil
}

func (e *evaluator) evalRange(n *rangeExpr, f *focus) (seq, error) {
	lo, err := e.eval(n.lo, f)
	if err != nil {
		return nil, err
	}
	hi, err := e.eval(n.hi, f)
	if err != nil {
		return nil, err
	}
	lf, err := toNumber(lo)
	if err != nil {
		return nil, err
	}
	hf, err := toNumber(hi)
	if err != nil {
		return nil, err
	}
	var out seq
	for i := int64(lf); i <= int64(hf); i++ {
		out = append(out, float64(i))
	}
	return out, nil
}

// evalQuantified evaluates "some/every $v in E ... satisfies B" by iterating the
// cartesian product of the binding sequences, binding the variables in the
// runtime scope. "every" is true unless some tuple makes B false; "some" is true
// as soon as one tuple makes B true.
func (e *evaluator) evalQuantified(n *quantified, f *focus) (seq, error) {
	res, err := e.iterateBindings(n.binds, 0, f, func() (bool, bool, error) {
		v, err := e.eval(n.body, f)
		if err != nil {
			return false, false, err
		}
		b := effectiveBool(v)
		// stop early: "every" stops on the first false, "some" on the first true.
		return b, b != n.every, nil
	}, n.every)
	if err != nil {
		return nil, err
	}
	return seq{res}, nil
}

// iterateBindings recurses over the binding clauses, calling visit for each
// tuple. visit returns (result, stop, err): result is the tuple's boolean, stop
// requests early termination. seed is the starting accumulator (true for every,
// false for some); the accumulator folds with "every"⇒AND, "some"⇒OR.
func (e *evaluator) iterateBindings(binds []binding, i int, f *focus, visit func() (bool, bool, error), every bool) (bool, error) {
	if i == len(binds) {
		r, _, err := visit()
		return r, err
	}
	s, err := e.eval(binds[i].seq, f)
	if err != nil {
		return false, err
	}
	acc := every
	saved, had := e.vars[binds[i].name], hasKey(e.vars, binds[i].name)
	defer func() {
		if had {
			e.vars[binds[i].name] = saved
		} else {
			delete(e.vars, binds[i].name)
		}
	}()
	for _, it := range s {
		e.vars[binds[i].name] = seq{it}
		r, err := e.iterateBindings(binds, i+1, f, visit, every)
		if err != nil {
			return false, err
		}
		if every {
			acc = acc && r
			if !acc {
				return false, nil
			}
		} else {
			acc = acc || r
			if acc {
				return true, nil
			}
		}
	}
	return acc, nil
}

// evalFor evaluates "for $v in E return B", concatenating B over each binding.
func (e *evaluator) evalFor(n *forExpr, f *focus) (seq, error) {
	return e.forProduct(n.binds, 0, f, n.body)
}

func (e *evaluator) forProduct(binds []binding, i int, f *focus, body exprNode) (seq, error) {
	if i == len(binds) {
		return e.eval(body, f)
	}
	s, err := e.eval(binds[i].seq, f)
	if err != nil {
		return nil, err
	}
	saved, had := e.vars[binds[i].name], hasKey(e.vars, binds[i].name)
	defer func() {
		if had {
			e.vars[binds[i].name] = saved
		} else {
			delete(e.vars, binds[i].name)
		}
	}()
	var out seq
	for _, it := range s {
		e.vars[binds[i].name] = seq{it}
		r, err := e.forProduct(binds, i+1, f, body)
		if err != nil {
			return nil, err
		}
		out = append(out, r...)
	}
	return out, nil
}

func hasKey(m map[string]seq, k string) bool { _, ok := m[k]; return ok }

func (e *evaluator) evalBinary(n *binary, f *focus) (seq, error) {
	switch n.op {
	case "and":
		return e.evalAnd(n, f)
	case "or":
		return e.evalOr(n, f)
	case "|":
		l, err := e.eval(n.l, f)
		if err != nil {
			return nil, err
		}
		r, err := e.eval(n.r, f)
		if err != nil {
			return nil, err
		}
		return unionNodes(l, r)
	case "+", "-", "*", "div", "mod", "idiv":
		return e.evalArith(n, f)
	default:
		l, err := e.eval(n.l, f)
		if err != nil {
			return nil, err
		}
		r, err := e.eval(n.r, f)
		if err != nil {
			return nil, err
		}
		res, err := compareSeq(n.op, l, r)
		if err != nil {
			return nil, err
		}
		return seq{res}, nil
	}
}

// evalAnd is the short-circuiting "and": a false left operand yields false
// without evaluating the right.
func (e *evaluator) evalAnd(n *binary, f *focus) (seq, error) {
	l, err := e.eval(n.l, f)
	if err != nil {
		return nil, err
	}
	if !effectiveBool(l) {
		return seq{false}, nil
	}
	r, err := e.eval(n.r, f)
	if err != nil {
		return nil, err
	}
	return seq{effectiveBool(r)}, nil
}

// evalOr is the short-circuiting "or": a true left operand yields true without
// evaluating the right.
func (e *evaluator) evalOr(n *binary, f *focus) (seq, error) {
	l, err := e.eval(n.l, f)
	if err != nil {
		return nil, err
	}
	if effectiveBool(l) {
		return seq{true}, nil
	}
	r, err := e.eval(n.r, f)
	if err != nil {
		return nil, err
	}
	return seq{effectiveBool(r)}, nil
}

// evalArith evaluates the numeric binary operators (+, -, *, div, mod, idiv)
// after coercing both operands to numbers.
func (e *evaluator) evalArith(n *binary, f *focus) (seq, error) {
	l, err := e.eval(n.l, f)
	if err != nil {
		return nil, err
	}
	r, err := e.eval(n.r, f)
	if err != nil {
		return nil, err
	}
	lf, err := toNumber(l)
	if err != nil {
		return nil, err
	}
	rf, err := toNumber(r)
	if err != nil {
		return nil, err
	}
	return seq{arith(n.op, lf, rf)}, nil
}

// unionNodes is the "|" operator: the union of two node sequences, de-duplicated
// by node identity. Both operands must contain only nodes; any atomic value
// makes it a type error, so the caller fails open.
func unionNodes(l, r seq) (seq, error) {
	seen := map[any]bool{}
	var out seq
	add := func(s seq) error {
		for _, it := range s {
			id := nodeIdentity(it)
			if id == nil {
				return errUnsupported
			}
			if !seen[id] {
				seen[id] = true
				out = append(out, it)
			}
		}
		return nil
	}
	if err := add(l); err != nil {
		return nil, err
	}
	if err := add(r); err != nil {
		return nil, err
	}
	return out, nil
}

// nodeIdentity returns a comparable identity for a node item, or nil if it is
// not a node.
func nodeIdentity(it item) any {
	switch v := it.(type) {
	case *nodeCtx:
		return v.node
	case *attrItem:
		return v.attr
	}
	return nil
}

func arith(op string, a, b float64) float64 {
	switch op {
	case "+":
		return a + b
	case "-":
		return a - b
	case "*":
		return a * b
	case "div":
		if b == 0 {
			return 0
		}
		return a / b
	case "mod":
		if b == 0 {
			return 0
		}
		return float64(int64(a) % int64(b))
	case "idiv":
		if b == 0 {
			return 0
		}
		return float64(int64(a / b))
	}
	return 0
}

func compareSeq(op string, l, r seq) (bool, error) {
	value := false
	switch op {
	case "eq":
		op, value = "=", true
	case "ne":
		op, value = "!=", true
	case "lt":
		op, value = "<", true
	case "le":
		op, value = "<=", true
	case "gt":
		op, value = ">", true
	case "ge":
		op, value = ">=", true
	}
	la := atomizeAll(l)
	ra := atomizeAll(r)
	if value {
		if len(la) != 1 || len(ra) != 1 {
			if len(la) == 0 || len(ra) == 0 {
				return false, nil
			}
			return false, errUnsupported
		}
		return cmpAtoms(op, la[0], ra[0]), nil
	}
	for _, a := range la {
		for _, b := range ra {
			if cmpAtoms(op, a, b) {
				return true, nil
			}
		}
	}
	return false, nil
}

func cmpAtoms(op string, a, b item) bool {
	if af, aok := asNumber(a); aok {
		if bf, bok := asNumber(b); bok {
			return cmpNum(op, af, bf)
		}
	}
	as, bs := atomString(a), atomString(b)
	switch op {
	case "=":
		return as == bs
	case "!=":
		return as != bs
	case "<":
		return as < bs
	case "<=":
		return as <= bs
	case ">":
		return as > bs
	case ">=":
		return as >= bs
	}
	return false
}

func cmpNum(op string, a, b float64) bool {
	switch op {
	case "=":
		return a == b
	case "!=":
		return a != b
	case "<":
		return a < b
	case "<=":
		return a <= b
	case ">":
		return a > b
	case ">=":
		return a >= b
	}
	return false
}

// ---- path evaluation ----

func (e *evaluator) evalPath(n *pathExpr, f *focus) (seq, error) {
	if n.fromRoot {
		// No document node exists above the assertion's root element, so an
		// absolute path selects nothing (XSD 1.1 §3.13.4.1 "stay in subtree").
		return seq{}, nil
	}
	var ctxItems seq
	if n.start != nil {
		s, err := e.eval(n.start, f)
		if err != nil {
			return nil, err
		}
		ctxItems = s
	} else {
		// A path relative to the context item ("." or a child step) is a dynamic
		// error when no context item is defined (simple-type assertion).
		if e.ctxAbsent {
			return nil, errDynamic
		}
		if f == nil || f.item == nil {
			return nil, errUnsupported
		}
		ctxItems = seq{f.item}
	}
	for _, st := range n.steps {
		next, err := e.evalStep(st, ctxItems)
		if err != nil {
			return nil, err
		}
		ctxItems = next
	}
	return ctxItems, nil
}

func (e *evaluator) evalFilter(n *filterExpr, f *focus) (seq, error) {
	v, err := e.eval(n.x, f)
	if err != nil {
		return nil, err
	}
	return e.applyPreds(v, n.preds)
}

// evalStep applies a location step to every context item: the axis and node
// test, then the predicates per context node (so positional predicates count
// within each node's own axis result), and finally a union that de-duplicates
// the combined result by node identity (XPath path-step semantics).
func (e *evaluator) evalStep(st pathStep, ctxItems seq) (seq, error) {
	var out seq
	seen := map[any]bool{}
	for _, it := range ctxItems {
		nodes, err := axisCandidates(st.axis, it)
		if err != nil {
			return nil, err
		}
		var matched seq
		for _, c := range nodes {
			if matchesTest(st.test, c) {
				matched = append(matched, c)
			}
		}
		kept, err := e.applyPreds(matched, st.preds)
		if err != nil {
			return nil, err
		}
		for _, c := range kept {
			id := nodeIdentity(c)
			if id == nil {
				out = append(out, c)
				continue
			}
			if !seen[id] {
				seen[id] = true
				out = append(out, c)
			}
		}
	}
	return out, nil
}

// axisCandidates returns the items on the given axis from a context item (before
// the node test). Attributes only support self and parent.
func axisCandidates(ax axisKind, it item) (seq, error) {
	switch v := it.(type) {
	case *attrItem:
		switch ax {
		case axSelf:
			return seq{v}, nil
		case axParent, axAncestor, axAncestorOrSelf:
			if v.owner == nil {
				return nil, nil
			}
			anc := ancestorsOf(v.owner, ax == axAncestorOrSelf)
			if ax == axParent {
				return seq{wrap(v.owner)}, nil
			}
			return anc, nil
		}
		return nil, errUnsupported
	case *nodeCtx:
		return axisFromNode(ax, v), nil
	}
	return nil, errUnsupported
}

func axisFromNode(ax axisKind, c *nodeCtx) seq {
	switch ax {
	case axSelf:
		return seq{c}
	case axChild:
		return wrapAll(childrenOf(c))
	case axAttribute:
		var out seq
		for _, a := range c.node.NodeAttrs() {
			out = append(out, &attrItem{attr: a, owner: c})
		}
		return out
	case axDescendant:
		return wrapAll(descendantsOf(c))
	case axDescendantOrSelf:
		return append(seq{c}, wrapAll(descendantsOf(c))...)
	case axParent:
		if c.parent == nil {
			return nil
		}
		return seq{c.parent}
	case axAncestor:
		return ancestorsOf(c, false)
	case axAncestorOrSelf:
		return ancestorsOf(c, true)
	case axFollowingSibling:
		return siblings(c, true)
	case axPrecedingSibling:
		return siblings(c, false)
	case axFollowing:
		return followingOf(c)
	case axPreceding:
		return precedingOf(c)
	}
	return nil
}

func matchesTest(t nodeTest, it item) bool {
	switch v := it.(type) {
	case *nodeCtx:
		switch t.kind {
		case tnStar, tnNode:
			return true
		case tnName:
			return v.node.NodeName().Local == t.name
		case tnText:
			return false // text nodes are not modelled as Nodes
		}
	case *attrItem:
		switch t.kind {
		case tnStar, tnNode:
			return true
		case tnName:
			return v.attr.AttrName().Local == t.name
		case tnText:
			return false
		}
	}
	return false
}

// applyPreds filters set by each predicate in turn, threading position/size so
// numeric predicates select by position and others by effective boolean.
func (e *evaluator) applyPreds(set seq, preds []exprNode) (seq, error) {
	for _, pred := range preds {
		var kept seq
		size := len(set)
		for idx, it := range set {
			f := &focus{item: it, pos: idx + 1, size: size}
			v, err := e.eval(pred, f)
			if err != nil {
				return nil, err
			}
			if len(v) == 1 {
				if fl, ok := v[0].(float64); ok {
					if int(fl) == idx+1 {
						kept = append(kept, it)
					}
					continue
				}
			}
			if effectiveBool(v) {
				kept = append(kept, it)
			}
		}
		set = kept
	}
	return set, nil
}

// ---- tree navigation over positioned nodes ----

func wrap(c *nodeCtx) item { return c }

func wrapAll(cs []*nodeCtx) seq {
	out := make(seq, len(cs))
	for i, c := range cs {
		out[i] = c
	}
	return out
}

// childrenOf returns c's element children, each positioned with c as parent and
// its index among them.
func childrenOf(c *nodeCtx) []*nodeCtx {
	kids := c.node.NodeChildren()
	out := make([]*nodeCtx, len(kids))
	for i, k := range kids {
		out[i] = &nodeCtx{node: k, parent: c, index: i}
	}
	return out
}

// descendantsOf returns c's descendants in document order (preorder).
func descendantsOf(c *nodeCtx) []*nodeCtx {
	var out []*nodeCtx
	for _, k := range childrenOf(c) {
		out = append(out, k)
		out = append(out, descendantsOf(k)...)
	}
	return out
}

// ancestorsOf returns c's ancestors nearest-first (reverse document order); when
// orSelf is set, c precedes them.
func ancestorsOf(c *nodeCtx, orSelf bool) seq {
	var out seq
	if orSelf {
		out = append(out, c)
	}
	for a := c.parent; a != nil; a = a.parent {
		out = append(out, a)
	}
	return out
}

// siblings returns c's following (forward=true) or preceding (forward=false)
// siblings; preceding siblings are returned in reverse document order, as the
// reverse axis requires.
func siblings(c *nodeCtx, forward bool) seq {
	if c.parent == nil {
		return nil
	}
	all := childrenOf(c.parent)
	var out seq
	if forward {
		for i := c.index + 1; i < len(all); i++ {
			out = append(out, all[i])
		}
	} else {
		for i := c.index - 1; i >= 0; i-- {
			out = append(out, all[i])
		}
	}
	return out
}

// followingOf returns the following axis: every node after c in document order,
// excluding c's descendants and ancestors.
func followingOf(c *nodeCtx) seq {
	var out seq
	for a := c; a != nil; a = a.parent {
		for _, sib := range siblings(a, true) {
			s := sib.(*nodeCtx)
			out = append(out, s)
			out = append(out, wrapAll(descendantsOf(s))...)
		}
	}
	return out
}

// precedingOf returns the preceding axis: every node before c in document order,
// excluding c's ancestors (reverse document order).
func precedingOf(c *nodeCtx) seq {
	var out seq
	for a := c; a != nil; a = a.parent {
		for _, sib := range siblings(a, false) {
			s := sib.(*nodeCtx)
			// the subtree in reverse document order: descendants (reverse) then self
			ds := descendantsOf(s)
			for i := len(ds) - 1; i >= 0; i-- {
				out = append(out, ds[i])
			}
			out = append(out, s)
		}
	}
	return out
}

// ---- type operators / functions ----

func (e *evaluator) evalTypeOp(n *typeOp, f *focus) (seq, error) {
	v, err := e.eval(n.x, f)
	if err != nil {
		return nil, err
	}
	atoms := atomizeAll(v)
	if n.kind == "cast" {
		// "x cast as T" yields the value cast to T; a non-castable value (or a
		// non-singleton) raises a dynamic error.
		if len(atoms) != 1 || e.ec.Castable == nil {
			return nil, errDynamic
		}
		s := atomString(atoms[0])
		if !e.ec.Castable(n.typ, s) {
			return nil, errDynamic
		}
		return seq{castValue(n.typ, s)}, nil
	}
	// "castable as" / "instance of" — a boolean test.
	if len(atoms) == 0 {
		return seq{false}, nil
	}
	if len(atoms) != 1 || e.ec.Castable == nil {
		return nil, errUnsupported
	}
	return seq{e.ec.Castable(n.typ, atomString(atoms[0]))}, nil
}

// castValue produces the in-evaluator representation of a lexical value cast to
// a built-in type. xs:boolean must become a real boolean (so its effective
// boolean value is the value, not "non-empty string"); other types keep the
// lexical, tagged with the right comparison kind.
func castValue(typeLocal, lex string) item {
	switch typeLocal {
	case "boolean":
		return lex == "true" || lex == "1"
	case "decimal", "integer", "int", "long", "short", "byte",
		"nonNegativeInteger", "nonPositiveInteger", "positiveInteger",
		"negativeInteger", "unsignedLong", "unsignedInt", "unsignedShort",
		"unsignedByte", "float", "double":
		return typedItem{lex: lex, kind: KindNumber}
	default:
		return typedItem{lex: lex, kind: KindString}
	}
}

func (e *evaluator) evalCall(n *call, f *focus) (seq, error) {
	// The XPath function library is partitioned into category helpers; each
	// reports whether it handled n.name, so the first match wins.
	if s, ok, err := e.evalLogicFn(n, f); ok {
		return s, err
	}
	if s, ok, err := e.evalStringFn(n, f); ok {
		return s, err
	}
	if s, ok, err := e.evalNumericFn(n, f); ok {
		return s, err
	}
	if s, ok, err := e.evalAggregateFn(n, f); ok {
		return s, err
	}
	if s, ok, err := e.evalDateFn(n, f); ok {
		return s, err
	}
	if s, ok, err := e.evalNodeFn(n, f); ok {
		return s, err
	}
	// Constructor functions: xs:integer(…), xs:decimal(…), etc. — the argument
	// must be castable to the named built-in type; the cast value is carried as
	// its lexical form, which numeric/string comparisons coerce as needed.
	if strings.IndexByte(n.name, ':') >= 0 && len(n.args) == 1 && e.ec.Castable != nil {
		v, err := e.arg(n, 0, f)
		if err != nil {
			return nil, err
		}
		s := seqString(v)
		if !e.ec.Castable(localPart(n.name), s) {
			// A constructor whose argument is not castable to the target type
			// raises a dynamic error (e.g. xs:date("not-a-date")), which makes a
			// containing assertion unsatisfied rather than failing open.
			return nil, errDynamic
		}
		return seq{s}, nil
	}
	return nil, errUnsupported
}

// evalLogicFn handles boolean, existence, count, positional, and data() calls.
func (e *evaluator) evalLogicFn(n *call, f *focus) (seq, bool, error) {
	switch n.name {
	case "true":
		return seq{true}, true, nil
	case "false":
		return seq{false}, true, nil
	case "not":
		v, err := e.arg(n, 0, f)
		if err != nil {
			return nil, true, err
		}
		return seq{!effectiveBool(v)}, true, nil
	case "boolean":
		v, err := e.arg(n, 0, f)
		if err != nil {
			return nil, true, err
		}
		return seq{effectiveBool(v)}, true, nil
	case "exists":
		v, err := e.arg(n, 0, f)
		if err != nil {
			return nil, true, err
		}
		return seq{len(v) > 0}, true, nil
	case "empty":
		v, err := e.arg(n, 0, f)
		if err != nil {
			return nil, true, err
		}
		return seq{len(v) == 0}, true, nil
	case "count":
		v, err := e.arg(n, 0, f)
		if err != nil {
			return nil, true, err
		}
		return seq{float64(len(v))}, true, nil
	case "position":
		if e.ctxAbsent {
			return nil, true, errDynamic
		}
		if f == nil {
			return nil, true, errUnsupported
		}
		return seq{float64(f.pos)}, true, nil
	case "last":
		if e.ctxAbsent {
			return nil, true, errDynamic
		}
		if f == nil {
			return nil, true, errUnsupported
		}
		return seq{float64(f.size)}, true, nil
	case "data":
		v, err := e.arg(n, 0, f)
		if err != nil {
			return nil, true, err
		}
		// data() over nodes would need each node's schema-typed value, which this
		// evaluator does not model (a node atomizes only to its untyped string).
		// Restrict data() to already-atomic operands (e.g. data($value)); a node
		// operand is unsupported, so the caller fails open rather than treating
		// the untyped string as a typed value.
		for _, it := range v {
			if nodeIdentity(it) != nil {
				return nil, true, errUnsupported
			}
		}
		return seq(atomizeAll(v)), true, nil
	}
	return nil, false, nil
}

// evalStringFn handles the string-valued functions.
func (e *evaluator) evalStringFn(n *call, f *focus) (seq, bool, error) {
	switch n.name {
	case "string":
		if len(n.args) == 0 {
			return seq{focusString(f)}, true, nil
		}
		v, err := e.arg(n, 0, f)
		if err != nil {
			return nil, true, err
		}
		return seq{seqString(v)}, true, nil
	case "number":
		v, err := e.arg(n, 0, f)
		if err != nil {
			return nil, true, err
		}
		fl, err := toNumber(v)
		if err != nil {
			return nil, true, err
		}
		return seq{fl}, true, nil
	case "string-length":
		s, err := e.strArgOrCtx(n, f)
		if err != nil {
			return nil, true, err
		}
		return seq{float64(len([]rune(s)))}, true, nil
	case "normalize-space":
		s, err := e.strArgOrCtx(n, f)
		if err != nil {
			return nil, true, err
		}
		return seq{collapse(s)}, true, nil
	case "contains", "starts-with", "ends-with":
		a, err := e.arg(n, 0, f)
		if err != nil {
			return nil, true, err
		}
		b, err := e.arg(n, 1, f)
		if err != nil {
			return nil, true, err
		}
		s1, s2 := seqString(a), seqString(b)
		switch n.name {
		case "contains":
			return seq{strings.Contains(s1, s2)}, true, nil
		case "starts-with":
			return seq{strings.HasPrefix(s1, s2)}, true, nil
		default:
			return seq{strings.HasSuffix(s1, s2)}, true, nil
		}
	case "concat":
		var b strings.Builder
		for i := range n.args {
			v, err := e.arg(n, i, f)
			if err != nil {
				return nil, true, err
			}
			b.WriteString(seqString(v))
		}
		return seq{b.String()}, true, nil
	case "substring":
		return e.evalSubstring(n, f)
	case "substring-before", "substring-after":
		return e.evalSubstringPart(n, f)
	case "matches":
		return e.evalMatches(n, f)
	case "replace":
		return e.evalReplace(n, f)
	case "tokenize":
		return e.evalTokenize(n, f)
	}
	return nil, false, nil
}

// evalSubstringPart implements fn:substring-before and fn:substring-after: the
// portion of $input before / after the first occurrence of $sub. An absent or
// empty search string yields "" per F&O.
func (e *evaluator) evalSubstringPart(n *call, f *focus) (seq, bool, error) {
	a, err := e.arg(n, 0, f)
	if err != nil {
		return nil, true, err
	}
	b, err := e.arg(n, 1, f)
	if err != nil {
		return nil, true, err
	}
	s, sub := seqString(a), seqString(b)
	i := strings.Index(s, sub)
	if sub == "" || i < 0 {
		return seq{""}, true, nil
	}
	if n.name == "substring-before" {
		return seq{s[:i]}, true, nil
	}
	return seq{s[i+len(sub):]}, true, nil
}

// evalSubstring implements fn:substring (1-based, rounding semantics per F&O):
// it returns the characters at positions p with round(start) <= p <
// round(start)+round(length); the 2-arg form has no upper bound. start/length
// are rounded with fn:round (half to positive infinity), positions count
// codepoints, and an empty string operand yields "".
func (e *evaluator) evalSubstring(n *call, f *focus) (seq, bool, error) {
	src, err := e.arg(n, 0, f)
	if err != nil {
		return nil, true, err
	}
	start, err := e.numArg(n, 1, f)
	if err != nil {
		return nil, true, err
	}
	runes := []rune(seqString(src))
	lo := xpathRound(start)
	hi := math.Inf(1)
	if len(n.args) >= 3 {
		length, err := e.numArg(n, 2, f)
		if err != nil {
			return nil, true, err
		}
		hi = lo + xpathRound(length)
	}
	var b strings.Builder
	for i, r := range runes {
		p := float64(i + 1) // 1-based position
		if p >= lo && p < hi {
			b.WriteRune(r)
		}
	}
	return seq{b.String()}, true, nil
}

// evalMatches implements fn:matches($input, $pattern[, $flags]) with the
// XPath/XQuery F&O regex flavor XSD 1.1 binds assertions to (via
// xsdregex.TranslateFO): `^`/`$` anchor, `.` excludes newline unless the `s`
// flag is set, and the `i` flag is case-insensitive. A bad/uncompilable
// pattern or an unsupported flag (`m`/`x`/`q`, back-references) is a dynamic
// error, so a failed match in an assertion counts as unsatisfied rather than
// failing open.
func (e *evaluator) evalMatches(n *call, f *focus) (seq, bool, error) {
	a, err := e.arg(n, 0, f)
	if err != nil {
		return nil, true, err
	}
	p, err := e.arg(n, 1, f)
	if err != nil {
		return nil, true, err
	}
	flags, err := e.regexFlags(n, 2, f)
	if err != nil {
		return nil, true, err
	}
	re, err := foCompile(seqString(p), flags)
	if err != nil {
		return nil, true, errDynamic
	}
	return seq{re.MatchString(seqString(a))}, true, nil
}

// regexFlags reads the optional flag-string argument at position i (matches /
// replace / tokenize all take it as their last argument), returning "" when the
// argument is absent.
func (e *evaluator) regexFlags(n *call, i int, f *focus) (string, error) {
	if len(n.args) <= i {
		return "", nil
	}
	fl, err := e.arg(n, i, f)
	if err != nil {
		return "", err
	}
	return seqString(fl), nil
}

// foCompile translates an F&O-flavor pattern+flags and compiles it to a Go
// regexp. Translation rejects constructs RE2 cannot express (m/x/q flags,
// back-references) so callers surface those as dynamic errors.
func foCompile(pattern, flags string) (*regexp.Regexp, error) {
	s, err := xsdregex.TranslateFO(pattern, flags)
	if err != nil {
		return nil, err
	}
	return regexp.Compile(s)
}

// evalReplace implements fn:replace($input, $pattern, $replacement[, $flags])
// (F&O 7.6.4). Each non-overlapping match of the F&O-flavor pattern is replaced
// by $replacement, in which `$N` (single digit) is the substring captured by
// group N (`$0` is the whole match) and `\$`/`\\` are literal `$`/`\`. A
// pattern that can match the zero-length string, an uncompilable pattern, or a
// malformed replacement string is a dynamic error (assertion unsatisfied).
func (e *evaluator) evalReplace(n *call, f *focus) (seq, bool, error) {
	a, err := e.arg(n, 0, f)
	if err != nil {
		return nil, true, err
	}
	p, err := e.arg(n, 1, f)
	if err != nil {
		return nil, true, err
	}
	r, err := e.arg(n, 2, f)
	if err != nil {
		return nil, true, err
	}
	flags, err := e.regexFlags(n, 3, f)
	if err != nil {
		return nil, true, err
	}
	re, err := foCompile(seqString(p), flags)
	if err != nil {
		return nil, true, errDynamic
	}
	// F&O raises FORX0003 when the pattern matches the empty string.
	if re.MatchString("") {
		return nil, true, errDynamic
	}
	repl, err := goReplacement(seqString(r))
	if err != nil {
		return nil, true, errDynamic
	}
	return seq{re.ReplaceAllString(seqString(a), repl)}, true, nil
}

// goReplacement rewrites an F&O 7.6.4 replacement string into Go's
// Regexp.ReplaceAllString template. F&O uses `$N` for a single-digit group
// reference and `\` to escape `$` or `\`; every other `\` or a `$` not followed
// by a digit is an error. Go's template uses `${N}` for a group and `$$` for a
// literal `$`, so group refs become `${N}` (avoiding Go's greedy multi-digit
// parse) and literal `$` becomes `$$`.
func goReplacement(s string) (string, error) {
	var b strings.Builder
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		c := runes[i]
		switch c {
		case '\\':
			if i+1 >= len(runes) || (runes[i+1] != '\\' && runes[i+1] != '$') {
				return "", fmt.Errorf("replace: invalid backslash in replacement")
			}
			i++
			if runes[i] == '$' {
				b.WriteString("$$")
				continue
			}
			b.WriteByte('\\')
		case '$':
			if i+1 >= len(runes) || runes[i+1] < '0' || runes[i+1] > '9' {
				return "", fmt.Errorf("replace: $ must be followed by a digit")
			}
			i++
			b.WriteString("${")
			b.WriteRune(runes[i])
			b.WriteByte('}')
		default:
			// A bare '$' would start a Go reference; c is never '$' here.
			b.WriteRune(c)
		}
	}
	return b.String(), nil
}

// evalTokenize implements fn:tokenize($input[, $pattern[, $flags]]) (F&O
// 7.6.6). It splits $input at each non-overlapping match of the separator
// pattern and returns the sequence of pieces. The 1-arg form splits on
// whitespace after collapsing (leading/trailing whitespace trimmed, runs of
// whitespace treated as one separator). A pattern that matches the empty string
// or fails to compile is a dynamic error.
func (e *evaluator) evalTokenize(n *call, f *focus) (seq, bool, error) {
	a, err := e.arg(n, 0, f)
	if err != nil {
		return nil, true, err
	}
	input := seqString(a)
	if len(n.args) < 2 {
		return tokenizeWhitespace(input), true, nil
	}
	p, err := e.arg(n, 1, f)
	if err != nil {
		return nil, true, err
	}
	flags, err := e.regexFlags(n, 2, f)
	if err != nil {
		return nil, true, err
	}
	re, err := foCompile(seqString(p), flags)
	if err != nil {
		return nil, true, errDynamic
	}
	// F&O raises FORX0003 when the separator can match the empty string.
	if re.MatchString("") {
		return nil, true, errDynamic
	}
	// F&O: tokenizing the empty string yields the empty sequence.
	if input == "" {
		return seq{}, true, nil
	}
	pieces := re.Split(input, -1)
	out := make(seq, len(pieces))
	for i, piece := range pieces {
		out[i] = piece
	}
	return out, true, nil
}

// tokenizeWhitespace implements the 1-argument fn:tokenize: collapse whitespace
// and split on it. The empty (or all-whitespace) input yields the empty
// sequence per F&O — collapse trims the ends, so an empty result means no
// tokens.
func tokenizeWhitespace(input string) seq {
	collapsed := collapse(input)
	if collapsed == "" {
		return seq{}
	}
	parts := strings.Split(collapsed, " ")
	out := make(seq, len(parts))
	for i, piece := range parts {
		out[i] = piece
	}
	return out
}

// evalNumericFn handles the single-operand numeric functions. Each rounds or
// transforms one xs:decimal/float/double argument; an empty sequence yields the
// empty sequence (F&O), and a non-numeric operand is a type error reported as
// errDynamic — so a bad cast in an assertion makes it unsatisfied, never fails
// open. The evaluator models all numbers as float64, so the result is numeric
// (an integral value prints without a fractional part), matching count/sum/number.
func (e *evaluator) evalNumericFn(n *call, f *focus) (seq, bool, error) {
	switch n.name {
	case "abs", "floor", "ceiling", "round":
		v, err := e.arg(n, 0, f)
		if err != nil {
			return nil, true, err
		}
		atoms := atomizeAll(v)
		if len(atoms) == 0 {
			return seq{}, true, nil
		}
		if len(atoms) != 1 {
			return nil, true, errDynamic
		}
		x, ok := asNumber(atoms[0])
		if !ok {
			return nil, true, errDynamic
		}
		switch n.name {
		case "abs":
			return seq{math.Abs(x)}, true, nil
		case "floor":
			return seq{math.Floor(x)}, true, nil
		case "ceiling":
			return seq{math.Ceil(x)}, true, nil
		default:
			return seq{xpathRound(x)}, true, nil
		}
	}
	return nil, false, nil
}

// numArg evaluates argument i as a single number. A missing or non-numeric
// argument is a dynamic (type) error.
func (e *evaluator) numArg(n *call, i int, f *focus) (float64, error) {
	v, err := e.arg(n, i, f)
	if err != nil {
		return 0, err
	}
	atoms := atomizeAll(v)
	if len(atoms) != 1 {
		return 0, errDynamic
	}
	x, ok := asNumber(atoms[0])
	if !ok {
		return 0, errDynamic
	}
	return x, nil
}

// xpathRound implements fn:round's rule of rounding a value of type
// xs:decimal/float/double to the nearest integer, with ties rounded toward
// positive infinity: round(2.5)=3, round(-2.5)=-2. That differs from Go's
// math.Round (ties away from zero), so it is computed as floor(x+0.5). NaN and
// the infinities are returned unchanged.
func xpathRound(x float64) float64 {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return x
	}
	return math.Floor(x + 0.5)
}

// evalAggregateFn handles sequence aggregates.
func (e *evaluator) evalAggregateFn(n *call, f *focus) (seq, bool, error) {
	switch n.name {
	case "sum":
		v, err := e.arg(n, 0, f)
		if err != nil {
			return nil, true, err
		}
		var sum float64
		for _, it := range atomizeAll(v) {
			fl, ok := asNumber(it)
			if !ok {
				return nil, true, errUnsupported
			}
			sum += fl
		}
		return seq{sum}, true, nil
	case "distinct-values":
		v, err := e.arg(n, 0, f)
		if err != nil {
			return nil, true, err
		}
		seen := map[string]bool{}
		var out seq
		for _, it := range atomizeAll(v) {
			s := atomString(it)
			if !seen[s] {
				seen[s] = true
				out = append(out, it)
			}
		}
		return out, true, nil
	}
	return nil, false, nil
}

// evalDateFn handles current-* and the *-from-* date/time accessors.
func (e *evaluator) evalDateFn(n *call, f *focus) (seq, bool, error) {
	switch n.name {
	case "current-date":
		return seq{time.Now().Format("2006-01-02")}, true, nil
	case "current-dateTime":
		return seq{time.Now().Format("2006-01-02T15:04:05")}, true, nil
	case "current-time":
		return seq{time.Now().Format("15:04:05")}, true, nil
	case "year-from-date", "month-from-date", "day-from-date",
		"year-from-dateTime", "month-from-dateTime", "day-from-dateTime",
		"hours-from-dateTime", "minutes-from-dateTime", "seconds-from-dateTime",
		"hours-from-time", "minutes-from-time", "seconds-from-time":
		v, err := e.arg(n, 0, f)
		if err != nil {
			return nil, true, err
		}
		atoms := atomizeAll(v)
		if len(atoms) == 0 {
			return seq{}, true, nil // fn:*-from-*(()) is the empty sequence
		}
		comp, ok := dateComponent(n.name, atomString(atoms[0]))
		if !ok {
			// The argument is already a validated date/time value, so a parse
			// failure means our reader is too strict — fail open, never reject.
			return nil, true, errUnsupported
		}
		return seq{comp}, true, nil
	}
	return nil, false, nil
}

// evalNodeFn handles the node-name accessors.
func (e *evaluator) evalNodeFn(n *call, f *focus) (seq, bool, error) {
	switch n.name {
	case "local-name", "name":
		var el *nodeCtx
		if len(n.args) == 0 {
			c, ok := f.node()
			if !ok {
				return seq{""}, true, nil
			}
			el = c
		} else {
			v, err := e.arg(n, 0, f)
			if err != nil {
				return nil, true, err
			}
			if len(v) == 0 {
				return seq{""}, true, nil
			}
			c, ok := v[0].(*nodeCtx)
			if !ok {
				return seq{""}, true, nil
			}
			el = c
		}
		return seq{el.node.NodeName().Local}, true, nil
	}
	return nil, false, nil
}

func (e *evaluator) arg(n *call, i int, f *focus) (seq, error) {
	if i >= len(n.args) {
		return nil, errUnsupported
	}
	return e.eval(n.args[i], f)
}

func (e *evaluator) strArgOrCtx(n *call, f *focus) (string, error) {
	if len(n.args) == 0 {
		return focusString(f), nil
	}
	v, err := e.arg(n, 0, f)
	if err != nil {
		return "", err
	}
	return seqString(v), nil
}

// focusString is the string value of the context item.
func focusString(f *focus) string {
	if f == nil || f.item == nil {
		return ""
	}
	return atomString(f.item)
}

// ---- atomization / coercion ----

func atomizeAll(s seq) []item {
	out := make([]item, 0, len(s))
	for _, it := range s {
		switch v := it.(type) {
		case *nodeCtx:
			out = append(out, v.node.StringValue())
		case *attrItem:
			out = append(out, v.attr.AttrValue())
		default:
			out = append(out, it)
		}
	}
	return out
}

func atomString(it item) string {
	switch v := it.(type) {
	case string:
		return v
	case typedItem:
		return v.lex
	case float64:
		return formatNumber(v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case *nodeCtx:
		return v.node.StringValue()
	case *attrItem:
		return v.attr.AttrValue()
	}
	return ""
}

func asNumber(it item) (float64, bool) {
	switch v := it.(type) {
	case float64:
		return v, true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return 0, false
		}
		return f, true
	case typedItem:
		// A string-typed atom never coerces to a number: an xs:string value of
		// "100" must compare as a string, not as the number 100.
		if v.kind == KindString {
			return 0, false
		}
		f, err := strconv.ParseFloat(strings.TrimSpace(v.lex), 64)
		if err != nil {
			return 0, false
		}
		return f, true
	}
	return 0, false
}

func toNumber(s seq) (float64, error) {
	a := atomizeAll(s)
	if len(a) != 1 {
		return 0, errUnsupported
	}
	f, ok := asNumber(a[0])
	if !ok {
		return 0, errUnsupported
	}
	return f, nil
}

func seqString(s seq) string {
	if len(s) == 0 {
		return ""
	}
	return atomString(s[0])
}

func formatNumber(f float64) string {
	if f == float64(int64(f)) {
		return strconv.FormatInt(int64(f), 10)
	}
	return strconv.FormatFloat(f, 'g', -1, 64)
}

func collapse(s string) string { return strings.Join(strings.Fields(s), " ") }

// dateComponent extracts the calendar component named by an fn:*-from-*
// function (year/month/day-from-date·DateTime, hours/minutes/seconds-from-time·
// DateTime) from an ISO date/dateTime/time lexical. The lexical comes from an
// already-validated date/time value, so it is parsed by the rigorous
// xsdtemporal reader shared with the value layer; ok is false when parsing
// fails, so the caller fails open.
func dateComponent(fn, lex string) (float64, bool) {
	lex = strings.TrimSpace(lex)
	var (
		dt  *xsdtemporal.DateTime
		err error
	)
	switch {
	case strings.Contains(fn, "from-dateTime"):
		dt, err = xsdtemporal.ParseDateTime(lex)
	case strings.Contains(fn, "from-time"):
		dt, err = xsdtemporal.ParseTime(lex)
	default: // *-from-date
		dt, err = xsdtemporal.ParseDate(lex)
	}
	if err != nil {
		return 0, false
	}
	switch fn[:strings.IndexByte(fn, '-')] {
	case "year":
		return float64(dt.Year), true
	case "month":
		return float64(dt.Month), true
	case "day":
		return float64(dt.Day), true
	case "hours":
		return float64(dt.Hour), true
	case "minutes":
		return float64(dt.Minute), true
	case "seconds":
		if dt.Second == nil {
			return 0, true
		}
		ss, _ := dt.Second.Float64()
		return ss, true
	}
	return 0, false
}

func localPart(qname string) string {
	if i := strings.LastIndexByte(qname, ':'); i >= 0 {
		return qname[i+1:]
	}
	return qname
}
