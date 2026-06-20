package gotype_test

import (
	"testing"
	"time"

	"github.com/kud360/goxsd5/builtin"
	"github.com/kud360/goxsd5/builtin/gotype"
	"github.com/kud360/goxsd5/xsd"
	"github.com/kud360/goxsd5/xsdedit"
)

// TestAllBuiltinsMirrorsStrict asserts the lax set has exactly the same member
// names as the strict set — same types, lax semantics.
func TestAllBuiltinsMirrorsStrict(t *testing.T) {
	lax := names(gotype.AllBuiltins())
	strict := map[xsd.QName]bool{}
	for _, st := range builtin.AllBuiltins() {
		strict[st.Name] = true
	}
	if len(lax) != len(strict) {
		t.Errorf("lax set has %d types, strict has %d", len(lax), len(strict))
	}
	for q := range strict {
		if !lax[q] {
			t.Errorf("lax set missing %s", q)
		}
	}
	for q := range lax {
		if !strict[q] {
			t.Errorf("lax set has unexpected %s", q)
		}
	}
}

func names(ts []*xsd.SimpleType) map[xsd.QName]bool {
	m := make(map[xsd.QName]bool, len(ts))
	for _, t := range ts {
		m[t.Name] = true
	}
	return m
}

// TestLaxValueKinds asserts each primitive lands its lexical on the Go-native
// value the lax space promises.
func TestLaxValueKinds(t *testing.T) {
	byName := index(gotype.AllBuiltins())
	cases := []struct {
		typ     string
		lexical string
		want    any // a sample of the want type, compared by type only
	}{
		{"string", "hi", "x"},
		{"boolean", "true", true},
		{"decimal", "1.5", float64(0)},
		{"double", "1e3", float64(0)},
		{"integer", "42", int64(0)},
		{"int", "7", int64(0)},
		{"dateTime", "2001-02-03T04:05:06", time.Time{}},
		{"date", "2001-02-03", time.Time{}},
		{"duration", "PT1H", time.Duration(0)},
		{"hexBinary", "deadbeef", []byte(nil)},
	}
	for _, c := range cases {
		st := byName[c.typ]
		if st == nil {
			t.Fatalf("no lax type %s", c.typ)
		}
		v, err := st.ParseValue(c.lexical, nil)
		if err != nil {
			t.Errorf("%s.ParseValue(%q): %v", c.typ, c.lexical, err)
			continue
		}
		if !sameKind(v, c.want) {
			t.Errorf("%s.ParseValue(%q) is %T, want %T", c.typ, c.lexical, v, c.want)
		}
	}
}

func sameKind(a, b any) bool {
	switch b.(type) {
	case string:
		// strVal is unexported; assert it stringifies and is not one of the others.
		_, isBool := a.(bool)
		_, isF := a.(float64)
		_, isI := a.(int64)
		return !isBool && !isF && !isI
	case bool:
		_, ok := a.(bool)
		return ok
	case float64:
		_, ok := a.(float64)
		return ok
	case int64:
		_, ok := a.(int64)
		return ok
	case time.Time:
		_, ok := a.(time.Time)
		return ok
	case time.Duration:
		_, ok := a.(time.Duration)
		return ok
	case []byte:
		_, isBin := a.([]byte)
		if isBin {
			return true
		}
		// binVal is unexported; accept any non-primitive, length-bearing value.
		_, ok := a.(xsd.Lengthed)
		_, isStr := a.(string)
		return ok && !isStr
	}
	return false
}

func index(ts []*xsd.SimpleType) map[string]*xsd.SimpleType {
	m := make(map[string]*xsd.SimpleType, len(ts))
	for _, t := range ts {
		m[t.Name.Local] = t
	}
	return m
}

// TestLaxBoundsOrder asserts the integer ladder's bounds are enforced through
// the lax int64 comparator: an in-range value validates, an out-of-range one is
// rejected.
func TestLaxBoundsOrder(t *testing.T) {
	byte8 := index(gotype.AllBuiltins())["byte"]
	if _, err := byte8.ParseValue("127", nil); err != nil {
		t.Errorf("byte rejected in-range 127: %v", err)
	}
	if _, err := byte8.ParseValue("200", nil); err == nil {
		t.Error("byte accepted out-of-range 200")
	}
}

// TestReplaceMixesStrictAndLax asserts Replace swaps named entries and leaves
// the rest strict, returning a same-membership set.
func TestReplaceMixesStrictAndLax(t *testing.T) {
	base := builtin.AllBuiltins()
	mixed := gotype.Replace(base, gotype.Double, gotype.Integer)
	if len(mixed) != len(base) {
		t.Fatalf("Replace changed membership: %d vs %d", len(mixed), len(base))
	}
	byQName := func(ts []*xsd.SimpleType) map[xsd.QName]*xsd.SimpleType {
		m := map[xsd.QName]*xsd.SimpleType{}
		for _, t := range ts {
			m[t.Name] = t
		}
		return m
	}
	got := byQName(mixed)
	if got[gotype.Double.Name] != gotype.Double {
		t.Error("Replace did not substitute the lax double")
	}
	if got[gotype.Integer.Name] != gotype.Integer {
		t.Error("Replace did not substitute the lax integer")
	}
	// A type not in the lax list stays the strict instance.
	strict := byQName(base)
	stringQ := xsd.QName{Namespace: xsd.XSDNS, Local: "string"}
	if got[stringQ] != strict[stringQ] {
		t.Error("Replace altered an untouched (string) entry")
	}
	// base must be untouched.
	if byQName(base)[gotype.Double.Name] == gotype.Double {
		t.Error("Replace mutated its base argument")
	}
}

// TestPrimitivesPassValidate asserts every lax primitive is a well-formed
// primitive per xsdedit.Validate (Applicable != 0, own Parse, sound facets).
func TestPrimitivesPassValidate(t *testing.T) {
	for _, st := range gotype.AllBuiltins() {
		if err := xsdedit.Validate(st); err != nil {
			t.Errorf("Validate(%s): %v", st.Name, err)
		}
	}
}
