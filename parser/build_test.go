package parser

// M6 builder (pass 2) tests: negatives for the constraints enforced during
// component construction, plus value-model assertions for lists, unions,
// simpleContent restrictions, and complex-type assembly.

import (
	"testing"

	"github.com/kud360/goxsd5/xsd"
)

// buildAll runs both passes over src. Fixtures must be structurally valid:
// a pass-1 error means the test exercises the wrong pass and is a test bug.
func buildAll(t *testing.T, src string) (*xsd.Schema, *xsd.ErrorList) {
	t.Helper()
	doc, reg, errs := load(t, src)
	if !errs.Empty() {
		t.Fatalf("pass 1 rejected the fixture (want pass-2 coverage): %v", errs.Err())
	}
	s := buildSchema(reg, doc, errs)
	if s == nil {
		t.Fatal("buildSchema returned nil")
	}
	return s, errs
}

func TestBuilderNegatives(t *testing.T) {
	cases := []struct {
		name string
		body string // wrapped in <xs:schema targetNamespace="urn:t" ...>
		ids  []string
	}{
		// Cyclic definitions.
		{"cyclic simpleType restriction chain",
			`<xs:simpleType name="a"><xs:restriction base="tns:b"/></xs:simpleType>
			 <xs:simpleType name="b"><xs:restriction base="tns:a"/></xs:simpleType>`,
			[]string{"st-props-correct"}},
		{"simpleType is a list of itself",
			`<xs:simpleType name="l"><xs:list itemType="tns:l"/></xs:simpleType>`,
			[]string{"st-props-correct"}},
		{"cyclic complexType derivation chain",
			`<xs:complexType name="a"><xs:complexContent><xs:extension base="tns:b"><xs:sequence/></xs:extension></xs:complexContent></xs:complexType>
			 <xs:complexType name="b"><xs:complexContent><xs:extension base="tns:a"><xs:sequence/></xs:extension></xs:complexContent></xs:complexType>`,
			[]string{"ct-props-correct"}},
		{"complexType restricts itself",
			`<xs:complexType name="a"><xs:complexContent><xs:restriction base="tns:a"><xs:sequence/></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"ct-props-correct"}},

		// Reference resolution (src-resolve).
		{"element type undeclared", `<xs:element name="e" type="tns:nope"/>`, []string{"src-resolve"}},
		{"element ref undeclared",
			`<xs:complexType name="c"><xs:sequence><xs:element ref="tns:nope"/></xs:sequence></xs:complexType>`,
			[]string{"src-resolve"}},
		{"attribute ref undeclared",
			`<xs:complexType name="c"><xs:attribute ref="tns:nope"/></xs:complexType>`,
			[]string{"src-resolve"}},
		{"group ref undeclared",
			`<xs:complexType name="c"><xs:sequence><xs:group ref="tns:nope"/></xs:sequence></xs:complexType>`,
			[]string{"src-resolve"}},
		{"attributeGroup ref undeclared",
			`<xs:complexType name="c"><xs:attributeGroup ref="tns:nope"/></xs:complexType>`,
			[]string{"src-resolve"}},
		{"substitutionGroup head undeclared",
			`<xs:element name="e" type="xs:int" substitutionGroup="tns:nope"/>`,
			[]string{"src-resolve"}},
		{"two-element circular substitution group",
			`<xs:element name="foo" type="xs:string" substitutionGroup="tns:bar"/>
			 <xs:element name="bar" type="xs:string" substitutionGroup="tns:foo"/>`,
			[]string{"e-props-correct"}},
		{"three-hop circular substitution group",
			`<xs:element name="foo" type="xs:string" substitutionGroup="tns:bar"/>
			 <xs:element name="bar" type="xs:string" substitutionGroup="tns:zot"/>
			 <xs:element name="zot" type="xs:string" substitutionGroup="tns:foo"/>`,
			[]string{"e-props-correct"}},
		{"member type not derived from head type",
			`<xs:element name="head" type="xs:integer"/>
			 <xs:element name="mem" type="xs:string" substitutionGroup="tns:head"/>`,
			[]string{"e-props-correct"}},
		{"notQName names a namespace excluded by notNamespace",
			`<xs:complexType name="c"><xs:sequence/><xs:anyAttribute notNamespace="http://www.w3.org/XML/1998/namespace" notQName="xml:space"/></xs:complexType>`,
			[]string{"w-props-correct"}},
		{"notQName names a namespace outside an enumeration",
			`<xs:complexType name="c"><xs:sequence><xs:any namespace="urn:ok" notQName="tns:foo"/></xs:sequence></xs:complexType>`,
			[]string{"w-props-correct"}},
		{"mixed extension adds content to an element-only base",
			`<xs:complexType name="b"><xs:sequence><xs:element name="x" type="xs:int"/></xs:sequence></xs:complexType>
			 <xs:complexType name="t" mixed="true"><xs:complexContent><xs:extension base="tns:b"><xs:sequence><xs:element name="y" type="xs:int"/></xs:sequence></xs:extension></xs:complexContent></xs:complexType>`,
			[]string{"cos-ct-extends"}},
		{"non-all model group nested in all",
			`<xs:group name="g"><xs:sequence><xs:element name="b" type="xs:int"/></xs:sequence></xs:group>
			 <xs:complexType name="c"><xs:all><xs:element name="a" type="xs:int"/><xs:group ref="tns:g"/></xs:all></xs:complexType>`,
			[]string{"cos-all-limited"}},
		{"nested all group with maxOccurs greater than one",
			`<xs:group name="g"><xs:all><xs:element name="b" type="xs:int"/></xs:all></xs:group>
			 <xs:complexType name="c"><xs:all><xs:element name="a" type="xs:int"/><xs:group ref="tns:g" maxOccurs="3"/></xs:all></xs:complexType>`,
			[]string{"cos-all-limited"}},
		{"same element twice in an all group",
			`<xs:element name="o" type="xs:int"/>
			 <xs:complexType name="c"><xs:all><xs:element ref="tns:o"/><xs:element name="x" type="xs:int"/><xs:element ref="tns:o"/></xs:all></xs:complexType>`,
			[]string{"cos-nonambig"}},
		{"substitutable element competes in an all group",
			`<xs:element name="o" type="xs:int"/>
			 <xs:element name="p" type="xs:int" substitutionGroup="tns:o"/>
			 <xs:complexType name="c"><xs:all><xs:element ref="tns:o"/><xs:element ref="tns:p"/></xs:all></xs:complexType>`,
			[]string{"cos-nonambig"}},
		{"overlapping wildcards in an all group",
			`<xs:complexType name="c"><xs:all><xs:any namespace="urn:a urn:b"/><xs:any namespace="urn:b urn:c"/></xs:all></xs:complexType>`,
			[]string{"cos-nonambig"}},
		{"repeated optional element then required same element (a*,a)",
			`<xs:complexType name="c"><xs:sequence><xs:element name="a" type="xs:int" minOccurs="0" maxOccurs="unbounded"/><xs:element name="a" type="xs:int"/></xs:sequence></xs:complexType>`,
			[]string{"cos-nonambig"}},
		{"choice of element and substitutable element",
			`<xs:element name="head" type="xs:int"/>
			 <xs:element name="mem" type="xs:int" substitutionGroup="tns:head"/>
			 <xs:complexType name="c"><xs:choice><xs:element ref="tns:head"/><xs:element ref="tns:mem"/></xs:choice></xs:complexType>`,
			[]string{"cos-nonambig"}},
		{"choice of overlapping wildcards",
			`<xs:complexType name="c"><xs:choice><xs:any namespace="urn:a urn:b"/><xs:any namespace="urn:b"/></xs:choice></xs:complexType>`,
			[]string{"cos-nonambig"}},
		{"all extension re-adds a base element (UPA)",
			`<xs:complexType name="b"><xs:all><xs:element name="x" type="xs:int"/></xs:all></xs:complexType>
			 <xs:complexType name="e"><xs:complexContent><xs:extension base="tns:b"><xs:all><xs:element name="x" type="xs:int"/></xs:all></xs:extension></xs:complexContent></xs:complexType>`,
			[]string{"cos-nonambig"}},
		{"sequence cannot extend an all group",
			`<xs:complexType name="b"><xs:all><xs:element name="x" type="xs:int"/></xs:all></xs:complexType>
			 <xs:complexType name="e"><xs:complexContent><xs:extension base="tns:b"><xs:sequence><xs:element name="y" type="xs:int"/></xs:sequence></xs:extension></xs:complexContent></xs:complexType>`,
			[]string{"cos-all-limited"}},
		{"all extension must keep the base minOccurs",
			`<xs:complexType name="b"><xs:all><xs:element name="x" type="xs:int"/></xs:all></xs:complexType>
			 <xs:complexType name="e"><xs:complexContent><xs:extension base="tns:b"><xs:all minOccurs="0"><xs:element name="y" type="xs:int"/></xs:all></xs:extension></xs:complexContent></xs:complexType>`,
			[]string{"cos-ct-extends"}},
		// Particle restriction (cos-particle-restrict, §3.4.6.4 / §3.9.6):
		// per-name occurrence/type bag check for flat all/sequence models.
		{"all restriction loosens a base element minOccurs",
			`<xs:complexType name="b"><xs:all><xs:element name="a" type="xs:int" minOccurs="1" maxOccurs="5"/><xs:element name="d" type="xs:int"/></xs:all></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:all><xs:element name="d" type="xs:int"/><xs:element name="a" type="xs:int" minOccurs="0" maxOccurs="4"/></xs:all></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"cos-particle-restrict"}},
		{"all restriction widens a base element maxOccurs",
			`<xs:complexType name="b"><xs:all><xs:element name="d" type="xs:int" minOccurs="1" maxOccurs="1"/></xs:all></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:all><xs:element name="d" type="xs:int" minOccurs="1" maxOccurs="5"/></xs:all></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"cos-particle-restrict"}},
		{"all restriction omits a required base element",
			`<xs:complexType name="b"><xs:all><xs:element name="b" type="xs:int" minOccurs="1"/><xs:element name="d" type="xs:int" minOccurs="1"/></xs:all></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:all><xs:element name="b" type="xs:int" minOccurs="1"/></xs:all></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"cos-particle-restrict"}},
		{"all restriction introduces an element the base disallows",
			`<xs:complexType name="b"><xs:all><xs:element name="b" type="xs:int"/></xs:all></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:all><xs:element name="b" type="xs:int"/><xs:element name="f" type="xs:int" minOccurs="0"/></xs:all></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"cos-particle-restrict"}},
		{"all restriction gives a child an unrelated type",
			`<xs:complexType name="b"><xs:all><xs:element name="c" type="xs:positiveInteger"/></xs:all></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:all><xs:element name="c" type="xs:int"/></xs:all></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"cos-particle-restrict"}},
		{"sequence restriction of an all loosens a base minOccurs",
			`<xs:complexType name="b"><xs:all><xs:element name="a" type="xs:int" minOccurs="1" maxOccurs="5"/></xs:all></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:sequence><xs:element name="a" type="xs:int" minOccurs="0" maxOccurs="4"/></xs:sequence></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"cos-particle-restrict"}},
		{"all restriction wildcard widens the base wildcard namespace",
			`<xs:complexType name="b"><xs:all><xs:any namespace="urn:y"/></xs:all></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:all><xs:any namespace="urn:x urn:y"/></xs:all></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"cos-particle-restrict"}},
		{"all restriction splits a base wildcard past its maxOccurs",
			`<xs:complexType name="b"><xs:all><xs:any namespace="urn:x urn:y" maxOccurs="5"/></xs:all></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:all><xs:any namespace="urn:x" maxOccurs="3"/><xs:any namespace="urn:y" maxOccurs="3"/></xs:all></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"cos-particle-restrict"}},
		{"all restriction splits a base wildcard below its minOccurs",
			`<xs:complexType name="b"><xs:all><xs:any namespace="urn:x urn:y" minOccurs="5" maxOccurs="unbounded"/></xs:all></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:all><xs:any namespace="urn:x" minOccurs="2" maxOccurs="unbounded"/><xs:any namespace="urn:y" minOccurs="2" maxOccurs="unbounded"/></xs:all></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"cos-particle-restrict"}},
		{"restriction introduces a content wildcard the base lacks",
			`<xs:complexType name="b"><xs:sequence><xs:element name="a" type="xs:int"/></xs:sequence></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:sequence><xs:element name="a" type="xs:int"/><xs:any namespace="##any" minOccurs="0"/></xs:sequence></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"cos-particle-restrict"}},
		{"all restriction splits a head whose substitutes overshoot maxOccurs",
			`<xs:element name="a" type="xs:int"/>
			 <xs:element name="A1" type="xs:int" substitutionGroup="tns:a"/>
			 <xs:element name="A2" type="xs:int" substitutionGroup="tns:a"/>
			 <xs:complexType name="b"><xs:all><xs:element ref="tns:a" minOccurs="10" maxOccurs="20"/></xs:all></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:all><xs:element ref="tns:A1" minOccurs="6" maxOccurs="15"/><xs:element ref="tns:A2" minOccurs="6" maxOccurs="15"/></xs:all></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"cos-particle-restrict"}},
		// Multi-base-wildcard restriction (cos-particle-restrict): two or more
		// base wildcards, so the per-name slot mapping gives way to the packing
		// solver. With disjoint base regions the region count check is exact.
		{"two base wildcards: restriction underflows a base wildcard minimum",
			`<xs:complexType name="b"><xs:all><xs:any namespace="urn:one urn:two" minOccurs="5" maxOccurs="unbounded"/><xs:any namespace="urn:three" minOccurs="0" maxOccurs="2"/></xs:all></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:all><xs:any namespace="urn:one" minOccurs="3" maxOccurs="unbounded"/><xs:any namespace="urn:two urn:three" minOccurs="2" maxOccurs="2"/></xs:all></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"cos-particle-restrict"}},
		{"two base wildcards: restriction overshoots a base wildcard maximum",
			`<xs:complexType name="b"><xs:all><xs:any namespace="urn:x" minOccurs="0" maxOccurs="2"/><xs:any namespace="urn:y" minOccurs="0" maxOccurs="2"/></xs:all></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:all><xs:any namespace="urn:x" minOccurs="0" maxOccurs="3"/><xs:any namespace="urn:y" minOccurs="0" maxOccurs="2"/></xs:all></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"cos-particle-restrict"}},
		{"two base wildcards: restriction wildcard escapes every base wildcard",
			`<xs:complexType name="b"><xs:all><xs:any namespace="urn:x" minOccurs="0" maxOccurs="2"/><xs:any namespace="urn:y" minOccurs="0" maxOccurs="2"/></xs:all></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:all><xs:any namespace="urn:x urn:z" minOccurs="0" maxOccurs="2"/></xs:all></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"cos-particle-restrict"}},
		{"two base wildcards with an overlapping element: restriction admits a notQName-excluded name",
			`<xs:complexType name="b"><xs:all><xs:element name="nm" type="xs:string"/><xs:any namespace="##local" notQName="a b c" minOccurs="0" maxOccurs="2" processContents="skip"/><xs:any notNamespace="##local" minOccurs="0" maxOccurs="2" processContents="skip"/></xs:all></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:sequence><xs:element name="nm" type="xs:string"/><xs:any namespace="##any" notQName="a b" minOccurs="1" maxOccurs="1" processContents="skip"/></xs:sequence></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"cos-particle-restrict"}},
		// Attribute wildcard restriction subset (derivation-ok-restriction /
		// Wildcard Subset, §3.4.6.3 / §3.10.6.2): R's wildcard ⊆ B's.
		{"restriction attribute wildcard widens an enumeration past a notNamespace base",
			`<xs:complexType name="b"><xs:sequence/><xs:anyAttribute notNamespace="urn:x urn:y"/></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:sequence/><xs:anyAttribute namespace="urn:z urn:x"/></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"derivation-ok-restriction"}},
		{"restriction attribute wildcard not is broader than the base not",
			`<xs:complexType name="b"><xs:sequence/><xs:anyAttribute notNamespace="urn:x urn:y"/></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:sequence/><xs:anyAttribute notNamespace="urn:x"/></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"derivation-ok-restriction"}},
		{"restriction attribute wildcard ##any over a constrained base",
			`<xs:complexType name="b"><xs:sequence/><xs:anyAttribute notNamespace="urn:x"/></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:sequence/><xs:anyAttribute namespace="##any"/></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"derivation-ok-restriction"}},
		{"restriction attribute wildcard drops the base ##defined disallowance",
			`<xs:complexType name="b"><xs:sequence/><xs:anyAttribute namespace="##any" notQName="##defined" processContents="skip"/></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:sequence/><xs:anyAttribute namespace="##local" processContents="skip"/></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"derivation-ok-restriction"}},
		{"restriction adds an attribute wildcard the base lacks",
			`<xs:complexType name="b"><xs:sequence/></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:sequence/><xs:anyAttribute namespace="##any"/></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"derivation-ok-restriction"}},
		{"all restriction by a choice whose branch overshoots a base maxOccurs",
			`<xs:complexType name="b"><xs:all><xs:element name="a" type="xs:int" minOccurs="0" maxOccurs="5"/><xs:element name="bb" type="xs:int" minOccurs="1" maxOccurs="5"/></xs:all></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:choice><xs:sequence><xs:element name="bb" type="xs:int" minOccurs="3" maxOccurs="4"/></xs:sequence><xs:sequence><xs:element name="a" type="xs:int" minOccurs="1" maxOccurs="8"/><xs:element name="bb" type="xs:int" minOccurs="3" maxOccurs="4"/></xs:sequence></xs:choice></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"cos-particle-restrict"}},
		{"choice restriction branch omits a required base element",
			`<xs:complexType name="b"><xs:sequence><xs:element name="a" type="xs:int"/><xs:element name="bb" type="xs:int"/></xs:sequence></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:choice><xs:element name="a" type="xs:int"/><xs:element name="bb" type="xs:int"/></xs:choice></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"cos-particle-restrict"}},
		// Open content restriction subset (derivation-ok-restriction §3.4.6.4
		// clause 9): R's open content may not out-reach the base's (open016-019).
		{"restriction adds open content the base lacks",
			`<xs:complexType name="b"><xs:sequence><xs:element name="a" type="xs:int"/></xs:sequence></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:openContent mode="suffix"><xs:any namespace="urn:o"/></xs:openContent><xs:sequence><xs:element name="a" type="xs:int"/></xs:sequence></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"derivation-ok-restriction"}},
		{"restriction open content widens the base wildcard namespace",
			`<xs:complexType name="b"><xs:openContent mode="suffix"><xs:any namespace="urn:o"/></xs:openContent><xs:sequence><xs:element name="a" type="xs:int"/></xs:sequence></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:openContent mode="suffix"><xs:any namespace="urn:o urn:p"/></xs:openContent><xs:sequence><xs:element name="a" type="xs:int"/></xs:sequence></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"derivation-ok-restriction"}},
		{"restriction open content weakens processContents",
			`<xs:complexType name="b"><xs:openContent mode="suffix"><xs:any namespace="urn:o" processContents="strict"/></xs:openContent><xs:sequence><xs:element name="a" type="xs:int"/></xs:sequence></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:openContent mode="suffix"><xs:any namespace="urn:o" processContents="lax"/></xs:openContent><xs:sequence><xs:element name="a" type="xs:int"/></xs:sequence></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"derivation-ok-restriction"}},
		{"restriction open content is more open (interleave) than the base (suffix)",
			`<xs:complexType name="b"><xs:openContent mode="suffix"><xs:any namespace="urn:o"/></xs:openContent><xs:sequence><xs:element name="a" type="xs:int"/></xs:sequence></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:openContent mode="interleave"><xs:any namespace="urn:o"/></xs:openContent><xs:sequence><xs:element name="a" type="xs:int"/></xs:sequence></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"derivation-ok-restriction"}},
		// Extension open content (cos-ct-extends §3.4.6.2 clause 1.4.3.2.2.3):
		// an extension may not narrow the base's interleaved open content to
		// suffix (open030/033/046).
		// Particle restriction subsumes clause 4.6: a restriction element's type
		// table must match the base element's (cta0043).
		{"restriction changes a matched element's type table",
			`<xs:complexType name="b"><xs:sequence><xs:element name="s"><xs:alternative test="@k='1'" type="xs:token"/></xs:element></xs:sequence></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:sequence><xs:element name="s"><xs:alternative test="@k='1'" type="xs:string"/></xs:element></xs:sequence></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"cos-particle-restrict"}},
		{"restriction element narrows a union to a smaller union (cos-st-derived-ok 2.2.4.2)",
			// union(date,time) is not validly derived from union(date,dateTime,time):
			// a smaller union is not a member of the larger (saxon simple011).
			`<xs:complexType name="b"><xs:sequence><xs:element name="c" type="tns:big"/></xs:sequence></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:sequence><xs:element name="c" type="tns:small"/></xs:sequence></xs:restriction></xs:complexContent></xs:complexType>
			 <xs:simpleType name="big"><xs:union memberTypes="xs:date xs:dateTime xs:time"/></xs:simpleType>
			 <xs:simpleType name="small"><xs:union memberTypes="xs:date xs:time"/></xs:simpleType>`,
			[]string{"cos-st-derived-ok"}},
		{"restriction element type is a member of a facet-restricted union base (cos-st-derived-ok 2.2.4.3)",
			// xs:date is in chap's transitive membership, but chap is a
			// pattern-restricted union, so the member is no longer substitutable
			// for it (saxon simple014).
			`<xs:complexType name="b"><xs:sequence><xs:element name="c" type="tns:chap"/></xs:sequence></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:sequence><xs:element name="c" type="xs:date"/></xs:sequence></xs:restriction></xs:complexContent></xs:complexType>
			 <xs:simpleType name="ddt"><xs:union memberTypes="xs:date xs:dateTime xs:time"/></xs:simpleType>
			 <xs:simpleType name="chap"><xs:restriction base="tns:ddt"><xs:pattern value=".*Z"/></xs:restriction></xs:simpleType>`,
			[]string{"cos-st-derived-ok"}},
		{"restriction element type is a member reached through a facet-restricted intervening union",
			// chap = union(dt, time); dt = pattern-restricted union(date,dateTime).
			// xs:date is in chap's transitive membership but the intervening union
			// dt carries a facet, so it is not substitutable (saxon simple015).
			`<xs:complexType name="b"><xs:sequence><xs:element name="c" type="tns:chap"/></xs:sequence></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:sequence><xs:element name="c" type="xs:date"/></xs:sequence></xs:restriction></xs:complexContent></xs:complexType>
			 <xs:simpleType name="chap"><xs:union memberTypes="tns:dt xs:time"/></xs:simpleType>
			 <xs:simpleType name="dt"><xs:restriction><xs:simpleType><xs:union memberTypes="xs:date xs:dateTime"/></xs:simpleType><xs:pattern value=".*Z"/></xs:restriction></xs:simpleType>`,
			[]string{"cos-st-derived-ok"}},
		{"all restriction drops a named element whose wildcard binds a global of an underived type (saxon wild069)",
			// In the <all> base, the named e (union(date,time)) shadows the lax
			// wildcard, so B always types <e> by the union. The restriction omits
			// the named e, routing <e> to its lax wildcard, which binds the global
			// e (xs:duration) — not derived from union(date,time), so <e>P1Y</e> is
			// valid in r but not b. Order-independence of <all> makes this unsound,
			// unlike the sequence case (wild068, in the valid-models test).
			`<xs:complexType name="b"><xs:all><xs:element name="e" form="qualified" minOccurs="0"><xs:simpleType><xs:union memberTypes="xs:date xs:time"/></xs:simpleType></xs:element><xs:element name="f" type="xs:integer"/><xs:any namespace="##targetNamespace" processContents="lax"/></xs:all></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:all><xs:element name="f" type="xs:integer"/><xs:any namespace="##targetNamespace" processContents="lax"/></xs:all></xs:restriction></xs:complexContent></xs:complexType>
			 <xs:element name="e" type="xs:duration"/>`,
			[]string{"cos-st-derived-ok"}},
		{"all restriction drops a named element whose skip wildcard leaves it unconstrained",
			// The restriction's skip wildcard accepts any content for <e>, which the
			// base's union-typed named e forbids — unsound regardless of any global.
			`<xs:complexType name="b"><xs:all><xs:element name="e" form="qualified" minOccurs="0"><xs:simpleType><xs:union memberTypes="xs:date xs:time"/></xs:simpleType></xs:element><xs:element name="f" type="xs:integer"/><xs:any namespace="##targetNamespace" processContents="skip"/></xs:all></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:all><xs:element name="f" type="xs:integer"/><xs:any namespace="##targetNamespace" processContents="skip"/></xs:all></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"cos-particle-restrict"}},
		{"extension narrows interleave open content to suffix (explicit)",
			`<xs:complexType name="b"><xs:openContent mode="interleave"><xs:any namespace="urn:o"/></xs:openContent><xs:sequence><xs:element name="a" type="xs:int"/></xs:sequence></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:extension base="tns:b"><xs:openContent mode="suffix"><xs:any namespace="urn:o"/></xs:openContent><xs:sequence/></xs:extension></xs:complexContent></xs:complexType>`,
			[]string{"cos-ct-extends"}},
		{"extension narrows interleave to suffix via defaultOpenContent on an empty extension",
			`<xs:defaultOpenContent mode="suffix"><xs:any namespace="urn:o"/></xs:defaultOpenContent>
			 <xs:complexType name="b"><xs:openContent mode="interleave"><xs:any namespace="urn:o"/></xs:openContent><xs:sequence><xs:element name="a" type="xs:int"/></xs:sequence></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:extension base="tns:b"><xs:sequence/></xs:extension></xs:complexContent></xs:complexType>`,
			[]string{"cos-ct-extends"}},
		{"keyref refer undeclared",
			`<xs:element name="e" type="xs:int"><xs:keyref name="r" refer="tns:nope"><xs:selector xpath="a"/><xs:field xpath="b"/></xs:keyref></xs:element>`,
			[]string{"src-resolve"}},
		{"identity constraint ref undeclared",
			`<xs:element name="e" type="xs:int"><xs:key ref="tns:nope"/></xs:element>`,
			[]string{"src-resolve"}},

		// Facet construction.
		{"fractionDigits on a string type",
			`<xs:simpleType name="t"><xs:restriction base="xs:string"><xs:fractionDigits value="2"/></xs:restriction></xs:simpleType>`,
			[]string{"cos-applicable-facets"}},
		{"length on a decimal type",
			`<xs:simpleType name="t"><xs:restriction base="xs:decimal"><xs:length value="3"/></xs:restriction></xs:simpleType>`,
			[]string{"cos-applicable-facets"}},
		{"enumeration value invalid against base",
			`<xs:simpleType name="t"><xs:restriction base="xs:int"><xs:enumeration value="abc"/></xs:restriction></xs:simpleType>`,
			[]string{"enumeration-valid-restriction"}},
		{"bound lexical invalid against base",
			`<xs:simpleType name="t"><xs:restriction base="xs:int"><xs:maxInclusive value="abc"/></xs:restriction></xs:simpleType>`,
			[]string{"maxInclusive-valid-restriction"}},
		{"bound outside the base range",
			`<xs:simpleType name="b"><xs:restriction base="xs:int"><xs:maxInclusive value="100"/></xs:restriction></xs:simpleType>
			 <xs:simpleType name="t"><xs:restriction base="tns:b"><xs:maxInclusive value="200"/></xs:restriction></xs:simpleType>`,
			[]string{"maxInclusive-valid-restriction"}},
		{"minLength loosened in restriction",
			`<xs:simpleType name="b"><xs:restriction base="xs:string"><xs:minLength value="5"/></xs:restriction></xs:simpleType>
			 <xs:simpleType name="t"><xs:restriction base="tns:b"><xs:minLength value="3"/></xs:restriction></xs:simpleType>`,
			[]string{"minLength-valid-restriction"}},
		{"totalDigits raised in restriction",
			`<xs:simpleType name="b"><xs:restriction base="xs:decimal"><xs:totalDigits value="3"/></xs:restriction></xs:simpleType>
			 <xs:simpleType name="t"><xs:restriction base="tns:b"><xs:totalDigits value="5"/></xs:restriction></xs:simpleType>`,
			[]string{"totalDigits-valid-restriction"}},
		{"whiteSpace loosened in restriction",
			`<xs:simpleType name="t"><xs:restriction base="xs:token"><xs:whiteSpace value="preserve"/></xs:restriction></xs:simpleType>`,
			[]string{"whiteSpace-valid-restriction"}},
		{"fixed facet changed in restriction",
			`<xs:simpleType name="b"><xs:restriction base="xs:int"><xs:minInclusive value="0" fixed="true"/></xs:restriction></xs:simpleType>
			 <xs:simpleType name="t"><xs:restriction base="tns:b"><xs:minInclusive value="1"/></xs:restriction></xs:simpleType>`,
			[]string{"fixed-facet-value"}},
		{"minInclusive greater than maxInclusive",
			`<xs:simpleType name="t"><xs:restriction base="xs:int"><xs:minInclusive value="5"/><xs:maxInclusive value="1"/></xs:restriction></xs:simpleType>`,
			[]string{"minInclusive-less-than-equal-to-maxInclusive"}},
		{"minLength greater than maxLength",
			`<xs:simpleType name="t"><xs:restriction base="xs:string"><xs:minLength value="5"/><xs:maxLength value="3"/></xs:restriction></xs:simpleType>`,
			[]string{"minLength-less-than-equal-to-maxLength"}},
		{"duplicate single-valued facet",
			`<xs:simpleType name="t"><xs:restriction base="xs:string"><xs:maxLength value="3"/><xs:maxLength value="4"/></xs:restriction></xs:simpleType>`,
			[]string{"src-single-facet-value"}},
		{"invalid pattern regex",
			`<xs:simpleType name="t"><xs:restriction base="xs:string"><xs:pattern value="["/></xs:restriction></xs:simpleType>`,
			[]string{"regex-valid"}},
		{"explicitTimezone widened from required to optional",
			`<xs:simpleType name="b"><xs:restriction base="xs:time"><xs:explicitTimezone value="required"/></xs:restriction></xs:simpleType>
			 <xs:simpleType name="t"><xs:restriction base="tns:b"><xs:explicitTimezone value="optional"/></xs:restriction></xs:simpleType>`,
			[]string{"explicitTimezone-valid-restriction"}},
		{"explicitTimezone changed from prohibited to required",
			`<xs:simpleType name="b"><xs:restriction base="xs:time"><xs:explicitTimezone value="prohibited"/></xs:restriction></xs:simpleType>
			 <xs:simpleType name="t"><xs:restriction base="tns:b"><xs:explicitTimezone value="required"/></xs:restriction></xs:simpleType>`,
			[]string{"explicitTimezone-valid-restriction"}},

		// Restriction of special / wrong-variety bases.
		{"restriction of anySimpleType",
			`<xs:simpleType name="t"><xs:restriction base="xs:anySimpleType"/></xs:simpleType>`,
			[]string{"cos-st-restricts"}},
		{"simple restriction of a complex type",
			`<xs:complexType name="c"><xs:sequence/></xs:complexType>
			 <xs:simpleType name="t"><xs:restriction base="tns:c"/></xs:simpleType>`,
			[]string{"cos-st-restricts"}},
		{"list whose item type is a list",
			`<xs:simpleType name="t"><xs:list itemType="xs:NMTOKENS"/></xs:simpleType>`,
			[]string{"cos-st-restricts"}},
		{"list item type final=list",
			`<xs:simpleType name="it" final="list"><xs:restriction base="xs:token"/></xs:simpleType>
			 <xs:simpleType name="t"><xs:list itemType="tns:it"/></xs:simpleType>`,
			[]string{"st-props-correct"}},
		{"union member final=union",
			`<xs:simpleType name="m" final="union"><xs:restriction base="xs:token"/></xs:simpleType>
			 <xs:simpleType name="t"><xs:union memberTypes="tns:m"/></xs:simpleType>`,
			[]string{"st-props-correct"}},

		// NOTATION.
		{"NOTATION-derived type without enumeration",
			`<xs:attribute name="a"><xs:simpleType><xs:restriction base="xs:NOTATION"/></xs:simpleType></xs:attribute>`,
			[]string{"enumeration-required-notation"}},
		{"list of NOTATION without enumeration",
			`<xs:simpleType name="t"><xs:list itemType="xs:NOTATION"/></xs:simpleType>`,
			[]string{"enumeration-required-notation"}},
		{"union with a bare NOTATION member",
			`<xs:simpleType name="t"><xs:union memberTypes="xs:string xs:NOTATION"/></xs:simpleType>`,
			[]string{"enumeration-required-notation"}},
		{"NOTATION enumeration value names no declared notation",
			`<xs:notation name="jpeg" public="image/jpeg"/>
			 <xs:attribute name="a"><xs:simpleType><xs:restriction base="xs:NOTATION"><xs:enumeration value="png"/></xs:restriction></xs:simpleType></xs:attribute>`,
			[]string{"enumeration-required-notation"}},

		// The special types must not be used as a list/union item type.
		{"list of xs:anyAtomicType",
			`<xs:simpleType name="t"><xs:list itemType="xs:anyAtomicType"/></xs:simpleType>`,
			[]string{"cos-st-restricts"}},
		{"union member xs:anyAtomicType",
			`<xs:simpleType name="t"><xs:union memberTypes="xs:anyAtomicType xs:string"/></xs:simpleType>`,
			[]string{"cos-st-restricts"}},

		// Element declarations.
		{"substitution excluded by head final",
			`<xs:complexType name="bt"><xs:sequence/></xs:complexType>
			 <xs:complexType name="dt"><xs:complexContent><xs:restriction base="tns:bt"><xs:sequence/></xs:restriction></xs:complexContent></xs:complexType>
			 <xs:element name="head" type="tns:bt" final="restriction"/>
			 <xs:element name="member" type="tns:dt" substitutionGroup="tns:head"/>`,
			[]string{"e-props-correct"}},
		{"element default invalid for its type",
			`<xs:element name="e" type="xs:int" default="abc"/>`,
			[]string{"cos-valid-default"}},
		{"element default with element-only content",
			`<xs:complexType name="c"><xs:sequence><xs:element name="x" type="xs:int"/></xs:sequence></xs:complexType>
			 <xs:element name="e" type="tns:c" default="x"/>`,
			[]string{"cos-valid-default"}},

		// Attribute declarations and uses.
		{"attribute typed by a complex type",
			`<xs:complexType name="c"><xs:sequence/></xs:complexType>
			 <xs:attribute name="a" type="tns:c"/>`,
			[]string{"a-props-correct"}},
		{"attribute default invalid for its type",
			`<xs:attribute name="a" type="xs:int" default="abc"/>`,
			[]string{"a-props-correct"}},
		{"use changes a fixed attribute value",
			`<xs:attribute name="ga" type="xs:string" fixed="a"/>
			 <xs:complexType name="c"><xs:attribute ref="tns:ga" fixed="b"/></xs:complexType>`,
			[]string{"au-props-correct"}},
		{"use adds a default to a fixed attribute",
			`<xs:attribute name="ga" type="xs:string" fixed="a"/>
			 <xs:complexType name="c"><xs:attribute ref="tns:ga" default="b"/></xs:complexType>`,
			[]string{"au-props-correct"}},

		// Complex types.
		{"extension redeclares a base attribute",
			`<xs:complexType name="b"><xs:attribute name="x" type="xs:int"/></xs:complexType>
			 <xs:complexType name="d"><xs:complexContent><xs:extension base="tns:b"><xs:attribute name="x" type="xs:int"/></xs:extension></xs:complexContent></xs:complexType>`,
			[]string{"ct-props-correct"}},
		{"simpleContent restriction of a simple type",
			`<xs:complexType name="c"><xs:simpleContent><xs:restriction base="xs:int"><xs:minInclusive value="0"/></xs:restriction></xs:simpleContent></xs:complexType>`,
			[]string{"src-ct"}},
		{"simpleContent extension of an element-only type",
			`<xs:complexType name="eo"><xs:sequence/></xs:complexType>
			 <xs:complexType name="c"><xs:simpleContent><xs:extension base="tns:eo"/></xs:simpleContent></xs:complexType>`,
			[]string{"src-ct"}},
		{"complexContent with a simple base",
			`<xs:complexType name="c"><xs:complexContent><xs:extension base="xs:int"><xs:sequence/></xs:extension></xs:complexContent></xs:complexType>`,
			[]string{"src-ct"}},
		{"mixed disagrees between complexType and complexContent",
			`<xs:complexType name="c" mixed="true"><xs:complexContent mixed="false"><xs:restriction base="xs:anyType"><xs:sequence/></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"src-ct"}},
		{"element content extension of a simpleContent type",
			`<xs:complexType name="sc"><xs:simpleContent><xs:extension base="xs:int"/></xs:simpleContent></xs:complexType>
			 <xs:complexType name="c"><xs:complexContent><xs:extension base="tns:sc"><xs:sequence><xs:element name="x" type="xs:int"/></xs:sequence></xs:extension></xs:complexContent></xs:complexType>`,
			[]string{"cos-ct-extends"}},
		{"extension of a complex type final for extension",
			`<xs:complexType name="b" final="extension"><xs:sequence/></xs:complexType>
			 <xs:complexType name="d"><xs:complexContent><xs:extension base="tns:b"><xs:sequence/></xs:extension></xs:complexContent></xs:complexType>`,
			[]string{"cos-ct-extends"}},
		{"restriction of a complex type final for restriction",
			`<xs:complexType name="b" final="restriction"><xs:sequence/></xs:complexType>
			 <xs:complexType name="d"><xs:complexContent><xs:restriction base="tns:b"><xs:sequence/></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"derivation-ok-restriction"}},
		{"restriction of a simple type final for restriction",
			`<xs:simpleType name="b" final="restriction"><xs:restriction base="xs:int"/></xs:simpleType>
			 <xs:simpleType name="d"><xs:restriction base="tns:b"><xs:minInclusive value="0"/></xs:restriction></xs:simpleType>`,
			[]string{"derivation-ok-restriction"}},
		{"simpleContent extension of a simple type final for extension",
			`<xs:simpleType name="b" final="extension"><xs:restriction base="xs:int"/></xs:simpleType>
			 <xs:complexType name="d"><xs:simpleContent><xs:extension base="tns:b"><xs:attribute name="a" type="xs:string"/></xs:extension></xs:simpleContent></xs:complexType>`,
			[]string{"cos-ct-extends"}},

		// Element Declarations Consistent (cos-element-consistent).
		{"extension reintroduces a base element name with a different type",
			`<xs:complexType name="b"><xs:sequence><xs:element name="x" type="xs:int"/></xs:sequence></xs:complexType>
			 <xs:complexType name="d"><xs:complexContent><xs:extension base="tns:b"><xs:sequence><xs:element name="x" type="xs:string"/></xs:sequence></xs:extension></xs:complexContent></xs:complexType>`,
			[]string{"cos-element-consistent"}},
		{"same element name with different types in one sequence",
			`<xs:complexType name="c"><xs:sequence><xs:element name="x" type="xs:int"/><xs:element name="x" type="xs:string"/></xs:sequence></xs:complexType>`,
			[]string{"cos-element-consistent"}},
		{"same element name, same type, differing type tables (one has none)",
			`<xs:complexType name="c"><xs:sequence>
			   <xs:element name="x" type="xs:string"><xs:alternative test="@a='1'" type="xs:token"/></xs:element>
			   <xs:element name="x" type="xs:string"/></xs:sequence></xs:complexType>`,
			[]string{"cos-element-consistent"}},
		{"same element name, same type, differing type-table alternatives",
			`<xs:complexType name="c"><xs:sequence>
			   <xs:element name="x" type="xs:string"><xs:alternative test="@a='1'" type="xs:token"/></xs:element>
			   <xs:element name="x" type="xs:string"><xs:alternative test="@a='2'" type="xs:token"/></xs:element></xs:sequence></xs:complexType>`,
			[]string{"cos-element-consistent"}},
		{"strict wildcard binds a global whose type table differs from a local",
			`<xs:complexType name="c"><xs:sequence>
			   <xs:element name="x" type="xs:string" form="qualified"/>
			   <xs:any namespace="##targetNamespace" processContents="strict"/></xs:sequence></xs:complexType>
			 <xs:element name="x" type="xs:string"><xs:alternative test="@a='1'" type="xs:token"/></xs:element>`,
			[]string{"cos-element-consistent"}},

		// Type alternatives (e-props-correct.7): the alternative's type must be
		// validly substitutable for the element's declared type.
		{"alternative type not derived from the declared type",
			`<xs:element name="e" type="xs:integer"><xs:alternative test="@k='s'" type="xs:string"/></xs:element>`,
			[]string{"e-props-correct"}},
		{"default alternative type not derived from the declared type",
			`<xs:element name="e" type="xs:integer"><xs:alternative type="xs:decimal"/></xs:element>`,
			[]string{"e-props-correct"}},
		{"alternative complex type not derived from the declared complex type",
			`<xs:complexType name="d"><xs:sequence/></xs:complexType>
			 <xs:complexType name="u"><xs:sequence/></xs:complexType>
			 <xs:element name="e" type="tns:d"><xs:alternative test="@k='u'" type="tns:u"/></xs:element>`,
			[]string{"e-props-correct"}},

		// Attribute restriction (derivation-ok-restriction §3.4.6.3 clause 3 /
		// subsumes clause 5.3): the inheritability of a redeclared attribute may
		// not change.
		{"restriction changes attribute inheritability to false",
			`<xs:complexType name="b"><xs:sequence/><xs:attribute name="a" type="xs:string" inheritable="true"/></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:sequence/><xs:attribute name="a" type="xs:string" inheritable="false"/></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"derivation-ok-restriction"}},
		{"restriction changes attribute inheritability to true",
			`<xs:complexType name="b"><xs:sequence/><xs:attribute name="a" type="xs:string"/></xs:complexType>
			 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:sequence/><xs:attribute name="a" type="xs:string" inheritable="true"/></xs:restriction></xs:complexContent></xs:complexType>`,
			[]string{"derivation-ok-restriction"}},

		// Group cycles.
		{"model group cycle",
			`<xs:group name="g1"><xs:sequence><xs:group ref="tns:g2"/></xs:sequence></xs:group>
			 <xs:group name="g2"><xs:sequence><xs:group ref="tns:g1"/></xs:sequence></xs:group>`,
			[]string{"mg-props-correct"}},
		{"model group cycle through a nested model group",
			`<xs:group name="g1"><xs:choice><xs:sequence><xs:group ref="tns:g1"/></xs:sequence></xs:choice></xs:group>`,
			[]string{"mg-props-correct"}},
		{"attribute group cycle",
			`<xs:attributeGroup name="ag1"><xs:attributeGroup ref="tns:ag2"/></xs:attributeGroup>
			 <xs:attributeGroup name="ag2"><xs:attributeGroup ref="tns:ag1"/></xs:attributeGroup>`,
			[]string{"src-attribute_group"}},

		// Identity constraints.
		{"keyref field count differs from its key",
			`<xs:element name="e" type="xs:anyType">
			   <xs:key name="k"><xs:selector xpath="a"/><xs:field xpath="b"/><xs:field xpath="c"/></xs:key>
			   <xs:keyref name="r" refer="tns:k"><xs:selector xpath="a"/><xs:field xpath="b"/></xs:keyref>
			 </xs:element>`,
			[]string{"c-props-correct"}},
		{"keyref refers to another keyref",
			`<xs:element name="e" type="xs:anyType">
			   <xs:key name="k"><xs:selector xpath="a"/><xs:field xpath="b"/></xs:key>
			   <xs:keyref name="r1" refer="tns:r2"><xs:selector xpath="a"/><xs:field xpath="b"/></xs:keyref>
			   <xs:keyref name="r2" refer="tns:k"><xs:selector xpath="a"/><xs:field xpath="b"/></xs:keyref>
			 </xs:element>`,
			[]string{"c-props-correct"}},
		{"unique ref names a key",
			`<xs:element name="e" type="xs:anyType">
			   <xs:key name="k"><xs:selector xpath="a"/><xs:field xpath="b"/></xs:key>
			   <xs:unique ref="tns:k"/>
			 </xs:element>`,
			[]string{"src-identity-constraint"}},

		// Type-alternative {test} XPath validity (src-ta, §3.12.3 clause 1:
		// no XPath 2.0 static errors).
		{"alternative test is malformed XPath",
			`<xs:element name="e" type="xs:string">
			   <xs:alternative test="12 5 2" type="xs:token"/>
			 </xs:element>`,
			[]string{"src-ta"}},
		{"alternative test uses uppercase AND",
			`<xs:element name="e" type="xs:string">
			   <xs:alternative test="@a AND @b" type="xs:token"/>
			 </xs:element>`,
			[]string{"src-ta"}},
		{"alternative test casts to a complex type",
			`<xs:complexType name="ct"><xs:sequence/></xs:complexType>
			 <xs:element name="e" type="xs:string">
			   <xs:alternative test="@x cast as tns:ct = 'y'" type="xs:token"/>
			 </xs:element>`,
			[]string{"src-ta"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, errs := buildAll(t, `<xs:schema `+xmlnsXS+` xmlns:tns="urn:t" targetNamespace="urn:t">`+tc.body+`</xs:schema>`)
			if errs.Empty() {
				t.Fatal("expected errors, got none")
			}
			wantIDs(t, errs, tc.ids...)
		})
	}
}

