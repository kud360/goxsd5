package xpath

// This file implements the XPath 2.0 grammar (REC-xpath20-20101214 Appendix A.1)
// as a recursive-descent parser. Production numbers from the EBNF are cited in
// comments. The parser does two jobs from one pass: it validates syntax (and
// records cast/castable/treat/instance-of type references on p.expr for the
// schema-time checks driven by Parse), and it builds the evaluable AST consumed
// by the evaluator (evalExpr). The AST node types live in eval.go.
//
// Constructs outside the evaluator's subset (treat-as, intersect/except, node
// comparison, non-atomic sequence types, kind-test node tests, …) are parsed for
// validity but flagged via p.markUnsupported(); the evaluator then fails open on
// the whole expression, matching the all-or-nothing behaviour of the former
// dedicated evaluator parser.

import "strconv"

// [2] Expr ::= ExprSingle ("," ExprSingle)*
func (p *parser) parseExpr() exprNode {
	first := p.parseExprSingle()
	if p.failed() || p.cur().kind != tokComma {
		return first
	}
	items := []exprNode{first}
	for !p.failed() && p.cur().kind == tokComma {
		p.advance()
		items = append(items, p.parseExprSingle())
	}
	return &seqExpr{items}
}

// [3] ExprSingle ::= ForExpr | QuantifiedExpr | IfExpr | OrExpr
func (p *parser) parseExprSingle() exprNode {
	if p.failed() {
		return nil
	}
	t := p.cur()
	switch {
	case isName(t, "for") && p.peek(1).kind == tokDollar:
		return p.parseForExpr()
	case (isName(t, "some") || isName(t, "every")) && p.peek(1).kind == tokDollar:
		return p.parseQuantifiedExpr()
	case isName(t, "if") && p.peek(1).kind == tokLParen:
		return p.parseIfExpr()
	default:
		return p.parseOrExpr()
	}
}

// [4] ForExpr ::= SimpleForClause "return" ExprSingle
// [5] SimpleForClause ::= "for" "$" VarName ("in" ExprSingle ("," "$" VarName "in" ExprSingle)*)
func (p *parser) parseForExpr() exprNode {
	p.advance() // "for"
	binds := p.parseBindings()
	if !p.expectName("return") {
		return nil
	}
	return &forExpr{binds, p.parseExprSingle()}
}

// [6] QuantifiedExpr ::= ("some" | "every") "$" VarName "in" ExprSingle
//
//	("," "$" VarName "in" ExprSingle)* "satisfies" ExprSingle
func (p *parser) parseQuantifiedExpr() exprNode {
	every := p.advance().text == "every"
	binds := p.parseBindings()
	if !p.expectName("satisfies") {
		return nil
	}
	return &quantified{every, binds, p.parseExprSingle()}
}

// parseBindings parses the shared "$" VarName "in" ExprSingle ("," …)* clause of
// for/some/every.
func (p *parser) parseBindings() []binding {
	var binds []binding
	for {
		p.expect(tokDollar, "'$'")
		name := p.parseVarName()
		if !p.expectName("in") {
			return binds
		}
		seq := p.parseExprSingle()
		binds = append(binds, binding{name, seq})
		if p.failed() || p.cur().kind != tokComma {
			return binds
		}
		p.advance() // ","
	}
}

// [7] IfExpr ::= "if" "(" Expr ")" "then" ExprSingle "else" ExprSingle
func (p *parser) parseIfExpr() exprNode {
	p.advance() // "if"
	p.expect(tokLParen, "'('")
	cond := p.parseExpr()
	p.expect(tokRParen, "')'")
	if !p.expectName("then") {
		return nil
	}
	then := p.parseExprSingle()
	if !p.expectName("else") {
		return nil
	}
	return &ifExpr{cond, then, p.parseExprSingle()}
}

