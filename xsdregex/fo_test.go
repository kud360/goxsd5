package xsdregex

import (
	"regexp"
	"testing"
)

// foMatch translates an F&O pattern+flags and reports whether input matches,
// failing the test if translation or compilation errors unexpectedly.
func foMatch(t *testing.T, pattern, flags, input string) bool {
	t.Helper()
	s, err := TranslateFO(pattern, flags)
	if err != nil {
		t.Fatalf("TranslateFO(%q, %q): unexpected error %v", pattern, flags, err)
	}
	re, err := regexp.Compile(s)
	if err != nil {
		t.Fatalf("TranslateFO(%q, %q) = %q: does not compile: %v", pattern, flags, s, err)
	}
	return re.MatchString(input)
}

func TestTranslateFO(t *testing.T) {
	cases := []struct {
		pattern string
		flags   string
		input   string
		want    bool
	}{
		// `^`/`$` are real anchors, not literals: an anchored pattern only
		// matches when the whole string satisfies it (the Appendix-G bug this
		// fixes was a false accept for anchored patterns).
		{"^[A-Z]+$", "", "ABC", true},
		{"^[A-Z]+$", "", "ABCx", false},
		{"^[A-Z]+$", "", "xABC", false},
		{"^abc", "", "abcdef", true},
		{"^abc", "", "xabc", false},
		{"abc$", "", "xxabc", true},
		{"abc$", "", "abcx", false},
		// Unanchored: matches a substring.
		{"abc", "", "xxabcyy", true},
		{"abc", "", "ab", false},
		// `.` excludes newline by default; the `s` flag makes it match.
		{"a.c", "", "abc", true},
		{"a.c", "", "a\nc", false},
		{"a.c", "s", "a\nc", true},
		// `i` flag: case-insensitive.
		{"^abc$", "i", "ABC", true},
		{"^abc$", "", "ABC", false},
		// Reluctant quantifiers compile and match.
		{"a+?", "", "aaa", true},
		{"^a{2,3}?", "", "aaa", true},
	}
	for _, c := range cases {
		if got := foMatch(t, c.pattern, c.flags, c.input); got != c.want {
			t.Errorf("TranslateFO(%q,%q) match %q = %v, want %v", c.pattern, c.flags, c.input, got, c.want)
		}
	}
}

func TestTranslateFOErrors(t *testing.T) {
	cases := []struct {
		pattern string
		flags   string
	}{
		{"^a$", "m"},     // multi-line: RE2 cannot express F&O line boundaries
		{"a b", "x"},     // free-spacing: not expressible in RE2
		{"abc", "q"},     // literal mode: not expressible in RE2
		{"abc", "z"},     // unknown flag
		{`(a)\1`, ""},    // back-reference: RE2 cannot express it
		{`(a)(b)\2`, ""}, // back-reference to group 2
		{"(a", ""},       // unterminated group (grammar error still surfaces)
		{`\p{Foo}`, ""},  // unknown category
	}
	for _, c := range cases {
		if _, err := TranslateFO(c.pattern, c.flags); err == nil {
			t.Errorf("TranslateFO(%q, %q) should return an error", c.pattern, c.flags)
		}
	}
}
