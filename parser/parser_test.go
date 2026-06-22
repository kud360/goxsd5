package parser

// M7 tests: composition (import/include/redefine/override) through the
// public Parse entry point with an injected resolver.

import (
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/kud360/goxsd5/xsd"
)

// mapResolver serves schema documents from a map keyed by resolved
// location (relative references join against the referencing document).
type mapResolver map[string]string

func (m mapResolver) Resolve(location, base string) (io.ReadCloser, error) {
	uri := resolveLocation(location, base)
	src, ok := m[uri]
	if !ok {
		return nil, fmt.Errorf("no such document: %s", uri)
	}
	return io.NopCloser(strings.NewReader(src)), nil
}

func parseMap(t *testing.T, files map[string]string, root string) ([]*xsd.Schema, error) {
	t.Helper()
	schemas, err := Parse(root, &Options{Resolver: mapResolver(files)})
	if schemas == nil {
		t.Fatalf("Parse returned no schemas (err: %v)", err)
	}
	return schemas, err
}

func wantErrIDs(t *testing.T, err error, ids ...string) {
	t.Helper()
	got := xsd.RefIDs(err)
	for _, id := range ids {
		if !slices.Contains(got, id) {
			t.Errorf("missing expected error id %q; got %v\nerror: %v", id, got, err)
		}
	}
}

func wantNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected errors: %v", err)
	}
}

// rejects asserts that the simple type refuses the lexical.
func rejects(t *testing.T, st *xsd.SimpleType, lexical string) {
	t.Helper()
	if _, err := st.ParseValue(lexical, nil); err == nil {
		t.Errorf("type %s accepted %q, want rejection", st.Name, lexical)
	}
}

func accepts(t *testing.T, st *xsd.SimpleType, lexical string) {
	t.Helper()
	if _, err := st.ParseValue(lexical, nil); err != nil {
		t.Errorf("type %s rejected %q: %v", st.Name, lexical, err)
	}
}

const xsNS = `xmlns:xs="http://www.w3.org/2001/XMLSchema"`

