package xsdwalk

import (
	"testing"

	"github.com/kud360/goxsd5/xsd"
)

// ct wraps a root content-model particle in a complex type with element content.
func ct(root *xsd.Particle) *xsd.ComplexType {
	return &xsd.ComplexType{Content: &xsd.ElementContent{Particle: root}}
}

// visit records each particle's term description and depth, walking with a
// predicate that decides descent.
func visit(t *xsd.ComplexType, descend func(*xsd.Particle, int) bool) ([]string, []int) {
	var terms []string
	var depths []int
	NewWalker().Walk(t, func(p *xsd.Particle, depth int) bool {
		terms = append(terms, termDesc(p.Term))
		depths = append(depths, depth)
		return descend(p, depth)
	})
	return terms, depths
}

func termDesc(t xsd.Term) string {
	switch t := t.(type) {
	case *xsd.ElementDecl:
		return "elem:" + t.Name.Local
	case *xsd.ModelGroup:
		return "group:" + t.Compositor.String()
	case *xsd.GroupRef:
		if t.Ref != nil {
			return "ref:" + t.Ref.Name.Local
		}
		return "ref:<nil>"
	case *xsd.Wildcard:
		return "any"
	}
	return "?"
}

func always(*xsd.Particle, int) bool { return true }

func elemNames(decls []*xsd.ElementDecl) []string {
	out := make([]string, len(decls))
	for i, d := range decls {
		out[i] = d.Name.Local
	}
	return out
}

func eqStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func eqInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestWalkFlatSequence(t *testing.T) {
	// sequence (a, b, c)
	root := part(1, 1, seqGroup(part(1, 1, elem("a")), part(1, 1, elem("b")), part(1, 1, elem("c"))))
	terms, depths := visit(ct(root), always)

	wantTerms := []string{"group:sequence", "elem:a", "elem:b", "elem:c"}
	wantDepths := []int{0, 1, 1, 1}
	if !eqStrings(terms, wantTerms) {
		t.Errorf("terms = %v, want %v", terms, wantTerms)
	}
	if !eqInts(depths, wantDepths) {
		t.Errorf("depths = %v, want %v", depths, wantDepths)
	}

	if got := elemNames(Elements(ct(root))); !eqStrings(got, []string{"a", "b", "c"}) {
		t.Errorf("Elements = %v, want [a b c]", got)
	}
}

func TestWalkNestedChoice(t *testing.T) {
	// sequence (a, choice(b, sequence(c, d)))
	inner := part(1, 1, seqGroup(part(1, 1, elem("c")), part(1, 1, elem("d"))))
	root := part(1, 1, seqGroup(
		part(1, 1, elem("a")),
		part(1, 1, choiceGroup(part(1, 1, elem("b")), inner)),
	))
	terms, depths := visit(ct(root), always)

	wantTerms := []string{"group:sequence", "elem:a", "group:choice", "elem:b", "group:sequence", "elem:c", "elem:d"}
	wantDepths := []int{0, 1, 1, 2, 2, 3, 3}
	if !eqStrings(terms, wantTerms) {
		t.Errorf("terms = %v, want %v", terms, wantTerms)
	}
	if !eqInts(depths, wantDepths) {
		t.Errorf("depths = %v, want %v", depths, wantDepths)
	}

	if got := elemNames(Elements(ct(root))); !eqStrings(got, []string{"a", "b", "c", "d"}) {
		t.Errorf("Elements = %v, want [a b c d]", got)
	}
}

func TestWalkPruneSubtree(t *testing.T) {
	// Returning false on the choice compositor must skip its children.
	root := part(1, 1, seqGroup(
		part(1, 1, elem("a")),
		part(1, 1, choiceGroup(part(1, 1, elem("b")), part(1, 1, elem("c")))),
		part(1, 1, elem("d")),
	))
	terms, _ := visit(ct(root), func(p *xsd.Particle, _ int) bool {
		g, ok := p.Term.(*xsd.ModelGroup)
		return !ok || g.Compositor != xsd.CompositorChoice
	})
	wantTerms := []string{"group:sequence", "elem:a", "group:choice", "elem:d"}
	if !eqStrings(terms, wantTerms) {
		t.Errorf("terms = %v, want %v", terms, wantTerms)
	}
}

func TestWalkGroupRef(t *testing.T) {
	// sequence (a, group-ref g{ sequence(b, c) }, d)
	g := &xsd.Group{Name: qn("g"), Group: seqGroup(part(1, 1, elem("b")), part(1, 1, elem("c")))}
	root := part(1, 1, seqGroup(
		part(1, 1, elem("a")),
		part(1, 1, &xsd.GroupRef{Ref: g}),
		part(1, 1, elem("d")),
	))
	terms, depths := visit(ct(root), always)

	// A group reference stands in for the group: descending it visits the
	// referenced group's particles directly (the bare ModelGroup compositor has
	// no enclosing particle of its own, mirroring how the Matcher enters it).
	wantTerms := []string{"group:sequence", "elem:a", "ref:g", "elem:b", "elem:c", "elem:d"}
	wantDepths := []int{0, 1, 1, 2, 2, 1}
	if !eqStrings(terms, wantTerms) {
		t.Errorf("terms = %v, want %v", terms, wantTerms)
	}
	if !eqInts(depths, wantDepths) {
		t.Errorf("depths = %v, want %v", depths, wantDepths)
	}

	if got := elemNames(Elements(ct(root))); !eqStrings(got, []string{"a", "b", "c", "d"}) {
		t.Errorf("Elements = %v, want [a b c d]", got)
	}
}

