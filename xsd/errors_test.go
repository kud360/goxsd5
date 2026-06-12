package xsd

import (
	"strings"
	"testing"
)

func TestErrorFormat(t *testing.T) {
	e := NewError(SpecMinLengthValid, Pos{URI: "a.xsd", Line: 3, Column: 7}, "too short: %d < %d", 2, 5)
	got := e.Error()
	want := "a.xsd:3:7: [cvc-minLength-valid] too short: 2 < 5"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestErrorSecondaryPos(t *testing.T) {
	e := &Error{
		Pos:      Pos{URI: "a.xsd", Line: 9, Column: 1},
		OtherPos: Pos{URI: "b.xsd", Line: 2, Column: 4},
		Msg:      "duplicate definition of foo",
	}
	got := e.Error()
	if !strings.Contains(got, "see also b.xsd:2:4") {
		t.Errorf("missing secondary pos: %q", got)
	}
}

func TestErrorListAndRefIDs(t *testing.T) {
	var l ErrorList
	if l.Err() != nil {
		t.Fatal("empty list should yield nil")
	}
	l.Addf(SpecPatternValid, Pos{URI: "x", Line: 1, Column: 1}, "bad pattern")
	l.Addf(SpecPatternValid, Pos{URI: "x", Line: 2, Column: 1}, "bad pattern again")
	l.Addf(SpecEnumerationValid, Pos{URI: "x", Line: 3, Column: 1}, "not in enum")
	err := l.Err()
	if err == nil {
		t.Fatal("expected error")
	}
	if n := len(AllErrors(err)); n != 3 {
		t.Errorf("AllErrors len = %d, want 3", n)
	}
	ids := RefIDs(err)
	if len(ids) != 2 || ids[0] != "cvc-pattern-valid" || ids[1] != "cvc-enumeration-valid" {
		t.Errorf("RefIDs = %v", ids)
	}
}

func TestRefsRegistry(t *testing.T) {
	r, ok := Refs["cvc-minLength-valid"]
	if !ok || r.Part != 2 {
		t.Errorf("registry missing cvc-minLength-valid: %+v", r)
	}
}
