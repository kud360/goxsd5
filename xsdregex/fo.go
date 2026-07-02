package xsdregex

import "fmt"

// XPath/XQuery 2.0 Functions & Operators (F&O) regex flavor, used by
// fn:matches / fn:replace / fn:tokenize in assertions. XSD 1.1 normatively
// binds those functions to the F&O grammar (F&O Appendix F), NOT the XSD
// Part 2 Appendix-G pattern-facet grammar that TranslateRegex/Matches use.
// The two flavors share character-class handling (\d \w \p{…}, [a-z-[m]],
// implicit sets) but differ in three ways this translator implements:
//
//   - `^` and `$` are ANCHORS (start/end of string), not literal characters:
//     `^` → \A, `$` → \z. In default mode there are no line boundaries, so
//     they anchor the whole string. (The `m` multi-line flag would change
//     that, but RE2 cannot express F&O's line semantics — see TranslateFO.)
//   - `.` matches every character EXCEPT newline by default; the `s`
//     (dot-all) flag makes it match newlines too — exactly RE2's own `.`.
//   - Reluctant quantifiers (`a+?`, `a{2,3}?`) exist and map straight to RE2;
//     back-references (`\1`) exist but RE2 cannot express them, so they are
//     rejected rather than mistranslated.

// TranslateFO translates an F&O-flavor pattern into a Go RE2 source string,
// applying the fn:matches/replace/tokenize flag string. The returned pattern
// is UNANCHORED (F&O matches a substring unless the pattern itself anchors
// with `^`/`$`); callers compile it directly.
//
// Supported flags: "i" (case-insensitive) and "s" (dot-all). The "m", "x" and
// "q" flags are NOT expressible in RE2 — F&O's multi-line line boundaries and
// free-spacing/literal modes have no RE2 equivalent — and any other flag
// character is undefined; all of these return an error so a caller surfaces a
// dynamic error (assertion unsatisfied) rather than silently ignoring the flag.
func TranslateFO(pattern, flags string) (string, error) {
	inline, dotAll, err := foFlags(flags)
	if err != nil {
		return "", err
	}
	p := &reParser{src: []rune(pattern), fo: true, dotAll: dotAll}
	s, err := p.regExp()
	if err != nil {
		return "", fmt.Errorf("invalid F&O regex %q at offset %d: %w", pattern, p.pos, err)
	}
	if p.pos != len(p.src) {
		return "", fmt.Errorf("invalid F&O regex %q: unexpected %q at offset %d", pattern, p.src[p.pos], p.pos)
	}
	return inline + `(?:` + s + `)`, nil
}

// foFlags translates an F&O flag string into an RE2 inline-flag prefix and
// reports whether dot-all is in effect. "i" → case-insensitive, "s" → dot-all
// (handled at translation time so `.` becomes an all-characters set). "m", "x"
// and "q" — and any other flag — are rejected: RE2 cannot express F&O's line
// boundaries (m), free-spacing (x) or literal (q) modes, so honouring them
// would give wrong answers.
func foFlags(flags string) (inline string, dotAll bool, err error) {
	caseless := false
	for _, c := range flags {
		switch c {
		case 'i':
			caseless = true
		case 's':
			dotAll = true
		default:
			return "", false, fmt.Errorf("xsdregex: unsupported F&O regex flag %q", c)
		}
	}
	if caseless {
		return "(?i)", dotAll, nil
	}
	return "", dotAll, nil
}
