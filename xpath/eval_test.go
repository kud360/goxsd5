package xpath

import "testing"

// testNode is a tiny in-memory Node for evaluator tests.
type testNode struct {
	name     Name
	attrs    []NodeAttr
	children []Node
	text     string
}

func (n *testNode) NodeName() Name        { return n.name }
func (n *testNode) NodeAttrs() []NodeAttr { return n.attrs }
func (n *testNode) NodeChildren() []Node  { return n.children }
func (n *testNode) StringValue() string {
	if n.text != "" || len(n.children) == 0 {
		return n.text
	}
	s := n.text
	for _, c := range n.children {
		s += c.StringValue()
	}
	return s
}

type testAttr struct {
	name Name
	val  string
}

func (a testAttr) AttrName() Name    { return a.name }
func (a testAttr) AttrValue() string { return a.val }

func at(local, val string) NodeAttr { return testAttr{Name{Local: local}, val} }
func el(local string) *testNode     { return &testNode{name: Name{Local: local}} }

// castInt accepts only integer-looking values, a stand-in for xs:integer.
func castInt(typ, val string) bool {
	if typ != "integer" {
		return typ == "string"
	}
	for i, c := range val {
		if c == '-' && i == 0 {
			continue
		}
		if c < '0' || c > '9' {
			return false
		}
	}
	return val != "" && val != "-"
}

func TestEvalBool(t *testing.T) {
	root := el("temp")
	root.attrs = []NodeAttr{at("x", "204"), at("y", "abc"), at("n", "5")}
	a := el("a")
	a.text = "1"
	b := el("b")
	b.text = "2"
	root.children = []Node{a, b}

	ec := EvalContext{Castable: castInt}
	cases := []struct {
		expr string
		want bool
		ok   bool
	}{
		{"@x > 300", false, true},
		{"@x > 100", true, true},
		{"@x", true, true},
		{"@missing", false, true},
		{"@x = 204", true, true},
		{"@x eq 204", true, true},
		{"@y = 'abc'", true, true},
		{"@x > 100 and @n = 5", true, true},
		{"@x < 100 or @n = 5", true, true},
		{"not(@x = 204)", false, true},
		{"count(a) = 1", true, true},
		{"count(*) = 2", true, true},
		{"exists(a)", true, true},
		{"empty(c)", true, true},
		{"a = 1", true, true},
		{"string-length(@y) = 3", true, true},
		{"100 div xs:integer(@n) > 50", false, true},
		{"100 div xs:integer(@n) >= 20", true, true},
		{"@x castable as xs:integer", true, true},
		{"@y castable as xs:integer", false, true},
		{"if (@n = 5) then @x = 204 else false()", true, true},
		{"contains(@y, 'b')", true, true},
		{"@x > unknown:func()", false, false}, // unsupported → fail open
		{"@@@", false, false},                 // garbage → fail open
	}
	for _, c := range cases {
		got, ok := EvalBool(c.expr, root, ec)
		if got != c.want || ok != c.ok {
			t.Errorf("EvalBool(%q) = (%v,%v), want (%v,%v)", c.expr, got, ok, c.want, c.ok)
		}
	}
}

// TestEvalSubtree exercises the XSD-assertion "stay in subtree" rule: a leading
// "/" or "//" is rooted at a document node that does not exist for a parentless
// assertion root, so it selects nothing; a relative ".//" descends the subtree
// (descendant-or-self), reaching nested element and attribute nodes.
func TestEvalSubtree(t *testing.T) {
	// ele2 / subElement2 / ele1[@attr1=2] — the inner ele1 is a descendant.
	inner := el("ele1")
	inner.attrs = []NodeAttr{at("attr1", "2")}
	sub := el("subElement2")
	sub.children = []Node{inner}
	root := el("ele2")
	root.attrs = []NodeAttr{at("attr2", "3")}
	root.children = []Node{sub}

	ec := EvalContext{Castable: castInt}
	cases := []struct {
		expr string
		want bool
		ok   bool
	}{
		{"count(.//@attr1) eq 1", true, true}, // descendant attribute reached
		{"count(.//ele1) eq 1", true, true},   // descendant element reached
		{"count(//ele1) eq 1", false, true},   // absolute // is empty → 0
		{"count(//ele1) eq 0", true, true},    // ... so eq 0 holds
		{"count(//@attr1) eq 1", false, true}, // absolute //@ is empty → 0
		{"//ele1", false, true},               // absolute path → empty → false
		{"exists(.//ele1)", true, true},       // relative descendant exists
		{"count(.//*) eq 2", true, true},      // subElement2 + ele1
	}
	for _, c := range cases {
		got, ok := EvalBool(c.expr, root, ec)
		if got != c.want || ok != c.ok {
			t.Errorf("EvalBool(%q) = (%v,%v), want (%v,%v)", c.expr, got, ok, c.want, c.ok)
		}
	}
}

