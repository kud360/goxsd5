package xsdregex

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// XSD regular expressions (Part 2 Appendix G) translated to Go's RE2
// syntax. The translator fully parses the XSD grammar: implicit anchoring,
// `.` = [^\n\r], class subtraction [a-z-[m]], \i \I \c \C \s \S \d \D \w \W,
// category escapes \p{L}, block escapes \p{IsBasicLatin}, and `^`/`$` as
// ordinary characters. Character classes are computed as explicit code
// point sets, so the emitted RE2 contains no \p escapes at all.

// CompileRegex translates an XSD pattern and compiles it.
func CompileRegex(pattern string) (*regexp.Regexp, error) {
	s, err := TranslateRegex(pattern)
	if err != nil {
		return nil, err
	}
	re, err := regexp.Compile(s)
	if err != nil {
		// Translation is supposed to emit only valid RE2; surface the XSD
		// source alongside the translated form if it ever does not.
		return nil, fmt.Errorf("XSD regex %q: translated form %q does not compile: %w", pattern, s, err)
	}
	return re, nil
}

// TranslateRegex translates an XSD pattern into Go regexp syntax, wrapped
// in \A(?:…)\z for the implicit anchoring of pattern facets.
func TranslateRegex(pattern string) (string, error) {
	s, err := translate(pattern)
	if err != nil {
		return "", err
	}
	return `\A(?:` + s + `)\z`, nil
}

// translate parses an XSD pattern into the equivalent RE2 source, without the
// anchoring wrapper. Pattern facets anchor the whole value (TranslateRegex);
// fn:matches is an unanchored substring test (Matches), so both share this core.
func translate(pattern string) (string, error) {
	p := &reParser{src: []rune(pattern)}
	s, err := p.regExp()
	if err != nil {
		return "", fmt.Errorf("invalid XSD regex %q at offset %d: %w", pattern, p.pos, err)
	}
	if p.pos != len(p.src) {
		return "", fmt.Errorf("invalid XSD regex %q: unexpected %q at offset %d", pattern, p.src[p.pos], p.pos)
	}
	return s, nil
}

// Matches implements fn:matches: it reports whether input contains a substring
// matching the XSD-syntax pattern. Unlike a pattern facet the match is NOT
// implicitly anchored, so the translated form is compiled without \A…\z.
//
// flags carries the fn:matches flag string. Only "i" (case-insensitive) is
// honoured: in XSD regex syntax "." already translates to an explicit
// newline-excluding set and "^"/"$" are ordinary characters, so the "s" and
// "m" flags would have no effect on the translated form — accepting them would
// silently give wrong answers. They, "x" (free-spacing, which RE2 cannot
// express), and any other flag character are therefore rejected with an error,
// which a caller surfaces as a dynamic error rather than ignoring the flag.
func Matches(pattern, flags, input string) (bool, error) {
	inline, err := inlineFlags(flags)
	if err != nil {
		return false, err
	}
	core, err := translate(pattern)
	if err != nil {
		return false, err
	}
	re, err := regexp.Compile(inline + `(?:` + core + `)`)
	if err != nil {
		return false, fmt.Errorf("XSD regex %q: translated form does not compile: %w", pattern, err)
	}
	return re.MatchString(input), nil
}

// inlineFlags translates an fn:matches flag string into an RE2 inline-flag
// prefix like "(?i)". Only "i" is supportable against XSD-translated regexes;
// any other flag is an error (see Matches).
func inlineFlags(flags string) (string, error) {
	for _, c := range flags {
		if c != 'i' {
			return "", fmt.Errorf("xsdregex: unsupported fn:matches flag %q", c)
		}
	}
	if flags == "" {
		return "", nil
	}
	return "(?i)", nil
}

