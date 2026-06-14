package xpath

// KindTest productions (REC-xpath20 Appendix A.1, [54]–[70]). A kind test is a
// node test naming a node kind; the type names it may contain (ElementTest /
// AttributeTest TypeName) are element/attribute type annotations, distinct from
// the atomic cast/instance-of targets, so they are not recorded as TypeRefs.

// [54] KindTest dispatch. The caller has already established that the current
// token is a kind-test keyword followed by "(".
func (p *parser) parseKindTest() {
	name := p.advance().text // the keyword
	p.expect(tokLParen, "'('")
	if p.failed() {
		return
	}
	switch name {
	case "node", "text", "comment":
		// [55]/[57]/[58]: empty argument list.
	case "processing-instruction":
		// [59] PITest ::= "processing-instruction" "(" (NCName | StringLiteral)? ")"
		switch p.cur().kind {
		case tokName, tokString:
			p.advance()
		}
	case "document-node":
		// [56] DocumentTest ::= "document-node" "(" (ElementTest | SchemaElementTest)? ")"
		if isName(p.cur(), "element") || isName(p.cur(), "schema-element") {
			if p.cur().kind == tokName && p.peek(1).kind == tokLParen {
				p.parseKindTest()
			}
		}
	case "schema-element", "schema-attribute":
		// [62]/[66]: a required element/attribute declaration (a QName).
		p.parseQName("a declaration name")
	case "element":
		p.parseElementTest()
	case "attribute":
		p.parseAttributeTest()
	}
	p.expect(tokRParen, "')'")
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
		p.parseQName("a type name")
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
		p.parseQName("a type name")
	}
}

// parseNameOrWildcard parses a QName or a lone "*".
func (p *parser) parseNameOrWildcard() {
	if p.cur().kind == tokStar {
		p.advance()
		return
	}
	p.parseQName("a name or '*'")
}