// TestEvalSeqUnion covers sequence construction "(a,b,c)", the range operator
// "to", and the node-set union "|".
func TestEvalSeqUnion(t *testing.T) {
	root := el("root")
	root.attrs = []NodeAttr{at("time", "9"), at("iterations", "3")}
	a, b := el("a"), el("b")
	a.text, b.text = "x", "y"
	root.children = []Node{a, b}

	ec := EvalContext{Castable: castInt}
	cases := []struct {
		expr string
		want bool
		ok   bool
	}{
		{"5 = (1 to 10, 20, 30)", true, true},          // range + sequence, existential =
		{"15 = (1 to 10, 20, 30)", false, true},        // not in the set
		{"20 = (1 to 10, 20, 30)", true, true},         // explicit member
		{"count(@time | @iterations) = 2", true, true}, // attribute union
		{"count(@time | @missing) = 1", true, true},    // union with empty side
		{"@time = ('1', '9', '7')", true, true},        // string sequence membership
		{"'z' = ('x', 'y')", false, true},              // not a member
		{"count(a | b) = 2", true, true},               // element union
	}
	for _, c := range cases {
		got, ok := EvalBool(c.expr, root, ec)
		if got != c.want || ok != c.ok {
			t.Errorf("EvalBool(%q) = (%v,%v), want (%v,%v)", c.expr, got, ok, c.want, c.ok)
		}
	}
}

// TestEvalVars covers "$name" variable references, as used by simple-type
// assertions binding $value to the value under validation.
func TestEvalVars(t *testing.T) {
	root := el("v")
	ec := EvalContext{
		Castable: castInt,
		Vars:     map[string][]string{"value": {"4"}, "list": {"a", "b", "c"}},
	}
	cases := []struct {
		expr string
		want bool
		ok   bool
	}{
		{"$value mod 2 = 0", true, true},       // even
		{"$value = 4", true, true},             // numeric compare
		{"ends-with($value, '4')", true, true}, // string function on $value
		{"$value castable as xs:integer", true, true},
		{"count($list) = 3", true, true}, // multi-item variable
		{"$list = 'b'", true, true},      // existential over sequence
		{"$missing > 1", false, false},   // unbound → fail open
		// current-date() as an ISO string compares chronologically; these hold
		// for any plausible run date.
		{"current-date() gt '1900-01-01'", true, true},
		{"current-date() lt '9999-01-01'", true, true},
	}
	for _, c := range cases {
		got, ok := EvalBool(c.expr, root, ec)
		if got != c.want || ok != c.ok {
			t.Errorf("EvalBool(%q) = (%v,%v), want (%v,%v)", c.expr, got, ok, c.want, c.ok)
		}
	}
}

// castDate is a Castable that accepts the ISO date forms used below.
func castDate(typ, val string) bool {
	if typ == "date" {
		// crude YYYY-MM-DD recogniser (optionally with a timezone suffix); any
		// stray character (e.g. the "!!!" of assert-simple007) makes it not a date
		if len(val) < 10 || val[4] != '-' || val[7] != '-' {
			return false
		}
		for i, c := range val {
			if i == 4 || i == 7 {
				continue
			}
			if (c < '0' || c > '9') && c != '+' && c != ':' && c != 'Z' {
				return false
			}
		}
		return true
	}
	if typ == "boolean" {
		return val == "true" || val == "false" || val == "0" || val == "1"
	}
	return castInt(typ, val)
}

