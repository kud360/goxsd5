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
