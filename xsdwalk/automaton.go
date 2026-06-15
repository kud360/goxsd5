package xsdwalk

import "github.com/kud360/goxsd5/xsd"

// MatchedTerm records how one instance child element was accounted for by the
// content model: either by a (resolved) element declaration — for an element
// term, with substitution groups already applied — or by a wildcard.
type MatchedTerm struct {
	Elem     *xsd.ElementDecl // resolved declaration when an element term matched
	Wildcard *xsd.Wildcard    // the wildcard when a wildcard term matched
}

// Matcher checks a sequence of child element names against a content model
// (cvc-complex-type.2 / cvc-particle, Part 1 §3.4.4 / §3.9.4), producing the
// per-child term assignment a validator needs to find each child's governing
// type. It is the demand-driven (pull) use of the model algebra; the same
// states drive a generated deserializer.
//
// LookupGlobal resolves an expanded name to a global element declaration, so
// substitution-group members can be recognised inside the model.
type Matcher struct {
	LookupGlobal func(xsd.QName) *xsd.ElementDecl

	// set per Match call
	children   []xsd.QName
	terms      []MatchedTerm
	interleave bool
	openW      *xsd.Wildcard
	siblings   []*xsd.ElementDecl // element decls in the content model (for ##definedSibling)
}

// Match attempts to account for every name in children using the particle p
// (which may be nil for empty content) plus any open content. On success it
// returns ok=true and a slice parallel to children giving each child's matched
// term. A nil particle accepts only the empty sequence (modulo open content).
func (m *Matcher) Match(p *xsd.Particle, children []xsd.QName, open *xsd.OpenContent) (terms []MatchedTerm, ok bool) {
	m.children = children
	m.terms = make([]MatchedTerm, len(children))
	m.openW, m.interleave = nil, false
	m.siblings = nil
	if p != nil {
		m.collectSiblings(p)
	}
	if open != nil && open.Wildcard != nil {
		switch open.Mode {
		case xsd.OpenContentInterleave:
			m.openW, m.interleave = open.Wildcard, true
		case xsd.OpenContentSuffix:
			m.openW = open.Wildcard // consumed as a trailing run
		}
	}

	final := func(pos int) bool { return m.finish(pos) }
	var matched bool
	if p == nil {
		matched = final(0)
	} else {
		matched = m.matchParticle(p, 0, final)
	}
	if !matched {
		return nil, false
	}
	return m.terms, true
}

// finish accepts position pos as a complete match: every remaining child must
// be absorbed by open content (interleave: trailing run; suffix: the run).
func (m *Matcher) finish(pos int) bool {
	if m.openW == nil {
		return pos == len(m.children)
	}
	for pos < len(m.children) {
		if !m.wildcardOK(m.openW, m.children[pos]) {
			return false
		}
		m.terms[pos] = MatchedTerm{Wildcard: m.openW}
		pos++
	}
	return true
}

// matchParticle matches p (with its occurrence range) starting at pos, then
// invokes cont for each reachable end position; returns whether any succeeded.
func (m *Matcher) matchParticle(p *xsd.Particle, pos int, cont func(int) bool) bool {
	var rec func(count, cur int) bool
	rec = func(count, cur int) bool {
		if count >= p.MinOccurs && cont(cur) {
			return true
		}
		if p.MaxOccurs != xsd.UnboundedOccurs && count >= p.MaxOccurs {
			return false
		}
		// Try one more occurrence that consumes at least one child. An empty
		// occurrence (next == cur) makes no progress; counting it would loop, so
		// it is tracked separately and used only to satisfy a still-unmet
		// minOccurs (a nullable term — e.g. an empty sequence — occurring the
		// required number of times without consuming anything).
		emptyOK := false
		if m.matchTerm(p.Term, cur, func(next int) bool {
			if next == cur {
				emptyOK = true
				return false
			}
			return rec(count+1, next)
		}) {
			return true
		}
		if emptyOK && count < p.MinOccurs {
			return cont(cur)
		}
		return false
	}
	return rec(0, pos)
}

func (m *Matcher) matchTerm(t xsd.Term, pos int, cont func(int) bool) bool {
	switch t := t.(type) {
	case *xsd.ElementDecl:
		return m.matchLeaf(pos, cont, func(name xsd.QName) (MatchedTerm, bool) {
			if d := m.acceptElement(t, name); d != nil {
				return MatchedTerm{Elem: d}, true
			}
			return MatchedTerm{}, false
		})
	case *xsd.Wildcard:
		return m.matchLeaf(pos, cont, func(name xsd.QName) (MatchedTerm, bool) {
			if m.wildcardOK(t, name) {
				return MatchedTerm{Wildcard: t}, true
			}
			return MatchedTerm{}, false
		})
	case *xsd.GroupRef:
		if t.Ref == nil || t.Ref.Group == nil {
			return cont(pos)
		}
		return m.matchGroup(t.Ref.Group, pos, cont)
	case *xsd.ModelGroup:
		return m.matchGroup(t, pos, cont)
	}
	return false
}

// matchLeaf matches a single child at pos with accept, recording the term, and
// in interleave mode may first absorb open-content children.
func (m *Matcher) matchLeaf(pos int, cont func(int) bool, accept func(xsd.QName) (MatchedTerm, bool)) bool {
	if pos < len(m.children) {
		if mt, ok := accept(m.children[pos]); ok {
			m.terms[pos] = mt
			if cont(pos + 1) {
				return true
			}
		}
	}
	if m.interleave && pos < len(m.children) && m.wildcardOK(m.openW, m.children[pos]) {
		m.terms[pos] = MatchedTerm{Wildcard: m.openW}
		if m.matchLeaf(pos+1, cont, accept) {
			return true
		}
	}
	return false
}