func TestImport(t *testing.T) {
	t.Run("cross-namespace reference", func(t *testing.T) {
		schemas, err := parseMap(t, map[string]string{
			"a.xsd": `<xs:schema ` + xsNS + ` xmlns:b="urn:b" targetNamespace="urn:a">
  <xs:import namespace="urn:b" schemaLocation="b.xsd"/>
  <xs:element name="e" type="b:T"/>
</xs:schema>`,
			"b.xsd": `<xs:schema ` + xsNS + ` targetNamespace="urn:b">
  <xs:simpleType name="T"><xs:restriction base="xs:string"/></xs:simpleType>
</xs:schema>`,
		}, "a.xsd")
		wantNoErr(t, err)
		if len(schemas) != 2 {
			t.Fatalf("got %d schemas, want 2", len(schemas))
		}
		if schemas[0].TargetNamespace != "urn:a" || schemas[1].TargetNamespace != "urn:b" {
			t.Fatalf("namespace order = %s, %s", schemas[0].TargetNamespace, schemas[1].TargetNamespace)
		}
		e := schemas[0].Elements[xsd.QName{Namespace: "urn:a", Local: "e"}]
		bt := schemas[1].Types[xsd.QName{Namespace: "urn:b", Local: "T"}]
		if e == nil || bt == nil || e.Type != bt {
			t.Errorf("element type not linked across the import")
		}
		if !slices.Contains(schemas[0].Imports, "urn:b") {
			t.Errorf("Imports = %v, want urn:b", schemas[0].Imports)
		}
	})

	t.Run("cyclic imports are legal", func(t *testing.T) {
		_, err := parseMap(t, map[string]string{
			"a.xsd": `<xs:schema ` + xsNS + ` xmlns:b="urn:b" targetNamespace="urn:a">
  <xs:import namespace="urn:b" schemaLocation="b.xsd"/>
  <xs:simpleType name="AT"><xs:restriction base="xs:string"/></xs:simpleType>
  <xs:element name="ea" type="b:BT"/>
</xs:schema>`,
			"b.xsd": `<xs:schema ` + xsNS + ` xmlns:a="urn:a" targetNamespace="urn:b">
  <xs:import namespace="urn:a" schemaLocation="a.xsd"/>
  <xs:simpleType name="BT"><xs:restriction base="xs:string"/></xs:simpleType>
  <xs:element name="eb" type="a:AT"/>
</xs:schema>`,
		}, "a.xsd")
		wantNoErr(t, err)
	})

	t.Run("reference without import fails", func(t *testing.T) {
		_, err := parseMap(t, map[string]string{
			"a.xsd": `<xs:schema ` + xsNS + ` xmlns:b="urn:b" targetNamespace="urn:a">
  <xs:import namespace="urn:b" schemaLocation="b.xsd"/>
  <xs:element name="ea" type="b:BT"/>
</xs:schema>`,
			// b references urn:a without importing it.
			"b.xsd": `<xs:schema ` + xsNS + ` xmlns:a="urn:a" targetNamespace="urn:b">
  <xs:simpleType name="BT"><xs:restriction base="xs:string"/></xs:simpleType>
  <xs:element name="eb" type="a:AT"/>
</xs:schema>`,
		}, "a.xsd")
		wantErrIDs(t, err, "src-resolve")
	})

	t.Run("import of own namespace fails", func(t *testing.T) {
		_, err := parseMap(t, map[string]string{
			"a.xsd": `<xs:schema ` + xsNS + ` targetNamespace="urn:a">
  <xs:import namespace="urn:a" schemaLocation="a.xsd"/>
</xs:schema>`,
		}, "a.xsd")
		wantErrIDs(t, err, "src-import")
	})

	t.Run("imported document namespace mismatch fails", func(t *testing.T) {
		_, err := parseMap(t, map[string]string{
			"a.xsd": `<xs:schema ` + xsNS + ` targetNamespace="urn:a">
  <xs:import namespace="urn:b" schemaLocation="c.xsd"/>
</xs:schema>`,
			"c.xsd": `<xs:schema ` + xsNS + ` targetNamespace="urn:c"/>`,
		}, "a.xsd")
		wantErrIDs(t, err, "src-import")
	})

	t.Run("unresolvable import location is tolerated", func(t *testing.T) {
		_, err := parseMap(t, map[string]string{
			"a.xsd": `<xs:schema ` + xsNS + ` targetNamespace="urn:a">
  <xs:import namespace="urn:b" schemaLocation="missing.xsd"/>
</xs:schema>`,
		}, "a.xsd")
		wantNoErr(t, err)
	})

	t.Run("import without namespace into no-namespace schema fails", func(t *testing.T) {
		_, err := parseMap(t, map[string]string{
			"a.xsd": `<xs:schema ` + xsNS + `>
  <xs:import schemaLocation="missing.xsd"/>
</xs:schema>`,
		}, "a.xsd")
		wantErrIDs(t, err, "src-import")
	})
}