type reParser struct {
	src []rune
	pos int

	// fo selects the XPath/XQuery Functions & Operators regex flavor used by
	// fn:matches/replace/tokenize (see fo.go). In the default (false) XSD
	// Appendix-G flavor `^`/`$` are ordinary characters and reluctant
	// quantifiers and back-references do not exist; fo mode enables them.
	fo bool
	// dotAll reflects the F&O `s` flag: `.` matches every character including
	// newlines. It is only consulted when fo is true.
	dotAll bool
}

func (p *reParser) eof() bool { return p.pos >= len(p.src) }

func (p *reParser) peek() rune {
	if p.eof() {
		return -1
	}
	return p.src[p.pos]
}

func (p *reParser) peekAt(n int) rune {
	if p.pos+n >= len(p.src) {
		return -1
	}
	return p.src[p.pos+n]
}

func (p *reParser) next() rune {
	r := p.peek()
	p.pos++
	return r
}

// regExp := branch ('|' branch)*
func (p *reParser) regExp() (string, error) {
	var parts []string
	for {
		b, err := p.branch()
		if err != nil {
			return "", err
		}
		parts = append(parts, b)
		if p.peek() != '|' {
			break
		}
		p.pos++
	}
	return strings.Join(parts, "|"), nil
}

// branch := piece*
func (p *reParser) branch() (string, error) {
	var b strings.Builder
	for !p.eof() && p.peek() != '|' && p.peek() != ')' {
		s, err := p.piece()
		if err != nil {
			return "", err
		}
		b.WriteString(s)
	}
	return b.String(), nil
}

// piece := atom quantifier?
func (p *reParser) piece() (string, error) {
	atom, err := p.atom()
	if err != nil {
		return "", err
	}
	switch p.peek() {
	case '?', '*', '+':
		return atom + string(p.next()) + p.reluctant(), nil
	case '{':
		p.pos++
		min, max, err := p.quantity()
		if err != nil {
			return "", err
		}
		return applyQuant(atom, min, max) + p.reluctant(), nil
	}
	return atom, nil
}

// reluctant consumes a trailing '?' reluctant-quantifier marker in F&O mode
// (e.g. `a+?`, `a{2,3}?`) and returns it for RE2, which supports non-greedy
// quantifiers natively. Outside F&O mode the XSD grammar has no reluctant
// quantifiers, so nothing is consumed.
func (p *reParser) reluctant() string {
	if !p.fo || p.peek() != '?' {
		return ""
	}
	p.pos++
	return "?"
}

// backReference reports whether the position just past a consumed backslash
// begins an F&O back-reference (\1..\9). RE2 cannot express back-references, so
// the caller rejects them rather than mistranslating. The digit is not
// consumed; it is returned for the error message.
func (p *reParser) backReference() (rune, bool) {
	c := p.peek()
	if c < '1' || c > '9' {
		return 0, false
	}
	return c, true
}

// quantity := digits (',' digits?)? '}'
func (p *reParser) quantity() (min, max int, err error) {
	min, ok := p.digits()
	if !ok {
		return 0, 0, fmt.Errorf("malformed {n,m} quantifier")
	}
	max = min
	if p.peek() == ',' {
		p.pos++
		if p.peek() == '}' {
			max = -1
		} else if max, ok = p.digits(); !ok {
			return 0, 0, fmt.Errorf("malformed {n,m} quantifier")
		}
	}
	if p.next() != '}' {
		return 0, 0, fmt.Errorf("unterminated {n,m} quantifier")
	}
	if max != -1 && max < min {
		return 0, 0, fmt.Errorf("quantifier {%d,%d} has max < min", min, max)
	}
	return min, max, nil
}

func (p *reParser) digits() (int, bool) {
	start := p.pos
	v := 0
	for !p.eof() && p.peek() >= '0' && p.peek() <= '9' {
		v = v*10 + int(p.next()-'0')
		if v > 1<<29 {
			return 0, false
		}
	}
	return v, p.pos > start
}

// goRepeatMax is RE2's maximum counted-repetition bound; larger XSD counts
// are decomposed (exactly for fixed/lower bounds, approximately for very
// wide ranges).
const goRepeatMax = 1000

