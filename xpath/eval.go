package xpath

// A small XPath-2.0-subset evaluator over an abstract instance tree, the
// "evaluator half" of this package (the static parser above feeds schema-time
// type checks; this drives xs:assert and xs:alternative/@test at validation
// time). It is intentionally partial: every unsupported construct returns an
// error so callers can fail open — treat an assertion as satisfied and a
// type-table alternative as unmatched — and never reject an instance for syntax
// the evaluator does not understand.
//
// The evaluator is format-agnostic: it walks the Node / NodeAttr interfaces, so
// any infoset (XML today, JSON/BER later) can be assessed without this package
// depending on it. Name tests match by local name; the static namespace context
// of the expression is not threaded in (callers that need built-in casts supply
// EvalContext.Castable).

import (
	"fmt"
	"strconv"
	"strings"
	"time"
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
	// validated.
	Vars map[string][]string
}

// item is one XPath item: float64, string, bool, Node, or NodeAttr.
type item any

type seq []item

// errUnsupported marks a construct outside the supported subset.
var errUnsupported = fmt.Errorf("xpath: unsupported construct")

// EvalBool evaluates expr against the context node and returns its effective
// boolean value. ok is false when the expression is outside the supported
// subset (or otherwise fails to evaluate), so callers can fail open.
func EvalBool(expr string, node Node, ec EvalContext) (result, ok bool) {
	v, err := evalExpr(expr, node, ec)
	if err != nil {
		return false, false
	}
	return effectiveBool(v), true
}