func TestInclude(t *testing.T) {
	t.Run("same namespace merge", func(t *testing.T) {
		schemas, err := parseMap(t, map[string]string{
			"a.xsd": `<xs:schema ` + xsNS + ` xmlns:tns="urn:t" targetNamespace="urn:t">
  <xs:include schemaLocation="b.xsd"/>
  <xs:element name="e" type="tns:T"/>
</xs:schema>`,
			"b.xsd": `<xs:schema ` + xsNS + ` targetNamespace="urn:t">
  <xs:simpleType name="T"><xs:restriction base="xs:string"/></xs:simpleType>
</xs:schema>`,
		}, "a.xsd")
		wantNoErr(t, err)
		if len(schemas) != 1 {
			t.Fatalf("got %d schemas, want 1 (include merges)", len(schemas))
		}
		s := schemas[0]
		e := s.Elements[xsd.QName{Namespace: "urn:t", Local: "e"}]
		ty := s.Types[xsd.QName{Namespace: "urn:t", Local: "T"}]
		if e == nil || ty == nil || e.Type != ty {
			t.Error("included type not merged into the including namespace")
		}
	})

	t.Run("chameleon include absorbs the including namespace", func(t *testing.T) {
		schemas, err := parseMap(t, map[string]string{
			"a.xsd": `<xs:schema ` + xsNS + ` xmlns:tns="urn:t" targetNamespace="urn:t">
  <xs:include schemaLocation="cham.xsd"/>
  <xs:element name="e" type="tns:small"/>
</xs:schema>`,
			// No targetNamespace; the unqualified references must be
			// remapped to urn:t.
			"cham.xsd": `<xs:schema ` + xsNS + `>
  <xs:simpleType name="small"><xs:restriction base="xs:string"><xs:maxLength value="3"/></xs:restriction></xs:simpleType>
  <xs:element name="inner" type="small"/>
</xs:schema>`,
		}, "a.xsd")
		wantNoErr(t, err)
		if len(schemas) != 1 {
			t.Fatalf("got %d schemas, want 1", len(schemas))
		}
		s := schemas[0]
		small, _ := s.Types[xsd.QName{Namespace: "urn:t", Local: "small"}].(*xsd.SimpleType)
		if small == nil {
			t.Fatal("chameleon type not absorbed into urn:t")
		}
		rejects(t, small, "abcd")
		inner := s.Elements[xsd.QName{Namespace: "urn:t", Local: "inner"}]
		if inner == nil || inner.Type != xsd.Type(small) {
			t.Error("unqualified reference inside the chameleon document not remapped")
		}
		if e := s.Elements[xsd.QName{Namespace: "urn:t", Local: "e"}]; e == nil || e.Type != xsd.Type(small) {
			t.Error("includer's reference to the absorbed type not linked")
		}
	})

	t.Run("namespace mismatch fails", func(t *testing.T) {
		_, err := parseMap(t, map[string]string{
			"a.xsd": `<xs:schema ` + xsNS + ` targetNamespace="urn:a">
  <xs:include schemaLocation="b.xsd"/>
</xs:schema>`,
			"b.xsd": `<xs:schema ` + xsNS + ` targetNamespace="urn:b"/>`,
		}, "a.xsd")
		wantErrIDs(t, err, "src-include")
	})

	t.Run("unresolvable include fails", func(t *testing.T) {
		_, err := parseMap(t, map[string]string{
			"a.xsd": `<xs:schema ` + xsNS + ` targetNamespace="urn:a">
  <xs:include schemaLocation="missing.xsd"/>
</xs:schema>`,
		}, "a.xsd")
		wantErrIDs(t, err, "src-include")
	})

	t.Run("cyclic include terminates", func(t *testing.T) {
		_, err := parseMap(t, map[string]string{
			"a.xsd": `<xs:schema ` + xsNS + ` targetNamespace="urn:t">
  <xs:include schemaLocation="b.xsd"/>
  <xs:simpleType name="A"><xs:restriction base="xs:string"/></xs:simpleType>
</xs:schema>`,
			"b.xsd": `<xs:schema ` + xsNS + ` targetNamespace="urn:t">
  <xs:include schemaLocation="a.xsd"/>
  <xs:simpleType name="B"><xs:restriction base="xs:string"/></xs:simpleType>
</xs:schema>`,
		}, "a.xsd")
		wantNoErr(t, err)
	})

	t.Run("relative locations resolve against the including document", func(t *testing.T) {
		_, err := parseMap(t, map[string]string{
			"dir/a.xsd": `<xs:schema ` + xsNS + ` targetNamespace="urn:t">
  <xs:include schemaLocation="sub/b.xsd"/>
</xs:schema>`,
			"dir/sub/b.xsd": `<xs:schema ` + xsNS + ` targetNamespace="urn:t">
  <xs:include schemaLocation="c.xsd"/>
</xs:schema>`,
			"dir/sub/c.xsd": `<xs:schema ` + xsNS + ` targetNamespace="urn:t"/>`,
		}, "dir/a.xsd")
		wantNoErr(t, err)
	})
}

