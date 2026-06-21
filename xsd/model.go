package xsd

// This file is the public schema component model, mirroring the component
// definitions of XSD 1.1 Part 1 §3 and Part 2 §3–4. Pure data: nothing here
// depends on the parser.

// Form is the value of elementFormDefault/attributeFormDefault and form.
type Form int

const (
	FormUnqualified Form = iota
	FormQualified
)

// Derivation is one method of type derivation.
type Derivation int

const (
	DeriveRestriction Derivation = 1 << iota
	DeriveExtension
	DeriveList
	DeriveUnion
	DeriveSubstitution // only meaningful in element block/final sets
)

// DerivationSet is a set of Derivation methods (block/final attributes).
type DerivationSet int

func (s DerivationSet) Has(d Derivation) bool { return int(s)&int(d) != 0 }

// AllDerivations is the expansion of #all.
const AllDerivations = DerivationSet(DeriveRestriction | DeriveExtension | DeriveList | DeriveUnion | DeriveSubstitution)

// Variety classifies a simple type definition (Part 2 §4.1.1).
type Variety int

const (
	VarietyAtomic Variety = iota
	VarietyList
	VarietyUnion
)

func (v Variety) String() string {
	switch v {
	case VarietyList:
		return "list"
	case VarietyUnion:
		return "union"
	default:
		return "atomic"
	}
}

// Type is a type definition: *SimpleType or *ComplexType.
type Type interface {
	TypeName() QName
	TypePos() Pos
	// Base returns the base type definition (nil only for xs:anyType).
	Base() Type
	isType()
}

// ForeignAttr is an attribute from a non-XSD namespace, preserved verbatim.
type ForeignAttr struct {
	Name  QName
	Value string
}

// ForeignNode is an element from a non-XSD namespace, preserved verbatim
// (appinfo/documentation content, vendor extensions, …).
type ForeignNode struct {
	Name     QName
	Attrs    []ForeignAttr
	Children []*ForeignNode
	CharData string
}

// Extensions carries the foreign attributes and elements found on a schema
// component, so unknown content survives a parse/mutate round trip.
type Extensions struct {
	Attrs []ForeignAttr
	Nodes []*ForeignNode
}

// Annotation is the content of an xs:annotation child.
type Annotation struct {
	Documentation []string
	AppInfo       []*ForeignNode
}

// Schema is one schema (one target namespace's worth of components as
// assembled from a schema document plus its includes/overrides/redefines).
type Schema struct {
	TargetNamespace string
	Location        string // URI of the defining document
	Pos             Pos

	Version                string
	ElementFormDefault     Form
	AttributeFormDefault   Form
	BlockDefault           DerivationSet
	FinalDefault           DerivationSet
	DefaultAttributesGroup QName // defaultAttributes attribute, if any

	Types           map[QName]Type
	Elements        map[QName]*ElementDecl
	Attributes      map[QName]*AttributeDecl
	Groups          map[QName]*Group
	AttributeGroups map[QName]*AttributeGroup
	Notations       map[QName]*Notation

	// Imports records the namespaces imported by this schema.
	Imports []string

	Annotations []*Annotation
	Extensions  Extensions
}

// SimpleType is a simple type definition (Part 2 §4.1). The facet engine
// hooks (ParseValue/Compare overrides, Primitive) are defined in the value
// and facet files of this package.
type SimpleType struct {
	Name QName
	Pos  Pos

	BaseType Type // *SimpleType, except xs:anySimpleType whose base is xs:anyType
	Variety  Variety

	// ItemType is the item type for list varieties.
	ItemType *SimpleType
	// DirectMembers is the spec {member type definitions} property of a union
	// (Part 2 §4.1.1): the member types exactly as declared, WITHOUT flattening
	// member unions. A restriction of a union inherits its base's DirectMembers.
	// This preserves the union-nesting needed to compute transitive membership
	// and the intervening unions of cos-st-derived-ok clause 2.2.4.
	DirectMembers []*SimpleType

	Final DerivationSet

	// DeclaredFacets are only the facets declared on this derivation step —
	// the canonical, authored facet state. The effective (merged, narrowing-
	// validated) facets are derived on demand from these plus the base chain by
	// EffectiveFacets(); they are not stored.
	DeclaredFacets Facets

	// Parse overrides lexical→value mapping; if nil, resolution walks to
	// the nearest ancestor that defines one (ultimately the primitive).
	Parse ParseFunc
	// Compare overrides value comparison, with the same inheritance rule.
	Compare CompareFunc

	// Applicable is the set of constraining facets this type's value space
	// admits (cos-applicable-facets). It is authored only on primitive types
	// (including custom ones registered through the API); derived, list, and
	// union types compute their applicable set from it via ApplicableFacets.
	// Because only primitives carry it, a non-zero Applicable is also what marks
	// a type as the primitive in its chain — PrimitiveType() detects it this way.
	Applicable FacetSet

	// FundamentalBase is the authored base case of the fundamental facets (Part 2
	// §F.1) — the primitive's own {ordered}/{bounded}/{cardinality}/{numeric}.
	// Like Applicable it is authored only on primitives (in builtin) and is nil
	// elsewhere; Fundamentals() copies it off PrimitiveType() and then recomputes
	// {bounded}/{cardinality} from the effective facets. Every primitive is
	// unbounded and countably infinite, so a primitive's stored {bounded}=false /
	// {cardinality}=CountablyInfinite are true of the primitive itself, not stale
	// derived state. Treated as immutable: Fundamentals() copies by value, so one
	// preset may be shared across primitives. list/union and the atomic ur-types
	// (no primitive) leave it nil and fall back to the §F defaults.
	FundamentalBase *Fundamentals

	// The fundamental facets (Part 2 §F) are not stored: they are derived on
	// demand by Fundamentals() from the variety, the primitive base case, and
	// the effective facets.

	Annotation *Annotation
	Extensions Extensions
}

