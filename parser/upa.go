package parser

// Unique Particle Attribution (cos-nonambig, §3.8.6.4) for sequence/choice
// content models, via the Glushkov position automaton.
//
// Each element/wildcard particle is a "position". The construction yields, for
// the content model, the set of start positions (first) and, for every
// position i, the set of positions that may immediately follow it (follow[i]).
// A content model violates UPA iff some such set contains two "competing"
// positions: two element particles that can accept a common element (by name
// or substitution group), or two wildcard particles with overlapping namespace
// constraints. (Element-vs-wildcard never compete: the element wins.)
//
// <all> groups are not regular expressions (they interleave); they are checked
// separately by checkAllUPA. If the automaton meets an <all> group it bails out
// (sets done=false) and no sequence/choice verdict is produced for that model.

import "github.com/kud360/goxsd5/xsd"

// upaLeaf is one element or wildcard particle, identified by its position index.
type upaLeaf struct {
	decl *xsd.ElementDecl // non-nil for an element particle
	wc   *xsd.Wildcard    // non-nil for a wildcard particle
	pos  xsd.Pos
}

type upaAutomaton struct {
	leaves []upaLeaf
	follow []map[int]bool
	bail   bool // an <all> group was encountered
}

func (u *upaAutomaton) addLeaf(l upaLeaf) int {
	i := len(u.leaves)
	u.leaves = append(u.leaves, l)
	u.follow = append(u.follow, map[int]bool{})
	return i
}

func (u *upaAutomaton) link(from, to []int) {
	for _, a := range from {
		for _, b := range to {
			u.follow[a][b] = true
		}
	}
}

// visit returns (nullable, first, last) for a particle subtree and records
// follow edges along the way.
func (u *upaAutomaton) visit(p *xsd.Particle) (nullable bool, first, last []int) {
	if p == nil || u.bail {
		return true, nil, nil
	}
	switch t := p.Term.(type) {
	case *xsd.ElementDecl:
		i := u.addLeaf(upaLeaf{decl: t, pos: p.Pos})
		nullable, first, last = false, []int{i}, []int{i}
	case *xsd.Wildcard:
		i := u.addLeaf(upaLeaf{wc: t, pos: p.Pos})
		nullable, first, last = false, []int{i}, []int{i}
	case *xsd.ModelGroup:
		nullable, first, last = u.visitGroup(t)
	case *xsd.GroupRef:
		if t.Ref == nil || t.Ref.Group == nil {
			return true, nil, nil
		}
		nullable, first, last = u.visitGroup(t.Ref.Group)
	default:
		return true, nil, nil
	}
	if u.bail {
		return true, nil, nil
	}
	// Repetition: after one iteration's last positions, another iteration's
	// first positions may follow.
	if p.MaxOccurs == xsd.UnboundedOccurs || p.MaxOccurs > 1 {
		u.link(last, first)
	}
	if p.MinOccurs == 0 {
		nullable = true
	}
	return nullable, first, last
}

func (u *upaAutomaton) visitGroup(mg *xsd.ModelGroup) (nullable bool, first, last []int) {
	if u.bail {
		return true, nil, nil
	}
	switch mg.Compositor {
	case xsd.CompositorAll:
		u.bail = true
		return true, nil, nil
	case xsd.CompositorChoice:
		// An empty choice matches nothing and contributes no positions.
		for _, c := range mg.Particles {
			cn, cf, cl := u.visit(c)
			first = append(first, cf...)
			last = append(last, cl...)
			if cn {
				nullable = true
			}
		}
		return nullable, first, last
	default: // sequence
		nullable = true
		var activeLast []int // tails so far that the next child may follow
		for _, c := range mg.Particles {
			cn, cf, cl := u.visit(c)
			if u.bail {
				return true, nil, nil
			}
			if nullable {
				first = append(first, cf...)
			}
			u.link(activeLast, cf)
			if cn {
				activeLast = append(activeLast, cl...)
			} else {
				activeLast = cl
			}
			nullable = nullable && cn
		}
		return nullable, first, activeLast
	}
}

// checkSeqChoiceUPA builds the position automaton for ct's content model and
// reports cos-nonambig violations among competing positions.
func (b *builder) checkSeqChoiceUPA(ct *xsd.ComplexType, accepted func(*xsd.ElementDecl) map[xsd.QName]bool) {
	ec, ok := ct.Content.(*xsd.ElementContent)
	if !ok || ec.Particle == nil {
		return
	}
	u := &upaAutomaton{}
	_, first, _ := u.visit(ec.Particle)
	if u.bail {
		return // <all> group present; handled by checkAllUPA
	}

	conflict := func(i, j int) bool {
		a, c := u.leaves[i], u.leaves[j]
		switch {
		case a.decl != nil && c.decl != nil:
			return namesOverlap(accepted(a.decl), accepted(c.decl))
		case a.wc != nil && c.wc != nil:
			return wildcardsOverlap(a.wc, c.wc)
		default:
			return false // element vs wildcard: element wins, no competition
		}
	}
	reported := map[[2]int]bool{}
	check := func(set map[int]bool) {
		idx := make([]int, 0, len(set))
		for i := range set {
			idx = append(idx, i)
		}
		for a := 0; a < len(idx); a++ {
			for c := a + 1; c < len(idx); c++ {
				i, j := idx[a], idx[c]
				if i > j {
					i, j = j, i
				}
				if reported[[2]int{i, j}] || !conflict(i, j) {
					continue
				}
				reported[[2]int{i, j}] = true
				b.errf(xsd.SpecCosNonambig, u.leaves[j].pos, "%s competes with %s in the content model of %s (Unique Particle Attribution)", upaLeafName(u.leaves[i]), upaLeafName(u.leaves[j]), describeCT(ct))
			}
		}
	}

	startSet := map[int]bool{}
	for _, i := range first {
		startSet[i] = true
	}
	check(startSet)
	for i := range u.leaves {
		check(u.follow[i])
	}
}

func upaLeafName(l upaLeaf) string {
	if l.decl != nil {
		return "element " + l.decl.Name.String()
	}
	return "a wildcard"
}
