package xpath

import "fmt"

// TypeRefKind identifies the operator that introduced a type-name reference.
type TypeRefKind int

const (
	// Cast is the target type of a "cast as" expression (a SingleType).
	Cast TypeRefKind = iota
	// Castable is the target type of a "castable as" expression (a SingleType).
	Castable
	// Treat is the target item type of a "treat as" expression, when that
	// item type is an AtomicType (a bare QName, not a kind test).
	Treat
	// InstanceOf is the target item type of an "instance of" expression, when
	// that item type is an AtomicType (a bare QName, not a kind test).
	InstanceOf
)

func (k TypeRefKind) String() string {
	switch k {
	case Cast:
		return "cast as"
	case Castable:
		return "castable as"
	case Treat:
		return "treat as"
	case InstanceOf:
		return "instance of"
	}
	return "unknown"
}

// TypeRef is a type name named by a cast/castable/treat/instance-of operator,
// as written in the source. The prefix is not resolved to a namespace URI;
// callers resolve it against their own namespace context.
type TypeRef struct {
	Prefix string // namespace prefix as written; "" if unprefixed
	Local  string // local part
	Kind   TypeRefKind
	Offset int // byte offset of the type name within the source expression
}

// Expr is a successfully parsed XPath 2.0 expression. It records the type-name
// references found while parsing so callers can apply schema-level checks; the
// full evaluable tree is not retained, since validation only needs syntax plus
// these references.
type Expr struct {
	Src      string
	TypeRefs []TypeRef
}

// Error is a static (lexical or syntactic) error in an XPath expression. The
// XSD 1.1 "Type Alternative Representation OK" (src-ta) and assertion
// constraints make any such static error a schema error.
type Error struct {
	Offset int // byte offset within the expression
	Msg    string
}

func (e *Error) Error() string {
	return fmt.Sprintf("xpath: %s (at offset %d)", e.Msg, e.Offset)
}

// Parse parses src as an XPath 2.0 expression. It returns a non-nil *Error if
// src contains a static syntax error (an unparseable expression, per the
// XPath 2.0 grammar). A nil error means src is well-formed XPath 2.0; any
// remaining schema-level checks are driven by the returned Expr.TypeRefs.
func Parse(src string) (*Expr, error) {
	toks, lerr := lex(src)
	if lerr != nil {
		return nil, lerr
	}
	p := &parser{src: src, toks: toks, expr: &Expr{Src: src}}
	p.parseExpr()
	if p.err != nil {
		return nil, p.err
	}
	if p.cur().kind != tokEOF {
		p.errorf(p.cur().offset, "unexpected %s", p.describe(p.cur()))
		return nil, p.err
	}
	return p.expr, nil
}

// parser is a recursive-descent parser over the pre-tokenised expression.
type parser struct {
	src  string
	toks []token
	pos  int
	expr *Expr
	err  *Error
}

func (p *parser) cur() token { return p.toks[p.pos] }
func (p *parser) peek(n int) token {
	i := p.pos + n
	if i >= len(p.toks) {
		return p.toks[len(p.toks)-1] // tokEOF
	}
	return p.toks[i]
}
func (p *parser) advance() token {
	t := p.toks[p.pos]
	if t.kind != tokEOF {
		p.pos++
	}
	return t
}

func (p *parser) errorf(off int, format string, args ...any) {
	if p.err == nil {
		p.err = &Error{Offset: off, Msg: fmt.Sprintf(format, args...)}
	}
}

func (p *parser) failed() bool { return p.err != nil }

// isName reports whether t is the keyword name kw (case-sensitive, as the
// XPath 2.0 grammar requires — "and" is the operator, "AND" is a name test).
func isName(t token, kw string) bool { return t.kind == tokName && t.text == kw }

func (p *parser) describe(t token) string {
	switch t.kind {
	case tokEOF:
		return "end of expression"
	case tokName:
		return fmt.Sprintf("name %q", t.text)
	case tokNumber:
		return fmt.Sprintf("number %q", t.text)
	case tokString:
		return "string literal"
	default:
		return fmt.Sprintf("token %q", tokText(t.kind))
	}
}

func tokText(k tokKind) string {
	switch k {
	case tokDollar:
		return "$"
	case tokLParen:
		return "("
	case tokRParen:
		return ")"
	case tokLBracket:
		return "["
	case tokRBracket:
		return "]"
	case tokAt:
		return "@"
	case tokComma:
		return ","
	case tokColon:
		return ":"
	case tokColonColon:
		return "::"
	case tokSlash:
		return "/"
	case tokSlashSlash:
		return "//"
	case tokDot:
		return "."
	case tokDotDot:
		return ".."
	case tokStar:
		return "*"
	case tokPlus:
		return "+"
	case tokMinus:
		return "-"
	case tokQuestion:
		return "?"
	case tokPipe:
		return "|"
	case tokEq:
		return "="
	case tokNe:
		return "!="
	case tokLt:
		return "<"
	case tokLe:
		return "<="
	case tokLtLt:
		return "<<"
	case tokGt:
		return ">"
	case tokGe:
		return ">="
	case tokGtGt:
		return ">>"
	}
	return "?"
}