func TestRedefine(t *testing.T) {
	files := func(redefineChildren string) map[string]string {
		return map[string]string{
			"main.xsd": `<xs:schema ` + xsNS + ` xmlns:tns="urn:t" targetNamespace="urn:t">
  <xs:redefine schemaLocation="base.xsd">` + redefineChildren + `</xs:redefine>
</xs:schema>`,
			"base.xsd": `<xs:schema ` + xsNS + ` xmlns:tns="urn:t" targetNamespace="urn:t">
  <xs:simpleType name="st"><xs:restriction base="xs:string"><xs:maxLength value="10"/></xs:restriction></xs:simpleType>
  <xs:element name="e" type="tns:st"/>
  <xs:group name="g"><xs:sequence><xs:element name="x" type="xs:int"/></xs:sequence></xs:group>
  <xs:attributeGroup name="ag"><xs:attribute name="a" type="xs:int"/></xs:attributeGroup>
</xs:schema>`,
		}
	}

	t.Run("type redefinition is pervasive and derives from the original", func(t *testing.T) {
		schemas, err := parseMap(t, files(
			`<xs:simpleType name="st"><xs:restriction base="tns:st"><xs:maxLength value="3"/></xs:restriction></xs:simpleType>`,
		), "main.xsd")
		wantNoErr(t, err)
		if len(schemas) != 1 {
			t.Fatalf("got %d schemas, want 1", len(schemas))
		}
		s := schemas[0]
		st, _ := s.Types[xsd.QName{Namespace: "urn:t", Local: "st"}].(*xsd.SimpleType)
		if st == nil {
			t.Fatal("redefined type missing")
		}
		rejects(t, st, "abcd") // the redefinition's maxLength=3
		base, _ := st.BaseType.(*xsd.SimpleType)
		if base == nil {
			t.Fatal("redefined type has no simple base")
		}
		accepts(t, base, "abcd") // the original's maxLength=10
		rejects(t, base, "abcdefghijk")
		// The base document's own reference sees the redefinition.
		if e := s.Elements[xsd.QName{Namespace: "urn:t", Local: "e"}]; e == nil || e.Type != xsd.Type(st) {
			t.Error("reference in the redefined document does not see the redefinition")
		}
	})

	t.Run("group redefinition resolves its self-reference to the original", func(t *testing.T) {
		schemas, err := parseMap(t, files(
			`<xs:group name="g"><xs:sequence><xs:group ref="tns:g"/><xs:element name="y" type="xs:int"/></xs:sequence></xs:group>`,
		), "main.xsd")
		wantNoErr(t, err)
		g := schemas[0].Groups[xsd.QName{Namespace: "urn:t", Local: "g"}]
		if g == nil || g.Group == nil {
			t.Fatal("redefined group missing")
		}
		if len(g.Group.Particles) != 2 {
			t.Fatalf("redefined group has %d particles, want 2", len(g.Group.Particles))
		}
		ref, _ := g.Group.Particles[0].Term.(*xsd.GroupRef)
		if ref == nil || ref.Ref == g {
			t.Error("group self-reference did not resolve to the original definition")
		}
	})

	t.Run("group redefinition without a self-reference is allowed", func(t *testing.T) {
		// No self-reference is the restriction case (src-redefine 6.2); its
		// subset check is deferred, so the redefinition is accepted.
		_, err := parseMap(t, files(
			`<xs:group name="g"><xs:sequence><xs:element name="x" type="xs:int"/></xs:sequence></xs:group>`,
		), "main.xsd")
		wantNoErr(t, err)
	})

	t.Run("group redefinition with two self-references fails", func(t *testing.T) {
		// src-redefine 6.1.1: at most one self-reference.
		_, err := parseMap(t, files(
			`<xs:group name="g"><xs:sequence><xs:group ref="tns:g"/><xs:group ref="tns:g"/></xs:sequence></xs:group>`,
		), "main.xsd")
		wantErrIDs(t, err, "src-redefine")
	})

	t.Run("group self-reference with maxOccurs > 1 fails", func(t *testing.T) {
		// src-redefine 6.1.2: the self-reference must have minOccurs = maxOccurs = 1.
		_, err := parseMap(t, files(
			`<xs:group name="g"><xs:sequence><xs:group ref="tns:g" maxOccurs="2"/></xs:sequence></xs:group>`,
		), "main.xsd")
		wantErrIDs(t, err, "src-redefine")
	})

	t.Run("attributeGroup redefinition with one self-reference is allowed", func(t *testing.T) {
		_, err := parseMap(t, files(
			`<xs:attributeGroup name="ag"><xs:attributeGroup ref="tns:ag"/><xs:attribute name="b" type="xs:int"/></xs:attributeGroup>`,
		), "main.xsd")
		wantNoErr(t, err)
	})

	t.Run("attributeGroup redefinition with two self-references fails", func(t *testing.T) {
		// src-redefine 7.1: at most one self-reference.
		_, err := parseMap(t, files(
			`<xs:attributeGroup name="ag"><xs:attributeGroup ref="tns:ag"/><xs:attributeGroup ref="tns:ag"/></xs:attributeGroup>`,
		), "main.xsd")
		wantErrIDs(t, err, "src-redefine")
	})

	t.Run("redefining an undeclared component fails", func(t *testing.T) {
		_, err := parseMap(t, files(
			`<xs:simpleType name="nope"><xs:restriction base="tns:nope"/></xs:simpleType>`,
		), "main.xsd")
		wantErrIDs(t, err, "src-redefine")
	})

	t.Run("type redefinition must derive from itself", func(t *testing.T) {
		_, err := parseMap(t, files(
			`<xs:simpleType name="st"><xs:restriction base="xs:string"/></xs:simpleType>`,
		), "main.xsd")
		wantErrIDs(t, err, "src-redefine")
	})
}