func evalExpr(expr string, node Node, ec EvalContext) (seq, error) {
	p := &exprParser{toks: lexExpr(expr)}
	root, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if p.cur().kind != tkEOF {
		return nil, errUnsupported
	}
	ev := &evaluator{ec: ec}
	return ev.eval(root, node)
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

// ---- lexer ----

type exprTokKind int

const (
	tkName exprTokKind = iota
	tkNum
	tkStr
	tkOp
	tkEOF
)

type exprTok struct {
	kind exprTokKind
	text string
}

func lexExpr(s string) []exprTok {
	var toks []exprTok
	i := 0
	for i < len(s) {
		c := s[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
		case c == '\'' || c == '"':
			j := i + 1
			for j < len(s) && s[j] != c {
				j++
			}
			toks = append(toks, exprTok{tkStr, s[i+1 : min(j, len(s))]})
			i = j + 1
		case c >= '0' && c <= '9' || (c == '.' && i+1 < len(s) && s[i+1] >= '0' && s[i+1] <= '9'):
			j := i
			for j < len(s) && (s[j] >= '0' && s[j] <= '9' || s[j] == '.') {
				j++
			}
			toks = append(toks, exprTok{tkNum, s[i:j]})
			i = j
		case xnameStart(c):
			j := i
			for j < len(s) && xnameChar(s[j]) {
				j++
			}
			toks = append(toks, exprTok{tkName, s[i:j]})
			i = j
		default:
			two := ""
			if i+1 < len(s) {
				two = s[i : i+2]
			}
			switch two {
			case "//", "<=", ">=", "!=":
				toks = append(toks, exprTok{tkOp, two})
				i += 2
				continue
			}
			toks = append(toks, exprTok{tkOp, string(c)})
			i++
		}
	}
	toks = append(toks, exprTok{tkEOF, ""})
	return toks
}

func xnameStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
func xnameChar(c byte) bool {
	return xnameStart(c) || (c >= '0' && c <= '9') || c == '-' || c == '.' || c == ':'
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
type pathExpr struct {
	fromRoot bool // path begins with "/" or "//": rooted at the document node
	steps    []pathStep
}
type pathStep struct {
	axisAttr bool
	name     string
	dot      bool
	descend  bool
	preds    []exprNode
}
type typeOp struct {
	x    exprNode
	kind string // "castable" or "instance"
	typ  string
}
type seqExpr struct{ items []exprNode } // (e1, e2, ...) sequence construction
type rangeExpr struct{ lo, hi exprNode } // e1 to e2
type varRef struct{ name string }        // $name variable reference

// ---- parser ----

type exprParser struct {
	toks []exprTok
	pos  int
}

func (p *exprParser) cur() exprTok  { return p.toks[p.pos] }
func (p *exprParser) next() exprTok { t := p.toks[p.pos]; p.pos++; return t }
func (p *exprParser) isOp(s string) bool {
	return p.cur().kind == tkOp && p.cur().text == s
}
func (p *exprParser) isKw(s string) bool {
	return p.cur().kind == tkName && p.cur().text == s
}

func (p *exprParser) parseExpr() (exprNode, error) { return p.parseOr() }

func (p *exprParser) parseOr() (exprNode, error) {
	l, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.isKw("or") {
		p.next()
		r, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		l = &binary{"or", l, r}
	}
	return l, nil
}

func (p *exprParser) parseAnd() (exprNode, error) {
	l, err := p.parseComparison()
	if err != nil {
		return nil, err
	}
	for p.isKw("and") {
		p.next()
		r, err := p.parseComparison()
		if err != nil {
			return nil, err
		}
		l = &binary{"and", l, r}
	}
	return l, nil
}

var cmpWords = map[string]bool{"eq": true, "ne": true, "lt": true, "le": true, "gt": true, "ge": true}

func (p *exprParser) parseComparison() (exprNode, error) {
	l, err := p.parseRange()
	if err != nil {
		return nil, err
	}
	t := p.cur()
	switch {
	case t.kind == tkOp && (t.text == "=" || t.text == "!=" || t.text == "<" || t.text == ">" || t.text == "<=" || t.text == ">="):
		p.next()
		r, err := p.parseRange()
		if err != nil {
			return nil, err
		}
		return &binary{t.text, l, r}, nil
	case t.kind == tkName && cmpWords[t.text]:
		p.next()
		r, err := p.parseRange()
		if err != nil {
			return nil, err
		}
		return &binary{t.text, l, r}, nil
	}
	return l, nil
}

// parseRange handles "AdditiveExpr to AdditiveExpr" (XPath 2.0 RangeExpr).
func (p *exprParser) parseRange() (exprNode, error) {
	l, err := p.parseAdditive()
	if err != nil {
		return nil, err
	}
	if p.isKw("to") {
		p.next()
		r, err := p.parseAdditive()
		if err != nil {
			return nil, err
		}
		return &rangeExpr{l, r}, nil
	}
	return l, nil
}

func (p *exprParser) parseAdditive() (exprNode, error) {
	l, err := p.parseMultiplicative()
	if err != nil {
		return nil, err
	}
	for p.isOp("+") || p.isOp("-") {
		op := p.next().text
		r, err := p.parseMultiplicative()
		if err != nil {
			return nil, err
		}
		l = &binary{op, l, r}
	}
	return l, nil
}

func (p *exprParser) parseMultiplicative() (exprNode, error) {
	l, err := p.parseUnion()
	if err != nil {
		return nil, err
	}
	for p.isOp("*") || p.isKw("div") || p.isKw("mod") || p.isKw("idiv") {
		op := p.next().text
		r, err := p.parseUnion()
		if err != nil {
			return nil, err
		}
		l = &binary{op, l, r}
	}
	return l, nil
}

// parseUnion handles "e | e" / "e union e" (XPath 2.0 UnionExpr), which binds
// tighter than the multiplicative operators.
func (p *exprParser) parseUnion() (exprNode, error) {
	l, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.isOp("|") || p.isKw("union") {
		p.next()
		r, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		l = &binary{"|", l, r}
	}
	return l, nil
}

func (p *exprParser) parseUnary() (exprNode, error) {
	if p.isOp("-") || p.isOp("+") {
		op := p.next().text
		x, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &unary{op, x}, nil
	}
	return p.parseTypeExpr()
}

func (p *exprParser) parseTypeExpr() (exprNode, error) {
	x, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	if p.isKw("castable") {
		p.next()
		if !p.isKw("as") {
			return nil, errUnsupported
		}
		p.next()
		typ, err := p.parseTypeName()
		if err != nil {
			return nil, err
		}
		return &typeOp{x, "castable", typ}, nil
	}
	if p.isKw("instance") {
		p.next()
		if !p.isKw("of") {
			return nil, errUnsupported
		}
		p.next()
		typ, err := p.parseTypeName()
		if err != nil {
			return nil, err
		}
		if p.isOp("?") || p.isOp("*") || p.isOp("+") {
			p.next()
		}
		return &typeOp{x, "instance", typ}, nil
	}
	return x, nil
}

func (p *exprParser) parseTypeName() (string, error) {
	if p.cur().kind != tkName {
		return "", errUnsupported
	}
	return localPart(p.next().text), nil
}

func (p *exprParser) parsePrimary() (exprNode, error) {
	t := p.cur()
	switch {
	case t.kind == tkNum:
		p.next()
		f, err := strconv.ParseFloat(t.text, 64)
		if err != nil {
			return nil, errUnsupported
		}
		return &litNum{f}, nil
	case t.kind == tkStr:
		p.next()
		return &litStr{t.text}, nil
	case p.isOp("("):
		p.next()
		if p.isOp(")") {
			p.next()
			return &call{name: "__empty__"}, nil
		}
		inner, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if p.isOp(",") {
			items := []exprNode{inner}
			for p.isOp(",") {
				p.next()
				it, err := p.parseExpr()
				if err != nil {
					return nil, err
				}
				items = append(items, it)
			}
			inner = &seqExpr{items}
		}
		if !p.isOp(")") {
			return nil, errUnsupported
		}
		p.next()
		return inner, nil
	case p.isOp("$"):
		p.next()
		if p.cur().kind != tkName {
			return nil, errUnsupported
		}
		return &varRef{localPart(p.next().text)}, nil
	case p.isKw("if"):
		return p.parseIf()
	case t.kind == tkName && p.peekIsCall():
		return p.parseCall()
	case t.kind == tkName, p.isOp("@"), p.isOp("."), p.isOp("*"), p.isOp("//"), p.isOp("/"):
		return p.parsePath()
	}
	return nil, errUnsupported
}

func (p *exprParser) peekIsCall() bool {
	return p.cur().kind == tkName && p.pos+1 < len(p.toks) &&
		p.toks[p.pos+1].kind == tkOp && p.toks[p.pos+1].text == "("
}

func (p *exprParser) parseIf() (exprNode, error) {
	p.next()
	if !p.isOp("(") {
		return nil, errUnsupported
	}
	p.next()
	cond, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if !p.isOp(")") {
		return nil, errUnsupported
	}
	p.next()
	if !p.isKw("then") {
		return nil, errUnsupported
	}
	p.next()
	then, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if !p.isKw("else") {
		return nil, errUnsupported
	}
	p.next()
	els, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	return &ifExpr{cond, then, els}, nil
}

func (p *exprParser) parseCall() (exprNode, error) {
	name := p.next().text
	p.next() // (
	c := &call{name: name}
	if p.isOp(")") {
		p.next()
		return c, nil
	}
	for {
		arg, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		c.args = append(c.args, arg)
		if p.isOp(",") {
			p.next()
			continue
		}
		break
	}
	if !p.isOp(")") {
		return nil, errUnsupported
	}
	p.next()
	return c, nil
}

func (p *exprParser) parsePath() (exprNode, error) {
	path := &pathExpr{}
	if p.isOp("/") {
		// A leading "/" is rooted at the document node. In an XSD assertion the
		// XDM root is the (parentless) element being assessed, so there is no
		// document node and the path selects nothing; a bare "/" is the root
		// itself, equally absent. Either way evalPath yields the empty sequence.
		p.next()
		path.fromRoot = true
		if !p.stepStarts() {
			return path, nil
		}
	} else if p.isOp("//") {
		p.next()
		path.fromRoot = true
	}
	for {
		st, err := p.parseStep()
		if err != nil {
			return nil, err
		}
		path.steps = append(path.steps, st)
		if p.isOp("/") {
			p.next()
			continue
		}
		if p.isOp("//") {
			p.next()
			next, err := p.parseStep()
			if err != nil {
				return nil, err
			}
			next.descend = true
			path.steps = append(path.steps, next)
			if p.isOp("/") {
				p.next()
				continue
			}
			break
		}
		break
	}
	return path, nil
}

// stepStarts reports whether the current token can begin a path step.
func (p *exprParser) stepStarts() bool {
	return p.cur().kind == tkName || p.isOp("@") || p.isOp(".") || p.isOp("*")
}

func (p *exprParser) parseStep() (pathStep, error) {
	st := pathStep{}
	switch {
	case p.isOp("@"):
		p.next()
		if p.cur().kind == tkName {
			st.axisAttr = true
			st.name = localPart(p.next().text)
		} else if p.isOp("*") {
			p.next()
			st.axisAttr = true
			st.name = "*"
		} else {
			return st, errUnsupported
		}
	case p.isOp("."):
		p.next()
		if p.isOp(".") {
			return st, errUnsupported
		}
		st.dot = true
	case p.isOp("*"):
		p.next()
		st.name = "*"
	case p.cur().kind == tkName:
		st.name = localPart(p.next().text)
	default:
		return st, errUnsupported
	}
	for p.isOp("[") {
		p.next()
		pred, err := p.parseExpr()
		if err != nil {
			return st, err
		}
		if !p.isOp("]") {
			return st, errUnsupported
		}
		p.next()
		st.preds = append(st.preds, pred)
	}
	return st, nil
}

// ---- evaluator ----

type evaluator struct {
	ec EvalContext
}

func (e *evaluator) eval(node exprNode, ctx Node) (seq, error) {
	switch n := node.(type) {
	case *litNum:
		return seq{n.v}, nil
	case *litStr:
		return seq{n.v}, nil
	case *unary:
		v, err := e.eval(n.x, ctx)
		if err != nil {
			return nil, err
		}
		f, err := toNumber(v)
		if err != nil {
			return nil, err
		}
		if n.op == "-" {
			f = -f
		}
		return seq{f}, nil
	case *binary:
		return e.evalBinary(n, ctx)
	case *ifExpr:
		c, err := e.eval(n.cond, ctx)
		if err != nil {
			return nil, err
		}
		if effectiveBool(c) {
			return e.eval(n.then, ctx)
		}
		return e.eval(n.els, ctx)
	case *call:
		return e.evalCall(n, ctx)
	case *pathExpr:
		return e.evalPath(n, ctx)
	case *typeOp:
		return e.evalTypeOp(n, ctx)
	case *seqExpr:
		var out seq
		for _, it := range n.items {
			v, err := e.eval(it, ctx)
			if err != nil {
				return nil, err
			}
			out = append(out, v...)
		}
		return out, nil
	case *varRef:
		vals, ok := e.ec.Vars[n.name]
		if !ok {
			return nil, errUnsupported
		}
		out := make(seq, len(vals))
		for i, v := range vals {
			out[i] = v
		}
		return out, nil
	case *rangeExpr:
		lo, err := e.eval(n.lo, ctx)
		if err != nil {
			return nil, err
		}
		hi, err := e.eval(n.hi, ctx)
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
	return nil, errUnsupported
}

func (e *evaluator) evalBinary(n *binary, ctx Node) (seq, error) {
	switch n.op {
	case "and":
		l, err := e.eval(n.l, ctx)
		if err != nil {
			return nil, err
		}
		if !effectiveBool(l) {
			return seq{false}, nil
		}
		r, err := e.eval(n.r, ctx)
		if err != nil {
			return nil, err
		}
		return seq{effectiveBool(r)}, nil
	case "or":
		l, err := e.eval(n.l, ctx)
		if err != nil {
			return nil, err
		}
		if effectiveBool(l) {
			return seq{true}, nil
		}
		r, err := e.eval(n.r, ctx)
		if err != nil {
			return nil, err
		}
		return seq{effectiveBool(r)}, nil
	case "|":
		l, err := e.eval(n.l, ctx)
		if err != nil {
			return nil, err
		}
		r, err := e.eval(n.r, ctx)
		if err != nil {
			return nil, err
		}
		return unionNodes(l, r)
	case "+", "-", "*", "div", "mod", "idiv":
		l, err := e.eval(n.l, ctx)
		if err != nil {
			return nil, err
		}
		r, err := e.eval(n.r, ctx)
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
	default:
		l, err := e.eval(n.l, ctx)
		if err != nil {
			return nil, err
		}
		r, err := e.eval(n.r, ctx)
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

// unionNodes is the "|" operator: the union of two node sequences, de-duplicated
// by node identity. Both operands must contain only nodes (Node/NodeAttr); any
// atomic value makes it a type error, so the caller fails open.
func unionNodes(l, r seq) (seq, error) {
	seen := map[item]bool{}
	var out seq
	add := func(s seq) error {
		for _, it := range s {
			switch it.(type) {
			case Node, NodeAttr:
			default:
				return errUnsupported
			}
			if !seen[it] {
				seen[it] = true
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

func (e *evaluator) evalPath(n *pathExpr, ctx Node) (seq, error) {
	if n.fromRoot {
		// No document node exists above the assertion's root element, so an
		// absolute path selects nothing (XSD 1.1 §3.13.4.1 "stay in subtree").
		return seq{}, nil
	}
	cur := []Node{ctx}
	for si, st := range n.steps {
		if st.dot {
			if err := e.applyPreds(&cur, st.preds); err != nil {
				return nil, err
			}
			continue
		}
		if st.axisAttr {
			if si != len(n.steps)-1 {
				return nil, errUnsupported
			}
			base := cur
			if st.descend {
				base = descendOrSelfAll(cur)
			}
			var out seq
			for _, el := range base {
				for _, at := range el.NodeAttrs() {
					if st.name == "*" || at.AttrName().Local == st.name {
						out = append(out, at)
					}
				}
			}
			return out, nil
		}
		var next []Node
		base := cur
		if st.descend {
			base = descendOrSelfAll(cur)
		}
		for _, el := range base {
			for _, c := range el.NodeChildren() {
				if st.name == "*" || c.NodeName().Local == st.name {
					next = append(next, c)
				}
			}
		}
		cur = next
		if err := e.applyPreds(&cur, st.preds); err != nil {
			return nil, err
		}
	}
	out := make(seq, len(cur))
	for i, el := range cur {
		out[i] = el
	}
	return out, nil
}

func (e *evaluator) applyPreds(set *[]Node, preds []exprNode) error {
	for _, pred := range preds {
		var kept []Node
		for idx, el := range *set {
			v, err := e.eval(pred, el)
			if err != nil {
				return err
			}
			if len(v) == 1 {
				if f, ok := v[0].(float64); ok {
					if int(f) == idx+1 {
						kept = append(kept, el)
					}
					continue
				}
			}
			if effectiveBool(v) {
				kept = append(kept, el)
			}
		}
		*set = kept
	}
	return nil
}

func (e *evaluator) evalTypeOp(n *typeOp, ctx Node) (seq, error) {
	v, err := e.eval(n.x, ctx)
	if err != nil {
		return nil, err
	}
	atoms := atomizeAll(v)
	if len(atoms) == 0 {
		return seq{false}, nil
	}
	if len(atoms) != 1 || e.ec.Castable == nil {
		return nil, errUnsupported
	}
	return seq{e.ec.Castable(n.typ, atomString(atoms[0]))}, nil
}

func (e *evaluator) evalCall(n *call, ctx Node) (seq, error) {
	switch n.name {
	case "__empty__":
		return seq{}, nil
	case "true":
		return seq{true}, nil
	case "false":
		return seq{false}, nil
	case "not":
		v, err := e.arg(n, 0, ctx)
		if err != nil {
			return nil, err
		}
		return seq{!effectiveBool(v)}, nil
	case "boolean":
		v, err := e.arg(n, 0, ctx)
		if err != nil {
			return nil, err
		}
		return seq{effectiveBool(v)}, nil
	case "exists":
		v, err := e.arg(n, 0, ctx)
		if err != nil {
			return nil, err
		}
		return seq{len(v) > 0}, nil
	case "empty":
		v, err := e.arg(n, 0, ctx)
		if err != nil {
			return nil, err
		}
		return seq{len(v) == 0}, nil
	case "count":
		v, err := e.arg(n, 0, ctx)
		if err != nil {
			return nil, err
		}
		return seq{float64(len(v))}, nil
	case "string":
		if len(n.args) == 0 {
			return seq{ctx.StringValue()}, nil
		}
		v, err := e.arg(n, 0, ctx)
		if err != nil {
			return nil, err
		}
		return seq{seqString(v)}, nil
	case "number":
		v, err := e.arg(n, 0, ctx)
		if err != nil {
			return nil, err
		}
		f, err := toNumber(v)
		if err != nil {
			return nil, err
		}
		return seq{f}, nil
	case "string-length":
		s, err := e.strArgOrCtx(n, ctx)
		if err != nil {
			return nil, err
		}
		return seq{float64(len([]rune(s)))}, nil
	case "normalize-space":
		s, err := e.strArgOrCtx(n, ctx)
		if err != nil {
			return nil, err
		}
		return seq{collapse(s)}, nil
	case "contains", "starts-with", "ends-with":
		a, err := e.arg(n, 0, ctx)
		if err != nil {
			return nil, err
		}
		b, err := e.arg(n, 1, ctx)
		if err != nil {
			return nil, err
		}
		s1, s2 := seqString(a), seqString(b)
		switch n.name {
		case "contains":
			return seq{strings.Contains(s1, s2)}, nil
		case "starts-with":
			return seq{strings.HasPrefix(s1, s2)}, nil
		default:
			return seq{strings.HasSuffix(s1, s2)}, nil
		}
	case "concat":
		var b strings.Builder
		for i := range n.args {
			v, err := e.arg(n, i, ctx)
			if err != nil {
				return nil, err
			}
			b.WriteString(seqString(v))
		}
		return seq{b.String()}, nil
	case "sum":
		v, err := e.arg(n, 0, ctx)
		if err != nil {
			return nil, err
		}
		var sum float64
		for _, it := range atomizeAll(v) {
			f, ok := asNumber(it)
			if !ok {
				return nil, errUnsupported
			}
			sum += f
		}
		return seq{sum}, nil
	case "current-date":
		return seq{time.Now().Format("2006-01-02")}, nil
	case "current-dateTime":
		return seq{time.Now().Format("2006-01-02T15:04:05")}, nil
	case "current-time":
		return seq{time.Now().Format("15:04:05")}, nil
	case "local-name", "name":
		var el Node
		if len(n.args) == 0 {
			el = ctx
		} else {
			v, err := e.arg(n, 0, ctx)
			if err != nil {
				return nil, err
			}
			if len(v) == 0 {
				return seq{""}, nil
			}
			e2, ok := v[0].(Node)
			if !ok {
				return seq{""}, nil
			}
			el = e2
		}
		return seq{el.NodeName().Local}, nil
	}
	// Constructor functions: xs:integer(…), xs:decimal(…), etc. — the argument
	// must be castable to the named built-in type; the cast value is carried as
	// its lexical form, which numeric/string comparisons coerce as needed.
	if strings.IndexByte(n.name, ':') >= 0 && len(n.args) == 1 && e.ec.Castable != nil {
		v, err := e.arg(n, 0, ctx)
		if err != nil {
			return nil, err
		}
		s := seqString(v)
		if !e.ec.Castable(localPart(n.name), s) {
			return nil, errUnsupported
		}
		return seq{s}, nil
	}
	return nil, errUnsupported
}

func (e *evaluator) arg(n *call, i int, ctx Node) (seq, error) {
	if i >= len(n.args) {
		return nil, errUnsupported
	}
	return e.eval(n.args[i], ctx)
}

func (e *evaluator) strArgOrCtx(n *call, ctx Node) (string, error) {
	if len(n.args) == 0 {
		return ctx.StringValue(), nil
	}
	v, err := e.arg(n, 0, ctx)
	if err != nil {
		return "", err
	}
	return seqString(v), nil
}

// ---- atomization / coercion ----

func atomizeAll(s seq) []item {
	out := make([]item, 0, len(s))
	for _, it := range s {
		switch v := it.(type) {
		case Node:
			out = append(out, v.StringValue())
		case NodeAttr:
			out = append(out, v.AttrValue())
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
	case float64:
		return formatNumber(v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case Node:
		return v.StringValue()
	case NodeAttr:
		return v.AttrValue()
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

func descendants(el Node) []Node {
	var out []Node
	for _, c := range el.NodeChildren() {
		out = append(out, c)
		out = append(out, descendants(c)...)
	}
	return out
}

func descendantsOrSelf(el Node) []Node {
	return append([]Node{el}, descendants(el)...)
}

// descendOrSelfAll is the "//" step expansion: descendant-or-self::node() over
// every node in the input set, preserving document order (self before descendants).
func descendOrSelfAll(set []Node) []Node {
	var out []Node
	for _, el := range set {
		out = append(out, descendantsOrSelf(el)...)
	}
	return out
}

func collapse(s string) string { return strings.Join(strings.Fields(s), " ") }

func localPart(qname string) string {
	if i := strings.LastIndexByte(qname, ':'); i >= 0 {
		return qname[i+1:]
	}
	return qname
}