// A derived bound may equal the corresponding base bound, even when that
// bound is exclusive: the *-valid-restriction rules permit equality (the
// facet value must not be validated against the base's own exclusive bound).
func TestBuildDerivedExclusiveBoundEqualsBase(t *testing.T) {
	_, errs := buildAll(t, `<xs:schema `+xmlnsXS+` xmlns:tns="urn:t" targetNamespace="urn:t">
		<xs:simpleType name="b"><xs:restriction base="xs:int"><xs:maxExclusive value="10"/></xs:restriction></xs:simpleType>
		<xs:simpleType name="d"><xs:restriction base="tns:b"><xs:maxExclusive value="10"/></xs:restriction></xs:simpleType>
	</xs:schema>`)
	wantClean(t, errs)
}

// A top-level attribute inherits the schema's target namespace; when that is
// the XML Schema instance namespace, the declaration is illegal (no-xsi,
// §3.2.6.4) even though there is no explicit targetNamespace attribute to catch
// in pass 1.
func TestBuildNoXsiGlobalAttribute(t *testing.T) {
	_, errs := buildAll(t, `<xs:schema `+xmlnsXS+` targetNamespace="http://www.w3.org/2001/XMLSchema-instance">
		<xs:attribute name="nonStandardAttribute" type="xs:string"/>
	</xs:schema>`)
	if errs.Empty() {
		t.Fatal("expected a no-xsi error, got none")
	}
	wantIDs(t, errs, "no-xsi")
}