// [8] OrExpr ::= AndExpr ( "or" AndExpr )*
func (p *parser) parseOrExpr() exprNode {
	l := p.parseAndExpr()
	for !p.failed() && isName(p.cur(), "or") {
		p.advance()
		l = &binary{"or", l, p.parseAndExpr()}
	}
	return l
}

// [9] AndExpr ::= ComparisonExpr ( "and" ComparisonExpr )*
func (p *parser) parseAndExpr() exprNode {
	l := p.parseComparisonExpr()
	for !p.failed() && isName(p.cur(), "and") {
		p.advance()
		l = &binary{"and", l, p.parseComparisonExpr()}
	}
	return l
}

// [10] ComparisonExpr ::= RangeExpr ( (ValueComp | GeneralComp | NodeComp) RangeExpr )?
// [22] GeneralComp ::= "=" | "!=" | "<" | "<=" | ">" | ">="
// [23] ValueComp   ::= "eq" | "ne" | "lt" | "le" | "gt" | "ge"
// [24] NodeComp    ::= "is" | "<<" | ">>"
func (p *parser) parseComparisonExpr() exprNode {
	l := p.parseRangeExpr()
	if p.failed() {
		return l
	}
	t := p.cur()
	op, ok := comparisonOp(t)
	if !ok {
		return l
	}
	// Node comparisons (is / << / >>) are outside the evaluator's subset.
	if t.kind == tokLtLt || t.kind == tokGtGt || isName(t, "is") {
		p.markUnsupported()
	}
	p.advance()
	return &binary{op, l, p.parseRangeExpr()}
}

// comparisonOp reports whether t is a comparison operator and returns the
// canonical op string the evaluator expects (symbol form for general/node
// comparisons, the word itself for value comparisons).
func comparisonOp(t token) (string, bool) {
	switch t.kind {
	case tokEq:
		return "=", true
	case tokNe:
		return "!=", true
	case tokLt:
		return "<", true
	case tokLe:
		return "<=", true
	case tokGt:
		return ">", true
	case tokGe:
		return ">=", true
	case tokLtLt:
		return "<<", true
	case tokGtGt:
		return ">>", true
	}
	if t.kind == tokName {
		switch t.text {
		case "eq", "ne", "lt", "le", "gt", "ge", "is":
			return t.text, true
		}
	}
	return "", false
}

// [11] RangeExpr ::= AdditiveExpr ( "to" AdditiveExpr )?
func (p *parser) parseRangeExpr() exprNode {
	l := p.parseAdditiveExpr()
	if !p.failed() && isName(p.cur(), "to") {
		p.advance()
		return &rangeExpr{l, p.parseAdditiveExpr()}
	}
	return l
}

// [12] AdditiveExpr ::= MultiplicativeExpr ( ("+" | "-") MultiplicativeExpr )*
func (p *parser) parseAdditiveExpr() exprNode {
	l := p.parseMultiplicativeExpr()
	for !p.failed() && (p.cur().kind == tokPlus || p.cur().kind == tokMinus) {
		op := "+"
		if p.advance().kind == tokMinus {
			op = "-"
		}
		l = &binary{op, l, p.parseMultiplicativeExpr()}
	}
	return l
}

// [13] MultiplicativeExpr ::= UnionExpr ( ("*" | "div" | "idiv" | "mod") UnionExpr )*
func (p *parser) parseMultiplicativeExpr() exprNode {
	l := p.parseUnionExpr()
	for !p.failed() {
		t := p.cur()
		var op string
		switch {
		case t.kind == tokStar:
			op = "*"
		case isName(t, "div"), isName(t, "idiv"), isName(t, "mod"):
			op = t.text
		default:
			return l
		}
		p.advance()
		l = &binary{op, l, p.parseUnionExpr()}
	}
	return l
}

