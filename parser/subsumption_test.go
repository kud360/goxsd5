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

// verdict reduces a finding set to what the conformance ratchet cares about: is
// the restriction rejected at all (any violation), and under which spec
// constraints (the set of distinct ref ids). A genuine unification of the
// fragments reaches the same verdict but may phrase or position its diagnostics
// differently, so messages and positions are deliberately not compared.
func verdict(vs []restrictViolation) (invalid bool, refs []string) {
	seen := map[string]bool{}
	for _, v := range vs {
		if !seen[v.ref.ID] {
			seen[v.ref.ID] = true
			refs = append(refs, v.ref.ID)
		}
	}
	sort.Strings(refs)
	return len(vs) > 0, refs
}

// diffViolations returns a human-readable description of how two finding sets
// disagree on the verdict (rejected-or-not, and which spec constraints), or ""
// when they agree.
func diffViolations(legacy, unified []restrictViolation) string {
	li, lr := verdict(legacy)
	ui, ur := verdict(unified)
	if li == ui && fmt.Sprint(lr) == fmt.Sprint(ur) {
		return ""
	}
	return fmt.Sprintf("legacy={invalid:%v refs:%v} unified={invalid:%v refs:%v}", li, lr, ui, ur)
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

// renderParticle renders a flat particle for divergence diagnostics.
func renderParticle(p *xsd.Particle) string {
	occ := func(p *xsd.Particle) string {
		mx := fmt.Sprint(p.MaxOccurs)
		if p.MaxOccurs == xsd.UnboundedOccurs {
			mx = "*"
		}
		return fmt.Sprintf("[%d,%s]", p.MinOccurs, mx)
	}
	mg, ok := p.Term.(*xsd.ModelGroup)
	if !ok {
		return "?"
	}
	out := mg.Compositor.String() + occ(p) + "{"
	for i, c := range mg.Particles {
		if i > 0 {
			out += " "
		}
		switch t := c.Term.(type) {
		case *xsd.ElementDecl:
			out += fmt.Sprintf("%s:%s%s", t.Name, t.Type.TypeName().Local, occ(c))
		case *xsd.Wildcard:
			out += fmt.Sprintf("any(m%d,ns%v,pc%d)%s", t.Mode, t.Namespaces, t.ProcessContents, occ(c))
		}
	}
	return out + "}"
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
	// Element names spread over the same namespaces the wildcards range over, so
	// wildcard/element overlap actually occurs.
	nss := []string{"", "n1", "n2"}
	var names []xsd.QName
	for _, ns := range nss {
		for _, l := range []string{"a", "b", "c"} {
			names = append(names, xsd.QName{Namespace: ns, Local: l})
		}
	}

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
	// A small pool of wildcards over the fuzz namespaces, plus the occasional
	// "no wildcard" so most particles are named elements.
	wildcard := func() *xsd.Wildcard {
		w := &xsd.Wildcard{ProcessContents: xsd.ProcessContents(rng.Intn(3))}
		switch rng.Intn(3) {
		case 0:
			w.Mode = xsd.NSConstraintAny
		case 1:
			w.Mode = xsd.NSConstraintEnumeration
			w.Namespaces = []string{nss[rng.Intn(len(nss))]}
		default:
			w.Mode = xsd.NSConstraintNot
			w.Namespaces = []string{nss[rng.Intn(len(nss))]}
		}
		return w
	}
	run := func(compositor xsd.Compositor) *xsd.Particle {
		k := 1 + rng.Intn(4)
		var parts []*xsd.Particle
		for i := 0; i < k; i++ {
			min, max := occ()
			var term xsd.Term
			if rng.Intn(4) == 0 {
				term = wildcard()
			} else {
				term = &xsd.ElementDecl{
					Name:     names[rng.Intn(len(names))],
					Type:     types[rng.Intn(len(types))],
					Nillable: rng.Intn(4) == 0,
				}
			}
			parts = append(parts, &xsd.Particle{MinOccurs: min, MaxOccurs: max, Term: term})
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
	for iter := 0; iter < 100000; iter++ {
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
			if mismatches < 5 {
				t.Errorf("iter %d diverges: %s\n  base=%s\n  restr=%s", iter, d,
					renderParticle(base.Content.(*xsd.ElementContent).Particle),
					renderParticle(r.Content.(*xsd.ElementContent).Particle))
			}
			mismatches++
		}
	}
	if mismatches > 0 {
		t.Fatalf("%d fuzz restriction pairs diverged between legacy and unified relations", mismatches)
	}
}