func applyQuant(atom string, min, max int) string {
	exact := func(n int) string {
		if n == 0 {
			return ""
		}
		if n == 1 {
			return atom
		}
		if n <= goRepeatMax {
			return fmt.Sprintf("%s{%d}", atom, n)
		}
		q, r := n/goRepeatMax, n%goRepeatMax
		s := fmt.Sprintf("(?:%s{%d}){%d}", atom, goRepeatMax, q)
		if r > 0 {
			s += fmt.Sprintf("%s{%d}", atom, r)
		}
		return s
	}
	switch {
	case max == -1:
		switch min {
		case 0:
			return atom + "*"
		case 1:
			return atom + "+"
		default:
			return exact(min) + atom + "*"
		}
	case min == max:
		if min == 0 {
			return "(?:)" // {0,0}: the empty string
		}
		return exact(min)
	default:
		opt := max - min
		var tail string
		if opt <= goRepeatMax {
			tail = fmt.Sprintf("%s{0,%d}", atom, opt)
		} else {
			// Approximate: allows up to the next multiple of the chunk
			// size. Wide ranges beyond RE2's limits are pathological.
			chunks := (opt + goRepeatMax - 1) / goRepeatMax
			tail = fmt.Sprintf("(?:%s{0,%d}){%d}", atom, goRepeatMax, chunks)
		}
		return exact(min) + tail
	}
}

// atom := NormalChar | charClass | ( regExp )
func (p *reParser) atom() (string, error) {
	switch c := p.peek(); c {
	case '(':
		p.pos++
		s, err := p.regExp()
		if err != nil {
			return "", err
		}
		if p.next() != ')' {
			return "", fmt.Errorf("unterminated group")
		}
		if p.fo {
			// F&O groups capture (fn:replace references them as $1, $2, …);
			// the XSD flavor has no back-references, so it uses non-capturing
			// groups to keep the emitted RE2 minimal.
			return "(" + s + ")", nil
		}
		return "(?:" + s + ")", nil
	case '[':
		set, err := p.charClassExpr()
		if err != nil {
			return "", err
		}
		return emitSet(set), nil
	case '.':
		p.pos++
		if p.fo {
			// F&O/XQuery `.` excludes only #x0A (\n); with the `s`
			// (dot-all) flag it matches every character. Unlike the XSD
			// Part 2 Appendix-G pattern-facet `.` (dotSet, which also
			// excludes \r), F&O `.` matches a carriage return.
			if p.dotAll {
				return emitSet(complement(nil)), nil
			}
			return emitSet(foDotSet()), nil
		}
		return emitSet(dotSet()), nil
	case '^':
		if p.fo {
			p.pos++
			return `\A`, nil
		}
		p.pos++
		return escLiteral(c), nil
	case '$':
		if p.fo {
			p.pos++
			return `\z`, nil
		}
		p.pos++
		return escLiteral(c), nil
	case '\\':
		p.pos++
		if p.fo {
			if r, ok := p.backReference(); ok {
				return "", fmt.Errorf(`back-reference \%c is not supported`, r)
			}
		}
		if lit, ok, err := p.singleCharEsc(); err != nil {
			return "", err
		} else if ok {
			return escLiteral(lit), nil
		}
		set, err := p.classEsc()
		if err != nil {
			return "", err
		}
		return emitSet(set), nil
	case '?', '*', '+', '{', '}', ')', '|', ']', -1:
		return "", fmt.Errorf("unescaped metacharacter or missing atom")
	default:
		p.pos++
		return escLiteral(c), nil
	}
}