// [14] UnionExpr ::= IntersectExceptExpr ( ("union" | "|") IntersectExceptExpr )*
func (p *parser) parseUnionExpr() exprNode {
	l := p.parseIntersectExceptExpr()
	for !p.failed() && (p.cur().kind == tokPipe || isName(p.cur(), "union")) {
		p.advance()
		l = &binary{"|", l, p.parseIntersectExceptExpr()}
	}
	return l
}

// [15] IntersectExceptExpr ::= InstanceofExpr ( ("intersect" | "except") InstanceofExpr )*
func (p *parser) parseIntersectExceptExpr() exprNode {
	l := p.parseInstanceofExpr()
	for !p.failed() && (isName(p.cur(), "intersect") || isName(p.cur(), "except")) {
		p.markUnsupported() // intersect/except are outside the evaluator's subset
		p.advance()
		l = &binary{"intersect", l, p.parseInstanceofExpr()}
	}
	return l
}

// [16] InstanceofExpr ::= TreatExpr ( "instance" "of" SequenceType )?
func (p *parser) parseInstanceofExpr() exprNode {
	x := p.parseTreatExpr()
	if p.failed() || !isName(p.cur(), "instance") {
		return x
	}
	p.advance()
	if !p.expectName("of") {
		return x
	}
	local, atomic := p.parseSequenceType(InstanceOf)
	if !atomic {
		p.markUnsupported()
	}
	return &typeOp{x, "instance", local}
}

// [17] TreatExpr ::= CastableExpr ( "treat" "as" SequenceType )?
func (p *parser) parseTreatExpr() exprNode {
	x := p.parseCastableExpr()
	if p.failed() || !isName(p.cur(), "treat") {
		return x
	}
	p.markUnsupported() // treat-as is outside the evaluator's subset
	p.advance()
	if !p.expectName("as") {
		return x
	}
	p.parseSequenceType(Treat)
	return x
}

// [18] CastableExpr ::= CastExpr ( "castable" "as" SingleType )?
func (p *parser) parseCastableExpr() exprNode {
	x := p.parseCastExpr()
	if p.failed() || !isName(p.cur(), "castable") {
		return x
	}
	p.advance()
	if !p.expectName("as") {
		return x
	}
	return &typeOp{x, "castable", p.parseSingleType(Castable)}
}

// [19] CastExpr ::= UnaryExpr ( "cast" "as" SingleType )?
func (p *parser) parseCastExpr() exprNode {
	x := p.parseUnaryExpr()
	if p.failed() || !isName(p.cur(), "cast") {
		return x
	}
	p.advance()
	if !p.expectName("as") {
		return x
	}
	return &typeOp{x, "cast", p.parseSingleType(Cast)}
}

// [20] UnaryExpr ::= ("-" | "+")* ValueExpr
// [21] ValueExpr ::= PathExpr
func (p *parser) parseUnaryExpr() exprNode {
	var ops []string
	for p.cur().kind == tokMinus || p.cur().kind == tokPlus {
		if p.advance().kind == tokMinus {
			ops = append(ops, "-")
		} else {
			ops = append(ops, "+")
		}
	}
	x := p.parsePathExpr()
	for i := len(ops) - 1; i >= 0; i-- {
		x = &unary{ops[i], x}
	}
	return x
}

// [25] PathExpr ::= ("/" RelativePathExpr?) | ("//" RelativePathExpr) | RelativePathExpr
//
//	(xgc: leading-lone-slash)
func (p *parser) parsePathExpr() exprNode {
	switch p.cur().kind {
	case tokSlash:
		p.advance()
		path := &pathExpr{fromRoot: true}
		if canStartStep(p.cur()) {
			if prim, isStep := p.parseStepExpr(path); !isStep {
				path.start = prim
			}
			p.parseTrailingSteps(path)
		}
		// otherwise a complete path expression "/" (the root).
		return path
	case tokSlashSlash:
		p.advance()
		path := &pathExpr{fromRoot: true}
		path.steps = append(path.steps, pathStep{axis: axDescendantOrSelf, test: nodeTest{kind: tnNode}})
		if prim, isStep := p.parseStepExpr(path); !isStep {
			path.start = prim
		}
		p.parseTrailingSteps(path)
		return path
	default:
		return p.parseRelativePathExpr()
	}
}

