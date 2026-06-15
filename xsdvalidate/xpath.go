package xsdvalidate

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kud360/goxsd5/builtin"
	"github.com/kud360/goxsd5/xsd"
)

// A small XPath-2.0-subset evaluator over the infoset, used for xs:assert and
// xs:alternative/@test (V4). It is intentionally partial: every unsupported
// construct returns an error, and the callers treat an evaluation error as
// "fail open" (an assertion is considered satisfied; a type-alternative test is
// considered not matched). That guarantees an instance is never rejected for a
// construct this evaluator cannot understand — coverage only ratchets up.
//
// Name tests match by local name (the same namespace approximation the identity
// constraint evaluator documents): the assertion's static namespace context is
// not retained on the compiled component.

// ---- value model ----

// item is one XPath item: float64, string, bool, Element, or Attribute.
type xpItem any

type xpSeq []xpItem

// errUnsupported marks a construct outside the supported subset; it triggers
// fail-open handling in the callers.
var errUnsupported = fmt.Errorf("xpath: unsupported construct")

// evalAssertion reports whether an assertion's test holds for the context
// element. ok is false when the expression could not be evaluated (fail open).
func evalAssertion(ctx Element, expr string) (result, ok bool) {
	v, err := evalXPath(ctx, expr)
	if err != nil {
		return false, false
	}
	return effectiveBool(v), true
}

func evalXPath(ctx Element, expr string) (xpSeq, error) {
	p := &xpParser{toks: lexXPath(expr)}
	node, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if p.cur().kind != tkEOF {
		return nil, errUnsupported
	}
	ev := &xpEval{ctx: ctx}
	return ev.eval(node, ctx)
}

func effectiveBool(seq xpSeq) bool {
	if len(seq) == 0 {
		return false
	}
	if len(seq) == 1 {
		switch v := seq[0].(type) {
		case bool:
			return v
		case float64:
			return v != 0 && !isNaN(v)
		case string:
			return v != ""
		}
	}
	// A non-empty sequence whose first item is a node has EBV true.
	switch seq[0].(type) {
	case Element, Attribute:
		return true
	}
	return true
}

func isNaN(f float64) bool { return f != f }

// ---- lexer ----

type xpTokKind int

const (
	tkName xpTokKind = iota
	tkNum
	tkStr
	tkOp
	tkEOF
)

type xpTok struct {
	kind xpTokKind
	text string
}

func lexXPath(s string) []xpTok {
	var toks []xpTok
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
			toks = append(toks, xpTok{tkStr, s[i+1 : min(j, len(s))]})
			i = j + 1
		case c >= '0' && c <= '9' || (c == '.' && i+1 < len(s) && s[i+1] >= '0' && s[i+1] <= '9'):
			j := i
			for j < len(s) && (s[j] >= '0' && s[j] <= '9' || s[j] == '.') {
				j++
			}
			toks = append(toks, xpTok{tkNum, s[i:j]})
			i = j
		case isNameStart(c):
			j := i
			for j < len(s) && isNameChar(s[j]) {
				j++
			}
			toks = append(toks, xpTok{tkName, s[i:j]})
			i = j
		default:
			// multi-char operators first
			two := ""
			if i+1 < len(s) {
				two = s[i : i+2]
			}
			switch two {
			case "//", "<=", ">=", "!=":
				toks = append(toks, xpTok{tkOp, two})
				i += 2
				continue
			}
			toks = append(toks, xpTok{tkOp, string(c)})
			i++
		}
	}
	toks = append(toks, xpTok{tkEOF, ""})
	return toks
}

func isNameStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
func isNameChar(c byte) bool {
	return isNameStart(c) || (c >= '0' && c <= '9') || c == '-' || c == '.' || c == ':'
}

// ---- AST ----

type xpNode interface{}

type xpBinary struct {
	op   string
	l, r xpNode
}
type xpUnary struct {
	op string
	x  xpNode
}
type xpLiteralNum struct{ v float64 }
type xpLiteralStr struct{ v string }
type xpIf struct{ cond, then, els xpNode }
type xpCall struct {
	name string
	args []xpNode
}
type xpPath struct {
	descendant bool // leading "//"
	steps      []xpStep
}
type xpStep struct {
	axisAttr bool
	name     string // "" with axisAttr=false and dot=true means "."
	dot      bool
	descend  bool // this step preceded by "//"
	preds    []xpNode
}
type xpTypeOp struct { // castable as / instance of
	x    xpNode
	kind string // "castable" or "instance"
	typ  string // local type name
}