// TestEvalTypedAtoms covers schema-typed $value bindings (TypedVars): a
// string-kind value never coerces to a number; numeric and date comparisons
// still work; data() atomises a bound sequence; and the simple-type-assertion
// dynamic-error rules (NoContextItem, failed casts) yield a definite false.
func TestEvalTypedAtoms(t *testing.T) {
	root := el("v")
	cases := []struct {
		expr      string
		typed     map[string][]TypedAtom
		noContext bool
		want, ok  bool
	}{
		// String-kind: "\n   100\n" is not the string "100" (no numeric coercion).
		{`$value eq '100'`, map[string][]TypedAtom{"value": {{"\n   100\n", KindString}}}, false, false, true},
		{`$value = '100'`, map[string][]TypedAtom{"value": {{"\n   100\n", KindString}}}, false, false, true},
		{`$value eq '100'`, map[string][]TypedAtom{"value": {{"100", KindString}}}, false, true, true},
		// Number-kind: a numeric-looking value compares numerically.
		{`$value mod 2 = 0`, map[string][]TypedAtom{"value": {{"4", KindNumber}}}, false, true, true},
		{`$value = 100`, map[string][]TypedAtom{"value": {{"100", KindNumber}}}, false, true, true},
		// Date as string-kind: ISO lexicals order chronologically by string.
		{`$value lt current-date()`, map[string][]TypedAtom{"value": {{"2001-01-01", KindString}}}, false, true, true},
		{`$value lt current-date()`, map[string][]TypedAtom{"value": {{"9999-10-10", KindString}}}, false, false, true},
		// data() over a bound sequence atomises it; a quantifier sees each item.
		{`count(data($value)) = 3`, map[string][]TypedAtom{"value": {{"2", KindNumber}, {"4", KindNumber}, {"6", KindNumber}}}, false, true, true},
		{`every $x in data($value) satisfies ($x mod 2 = 0)`, map[string][]TypedAtom{"value": {{"2", KindNumber}, {"7", KindNumber}}}, false, false, true},
		// NoContextItem: referencing the absent context item is a dynamic error,
		// reported as a definite false (ok true) so the assertion is unsatisfied.
		{`position() le 50`, map[string][]TypedAtom{"value": {{"x", KindString}}}, true, false, true},
		{`last() le 50`, map[string][]TypedAtom{"value": {{"x", KindString}}}, true, false, true},
		{`. castable as xs:date`, map[string][]TypedAtom{"value": {{"2008-05-01", KindString}}}, true, false, true},
		// A failed constructor cast is a dynamic error (definite false), not a
		// fail-open unsupported construct.
		{`xs:date(concat($value, '!!!')) gt xs:date('1900-01-01')`, map[string][]TypedAtom{"value": {{"2001-01-01", KindString}}}, true, false, true},
		// fn:*-from-date/dateTime/time component extraction.
		{`year-from-date($value) eq 2008`, map[string][]TypedAtom{"value": {{"2008-07-28", KindString}}}, false, true, true},
		{`year-from-date($value) eq 2008`, map[string][]TypedAtom{"value": {{"2018-07-28", KindString}}}, false, false, true},
		{`month-from-date($value) = 11`, map[string][]TypedAtom{"value": {{"2008-11-05+01:00", KindString}}}, false, true, true},
		{`day-from-dateTime($value) = 5`, map[string][]TypedAtom{"value": {{"2008-11-05T12:30:00", KindString}}}, false, true, true},
		{`hours-from-time($value) = 23`, map[string][]TypedAtom{"value": {{"23:15:00Z", KindString}}}, false, true, true},
		// "cast as": xs:boolean yields a real boolean; a numeric kind a number;
		// a non-castable value is a dynamic error (definite false).
		{`$value cast as xs:boolean`, map[string][]TypedAtom{"value": {{"false", KindString}}}, false, false, true},
		{`$value cast as xs:boolean`, map[string][]TypedAtom{"value": {{"true", KindString}}}, false, true, true},
		{`($value cast as xs:integer) mod 2 = 0`, map[string][]TypedAtom{"value": {{"11", KindString}}}, false, false, true},
		{`$value cast as xs:integer`, map[string][]TypedAtom{"value": {{"notanint", KindString}}}, false, false, true},
	}
	for _, c := range cases {
		ec := EvalContext{Castable: castDate, TypedVars: c.typed, NoContextItem: c.noContext}
		got, ok := EvalBool(c.expr, root, ec)
		if got != c.want || ok != c.ok {
			t.Errorf("EvalBool(%q) = (%v,%v), want (%v,%v)", c.expr, got, ok, c.want, c.ok)
		}
	}
}

// TestEvalAxes covers the reverse and sibling axes (synthesized from the
// positioned-node parent chain) plus quantified and for expressions. The tree:
//
//	game / { white(e4) black(e5) white(d4) white(c3) result(draw) }
func TestEvalAxes(t *testing.T) {
	mk := func(name, text string) *testNode { n := el(name); n.text = text; return n }
	w1, b1 := mk("white", "e4"), mk("black", "e5")
	w2, w3 := mk("white", "d4"), mk("white", "c3")
	res := mk("result", "draw")
	game := el("game")
	game.children = []Node{w1, b1, w2, w3, res}

	ec := EvalContext{Castable: castInt}
	cases := []struct {
		expr string
		want bool
		ok   bool
	}{
		// following-sibling: w2(d4) is immediately followed by w3, another white.
		{"some $w in white satisfies $w/following-sibling::*[1][self::white]", true, true},
		// every white is immediately followed by a white? no (w1 is followed by black).
		{"every $w in white satisfies $w/following-sibling::*[1][self::white]", false, true},
		// no two consecutive whites? false here (d4 then c3).
		{"every $w in white satisfies not($w/following-sibling::*[1][self::white])", false, true},
		{"count(white) = 3", true, true},
		{"result/preceding-sibling::white[1] = 'c3'", true, true}, // nearest preceding white
		{"count(white/..) = 1", true, true},                       // parent of all whites is the one game
		{"every $w in white satisfies $w/.. = $w/parent::game", true, true},
		{"result/parent::game/result = 'draw'", true, true},
		{"count(result/preceding::white) = 3", true, true},             // preceding axis
		{"count(white[1]/following::*) = 4", true, true},               // following axis from first white
		{"(for $w in white return string-length($w)) = 2", true, true}, // for + string-length
		{"some $w in white satisfies $w/ancestor::game", true, true},   // ancestor axis
	}
	for _, c := range cases {
		got, ok := EvalBool(c.expr, game, ec)
		if got != c.want || ok != c.ok {
			t.Errorf("EvalBool(%q) = (%v,%v), want (%v,%v)", c.expr, got, ok, c.want, c.ok)
		}
	}
}
