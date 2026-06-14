// Package xpath provides a syntactic parser for XPath 2.0 expressions,
// sufficient to validate the {test} and {select} expressions that appear in
// XSD 1.1 <alternative>, <assert>, and identity-constraint components. It
// implements the full XPath 2.0 grammar (W3C REC-xpath20-20101214, Appendix A)
// and reports static syntax errors. It also surfaces the type-name references
// introduced by the cast/castable/treat/instance-of operators so that callers
// can apply schema-level static checks (for example, that a cast target denotes
// an atomic type and not a complex type).
//
// The package is deliberately decoupled from the xsd model: it neither resolves
// namespace prefixes nor looks up types. Prefix and type resolution belong to
// the caller, which has the schema's namespace context and component registry;
// Parse merely exposes the lexical TypeRefs needed for those checks.
package xpath

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

// tokKind enumerates the XPath 2.0 terminal symbols the lexer distinguishes.
// Operator keywords (and, or, div, eq, …) are NOT pre-classified: they are
// emitted as tokName and recognised by the parser according to grammar
// position, which is how the XPath 2.0 operator/name-test ambiguity is resolved.
type tokKind int

const (
	tokEOF        tokKind = iota
	tokName               // NCName (a QName is two tokName separated by tokColon)
	tokNumber             // IntegerLiteral | DecimalLiteral | DoubleLiteral
	tokString             // StringLiteral
	tokDollar             // $
	tokLParen             // (
	tokRParen             // )
	tokLBracket           // [
	tokRBracket           // ]
	tokAt                 // @
	tokComma              // ,
	tokColon              // :
	tokColonColon         // ::
	tokSlash              // /
	tokSlashSlash         // //
	tokDot                // .
	tokDotDot             // ..
	tokStar               // *
	tokPlus               // +
	tokMinus              // -
	tokQuestion           // ?
	tokPipe               // |
	tokEq                 // =
	tokNe                 // !=
	tokLt                 // <
	tokLe                 // <=
	tokLtLt               // <<
	tokGt                 // >
	tokGe                 // >=
	tokGtGt               // >>
)

// token is one lexical token. text carries the literal source for names,
// numbers, and string literals. wsBefore records whether whitespace or a
// comment preceded the token; the parser uses it to forbid whitespace inside a
// QName (the prefix ':' local-part must be contiguous, per the xml-version
// grammar note).
type token struct {
	kind     tokKind
	text     string
	offset   int
	wsBefore bool
}

// lexer turns an XPath 2.0 expression into a slice of tokens. Tokenising the
// whole input up front lets the recursive-descent parser look ahead freely,
// which the grammar requires (function-call vs kind-test vs name-test, the
// leading-lone-slash rule, and for/some/every disambiguation).
type lexer struct {
	src  string
	pos  int
	toks []token
	err  *Error
}

// lex tokenises src. On a lexical error it returns the tokens scanned so far
// and a non-nil *Error.
func lex(src string) ([]token, *Error) {
	l := &lexer{src: src}
	l.run()
	if l.err != nil {
		return l.toks, l.err
	}
	return l.toks, nil
}

func (l *lexer) errorf(off int, format string, args ...any) {
	if l.err == nil {
		l.err = &Error{Offset: off, Msg: fmt.Sprintf(format, args...)}
	}
}

func (l *lexer) run() {
	for l.err == nil {
		ws := l.skipSeparators()
		if l.pos >= len(l.src) {
			l.toks = append(l.toks, token{kind: tokEOF, offset: l.pos, wsBefore: ws})
			return
		}
		start := l.pos
		t := l.scanToken()
		if l.err != nil {
			return
		}
		t.offset = start
		t.wsBefore = ws
		l.toks = append(l.toks, t)
	}
}

// skipSeparators consumes ignorable whitespace and (: nested comments :),
// reporting whether anything was skipped.
func (l *lexer) skipSeparators() bool {
	skipped := false
	for l.pos < len(l.src) {
		c := l.src[l.pos]
		switch {
		case c == ' ' || c == '\t' || c == '\r' || c == '\n':
			l.pos++
			skipped = true
		case c == '(' && l.pos+1 < len(l.src) && l.src[l.pos+1] == ':':
			l.skipComment()
			skipped = true
		default:
			return skipped
		}
	}
	return skipped
}

// skipComment consumes a "(:" comment, which may nest, up to its matching ":)".
func (l *lexer) skipComment() {
	start := l.pos
	l.pos += 2 // consume "(:"
	depth := 1
	for l.pos < len(l.src) {
		if l.src[l.pos] == '(' && l.pos+1 < len(l.src) && l.src[l.pos+1] == ':' {
			depth++
			l.pos += 2
			continue
		}
		if l.src[l.pos] == ':' && l.pos+1 < len(l.src) && l.src[l.pos+1] == ')' {
			depth--
			l.pos += 2
			if depth == 0 {
				return
			}
			continue
		}
		l.pos++
	}
	l.errorf(start, "unterminated comment")
}

