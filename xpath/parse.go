package xpath

// This file implements the XPath 2.0 grammar (REC-xpath20-20101214 Appendix A.1)
// as a recursive-descent parser. Production numbers from the EBNF are cited in
// comments. The parser validates syntax only; it builds no evaluable tree, but
// records cast/castable/treat/instance-of type references on p.expr.

// [2] Expr ::= ExprSingle ("," ExprSingle)*
func (p *parser) parseExpr() {
	p.parseExprSingle()
	for !p.failed() && p.cur().kind == tokComma {
		p.advance()
		p.parseExprSingle()
	}
}

// [3] ExprSingle ::= ForExpr | QuantifiedExpr | IfExpr | OrExpr
func (p *parser) parseExprSingle() {
	if p.failed() {
		return
	}
	t := p.cur()
	switch {
	case isName(t, "for") && p.peek(1).kind == tokDollar:
		p.parseForExpr()
	case (isName(t, "some") || isName(t, "every")) && p.peek(1).kind == tokDollar:
		p.parseQuantifiedExpr()
	case isName(t, "if") && p.peek(1).kind == tokLParen:
		p.parseIfExpr()
	default:
		p.parseOrExpr()
	}
}

// [4] ForExpr ::= SimpleForClause "return" ExprSingle
// [5] SimpleForClause ::= "for" "$" VarName ("in" ExprSingle ("," "$" VarName "in" ExprSingle)*)
func (p *parser) parseForExpr() {
	p.advance() // "for"
	for {
		p.expect(tokDollar, "'$'")
		p.parseQName("variable name")
		if !p.expectName("in") {
			return
		}
		p.parseExprSingle()
		if p.failed() || p.cur().kind != tokComma {
			break
		}
		p.advance() // ","
	}
	if !p.expectName("return") {
		return
	}
	p.parseExprSingle()
}

// [6] QuantifiedExpr ::= ("some" | "every") "$" VarName "in" ExprSingle
//
//	("," "$" VarName "in" ExprSingle)* "satisfies" ExprSingle
func (p *parser) parseQuantifiedExpr() {
	p.advance() // "some" | "every"
	for {
		p.expect(tokDollar, "'$'")
		p.parseQName("variable name")
		if !p.expectName("in") {
			return
		}
		p.parseExprSingle()
		if p.failed() || p.cur().kind != tokComma {
			break
		}
		p.advance() // ","
	}
	if !p.expectName("satisfies") {
		return
	}
	p.parseExprSingle()
}

// [7] IfExpr ::= "if" "(" Expr ")" "then" ExprSingle "else" ExprSingle
func (p *parser) parseIfExpr() {
	p.advance() // "if"
	p.expect(tokLParen, "'('")
	p.parseExpr()
	p.expect(tokRParen, "')'")
	if !p.expectName("then") {
		return
	}
	p.parseExprSingle()
	if !p.expectName("else") {
		return
	}
	p.parseExprSingle()
}

// [8] OrExpr ::= AndExpr ( "or" AndExpr )*
func (p *parser) parseOrExpr() {
	p.parseAndExpr()
	for !p.failed() && isName(p.cur(), "or") {
		p.advance()
		p.parseAndExpr()
	}
}

// [9] AndExpr ::= ComparisonExpr ( "and" ComparisonExpr )*
func (p *parser) parseAndExpr() {
	p.parseComparisonExpr()
	for !p.failed() && isName(p.cur(), "and") {
		p.advance()
		p.parseComparisonExpr()
	}
}

// [10] ComparisonExpr ::= RangeExpr ( (ValueComp | GeneralComp | NodeComp) RangeExpr )?
// [22] GeneralComp ::= "=" | "!=" | "<" | "<=" | ">" | ">="
// [23] ValueComp   ::= "eq" | "ne" | "lt" | "le" | "gt" | "ge"
// [24] NodeComp    ::= "is" | "<<" | ">>"
func (p *parser) parseComparisonExpr() {
	p.parseRangeExpr()
	if p.failed() {
		return
	}
	t := p.cur()
	if isComparisonOp(t) {
		p.advance()
		p.parseRangeExpr()
	}
}

