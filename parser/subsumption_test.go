package parser

// Track-B differential tests: the unified particle-restriction relation
// (particleRestrictUnified) must agree with the legacy fragment relation
// (checkParticleRestrictLegacy) on every restriction the analysis sees. Two
// oracles drive the comparison: the full W3C conformance corpus (every
// complexContent restriction the suite builds), and a structural fuzzer over
// random flat element-run pairs. The invariant is exact agreement on the set of
// findings, so any divergence — a ported case that decides differently, or a
// coverage regression — fails loudly. These run with restrictDiff installed; it
// is nil in production.

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"testing"

	"github.com/kud360/goxsd5/xsd"
)

// vkey renders a finding as a stable comparison key (spec id + position +
// message); the multiset of keys is what must match between the two relations.
func vkey(v restrictViolation) string {
	return v.ref.ID + "|" + v.pos.String() + "|" + v.msg
}

// diffViolations returns a human-readable description of how two finding sets
// differ, or "" when they are equal as multisets.
func diffViolations(legacy, unified []restrictViolation) string {
	lk := make([]string, len(legacy))
	uk := make([]string, len(unified))
	for i, v := range legacy {
		lk[i] = vkey(v)
	}
	for i, v := range unified {
		uk[i] = vkey(v)
	}
	sort.Strings(lk)
	sort.Strings(uk)
	if fmt.Sprint(lk) == fmt.Sprint(uk) {
		return ""
	}
	return fmt.Sprintf("legacy=%v unified=%v", lk, uk)
}

// TestRestrictDifferentialSuite installs the differential hook and rebuilds the
// entire conformance corpus, asserting the unified and legacy relations agree on
// every complexContent restriction.
func TestRestrictDifferentialSuite(t *testing.T) {
	if _, err := os.Stat(suiteRoot); err != nil {
		t.Skipf("W3C suite not checked out (%v); run testdata/fetch-xsdtests.sh", err)
	}
	cases := collectSuiteCases(t)
	if len(cases) == 0 {
		t.Fatal("no XSD 1.1 schema cases found under " + suiteRoot)
	}

	var mismatches []string
	var compared int
	restrictDiff = func(ct *xsd.ComplexType, legacy, unified []restrictViolation) {
		compared++
		if d := diffViolations(legacy, unified); d != "" {
			mismatches = append(mismatches, describeCT(ct)+": "+d)
		}
	}
	defer func() { restrictDiff = nil }()

	for _, c := range cases {
		runSuiteCase(c)
	}

	if len(mismatches) > 0 {
		sort.Strings(mismatches)
		if len(mismatches) > 40 {
			mismatches = mismatches[:40]
		}
		t.Fatalf("%d/%d restrictions diverge between legacy and unified relations:\n%s",
			len(mismatches), compared, joinLines(mismatches))
	}
	t.Logf("compared %d restrictions across the suite; legacy and unified agree", compared)
}

func joinLines(ss []string) string {
	out := ""
	for _, s := range ss {
		out += "  " + s + "\n"
	}
	return out
}

// TestRestrictDifferentialFuzz compares the two relations on random flat
// element-run restriction pairs (the slice the unified relation handles
// natively), so the transcription is exercised on shapes the suite never names.
func TestRestrictDifferentialFuzz(t *testing.T) {
	b := &builder{errs: &xsd.ErrorList{}}

	// A small derivation chain so type-derivation checks have both valid and
	// invalid pairings to find: ct0 <- ct1 <- ct2 by restriction, plus an
	// unrelated ctx.
	ct0 := &xsd.ComplexType{Name: xsd.QName{Local: "T0"}}
	ct1 := &xsd.ComplexType{Name: xsd.QName{Local: "T1"}, BaseType: ct0, DerivationMethod: xsd.DeriveRestriction}
	ct2 := &xsd.ComplexType{Name: xsd.QName{Local: "T2"}, BaseType: ct1, DerivationMethod: xsd.DeriveRestriction}
	ctx := &xsd.ComplexType{Name: xsd.QName{Local: "TX"}}
	types := []xsd.Type{ct0, ct1, ct2, ctx}
	names := []xsd.QName{{Local: "a"}, {Local: "b"}, {Local: "c"}, {Local: "d"}}

	// accepted with no substitution groups: an element particle accepts only its
	// own name (the suite test covers the substitution-group routing).
	accepted := func(e *xsd.ElementDecl) map[xsd.QName]bool {
		return map[xsd.QName]bool{e.Name: true}
	}
	globals := map[xsd.QName]*xsd.ElementDecl{}

	rng := rand.New(rand.NewSource(1))
	occ := func() (int, int) {
		min := rng.Intn(3)
		switch rng.Intn(3) {
		case 0:
			return min, min + rng.Intn(3)
		case 1:
			return min, xsd.UnboundedOccurs
		default:
			return min, min
		}
	}
	run := func(compositor xsd.Compositor) *xsd.Particle {
		k := 1 + rng.Intn(4)
		var parts []*xsd.Particle
		for i := 0; i < k; i++ {
			min, max := occ()
			e := &xsd.ElementDecl{
				Name:     names[rng.Intn(len(names))],
				Type:     types[rng.Intn(len(types))],
				Nillable: rng.Intn(4) == 0,
			}
			parts = append(parts, &xsd.Particle{MinOccurs: min, MaxOccurs: max, Term: e})
		}
		tmin, tmax := occ()
		if tmax == 0 {
			tmax = 1
		}
		return &xsd.Particle{MinOccurs: tmin, MaxOccurs: tmax, Term: &xsd.ModelGroup{Compositor: compositor, Particles: parts}}
	}
	pick := func() xsd.Compositor {
		if rng.Intn(2) == 0 {
			return xsd.CompositorSequence
		}
		return xsd.CompositorAll
	}

	mismatches := 0
	for iter := 0; iter < 20000; iter++ {
		base := &xsd.ComplexType{Name: xsd.QName{Local: "B"}, Content: &xsd.ElementContent{Particle: run(pick())}}
		r := &xsd.ComplexType{
			Name:             xsd.QName{Local: "R"},
			BaseType:         base,
			DerivationMethod: xsd.DeriveRestriction,
			Content:          &xsd.ElementContent{Particle: run(pick())},
		}
		legacy := b.collectLegacyRestrict(r, accepted, globals)
		unified := b.particleRestrictUnified(r, accepted, globals)
		if d := diffViolations(legacy, unified); d != "" {
			if mismatches < 10 {
				t.Errorf("iter %d diverges: %s", iter, d)
			}
			mismatches++
		}
	}
	if mismatches > 0 {
		t.Fatalf("%d fuzz restriction pairs diverged between legacy and unified relations", mismatches)
	}
}