func TestBuildSimpleTypeModels(t *testing.T) {
	s, errs := buildAll(t, `<xs:schema `+xmlnsXS+` xmlns:tns="urn:t" targetNamespace="urn:t">
  <xs:simpleType name="size">
    <xs:restriction base="xs:token"><xs:enumeration value="small"/><xs:enumeration value="large"/></xs:restriction>
  </xs:simpleType>
  <xs:simpleType name="sizes"><xs:list itemType="tns:size"/></xs:simpleType>
  <xs:simpleType name="shortSizes"><xs:restriction base="tns:sizes"><xs:maxLength value="3"/></xs:restriction></xs:simpleType>
  <xs:simpleType name="sizeOrInt"><xs:union memberTypes="tns:size xs:int"/></xs:simpleType>
  <xs:simpleType name="extended"><xs:union memberTypes="tns:sizeOrInt xs:date"/></xs:simpleType>
</xs:schema>`)
	wantClean(t, errs)
	st := func(local string) *xsd.SimpleType {
		t.Helper()
		v, ok := s.Types[xsd.QName{Namespace: "urn:t", Local: local}].(*xsd.SimpleType)
		if !ok {
			t.Fatalf("simple type %s missing", local)
		}
		return v
	}

	size := st("size")
	if !size.Facets.HasEnumeration || len(size.Facets.Enumeration) != 2 || size.Facets.Enumeration[0].Lexical != "small" {
		t.Errorf("size enumeration not built: %+v", size.Facets.Enumeration)
	}

	sizes := st("sizes")
	if sizes.Variety != xsd.VarietyList || sizes.ItemType != size {
		t.Errorf("sizes: variety=%v itemType=%v, want list of size", sizes.Variety, sizes.ItemType)
	}
	if sizes.Facets.WhiteSpace != xsd.WSCollapse || !sizes.Facets.WhiteSpaceFixed {
		t.Errorf("list whiteSpace = %v (fixed=%v), want fixed collapse", sizes.Facets.WhiteSpace, sizes.Facets.WhiteSpaceFixed)
	}

	short := st("shortSizes")
	if short.Variety != xsd.VarietyList || short.ItemType != size {
		t.Errorf("shortSizes did not inherit the list variety/item type")
	}
	if short.Facets.MaxLength == nil || short.Facets.MaxLength.Value != 3 {
		t.Errorf("shortSizes maxLength = %+v, want 3", short.Facets.MaxLength)
	}
	if short.Facets.WhiteSpace != xsd.WSCollapse {
		t.Errorf("shortSizes lost the inherited whiteSpace facet")
	}

	union := st("sizeOrInt")
	if union.Variety != xsd.VarietyUnion || len(union.BasicMembers()) != 2 || union.BasicMembers()[0] != size {
		t.Errorf("sizeOrInt members = %v", union.BasicMembers())
	}
	// A union member of union variety is flattened into its own members.
	ext := st("extended")
	if len(ext.BasicMembers()) != 3 || ext.BasicMembers()[0] != size {
		t.Errorf("extended members = %v, want [size int date]", ext.BasicMembers())
	}
}

