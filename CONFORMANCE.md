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
| src-qname | §3.15.3 | M1 | done | |
| src-element | §3.3.3 | M5 | wip (4.3 → M6) | |
| src-attribute | §3.2.3 | M5 | wip (6.3 → M6) | |
| src-ct | §3.4.3 | M5 | wip (semantic clauses → M6) | |
| src-simple-type | §3.16.3 | M5 | done | |
| src-restriction-base-or-simpleType | §3.16.3 | M5 | done | |
| src-list-itemType-or-simpleType | §3.16.3 | M5 | done | |
| src-union-memberTypes-or-simpleTypes | §3.16.3 | M5 | done | |
| src-attribute_group | §3.6.3 | M5 | done | |
| src-model_group_defn | §3.7.3 | M5 | done | |
| src-identity-constraint | §3.11.3 | M5 | wip (5 → M6) | |
| src-ta | §3.12.3 | M5 | done | |
| src-schema | §3.17.3 | M5 | done | |
| src-annotation | §3.15.3 | M5 | done | |
| src-wildcard | §3.10.3 | M5 | done | |
| src-id | §3.17.3 | M5 | done | |
| no-xmlns | §3.2.6.3 | M5 | done | |
| no-xsi | §3.2.6.4 | M5 | wip (ref'd uses → M6) | |
| sch-props-correct.2 (unique globals) | §3.17.6 | M5 | done | |
| p-props-correct.2.1 (min ≤ max) | §3.9.6 | M5 | done | |
| n-props-correct (public/system) | §3.14.6 | M5 | done | |
| src-import | §4.2.6.2 | M5/M7 | wip (clause 1 done) | |
| src-include | §4.2.3 | M5/M7 | wip (representation done) | |
| src-include.2 (chameleon) | §4.2.3 | M7 | | |
| src-redefine | §4.2.5 | M5/M7 | wip (representation done) | |
| src-override | §4.2.4 | M5/M7 | wip (representation done) | |
| src-resolve | §3.15.3 | M7 | | |

### Schema Component Constraints — props-correct (`*-props-correct`)

| Constraint ID | Section | Milestone | Status | Impl (file:line) |
|---------------|---------|-----------|--------|------------------|
| st-props-correct | §3.16.6 | M4/M6 | | |
| ct-props-correct | §3.4.6 | M6 | | |
| ct-props-correct.3 (no circular defs) | §3.4.6 | M6 | | |
| e-props-correct | §3.3.6 | M6 | | |
| a-props-correct | §3.2.6 | M6 | | |
| au-props-correct | §3.5.6 | M6 | | |

### Derivation validity (`cos-*`, `derivation-ok-*`)

| Constraint ID | Section | Milestone | Status | Impl (file:line) |
|---------------|---------|-----------|--------|------------------|
| cos-ct-extends | §3.4.6 | M6 | | |
| cos-ct-restricts | §3.4.6 | M6 | | |
| derivation-ok-restriction | §3.4.6 | M6 | | |
| cos-equiv-class (substitution groups) | §3.3.6 | M6 | | |

---

## Part 2 — Datatypes

### Facet validation rules (`cvc-*-valid`)

| Constraint ID | Section | Milestone | Status | Impl (file:line) |
|---------------|---------|-----------|--------|------------------|
| cvc-whiteSpace | §4.3.6 | M3 | | |
| cvc-pattern-valid | §4.3.4.4 | M3 | | |
| cvc-enumeration-valid | §4.3.5.4 | M3 | | |
| cvc-length-valid | §4.3.1.4 | M3 | | |
| cvc-minLength-valid | §4.3.2.4 | M3 | | |
| cvc-maxLength-valid | §4.3.3.4 | M3 | | |
| cvc-minInclusive-valid | §4.3.10.4 | M3 | | |
| cvc-maxInclusive-valid | §4.3.7.4 | M3 | | |
| cvc-minExclusive-valid | §4.3.9.4 | M3 | | |
| cvc-maxExclusive-valid | §4.3.8.4 | M3 | | |
| cvc-totalDigits-valid | §4.3.11.4 | M3 | | |
| cvc-fractionDigits-valid | §4.3.12.4 | M3 | | |

### Intra-facet consistency (facet "Constraints on … Schema Components")

| Constraint ID | Section | Milestone | Status | Impl (file:line) |
|---------------|---------|-----------|--------|------------------|
| length-minLength-maxLength | §4.3.1.5 | M3 | | |
| minLength-less-than-equal-to-maxLength | §4.3.2.5 | M3 | | |
| minInclusive-less-than-equal-to-maxInclusive | §4.3.10.5 | M3 | | |
| minExclusive-less-than-equal-to-maxExclusive | §4.3.9.5 | M3 | | |
| fractionDigits-less-than-equal-to-totalDigits | §4.3.12.5 | M3 | | |
| enumeration-valid-restriction | §4.3.5.5 | M3 | | |

### Restriction / narrowing (`cos-st-restricts`, `rcase-*`)

| Constraint ID | Section | Milestone | Status | Impl (file:line) |
|---------------|---------|-----------|--------|------------------|
| cos-st-restricts | §4.1.6 | M3/M6 | | |
| cos-applicable-facets | §4.1.6 | M3 | | |
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
