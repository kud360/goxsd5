package parser

// Annotation and foreign-content capture: every component keeps its
// xs:annotation content and any foreign-namespace attributes/elements so a
// parse/mutate round trip does not lose unknown content.

import (
	"github.com/kud360/goxsd5/parser/xmltree"
	"github.com/kud360/goxsd5/xsd"
)

// annotationOf maps n's first xs:annotation child (the only one the content
// models admit before other content) to the model.
func annotationOf(n *xmltree.Node, doc *schemaDoc) *xsd.Annotation {
	c := firstChild(n, doc, "annotation")
	if c == nil {
		return nil
	}
	return buildAnnotation(c, doc)
}

func buildAnnotation(c *xmltree.Node, doc *schemaDoc) *xsd.Annotation {
	a := &xsd.Annotation{}
	for _, k := range xsdElems(c, doc) {
		switch k.Name.Local {
		case "documentation":
			a.Documentation = append(a.Documentation, k.CharData)
		case "appinfo":
			for _, f := range k.Children {
				a.AppInfo = append(a.AppInfo, foreignNode(f))
			}
		}
	}
	return a
}

// extensionsOf collects n's foreign-namespace attributes and child
// elements.
func extensionsOf(n *xmltree.Node) xsd.Extensions {
	var ext xsd.Extensions
	for i := range n.Attrs {
		a := &n.Attrs[i]
		if a.Name.Space != "" && a.Name.Space != xsd.XSDNS {
			ext.Attrs = append(ext.Attrs, xsd.ForeignAttr{
				Name:  xsd.QName{Namespace: a.Name.Space, Local: a.Name.Local},
				Value: a.Value,
			})
		}
	}
	for _, c := range n.Children {
		if c.Name.Space != xsd.XSDNS {
			ext.Nodes = append(ext.Nodes, foreignNode(c))
		}
	}
	return ext
}

// foreignNode deep-copies an xmltree subtree into the model's
// parser-independent ForeignNode shape.
func foreignNode(n *xmltree.Node) *xsd.ForeignNode {
	f := &xsd.ForeignNode{
		Name:     xsd.QName{Namespace: n.Name.Space, Local: n.Name.Local},
		CharData: n.CharData,
	}
	for i := range n.Attrs {
		a := &n.Attrs[i]
		f.Attrs = append(f.Attrs, xsd.ForeignAttr{
			Name:  xsd.QName{Namespace: a.Name.Space, Local: a.Name.Local},
			Value: a.Value,
		})
	}
	for _, c := range n.Children {
		f.Children = append(f.Children, foreignNode(c))
	}
	return f
}