func (t *SimpleType) TypeName() QName { return t.Name }
func (t *SimpleType) TypePos() Pos    { return t.Pos }
func (t *SimpleType) Base() Type      { return t.BaseType }
func (t *SimpleType) isType()         {}

// PrimitiveType returns the primitive ancestor for an atomic type (itself for a
// primitive), or nil for list/union varieties and for the atomic ur-types
// (anySimpleType/anyAtomicType, which have no primitive). It is a derived view
// of the BaseType chain: the primitive is the nearest atomic ancestor that
// authors an applicable-facet set. Only primitives carry one (see Applicable),
// so a non-zero Applicable marks the primitive boundary; the atomic ur-types
// author none and are not primitives.
func (t *SimpleType) PrimitiveType() *SimpleType {
	if t.Variety != VarietyAtomic {
		return nil
	}
	for cur := t; cur != nil; cur, _ = cur.BaseType.(*SimpleType) {
		if cur.Applicable != 0 {
			return cur
		}
	}
	return nil
}

// ApplicableFacets returns the set of constraining facets that may appear in a
// restriction of t (cos-applicable-facets, Part 2 §4.1.6). List and union
// varieties have fixed sets; an atomic type delegates to its primitive's
// authored Applicable set. The atomic ur-types (anySimpleType/anyAtomicType,
// no primitive) admit everything, deferring the real check to their concrete
// descendants.
func (t *SimpleType) ApplicableFacets() FacetSet {
	switch t.Variety {
	case VarietyList:
		return FacetsCommon | FacetsLength
	case VarietyUnion:
		return FacetPattern | FacetEnumeration | FacetAssertion
	}
	prim := t.PrimitiveType()
	if prim == nil {
		return AllFacets
	}
	return prim.Applicable
}

// BasicMembers returns the *basic members* of a union (Part 2 §4.1.6): the
// member types with every member-union flattened away, so each entry is a
// non-union type. It is a derived view of the canonical DirectMembers (the
// un-flattened {member type definitions}); a non-union type has none.
func (t *SimpleType) BasicMembers() []*SimpleType {
	if t.Variety != VarietyUnion {
		return nil
	}
	var flat []*SimpleType
	for _, m := range t.DirectMembers {
		if m.Variety == VarietyUnion {
			flat = append(flat, m.BasicMembers()...)
		} else {
			flat = append(flat, m)
		}
	}
	return flat
}

// OrderedFacet is the HFP `ordered` fundamental facet.
type OrderedFacet int

const (
	OrderedFalse OrderedFacet = iota
	OrderedPartial
	OrderedTotal
)

// Cardinality is the HFP `cardinality` fundamental facet.
type Cardinality int

const (
	CardinalityCountablyInfinite Cardinality = iota
	CardinalityFinite
)

// Fundamentals are the four fundamental facets of a simple type (Part 2 §F):
// {ordered}, {bounded}, {cardinality}, {numeric}.
type Fundamentals struct {
	Ordered     OrderedFacet
	Bounded     bool
	Cardinality Cardinality
	Numeric     bool
}