// canStartStep reports whether t can begin a StepExpr, used for the
// leading-lone-slash disambiguation.
func canStartStep(t token) bool {
	switch t.kind {
	case tokName, tokStar, tokAt, tokDot, tokDotDot, tokLParen, tokDollar, tokString, tokNumber:
		return true
	}
	return false
}

// [26] RelativePathExpr ::= StepExpr (("/" | "//") StepExpr)*
//
// A leading StepExpr that is a primary expression (FilterExpr) with no following
// "/" is returned as a bare primary; otherwise the result is a pathExpr whose
// steps navigate from the context node (or from a leading primary in path.start).
func (p *parser) parseRelativePathExpr() exprNode {
	path := &pathExpr{}
	if prim, isStep := p.parseStepExpr(path); !isStep {
		if p.cur().kind != tokSlash && p.cur().kind != tokSlashSlash {
			return prim // a primary with no following steps
		}
		path.start = prim
	}
	p.parseTrailingSteps(path)
	return path
}

// parseTrailingSteps appends the ("/" | "//") StepExpr steps that follow the
// first one. A trailing StepExpr that is a primary expression (e.g. "a/(b|c)")
// cannot be modelled as an axis step, so it is parsed for validity but marks the
// expression unsupported.
func (p *parser) parseTrailingSteps(path *pathExpr) {
	for !p.failed() {
		switch p.cur().kind {
		case tokSlashSlash:
			p.advance()
			path.steps = append(path.steps, pathStep{axis: axDescendantOrSelf, test: nodeTest{kind: tnNode}})
		case tokSlash:
			p.advance()
		default:
			return
		}
		if _, isStep := p.parseStepExpr(path); !isStep {
			p.markUnsupported()
		}
	}
}

// [27] StepExpr ::= FilterExpr | AxisStep
//
// When the step is an axis step it is appended to path.steps and isStep is true;
// when it is a primary expression (FilterExpr) the primary is returned and
// isStep is false (the caller decides whether it is a bare primary or the start
// of a path). FilterExpr starts a PrimaryExpr (literal, $var, "(", ".", or a
// function call); AxisStep starts an axis, "@", "..", or a NodeTest.
func (p *parser) parseStepExpr(path *pathExpr) (prim exprNode, isStep bool) {
	if p.failed() {
		return nil, false
	}
	t := p.cur()
	switch t.kind {
	case tokString, tokNumber, tokDollar, tokLParen:
		// PrimaryExpr (FilterExpr).
		return p.parseFilterExpr(), false
	case tokDot:
		// "." is the context item as a primary, but in step position it is the
		// self::node() abbreviation. Treat a following predicate/path as a step.
		p.advance()
		path.steps = append(path.steps, p.finishStep(pathStep{axis: axSelf, test: nodeTest{kind: tnNode}}))
		return nil, true
	case tokDotDot:
		// [34] AbbrevReverseStep
		p.advance()
		path.steps = append(path.steps, p.finishStep(pathStep{axis: axParent, test: nodeTest{kind: tnNode}}))
		return nil, true
	case tokAt:
		// [31] AbbrevForwardStep ::= "@"? NodeTest
		p.advance()
		st := pathStep{axis: axAttribute, test: p.parseNodeTest()}
		path.steps = append(path.steps, p.finishStep(st))
		return nil, true
	case tokStar:
		st := pathStep{axis: axChild, test: p.parseNameTest()}
		path.steps = append(path.steps, p.finishStep(st))
		return nil, true
	case tokName:
		// Axis step?  AxisName "::" ...
		if p.peek(1).kind == tokColonColon && isAxis(t.text) {
			ax, ok := axisNames[t.text]
			if !ok {
				p.markUnsupported() // e.g. the namespace axis
				ax = axChild
			}
			p.advance() // axis name
			p.advance() // "::"
			st := pathStep{axis: ax, test: p.parseNodeTest()}
			path.steps = append(path.steps, p.finishStep(st))
			return nil, true
		}
		// Function call vs kind test vs name test.
		prefix, local, n, _, ok := p.peekQName(p.pos)
		if !ok {
			p.errorf(t.offset, "expected a name test or step")
			return nil, false
		}
		if p.peek(n).kind == tokLParen {
			if prefix == "" && isKindTestKeyword(local) {
				st := pathStep{axis: axChild, test: p.parseNodeTest()} // KindTest
				path.steps = append(path.steps, p.finishStep(st))
				return nil, true
			}
			return p.parseFilterExpr(), false // FunctionCall is a primary
		}
		st := pathStep{axis: axChild, test: p.parseNameTest()}
		path.steps = append(path.steps, p.finishStep(st))
		return nil, true
	}
	p.errorf(t.offset, "expected an expression, found %s", p.describe(t))
	return nil, false
}

