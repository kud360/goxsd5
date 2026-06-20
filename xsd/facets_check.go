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
// facets unchanged. cmp compares values (base value space). Each facet family
// is checked by its own helper; see the *-valid-restriction rules in
// XSD 1.1 Part 2 §4.3.
func CheckFacetRestriction(declared, base *Facets, cmp CompareFunc) error {
	var errs ErrorList
	checkLengthRestriction(&errs, declared, base)
	checkWhiteSpaceRestriction(&errs, declared, base)
	checkBoundsRestriction(&errs, declared, base, cmp)
	checkDigitsRestriction(&errs, declared, base)
	checkTimezoneRestriction(&errs, declared, base)
	return errs.Err()
}

// checkFixedInt reports when a declared integer facet differs from a base facet
// whose value the base fixed.
// spec: fixed-facet-value — XSD 1.1 Part 2 §4.3 (xmlschema11-2.md#fixed-facet-value)
func checkFixedInt(errs *ErrorList, d, b *IntFacet, name string) {
	if d != nil && b != nil && b.Fixed && d.Value != b.Value {
		errs.Addf(SpecFixedFacetValue, d.Pos, "facet %s is fixed to %d in the base type", name, b.Value)
	}
}

// checkLengthRestriction enforces length/minLength/maxLength narrowing.
// spec: {length,minLength,maxLength}-valid-restriction — XSD 1.1 Part 2 §4.3.{1,2,3}.5
func checkLengthRestriction(errs *ErrorList, declared, base *Facets) {
	checkLengthFacet(errs, declared, base)
	checkMinLengthFacet(errs, declared, base)
	checkMaxLengthFacet(errs, declared, base)
}

// checkLengthFacet enforces the declared length against the base length and
// length range, plus the fixed-facet rule.
func checkLengthFacet(errs *ErrorList, declared, base *Facets) {
	if declared.Length == nil {
		return
	}
	if base.Length != nil && declared.Length.Value != base.Length.Value {
		errs.Addf(SpecLengthValidRestriction, declared.Length.Pos, "length %d differs from base length %d", declared.Length.Value, base.Length.Value)
	}
	if base.MinLength != nil && declared.Length.Value < base.MinLength.Value {
		errs.Addf(SpecLengthValidRestriction, declared.Length.Pos, "length %d < base minLength %d", declared.Length.Value, base.MinLength.Value)
	}
	if base.MaxLength != nil && declared.Length.Value > base.MaxLength.Value {
		errs.Addf(SpecLengthValidRestriction, declared.Length.Pos, "length %d > base maxLength %d", declared.Length.Value, base.MaxLength.Value)
	}
	if base.Length != nil {
		checkFixedInt(errs, declared.Length, base.Length, "length")
	}
}

// checkMinLengthFacet enforces the declared minLength against the base
// length/minLength/maxLength, plus the fixed-facet rule.
func checkMinLengthFacet(errs *ErrorList, declared, base *Facets) {
	if declared.MinLength == nil {
		return
	}
	if base.MinLength != nil && declared.MinLength.Value < base.MinLength.Value {
		errs.Addf(SpecMinLengthValidRestriction, declared.MinLength.Pos, "minLength %d < base minLength %d", declared.MinLength.Value, base.MinLength.Value)
	}
	if base.MaxLength != nil && declared.MinLength.Value > base.MaxLength.Value {
		errs.Addf(SpecMinLengthValidRestriction, declared.MinLength.Pos, "minLength %d > base maxLength %d", declared.MinLength.Value, base.MaxLength.Value)
	}
	if base.Length != nil && declared.MinLength.Value > base.Length.Value {
		errs.Addf(SpecMinLengthValidRestriction, declared.MinLength.Pos, "minLength %d > base length %d", declared.MinLength.Value, base.Length.Value)
	}
	checkFixedInt(errs, declared.MinLength, base.MinLength, "minLength")
}

// checkMaxLengthFacet enforces the declared maxLength against the base
// length/minLength/maxLength, plus the fixed-facet rule.
func checkMaxLengthFacet(errs *ErrorList, declared, base *Facets) {
	if declared.MaxLength == nil {
		return
	}
	if base.MaxLength != nil && declared.MaxLength.Value > base.MaxLength.Value {
		errs.Addf(SpecMaxLengthValidRestriction, declared.MaxLength.Pos, "maxLength %d > base maxLength %d", declared.MaxLength.Value, base.MaxLength.Value)
	}
	if base.MinLength != nil && declared.MaxLength.Value < base.MinLength.Value {
		errs.Addf(SpecMaxLengthValidRestriction, declared.MaxLength.Pos, "maxLength %d < base minLength %d", declared.MaxLength.Value, base.MinLength.Value)
	}
	if base.Length != nil && declared.MaxLength.Value < base.Length.Value {
		errs.Addf(SpecMaxLengthValidRestriction, declared.MaxLength.Pos, "maxLength %d < base length %d", declared.MaxLength.Value, base.Length.Value)
	}
	checkFixedInt(errs, declared.MaxLength, base.MaxLength, "maxLength")
}