// Fundamentals derives the four fundamental facets (Part 2 §F) on demand,
// replacing what were stored, hand-copied fields. {ordered} and {numeric} come
// from the primitive's authored FundamentalBase (the §F defaults — false /
// unbounded / countably infinite — for list/union and the atomic ur-types, which
// have no primitive); {bounded} and {cardinality} are then recomputed from the
// effective facets. The values accumulate down a restriction chain because the
// effective facets are the merged set, so no recursion into the base is needed.
func (t *SimpleType) Fundamentals() Fundamentals {
	f := Fundamentals{Cardinality: CardinalityCountablyInfinite}
	if prim := t.PrimitiveType(); prim != nil && prim.FundamentalBase != nil {
		f = *prim.FundamentalBase
	}
	ef := t.EffectiveFacets()
	// spec: §F — {bounded} is true iff a lower and an upper bound both apply.
	if (ef.MinInclusive != nil || ef.MinExclusive != nil) &&
		(ef.MaxInclusive != nil || ef.MaxExclusive != nil) {
		f.Bounded = true
	}
	// spec: §F — {cardinality} is finite once a value-enumerating, length, or
	// digit-count facet constrains the value space.
	if ef.HasEnumeration() || ef.Length != nil || ef.MaxLength != nil || ef.TotalDigits != nil {
		f.Cardinality = CardinalityFinite
	}
	return f
}

// ComplexType is a complex type definition (Part 1 §3.4).
type ComplexType struct {
	Name QName
	Pos  Pos

	BaseType         Type
	DerivationMethod Derivation // DeriveRestriction or DeriveExtension

	Abstract bool
	Final    DerivationSet
	Block    DerivationSet // prohibitedSubstitutions

	// Content is the content type: *SimpleContent, *ElementContent (with
	// possibly empty particle), or EmptyContent (nil Particle).
	Content Content

	AttributeUses     []*AttributeUse
	AttributeWildcard *Wildcard

	// Assertions are XSD 1.1 xs:assert children, stored unevaluated.
	Assertions []Assertion

	Annotation *Annotation
	Extensions Extensions
}

func (t *ComplexType) TypeName() QName { return t.Name }
func (t *ComplexType) TypePos() Pos    { return t.Pos }
func (t *ComplexType) Base() Type      { return t.BaseType }
func (t *ComplexType) isType()         {}

// IsMixed reports whether t has mixed {content type}. This is a derived view of
// the canonical ElementContent.Mixed: only element content can be mixed
// (SimpleContent and empty/non-ElementContent are never mixed).
func (t *ComplexType) IsMixed() bool {
	ec, ok := t.Content.(*ElementContent)
	return ok && ec.Mixed
}

// Content is the content type of a complex type.
type Content interface{ isContent() }

// SimpleContent: the complex type has character-only content of SimpleType.
type SimpleContent struct {
	Type *SimpleType
}

func (*SimpleContent) isContent() {}

// ElementContent: element-only or mixed content with a particle. A nil
// Particle means empty content.
type ElementContent struct {
	Mixed    bool
	Particle *Particle
	// Open content (XSD 1.1).
	OpenContent *OpenContent
}

func (*ElementContent) isContent() {}

// OpenContent is an XSD 1.1 open content property.
type OpenContent struct {
	Mode     OpenContentMode
	Wildcard *Wildcard
	Pos      Pos
}

type OpenContentMode int

const (
	OpenContentInterleave OpenContentMode = iota
	OpenContentSuffix
	OpenContentNone
)

// Particle is a particle: occurrence range plus a term.
type Particle struct {
	MinOccurs int
	MaxOccurs int // UnboundedOccurs for unbounded
	Term      Term
	Pos       Pos
}

// UnboundedOccurs is the MaxOccurs encoding of maxOccurs="unbounded".
const UnboundedOccurs = -1

// Term is *ElementDecl, *ModelGroup, *GroupRef, or *Wildcard.
type Term interface{ isTerm() }

// Compositor is the kind of a model group.
type Compositor int

const (
	CompositorSequence Compositor = iota
	CompositorChoice
	CompositorAll
)

func (c Compositor) String() string {
	switch c {
	case CompositorChoice:
		return "choice"
	case CompositorAll:
		return "all"
	default:
		return "sequence"
	}
}

// ModelGroup is a sequence/choice/all group.
type ModelGroup struct {
	Compositor Compositor
	Particles  []*Particle
	Pos        Pos
}

func (*ModelGroup) isTerm() {}

// GroupRef is a reference to a named model group definition; kept as a
// distinct term so the model can round-trip group references.
type GroupRef struct {
	Ref *Group
	Pos Pos
}

func (*GroupRef) isTerm() {}

// Group is a named model group definition (Part 1 §3.7).
type Group struct {
	Name       QName
	Pos        Pos
	Group      *ModelGroup
	Annotation *Annotation
	Extensions Extensions
}