// finishStep parses and attaches the predicate list of a step.
func (p *parser) finishStep(st pathStep) pathStep {
	st.preds = p.parsePredicateList()
	return st
}

// parseFilterExpr parses a PrimaryExpr followed by its predicate list, wrapping
// it in a filterExpr when any predicate is present.
func (p *parser) parseFilterExpr() exprNode {
	prim := p.parsePrimaryExpr()
	preds := p.parsePredicateList()
	if len(preds) == 0 {
		return prim
	}
	return &filterExpr{prim, preds}
}

// [39] PredicateList ::= Predicate*
// [40] Predicate ::= "[" Expr "]"
func (p *parser) parsePredicateList() []exprNode {
	var preds []exprNode
	for !p.failed() && p.cur().kind == tokLBracket {
		p.advance()
		preds = append(preds, p.parseExpr())
		p.expect(tokRBracket, "']'")
	}
	return preds
}

// [41] PrimaryExpr ::= Literal | VarRef | ParenthesizedExpr | ContextItemExpr | FunctionCall
func (p *parser) parsePrimaryExpr() exprNode {
	t := p.cur()
	switch t.kind {
	case tokNumber:
		p.advance() // [42] Literal
		f, err := strconv.ParseFloat(t.text, 64)
		if err != nil {
			p.markUnsupported()
		}
		return &litNum{f}
	case tokString:
		p.advance()
		return &litStr{stringValue(t.text)}
	case tokDollar:
		// [44] VarRef ::= "$" VarName
		p.advance()
		return &varRef{p.parseVarName()}
	case tokLParen:
		// [46] ParenthesizedExpr ::= "(" Expr? ")"
		p.advance()
		if p.cur().kind == tokRParen {
			p.advance()
			return &seqExpr{} // the empty sequence
		}
		inner := p.parseExpr()
		p.expect(tokRParen, "')'")
		return inner
	case tokName:
		// [48] FunctionCall
		return p.parseFunctionCall()
	default:
		p.errorf(t.offset, "expected a primary expression, found %s", p.describe(t))
		return nil
	}
}

// [48] FunctionCall ::= QName "(" (ExprSingle ("," ExprSingle)*)? ")"
func (p *parser) parseFunctionCall() exprNode {
	name := p.parseFuncName()
	c := &call{name: name}
	p.expect(tokLParen, "'('")
	if p.cur().kind != tokRParen {
		c.args = append(c.args, p.parseExprSingle())
		for !p.failed() && p.cur().kind == tokComma {
			p.advance()
			c.args = append(c.args, p.parseExprSingle())
		}
	}
	p.expect(tokRParen, "')'")
	return c
}

// [35] NodeTest ::= KindTest | NameTest
func (p *parser) parseNodeTest() nodeTest {
	if p.failed() {
		return nodeTest{}
	}
	t := p.cur()
	if t.kind == tokName && p.peek(1).kind == tokLParen && isKindTestKeyword(t.text) {
		return p.parseKindTest()
	}
	return p.parseNameTest()
}

