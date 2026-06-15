package xsdwalk

import (
	"testing"

	"github.com/kud360/goxsd5/xsd"
)

func qn(local string) xsd.QName { return xsd.QName{Local: local} }

func elem(local string) *xsd.ElementDecl {
	return &xsd.ElementDecl{Name: qn(local)}
}

// part wraps a term in a particle with the given occurrence range.
func part(min, max int, t xsd.Term) *xsd.Particle {
	return &xsd.Particle{MinOccurs: min, MaxOccurs: max, Term: t}
}

func seqGroup(ps ...*xsd.Particle) *xsd.ModelGroup {
	return &xsd.ModelGroup{Compositor: xsd.CompositorSequence, Particles: ps}
}
func choiceGroup(ps ...*xsd.Particle) *xsd.ModelGroup {
	return &xsd.ModelGroup{Compositor: xsd.CompositorChoice, Particles: ps}
}
func allGroup(ps ...*xsd.Particle) *xsd.ModelGroup {
	return &xsd.ModelGroup{Compositor: xsd.CompositorAll, Particles: ps}
}

func names(locals ...string) []xsd.QName {
	out := make([]xsd.QName, len(locals))
	for i, l := range locals {
		out[i] = qn(l)
	}
	return out
}

func TestMatcher(t *testing.T) {
	m := &Matcher{}

	// sequence (a, b)
	abSeq := part(1, 1, seqGroup(part(1, 1, elem("a")), part(1, 1, elem("b"))))
	// sequence (a, b?, c*)
	flexSeq := part(1, 1, seqGroup(
		part(1, 1, elem("a")),
		part(0, 1, elem("b")),
		part(0, xsd.UnboundedOccurs, elem("c")),
	))
	// choice (a | b), 1..2 times
	choice := part(1, 2, choiceGroup(part(1, 1, elem("a")), part(1, 1, elem("b"))))
	// all { a, b? }
	allg := part(1, 1, allGroup(part(1, 1, elem("a")), part(0, 1, elem("b"))))
	// empty sequence, required
	empty := part(1, 1, seqGroup())

	cases := []struct {
		name string
		p    *xsd.Particle
		in   []xsd.QName
		want bool
	}{
		{"seq ab ok", abSeq, names("a", "b"), true},
		{"seq ab missing b", abSeq, names("a"), false},
		{"seq ab extra", abSeq, names("a", "b", "c"), false},
		{"seq ab wrong order", abSeq, names("b", "a"), false},
		{"flex a", flexSeq, names("a"), true},
		{"flex a b", flexSeq, names("a", "b"), true},
		{"flex a c c", flexSeq, names("a", "c", "c"), true},
		{"flex a b c c", flexSeq, names("a", "b", "c", "c"), true},
		{"flex missing a", flexSeq, names("c"), false},
		{"choice a", choice, names("a"), true},
		{"choice a b", choice, names("a", "b"), true},
		{"choice three", choice, names("a", "b", "a"), false},
		{"choice none", choice, nil, false},
		{"all ab", allg, names("a", "b"), true},
		{"all ba", allg, names("b", "a"), true},
		{"all a", allg, names("a"), true},
		{"all b only", allg, names("b"), false}, // a required
		{"all dup a", allg, names("a", "a"), false},
		{"empty ok", empty, nil, true},
		{"empty with child", empty, names("a"), false},
	}
	for _, c := range cases {
		_, ok := m.Match(c.p, c.in, nil)
		if ok != c.want {
			t.Errorf("%s: Match=%v want %v", c.name, ok, c.want)
		}
	}
}

func TestMatcherDefinedSiblingWildcard(t *testing.T) {
	// Model: sequence(s, n, any[notQName=##definedSibling]*), with s1
	// substitutable for s. The trailing wildcard must exclude s, n, and s1
	// (definedSibling + substitutables) but admit unrelated names like n1/x.
	s := &xsd.ElementDecl{Name: qn("s"), Global: true}
	n := &xsd.ElementDecl{Name: qn("n"), Global: true}
	s1 := &xsd.ElementDecl{Name: qn("s1"), Global: true, SubstitutionGroups: []*xsd.ElementDecl{s}}
	n1 := &xsd.ElementDecl{Name: qn("n1"), Global: true}
	globals := map[xsd.QName]*xsd.ElementDecl{
		qn("s"): s, qn("n"): n, qn("s1"): s1, qn("n1"): n1,
	}
	m := &Matcher{LookupGlobal: func(q xsd.QName) *xsd.ElementDecl { return globals[q] }}
	wc := &xsd.Wildcard{Mode: xsd.NSConstraintAny, ProcessContents: xsd.ProcessLax,
		NotQName: []xsd.QName{{Local: "##definedSibling"}}}
	p := part(1, 1, seqGroup(part(1, 1, s), part(1, 1, n), part(0, xsd.UnboundedOccurs, wc)))

	cases := []struct {
		name string
		in   []xsd.QName
		want bool
	}{
		{"trailing n1 allowed", names("s", "n", "n1"), true},
		{"trailing x allowed", names("s", "n", "x"), true},
		{"trailing s excluded (sibling)", names("s", "n", "s"), false},
		{"trailing n excluded (sibling)", names("s", "n", "n"), false},
		{"trailing s1 excluded (substitutable for s)", names("s", "n", "s1"), false},
	}
	for _, c := range cases {
		if _, ok := m.Match(p, c.in, nil); ok != c.want {
			t.Errorf("%s: Match=%v want %v", c.name, ok, c.want)
		}
	}
}

func TestMatcherWildcardAndTerms(t *testing.T) {
	m := &Matcher{}
	// sequence (a, any*)
	wc := &xsd.Wildcard{Mode: xsd.NSConstraintAny, ProcessContents: xsd.ProcessLax}
	p := part(1, 1, seqGroup(part(1, 1, elem("a")), part(0, xsd.UnboundedOccurs, wc)))

	terms, ok := m.Match(p, names("a", "x", "y"), nil)
	if !ok {
		t.Fatal("expected match")
	}
	if terms[0].Elem == nil || terms[0].Elem.Name.Local != "a" {
		t.Errorf("child 0 should match element a, got %+v", terms[0])
	}
	if terms[1].Wildcard != wc || terms[2].Wildcard != wc {
		t.Errorf("children 1,2 should match the wildcard")
	}
}