// ---- parser ----

type xpParser struct {
	toks []xpTok
	pos  int
}

func (p *xpParser) cur() xpTok  { return p.toks[p.pos] }
func (p *xpParser) next() xpTok { t := p.toks[p.pos]; p.pos++; return t }
func (p *xpParser) isOp(s string) bool {
	return p.cur().kind == tkOp && p.cur().text == s
}
func (p *xpParser) isKw(s string) bool {
	return p.cur().kind == tkName && p.cur().text == s
}

func (p *xpParser) parseExpr() (xpNode, error) { return p.parseOr() }

func (p *xpParser) parseOr() (xpNode, error) {
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
		l = &xpBinary{"or", l, r}
	}
	return l, nil
}

func (p *xpParser) parseAnd() (xpNode, error) {
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
		l = &xpBinary{"and", l, r}
	}
	return l, nil
}

var cmpWords = map[string]bool{"eq": true, "ne": true, "lt": true, "le": true, "gt": true, "ge": true}

func (p *xpParser) parseComparison() (xpNode, error) {
	l, err := p.parseAdditive()
	if err != nil {
		return nil, err
	}
	t := p.cur()
	switch {
	case t.kind == tkOp && (t.text == "=" || t.text == "!=" || t.text == "<" || t.text == ">" || t.text == "<=" || t.text == ">="):
		p.next()
		r, err := p.parseAdditive()
		if err != nil {
			return nil, err
		}
		return &xpBinary{t.text, l, r}, nil
	case t.kind == tkName && cmpWords[t.text]:
		p.next()
		r, err := p.parseAdditive()
		if err != nil {
			return nil, err
		}
		return &xpBinary{t.text, l, r}, nil
	}
	return l, nil
}

func (p *xpParser) parseAdditive() (xpNode, error) {
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
		l = &xpBinary{op, l, r}
	}
	return l, nil
}

func (p *xpParser) parseMultiplicative() (xpNode, error) {
	l, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.isOp("*") || p.isKw("div") || p.isKw("mod") || p.isKw("idiv") {
		op := p.next().text
		r, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		l = &xpBinary{op, l, r}
	}
	return l, nil
}

func (p *xpParser) parseUnary() (xpNode, error) {
	if p.isOp("-") || p.isOp("+") {
		op := p.next().text
		x, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &xpUnary{op, x}, nil
	}
	return p.parseTypeExpr()
}

func (p *xpParser) parseTypeExpr() (xpNode, error) {
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
		return &xpTypeOp{x, "castable", typ}, nil
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
		// optional occurrence indicator
		if p.isOp("?") || p.isOp("*") || p.isOp("+") {
			p.next()
		}
		return &xpTypeOp{x, "instance", typ}, nil
	}
	return x, nil
}

func (p *xpParser) parseTypeName() (string, error) {
	if p.cur().kind != tkName {
		return "", errUnsupported
	}
	return localPart(p.next().text), nil
}

func (p *xpParser) parsePrimary() (xpNode, error) {
	t := p.cur()
	switch {
	case t.kind == tkNum:
		p.next()
		f, err := strconv.ParseFloat(t.text, 64)
		if err != nil {
			return nil, errUnsupported
		}
		return &xpLiteralNum{f}, nil
	case t.kind == tkStr:
		p.next()
		return &xpLiteralStr{t.text}, nil
	case p.isOp("("):
		p.next()
		if p.isOp(")") { // empty sequence ()
			p.next()
			return &xpCall{name: "__empty__"}, nil
		}
		inner, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if !p.isOp(")") {
			return nil, errUnsupported
		}
		p.next()
		return inner, nil
	case p.isKw("if"):
		return p.parseIf()
	case t.kind == tkName && p.peekIsCall():
		return p.parseCall()
	case t.kind == tkName, p.isOp("@"), p.isOp("."), p.isOp("*"), p.isOp("//"), p.isOp("/"):
		return p.parsePath()
	}
	return nil, errUnsupported
}