// ElementDecl is an element declaration (Part 1 §3.3).
type ElementDecl struct {
	Name   QName
	Pos    Pos
	Global bool
	Form   Form

	Type Type
	// TypeAlternatives are XSD 1.1 conditional type assignments (stored,
	// not evaluated by the parser).
	TypeAlternatives []*TypeAlternative

	Nillable bool
	Abstract bool

	Default *string
	Fixed   *string

	Final DerivationSet // restriction/extension only
	Block DerivationSet // restriction/extension/substitution

	// SubstitutionGroups are the heads this declaration is substitutable
	// for (XSD 1.1 allows several).
	SubstitutionGroups []*ElementDecl

	IdentityConstraints []*IdentityConstraint

	Annotation *Annotation
	Extensions Extensions
}

func (*ElementDecl) isTerm() {}

// TypeAlternative is an XSD 1.1 xs:alternative.
type TypeAlternative struct {
	Test string // XPath; empty on the final default alternative
	Type Type
	Pos  Pos
}

// AttributeDecl is an attribute declaration (Part 1 §3.2).
type AttributeDecl struct {
	Name   QName
	Pos    Pos
	Global bool
	Form   Form

	Type *SimpleType

	Default     *string
	Fixed       *string
	Inheritable bool // XSD 1.1

	Annotation *Annotation
	Extensions Extensions
}

// AttributeUse attaches an attribute declaration to a complex type or
// attribute group (Part 1 §3.5).
type AttributeUse struct {
	Decl     *AttributeDecl
	Required bool
	// Default/Fixed here override the declaration's (for ref= uses).
	Default *string
	Fixed   *string
	// Inheritable is the use's {inheritable} (XSD 1.1): the value declared on
	// the use if present, else inherited from the declaration. It governs
	// whether the attribute is supplied to descendant conditional type
	// assignment / assertions.
	Inheritable bool
	Pos         Pos
}

// AttributeGroup is an attribute group definition (Part 1 §3.6).
type AttributeGroup struct {
	Name       QName
	Pos        Pos
	Uses       []*AttributeUse
	Wildcard   *Wildcard
	Annotation *Annotation
	Extensions Extensions
}

// NamespaceConstraintMode is the variety of a wildcard namespace constraint.
type NamespaceConstraintMode int

const (
	NSConstraintAny NamespaceConstraintMode = iota
	NSConstraintNot
	NSConstraintEnumeration
)

// ProcessContents is a wildcard's processContents property.
type ProcessContents int

const (
	ProcessStrict ProcessContents = iota
	ProcessLax
	ProcessSkip
)

// Wildcard is an xs:any / xs:anyAttribute term (Part 1 §3.10).
type Wildcard struct {
	Mode NamespaceConstraintMode
	// Namespaces lists the (dis)allowed namespaces for Not/Enumeration
	// modes; "" means absent (##local), and ##targetNamespace has already
	// been substituted.
	Namespaces []string
	// NotQName (XSD 1.1): disallowed expanded names; entries with the
	// sentinel locals "##defined"/"##definedSibling" are kept verbatim in
	// the Local field with empty namespace.
	NotQName        []QName
	ProcessContents ProcessContents
	Pos             Pos
	Annotation      *Annotation
	Extensions      Extensions
}

func (*Wildcard) isTerm() {}

// IdentityConstraint is a unique/key/keyref definition (Part 1 §3.11).
type IdentityConstraint struct {
	Name     QName
	Pos      Pos
	Category ICCategory
	Selector string
	Fields   []string
	// NamespaceBindings maps each namespace prefix in scope at the constraint's
	// declaration to its namespace URI (the "" key is the default namespace),
	// used to resolve prefixed name tests in the selector/field XPath subset
	// (§3.11.6). Empty/nil ⇒ name tests match by local name only.
	NamespaceBindings map[string]string
	// Refer is the referenced key for keyrefs.
	Refer      *IdentityConstraint
	Annotation *Annotation
	Extensions Extensions
}

type ICCategory int

const (
	ICUnique ICCategory = iota
	ICKey
	ICKeyref
)

// Notation is a notation declaration (Part 1 §3.14).
type Notation struct {
	Name       QName
	Pos        Pos
	System     string
	Public     string
	Annotation *Annotation
	Extensions Extensions
}

// Assertion is an XSD 1.1 assertion (xs:assert / xs:assertion), stored with
// its XPath text and namespace context but not evaluated by the parser.
type Assertion struct {
	Test string
	// XPathDefaultNamespace as resolved at the assertion.
	DefaultNamespace string
	Pos              Pos
	Extensions       Extensions
}