func TestBuildSimpleContentRestrictionFacets(t *testing.T) {
	doc, reg, errs := load(t, kitchenSink)
	s := buildSchema(reg, doc, errs)
	wantClean(t, errs)

	measured, ok := s.Types[xsd.QName{Namespace: "urn:test", Local: "measured"}].(*xsd.ComplexType)
	if !ok {
		t.Fatal("measured type missing")
	}
	sc, ok := measured.Content.(*xsd.SimpleContent)
	if !ok {
		t.Fatalf("measured content = %T, want simple content", measured.Content)
	}
	// Effective facets: totalDigits from the inline effective base, plus the
	// declared minInclusive/fractionDigits.
	f := sc.Type.Facets
	if f.TotalDigits == nil || f.TotalDigits.Value != 5 {
		t.Errorf("totalDigits = %+v, want 5 (from the inline simpleType)", f.TotalDigits)
	}
	if f.MinInclusive == nil || f.MinInclusive.Lexical != "0" {
		t.Errorf("minInclusive = %+v, want 0", f.MinInclusive)
	}
	if f.FractionDigits == nil || f.FractionDigits.Value != 2 {
		t.Errorf("fractionDigits = %+v, want 2", f.FractionDigits)
	}
	// The restriction's own unit use (required) overrides the base's optional
	// one; the base's anyAttribute is inherited.
	if len(measured.AttributeUses) != 1 || !measured.AttributeUses[0].Required ||
		measured.AttributeUses[0].Decl.Name.Local != "unit" {
		t.Errorf("measured attribute uses = %+v, want one required unit", measured.AttributeUses)
	}
	if measured.AttributeWildcard == nil {
		t.Error("measured did not inherit the base attribute wildcard")
	}
}