func (p *xpParser) peekIsCall() bool {
	// a name immediately followed by '(' is a function call (no reserved-word
	// node test like text()/node() supported here → those fail open).
	return p.cur().kind == tkName && p.pos+1 < len(p.toks) &&
		p.toks[p.pos+1].kind == tkOp && p.toks[p.pos+1].text == "("
}

func (p *xpParser) parseIf() (xpNode, error) {
	p.next() // if
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
	return &xpIf{cond, then, els}, nil
}

func (p *xpParser) parseCall() (xpNode, error) {
	name := p.next().text
	p.next() // (
	call := &xpCall{name: name}
	if p.isOp(")") {
		p.next()
		return call, nil
	}
	for {
		arg, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		call.args = append(call.args, arg)
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
	return call, nil
}

func (p *xpParser) parsePath() (xpNode, error) {
	path := &xpPath{}
	if p.isOp("/") { // absolute path — root not reachable from the infoset
		return nil, errUnsupported
	}
	if p.isOp("//") {
		p.next()
		path.descendant = true
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
			// mark the next step as a descendant step
			next, err := p.parseStep()
			if err != nil {
				return nil, err
			}
			next.descend = true
			path.steps = append(path.steps, next)
			if p.isOp("/") || p.isOp("//") {
				// chained — keep looping
				if p.isOp("/") {
					p.next()
				}
				continue
			}
			break
		}
		break
	}
	return path, nil
}

func (p *xpParser) parseStep() (xpStep, error) {
	st := xpStep{}
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
			return st, errUnsupported // ".." parent axis unsupported
		}
		st.dot = true
	case p.isOp("*"):
		p.next()
		st.name = "*"
	case p.cur().kind == tkName:
		// reject "child::" style axes (name followed by "::")
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

type xpEval struct {
	ctx Element
}

func (e *xpEval) eval(node xpNode, ctx Element) (xpSeq, error) {
	switch n := node.(type) {
	case *xpLiteralNum:
		return xpSeq{n.v}, nil
	case *xpLiteralStr:
		return xpSeq{n.v}, nil
	case *xpUnary:
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
		return xpSeq{f}, nil
	case *xpBinary:
		return e.evalBinary(n, ctx)
	case *xpIf:
		c, err := e.eval(n.cond, ctx)
		if err != nil {
			return nil, err
		}
		if effectiveBool(c) {
			return e.eval(n.then, ctx)
		}
		return e.eval(n.els, ctx)
	case *xpCall:
		return e.evalCall(n, ctx)
	case *xpPath:
		return e.evalPath(n, ctx)
	case *xpTypeOp:
		return e.evalTypeOp(n, ctx)
	}
	return nil, errUnsupported
}