// checkWhiteSpaceRestriction enforces whiteSpace narrowing + fixed.
// spec: whiteSpace-valid-restriction — XSD 1.1 Part 2 §4.3.6.5 (xmlschema11-2.md#whiteSpace-valid-restriction)
func checkWhiteSpaceRestriction(errs *ErrorList, declared, base *Facets) {
	if declared.WhiteSpace == WSUnset || base.WhiteSpace == WSUnset {
		return
	}
	if wsRank(declared.WhiteSpace) < wsRank(base.WhiteSpace) {
		errs.Addf(SpecWhiteSpaceValidRestriction, declared.WhiteSpacePos, "whiteSpace %s loosens base whiteSpace %s", declared.WhiteSpace, base.WhiteSpace)
	}
	if base.WhiteSpaceFixed && declared.WhiteSpace != base.WhiteSpace {
		errs.Addf(SpecFixedFacetValue, declared.WhiteSpacePos, "facet whiteSpace is fixed to %s in the base type", base.WhiteSpace)
	}
}

// checkBoundsRestriction enforces that each declared range bound lies within
// the base's effective bounds.
// spec: {min,max}{Inclusive,Exclusive}-valid-restriction — XSD 1.1 Part 2 §4.3.{7,8,9,10}.5
func checkBoundsRestriction(errs *ErrorList, declared, base *Facets, cmp CompareFunc) {
	checkBoundRestriction(errs, declared.MinInclusive, "minInclusive", true, true, SpecMinInclValidRestriction, base.MinInclusive, base, cmp)
	checkBoundRestriction(errs, declared.MaxInclusive, "maxInclusive", false, true, SpecMaxInclValidRestriction, base.MaxInclusive, base, cmp)
	checkBoundRestriction(errs, declared.MinExclusive, "minExclusive", true, false, SpecMinExclValidRestriction, base.MinExclusive, base, cmp)
	checkBoundRestriction(errs, declared.MaxExclusive, "maxExclusive", false, false, SpecMaxExclValidRestriction, base.MaxExclusive, base, cmp)
}

// checkBoundRestriction validates one declared bound against the base bounds.
// Incomparable counts as outside; exclusive-vs-exclusive equality on the same
// side is the one permitted equality against an exclusive base bound.
func checkBoundRestriction(errs *ErrorList, b *Bound, name string, minSide, incl bool, ref SpecRef, baseSame *Bound, base *Facets, cmp CompareFunc) {
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
		if !ok || o == OrderLess || (o == OrderEqual && (!minSide || incl)) {
			errs.Addf(ref, b.Pos, "%s %s <= base minExclusive %s", name, b.Lexical, base.MinExclusive.Lexical)
		}
	}
	if base.MaxExclusive != nil {
		o, ok := cmp(b.Value, base.MaxExclusive.Value)
		if !ok || o == OrderGreater || (o == OrderEqual && (minSide || incl)) {
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

// checkDigitsRestriction enforces totalDigits/fractionDigits narrowing.
// spec: {total,fraction}Digits-valid-restriction — XSD 1.1 Part 2 §4.3.{11,12}.5
func checkDigitsRestriction(errs *ErrorList, declared, base *Facets) {
	// spec: totalDigits-valid-restriction — XSD 1.1 Part 2 §4.3.11.5 (xmlschema11-2.md#totalDigits-valid-restriction)
	if declared.TotalDigits != nil && base.TotalDigits != nil && declared.TotalDigits.Value > base.TotalDigits.Value {
		errs.Addf(SpecTotalDigitsValidRestriction, declared.TotalDigits.Pos, "totalDigits %d > base totalDigits %d", declared.TotalDigits.Value, base.TotalDigits.Value)
	}
	checkFixedInt(errs, declared.TotalDigits, base.TotalDigits, "totalDigits")
	// spec: fractionDigits-valid-restriction — XSD 1.1 Part 2 §4.3.12.5 (xmlschema11-2.md#fractionDigits-valid-restriction)
	if declared.FractionDigits != nil {
		if base.FractionDigits != nil && declared.FractionDigits.Value > base.FractionDigits.Value {
			errs.Addf(SpecFractionDigitsValidRestriction, declared.FractionDigits.Pos, "fractionDigits %d > base fractionDigits %d", declared.FractionDigits.Value, base.FractionDigits.Value)
		}
		if base.TotalDigits != nil && declared.FractionDigits.Value > base.TotalDigits.Value {
			errs.Addf(SpecFractionDigitsValidRestriction, declared.FractionDigits.Pos, "fractionDigits %d > base totalDigits %d", declared.FractionDigits.Value, base.TotalDigits.Value)
		}
		checkFixedInt(errs, declared.FractionDigits, base.FractionDigits, "fractionDigits")
	}
}

// checkTimezoneRestriction enforces explicitTimezone fixed + narrowing: a
// restriction may not widen the base (from required only required is allowed,
// from prohibited only prohibited, from optional anything).
// spec: explicitTimezone-valid-restriction — XSD 1.1 Part 2 §4.3.16.5 (xmlschema11-2.md#explicitTimezone-valid-restriction)
func checkTimezoneRestriction(errs *ErrorList, declared, base *Facets) {
	if declared.ExplicitTimezone == ETZUnset || base.ExplicitTimezone == ETZUnset {
		return
	}
	// spec: fixed-facet-value — XSD 1.1 Part 2 §4.3 (xmlschema11-2.md#fixed-facet-value)
	if base.ExplicitTimezoneFixed && declared.ExplicitTimezone != base.ExplicitTimezone {
		errs.Addf(SpecFixedFacetValue, declared.ExplicitTimezonePos, "facet explicitTimezone is fixed in the base type")
	}
	if base.ExplicitTimezone != ETZOptional && declared.ExplicitTimezone != base.ExplicitTimezone {
		errs.Addf(SpecETZValidRestriction, declared.ExplicitTimezonePos,
			"explicitTimezone %q cannot restrict base explicitTimezone %q",
			declared.ExplicitTimezone, base.ExplicitTimezone)
	}
}