func TestBuildComplexContentAssembly(t *testing.T) {
	doc, reg, errs := load(t, kitchenSink)
	s := buildSchema(reg, doc, errs)
	wantClean(t, errs)
	qn := func(l string) xsd.QName { return xsd.QName{Namespace: "urn:test", Local: l} }

	base := s.Types[qn("base")].(*xsd.ComplexType)
	bec, ok := base.Content.(*xsd.ElementContent)
	if !ok || !bec.Mixed {
		t.Fatalf("base content = %+v, want mixed element content", base.Content)
	}
	if bec.OpenContent == nil || bec.OpenContent.Mode != xsd.OpenContentInterleave {
		t.Errorf("base openContent = %+v, want interleave", bec.OpenContent)
	}
	seq, ok := bec.Particle.Term.(*xsd.ModelGroup)
	if !ok || seq.Compositor != xsd.CompositorSequence || len(seq.Particles) != 3 {
		t.Fatalf("base particle = %+v, want sequence of 3", bec.Particle.Term)
	}
	if a := seq.Particles[0]; a.MinOccurs != 0 || a.MaxOccurs != xsd.UnboundedOccurs {
		t.Errorf("element a occurs = %d..%d, want 0..unbounded", a.MinOccurs, a.MaxOccurs)
	}
	choice := seq.Particles[1].Term.(*xsd.ModelGroup)
	if choice.Compositor != xsd.CompositorChoice || len(choice.Particles) != 2 {
		t.Fatalf("choice = %+v", choice)
	}
	if ref, ok := choice.Particles[0].Term.(*xsd.ElementDecl); !ok || ref != s.Elements[qn("doc")] {
		t.Errorf("element ref does not share the global declaration")
	}
	wc, ok := choice.Particles[1].Term.(*xsd.Wildcard)
	if !ok || wc.Mode != xsd.NSConstraintNot {
		t.Fatalf("wildcard = %+v, want notNamespace constraint", choice.Particles[1].Term)
	}
	if len(wc.Namespaces) != 2 || wc.Namespaces[0] != "urn:test" || wc.Namespaces[1] != "" {
		t.Errorf("wildcard namespaces = %v, want [urn:test, absent]", wc.Namespaces)
	}
	if len(wc.NotQName) != 1 || wc.NotQName[0].Local != "##defined" {
		t.Errorf("wildcard notQName = %v, want the ##defined sentinel", wc.NotQName)
	}
	if gr, ok := seq.Particles[2].Term.(*xsd.GroupRef); !ok || gr.Ref != s.Groups[qn("body")] || seq.Particles[2].MaxOccurs != 2 {
		t.Errorf("group particle = %+v, want ref to body with maxOccurs 2", seq.Particles[2])
	}

	derived := s.Types[qn("derived")].(*xsd.ComplexType)
	dec := derived.Content.(*xsd.ElementContent)
	if !dec.Mixed {
		t.Error("derived must be mixed (complexContent mixed=\"true\", matching the mixed base per cos-ct-extends.1.4.3.2.2.1)")
	}
	// Extension particle = sequence(base particle, own particle), sharing the
	// base's particle component.
	dseq, ok := dec.Particle.Term.(*xsd.ModelGroup)
	if !ok || dseq.Compositor != xsd.CompositorSequence || len(dseq.Particles) != 2 || dseq.Particles[0] != bec.Particle {
		t.Fatalf("derived particle does not prepend the base particle: %+v", dec.Particle.Term)
	}
	// No explicit openContent on derived: the schema's defaultOpenContent
	// (mode=interleave) applies. (The default is interleave, not suffix,
	// because base's open content is interleave and an extension may not
	// narrow it to suffix — cos-ct-extends §3.4.6.2 clause 1.4.3.2.2.3.)
	if dec.OpenContent == nil || dec.OpenContent.Mode != xsd.OpenContentInterleave {
		t.Errorf("derived openContent = %+v, want the interleave defaultOpenContent", dec.OpenContent)
	}
	// Extension attribute uses = own (version) + inherited (id, lang, globalAttr).
	if len(derived.AttributeUses) != 4 {
		t.Errorf("derived has %d attribute uses, want 4", len(derived.AttributeUses))
	}
	if derived.AttributeWildcard == nil || derived.AttributeWildcard.Mode != xsd.NSConstraintNot {
		t.Errorf("derived attribute wildcard = %+v, want the base's anyAttribute", derived.AttributeWildcard)
	}

	// Identity constraints: refer linkage and component sharing via ref=.
	docElem := s.Elements[qn("doc")]
	ics := docElem.IdentityConstraints
	if len(ics) != 3 {
		t.Fatalf("doc identity constraints = %d, want 3", len(ics))
	}
	if ics[1].Category != xsd.ICKeyref || ics[1].Refer != ics[0] {
		t.Errorf("docRef.Refer not linked to docKey")
	}
	if ics[2] != ics[0] {
		t.Error("key ref= form must share the referenced component")
	}
}

