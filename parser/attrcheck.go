package parser

// Attribute-value checks for the structural validation table. Each checker
// validates one attribute value against the type the schema for schemas
// gives it. Checkers that parse values needed again later (occurrence
// ranges, derivation sets) are split into parse helpers reused by pass 2.

import (
	"fmt"
	"math/big"
	"slices"
	"strings"

	"github.com/kud360/goxsd5/parser/xmltree"
	"github.com/kud360/goxsd5/xsd"
)

// valueCheck validates one attribute value. n is the owning element, used
// for in-scope namespace resolution of QName-valued attributes.
type valueCheck func(v string, n *xmltree.Node) error

// refError overrides the SpecRef a failed check is reported under; without
// it the walker uses the owning element's blanket src-* constraint.
type refError struct {
	ref xsd.SpecRef
	err error
}

func (e *refError) Error() string { return e.err.Error() }
func (e *refError) Unwrap() error { return e.err }

func vcAny(string, *xmltree.Node) error { return nil }

func vcNCName(v string, _ *xmltree.Node) error {
	if !xmltree.IsNCName(strings.TrimSpace(v)) {
		return fmt.Errorf("%q is not an NCName", v)
	}
	return nil
}

// vcID is the same lexical check as NCName; per-document uniqueness
// (src-id) is enforced by the walker, which sees every id attribute.
var vcID = vcNCName

func vcQName(v string, n *xmltree.Node) error {
	if _, err := n.ResolveQName(strings.TrimSpace(v)); err != nil {
		// spec: src-qname — XSD 1.1 Part 1 §3.15.3 (src-qname)
		return &refError{xsd.SpecSrcQName, err}
	}
	return nil
}

// vcQNameList accepts a whitespace-separated, possibly empty, list of
// resolvable QNames (substitutionGroup, memberTypes).
func vcQNameList(v string, n *xmltree.Node) error {
	for _, tok := range strings.Fields(v) {
		if _, err := n.ResolveQName(tok); err != nil {
			return &refError{xsd.SpecSrcQName, err}
		}
	}
	return nil
}

func vcBool(v string, _ *xmltree.Node) error {
	_, err := parseBool(v)
	return err
}

func parseBool(v string) (bool, error) {
	switch strings.TrimSpace(v) {
	case "true", "1":
		return true, nil
	case "false", "0":
		return false, nil
	}
	return false, fmt.Errorf("%q is not a boolean", v)
}

func vcNonNegInt(v string, _ *xmltree.Node) error {
	_, err := parseNonNegInt(v)
	return err
}

func vcPosInt(v string, _ *xmltree.Node) error {
	i, err := parseNonNegInt(v)
	if err == nil && i == 0 {
		err = fmt.Errorf("%q is not a positive integer", v)
	}
	return err
}

// occursCap saturates absurdly large occurrence values so they fit an int
// without rejecting schemas that are technically valid.
const occursCap = 1 << 31

// parseNonNegInt parses an xs:nonNegativeInteger (optional sign, unbounded
// magnitude), saturating at occursCap.
func parseNonNegInt(v string) (int, error) {
	i, ok := new(big.Int).SetString(strings.TrimSpace(v), 10)
	if !ok {
		return 0, fmt.Errorf("%q is not an integer", v)
	}
	if i.Sign() < 0 {
		return 0, fmt.Errorf("%q is negative", v)
	}
	if !i.IsInt64() || i.Int64() > occursCap {
		return occursCap, nil
	}
	return int(i.Int64()), nil
}

func vcMinOccurs(v string, _ *xmltree.Node) error {
	_, err := parseNonNegInt(v)
	return err
}

func vcMaxOccurs(v string, _ *xmltree.Node) error {
	_, err := parseMaxOccurs(v)
	return err
}

// parseMaxOccurs parses (nonNegativeInteger | unbounded).
func parseMaxOccurs(v string) (int, error) {
	if strings.TrimSpace(v) == "unbounded" {
		return xsd.UnboundedOccurs, nil
	}
	return parseNonNegInt(v)
}

