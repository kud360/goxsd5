package xpath

import "testing"

// valid lists representative XPath 2.0 expressions, at least one per grammar
// production, drawn from the W3C XSD test suite's <alternative>/<assert> tests.
// Every one must parse without error. (The conformance ratchet exercises the
// full suite corpus end to end; this list guards the grammar directly.)
var valid = []string{
	// Literals, context item, variable references.
	"0 = 1",
	"@value eq 100",
	"y = 'Hello World'",
	"$value lt 500",
	".",
	". mod 2 = 0",
	// Paths, axes, predicates, wildcards.
	"@country = 'us'",
	"//e1InB",
	"/root = 'present'",
	"count(//ele1) eq 2",
	"self::message",
	"self::c2:message",
	"data(child::d[1]) instance of xs:date",
	"every $y in y[position() lt last()] satisfies ($y lt $y/following-sibling::y[1])",
	"not(a[preceding::a[not(b)]])",
	"empty(following::*) and empty(preceding::*)",
	"in-scope-prefixes(.) = 'a'",
	"count(@time | @iterations) = 1",
	"@test1:a mod 2 = 0",
	// Operators across every precedence level.
	"@numberOfChildren < 5 and @numberOfChildren > 0",
	"a and b and c and d",
	"(@kind = 'square') or (@kind = 'rectangle')",
	"@x > @y",
	"@min le @max",
	"@end-time <= 10",
	". = (1 to 10, 20, 30)",
	"bill-amount = sum(items/item/@price) + tax",
	"tax = sum(items/item/@price) * 0.1",
	"1 idiv string-length(@type) gt 0",
	"100 div xs:integer(@x) > 50",
	"@value mod 2 = 0",
	"@a is @b",
	". is root()",
	// for / quantified / conditional.
	"every $x in data($value) satisfies ($x mod 2 = 0)",
	"if (@kind = 'rectangle') then (a = c and b = d) else true()",
	// cast / castable / treat / instance of, occurrence indicators, kind tests.
	"@a cast as xs:boolean",
	"$value castable as xs:date",
	"@length cast as xs:int > @width cast as xs:int",
	"$value instance of xs:date",
	"data(event/d) instance of xs:date*",
	"not($value instance of xs:string)",
	". instance of element(*, xs:untyped)",
	"@type instance of attribute(*, xs:untypedAtomic)",
	"data(.) instance of xs:untypedAtomic",
	"@c2:min instance of c2:smallInteger",
	// Functions (prefixed, unprefixed, nested, no-arg), comments.
	"string() = '' and empty(..)",
	"fn:node-name(.) = node-name(.) (: tests default function namespace :)",
	"resolve-QName(@kind,.)=xs:QName('xs:date')",
	"namespace-uri-from-QName(xs:QName('p:ppp')) = 'http://cta023.com/p'",
	"((2 + position()) instance of xs:integer) (: trailing comment :)",
	"empty(.//comment())",
	"chess:result = ('black wins', 'white wins', 'draw')",
	// Syntactically valid even if semantically dubious (only syntax is checked).
	"()",
	"string cast as string",
	"double('3' cast as float > 2)",
	"@kind cast as messageTypeString='string'",
}

// invalid lists expressions that contain a static (syntax) error: every one
// must be rejected. These are the malformed {test} expressions from the suite's
// schema-invalid type-alternative cases (ibm S3_12 si04/si05/si06).
var invalid = []string{
	`3 cast as "3" ?`,         // cast target must be a QName, not a string
	`((7>=6)`,                 // unbalanced parentheses
	`@6='hi'`,                 // node test cannot be a number
	`@a:kind 's' 'a'`,         // juxtaposed operands
	`@a:kind 1 2`,             // juxtaposed operands
	`@a:kind 1 <`,             // trailing comparison operator
	`12 's' 'u'`,              // juxtaposed operands
	`12 5 2`,                  // juxtaposed operands
	`3 cast as 3`,             // cast target must be a QName, not a number
	`string cast as 'string'`, // cast target must be a QName, not a string
	`cast as decimal 3`,       // missing cast operand
	`3 cast as @a:kind > 1`,   // cast target must be a QName, not "@..."
	`6 > cast as decimal`,     // "cast" with no operand
	`3 cast 'as' decimal`,     // "as" keyword expected, found a string
	`)(`,                      // not an expression
	`>`,                       // not an expression
	`@numberOfChildren < 5 AND @numberOfChildren > 0`, // "AND" is not "and"
	`(@type cast as xs1::double)='double'`,            // "::" in a cast target QName
}

func TestParseValid(t *testing.T) {
	for _, src := range valid {
		if _, err := Parse(src); err != nil {
			t.Errorf("Parse(%q) returned error %v, want nil", src, err)
		}
	}
}

func TestParseInvalid(t *testing.T) {
	for _, src := range invalid {
		if _, err := Parse(src); err == nil {
			t.Errorf("Parse(%q) returned nil error, want a syntax error", src)
		}
	}
}

func TestTypeRefs(t *testing.T) {
	tests := []struct {
		src    string
		prefix string
		local  string
		kind   TypeRefKind
	}{
		{"@kind cast as messageTypeString='string'", "", "messageTypeString", Cast},
		{"$value castable as xs:date", "xs", "date", Castable},
		{"@c2:min instance of c2:smallInteger", "c2", "smallInteger", InstanceOf},
		{"$v treat as my:t", "my", "t", Treat},
	}
	for _, tt := range tests {
		e, err := Parse(tt.src)
		if err != nil {
			t.Errorf("Parse(%q): unexpected error %v", tt.src, err)
			continue
		}
		if len(e.TypeRefs) != 1 {
			t.Errorf("Parse(%q): got %d TypeRefs, want 1: %v", tt.src, len(e.TypeRefs), e.TypeRefs)
			continue
		}
		r := e.TypeRefs[0]
		if r.Prefix != tt.prefix || r.Local != tt.local || r.Kind != tt.kind {
			t.Errorf("Parse(%q): TypeRef = %+v, want {%q %q %v}", tt.src, r, tt.prefix, tt.local, tt.kind)
		}
	}
}

// Kind-test type annotations (element(*, T), attribute(*, T)) are NOT atomic
// cast/instance-of targets and must not be reported as TypeRefs, lest a valid
// element(*, xs:untyped) be mistaken for a cast to a non-atomic type.
func TestKindTestTypeNamesNotRefs(t *testing.T) {
	for _, src := range []string{
		". instance of element(*, xs:untyped)",
		"@type instance of attribute(*, xs:untypedAtomic)",
		". instance of element(*, foo:bar)?",
	} {
		e, err := Parse(src)
		if err != nil {
			t.Errorf("Parse(%q): unexpected error %v", src, err)
			continue
		}
		if len(e.TypeRefs) != 0 {
			t.Errorf("Parse(%q): got TypeRefs %v, want none", src, e.TypeRefs)
		}
	}
}