// TestRedefineRestriction exercises src-redefine clauses 6.2 / 7.2: a
// no-self-reference <group>/<attributeGroup> redefinition must be a restriction
// (subset) of the original it redefines.
func TestRedefineRestriction(t *testing.T) {
	files := func(base, redefineChildren string) map[string]string {
		return map[string]string{
			"main.xsd": `<xs:schema ` + xsNS + ` xmlns:tns="urn:t" targetNamespace="urn:t">
  <xs:redefine schemaLocation="base.xsd">` + redefineChildren + `</xs:redefine>
</xs:schema>`,
			"base.xsd": `<xs:schema ` + xsNS + ` xmlns:tns="urn:t" targetNamespace="urn:t">` + base + `</xs:schema>`,
		}
	}

	t.Run("group redefinition that narrows occurrence is a valid restriction", func(t *testing.T) {
		_, err := parseMap(t, files(
			`<xs:group name="g"><xs:sequence><xs:element name="x" type="xs:int" maxOccurs="3"/></xs:sequence></xs:group>`,
			`<xs:group name="g"><xs:sequence><xs:element name="x" type="xs:int" maxOccurs="2"/></xs:sequence></xs:group>`,
		), "main.xsd")
		wantNoErr(t, err)
	})

	t.Run("group redefinition that widens occurrence fails src-redefine", func(t *testing.T) {
		_, err := parseMap(t, files(
			`<xs:group name="g"><xs:sequence><xs:element name="x" type="xs:int" maxOccurs="2"/></xs:sequence></xs:group>`,
			`<xs:group name="g"><xs:sequence><xs:element name="x" type="xs:int" maxOccurs="5"/></xs:sequence></xs:group>`,
		), "main.xsd")
		wantErrIDs(t, err, "src-redefine")
	})

	t.Run("group redefinition that adds a new element fails src-redefine", func(t *testing.T) {
		_, err := parseMap(t, files(
			`<xs:group name="g"><xs:sequence><xs:element name="x" type="xs:int"/></xs:sequence></xs:group>`,
			`<xs:group name="g"><xs:sequence><xs:element name="x" type="xs:int"/><xs:element name="y" type="xs:int"/></xs:sequence></xs:group>`,
		), "main.xsd")
		wantErrIDs(t, err, "src-redefine")
	})

	t.Run("group redefinition dropping an optional element is a valid restriction", func(t *testing.T) {
		_, err := parseMap(t, files(
			`<xs:group name="g"><xs:sequence><xs:element name="x" type="xs:int"/><xs:element name="y" type="xs:int" minOccurs="0"/></xs:sequence></xs:group>`,
			`<xs:group name="g"><xs:sequence><xs:element name="x" type="xs:int"/></xs:sequence></xs:group>`,
		), "main.xsd")
		wantNoErr(t, err)
	})

	t.Run("attributeGroup redefinition that is a strict subset is valid", func(t *testing.T) {
		_, err := parseMap(t, files(
			`<xs:attributeGroup name="ag"><xs:attribute name="a" type="xs:int"/><xs:attribute name="b" type="xs:int"/></xs:attributeGroup>`,
			`<xs:attributeGroup name="ag"><xs:attribute name="a" type="xs:int"/></xs:attributeGroup>`,
		), "main.xsd")
		wantNoErr(t, err)
	})

	t.Run("attributeGroup redefinition adding a required attribute fails src-redefine", func(t *testing.T) {
		_, err := parseMap(t, files(
			`<xs:attributeGroup name="ag"><xs:attribute name="a" type="xs:int"/></xs:attributeGroup>`,
			`<xs:attributeGroup name="ag"><xs:attribute name="a" type="xs:int"/><xs:attribute name="b" type="xs:int" use="required"/></xs:attributeGroup>`,
		), "main.xsd")
		wantErrIDs(t, err, "src-redefine")
	})

	t.Run("attributeGroup redefinition dropping a required attribute fails src-redefine", func(t *testing.T) {
		_, err := parseMap(t, files(
			`<xs:attributeGroup name="ag"><xs:attribute name="a" type="xs:int" use="required"/></xs:attributeGroup>`,
			`<xs:attributeGroup name="ag"><xs:attribute name="a" type="xs:int"/></xs:attributeGroup>`,
		), "main.xsd")
		wantErrIDs(t, err, "src-redefine")
	})
}

