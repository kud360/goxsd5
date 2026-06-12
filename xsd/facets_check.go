package xsd

// Intra-facet consistency and restriction-narrowing validation
// (Part 2 §4.3.*.5 "Constraints on … Schema Components" and
// cos-st-restricts / cos-applicable-facets).

// ValidateFacetSet checks intra-facet consistency of an effective facet
// set. cmp compares values in the type's own value space.
func ValidateFacetSet(f *Facets, cmp CompareFunc) error {
	var errs ErrorList

	// spec: length-minLength-maxLength — XSD 1.1 Part 2 §4.3.1.5 (xmlschema11-2.md#length-minLength-maxLength)
	if f.Length != nil && f.MinLength != nil && f.MinLength.Value > f.Length.Value {
		errs.Addf(SpecLengthMinMax, f.Length.Pos, "length %d inconsistent with minLength %d", f.Length.Value, f.MinLength.Value)
	}
	if f.Length != nil && f.MaxLength != nil && f.MaxLength.Value < f.Length.Value {
		errs.Addf(SpecLengthMinMax, f.Length.Pos, "length %d inconsistent with maxLength %d", f.Length.Value, f.MaxLength.Value)
	}
	// spec: minLength-less-than-equal-to-maxLength — XSD 1.1 Part 2 §4.3.2.5 (xmlschema11-2.md#minLength-less-than-equal-to-maxLength)
	if f.MinLength != nil && f.MaxLength != nil && f.MinLength.Value > f.MaxLength.Value {
		errs.Addf(SpecMinLELMaxLength, f.MinLength.Pos, "minLength %d > maxLength %d", f.MinLength.Value, f.MaxLength.Value)
	}

	// spec: minInclusive-minExclusive — XSD 1.1 Part 2 §4.3.10.5 (xmlschema11-2.md#minInclusive-minExclusive)
	if f.MinInclusive != nil && f.MinExclusive != nil {
		errs.Addf(SpecMinInclExclusive, f.MinInclusive.Pos, "both minInclusive and minExclusive specified")
	}
	// spec: maxInclusive-maxExclusive — XSD 1.1 Part 2 §4.3.7.5 (xmlschema11-2.md#maxInclusive-maxExclusive)
	if f.MaxInclusive != nil && f.MaxExclusive != nil {
		errs.Addf(SpecMaxInclExclusive, f.MaxInclusive.Pos, "both maxInclusive and maxExclusive specified")
	}

	cmpBounds := func(a, b *Bound) (Order, bool) {
		if a == nil || b == nil {
			return 0, false
		}
		return cmp(a.Value, b.Value)
	}
	// spec: minInclusive-less-than-equal-to-maxInclusive — XSD 1.1 Part 2 §4.3.10.5 (xmlschema11-2.md#minInclusive-less-than-equal-to-maxInclusive)
	if o, ok := cmpBounds(f.MinInclusive, f.MaxInclusive); ok && o == OrderGreater {
		errs.Addf(SpecMinInclLEMaxIncl, f.MinInclusive.Pos, "minInclusive %s > maxInclusive %s", f.MinInclusive.Lexical, f.MaxInclusive.Lexical)
	}
	// spec: minExclusive-less-than-equal-to-maxExclusive — XSD 1.1 Part 2 §4.3.9.5 (xmlschema11-2.md#minExclusive-less-than-equal-to-maxExclusive)
	if o, ok := cmpBounds(f.MinExclusive, f.MaxExclusive); ok && o == OrderGreater {
		errs.Addf(SpecMinExclLEMaxExcl, f.MinExclusive.Pos, "minExclusive %s > maxExclusive %s", f.MinExclusive.Lexical, f.MaxExclusive.Lexical)
	}
	// spec: minInclusive-less-than-maxExclusive — XSD 1.1 Part 2 §4.3.10.5 (xmlschema11-2.md#minInclusive-less-than-maxExclusive)
	if o, ok := cmpBounds(f.MinInclusive, f.MaxExclusive); ok && o != OrderLess {
		errs.Addf(SpecMinInclLTMaxExcl, f.MinInclusive.Pos, "minInclusive %s >= maxExclusive %s", f.MinInclusive.Lexical, f.MaxExclusive.Lexical)
	}
	// spec: minExclusive-less-than-maxInclusive — XSD 1.1 Part 2 §4.3.9.5 (xmlschema11-2.md#minExclusive-less-than-maxInclusive)
	if o, ok := cmpBounds(f.MinExclusive, f.MaxInclusive); ok && o != OrderLess {
		errs.Addf(SpecMinExclLTMaxIncl, f.MinExclusive.Pos, "minExclusive %s >= maxInclusive %s", f.MinExclusive.Lexical, f.MaxInclusive.Lexical)
	}

	// spec: fractionDigits-less-than-equal-to-totalDigits — XSD 1.1 Part 2 §4.3.12.5 (xmlschema11-2.md#fractionDigits-less-than-equal-to-totalDigits)
	if f.TotalDigits != nil && f.FractionDigits != nil && f.FractionDigits.Value > f.TotalDigits.Value {
		errs.Addf(SpecFracLETotalDigits, f.FractionDigits.Pos, "fractionDigits %d > totalDigits %d", f.FractionDigits.Value, f.TotalDigits.Value)
	}
	return errs.Err()
}

