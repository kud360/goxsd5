package xsd

import "testing"

func TestRegexTranslate(t *testing.T) {
	cases := []struct {
		pattern string
		match   []string
		nomatch []string
	}{
		// Implicit anchoring.
		{"abc", []string{"abc"}, []string{"xabc", "abcx", "ab"}},
		// ^ and $ are ordinary characters in XSD regexes.
		{"a^b", []string{"a^b"}, []string{"ab"}},
		{"a$b", []string{"a$b"}, []string{"ab"}},
		// Wildcard excludes newlines.
		{"a.c", []string{"abc", "a c"}, []string{"a\nc", "a\rc", "ac"}},
		// Quantifiers.
		{"a{2,3}", []string{"aa", "aaa"}, []string{"a", "aaaa"}},
		{"a{2}", []string{"aa"}, []string{"a", "aaa"}},
		{"a{2,}", []string{"aa", "aaaaa"}, []string{"a"}},
		{"(ab)+", []string{"ab", "abab"}, []string{"", "aba"}},
		{"a?b", []string{"b", "ab"}, []string{"aab"}},
		// Large counted repeats (beyond RE2's 1000 limit).
		{"a{1500}", []string{rep("a", 1500)}, []string{rep("a", 1499), rep("a", 1501)}},
		// Char classes and ranges.
		{"[a-c]+", []string{"abc", "a"}, []string{"d", ""}},
		{"[-abc]", []string{"-", "a"}, []string{"d"}},
		{"[abc-]", []string{"-", "c"}, []string{"d"}},
		{"[^a-c]", []string{"d", "-"}, []string{"a", "b"}},
		// Class subtraction.
		{"[a-z-[aeiou]]+", []string{"bcd", "xyz"}, []string{"a", "bea"}},
		{"[a-z-[m-p-[o]]]+", []string{"abo", "z"}, []string{"m", "n", "p"}},
		{"[^a-z-[aeiou]]+", []string{"123", "AB"}, []string{"b", "a"}},
		// Multi-char escapes.
		{`\d+`, []string{"42", "٤٢"}, []string{"x", "4x"}},
		{`\D`, []string{"x"}, []string{"4"}},
		{`\s+`, []string{" \t\n\r"}, []string{"x", "\v"}},
		{`\w+`, []string{"abc42"}, []string{"a b", "a,b"}},
		{`\i\c*`, []string{"foo", "_x:y", "a-b.c"}, []string{"1ab", "-ab"}},
		{`[\i-[:]][\c-[:]]*`, []string{"foo", "_a1"}, []string{"a:b", "1a"}},
		// Category and block escapes.
		{`\p{Lu}+`, []string{"ABC", "ÉÀ"}, []string{"abc", "1"}},
		{`\p{L}+`, []string{"abcÉ"}, []string{"a1"}},
		{`\P{Lu}+`, []string{"abc1"}, []string{"A"}},
		{`\p{IsBasicLatin}+`, []string{"abc!"}, []string{"é"}},
		{`\P{IsBasicLatin}+`, []string{"éé"}, []string{"a"}},
		{`\p{IsGreek}`, []string{"α"}, []string{"a"}},
		// Escaped metacharacters.
		{`\.\*\+\?\(\)\{\}\[\]\\\|\-\^`, []string{`.*+?(){}[]\|-^`}, nil},
		{`a\nb`, []string{"a\nb"}, []string{"anb"}},
		// Alternation and grouping.
		{"a|b|", []string{"a", "b", ""}, []string{"c"}},
		{"(a|b)c", []string{"ac", "bc"}, []string{"c"}},
		// Empty pattern matches only the empty string.
		{"", []string{""}, []string{"a"}},
	}
	for _, tc := range cases {
		re, err := CompileRegex(tc.pattern)
		if err != nil {
			t.Errorf("CompileRegex(%q): %v", tc.pattern, err)
			continue
		}
		for _, s := range tc.match {
			if !re.MatchString(s) {
				t.Errorf("pattern %q should match %q (go: %s)", tc.pattern, s, re)
			}
		}
		for _, s := range tc.nomatch {
			if re.MatchString(s) {
				t.Errorf("pattern %q should NOT match %q (go: %s)", tc.pattern, s, re)
			}
		}
	}
}

func rep(s string, n int) string {
	out := make([]byte, 0, n*len(s))
	for range n {
		out = append(out, s...)
	}
	return string(out)
}

func TestRegexInvalid(t *testing.T) {
	for _, pattern := range []string{
		"a{2,1}",   // max < min
		"a{",       // malformed quantifier
		"+a",       // quantifier without atom
		"a**",      // double quantifier
		"(a",       // unterminated group
		"a)",       // stray close
		"[a",       // unterminated class
		"[]",       // empty class
		"[^]",      // empty negated class
		`\q`,       // invalid escape
		`\$`,       // $ is not escapable in XSD
		`\p{Foo}`,  // unknown category
		`\p{Lx}`,   // unknown category
		`\px`,      // malformed \p
		"[z-a]",    // reversed range
		`[\d-x]`,   // multi-char escape as range start
		"a\\",      // trailing backslash
		"[a-z-b]",  // misplaced - in class
		"]",        // unescaped ] outside class
		"}",        // unescaped } outside class (strict)
	} {
		if _, err := CompileRegex(pattern); err == nil {
			t.Errorf("CompileRegex(%q) should fail", pattern)
		}
	}
}