func TestBuildSubstitutionTypeFallback(t *testing.T) {
	s, errs := buildAll(t, `<xs:schema `+xmlnsXS+` xmlns:tns="urn:t" targetNamespace="urn:t">
  <xs:element name="head" type="xs:int"/>
  <xs:element name="member" substitutionGroup="tns:head"/>
</xs:schema>`)
	wantClean(t, errs)
	head := s.Elements[xsd.QName{Namespace: "urn:t", Local: "head"}]
	member := s.Elements[xsd.QName{Namespace: "urn:t", Local: "member"}]
	if len(member.SubstitutionGroups) != 1 || member.SubstitutionGroups[0] != head {
		t.Fatalf("member substitution groups = %v", member.SubstitutionGroups)
	}
	if member.Type != head.Type {
		t.Errorf("member type = %v, want the head's type", member.Type)
	}
}

func TestBuildAttrRestrictionProhibited(t *testing.T) {
	s, errs := buildAll(t, `<xs:schema `+xmlnsXS+` xmlns:tns="urn:t" targetNamespace="urn:t">
  <xs:complexType name="b"><xs:attribute name="a" type="xs:int"/><xs:attribute name="keep" type="xs:int"/></xs:complexType>
  <xs:complexType name="d">
    <xs:complexContent><xs:restriction base="tns:b"><xs:sequence/><xs:attribute name="a" use="prohibited"/></xs:restriction></xs:complexContent>
  </xs:complexType>
</xs:schema>`)
	wantClean(t, errs)
	d := s.Types[xsd.QName{Namespace: "urn:t", Local: "d"}].(*xsd.ComplexType)
	if len(d.AttributeUses) != 1 || d.AttributeUses[0].Decl.Name.Local != "keep" {
		t.Errorf("restricted attribute uses = %+v, want only keep", d.AttributeUses)
	}
}

func TestBuildDefaultAttributesGroup(t *testing.T) {
	t.Run("applied unless the type opts out", func(t *testing.T) {
		s, errs := buildAll(t, `<xs:schema `+xmlnsXS+` xmlns:tns="urn:t" targetNamespace="urn:t" defaultAttributes="tns:g">
  <xs:attributeGroup name="g"><xs:attribute name="ver" type="xs:int"/></xs:attributeGroup>
  <xs:complexType name="c"><xs:sequence/></xs:complexType>
  <xs:complexType name="opt" defaultAttributesApply="false"><xs:sequence/></xs:complexType>
</xs:schema>`)
		wantClean(t, errs)
		c := s.Types[xsd.QName{Namespace: "urn:t", Local: "c"}].(*xsd.ComplexType)
		if len(c.AttributeUses) != 1 || c.AttributeUses[0].Decl.Name.Local != "ver" {
			t.Errorf("c attribute uses = %+v, want the default ver", c.AttributeUses)
		}
		opt := s.Types[xsd.QName{Namespace: "urn:t", Local: "opt"}].(*xsd.ComplexType)
		if len(opt.AttributeUses) != 0 {
			t.Errorf("opt attribute uses = %+v, want none (opted out)", opt.AttributeUses)
		}
	})
	t.Run("undeclared group is src-resolve", func(t *testing.T) {
		_, errs := buildAll(t, `<xs:schema `+xmlnsXS+` xmlns:tns="urn:t" targetNamespace="urn:t" defaultAttributes="tns:nope">
  <xs:complexType name="c"><xs:sequence/></xs:complexType>
</xs:schema>`)
		wantIDs(t, errs, "src-resolve")
	})
	t.Run("collision with an explicit attribute is a duplicate", func(t *testing.T) {
		_, errs := buildAll(t, `<xs:schema `+xmlnsXS+` xmlns:tns="urn:t" targetNamespace="urn:t" defaultAttributes="tns:g">
  <xs:attributeGroup name="g"><xs:attribute name="ver" type="xs:int"/></xs:attributeGroup>
  <xs:complexType name="c"><xs:sequence/><xs:attribute name="ver" type="xs:string"/></xs:complexType>
</xs:schema>`)
		wantIDs(t, errs, "ct-props-correct")
	})
}