// singleCharEsc handles \n \r \t and the escaped metacharacters; the
// backslash has already been consumed. Returns ok=false (without consuming)
// if the escape is a class escape instead.
func (p *reParser) singleCharEsc() (rune, bool, error) {
	switch c := p.peek(); c {
	case 'n':
		p.pos++
		return '\n', true, nil
	case 'r':
		p.pos++
		return '\r', true, nil
	case 't':
		p.pos++
		return '\t', true, nil
	case '\\', '|', '.', '?', '*', '+', '(', ')', '{', '}', '-', '[', ']', '^':
		p.pos++
		return c, true, nil
	case 's', 'S', 'i', 'I', 'c', 'C', 'd', 'D', 'w', 'W', 'p', 'P':
		return 0, false, nil
	case -1:
		return 0, false, fmt.Errorf("trailing backslash")
	default:
		return 0, false, fmt.Errorf(`invalid escape \%c`, c)
	}
}

// classEsc handles \s \S \i \I \c \C \d \D \w \W \p{…} \P{…}; the
// backslash has already been consumed.
func (p *reParser) classEsc() (rangeSet, error) {
	switch c := p.next(); c {
	case 's':
		return xmlSpaceSet(), nil
	case 'S':
		return complement(xmlSpaceSet()), nil
	case 'i':
		return nameStartSet(), nil
	case 'I':
		return complement(nameStartSet()), nil
	case 'c':
		return nameCharSet(), nil
	case 'C':
		return complement(nameCharSet()), nil
	case 'd':
		return categorySet("Nd")
	case 'D':
		s, err := categorySet("Nd")
		if err != nil {
			return nil, err
		}
		return complement(s), nil
	case 'w':
		return wordSet()
	case 'W':
		s, err := wordSet()
		if err != nil {
			return nil, err
		}
		return complement(s), nil
	case 'p', 'P':
		if p.next() != '{' {
			return nil, fmt.Errorf(`\%c must be followed by {…}`, c)
		}
		start := p.pos
		for !p.eof() && p.peek() != '}' {
			p.pos++
		}
		if p.eof() {
			return nil, fmt.Errorf(`unterminated \%c{…}`, c)
		}
		name := string(p.src[start:p.pos])
		p.pos++ // '}'
		set, err := propertySet(name)
		if err != nil {
			return nil, err
		}
		if c == 'P' {
			return complement(set), nil
		}
		return set, nil
	default:
		return nil, fmt.Errorf(`invalid escape \%c`, c)
	}
}

// charClassExpr := '[' charGroup ']'   (with optional ^ negation and
// trailing -[…] subtraction)
func (p *reParser) charClassExpr() (rangeSet, error) {
	if p.next() != '[' {
		return nil, fmt.Errorf("expected [")
	}
	neg := false
	if p.peek() == '^' {
		neg = true
		p.pos++
	}
	var set rangeSet
	var sub rangeSet
	first := true
	for {
		done, added, newSet, newSub, err := p.charGroupStep(set, first)
		if err != nil {
			return nil, err
		}
		if done {
			break
		}
		if added {
			first = false
		}
		if newSub != nil {
			sub = newSub
		} else {
			set = newSet
		}
	}
	if neg {
		set = complement(set)
	}
	if sub != nil {
		set = subtract(set, sub)
	}
	return set, nil
}

// charGroupStep processes one iteration of the character group loop.
// done=true means the closing ']' was consumed; added=true means a real item
// was consumed and first should become false in the caller. Subtraction (-[…])
// does NOT set added because the left-hand charGroup must be non-empty first.
// newSub is non-nil only when a -[…] subtraction was parsed.
func (p *reParser) charGroupStep(set rangeSet, first bool) (done, added bool, newSet, newSub rangeSet, err error) {
	switch c := p.peek(); c {
	case -1:
		return false, false, nil, nil, fmt.Errorf("unterminated character class")
	case ']':
		if first {
			return false, false, nil, nil, fmt.Errorf("empty character class")
		}
		p.pos++
		return true, false, nil, nil, nil
	case '[':
		return false, false, nil, nil, fmt.Errorf("unescaped '[' in character class")
	case '-':
		newSet, newSub, err = p.charGroupDash(set)
		// Subtraction is only valid after at least one item (first==false).
		// Do not set added so first remains true when subtraction is the
		// very first token — the subsequent ']' will then report "empty class".
		isSub := newSub != nil
		return false, err == nil && !isSub, newSet, newSub, err
	default:
		newSet, err = p.charGroupDefault(set)
		return false, err == nil, newSet, nil, err
	}
}