// [36] NameTest ::= QName | Wildcard
// [37] Wildcard ::= "*" | (NCName ":" "*") | ("*" ":" NCName)
//
// The evaluator matches by local name only, so a prefixed name and a "*:local"
// wildcard both reduce to a local-name test. A bare "*" is the any-element test;
// a "prefix:*" wildcard cannot be expressed as a local-name test, so it falls
// outside the evaluable subset.
func (p *parser) parseNameTest() nodeTest {
	t := p.cur()
	if t.kind == tokStar {
		p.advance()
		if p.cur().kind == tokColon && !p.cur().wsBefore {
			p.advance()
			if p.cur().kind != tokName || p.cur().wsBefore {
				p.errorf(p.cur().offset, "expected a local name after '*:'")
				return nodeTest{}
			}
			return nodeTest{kind: tnName, name: p.advance().text} // "*:local"
		}
		return nodeTest{kind: tnStar}
	}
	if t.kind == tokName {
		first := p.advance().text
		if p.cur().kind == tokColon && !p.cur().wsBefore {
			p.advance()
			switch {
			case p.cur().kind == tokStar && !p.cur().wsBefore:
				p.advance()
				p.markUnsupported() // "prefix:*": any local in a namespace
				return nodeTest{kind: tnStar}
			case p.cur().kind == tokName && !p.cur().wsBefore:
				return nodeTest{kind: tnName, name: p.advance().text} // prefix:local
			default:
				p.errorf(p.cur().offset, "expected a name or '*' after ':'")
				return nodeTest{}
			}
		}
		return nodeTest{kind: tnName, name: first}
	}
	p.errorf(t.offset, "expected a name test, found %s", p.describe(t))
	return nodeTest{}
}

// [49] SingleType ::= AtomicType "?"?   ;  [53] AtomicType ::= QName
func (p *parser) parseSingleType(kind TypeRefKind) string {
	local := p.parseTypeName(kind)
	if !p.failed() && p.cur().kind == tokQuestion {
		p.advance()
	}
	return local
}

// [50] SequenceType ::= ("empty-sequence" "(" ")") | (ItemType OccurrenceIndicator?)
// [51] OccurrenceIndicator ::= "?" | "*" | "+"
//
// It returns the atomic type's local name and whether the item type is a bare
// AtomicType (the only form the evaluator can test); empty-sequence and kind
// tests yield atomic=false.
func (p *parser) parseSequenceType(kind TypeRefKind) (local string, atomic bool) {
	if isName(p.cur(), "empty-sequence") && p.peek(1).kind == tokLParen {
		p.advance() // "empty-sequence"
		p.expect(tokLParen, "'('")
		p.expect(tokRParen, "')'")
		return "", false
	}
	local, atomic = p.parseItemType(kind)
	if p.failed() {
		return local, atomic
	}
	switch p.cur().kind {
	case tokQuestion, tokStar, tokPlus:
		p.advance()
	}
	return local, atomic
}

// [52] ItemType ::= KindTest | ("item" "(" ")") | AtomicType
func (p *parser) parseItemType(kind TypeRefKind) (local string, atomic bool) {
	t := p.cur()
	if t.kind == tokName && p.peek(1).kind == tokLParen {
		switch {
		case t.text == "item":
			p.advance()
			p.expect(tokLParen, "'('")
			p.expect(tokRParen, "')'")
			return "", false
		case isKindTestKeyword(t.text):
			p.parseKindTest()
			return "", false
		}
	}
	// AtomicType ::= QName
	return p.parseTypeName(kind), true
}