func TestBuildUPAValidModels(t *testing.T) {
	// Content models that are deterministic and must NOT trip cos-nonambig.
	cases := []string{
		// Same element twice in sequence (distinct positions, no shared follow).
		`<xs:complexType name="c"><xs:sequence><xs:element name="a" type="xs:int"/><xs:element name="a" type="xs:int"/></xs:sequence></xs:complexType>`,
		// Optional then a different element.
		`<xs:complexType name="c"><xs:sequence><xs:element name="a" type="xs:int" minOccurs="0"/><xs:element name="b" type="xs:int"/></xs:sequence></xs:complexType>`,
		// Element competes with a wildcard: the element wins, no violation.
		`<xs:complexType name="c"><xs:choice><xs:element name="a" type="xs:int"/><xs:any namespace="##any"/></xs:choice></xs:complexType>`,
		// Disjoint wildcards in a choice.
		`<xs:complexType name="c"><xs:choice><xs:any namespace="urn:a"/><xs:any namespace="urn:b"/></xs:choice></xs:complexType>`,
	}
	for _, body := range cases {
		_, errs := buildAll(t, `<xs:schema `+xmlnsXS+` xmlns:tns="urn:t" targetNamespace="urn:t">`+body+`</xs:schema>`)
		wantClean(t, errs)
	}
}

func TestBuildGroupRecursionThroughElement(t *testing.T) {
	// A model group whose content reaches back to itself ONLY through an
	// element declaration's type is NOT a circular group: the element breaks
	// the chain (over030). It must build and must not trip mg-props-correct.
	_, errs := buildAll(t, `<xs:schema `+xmlnsXS+` xmlns:tns="urn:t" targetNamespace="urn:t">
	  <xs:group name="g"><xs:sequence><xs:element ref="tns:e"/></xs:sequence></xs:group>
	  <xs:element name="e" type="tns:t"/>
	  <xs:complexType name="t"><xs:sequence><xs:group ref="tns:g" minOccurs="0"/></xs:sequence></xs:complexType>`+
		`</xs:schema>`)
	wantClean(t, errs)
}

func TestBuildElementConsistentValidModels(t *testing.T) {
	// Content models that must NOT trip cos-element-consistent: a strict/lax
	// wildcard binding a like-named global whose TYPE differs is a dynamic
	// check (cvc-complex-type.5), leaving the schema valid; only a differing
	// TYPE TABLE is a static violation.
	cases := []string{
		// Local element and a wildcard-bound global of the same name differ only
		// in type — neither has a type table (wild061/062/075).
		`<xs:complexType name="c"><xs:sequence><xs:element name="x" type="xs:date" form="qualified"/><xs:any namespace="##targetNamespace" processContents="strict"/></xs:sequence></xs:complexType>
		 <xs:element name="x" type="xs:time"/>`,
		// Same, with processContents=lax.
		`<xs:complexType name="c"><xs:sequence><xs:element name="x" type="xs:date" form="qualified"/><xs:any namespace="##targetNamespace" processContents="lax"/></xs:sequence></xs:complexType>
		 <xs:element name="x" type="xs:time"/>`,
		// A skip wildcard never binds a declaration, so no consistency applies
		// even when type tables would differ.
		`<xs:complexType name="c"><xs:sequence><xs:element name="x" type="xs:string" form="qualified"><xs:alternative test="@a='1'" type="xs:token"/></xs:element><xs:any namespace="##targetNamespace" processContents="skip"/></xs:sequence></xs:complexType>
		 <xs:element name="x" type="xs:string"/>`,
		// Local and wildcard-bound global share an (identical) type table.
		`<xs:complexType name="c"><xs:sequence><xs:element name="x" type="xs:string" form="qualified"><xs:alternative test="@a='1'" type="xs:token"/></xs:element><xs:any namespace="##targetNamespace" processContents="strict"/></xs:sequence></xs:complexType>
		 <xs:element name="x" type="xs:string"><xs:alternative test="@a='1'" type="xs:token"/></xs:element>`,
	}
	for _, body := range cases {
		_, errs := buildAll(t, `<xs:schema `+xmlnsXS+` xmlns:tns="urn:t" targetNamespace="urn:t">`+body+`</xs:schema>`)
		wantClean(t, errs)
	}
}

func TestBuildTypeAlternativeAndAttrRestrictValid(t *testing.T) {
	// Schemas that must NOT trip e-props-correct.7 or the attribute
	// inheritability subsumption (derivation-ok-restriction).
	cases := []string{
		// Alternative type is a restriction of the declared type (token <: string).
		`<xs:element name="e" type="xs:string"><xs:alternative test="@k='t'" type="xs:token"/></xs:element>`,
		// No declared type ⇒ {type definition} is xs:anyType ⇒ any alternative
		// is substitutable.
		`<xs:element name="e"><xs:alternative test="@k='i'" type="xs:int"/><xs:alternative type="xs:string"/></xs:element>`,
		// Declared type is a union; an alternative naming a member of the union
		// is validly derived from it (cos-st-derived-ok clause 2.2.4).
		`<xs:simpleType name="u"><xs:union memberTypes="xs:int xs:date"/></xs:simpleType>
		 <xs:element name="e" type="tns:u"><xs:alternative test="@k='i'" type="xs:int"/></xs:element>`,
		// xs:error is always permitted as an alternative type (clause 7.2).
		`<xs:element name="e" type="xs:integer"><xs:alternative test="@k='x'" type="xs:error"/></xs:element>`,
		// Well-formed XPath {test} expressions must not trip src-ta: a cast to a
		// simple type, an instance-of test, and a function call are all valid.
		`<xs:element name="e" type="xs:integer">
		   <xs:alternative test="@k cast as xs:int gt 0" type="xs:nonNegativeInteger"/>
		   <xs:alternative test="@k instance of xs:int and count(.//x) eq 1" type="xs:int"/>
		 </xs:element>`,
		// A cast to a user-declared SIMPLE type is fine (only a complex target
		// is a src-ta violation).
		`<xs:simpleType name="st"><xs:restriction base="xs:int"/></xs:simpleType>
		 <xs:element name="e" type="xs:integer"><xs:alternative test="@k cast as tns:st = 1" type="xs:int"/></xs:element>`,
		// Restriction redeclares an attribute keeping the same inheritability.
		`<xs:complexType name="b"><xs:sequence/><xs:attribute name="a" type="xs:string" inheritable="true"/></xs:complexType>
		 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:sequence/><xs:attribute name="a" type="xs:token" inheritable="true"/></xs:restriction></xs:complexContent></xs:complexType>`,
		// Extension keeps the base's interleave open content as interleave
		// (widening the wildcard namespace is fine — cos-ct-extends 1.4.3.2.2.4
		// is met by the wildcard union).
		`<xs:complexType name="b"><xs:openContent mode="interleave"><xs:any namespace="urn:o"/></xs:openContent><xs:sequence><xs:element name="a" type="xs:int"/></xs:sequence></xs:complexType>
		 <xs:complexType name="r"><xs:complexContent><xs:extension base="tns:b"><xs:openContent mode="interleave"><xs:any namespace="urn:p"/></xs:openContent><xs:sequence/></xs:extension></xs:complexContent></xs:complexType>`,
		// Extension may widen suffix to interleave (the reverse narrowing is
		// the only thing barred); base suffix + extension interleave is valid.
		`<xs:complexType name="b"><xs:openContent mode="suffix"><xs:any namespace="urn:o"/></xs:openContent><xs:sequence><xs:element name="a" type="xs:int"/></xs:sequence></xs:complexType>
		 <xs:complexType name="r"><xs:complexContent><xs:extension base="tns:b"><xs:openContent mode="interleave"><xs:any namespace="urn:o"/></xs:openContent><xs:sequence/></xs:extension></xs:complexContent></xs:complexType>`,
	}
	for _, body := range cases {
		_, errs := buildAll(t, `<xs:schema `+xmlnsXS+` xmlns:tns="urn:t" targetNamespace="urn:t">`+body+`</xs:schema>`)
		wantClean(t, errs)
	}
}