func (e *xpEval) evalBinary(n *xpBinary, ctx Element) (xpSeq, error) {
	switch n.op {
	case "and":
		l, err := e.eval(n.l, ctx)
		if err != nil {
			return nil, err
		}
		if !effectiveBool(l) {
			return xpSeq{false}, nil
		}
		r, err := e.eval(n.r, ctx)
		if err != nil {
			return nil, err
		}
		return xpSeq{effectiveBool(r)}, nil
	case "or":
		l, err := e.eval(n.l, ctx)
		if err != nil {
			return nil, err
		}
		if effectiveBool(l) {
			return xpSeq{true}, nil
		}
		r, err := e.eval(n.r, ctx)
		if err != nil {
			return nil, err
		}
		return xpSeq{effectiveBool(r)}, nil
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
		return xpSeq{arith(n.op, lf, rf)}, nil
	default: // comparisons
		l, err := e.eval(n.l, ctx)
		if err != nil {
			return nil, err
		}
		r, err := e.eval(n.r, ctx)
		if err != nil {
			return nil, err
		}
		res, err := compare(n.op, l, r)
		if err != nil {
			return nil, err
		}
		return xpSeq{res}, nil
	}
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

// compare implements both general comparison (=, !=, <, …) over sequences and
// value comparison (eq, ne, …) over singletons.
func compare(op string, l, r xpSeq) (bool, error) {
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
	// general comparison: exists a pair that satisfies op
	for _, a := range la {
		for _, b := range ra {
			if cmpAtoms(op, a, b) {
				return true, nil
			}
		}
	}
	return false, nil
}

func cmpAtoms(op string, a, b xpItem) bool {
	// numeric comparison when both atoms are numbers (or numeric strings)
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

func (e *xpEval) evalPath(n *xpPath, ctx Element) (xpSeq, error) {
	cur := []Element{ctx}
	if n.descendant {
		cur = descendantsOrSelf(ctx)
	}
	var attrResult []Attribute
	for si, st := range n.steps {
		if st.dot {
			// context unchanged
			if err := e.applyPreds(&cur, st.preds, ctx); err != nil {
				return nil, err
			}
			continue
		}
		if st.axisAttr {
			if si != len(n.steps)-1 {
				return nil, errUnsupported // attribute must be the final step
			}
			for _, el := range cur {
				for _, at := range el.Attributes() {
					if at.Name().Namespace == xsd.XMLNSNS {
						continue
					}
					if st.name == "*" || at.Name().Local == st.name {
						attrResult = append(attrResult, at)
					}
				}
			}
			out := make(xpSeq, len(attrResult))
			for i, a := range attrResult {
				out[i] = a
			}
			return out, nil
		}
		var next []Element
		base := cur
		if st.descend {
			var d []Element
			for _, el := range cur {
				d = append(d, descendants(el)...)
			}
			base = d
		}
		for _, el := range base {
			for _, c := range elementChildren(el) {
				if st.name == "*" || c.Name().Local == st.name {
					next = append(next, c)
				}
			}
		}
		cur = next
		if err := e.applyPreds(&cur, st.preds, ctx); err != nil {
			return nil, err
		}
	}
	out := make(xpSeq, len(cur))
	for i, el := range cur {
		out[i] = el
	}
	return out, nil
}

func (e *xpEval) applyPreds(set *[]Element, preds []xpNode, _ Element) error {
	for _, pred := range preds {
		var kept []Element
		for idx, el := range *set {
			v, err := e.eval(pred, el)
			if err != nil {
				return err
			}
			// numeric predicate → positional; else EBV
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

func (e *xpEval) evalTypeOp(n *xpTypeOp, ctx Element) (xpSeq, error) {
	v, err := e.eval(n.x, ctx)
	if err != nil {
		return nil, err
	}
	t := builtin.Lookup(n.typ)
	if t == nil {
		return nil, errUnsupported
	}
	atoms := atomizeAll(v)
	if len(atoms) == 0 {
		return xpSeq{n.kind == "instance" && false}, nil
	}
	if len(atoms) != 1 {
		return nil, errUnsupported
	}
	s := atomString(atoms[0])
	_, perr := t.ParseValue(s, nsContext{ctx})
	return xpSeq{perr == nil}, nil
}

func (e *xpEval) evalCall(n *xpCall, ctx Element) (xpSeq, error) {
	switch n.name {
	case "__empty__":
		return xpSeq{}, nil
	case "true":
		return xpSeq{true}, nil
	case "false":
		return xpSeq{false}, nil
	case "not":
		v, err := e.arg(n, 0, ctx)
		if err != nil {
			return nil, err
		}
		return xpSeq{!effectiveBool(v)}, nil
	case "boolean":
		v, err := e.arg(n, 0, ctx)
		if err != nil {
			return nil, err
		}
		return xpSeq{effectiveBool(v)}, nil
	case "exists":
		v, err := e.arg(n, 0, ctx)
		if err != nil {
			return nil, err
		}
		return xpSeq{len(v) > 0}, nil
	case "empty":
		v, err := e.arg(n, 0, ctx)
		if err != nil {
			return nil, err
		}
		return xpSeq{len(v) == 0}, nil
	case "count":
		v, err := e.arg(n, 0, ctx)
		if err != nil {
			return nil, err
		}
		return xpSeq{float64(len(v))}, nil
	case "string":
		if len(n.args) == 0 {
			return xpSeq{nodeString(ctx)}, nil
		}
		v, err := e.arg(n, 0, ctx)
		if err != nil {
			return nil, err
		}
		return xpSeq{seqString(v)}, nil
	case "number":
		v, err := e.arg(n, 0, ctx)
		if err != nil {
			return nil, err
		}
		f, err := toNumber(v)
		if err != nil {
			return nil, err
		}
		return xpSeq{f}, nil
	case "string-length":
		var s string
		if len(n.args) == 0 {
			s = nodeString(ctx)
		} else {
			v, err := e.arg(n, 0, ctx)
			if err != nil {
				return nil, err
			}
			s = seqString(v)
		}
		return xpSeq{float64(len([]rune(s)))}, nil
	case "normalize-space":
		var s string
		if len(n.args) == 0 {
			s = nodeString(ctx)
		} else {
			v, err := e.arg(n, 0, ctx)
			if err != nil {
				return nil, err
			}
			s = seqString(v)
		}
		return xpSeq{collapse(s)}, nil
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
			return xpSeq{strings.Contains(s1, s2)}, nil
		case "starts-with":
			return xpSeq{strings.HasPrefix(s1, s2)}, nil
		default:
			return xpSeq{strings.HasSuffix(s1, s2)}, nil
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
		return xpSeq{b.String()}, nil
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
		return xpSeq{sum}, nil
	case "local-name", "name":
		var el Element
		if len(n.args) == 0 {
			el = ctx
		} else {
			v, err := e.arg(n, 0, ctx)
			if err != nil {
				return nil, err
			}
			if len(v) == 0 {
				return xpSeq{""}, nil
			}
			if e2, ok := v[0].(Element); ok {
				el = e2
			} else {
				return xpSeq{""}, nil
			}
		}
		return xpSeq{el.Name().Local}, nil
	}
	// Constructor functions: xs:integer(…), xs:decimal(…), xs:date(…), etc.
	// The argument is cast to the named built-in type; an uncastable value is an
	// evaluation error (fail open). The cast value is carried as its lexical
	// form, which the numeric/string comparisons coerce as needed.
	if strings.IndexByte(n.name, ':') >= 0 && len(n.args) == 1 {
		if t := builtin.Lookup(localPart(n.name)); t != nil {
			v, err := e.arg(n, 0, ctx)
			if err != nil {
				return nil, err
			}
			s := seqString(v)
			if _, perr := t.ParseValue(s, nsContext{ctx}); perr != nil {
				return nil, perr
			}
			return xpSeq{s}, nil
		}
	}
	return nil, errUnsupported
}

func (e *xpEval) arg(n *xpCall, i int, ctx Element) (xpSeq, error) {
	if i >= len(n.args) {
		return nil, errUnsupported
	}
	return e.eval(n.args[i], ctx)
}

// ---- atomization / coercion ----

func atomizeAll(seq xpSeq) []xpItem {
	out := make([]xpItem, 0, len(seq))
	for _, it := range seq {
		switch v := it.(type) {
		case Element:
			out = append(out, nodeString(v))
		case Attribute:
			out = append(out, v.Value())
		default:
			out = append(out, it)
		}
	}
	return out
}

func atomString(it xpItem) string {
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
	case Element:
		return nodeString(v)
	case Attribute:
		return v.Value()
	}
	return ""
}

func asNumber(it xpItem) (float64, bool) {
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

func toNumber(seq xpSeq) (float64, error) {
	a := atomizeAll(seq)
	if len(a) != 1 {
		return 0, errUnsupported
	}
	f, ok := asNumber(a[0])
	if !ok {
		return 0, errUnsupported
	}
	return f, nil
}

func seqString(seq xpSeq) string {
	if len(seq) == 0 {
		return ""
	}
	return atomString(seq[0])
}

func formatNumber(f float64) string {
	if f == float64(int64(f)) {
		return strconv.FormatInt(int64(f), 10)
	}
	return strconv.FormatFloat(f, 'g', -1, 64)
}

// nodeString is the XPath string value of an element: the concatenation of all
// descendant character data.
func nodeString(el Element) string {
	var b strings.Builder
	var walk func(e Element)
	walk = func(e Element) {
		for _, c := range e.Children() {
			switch c := c.(type) {
			case Text:
				b.WriteString(c.Data())
			case Element:
				walk(c)
			}
		}
	}
	walk(el)
	return b.String()
}

func descendants(el Element) []Element {
	var out []Element
	for _, c := range elementChildren(el) {
		out = append(out, c)
		out = append(out, descendants(c)...)
	}
	return out
}

func descendantsOrSelf(el Element) []Element {
	return append([]Element{el}, descendants(el)...)
}