func isComparisonOp(t token) bool {
	switch t.kind {
	case tokEq, tokNe, tokLt, tokLe, tokGt, tokGe, tokLtLt, tokGtGt:
		return true
	}
	return isName(t, "eq") || isName(t, "ne") || isName(t, "lt") ||
		isName(t, "le") || isName(t, "gt") || isName(t, "ge") || isName(t, "is")
}

// [11] RangeExpr ::= AdditiveExpr ( "to" AdditiveExpr )?
func (p *parser) parseRangeExpr() {
	p.parseAdditiveExpr()
	if !p.failed() && isName(p.cur(), "to") {
		p.advance()
		p.parseAdditiveExpr()
	}
}

// [12] AdditiveExpr ::= MultiplicativeExpr ( ("+" | "-") MultiplicativeExpr )*
func (p *parser) parseAdditiveExpr() {
	p.parseMultiplicativeExpr()
	for !p.failed() && (p.cur().kind == tokPlus || p.cur().kind == tokMinus) {
		p.advance()
		p.parseMultiplicativeExpr()
	}
}

// [13] MultiplicativeExpr ::= UnionExpr ( ("*" | "div" | "idiv" | "mod") UnionExpr )*
func (p *parser) parseMultiplicativeExpr() {
	p.parseUnionExpr()
	for !p.failed() {
		t := p.cur()
		if t.kind == tokStar || isName(t, "div") || isName(t, "idiv") || isName(t, "mod") {
			p.advance()
			p.parseUnionExpr()
			continue
		}
		break
	}
}

// [14] UnionExpr ::= IntersectExceptExpr ( ("union" | "|") IntersectExceptExpr )*
func (p *parser) parseUnionExpr() {
	p.parseIntersectExceptExpr()
	for !p.failed() && (p.cur().kind == tokPipe || isName(p.cur(), "union")) {
		p.advance()
		p.parseIntersectExceptExpr()
	}
}

// [15] IntersectExceptExpr ::= InstanceofExpr ( ("intersect" | "except") InstanceofExpr )*
func (p *parser) parseIntersectExceptExpr() {
	p.parseInstanceofExpr()
	for !p.failed() && (isName(p.cur(), "intersect") || isName(p.cur(), "except")) {
		p.advance()
		p.parseInstanceofExpr()
	}
}

// [16] InstanceofExpr ::= TreatExpr ( "instance" "of" SequenceType )?
func (p *parser) parseInstanceofExpr() {
	p.parseTreatExpr()
	if !p.failed() && isName(p.cur(), "instance") {
		p.advance()
		if !p.expectName("of") {
			return
		}
		p.parseSequenceType(InstanceOf)
	}
}

// [17] TreatExpr ::= CastableExpr ( "treat" "as" SequenceType )?
func (p *parser) parseTreatExpr() {
	p.parseCastableExpr()
	if !p.failed() && isName(p.cur(), "treat") {
		p.advance()
		if !p.expectName("as") {
			return
		}
		p.parseSequenceType(Treat)
	}
}

// [18] CastableExpr ::= CastExpr ( "castable" "as" SingleType )?
func (p *parser) parseCastableExpr() {
	p.parseCastExpr()
	if !p.failed() && isName(p.cur(), "castable") {
		p.advance()
		if !p.expectName("as") {
			return
		}
		p.parseSingleType(Castable)
	}
}

// [19] CastExpr ::= UnaryExpr ( "cast" "as" SingleType )?
func (p *parser) parseCastExpr() {
	p.parseUnaryExpr()
	if !p.failed() && isName(p.cur(), "cast") {
		p.advance()
		if !p.expectName("as") {
			return
		}
		p.parseSingleType(Cast)
	}
}

// [20] UnaryExpr ::= ("-" | "+")* ValueExpr
// [21] ValueExpr ::= PathExpr
func (p *parser) parseUnaryExpr() {
	for p.cur().kind == tokMinus || p.cur().kind == tokPlus {
		p.advance()
	}
	p.parsePathExpr()
}

