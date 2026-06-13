# goxsd5 — Spec Conformance Coverage Matrix

This file is the authoritative map from XSD 1.1 named constraints to their
implementation. It is kept green by a test that walks the `SpecRef` constant
table and fails if any constraint is **declared but unreferenced** or
**referenced but undeclared**.

**Specs.** Part 1 (Structures) → `docs/clean/xmlschema11-1.md`; Part 2
(Datatypes) → `docs/clean/xmlschema11-2.md`.

**Status legend.**
- `done` — implemented and enforced, with a `// spec:` annotation at the site.
- `wip` — partially implemented.
- `deferred` — recognized but not yet enforced (e.g. instance-validation only).
- `N/A` — out of scope for the parser (instance validation / codegen concern).
- *(blank)* — not yet started.

Add a row when a constraint becomes relevant to a milestone; fill `Impl
(file:line)` when enforced. Keep the `SpecRef.ID` column identical to the
constant name in code.

---

## Part 1 — Structures

### Schema Representation Constraints (`src-*`)

| Constraint ID | Section | Milestone | Status | Impl (file:line) |
|---------------|---------|-----------|--------|------------------|
| src-qname | §3.15.3 | M1 | done | parser/attrcheck.go:47 |
| src-element | §3.3.3 | M5 | wip (4.3 → M6) | parser/elemtable.go:698 |
| src-attribute | §3.2.3 | M5 | wip (6.3 → M6) | parser/elemtable.go:726 |
| src-ct | §3.4.3 | M5/M6 | done (mixed-emptiable clause of 2.2 deferred) | parser/buildcomplex.go:134 |
| src-simple-type | §3.16.3 | M5 | done | |
| src-restriction-base-or-simpleType | §3.16.3 | M5 | done | parser/elemtable.go:207 |
| src-list-itemType-or-simpleType | §3.16.3 | M5 | done | parser/elemtable.go:224 |
| src-union-memberTypes-or-simpleTypes | §3.16.3 | M5 | done | parser/elemtable.go:240 |
| src-attribute_group | §3.6.3 | M5/M6 | done (3 = circularity in M6) | parser/buildterms.go:378 |
| src-model_group_defn | §3.7.3 | M5 | done | |
| src-identity-constraint | §3.11.3 | M5/M6 | done (5 = ref category match in M6) | parser/elemtable.go:663 |
| src-ta | §3.12.3 | M5 | done | parser/elemtable.go:413 |
| src-schema | §3.17.3 | M5 | done | parser/elemtable.go:113 |
| src-annotation | §3.15.3 | M5 | done | |
| src-wildcard | §3.10.3 | M5 | done | parser/elemtable.go:761 |
| src-id | §3.17.3 | M5 | done | parser/validate.go:139 |
| cip | §4.2.2 | M9 | done (conditional inclusion: vc:minVersion/maxVersion + typeAvailable/Unavailable + facetAvailable/Unavailable evaluation and value validity) | parser/validate.go:170 |
| no-xmlns | §3.2.6.3 | M5 | done | parser/elemtable.go:750 |
| no-xsi | §3.2.6.4 | M5 | wip (ref'd uses → M6) | parser/elemtable.go:754 |
| sch-props-correct.2 (unique globals) | §3.17.6 | M5 | done | |
| p-props-correct.2.1 (min ≤ max) | §3.9.6 | M5 | done | |
| n-props-correct (public/system) | §3.14.6 | M5 | done | |
| src-import | §4.2.6.2 | M5/M7 | done (1.1/1.2 same-ns + no-ns, 3 imported-doc ns match; unresolvable location tolerated) | parser/elemtable.go:136 |
| src-include | §4.2.3 | M5/M7 | done (2.1 ns match, unresolvable location errors) | parser/loader.go:213 |
| src-include.2 (chameleon) | §4.2.3 | M7 | done (absorbed ns + unqualified-reference remapping) | |
| src-redefine | §4.2.5 | M5/M7 | wip (pervasive replacement, existence check, type self-derivation; group/attrGroup occurrence checks deferred) | parser/loader.go:337 |
| src-override | §4.2.4 | M5/M7 | done (pervasive transitive replacement, unmatched children ignored) | |
| src-resolve | §3.15.3 | M6/M7 | done (cross-document via global registry; 4.2 namespace-not-imported check) | parser/builder.go:101 |

### Schema Component Constraints — props-correct (`*-props-correct`)

| Constraint ID | Section | Milestone | Status | Impl (file:line) |
|---------------|---------|-----------|--------|------------------|
| st-props-correct | §3.16.6 | M4/M6 | done (2 = cycles; 3 = final exclusion for list/union) | parser/builder.go:147 |
| ct-props-correct | §3.4.6 | M6 | done (4 = duplicate attr uses; the 1.0 "≤1 ID attr" rule was dropped in 1.1) | parser/buildschema.go:261 |
| ct-props-correct.3 (no circular defs) | §3.4.6 | M6 | done (post-pass; cycle broken to keep the model walkable) | |
| e-props-correct | §3.3.6 | M6 | done (2 = value constraint; 4 = subst final exclusion; the 1.0 "no ID default" rule was dropped in 1.1) | parser/buildterms.go:98 |
| a-props-correct | §3.2.6 | M6 | done (2 = simple type + value constraint; the 1.0 "no ID default" rule was dropped in 1.1) | parser/buildterms.go:259 |
| au-props-correct | §3.5.6 | M6 | done (2 = fixed consistency) | parser/buildterms.go:357 |
| mg-props-correct | §3.8.6 | M6 | done (2 = circular model groups) | parser/buildterms.go:484 |
| c-props-correct | §3.11.6 | M6 | done (keyref category + field arity; ref cycles) | parser/buildterms.go:578 |

### Derivation validity (`cos-*`, `derivation-ok-*`)

| Constraint ID | Section | Milestone | Status | Impl (file:line) |
|---------------|---------|-----------|--------|------------------|
| cos-ct-extends | §3.4.6 | M6 | wip (1.4.2 + attr-use conflicts done; particle/mixed consistency deferred) | parser/buildcomplex.go:188 |
| cos-ct-restricts | §3.4.6 | M6/M9 | deferred (cos-particle-restrict; expectations ratchet tolerates) | |
| derivation-ok-restriction | §3.4.6 | M6/M9 | deferred | |
| cos-equiv-class (substitution groups) | §3.3.6 | M6 | deferred (the enforced part — substitution-group final exclusion — is reported under e-props-correct.4; full equivalence-class membership not yet checked under this ID) | |
| cos-valid-default | §3.3.6 | M6 | done (mixed-emptiable clause deferred) | parser/buildterms.go:110 |
| enumeration-required-notation | Part 2 §3.3.19 | M6 | done | parser/buildterms.go:226 |

---

## Part 2 — Datatypes

### Facet validation rules (`cvc-*-valid`)

| Constraint ID | Section | Milestone | Status | Impl (file:line) |
|---------------|---------|-----------|--------|------------------|
| cvc-whiteSpace | §4.3.6 | M3 | | |
| cvc-pattern-valid | §4.3.4.4 | M3 |  | xsd/facets.go:240 |
| cvc-enumeration-valid | §4.3.5.4 | M3 |  | xsd/facets.go:320 |
| cvc-length-valid | §4.3.1.4 | M3 |  | xsd/facets.go:254 |
| cvc-minLength-valid | §4.3.2.4 | M3 |  | xsd/facets.go:258 |
| cvc-maxLength-valid | §4.3.3.4 | M3 |  | xsd/facets.go:262 |
| cvc-minInclusive-valid | §4.3.10.4 | M3 |  | xsd/facets.go:281 |
| cvc-maxInclusive-valid | §4.3.7.4 | M3 |  | xsd/facets.go:285 |
| cvc-minExclusive-valid | §4.3.9.4 | M3 |  | xsd/facets.go:289 |
| cvc-maxExclusive-valid | §4.3.8.4 | M3 |  | xsd/facets.go:293 |
| cvc-totalDigits-valid | §4.3.11.4 | M3 |  | xsd/facets.go:299 |
| cvc-fractionDigits-valid | §4.3.12.4 | M3 |  | xsd/facets.go:303 |

### Intra-facet consistency (facet "Constraints on … Schema Components")

| Constraint ID | Section | Milestone | Status | Impl (file:line) |
|---------------|---------|-----------|--------|------------------|
| length-minLength-maxLength | §4.3.1.5 | M3 |  | xsd/facets_check.go:12 |
| minLength-less-than-equal-to-maxLength | §4.3.2.5 | M3 |  | xsd/facets_check.go:19 |
| minInclusive-less-than-equal-to-maxInclusive | §4.3.10.5 | M3 |  | xsd/facets_check.go:39 |
| minExclusive-less-than-equal-to-maxExclusive | §4.3.9.5 | M3 |  | xsd/facets_check.go:43 |
| fractionDigits-less-than-equal-to-totalDigits | §4.3.12.5 | M3 | done | xsd/facets_check.go:56 |
| enumeration-valid-restriction | §4.3.5.5 | M3/M6 | done (enums parsed with the base type at construction) | parser/buildsimple.go:338 |
| src-single-facet-value | §4.3 | M6 | done | parser/buildsimple.go:275 |
| fixed-facet-value | §4.3 | M3/M6 | done | xsd/facets_check.go:84 |
| regex-valid (pattern is a valid regex) | App. G | M3/M6 | done | |
| length-valid-restriction | §4.3.1.5 | M3/M6 | done | xsd/facets_check.go:89 |
| minLength-valid-restriction | §4.3.2.5 | M3/M6 | done | xsd/facets_check.go:101 |
| maxLength-valid-restriction | §4.3.3.5 | M3/M6 | done | xsd/facets_check.go:114 |
| whiteSpace-valid-restriction | §4.3.6.5 | M3/M6 | done | xsd/facets_check.go:131 |
| minInclusive-valid-restriction | §4.3.10.5 | M3/M6 | done (incl. base-membership of the lexical) | xsd/facets_check.go:178 |
| maxInclusive-valid-restriction | §4.3.7.5 | M3/M6 | done (incl. base-membership of the lexical) | xsd/facets_check.go:180 |
| minExclusive-valid-restriction | §4.3.9.5 | M3/M6 | done (incl. base-membership of the lexical) | xsd/facets_check.go:182 |
| maxExclusive-valid-restriction | §4.3.8.5 | M3/M6 | done (incl. base-membership of the lexical) | xsd/facets_check.go:184 |
| totalDigits-valid-restriction | §4.3.11.5 | M3/M6 | done | xsd/facets_check.go:187 |
| fractionDigits-valid-restriction | §4.3.12.5 | M3/M6 | done | xsd/facets_check.go:192 |
| explicitTimezone-valid-restriction | §4.3.16.5 | post-M9 | done (required/prohibited base cannot widen) | xsd/facets_check.go:209 |

### Restriction / narrowing (`cos-st-restricts`, `rcase-*`)

| Constraint ID | Section | Milestone | Status | Impl (file:line) |
|---------------|---------|-----------|--------|------------------|
| cos-st-restricts | §4.1.6 | M3/M6 | wip (special bases, list item / union member variety rules done) | parser/buildsimple.go:45 |
| cos-applicable-facets | §4.1.6 | M3/M6 | done (per-primitive applicability table in buildFacets) | parser/buildsimple.go:269 |
| rcase-LengthAndLength | §4.3 | M3 | | |
| rcase-MinLength | §4.3 | M3 | | |
| rcase-MaxLength | §4.3 | M3 | | |
| rcase-MinInclusive | §4.3 | M3 | | |
| rcase-Enumeration | §4.3 | M3 | | |
| rcase-Pattern | §4.3 | M3 | | |

### Builtin datatypes (HFP fundamental facets + per-type facets)

| Constraint ID | Section | Milestone | Status | Impl (file:line) |
|---------------|---------|-----------|--------|------------------|
| fundamental-facets (ordered/bounded/cardinality/numeric) | §F | M4 | | |
| qname-special (no context-free lexical→value) | §3.4.18 | M3 | | |
| notation-special | §3.4.19 | M3 | | |

---

## Constraints reconciled in M8

These named constraints had no row above when the `xsd.Refs` registry was
cross-checked against this file (Milestone 8). The conformance test
(`xsd/conformance_test.go`) keeps the registry and this matrix in sync from
now on: every `SpecRef` must appear here, every `done`/`wip` row must name
an enforcing file that carries the matching `// spec:` annotation, and every
`// spec:` annotation must map to a declared `SpecRef`.

### Enforced (were untracked)

| Constraint ID | Section | Milestone | Status | Impl (file:line) |
|---------------|---------|-----------|--------|------------------|
| cvc-datatype-valid | §3.1.4 | M3 | done (lexical→value mapping in the facet pipeline) | xsd/facets.go |
| cvc-explicitTimezone-valid | §3.2.7 | M3 | done | xsd/facets.go |
| minInclusive-minExclusive | §4.3.9.5 | M3 | done | xsd/facets_check.go |
| maxInclusive-maxExclusive | §4.3.7.5 | M3 | done | xsd/facets_check.go |
| minInclusive-less-than-maxExclusive | §4.3.10.5 | M3 | done | xsd/facets_check.go |
| minExclusive-less-than-maxInclusive | §4.3.9.5 | M3 | done | xsd/facets_check.go |

### Deferred (recognized, not yet enforced)

| Constraint ID | Section | Milestone | Status | Impl (file:line) |
|---------------|---------|-----------|--------|------------------|
| cos-nonambig (UPA) | §3.8.6 | M9+ | deferred (unique particle attribution) | |
| cos-particle-restrict | §3.9.6 | M9+ | deferred (particle restriction recursion) | |
| cos-element-consistent (EDC) | §3.8.6 | M9+ | deferred (element declarations consistent) | |
| cos-ct-derived-ok | §3.4.6 | M9+ | deferred (complex-type derivation OK) | |
| cos-st-derived-ok | §3.16.6 | M9+ | deferred (simple-type derivation OK) | |
| cos-aw-union | §3.10.6 | M9+ | deferred (wildcard union; first-wildcard-wins approximation) | |
| cos-aw-intersect | §3.10.6 | M9+ | deferred (wildcard intersection) | |
| cos-ns-subset | §3.10.6 | M9+ | deferred (wildcard subset) | |
| cos-all-limited | §3.8.6 | M9+ | deferred (all-group occurrence limits) | |
| cvc-assertions-valid | §3.13.4 | N/A | deferred (assertions stored, evaluated only at instance validation) | |
| ag-props-correct | §3.6.6 | M9+ | deferred (attribute group definition properties) | |
| w-props-correct | §3.10.6 | M9+ | deferred (wildcard properties) | |
| mgd-props-correct | §3.7.6 | M9+ | deferred (model group definition properties) | |
| st-restrict-facets | §4.1.6 | M9+ | deferred (companion to cos-st-restricts facet recursion) | |
| src-expredef | §4.2.5 | M9+ | deferred (redefine/override is a legal redefinition) | |

## Built-in declarations present in every schema (M9)

These are not constraints but components the spec mandates be available
without an explicit `<import>`. They were added in M9 so the W3C suite's
schemas that reference them resolve correctly.

| Component | Section | Status | Impl (file:line) |
|-----------|---------|--------|------------------|
| `xs:error` special simple type | §3.16.7.3 | done (no valid instances; resolvable as a type) | builtin/builtin.go (ErrorType) |
| Built-in XSI attribute declarations (`xsi:type`, `xsi:nil`, `xsi:schemaLocation`, `xsi:noNamespaceSchemaLocation`) | §3.2.7 | done (reference needs no import) | builtin/builtin.go (XSIAttributes) |
| xml namespace schema (`xml:lang`/`space`/`base`/`id`, `specialAttrs`) | §F.1 / xml.xsd | done (well-known w3.org URLs served by builtinResolver) | parser/builtinschemas.go |

## Conformance harness (M9)

The W3C XSD test suite is wired through `parser/TestConformanceSuite`. For
every schema test applicable to XSD 1.1 it runs the full pipeline and compares
our verdict to the suite's declared validity, gated against the committed
ratchet `testdata/xsd11-expectations.txt` (`pass` lines auto-ratchet up,
`skip:` lines are curated for deferred features). Regressions and unrecorded
passes both fail; see PLAN.md M9. Run with `-update-expectations` to re-baseline.