func (l *lexer) scanToken() token {
	c := l.src[l.pos]
	switch c {
	case '$':
		l.pos++
		return token{kind: tokDollar}
	case '(':
		l.pos++
		return token{kind: tokLParen}
	case ')':
		l.pos++
		return token{kind: tokRParen}
	case '[':
		l.pos++
		return token{kind: tokLBracket}
	case ']':
		l.pos++
		return token{kind: tokRBracket}
	case '@':
		l.pos++
		return token{kind: tokAt}
	case ',':
		l.pos++
		return token{kind: tokComma}
	case '|':
		l.pos++
		return token{kind: tokPipe}
	case '+':
		l.pos++
		return token{kind: tokPlus}
	case '-':
		l.pos++
		return token{kind: tokMinus}
	case '*':
		l.pos++
		return token{kind: tokStar}
	case '?':
		l.pos++
		return token{kind: tokQuestion}
	case '=':
		l.pos++
		return token{kind: tokEq}
	case ':':
		l.pos++
		if l.pos < len(l.src) && l.src[l.pos] == ':' {
			l.pos++
			return token{kind: tokColonColon}
		}
		return token{kind: tokColon}
	case '/':
		l.pos++
		if l.pos < len(l.src) && l.src[l.pos] == '/' {
			l.pos++
			return token{kind: tokSlashSlash}
		}
		return token{kind: tokSlash}
	case '!':
		l.pos++
		if l.pos < len(l.src) && l.src[l.pos] == '=' {
			l.pos++
			return token{kind: tokNe}
		}
		l.errorf(l.pos-1, "expected '=' after '!'")
		return token{}
	case '<':
		l.pos++
		if l.pos < len(l.src) {
			switch l.src[l.pos] {
			case '=':
				l.pos++
				return token{kind: tokLe}
			case '<':
				l.pos++
				return token{kind: tokLtLt}
			}
		}
		return token{kind: tokLt}
	case '>':
		l.pos++
		if l.pos < len(l.src) {
			switch l.src[l.pos] {
			case '=':
				l.pos++
				return token{kind: tokGe}
			case '>':
				l.pos++
				return token{kind: tokGtGt}
			}
		}
		return token{kind: tokGt}
	case '"', '\'':
		return l.scanString()
	case '.':
		// "." may begin a decimal/double literal (". Digits"), the ".."
		// step, or the context-item ".".
		if l.pos+1 < len(l.src) && isDigit(l.src[l.pos+1]) {
			return l.scanNumber()
		}
		l.pos++
		if l.pos < len(l.src) && l.src[l.pos] == '.' {
			l.pos++
			return token{kind: tokDotDot}
		}
		return token{kind: tokDot}
	}
	switch {
	case isDigit(c):
		return l.scanNumber()
	case isNameStart(rune(c)) || c >= utf8.RuneSelf:
		return l.scanName()
	}
	l.errorf(l.pos, "unexpected character %q", string(rune(c)))
	return token{}
}

// scanString scans a string literal, honouring the doubled-quote escapes
// (” inside '…' and "" inside "…").
func (l *lexer) scanString() token {
	quote := l.src[l.pos]
	start := l.pos
	l.pos++
	for l.pos < len(l.src) {
		if l.src[l.pos] == quote {
			if l.pos+1 < len(l.src) && l.src[l.pos+1] == quote {
				l.pos += 2 // escaped quote
				continue
			}
			l.pos++
			return token{kind: tokString, text: l.src[start:l.pos]}
		}
		l.pos++
	}
	l.errorf(start, "unterminated string literal")
	return token{}
}

// scanNumber scans an IntegerLiteral, DecimalLiteral, or DoubleLiteral.
func (l *lexer) scanNumber() token {
	start := l.pos
	for l.pos < len(l.src) && isDigit(l.src[l.pos]) {
		l.pos++
	}
	if l.pos < len(l.src) && l.src[l.pos] == '.' {
		l.pos++
		for l.pos < len(l.src) && isDigit(l.src[l.pos]) {
			l.pos++
		}
	}
	if l.pos < len(l.src) && (l.src[l.pos] == 'e' || l.src[l.pos] == 'E') {
		save := l.pos
		l.pos++
		if l.pos < len(l.src) && (l.src[l.pos] == '+' || l.src[l.pos] == '-') {
			l.pos++
		}
		if l.pos < len(l.src) && isDigit(l.src[l.pos]) {
			for l.pos < len(l.src) && isDigit(l.src[l.pos]) {
				l.pos++
			}
		} else {
			// "e" not followed by an exponent: not part of the number.
			l.pos = save
		}
	}
	return token{kind: tokNumber, text: l.src[start:l.pos]}
}

// scanName scans an NCName. Per XPath 2.0 lexical rules a name greedily
// consumes '-' and '.' (both are XML NameChars), so "foo-bar" and "a.b" each
// lex as a single name.
func (l *lexer) scanName() token {
	start := l.pos
	// First rune is already known to be a name-start (or non-ASCII) rune.
	_, sz := utf8.DecodeRuneInString(l.src[l.pos:])
	l.pos += sz
	for l.pos < len(l.src) {
		r, sz := utf8.DecodeRuneInString(l.src[l.pos:])
		if !isNameChar(r) {
			break
		}
		l.pos += sz
	}
	return token{kind: tokName, text: l.src[start:l.pos]}
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

// isNameStart and isNameChar approximate the XML NCName productions. They admit
// every ASCII letter and underscore as a start character, plus any non-ASCII
// letter, and additionally '-', '.', and digits as continuation characters.
// This is intentionally permissive: the goal is to recognise valid names
// without rejecting any, never to validate name characters precisely.
func isNameStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isNameChar(r rune) bool {
	return r == '_' || r == '-' || r == '.' || unicode.IsLetter(r) || unicode.IsDigit(r) ||
		r == 0xB7 || unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r)
}