// [25] PathExpr ::= ("/" RelativePathExpr?) | ("//" RelativePathExpr) | RelativePathExpr
//
//	(xgc: leading-lone-slash)
func (p *parser) parsePathExpr() {
	switch p.cur().kind {
	case tokSlash:
		p.advance()
		if canStartStep(p.cur()) {
			p.parseRelativePathExpr()
		}
		// otherwise a complete path expression "/" (the root).
	case tokSlashSlash:
		p.advance()
		p.parseRelativePathExpr()
	default:
		p.parseRelativePathExpr()
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
func (p *parser) parseRelativePathExpr() {
	p.parseStepExpr()
	for !p.failed() && (p.cur().kind == tokSlash || p.cur().kind == tokSlashSlash) {
		p.advance()
		p.parseStepExpr()
	}
}

// [27] StepExpr ::= FilterExpr | AxisStep
//
//	FilterExpr starts a PrimaryExpr (literal, $var, "(", ".", or function
//	call); AxisStep starts an axis, "@", "..", or a NodeTest. The two
//	overlap only on a leading name, resolved by lookahead below.
func (p *parser) parseStepExpr() {
	if p.failed() {
		return
	}
	t := p.cur()
	switch t.kind {
	case tokString, tokNumber, tokDollar, tokLParen, tokDot:
		// PrimaryExpr (FilterExpr). "." is the context item.
		p.parsePrimaryExpr()
		p.parsePredicateList()
		return
	case tokDotDot:
		// [34] AbbrevReverseStep
		p.advance()
		p.parsePredicateList()
		return
	case tokAt:
		// [31] AbbrevForwardStep ::= "@"? NodeTest
		p.advance()
		p.parseNodeTest()
		p.parsePredicateList()
		return
	case tokStar:
		// Wildcard NameTest (AxisStep).
		p.parseNameTest()
		p.parsePredicateList()
		return
	case tokName:
		// Axis step?  AxisName "::" ...
		if p.peek(1).kind == tokColonColon && isAxis(t.text) {
			p.advance() // axis name
			p.advance() // "::"
			p.parseNodeTest()
			p.parsePredicateList()
			return
		}
		// Function call vs kind test vs name test.
		prefix, local, n, _, ok := p.peekQName(p.pos)
		if !ok {
			p.errorf(t.offset, "expected a name test or step")
			return
		}
		if p.peek(n).kind == tokLParen {
			if prefix == "" && isKindTestKeyword(local) {
				p.parseNodeTest() // KindTest
				p.parsePredicateList()
				return
			}
			p.parseFunctionCall()
			p.parsePredicateList()
			return
		}
		p.parseNameTest()
		p.parsePredicateList()
		return
	}
	p.errorf(t.offset, "expected an expression, found %s", p.describe(t))
}

// [39] PredicateList ::= Predicate*
// [40] Predicate ::= "[" Expr "]"
func (p *parser) parsePredicateList() {
	for !p.failed() && p.cur().kind == tokLBracket {
		p.advance()
		p.parseExpr()
		p.expect(tokRBracket, "']'")
	}
}

// [41] PrimaryExpr ::= Literal | VarRef | ParenthesizedExpr | ContextItemExpr | FunctionCall
func (p *parser) parsePrimaryExpr() {
	t := p.cur()
	switch t.kind {
	case tokNumber, tokString:
		p.advance() // [42] Literal
	case tokDollar:
		// [44] VarRef ::= "$" VarName
		p.advance()
		p.parseQName("variable name")
	case tokDot:
		p.advance() // [47] ContextItemExpr
	case tokLParen:
		// [46] ParenthesizedExpr ::= "(" Expr? ")"
		p.advance()
		if p.cur().kind != tokRParen {
			p.parseExpr()
		}
		p.expect(tokRParen, "')'")
	case tokName:
		// [48] FunctionCall
		p.parseFunctionCall()
	default:
		p.errorf(t.offset, "expected a primary expression, found %s", p.describe(t))
	}
}

// [48] FunctionCall ::= QName "(" (ExprSingle ("," ExprSingle)*)? ")"
func (p *parser) parseFunctionCall() {
	p.parseQName("function name")
	p.expect(tokLParen, "'('")
	if p.cur().kind != tokRParen {
		p.parseExprSingle()
		for !p.failed() && p.cur().kind == tokComma {
			p.advance()
			p.parseExprSingle()
		}
	}
	p.expect(tokRParen, "')'")
}

// [35] NodeTest ::= KindTest | NameTest
func (p *parser) parseNodeTest() {
	if p.failed() {
		return
	}
	t := p.cur()
	if t.kind == tokName && p.peek(1).kind == tokLParen && isKindTestKeyword(t.text) {
		p.parseKindTest()
		return
	}
	p.parseNameTest()
}

// [36] NameTest ::= QName | Wildcard
// [37] Wildcard ::= "*" | (NCName ":" "*") | ("*" ":" NCName)
func (p *parser) parseNameTest() {
	t := p.cur()
	if t.kind == tokStar {
		p.advance()
		if p.cur().kind == tokColon && !p.cur().wsBefore {
			p.advance()
			if p.cur().kind != tokName || p.cur().wsBefore {
				p.errorf(p.cur().offset, "expected a local name after '*:'")
				return
			}
			p.advance()
		}
		return
	}
	if t.kind == tokName {
		p.advance()
		if p.cur().kind == tokColon && !p.cur().wsBefore {
			p.advance()
			switch {
			case p.cur().kind == tokStar && !p.cur().wsBefore:
				p.advance() // NCName ":" "*"
			case p.cur().kind == tokName && !p.cur().wsBefore:
				p.advance() // QName prefix ":" local
			default:
				p.errorf(p.cur().offset, "expected a name or '*' after ':'")
			}
		}
		return
	}
	p.errorf(t.offset, "expected a name test, found %s", p.describe(t))
}

// [49] SingleType ::= AtomicType "?"?   ;  [53] AtomicType ::= QName
func (p *parser) parseSingleType(kind TypeRefKind) {
	p.parseTypeName(kind)
	if !p.failed() && p.cur().kind == tokQuestion {
		p.advance()
	}
}

// [50] SequenceType ::= ("empty-sequence" "(" ")") | (ItemType OccurrenceIndicator?)
// [51] OccurrenceIndicator ::= "?" | "*" | "+"
func (p *parser) parseSequenceType(kind TypeRefKind) {
	if isName(p.cur(), "empty-sequence") && p.peek(1).kind == tokLParen {
		p.advance() // "empty-sequence"
		p.expect(tokLParen, "'('")
		p.expect(tokRParen, "')'")
		return
	}
	p.parseItemType(kind)
	if p.failed() {
		return
	}
	switch p.cur().kind {
	case tokQuestion, tokStar, tokPlus:
		p.advance()
	}
}

// [52] ItemType ::= KindTest | ("item" "(" ")") | AtomicType
func (p *parser) parseItemType(kind TypeRefKind) {
	t := p.cur()
	if t.kind == tokName && p.peek(1).kind == tokLParen {
		switch {
		case t.text == "item":
			p.advance()
			p.expect(tokLParen, "'('")
			p.expect(tokRParen, "')'")
			return
		case isKindTestKeyword(t.text):
			p.parseKindTest()
			return
		}
	}
	// AtomicType ::= QName
	p.parseTypeName(kind)
}

// parseTypeName parses a QName naming a type and records it as a TypeRef. It
// rejects a missing or malformed type name (catching e.g. "cast as 3" and
// "cast as 'string'").
func (p *parser) parseTypeName(kind TypeRefKind) {
	t := p.cur()
	if t.kind != tokName {
		p.errorf(t.offset, "expected a type name, found %s", p.describe(t))
		return
	}
	prefix, local, ok := p.scanQName()
	if !ok {
		return
	}
	p.expr.TypeRefs = append(p.expr.TypeRefs, TypeRef{
		Prefix: prefix, Local: local, Kind: kind, Offset: t.offset,
	})
}

// parseQName parses and consumes a QName, discarding it (used where the name
// has no schema significance, e.g. variable and function names).
func (p *parser) parseQName(what string) {
	if p.cur().kind != tokName {
		p.errorf(p.cur().offset, "expected %s, found %s", what, p.describe(p.cur()))
		return
	}
	p.scanQName()
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
