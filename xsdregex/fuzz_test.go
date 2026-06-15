package xsdregex

import (
	"regexp"
	"testing"
)

// regexSeeds are lexical patterns exercising the XSD-regex surface the
// translator has to bridge to Go's RE2 syntax: anchoring, the XSD-only
// metacharacters and escapes, multi-char escapes, char-class subtraction,
// quantifiers, and the Unicode block/category shorthands.
var regexSeeds = []string{
	"",
	"abc",
	"a.c",
	"a^b", "a$b", // ^ and $ are ordinary characters in XSD regex
	"a{2,3}", "a{2}", "a{2,}", "a*", "a+", "a?",
	`\d`, `\D`, `\w`, `\W`, `\s`, `\S`,
	`\i`, `\I`, `\c`, `\C`, // XSD name-char escapes
	`[a-z]`, `[^a-z]`, `[a-z-[aeiou]]`, // class subtraction
	`\p{L}`, `\P{L}`, `\p{Nd}`, `\p{IsBasicLatin}`, `\p{IsGreek}`,
	`(a|b)c`, `(?:ab)+`, `a\.b`, `\(`, `\)`, `\[`, `\]`, `\{`, `\}`,
	`\\`, `\n`, `\r`, `\t`,
	`[`, `]`, `(`, `)`, `{`, `}`, `\`, `\p{`, `[a-`, `a{`, `a{,2}`,
	`.*`, `[\d-a]`, `\p{IsNoSuchBlock}`,
}

// FuzzTranslateRegex asserts the translator never panics and, on success,
// always emits a pattern Go's regexp accepts — TranslateRegex's contract is
// that a non-error result is RE2-compilable. Determinism is checked too.
func FuzzTranslateRegex(f *testing.F) {
	for _, s := range regexSeeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, pattern string) {
		out, err := TranslateRegex(pattern)
		if err != nil {
			return
		}
		if _, cerr := regexp.Compile(out); cerr != nil {
			t.Fatalf("TranslateRegex(%q) = %q, but regexp.Compile rejected it: %v", pattern, out, cerr)
		}
		// Translation must be deterministic.
		if out2, err2 := TranslateRegex(pattern); err2 != nil || out2 != out {
			t.Fatalf("TranslateRegex(%q) non-deterministic: (%q,%v) then (%q,%v)", pattern, out, err, out2, err2)
		}
	})
}

// FuzzCompileRegex drives the full compile path (translate + RE2 build) and
// asserts no panic. A successful compile must yield a usable, non-nil matcher
// that itself does not panic on arbitrary input.
func FuzzCompileRegex(f *testing.F) {
	for _, s := range regexSeeds {
		f.Add(s, "abc")
	}
	f.Fuzz(func(t *testing.T, pattern, subject string) {
		re, err := CompileRegex(pattern)
		if err != nil {
			return
		}
		if re == nil {
			t.Fatalf("CompileRegex(%q) returned nil matcher, nil error", pattern)
		}
		_ = re.MatchString(subject)
	})
}
