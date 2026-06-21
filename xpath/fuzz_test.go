package xpath

// Fuzz coverage for the XPath front-end (Parse) and the assertion evaluator
// (EvalBool). Both run over untrusted input: Parse is fed schema-author
// expressions from xs:assert/xs:alternative, and EvalBool evaluates them with a
// $value binding taken from the instance document being validated. A panic in
// either is a denial-of-service vector for any caller of xsdvalidate.

import (
	"strings"
	"testing"
)

// parseSeeds combine the syntactically valid and invalid expressions guarded by
// xpath_test.go (one per grammar production) with a handful of structurally
// broken fragments, so the fuzzer starts from full grammar coverage.
var parseSeeds = func() []string {
	seeds := make([]string, 0, len(valid)+len(invalid)+8)
	seeds = append(seeds, valid...)
	seeds = append(seeds, invalid...)
	seeds = append(seeds,
		"", " ", "(", ")", "$", "@", "//", "::",
	)
	return seeds
}()

// FuzzParseXPath asserts the parser never panics on arbitrary input and that any
// error it returns is the package's typed *Error (never an untyped recovery),
// while any successful parse yields a non-nil Expr whose recorded source matches
// the input.
func FuzzParseXPath(f *testing.F) {
	for _, s := range parseSeeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, src string) {
		expr, err := Parse(src)
		if err != nil {
			if _, ok := err.(*Error); !ok {
				t.Fatalf("Parse(%q) returned %T, want *xpath.Error: %v", src, err, err)
			}
			return
		}
		if expr == nil {
			t.Fatalf("Parse(%q) returned nil Expr, nil error", src)
		}
		if expr.Src != src {
			t.Fatalf("Parse(%q): Expr.Src = %q, want the input", src, expr.Src)
		}
	})
}

// evalExprs are representative assertion expressions, one per major XPath
// production drawn from xpath_test.go, each referencing $value so the fuzzed
// binding actually flows through evaluation. EvalBool is documented to never
// return an error (it folds dynamic errors into a definite false and out-of-
// subset constructs into ok=false); its contract here is simply that no $value,
// however strange, makes it panic.
var evalExprs = []string{
	"$value lt 500",
	"$value mod 2 = 0",
	"$value = 'Hello World'",
	"$value castable as xs:date",
	"$value instance of xs:string",
	"not($value instance of xs:string)",
	"every $x in data($value) satisfies ($x mod 2 = 0)",
	"if ($value = '1') then true() else false()",
	"string-length($value) gt 0",
	"$value cast as xs:int > 0",
	". = $value",
}

// valueSeeds are the numeric, date/time, and string lexicals used across the
// assertion conformance tests, plus a few hostile shapes (empty, whitespace,
// non-UTF-8-ish, very long).
var valueSeeds = []string{
	"", " ", "0", "1", "-1", "100", "500", "3.14", "1e9",
	"NaN", "INF", "-INF",
	"2001-10-26", "2001-10-26T21:32:52", "P1Y2M3D", "13:20:00",
	"Hello World", "true", "false", "us", "draw",
	"a b c", "\t\n", "café", strings.Repeat("9", 64),
}

// fuzzNode is a minimal Node: a leaf element whose string value is the fuzzed
// $value, giving "." and string functions something to read.
type fuzzNode struct{ value string }

func (n fuzzNode) NodeName() Name        { return Name{Local: "e"} }
func (n fuzzNode) NodeAttrs() []NodeAttr { return nil }
func (n fuzzNode) NodeChildren() []Node  { return nil }
func (n fuzzNode) StringValue() string   { return n.value }

// FuzzEvalXPath binds an arbitrary $value into each representative assertion
// expression and evaluates it. The sole invariant is that EvalBool never panics
// for any binding; its two-valued (result, ok) return is by contract never an
// untyped panic recovery.
func FuzzEvalXPath(f *testing.F) {
	for _, v := range valueSeeds {
		f.Add(v)
	}
	f.Fuzz(func(t *testing.T, value string) {
		node := fuzzNode{value: value}
		ec := EvalContext{
			Vars:      map[string][]string{"value": {value}},
			TypedVars: map[string][]TypedAtom{"value": {{Lexical: value, Kind: KindUntyped}}},
			Castable:  func(typeLocal, val string) bool { return val != "" },
		}
		for _, expr := range evalExprs {
			// Must not panic for any expr/value pair.
			_, _ = EvalBool(expr, node, ec)
		}
	})
}