// charGroupDash handles a '-' character at the current position inside a
// character class group. It distinguishes subtraction (-[…]), a forbidden
// double-dash lower bound, and a literal dash.
func (p *reParser) charGroupDash(set rangeSet) (newSet, newSub rangeSet, err error) {
	switch {
	case p.peekAt(1) == '[':
		// Subtraction: must be the last item of the group.
		p.pos++
		s, err := p.charClassExpr()
		if err != nil {
			return nil, nil, err
		}
		if p.peek() != ']' {
			return nil, nil, fmt.Errorf("subtraction must end the character class")
		}
		return set, s, nil
	case p.peekAt(1) == '-' && p.peekAt(2) != ']' && p.peekAt(2) != '[':
		// '-' '-' X would be a charRange whose lower bound is an
		// unescaped '-', which the spec forbids, e.g. [--z].
		return nil, nil, fmt.Errorf("'-' cannot be the lower bound of a range")
	default:
		// A '-' at the start of a group item is a literal: it is
		// leading, trailing ([-]), or follows a completed range or
		// other item (e.g. [a-z-+] = a–z, '-', '+'). A '-' can never
		// begin a range, so it is always literal here.
		p.pos++
		return addRange(set, '-', '-'), nil, nil
	}
}

// charGroupDefault handles the default case inside a character class group:
// a character or escape sequence, possibly followed by a '-' range operator.
func (p *reParser) charGroupDefault(set rangeSet) (rangeSet, error) {
	lo, isChar, cset, err := p.classChar()
	if err != nil {
		return nil, err
	}
	if !isChar {
		return union(set, cset), nil
	}
	// Possible range: x-y, unless the '-' starts a subtraction
	// or is the trailing literal '-'.
	if p.peek() != '-' || p.peekAt(1) == '[' || p.peekAt(1) == ']' {
		return addRange(set, lo, lo), nil
	}
	p.pos++
	if p.peek() == '-' {
		// A literal (unescaped) '-' may not be the upper bound of
		// a range, e.g. [!--] is invalid (only \- could end one).
		return nil, fmt.Errorf("'-' cannot be the upper bound of a range")
	}
	hi, isChar2, _, err := p.classChar()
	if err != nil {
		return nil, err
	}
	if !isChar2 {
		return nil, fmt.Errorf("multi-character escape cannot end a range")
	}
	if hi < lo {
		return nil, fmt.Errorf("invalid range %q-%q", lo, hi)
	}
	return addRange(set, lo, hi), nil
}

// classChar parses one character or escape inside a class. Returns either
// a single rune (isChar) or a set (multi-char/category escape).
func (p *reParser) classChar() (rune, bool, rangeSet, error) {
	c := p.next()
	if c != '\\' {
		return c, true, nil, nil
	}
	if lit, ok, err := p.singleCharEsc(); err != nil {
		return 0, false, nil, err
	} else if ok {
		return lit, true, nil, nil
	}
	set, err := p.classEsc()
	return 0, false, set, err
}

// ---- code point sets ----

// rangeSet is a sorted, disjoint list of inclusive code point ranges.
type rangeSet [][2]rune

const maxRune = 0x10FFFF

func addRange(s rangeSet, lo, hi rune) rangeSet {
	return union(s, rangeSet{{lo, hi}})
}

