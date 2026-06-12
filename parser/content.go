package parser

// Content-model matching for the structural (src-*) validation of schema
// documents. Each xs:* element's permitted children are described by a tiny
// grammar over child local names, transcribed from the "XML Representation
// Summary" blocks of XSD 1.1 Part 1 §3 and Part 2 §4.3. The matcher is a
// set-of-positions NFA walk: tiny grammars, child lists of any length,
// linear in practice because the grammars are deterministic.

import (
	"slices"
	"sort"
)

// cm is one content-model node. advance returns every index reachable from
// i after consuming this node, recording failed expectations in m.
type cm interface {
	advance(m *matcher, i int) []int
}

type matcher struct {
	toks []string
	// furthest is the highest index at which a token failed to match;
	// expected accumulates the names that would have matched there.
	furthest int
	expected []string
}

// matchContent matches toks against c. On failure it reports the index of
// the first offending child (== len(toks) when a required child is missing)
// and the names that were expected there.
func matchContent(c cm, toks []string) (ok bool, failIdx int, expected []string) {
	m := &matcher{toks: toks}
	for _, end := range c.advance(m, 0) {
		if end == len(toks) {
			return true, 0, nil
		}
		// Matched a prefix; the next token is the offender.
		m.expect(end, "(end of content)")
	}
	return false, m.furthest, m.expected
}

func (m *matcher) expect(i int, name string) {
	if i > m.furthest {
		m.furthest = i
		m.expected = m.expected[:0]
	}
	if i == m.furthest && !slices.Contains(m.expected, name) {
		m.expected = append(m.expected, name)
		sort.Strings(m.expected)
	}
}

// merge unions two ascending index slices.
func merge(a, b []int) []int {
	if len(b) == 0 {
		return a
	}
	if len(a) == 0 {
		return append(a, b...)
	}
	out := make([]int, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) || j < len(b) {
		switch {
		case j == len(b) || (i < len(a) && a[i] < b[j]):
			out = append(out, a[i])
			i++
		case i == len(a) || b[j] < a[i]:
			out = append(out, b[j])
			j++
		default:
			out = append(out, a[i])
			i++
			j++
		}
	}
	return out
}

// one matches exactly one child with the given local name.
type cmOne string

func (c cmOne) advance(m *matcher, i int) []int {
	if i < len(m.toks) && m.toks[i] == string(c) {
		return []int{i + 1}
	}
	m.expect(i, "xs:"+string(c))
	return nil
}

type cmSeq []cm

func (c cmSeq) advance(m *matcher, i int) []int {
	states := []int{i}
	for _, part := range c {
		var next []int
		for _, s := range states {
			next = merge(next, part.advance(m, s))
		}
		if len(next) == 0 {
			return nil
		}
		states = next
	}
	return states
}

type cmChoice []cm

func (c cmChoice) advance(m *matcher, i int) []int {
	var out []int
	for _, alt := range c {
		out = merge(out, alt.advance(m, i))
	}
	return out
}

type cmEmpty struct{}

func (cmEmpty) advance(_ *matcher, i int) []int { return []int{i} }

type cmStar struct{ c cm }

func (c cmStar) advance(m *matcher, i int) []int {
	states := []int{i}
	work := []int{i}
	for len(work) > 0 {
		s := work[0]
		work = work[1:]
		for _, n := range c.c.advance(m, s) {
			if !slices.Contains(states, n) {
				states = merge(states, []int{n})
				work = append(work, n)
			}
		}
	}
	return states
}

func one(name string) cm   { return cmOne(name) }
func seq(parts ...cm) cm   { return cmSeq(parts) }
func choice(alts ...cm) cm { return cmChoice(alts) }
func star(c cm) cm         { return cmStar{c} }
func opt(c cm) cm          { return cmChoice{cmEmpty{}, c} }
func plus(c cm) cm         { return cmSeq{c, cmStar{c}} }
func names(ns ...string) cm {
	alts := make(cmChoice, len(ns))
	for i, n := range ns {
		alts[i] = cmOne(n)
	}
	return alts
}
