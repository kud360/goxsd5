package xpath

// KindTest productions (REC-xpath20 Appendix A.1, [54]–[70]). A kind test is a
// node test naming a node kind; the type names it may contain (ElementTest /
// AttributeTest TypeName) are element/attribute type annotations, distinct from
// the atomic cast/instance-of targets, so they are not recorded as TypeRefs.
//
// Of the kind tests only node() and text() map to the evaluator's node-test
// model; the rest are parsed for validity but mark the expression unsupported so
// the evaluator fails open.

// [54] KindTest dispatch. The caller has already established that the current
// token is a kind-test keyword followed by "(". It returns the equivalent
// nodeTest for the evaluator (tnNode for node(), tnText for text()), flagging
// every other kind test as outside the evaluable subset.
func (p *parser) parseKindTest() nodeTest {
	name := p.advance().text // the keyword
	p.expect(tokLParen, "'('")
	if p.failed() {
		return nodeTest{}
	}
	test := nodeTest{kind: tnNode}
	switch name {
	case "node":
		// [55]: empty argument list.
	case "text":
		// [57]: empty argument list.
		test.kind = tnText
	case "comment":
		// [58]: empty argument list.
		p.markUnsupported()
	case "processing-instruction":
		// [59] PITest ::= "processing-instruction" "(" (NCName | StringLiteral)? ")"
		switch p.cur().kind {
		case tokName, tokString:
			p.advance()
		}
		p.markUnsupported()
	case "document-node":
		// [56] DocumentTest ::= "document-node" "(" (ElementTest | SchemaElementTest)? ")"
		if (isName(p.cur(), "element") || isName(p.cur(), "schema-element")) && p.peek(1).kind == tokLParen {
			p.parseKindTest()
		}
		p.markUnsupported()
	case "schema-element", "schema-attribute":
		// [62]/[66]: a required element/attribute declaration (a QName).
		p.discardQName("a declaration name")
		p.markUnsupported()
	case "element":
		p.parseElementTest()
		p.markUnsupported()
	case "attribute":
		p.parseAttributeTest()
		p.markUnsupported()
	}
	p.expect(tokRParen, "')'")
	return test
}

// [64] ElementTest ::= "element" "(" (ElementNameOrWildcard ("," TypeName "?"?)?)? ")"
// [65] ElementNameOrWildcard ::= ElementName | "*"
func (p *parser) parseElementTest() {
	if p.cur().kind == tokRParen {
		return
	}
	p.parseNameOrWildcard()
	if p.cur().kind == tokComma {
		p.advance()
		p.discardQName("a type name")
		if p.cur().kind == tokQuestion {
			p.advance()
		}
	}
}

// [60] AttributeTest ::= "attribute" "(" (AttribNameOrWildcard ("," TypeName)?)? ")"
// [61] AttribNameOrWildcard ::= AttributeName | "*"
func (p *parser) parseAttributeTest() {
	if p.cur().kind == tokRParen {
		return
	}
	p.parseNameOrWildcard()
	if p.cur().kind == tokComma {
		p.advance()
		p.discardQName("a type name")
	}
}

// parseNameOrWildcard parses a QName or a lone "*".
func (p *parser) parseNameOrWildcard() {
	if p.cur().kind == tokStar {
		p.advance()
		return
	}
	p.discardQName("a name or '*'")
}

// discardQName consumes a QName whose value has no schema significance here
// (declaration and type-annotation names inside kind tests).
func (p *parser) discardQName(what string) {
	if p.cur().kind != tokName {
		p.errorf(p.cur().offset, "expected %s, found %s", what, p.describe(p.cur()))
		return
	}
	p.scanQName()
}