// parseTypeName parses a QName naming a type, records it as a TypeRef, and
// returns its local part (the form the evaluator's cast/instance-of uses). It
// rejects a missing or malformed type name (catching e.g. "cast as 3").
func (p *parser) parseTypeName(kind TypeRefKind) string {
	t := p.cur()
	if t.kind != tokName {
		p.errorf(t.offset, "expected a type name, found %s", p.describe(t))
		return ""
	}
	prefix, local, ok := p.scanQName()
	if !ok {
		return ""
	}
	p.expr.TypeRefs = append(p.expr.TypeRefs, TypeRef{
		Prefix: prefix, Local: local, Kind: kind, Offset: t.offset,
	})
	return local
}

// parseVarName parses a variable QName and returns its local part (variables are
// bound and matched by local name).
func (p *parser) parseVarName() string {
	if p.cur().kind != tokName {
		p.errorf(p.cur().offset, "expected a variable name, found %s", p.describe(p.cur()))
		return ""
	}
	_, local, _ := p.scanQName()
	return local
}

// parseFuncName parses a function-name QName and returns it as written
// ("prefix:local" or "local"). The evaluator recognises built-in functions by
// their unprefixed local name and constructor functions by their prefixed form.
func (p *parser) parseFuncName() string {
	prefix, local, ok := p.scanQName()
	if !ok {
		return ""
	}
	if prefix == "" {
		return local
	}
	return prefix + ":" + local
}

// scanQName consumes a "prefix:local" or bare "local" QName from the current
// position. The prefix, colon, and local part must be contiguous (no
// whitespace), per the xml-version grammar note.
func (p *parser) scanQName() (prefix, local string, ok bool) {
	first := p.advance().text
	if p.cur().kind == tokColon && !p.cur().wsBefore {
		p.advance()
		if p.cur().kind != tokName || p.cur().wsBefore {
			p.errorf(p.cur().offset, "expected a local name after ':'")
			return "", "", false
		}
		return first, p.advance().text, true
	}
	return "", first, true
}

// peekQName inspects (without consuming) the QName beginning at token index i,
// returning its prefix/local, the number of tokens it spans, whether it is a
// wildcard, and whether a name was found at all.
func (p *parser) peekQName(i int) (prefix, local string, ntoks int, wildcard, ok bool) {
	tok := func(j int) token {
		if i+j >= len(p.toks) {
			return p.toks[len(p.toks)-1]
		}
		return p.toks[i+j]
	}
	if tok(0).kind != tokName {
		return "", "", 0, false, false
	}
	if tok(1).kind == tokColon && !tok(1).wsBefore {
		if tok(2).kind == tokName && !tok(2).wsBefore {
			return tok(0).text, tok(2).text, 3, false, true
		}
		if tok(2).kind == tokStar && !tok(2).wsBefore {
			return tok(0).text, "*", 3, true, true
		}
	}
	return "", tok(0).text, 1, false, true
}

// expect consumes a token of kind k, or records an error naming want.
func (p *parser) expect(k tokKind, want string) {
	if p.failed() {
		return
	}
	if p.cur().kind != k {
		p.errorf(p.cur().offset, "expected %s, found %s", want, p.describe(p.cur()))
		return
	}
	p.advance()
}

// expectName consumes the keyword name kw, or records an error.
func (p *parser) expectName(kw string) bool {
	if p.failed() {
		return false
	}
	if !isName(p.cur(), kw) {
		p.errorf(p.cur().offset, "expected %q, found %s", kw, p.describe(p.cur()))
		return false
	}
	p.advance()
	return true
}

func isAxis(name string) bool {
	switch name {
	case "child", "descendant", "attribute", "self", "descendant-or-self",
		"following-sibling", "following", "namespace",
		"parent", "ancestor", "preceding-sibling", "preceding", "ancestor-or-self":
		return true
	}
	return false
}

func isKindTestKeyword(name string) bool {
	switch name {
	case "node", "text", "comment", "processing-instruction",
		"document-node", "element", "attribute", "schema-element", "schema-attribute":
		return true
	}
	return false
}