func (m *Matcher) matchGroup(g *xsd.ModelGroup, pos int, cont func(int) bool) bool {
	switch g.Compositor {
	case xsd.CompositorChoice:
		for _, p := range g.Particles {
			if m.matchParticle(p, pos, cont) {
				return true
			}
		}
		// An empty choice (no particles) matches nothing but the empty string,
		// which a 0-occurrence wrapping particle already allows; here it fails.
		return len(g.Particles) == 0 && cont(pos)
	case xsd.CompositorAll:
		return m.matchAll(g, pos, cont)
	default: // sequence
		return m.matchSeq(g.Particles, 0, pos, cont)
	}
}

func (m *Matcher) matchSeq(parts []*xsd.Particle, i, pos int, cont func(int) bool) bool {
	if i >= len(parts) {
		return cont(pos)
	}
	return m.matchParticle(parts[i], pos, func(next int) bool {
		return m.matchSeq(parts, i+1, next, cont)
	})
}

// matchAll matches an xs:all group: each member particle consumes children in
// any order, within its occurrence range (cvc-complex-type for all). Members
// have distinct acceptors (UPA/cos-all-limited), so a child is matched against
// whichever member accepts it. Interleave open content is absorbed inline.
func (m *Matcher) matchAll(g *xsd.ModelGroup, pos int, cont func(int) bool) bool {
	used := make([]int, len(g.Particles))
	var rec func(cur int) bool
	rec = func(cur int) bool {
		if m.allMinSatisfied(g, used) && cont(cur) {
			return true
		}
		if cur >= len(m.children) {
			return false
		}
		name := m.children[cur]
		for i, p := range g.Particles {
			if p.MaxOccurs != xsd.UnboundedOccurs && used[i] >= p.MaxOccurs {
				continue
			}
			mt, ok := m.allAccept(p.Term, name)
			if !ok {
				continue
			}
			used[i]++
			m.terms[cur] = mt
			if rec(cur + 1) {
				return true
			}
			used[i]--
		}
		// Interleave open content: skip a child the model can't place here.
		if m.interleave && m.wildcardOK(m.openW, name) {
			m.terms[cur] = MatchedTerm{Wildcard: m.openW}
			if rec(cur + 1) {
				return true
			}
		}
		return false
	}
	return rec(pos)
}

func (m *Matcher) allMinSatisfied(g *xsd.ModelGroup, used []int) bool {
	for i, p := range g.Particles {
		if used[i] < p.MinOccurs {
			return false
		}
	}
	return true
}

func (m *Matcher) allAccept(t xsd.Term, name xsd.QName) (MatchedTerm, bool) {
	switch t := t.(type) {
	case *xsd.ElementDecl:
		if d := m.acceptElement(t, name); d != nil {
			return MatchedTerm{Elem: d}, true
		}
	case *xsd.Wildcard:
		if m.wildcardOK(t, name) {
			return MatchedTerm{Wildcard: t}, true
		}
	}
	return MatchedTerm{}, false
}

// acceptElement reports whether a child named name is accepted by the element
// term, returning the resolved declaration (the term itself, or a substitution-
// group member). An abstract element is not directly instantiable.
func (m *Matcher) acceptElement(term *xsd.ElementDecl, name xsd.QName) *xsd.ElementDecl {
	if name == term.Name && !term.Abstract {
		return term
	}
	if term.Global && m.LookupGlobal != nil {
		if g := m.LookupGlobal(name); g != nil && !g.Abstract && SubstitutableFor(g, term, term.Block) {
			return g
		}
	}
	return nil
}

// wildcardOK reports whether the wildcard w admits an element named q in the
// current content model, applying the context-dependent ##defined/##definedSibling
// keyword exclusions (cvc-wildcard clauses 2 & 3) on top of WildcardAllows.
func (m *Matcher) wildcardOK(w *xsd.Wildcard, q xsd.QName) bool {
	if w == nil || !WildcardAllows(w, q) {
		return false
	}
	for _, d := range w.NotQName {
		if d.Namespace != "" {
			continue
		}
		switch d.Local {
		case "##defined":
			// clause 2.1: q must not resolve to a global element declaration.
			if m.LookupGlobal != nil && m.LookupGlobal(q) != nil {
				return false
			}
		case "##definedSibling":
			// clause 3: q must not match any element declaration contained in the
			// content model (directly, or implicitly via substitution groups).
			if m.matchesSibling(q) {
				return false
			}
		}
	}
	return true
}

// matchesSibling reports whether q is the name of, or is substitutable for, any
// element declaration appearing in the content model being matched.
func (m *Matcher) matchesSibling(q xsd.QName) bool {
	var g *xsd.ElementDecl
	resolved := false
	for _, d := range m.siblings {
		if q == d.Name {
			return true
		}
		if !resolved && m.LookupGlobal != nil {
			g, resolved = m.LookupGlobal(q), true
		}
		if g != nil && SubstitutableFor(g, d, d.Block) {
			return true
		}
	}
	return false
}

// collectSiblings gathers every element declaration appearing in the content
// model particle p (recursing through groups), for the ##definedSibling check.
func (m *Matcher) collectSiblings(p *xsd.Particle) {
	switch t := p.Term.(type) {
	case *xsd.ElementDecl:
		m.siblings = append(m.siblings, t)
	case *xsd.ModelGroup:
		for _, sub := range t.Particles {
			m.collectSiblings(sub)
		}
	case *xsd.GroupRef:
		if t.Ref != nil && t.Ref.Group != nil {
			for _, sub := range t.Ref.Group.Particles {
				m.collectSiblings(sub)
			}
		}
	}
}
