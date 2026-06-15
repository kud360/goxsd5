package xsdvalidate

import (
	"strings"

	"github.com/kud360/goxsd5/builtin"
	"github.com/kud360/goxsd5/xpath"
	"github.com/kud360/goxsd5/xsd"
)

// This file adapts the abstract infoset (Element/Attribute/Text) to the
// xpath package's Node/NodeAttr model and supplies the type-cast callback, so
// xs:assert and xs:alternative tests are evaluated by the shared XPath
// evaluator (the engine itself stays format-agnostic; xpath imports neither
// xsd nor builtin).

// xpNode adapts an infoset Element to xpath.Node.
type xpNode struct{ el Element }

func (n xpNode) NodeName() xpath.Name { return xpath.Name(n.el.Name()) }

func (n xpNode) NodeAttrs() []xpath.NodeAttr {
	own := n.el.Attributes()
	out := make([]xpath.NodeAttr, 0, len(own))
	for _, a := range own {
		if a.Name().Namespace == xsd.XMLNSNS {
			continue
		}
		out = append(out, xpAttr{a})
	}
	return out
}

func (n xpNode) NodeChildren() []xpath.Node {
	kids := elementChildren(n.el)
	out := make([]xpath.Node, len(kids))
	for i, k := range kids {
		out[i] = xpNode{k}
	}
	return out
}

func (n xpNode) StringValue() string { return nodeString(n.el) }

// xpAttr adapts an infoset Attribute to xpath.NodeAttr.
type xpAttr struct{ at Attribute }

func (a xpAttr) AttrName() xpath.Name { return xpath.Name(a.at.Name()) }
func (a xpAttr) AttrValue() string    { return a.at.Value() }

// nodeString is the XPath string value of an element: the concatenation of all
// descendant character data.
func nodeString(el Element) string {
	var b strings.Builder
	var walk func(e Element)
	walk = func(e Element) {
		for _, c := range e.Children() {
			switch c := c.(type) {
			case Text:
				b.WriteString(c.Data())
			case Element:
				walk(c)
			}
		}
	}
	walk(el)
	return b.String()
}

// evalAssertion evaluates an xs:assert / xs:alternative test against the context
// element. ok is false when the expression is outside the evaluator's supported
// subset, so callers fail open. Built-in casts resolve QNames against ctx's
// in-scope namespaces.
func evalAssertion(ctx Element, expr string) (result, ok bool) {
	return xpath.EvalBool(expr, xpNode{ctx}, castableContext(ctx, nil))
}

// evalSimpleAssertion evaluates a simple-type xs:assertion (XSD 1.1 §3.13.4)
// against the element carrying the value, with $value bound to the validated
// value. The element node serves as the context item, so "." atomizes to the
// same value. ok is false for any unsupported construct (fail open).
func evalSimpleAssertion(ctx Element, expr string, value []string) (result, ok bool) {
	return xpath.EvalBool(expr, xpNode{ctx}, castableContext(ctx, map[string][]string{"value": value}))
}

// castableContext builds the shared EvalContext: built-in casts resolve QNames
// against ctx's in-scope namespaces, plus any variable bindings.
func castableContext(ctx Element, vars map[string][]string) xpath.EvalContext {
	return xpath.EvalContext{
		Castable: func(local, val string) bool {
			t := builtin.Lookup(local)
			if t == nil {
				return false
			}
			_, err := t.ParseValue(val, nsContext{ctx})
			return err == nil
		},
		Vars: vars,
	}
}
