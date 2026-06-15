package parser

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kud360/goxsd5/parser/xmltree"
)

func TestSchemaLocationHints(t *testing.T) {
	cases := []struct {
		name string
		doc  string
		want []string
	}{
		{
			name: "schemaLocation pairs keep only locations",
			doc: `<r xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
			        xsi:schemaLocation="ns/a a.xsd ns/b b.xsd"/>`,
			want: []string{"a.xsd", "b.xsd"},
		},
		{
			name: "noNamespaceSchemaLocation",
			doc: `<r xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
			        xsi:noNamespaceSchemaLocation="local.xsd"/>`,
			want: []string{"local.xsd"},
		},
		{
			name: "both, schemaLocation first",
			doc: `<r xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
			        xsi:schemaLocation="ns/a  a.xsd"
			        xsi:noNamespaceSchemaLocation="local.xsd"/>`,
			want: []string{"a.xsd", "local.xsd"},
		},
		{
			name: "none",
			doc:  `<r/>`,
			want: nil,
		},
		{
			name: "non-xsi schemaLocation attribute ignored",
			doc:  `<r schemaLocation="not-a-hint.xsd"/>`,
			want: nil,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			root, err := xmltree.Parse(strings.NewReader(c.doc), c.name)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			got := SchemaLocationHints(root)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("SchemaLocationHints = %q, want %q", got, c.want)
			}
		})
	}
}