func union(a, b rangeSet) rangeSet {
	m := append(append(rangeSet{}, a...), b...)
	if len(m) == 0 {
		return nil
	}
	// Sort by lo.
	for i := 1; i < len(m); i++ {
		for j := i; j > 0 && m[j][0] < m[j-1][0]; j-- {
			m[j], m[j-1] = m[j-1], m[j]
		}
	}
	out := rangeSet{m[0]}
	for _, r := range m[1:] {
		last := &out[len(out)-1]
		if r[0] <= last[1]+1 {
			if r[1] > last[1] {
				last[1] = r[1]
			}
		} else {
			out = append(out, r)
		}
	}
	return out
}

func complement(s rangeSet) rangeSet {
	var out rangeSet
	next := rune(0)
	for _, r := range s {
		if r[0] > next {
			out = append(out, [2]rune{next, r[0] - 1})
		}
		if r[1]+1 > next {
			next = r[1] + 1
		}
	}
	if next <= maxRune {
		out = append(out, [2]rune{next, maxRune})
	}
	return out
}

func subtract(a, b rangeSet) rangeSet {
	return intersect(a, complement(b))
}

func intersect(a, b rangeSet) rangeSet {
	var out rangeSet
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		lo := max(a[i][0], b[j][0])
		hi := min(a[i][1], b[j][1])
		if lo <= hi {
			out = append(out, [2]rune{lo, hi})
		}
		if a[i][1] < b[j][1] {
			i++
		} else {
			j++
		}
	}
	return out
}

// emitSet renders a set as a Go regexp character class, with surrogates
// removed (they cannot appear in UTF-8 input anyway).
func emitSet(s rangeSet) string {
	s = subtract(s, rangeSet{{0xD800, 0xDFFF}})
	if len(s) == 0 {
		// Match nothing.
		return `[^\x{0}-\x{10FFFF}]`
	}
	var b strings.Builder
	b.WriteByte('[')
	for _, r := range s {
		if r[0] == r[1] {
			fmt.Fprintf(&b, `\x{%X}`, r[0])
		} else {
			fmt.Fprintf(&b, `\x{%X}-\x{%X}`, r[0], r[1])
		}
	}
	b.WriteByte(']')
	return b.String()
}