func TestWalkWildcard(t *testing.T) {
	// sequence (a, any)
	wc := &xsd.Wildcard{Mode: xsd.NSConstraintAny, ProcessContents: xsd.ProcessLax}
	root := part(1, 1, seqGroup(part(1, 1, elem("a")), part(0, xsd.UnboundedOccurs, wc)))
	terms, _ := visit(ct(root), always)

	wantTerms := []string{"group:sequence", "elem:a", "any"}
	if !eqStrings(terms, wantTerms) {
		t.Errorf("terms = %v, want %v", terms, wantTerms)
	}
	// A wildcard term is not an element declaration, so Elements omits it.
	if got := elemNames(Elements(ct(root))); !eqStrings(got, []string{"a"}) {
		t.Errorf("Elements = %v, want [a]", got)
	}
}

func TestWalkAllCompositor(t *testing.T) {
	// The `all` compositor is walked through the same walkGroup path as
	// sequence/choice: its particles are visited as direct children, in order.
	root := part(1, 1, allGroup(part(1, 1, elem("a")), part(0, 1, elem("b")), part(1, 1, elem("c"))))
	terms, depths := visit(ct(root), always)

	wantTerms := []string{"group:all", "elem:a", "elem:b", "elem:c"}
	wantDepths := []int{0, 1, 1, 1}
	if !eqStrings(terms, wantTerms) {
		t.Errorf("terms = %v, want %v", terms, wantTerms)
	}
	if !eqInts(depths, wantDepths) {
		t.Errorf("depths = %v, want %v", depths, wantDepths)
	}

	if got := elemNames(Elements(ct(root))); !eqStrings(got, []string{"a", "b", "c"}) {
		t.Errorf("Elements = %v, want [a b c]", got)
	}
}

func TestWalkSiblingGroupRefNotOverPruned(t *testing.T) {
	// A non-recursive named group referenced twice in disjoint sibling branches
	// must be walked BOTH times: the cycle guard is path/stack-scoped (deleted on
	// exit), not a global visited-set. choice( ref:g, ref:g ) with g{ (x, y) }.
	g := &xsd.Group{Name: qn("g"), Group: seqGroup(part(1, 1, elem("x")), part(1, 1, elem("y")))}
	root := part(1, 1, choiceGroup(
		part(1, 1, &xsd.GroupRef{Ref: g}),
		part(1, 1, &xsd.GroupRef{Ref: g}),
	))
	terms, _ := visit(ct(root), always)

	// Both sibling references expand: a global visited-set would prune the second.
	wantTerms := []string{"group:choice", "ref:g", "elem:x", "elem:y", "ref:g", "elem:x", "elem:y"}
	if !eqStrings(terms, wantTerms) {
		t.Errorf("terms = %v, want %v", terms, wantTerms)
	}
	if got := elemNames(Elements(ct(root))); !eqStrings(got, []string{"x", "y", "x", "y"}) {
		t.Errorf("Elements = %v, want [x y x y]", got)
	}
}

func TestWalkRecursiveGroupRefTerminates(t *testing.T) {
	// group g{ sequence(a, group-ref g) } referenced from the type: the self
	// reference must be detected and skipped, not looped forever.
	g := &xsd.Group{Name: qn("g")}
	g.Group = seqGroup(part(1, 1, elem("a")), part(1, 1, &xsd.GroupRef{Ref: g}))
	root := part(1, 1, &xsd.GroupRef{Ref: g})

	terms, _ := visit(ct(root), always)
	// Outer ref:g -> its sequence's particles (elem:a, inner ref:g); the inner
	// ref:g is visited but its re-entry is skipped by the cycle guard.
	wantTerms := []string{"ref:g", "elem:a", "ref:g"}
	if !eqStrings(terms, wantTerms) {
		t.Errorf("terms = %v, want %v", terms, wantTerms)
	}
	if got := elemNames(Elements(ct(root))); !eqStrings(got, []string{"a"}) {
		t.Errorf("Elements = %v, want [a]", got)
	}
}

func TestWalkReusableAcrossCalls(t *testing.T) {
	// One Walker reused: the second walk of a recursive model must behave
	// identically (cycle-guard state is reset per Walk).
	g := &xsd.Group{Name: qn("g")}
	g.Group = seqGroup(part(1, 1, elem("a")), part(1, 1, &xsd.GroupRef{Ref: g}))
	t1 := ct(part(1, 1, &xsd.GroupRef{Ref: g}))

	w := NewWalker()
	var first, second []string
	collect := func(out *[]string) ParticleFunc {
		return func(p *xsd.Particle, _ int) bool {
			*out = append(*out, termDesc(p.Term))
			return true
		}
	}
	w.Walk(t1, collect(&first))
	w.Walk(t1, collect(&second))
	if !eqStrings(first, second) {
		t.Errorf("reuse mismatch: first=%v second=%v", first, second)
	}
}

func TestWalkNonElementContent(t *testing.T) {
	// Simple, empty, and absent content yield no particles.
	cases := []*xsd.ComplexType{
		{Content: &xsd.SimpleContent{}},
		{Content: &xsd.ElementContent{Particle: nil}},
		{Content: nil},
		nil,
	}
	for i, c := range cases {
		if got := Elements(c); got != nil {
			t.Errorf("case %d: Elements = %v, want nil", i, got)
		}
		var n int
		NewWalker().Walk(c, func(*xsd.Particle, int) bool { n++; return true })
		if n != 0 {
			t.Errorf("case %d: visited %d particles, want 0", i, n)
		}
	}
}