func TestBuildParticleRestrictValidModels(t *testing.T) {
	// Restrictions that ARE valid and must NOT trip cos-particle-restrict.
	cases := []string{
		// All-to-all: narrows every occurrence range, drops an optional element.
		`<xs:complexType name="b"><xs:all><xs:element name="a" type="xs:int" minOccurs="0" maxOccurs="5"/><xs:element name="d" type="xs:int" minOccurs="1"/></xs:all></xs:complexType>
		 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:all><xs:element name="d" type="xs:int" minOccurs="1"/><xs:element name="a" type="xs:int" minOccurs="1" maxOccurs="3"/></xs:all></xs:restriction></xs:complexContent></xs:complexType>`,
		// Sequence-to-sequence: narrows the child type by restriction.
		`<xs:complexType name="b"><xs:sequence><xs:element name="n" type="xs:integer"/></xs:sequence></xs:complexType>
		 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:sequence><xs:element name="n" type="xs:positiveInteger"/></xs:sequence></xs:restriction></xs:complexContent></xs:complexType>`,
		// Restriction reproduces the base element's type table unchanged
		// (subsumes clause 4.6 is satisfied by an equivalent type table).
		`<xs:complexType name="b"><xs:sequence><xs:element name="s" type="xs:string"><xs:alternative test="@k='1'" type="xs:token"/></xs:element></xs:sequence></xs:complexType>
		 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:sequence><xs:element name="s" type="xs:string"><xs:alternative test="@k='1'" type="xs:token"/></xs:element></xs:sequence></xs:restriction></xs:complexContent></xs:complexType>`,
		// Substitution-group split whose summed occurrences stay within range.
		`<xs:element name="a" type="xs:int"/>
		 <xs:element name="A1" type="xs:int" substitutionGroup="tns:a"/>
		 <xs:element name="A2" type="xs:int" substitutionGroup="tns:a"/>
		 <xs:complexType name="b"><xs:all><xs:element ref="tns:a" minOccurs="4" maxOccurs="20"/></xs:all></xs:complexType>
		 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:all><xs:element ref="tns:A1" minOccurs="2" maxOccurs="8"/><xs:element ref="tns:A2" minOccurs="2" maxOccurs="8"/></xs:all></xs:restriction></xs:complexContent></xs:complexType>`,
		// Union-typed child restricted to one of the union's members: validly
		// derived because the facet-free union imposes no intervening facet
		// (cos-st-derived-ok clause 2.2.4 holds — the positive of simple011/014/015).
		`<xs:simpleType name="u"><xs:union memberTypes="xs:date xs:time"/></xs:simpleType>
		 <xs:complexType name="b"><xs:sequence><xs:element name="c" type="tns:u"/></xs:sequence></xs:complexType>
		 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:sequence><xs:element name="c" type="xs:date"/></xs:sequence></xs:restriction></xs:complexContent></xs:complexType>`,
		// Member reached through a facet-free intervening union: still validly
		// derived (clause 2.2.4.3 holds because no union on the path carries a facet).
		`<xs:simpleType name="inner"><xs:union memberTypes="xs:date xs:dateTime"/></xs:simpleType>
		 <xs:simpleType name="outer"><xs:union memberTypes="tns:inner xs:time"/></xs:simpleType>
		 <xs:complexType name="b"><xs:sequence><xs:element name="c" type="tns:outer"/></xs:sequence></xs:complexType>
		 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:sequence><xs:element name="c" type="xs:date"/></xs:sequence></xs:restriction></xs:complexContent></xs:complexType>`,
		// Attribute wildcard narrowed to a subset (notNamespace widened set).
		`<xs:complexType name="b"><xs:sequence/><xs:anyAttribute notNamespace="urn:x"/></xs:complexType>
		 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:sequence/><xs:anyAttribute notNamespace="urn:x urn:y"/></xs:restriction></xs:complexContent></xs:complexType>`,
		// Attribute wildcard dropped entirely by the restriction (allowed).
		`<xs:complexType name="b"><xs:sequence/><xs:anyAttribute namespace="##any"/></xs:complexType>
		 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:sequence/></xs:restriction></xs:complexContent></xs:complexType>`,
		// Content wildcard narrowed to a subset namespace (NSSubset).
		`<xs:complexType name="b"><xs:all><xs:any namespace="urn:x urn:y"/></xs:all></xs:complexType>
		 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:all><xs:any namespace="urn:x"/></xs:all></xs:restriction></xs:complexContent></xs:complexType>`,
		// Base wildcard replaced by a concrete element it admits (NSCompat).
		`<xs:complexType name="b"><xs:sequence><xs:any namespace="##any" maxOccurs="3"/></xs:sequence></xs:complexType>
		 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:sequence><xs:element name="e" type="xs:int"/></xs:sequence></xs:restriction></xs:complexContent></xs:complexType>`,
		// All restricted by a choice of branches, each within the base ranges.
		`<xs:complexType name="b"><xs:all><xs:element name="a" type="xs:int" minOccurs="0" maxOccurs="3"/><xs:element name="bb" type="xs:int" minOccurs="0" maxOccurs="3"/></xs:all></xs:complexType>
		 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:choice><xs:element name="a" type="xs:int" minOccurs="1" maxOccurs="2"/><xs:sequence><xs:element name="bb" type="xs:int" minOccurs="1" maxOccurs="2"/></xs:sequence></xs:choice></xs:restriction></xs:complexContent></xs:complexType>`,
		// Open content narrowed: suffix mode kept, wildcard namespace reduced.
		`<xs:complexType name="b"><xs:openContent mode="interleave"><xs:any namespace="urn:o urn:p"/></xs:openContent><xs:sequence><xs:element name="a" type="xs:int"/></xs:sequence></xs:complexType>
		 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:openContent mode="suffix"><xs:any namespace="urn:o"/></xs:openContent><xs:sequence><xs:element name="a" type="xs:int"/></xs:sequence></xs:restriction></xs:complexContent></xs:complexType>`,
		// More-open mode (interleave over suffix) is harmless when the restriction
		// content model is empty: interleave and suffix coincide (open020).
		`<xs:complexType name="b"><xs:openContent mode="suffix"><xs:any namespace="urn:o"/></xs:openContent><xs:sequence><xs:element name="a" type="xs:int" minOccurs="0" maxOccurs="unbounded"/></xs:sequence></xs:complexType>
		 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:openContent mode="interleave"><xs:any namespace="urn:o"/></xs:openContent><xs:sequence/></xs:restriction></xs:complexContent></xs:complexType>`,
		// Base has no open content but its content model is a wildcard that absorbs
		// the restriction's open-content elements, so it stays valid (open022).
		`<xs:complexType name="b"><xs:sequence><xs:any namespace="urn:o" processContents="lax" minOccurs="0" maxOccurs="unbounded"/></xs:sequence></xs:complexType>
		 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:openContent mode="interleave"><xs:any namespace="urn:o"/></xs:openContent><xs:sequence/></xs:restriction></xs:complexContent></xs:complexType>`,
		// Two disjoint base wildcards; a restriction wildcard straddles them yet
		// the summed counts stay within both regions (multi-wildcard packing).
		`<xs:complexType name="b"><xs:all><xs:any namespace="urn:x urn:y" minOccurs="0" maxOccurs="4"/><xs:any namespace="urn:z" minOccurs="0" maxOccurs="4"/></xs:all></xs:complexType>
		 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:all><xs:any namespace="urn:x" minOccurs="0" maxOccurs="2"/><xs:any namespace="urn:y urn:z" minOccurs="0" maxOccurs="2"/></xs:all></xs:restriction></xs:complexContent></xs:complexType>`,
		// Sequence analog of wild069 (saxon wild068, VALID): the named e sits
		// before f, so a later <e> routes to the lax wildcard in BOTH base and
		// restriction — precedence in a sequence is positional, not global, so
		// dropping the named e is a sound restriction. The shadow check must not
		// fire here (it is gated to <all> bases).
		`<xs:complexType name="b"><xs:sequence><xs:element name="e" form="qualified" minOccurs="0"><xs:simpleType><xs:union memberTypes="xs:date xs:time"/></xs:simpleType></xs:element><xs:element name="f" type="xs:integer"/><xs:any namespace="##targetNamespace" processContents="lax"/></xs:sequence></xs:complexType>
		 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:sequence><xs:element name="f" type="xs:integer"/><xs:any namespace="##targetNamespace" processContents="lax"/></xs:sequence></xs:restriction></xs:complexContent></xs:complexType>
		 <xs:element name="e" type="xs:duration"/>`,
		// All-group analog where the wildcard binds a global VALIDLY derived from
		// the base named type (xs:date restricts the union(date,time)), so dropping
		// the named e stays sound — the shadow check accepts it.
		`<xs:complexType name="b"><xs:all><xs:element name="e" form="qualified" minOccurs="0"><xs:simpleType><xs:union memberTypes="xs:date xs:time"/></xs:simpleType></xs:element><xs:element name="f" type="xs:integer"/><xs:any namespace="##targetNamespace" processContents="lax"/></xs:all></xs:complexType>
		 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:all><xs:element name="f" type="xs:integer"/><xs:any namespace="##targetNamespace" processContents="lax"/></xs:all></xs:restriction></xs:complexContent></xs:complexType>
		 <xs:element name="e" type="xs:date"/>`,
		// All-group where the shadowed base element is anyType: even an
		// unconstrained (skip) wildcard binding cannot exceed it, so it stays valid.
		`<xs:complexType name="b"><xs:all><xs:element name="e" form="qualified" type="xs:anyType" minOccurs="0"/><xs:element name="f" type="xs:integer"/><xs:any namespace="##targetNamespace" processContents="skip"/></xs:all></xs:complexType>
		 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:all><xs:element name="f" type="xs:integer"/><xs:any namespace="##targetNamespace" processContents="skip"/></xs:all></xs:restriction></xs:complexContent></xs:complexType>`,
		// A base wildcard overlaps the base element beside it (legal in an <all>),
		// so the regions are not disjoint and only the always-sound outside-name
		// check runs; the restriction wildcard excludes more than the base's, so
		// it admits nothing new (wild047 shape).
		`<xs:complexType name="b"><xs:all><xs:element name="nm" type="xs:string"/><xs:any namespace="##local" notQName="a b c" minOccurs="0" maxOccurs="2" processContents="skip"/><xs:any namespace="urn:x" minOccurs="0" maxOccurs="2" processContents="skip"/></xs:all></xs:complexType>
		 <xs:complexType name="r"><xs:complexContent><xs:restriction base="tns:b"><xs:sequence><xs:element name="nm" type="xs:string"/><xs:any namespace="##local" notQName="a b c d e" minOccurs="1" maxOccurs="1" processContents="skip"/></xs:sequence></xs:restriction></xs:complexContent></xs:complexType>`,
	}
	for _, body := range cases {
		_, errs := buildAll(t, `<xs:schema `+xmlnsXS+` xmlns:tns="urn:t" targetNamespace="urn:t">`+body+`</xs:schema>`)
		wantClean(t, errs)
	}
}

func TestBuildMaxOccursZeroParticle(t *testing.T) {
	s, errs := buildAll(t, `<xs:schema `+xmlnsXS+` xmlns:tns="urn:t" targetNamespace="urn:t">
  <xs:complexType name="c"><xs:sequence><xs:element name="x" type="xs:int" minOccurs="0" maxOccurs="0"/></xs:sequence></xs:complexType>
</xs:schema>`)
	wantClean(t, errs)
	c := s.Types[xsd.QName{Namespace: "urn:t", Local: "c"}].(*xsd.ComplexType)
	mg := c.Content.(*xsd.ElementContent).Particle.Term.(*xsd.ModelGroup)
	if len(mg.Particles) != 0 {
		t.Errorf("maxOccurs=0 particle must map to no component; got %+v", mg.Particles)
	}
}