func escLiteral(c rune) string {
	switch c {
	case '.', '+', '*', '?', '(', ')', '|', '[', ']', '{', '}', '^', '$', '\\':
		return `\` + string(c)
	}
	if c < 0x20 {
		return fmt.Sprintf(`\x{%X}`, c)
	}
	return string(c)
}

func dotSet() rangeSet {
	return complement(rangeSet{{'\n', '\n'}, {'\r', '\r'}})
}

// foDotSet is the F&O/XQuery default `.`: every character except #x0A (\n).
// It differs from dotSet (the XSD Part 2 Appendix-G pattern-facet `.`) by
// still matching a carriage return (\r). This equals RE2's own default `.`.
func foDotSet() rangeSet {
	return complement(rangeSet{{'\n', '\n'}})
}

func xmlSpaceSet() rangeSet {
	return rangeSet{{'\t', '\n'}, {'\r', '\r'}, {' ', ' '}}
}

// nameStartSet is XML NameStartChar including ':'.
func nameStartSet() rangeSet {
	return rangeSet{
		{':', ':'}, {'A', 'Z'}, {'_', '_'}, {'a', 'z'},
		{0xC0, 0xD6}, {0xD8, 0xF6}, {0xF8, 0x2FF}, {0x370, 0x37D},
		{0x37F, 0x1FFF}, {0x200C, 0x200D}, {0x2070, 0x218F},
		{0x2C00, 0x2FEF}, {0x3001, 0xD7FF}, {0xF900, 0xFDCF},
		{0xFDF0, 0xFFFD}, {0x10000, 0xEFFFF},
	}
}

// nameCharSet is XML NameChar including ':'.
func nameCharSet() rangeSet {
	return union(nameStartSet(), rangeSet{
		{'-', '-'}, {'.', '.'}, {'0', '9'}, {0xB7, 0xB7},
		{0x300, 0x36F}, {0x203F, 0x2040},
	})
}

// wordSet is \w: everything minus punctuation, separators and other.
func wordSet() (rangeSet, error) {
	var excl rangeSet
	for _, name := range [3]string{"P", "Z", "C"} {
		s, err := categorySet(name)
		if err != nil {
			return nil, err
		}
		excl = union(excl, s)
	}
	return complement(excl), nil
}

func rangeTableSet(t *unicode.RangeTable) rangeSet {
	var s rangeSet
	for _, r := range t.R16 {
		if r.Stride == 1 {
			s = append(s, [2]rune{rune(r.Lo), rune(r.Hi)})
		} else {
			for c := rune(r.Lo); c <= rune(r.Hi); c += rune(r.Stride) {
				s = append(s, [2]rune{c, c})
			}
		}
	}
	for _, r := range t.R32 {
		if r.Stride == 1 {
			s = append(s, [2]rune{rune(r.Lo), rune(r.Hi)})
		} else {
			for c := rune(r.Lo); c <= rune(r.Hi); c += rune(r.Stride) {
				s = append(s, [2]rune{c, c})
			}
		}
	}
	return union(nil, s)
}

var categoryCache = map[string]rangeSet{}

func categorySet(name string) (rangeSet, error) {
	if s, ok := categoryCache[name]; ok {
		return s, nil
	}
	var s rangeSet
	if t, ok := unicode.Categories[name]; ok {
		s = rangeTableSet(t)
	} else if len(name) == 1 {
		// Group category: union of all specific categories with the prefix.
		found := false
		for n, t := range unicode.Categories {
			if len(n) == 2 && n[0] == name[0] {
				found = true
				s = union(s, rangeTableSet(t))
			}
		}
		if !found {
			return nil, fmt.Errorf("unknown character category %q", name)
		}
		if name == "C" {
			// Go's C tables omit Cn (unassigned); XSD's C includes it.
			cn, err := categorySet("Cn")
			if err != nil {
				return nil, err
			}
			s = union(s, cn)
		}
	} else if name == "Cn" {
		var assigned rangeSet
		for n, t := range unicode.Categories {
			if len(n) == 2 {
				assigned = union(assigned, rangeTableSet(t))
			}
		}
		s = complement(assigned)
	} else {
		return nil, fmt.Errorf("unknown character category %q", name)
	}
	categoryCache[name] = s
	return s, nil
}

var validCategories = map[string]bool{
	"L": true, "Lu": true, "Ll": true, "Lt": true, "Lm": true, "Lo": true,
	"M": true, "Mn": true, "Mc": true, "Me": true,
	"N": true, "Nd": true, "Nl": true, "No": true,
	"P": true, "Pc": true, "Pd": true, "Ps": true, "Pe": true, "Pi": true, "Pf": true, "Po": true,
	"Z": true, "Zs": true, "Zl": true, "Zp": true,
	"S": true, "Sm": true, "Sc": true, "Sk": true, "So": true,
	"C": true, "Cc": true, "Cf": true, "Co": true, "Cn": true,
}

// propertySet resolves the name in \p{…}: a category or an Is-block.
func propertySet(name string) (rangeSet, error) {
	if strings.HasPrefix(name, "Is") {
		block := name[2:]
		if !validBlockName(block) {
			return nil, fmt.Errorf("malformed block name %q", name)
		}
		if s, ok := unicodeBlocks[block]; ok {
			return s, nil
		}
		// Per Part 2 Appendix G, an unrecognized block name defaults to
		// matching any character (both for \p and \P).
		return complement(nil), nil
	}
	if !validCategories[name] {
		return nil, fmt.Errorf("unknown character property %q", name)
	}
	return categorySet(name)
}

// validBlockName: block names in escapes are [a-zA-Z0-9-]+ (spaces and
// underscores removed from the Unicode names, hyphens kept).
func validBlockName(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '-' {
			return false
		}
	}
	return true
}