func TestOverride(t *testing.T) {
	t.Run("override replaces pervasively, unmatched children are ignored", func(t *testing.T) {
		schemas, err := parseMap(t, map[string]string{
			"main.xsd": `<xs:schema ` + xsNS + ` xmlns:tns="urn:t" targetNamespace="urn:t">
  <xs:override schemaLocation="base.xsd">
    <xs:simpleType name="st"><xs:restriction base="xs:string"><xs:maxLength value="3"/></xs:restriction></xs:simpleType>
    <xs:element name="unmatched" type="xs:int"/>
  </xs:override>
</xs:schema>`,
			"base.xsd": `<xs:schema ` + xsNS + ` xmlns:tns="urn:t" targetNamespace="urn:t">
  <xs:simpleType name="st"><xs:restriction base="xs:string"><xs:maxLength value="10"/></xs:restriction></xs:simpleType>
  <xs:simpleType name="keep"><xs:restriction base="xs:int"/></xs:simpleType>
  <xs:element name="e" type="tns:st"/>
</xs:schema>`,
		}, "main.xsd")
		wantNoErr(t, err)
		s := schemas[0]
		st, _ := s.Types[xsd.QName{Namespace: "urn:t", Local: "st"}].(*xsd.SimpleType)
		if st == nil {
			t.Fatal("overriding type missing")
		}
		rejects(t, st, "abcd") // the override's maxLength=3, not the original's 10
		// The overridden document's own reference sees the override.
		if e := s.Elements[xsd.QName{Namespace: "urn:t", Local: "e"}]; e == nil || e.Type != xsd.Type(st) {
			t.Error("reference in the overridden document does not see the override")
		}
		if s.Types[xsd.QName{Namespace: "urn:t", Local: "keep"}] == nil {
			t.Error("non-overridden component lost")
		}
		// spec: override transformation — children matching nothing in the
		// overridden document do not become components.
		if s.Elements[xsd.QName{Namespace: "urn:t", Local: "unmatched"}] != nil {
			t.Error("unmatched override child became a component")
		}
	})

	t.Run("override applies transitively through includes", func(t *testing.T) {
		schemas, err := parseMap(t, map[string]string{
			"main.xsd": `<xs:schema ` + xsNS + ` targetNamespace="urn:t">
  <xs:override schemaLocation="mid.xsd">
    <xs:simpleType name="st"><xs:restriction base="xs:string"><xs:maxLength value="3"/></xs:restriction></xs:simpleType>
  </xs:override>
</xs:schema>`,
			"mid.xsd": `<xs:schema ` + xsNS + ` targetNamespace="urn:t">
  <xs:include schemaLocation="deep.xsd"/>
</xs:schema>`,
			"deep.xsd": `<xs:schema ` + xsNS + ` targetNamespace="urn:t">
  <xs:simpleType name="st"><xs:restriction base="xs:string"><xs:maxLength value="10"/></xs:restriction></xs:simpleType>
</xs:schema>`,
		}, "main.xsd")
		wantNoErr(t, err)
		st, _ := schemas[0].Types[xsd.QName{Namespace: "urn:t", Local: "st"}].(*xsd.SimpleType)
		if st == nil {
			t.Fatal("overriding type missing")
		}
		rejects(t, st, "abcd")
	})

	t.Run("override of an override registers only the outermost replacement", func(t *testing.T) {
		// top overrides mid, mid overrides deep; all three replace "st". The
		// override transformation is applied outermost-last, so only top's
		// replacement becomes a component — registering mid's too would be a
		// spurious duplicate (over009).
		schemas, err := parseMap(t, map[string]string{
			"top.xsd": `<xs:schema ` + xsNS + ` targetNamespace="urn:t">
  <xs:override schemaLocation="mid.xsd">
    <xs:simpleType name="st"><xs:restriction base="xs:string"><xs:maxLength value="3"/></xs:restriction></xs:simpleType>
  </xs:override>
</xs:schema>`,
			"mid.xsd": `<xs:schema ` + xsNS + ` targetNamespace="urn:t">
  <xs:override schemaLocation="deep.xsd">
    <xs:simpleType name="st"><xs:restriction base="xs:string"><xs:maxLength value="6"/></xs:restriction></xs:simpleType>
  </xs:override>
</xs:schema>`,
			"deep.xsd": `<xs:schema ` + xsNS + ` targetNamespace="urn:t">
  <xs:simpleType name="st"><xs:restriction base="xs:string"><xs:maxLength value="10"/></xs:restriction></xs:simpleType>
</xs:schema>`,
		}, "top.xsd")
		wantNoErr(t, err)
		st, _ := schemas[0].Types[xsd.QName{Namespace: "urn:t", Local: "st"}].(*xsd.SimpleType)
		if st == nil {
			t.Fatal("overriding type missing")
		}
		rejects(t, st, "abcd") // top's maxLength=3 wins
	})
}

