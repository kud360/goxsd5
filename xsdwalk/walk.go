package xsdwalk

import "github.com/kud360/goxsd5/xsd"

// ParticleFunc is invoked for each particle reached while walking a content
// model. depth is the nesting level: 0 for the particles directly under the
// content-model root, incrementing by one for each model group (or referenced
// group) descended into. Returning false stops the walk of this particle's
// subtree (a compositor's children are not visited); returning true descends.
type ParticleFunc func(p *xsd.Particle, depth int) bool

// Walker is the push (exhaustive, schema-only) driver of the model algebra: it
// performs a depth-first traversal of a complex type's effective content model,
// enumerating every particle reachable through its sequences, choices, all
// groups, and named-group references. It is the schema-side counterpart to the
// instance-driven Matcher in automaton.go; both decode the same particle/term
// union, the reusable core being the algebra, not the driver.
//
// A Walker is reusable across calls. Each Walk resets the internal cycle-guard
// state, so callers may share one Walker; it holds no per-call state between
// invocations. It is not safe for concurrent use (the package is single-
// threaded by contract).
type Walker struct {
	// visited records the named groups currently on the recursion stack, keyed on
	// group identity, so a recursive xs:group reference is detected and its
	// re-entry skipped rather than looping forever.
	visited map[*xsd.Group]bool
}

// NewWalker returns a Walker ready for use. The cycle-guard set is allocated
// once here and reset at the start of each Walk.
func NewWalker() *Walker {
	return &Walker{visited: map[*xsd.Group]bool{}}
}

// Walk does a depth-first traversal of t's effective content model, calling fn
// for each particle in document-model order (a compositor particle is visited
// before its children). fn's return value governs descent: false prunes the
// subtree, true descends.
//
// The traversal of a well-formed compiled model cannot fail, so Walk has no
// error return; a caller decides what to do with the visited particles. (The
// issue that introduced this driver suggested an error return for symmetry with
// future fallible traversals; there is no failure mode today, so the minimal
// signature is preferred per the package's API conventions.)
//
// Only element content carries a particle; simple, empty, and absent content
// models yield no particles and Walk returns nil. Named-group references are
// followed into the referenced group's particles, with a cycle guard on group
// identity so recursive groups terminate. Substitution groups are NOT expanded:
// element terms (including abstract or head elements) are visited as authored —
// substitution is an instance-time concern handled by the Matcher.
//
// Walk traverses the explicit particle tree only. A wildcard term authored in
// the model (xs:any) is visited as an ordinary particle, but the XSD 1.1 open
// content property (xs:openContent, ElementContent.OpenContent) is a separate
// component, not a particle, and is not visited; read it off the content type
// directly if needed.
func (w *Walker) Walk(t *xsd.ComplexType, fn ParticleFunc) {
	if w.visited == nil {
		w.visited = map[*xsd.Group]bool{}
	}
	clear(w.visited)
	if t == nil {
		return
	}
	ec, ok := t.Content.(*xsd.ElementContent)
	if !ok || ec.Particle == nil {
		return
	}
	w.walkParticle(ec.Particle, 0, fn)
}

// walkParticle visits one particle then, if fn allowed descent, its term's
// children.
func (w *Walker) walkParticle(p *xsd.Particle, depth int, fn ParticleFunc) {
	if !fn(p, depth) {
		return
	}
	switch t := p.Term.(type) {
	case *xsd.ModelGroup:
		w.walkGroup(t, depth+1, fn)
	case *xsd.GroupRef:
		w.walkRef(t, depth+1, fn)
	}
}

// walkGroup visits each particle of a model group.
func (w *Walker) walkGroup(g *xsd.ModelGroup, depth int, fn ParticleFunc) {
	for _, p := range g.Particles {
		w.walkParticle(p, depth, fn)
	}
}

// walkRef follows a named-group reference into its model group, guarding against
// recursive references on group identity.
func (w *Walker) walkRef(ref *xsd.GroupRef, depth int, fn ParticleFunc) {
	if ref.Ref == nil || ref.Ref.Group == nil {
		return
	}
	if w.visited[ref.Ref] {
		return
	}
	w.visited[ref.Ref] = true
	w.walkGroup(ref.Ref.Group, depth, fn)
	delete(w.visited, ref.Ref)
}

// Elements returns every element declaration reachable in t's content model, in
// document-model order (the depth-first order a Walker visits them). Group
// references are followed (recursive ones terminate via the cycle guard);
// substitution groups are not expanded, so each declaration appears exactly as
// authored. A type with simple, empty, or absent content yields nil.
func Elements(t *xsd.ComplexType) []*xsd.ElementDecl {
	var out []*xsd.ElementDecl
	NewWalker().Walk(t, func(p *xsd.Particle, _ int) bool {
		if e, ok := p.Term.(*xsd.ElementDecl); ok {
			out = append(out, e)
		}
		return true
	})
	return out
}