// vcZeroOne is minOccurs/maxOccurs on xs:all, whose range is (0 | 1).
func vcZeroOne(v string, _ *xmltree.Node) error {
	switch strings.TrimSpace(v) {
	case "0", "1":
		return nil
	}
	return fmt.Errorf("%q is not 0 or 1", v)
}

// vcEnum builds a checker for keyword-valued attributes.
func vcEnum(allowed ...string) valueCheck {
	return func(v string, _ *xmltree.Node) error {
		v = strings.TrimSpace(v)
		if slices.Contains(allowed, v) {
			return nil
		}
		return fmt.Errorf("%q is not one of %s", v, strings.Join(allowed, ", "))
	}
}

var derivationKeywords = map[string]xsd.Derivation{
	"restriction":  xsd.DeriveRestriction,
	"extension":    xsd.DeriveExtension,
	"list":         xsd.DeriveList,
	"union":        xsd.DeriveUnion,
	"substitution": xsd.DeriveSubstitution,
}

// parseDerivationSet parses block/final/blockDefault/finalDefault values.
// allowed is the subset of keywords legal for the declaring attribute;
// "#all" expands to exactly that subset.
func parseDerivationSet(v string, allowed xsd.DerivationSet) (xsd.DerivationSet, error) {
	v = strings.TrimSpace(v)
	if v == "#all" {
		return allowed, nil
	}
	var set xsd.DerivationSet
	for _, tok := range strings.Fields(v) {
		d, ok := derivationKeywords[tok]
		if !ok || !allowed.Has(d) {
			return 0, fmt.Errorf("%q is not a valid control keyword here", tok)
		}
		set |= xsd.DerivationSet(d)
	}
	return set, nil
}

func vcDerivSet(allowed xsd.DerivationSet) valueCheck {
	return func(v string, _ *xmltree.Node) error {
		_, err := parseDerivationSet(v, allowed)
		return err
	}
}

// vcWildcardNS checks the namespace attribute of xs:any/xs:anyAttribute:
// ##any or ##other standalone, or a list of (anyURI|##targetNamespace|##local).
func vcWildcardNS(v string, _ *xmltree.Node) error {
	fields := strings.Fields(v)
	for _, tok := range fields {
		if tok == "##any" || tok == "##other" {
			if len(fields) != 1 {
				return fmt.Errorf("%s must stand alone in a namespace constraint", tok)
			}
		} else if strings.HasPrefix(tok, "##") && tok != "##targetNamespace" && tok != "##local" {
			return fmt.Errorf("unknown namespace keyword %q", tok)
		}
	}
	return nil
}

// vcNotNamespace checks notNamespace: a non-empty list with no ##any/##other.
func vcNotNamespace(v string, _ *xmltree.Node) error {
	fields := strings.Fields(v)
	if len(fields) == 0 {
		return fmt.Errorf("notNamespace must not be empty")
	}
	for _, tok := range fields {
		if strings.HasPrefix(tok, "##") && tok != "##targetNamespace" && tok != "##local" {
			return fmt.Errorf("unknown namespace keyword %q", tok)
		}
	}
	return nil
}

// vcNotQName checks notQName: a list of QNames plus ##defined and, on
// xs:any only, ##definedSibling.
func vcNotQName(allowSibling bool) valueCheck {
	return func(v string, n *xmltree.Node) error {
		for _, tok := range strings.Fields(v) {
			if tok == "##defined" || (allowSibling && tok == "##definedSibling") {
				continue
			}
			if strings.HasPrefix(tok, "##") {
				return fmt.Errorf("unknown keyword %q in notQName", tok)
			}
			if _, err := n.ResolveQName(tok); err != nil {
				return err
			}
		}
		return nil
	}
}

// vcXPath only requires a non-empty expression; XPath syntax is not parsed
// (assertions and CTA are stored, never evaluated — see NOTES.md).
func vcXPath(v string, _ *xmltree.Node) error {
	if strings.TrimSpace(v) == "" {
		return fmt.Errorf("XPath expression must not be empty")
	}
	return nil
}