func TestParseFileResolver(t *testing.T) {
	// The default resolver reads from disk; exercise it on a small pair of
	// fixture files under testdata.
	schemas, err := Parse("testdata/m7/main.xsd", nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(schemas) != 1 {
		t.Fatalf("got %d schemas, want 1", len(schemas))
	}
	if schemas[0].Types[xsd.QName{Namespace: "urn:m7", Local: "T"}] == nil {
		t.Error("included type missing")
	}
}

func TestParseMissingRoot(t *testing.T) {
	if _, err := Parse("testdata/m7/definitely-missing.xsd", nil); err == nil {
		t.Fatal("Parse of a missing root succeeded")
	}
}

func TestParseMultiple(t *testing.T) {
	t.Run("two docs spanning two namespaces produce the union schema set", func(t *testing.T) {
		files := mapResolver{
			"a.xsd": `<xs:schema ` + xsNS + ` targetNamespace="urn:a">
  <xs:element name="ea" type="xs:string"/>
</xs:schema>`,
			"b.xsd": `<xs:schema ` + xsNS + ` targetNamespace="urn:b">
  <xs:element name="eb" type="xs:int"/>
</xs:schema>`,
		}
		schemas, err := ParseMultiple([]string{"a.xsd", "b.xsd"}, &Options{Resolver: files})
		wantNoErr(t, err)
		if len(schemas) != 2 {
			t.Fatalf("got %d schemas, want 2", len(schemas))
		}
		ns := map[string]bool{}
		for _, s := range schemas {
			ns[s.TargetNamespace] = true
		}
		if !ns["urn:a"] || !ns["urn:b"] {
			t.Errorf("namespace set = %v, want {urn:a, urn:b}", ns)
		}
		// Verify elements from both docs are present.
		var sa, sb *xsd.Schema
		for _, s := range schemas {
			switch s.TargetNamespace {
			case "urn:a":
				sa = s
			case "urn:b":
				sb = s
			}
		}
		if sa == nil || sa.Elements[xsd.QName{Namespace: "urn:a", Local: "ea"}] == nil {
			t.Error("element ea from a.xsd not in schema set")
		}
		if sb == nil || sb.Elements[xsd.QName{Namespace: "urn:b", Local: "eb"}] == nil {
			t.Error("element eb from b.xsd not in schema set")
		}
	})

	t.Run("duplicate location is loaded once", func(t *testing.T) {
		files := mapResolver{
			"a.xsd": `<xs:schema ` + xsNS + ` targetNamespace="urn:a">
  <xs:element name="ea" type="xs:string"/>
</xs:schema>`,
		}
		schemas, err := ParseMultiple([]string{"a.xsd", "a.xsd"}, &Options{Resolver: files})
		wantNoErr(t, err)
		if len(schemas) != 1 {
			t.Fatalf("got %d schemas, want 1 (deduped)", len(schemas))
		}
	})

	t.Run("missing root returns error", func(t *testing.T) {
		files := mapResolver{}
		if _, err := ParseMultiple([]string{"missing.xsd"}, &Options{Resolver: files}); err == nil {
			t.Fatal("ParseMultiple of a missing root succeeded")
		}
	})
}