// wsRank orders whiteSpace strength: collapse > replace > preserve.
func wsRank(w WhiteSpace) int {
	switch w {
	case WSPreserve:
		return 0
	case WSReplace:
		return 1
	case WSCollapse:
		return 2
	}
	return -1
}

// CheckFacetRestriction validates the facets declared on one restriction
// step against the base type's effective facets: narrowing only, and fixed
// facets unchanged. cmp compares values (base value space).
func CheckFacetRestriction(declared, base *Facets, cmp CompareFunc) error {
	var errs ErrorList

	checkFixedInt := func(d, b *IntFacet, name string) {
		if d != nil && b != nil && b.Fixed && d.Value != b.Value {
			// spec: fixed-facet-value — XSD 1.1 Part 2 §4.3 (xmlschema11-2.md#fixed-facet-value)
			errs.Addf(SpecFixedFacetValue, d.Pos, "facet %s is fixed to %d in the base type", name, b.Value)
		}
	}

	// spec: length-valid-restriction — XSD 1.1 Part 2 §4.3.1.5 (xmlschema11-2.md#length-valid-restriction)
	if declared.Length != nil && base.Length != nil && declared.Length.Value != base.Length.Value {
		errs.Addf(SpecLengthValidRestriction, declared.Length.Pos, "length %d differs from base length %d", declared.Length.Value, base.Length.Value)
	}
	if declared.Length != nil {
		if base.MinLength != nil && declared.Length.Value < base.MinLength.Value {
			errs.Addf(SpecLengthValidRestriction, declared.Length.Pos, "length %d < base minLength %d", declared.Length.Value, base.MinLength.Value)
		}
		if base.MaxLength != nil && declared.Length.Value > base.MaxLength.Value {
			errs.Addf(SpecLengthValidRestriction, declared.Length.Pos, "length %d > base maxLength %d", declared.Length.Value, base.MaxLength.Value)
		}
	}
	// spec: minLength-valid-restriction — XSD 1.1 Part 2 §4.3.2.5 (xmlschema11-2.md#minLength-valid-restriction)
	if declared.MinLength != nil {
		if base.MinLength != nil && declared.MinLength.Value < base.MinLength.Value {
			errs.Addf(SpecMinLengthValidRestriction, declared.MinLength.Pos, "minLength %d < base minLength %d", declared.MinLength.Value, base.MinLength.Value)
		}
		if base.MaxLength != nil && declared.MinLength.Value > base.MaxLength.Value {
			errs.Addf(SpecMinLengthValidRestriction, declared.MinLength.Pos, "minLength %d > base maxLength %d", declared.MinLength.Value, base.MaxLength.Value)
		}
		if base.Length != nil && declared.MinLength.Value > base.Length.Value {
			errs.Addf(SpecMinLengthValidRestriction, declared.MinLength.Pos, "minLength %d > base length %d", declared.MinLength.Value, base.Length.Value)
		}
		checkFixedInt(declared.MinLength, base.MinLength, "minLength")
	}
	// spec: maxLength-valid-restriction — XSD 1.1 Part 2 §4.3.3.5 (xmlschema11-2.md#maxLength-valid-restriction)
	if declared.MaxLength != nil {
		if base.MaxLength != nil && declared.MaxLength.Value > base.MaxLength.Value {
			errs.Addf(SpecMaxLengthValidRestriction, declared.MaxLength.Pos, "maxLength %d > base maxLength %d", declared.MaxLength.Value, base.MaxLength.Value)
		}
		if base.MinLength != nil && declared.MaxLength.Value < base.MinLength.Value {
			errs.Addf(SpecMaxLengthValidRestriction, declared.MaxLength.Pos, "maxLength %d < base minLength %d", declared.MaxLength.Value, base.MinLength.Value)
		}
		if base.Length != nil && declared.MaxLength.Value < base.Length.Value {
			errs.Addf(SpecMaxLengthValidRestriction, declared.MaxLength.Pos, "maxLength %d < base length %d", declared.MaxLength.Value, base.Length.Value)
		}
		checkFixedInt(declared.MaxLength, base.MaxLength, "maxLength")
	}
	if declared.Length != nil && base.Length != nil {
		checkFixedInt(declared.Length, base.Length, "length")
	}

	// spec: whiteSpace-valid-restriction — XSD 1.1 Part 2 §4.3.6.5 (xmlschema11-2.md#whiteSpace-valid-restriction)
	if declared.WhiteSpace != WSUnset && base.WhiteSpace != WSUnset {
		if wsRank(declared.WhiteSpace) < wsRank(base.WhiteSpace) {
			errs.Addf(SpecWhiteSpaceValidRestriction, declared.WhiteSpacePos, "whiteSpace %s loosens base whiteSpace %s", declared.WhiteSpace, base.WhiteSpace)
		}
		if base.WhiteSpaceFixed && declared.WhiteSpace != base.WhiteSpace {
			errs.Addf(SpecFixedFacetValue, declared.WhiteSpacePos, "facet whiteSpace is fixed to %s in the base type", base.WhiteSpace)
		}
	}

	// Bounds narrowing: each declared bound must lie within the base's
	// effective bounds (incomparable counts as outside). Exclusive-vs-
	// exclusive equality on the same side is the one permitted equality
	// against an exclusive base bound.
	checkBound := func(b *Bound, name string, minSide, incl bool, ref SpecRef, baseSame *Bound) {
		if b == nil {
			return
		}
		if base.MinInclusive != nil {
			if o, ok := cmp(b.Value, base.MinInclusive.Value); !ok || o == OrderLess {
				errs.Addf(ref, b.Pos, "%s %s < base minInclusive %s", name, b.Lexical, base.MinInclusive.Lexical)
			}
		}
		if base.MaxInclusive != nil {
			if o, ok := cmp(b.Value, base.MaxInclusive.Value); !ok || o == OrderGreater {
				errs.Addf(ref, b.Pos, "%s %s > base maxInclusive %s", name, b.Lexical, base.MaxInclusive.Lexical)
			}
		}
		if base.MinExclusive != nil {
			o, ok := cmp(b.Value, base.MinExclusive.Value)
			if !ok || o == OrderLess || (o == OrderEqual && !(minSide && !incl)) {
				errs.Addf(ref, b.Pos, "%s %s <= base minExclusive %s", name, b.Lexical, base.MinExclusive.Lexical)
			}
		}
		if base.MaxExclusive != nil {
			o, ok := cmp(b.Value, base.MaxExclusive.Value)
			if !ok || o == OrderGreater || (o == OrderEqual && !(!minSide && !incl)) {
				errs.Addf(ref, b.Pos, "%s %s >= base maxExclusive %s", name, b.Lexical, base.MaxExclusive.Lexical)
			}
		}
		if baseSame != nil && baseSame.Fixed {
			// spec: fixed-facet-value — XSD 1.1 Part 2 §4.3 (xmlschema11-2.md#fixed-facet-value)
			if o, ok := cmp(b.Value, baseSame.Value); !ok || o != OrderEqual {
				errs.Addf(SpecFixedFacetValue, b.Pos, "facet %s is fixed in the base type", name)
			}
		}
	}
	// spec: minInclusive-valid-restriction — XSD 1.1 Part 2 §4.3.10.5 (xmlschema11-2.md#minInclusive-valid-restriction)
	checkBound(declared.MinInclusive, "minInclusive", true, true, SpecMinInclValidRestriction, base.MinInclusive)
	// spec: maxInclusive-valid-restriction — XSD 1.1 Part 2 §4.3.7.5 (xmlschema11-2.md#maxInclusive-valid-restriction)
	checkBound(declared.MaxInclusive, "maxInclusive", false, true, SpecMaxInclValidRestriction, base.MaxInclusive)
	// spec: minExclusive-valid-restriction — XSD 1.1 Part 2 §4.3.9.5 (xmlschema11-2.md#minExclusive-valid-restriction)
	checkBound(declared.MinExclusive, "minExclusive", true, false, SpecMinExclValidRestriction, base.MinExclusive)
	// spec: maxExclusive-valid-restriction — XSD 1.1 Part 2 §4.3.8.5 (xmlschema11-2.md#maxExclusive-valid-restriction)
	checkBound(declared.MaxExclusive, "maxExclusive", false, false, SpecMaxExclValidRestriction, base.MaxExclusive)

	// spec: totalDigits-valid-restriction — XSD 1.1 Part 2 §4.3.11.5 (xmlschema11-2.md#totalDigits-valid-restriction)
	if declared.TotalDigits != nil && base.TotalDigits != nil && declared.TotalDigits.Value > base.TotalDigits.Value {
		errs.Addf(SpecTotalDigitsValidRestriction, declared.TotalDigits.Pos, "totalDigits %d > base totalDigits %d", declared.TotalDigits.Value, base.TotalDigits.Value)
	}
	checkFixedInt(declared.TotalDigits, base.TotalDigits, "totalDigits")
	// spec: fractionDigits-valid-restriction — XSD 1.1 Part 2 §4.3.12.5 (xmlschema11-2.md#fractionDigits-valid-restriction)
	if declared.FractionDigits != nil {
		if base.FractionDigits != nil && declared.FractionDigits.Value > base.FractionDigits.Value {
			errs.Addf(SpecFractionDigitsValidRestriction, declared.FractionDigits.Pos, "fractionDigits %d > base fractionDigits %d", declared.FractionDigits.Value, base.FractionDigits.Value)
		}
		if base.TotalDigits != nil && declared.FractionDigits.Value > base.TotalDigits.Value {
			errs.Addf(SpecFractionDigitsValidRestriction, declared.FractionDigits.Pos, "fractionDigits %d > base totalDigits %d", declared.FractionDigits.Value, base.TotalDigits.Value)
		}
		checkFixedInt(declared.FractionDigits, base.FractionDigits, "fractionDigits")
	}

	// spec: fixed-facet-value — XSD 1.1 Part 2 §4.3 (xmlschema11-2.md#fixed-facet-value)
	if declared.ExplicitTimezone != ETZUnset && base.ExplicitTimezone != ETZUnset &&
		base.ExplicitTimezoneFixed && declared.ExplicitTimezone != base.ExplicitTimezone {
		errs.Addf(SpecFixedFacetValue, declared.ExplicitTimezonePos, "facet explicitTimezone is fixed in the base type")
	}

	return errs.Err()
}
