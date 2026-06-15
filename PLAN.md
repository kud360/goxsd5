# goxsd5 — Parser Project Plan

XSD 1.1 schema parser, in multiple packages, built bottom-up. The parser emits a
fully linked, validated model from the **`xsd`** package, which is the public
interface. Codegen is a future consumer of that model and is out of scope here
except for keeping its seam clean.

This plan governs the parser only (`xsd`, `builtin`, `parser` + `parser/xmltree`).
Every milestone is bound by the **Spec Traceability** rules below.

---

## Spec Traceability (standing convention — applies to every milestone)

**Source of truth.** XSD 1.1 Part 1 (Structures) and Part 2 (Datatypes), cleaned
into `docs/clean/xmlschema11-1.md` and `docs/clean/xmlschema11-2.md`. We adhere
strictly to these. Every citation names the **stable constraint ID** *and* the
**section anchor** — constraint IDs (`src-…`, `cos-…`, `cvc-…`, `st-props-correct`,
…) are stable; section numbers drift.

**Annotation form.** Every validation is annotated at its point of enforcement
with the constraint it implements:

```go
// spec: cvc-minLength-valid — XSD 1.1 Part 2 §4.3.1.4 (xmlschema11-2.md#minLength)
if length(v) < f.MinLength {
    return specErr(SpecMinLengthValid, pos, …)
}
```

**Machine-checkable, not just comments.** Two mechanisms make "strictly to spec"
verifiable rather than aspirational:

1. **`SpecRef` on every error.** The error type carries a `Ref SpecRef`
   (`{ID, Part, Section, Anchor}`). A failure tells the user the exact clause
   violated; negative-conformance tests assert the *expected* constraint ID, not
   merely "an error occurred."

2. **`CONFORMANCE.md` coverage matrix.** Enumerates every named constraint from
   both specs → the implementing file:line (or `N/A` / `deferred`). A test walks
   the `SpecRef` constant table and fails if any constraint is
   declared-but-unreferenced or referenced-but-undeclared. This keeps coverage
   honest as the code grows.

Each milestone below tags its validations with their constraint families.

---

## Decisions (settled)

- **Model package is `xsd`** (the public interface). Supersedes the old
  `DESIGN.md`, which called it `schema`.
- **Support both `xs:override` (1.1, primary) and `xs:redefine` (1.0)** via the
  same per-schema scoped-registry mechanism (`src-override` / `src-redefine`).
- **Tree-based front end** (`parser/xmltree`) rather than the streaming handler
  stack from the old `DESIGN.md`.
- **Two passes over a tree**, wrapped by an import-discovery step.

---

## Architecture

Bottom-up dependency order. Nothing depends on `parser` except `cmd` and (future)
`codegen`.

```
xsd/            ← public model + facet/value engine (the public interface)
builtin/        ← XSD 1.1 builtins expressed in xsd terms (facets + parse + compare)
parser/
  xmltree/      ← generic XML → node tree (stdlib), namespaces + line/col
  parser.go     ← orchestration (discover → pass 1 → pass 2)
codegen/        ← future; planned seam only
cmd/goxsd5/
```

**Flow:**

1. **Discover** — follow `import` / `include` / `override` / `redefine` to load
   the full document set. Cyclic *imports* are allowed; load each URI once.
2. **Pass 1 (per document)** — `xmltree` → internal component structs, one per
   schema component, **references left dangling**. Register globals; record
   `override` / `redefine` clauses in their own per-schema scope.
3. **Pass 2 (global)** — resolve references, **topologically sort by
   derivation/inheritance**, build `xsd` models in that order so each base is
   fully built (with effective facets) before its derivatives; validate facet
   narrowing and intra-facet consistency as you go; emit the public model.

Cyclic *imports* are fine; cyclic *type definitions* surface as a back-edge in
the topo sort → hard error with both positions. Ordering avoids cycles; explicit
cycle detection is retained in the topo sort as the safety net (free there).

---

## Milestone 0 — Foundations & decisions

**Deliverables**
- `xsd.Pos{URI, Line, Column}`, `xsd.QName{Namespace, Local}`.
- `SpecRef{ID, Part, Section, Anchor}` and the constant table seed.
- A single structured error type carrying `Pos` (plus a secondary `Pos` for
  "conflicting definition" errors) and `Ref SpecRef`; an `errors.Join`-based
  multi-error so one parse reports many problems.
- Constants: `XSDNS`, `XSINS`, `XMLNS`.

**Spec families:** none enforced yet; establishes the `SpecRef` machinery that
all later milestones use.

**Exit:** types compile; error formatting renders `uri:line:col: [ID] message`.

---

## Milestone 1 — `parser/xmltree`

Generic, schema-agnostic XML tree over `encoding/xml`.

**Deliverables**
- `Node{Name xml.Name, Attrs []Attr, Children []*Node, CharData, Pos, EndPos, NS *NSContext}`.
- Namespace capture: in-scope prefix→URI map per node (`NSContext`), so
  attribute *values* (`type="tns:Foo"`) resolve later — `encoding/xml` resolves
  names but not attribute-value QNames. `node.ResolveQName("tns:Foo") (QName, error)`.
- Line/column via `Decoder.InputOffset()` mapped to line:col (lazy offset index);
  record start and end positions.
- Preserve **foreign attributes/elements** (anything not in `XSDNS`) verbatim —
  feeds the "unknown content via extension" capture later.

**Spec families:** `src-qname` (Part 1 §3.15.3) for attribute-value QName
resolution. Structural `src-…` constraints are *recorded* here, *enforced* later.

**Tests:** position round-trip; QName-in-attribute resolution; mixed namespaces;
CDATA in `xs:documentation`.

---

## Milestone 2 — `xsd` model skeleton (public surface)

Pure data, zero dependency on `parser`.

**Deliverables**
- `Schema`, `ElementDecl`, `AttributeDecl`, `ComplexType`, `SimpleType`,
  `AttributeGroup`, `Group`, `Notation`; content model (`Sequence` / `Choice` /
  `All` / `SimpleContent`), particles, `Wildcard`; `DerivationSet`, `Form`.
- Derivation chain via typed fields: `BaseType`, `Derivation =
  *Restriction | *Extension | *List | *Union`.
- `Extensions` field on every component for the foreign content captured in M1.
- `SimpleType` carries the engine hooks: `Primitive *SimpleType`, `ParseValue`,
  `Compare`, `ConvertToBase` (defined in M3).

**Spec families:** none enforced; the struct shapes mirror the component
definitions of Part 1 §3 and Part 2 §3–4.

**Exit:** model compiles standalone.

---

## Milestone 3 — Value space, facet pipeline & QName/NOTATION (in `xsd`)

The conceptual core. Must be correct before builtins.

### 3a. Value space
```go
type Value interface{ isValue() }                 // sealed
type ValueContext interface {                      // supplies what lexical→value can't carry
    ResolveQName(prefix, local string) (QName, bool)
}
type ParseFunc   func(lexical string, ctx ValueContext) (Value, error)
type CompareFunc func(a, b Value) (Order, bool)    // bool = comparable (partial orders)
```
Inherited types may **override** `ParseFunc` / `CompareFunc`; resolution walks to
the primitive ancestor unless overridden. Value-space reps (`Decimal` via
`big.Rat`; dates/times per Appendix E day-pinning, **not** `time.AddDate`) live
here.

### 3b. QName / NOTATION special handling
Cannot go lexical→value without namespace context, and that context differs
between schema and instance:
- **Schema** (e.g. a `QName`-typed enumeration facet): resolve using the
  `NSContext` captured by `xmltree` at that node, in Pass 2; store the resolved
  `QName` **plus** the original lexical form.
- **Instance** validation: caller supplies a `ValueContext`. `ParseValue` for
  QName/NOTATION requires non-nil context (typed error otherwise).

### 3c. Ordered facet pipeline (space-correct)
A struct (not a slice) enforces order by construction; each stage runs in its
proper space:
```
1. whiteSpace        (lexical)   preserve|replace|collapse
2. pattern[]         (lexical)   AND across derivation steps, OR within one step
3. <ParseValue>      (lexical → value)   ← the space boundary
4. length/min/maxLength  (value: chars; octets for hexBinary/base64Binary)
5. bounds (min/max In/Exclusive), totalDigits, fractionDigits  (value)
6. enumeration       (value)     membership via CompareFunc / value equality
7. assert[]          (value, XSD 1.1)  delegated to an AssertionEvaluator
```

### 3d. Validations the engine owns
- **Intra-facet consistency:** `minInclusive ≤ maxInclusive`, exclusive/inclusive
  rules, `minLength ≤ maxLength`, `length` exclusive with min/maxLength,
  `fractionDigits ≤ totalDigits`, each enumeration valid in the type's own value
  space, pattern compiles.
- **Facet narrowing** vs base's *effective* facets: child `maxLength ≤` parent's,
  child bounds inside parent's, child enumeration ⊆ parent value space,
  `whiteSpace` can't loosen, `fixed` facets can't change.
- **Pattern inheritance (explicit):** multiple `<pattern>` in one restriction →
  union (OR); patterns across derivation steps → all-must-match (AND). Store
  per-step pattern *groups*; AND the groups.
- **Enum inheritance & `ConvertToBase`:** `(*SimpleType) ConvertToBase(v Value)
  (Value, error)` — identity for plain restriction, real conversion for
  list/union and differing reps; used to check a derived enumeration value is
  admissible in the base.
- **XSD regex (Appendix F):** translator emits RE2 wrapped `\A(?:…)\z`; handles
  implicit anchoring, class subtraction `[a-[b]]`, `\i \I \c \C`, `\p{IsBlock}`,
  `.`=`[^\n\r]`. Block table sourced from `docs/clean/xmlschema11-2.md`.

**Spec families:** `cvc-pattern-valid`, `cvc-enumeration-valid`,
`cvc-length/minLength/maxLength-valid`, `cvc-minInclusive/maxInclusive/…-valid`,
`cvc-fractionDigits/totalDigits-valid`, `cvc-whiteSpace` (Part 2 §4.3);
intra-facet `length-minLength-maxLength`,
`minInclusive-less-than-equal-to-maxInclusive`, `fractionDigits-totalDigits`,
`enumeration-valid-restriction`; narrowing `cos-st-restricts`, `rcase-*`,
`cos-applicable-facets` (Part 2 §4.3.4.3); QName/NOTATION Part 2 §3.4.18–19.

**Tests:** pipeline ordering; space correctness (hexBinary length in octets);
QName with/without context; narrowing accept/reject matrix; pattern AND/OR;
Appendix-F translation cases.

---

## Milestone 4 — `builtin` package (bootstrap from HFP)

Each XSD 1.1 builtin as a package-level `var *xsd.SimpleType`, defined via
**facets + ParseFunc + CompareFunc**, derived by hand from the datatypes
hierarchy in `docs/clean/xmlschema11-2.md` (HFP fundamental facets `ordered`,
`bounded`, `cardinality`, `numeric` + per-primitive constraining facets). Go's
in-package `var` init order lets derived builtins reference their base by pointer
— no `init()`, no maps, no cycles.

- Primitives: `string`, `decimal`, `boolean`, `float`, `double`, the date/time
  family, `duration`, `hexBinary`, `base64Binary`, `anyURI`, `QName`, `NOTATION`,
  plus 1.1 `dateTimeStamp`, `yearMonthDuration`, `dayTimeDuration`,
  `anyAtomicType`.
- Derived: `integer`→`long`→`int`→… ladder; `normalizedString`→`token`→
  `Name`/`NCName`/`ID`/… ; the unsigned/`*Integer` chains — restrictions with
  fixed facets.
- `AllBuiltins() []*xsd.SimpleType` to seed the registry.

**Bootstrapping check:** a test parses the W3C `datatypes.xsd` (schema-for-schemas)
and asserts our hand-written builtin facet tables match it — verified, not just
asserted.

**Spec families:** each builtin annotated to its Part 2 datatype section
(`§3.3.x`); fundamental facets to Part 2 §F; `st-props-correct` evidenced by the
datatypes-schema cross-check.

**Exit:** a fully working primitive type system, independently testable, no
`parser` involved.

---

## Milestone 5 — `parser` Pass 1 (tree → dangling components)

Per discovered document: walk `xmltree`, emit internal component structs (one per
schema component); leave references as dangling `QName`s (`type=`, `ref=`,
`base=`, `itemType=`, `memberTypes=`, `substitutionGroup=`).

- `refRegistry` of globals (types, elements, attributes, groups, attrGroups,
  notations), seeded from `builtin.AllBuiltins()`.
- Record `override` / `redefine` declarations into a **per-schema scoped
  registry** (not global), so replacements are local to the declaring schema.
- Capture foreign attrs/elements into each component's `Extensions`.
- Collect a flat `[]unresolvedRef` (name, kind, pos, owning-doc, target pointer)
  in declaration order for deterministic Pass 2.

**Spec families:** `src-element`, `src-attribute`, `src-ct`, `src-simple-type`,
`src-redefine`, `src-override`, `src-import`, `src-include`.

**Tests:** golden internal-struct dumps; override/redefine scoping; foreign-content
capture.

---

## Milestone 6 — `parser` Pass 2 (resolve, topo-sort, emit)

1. **Resolve** each `unresolvedRef` via the scoped registry (override/redefine
   shadows global) → wire pointers. Unresolved → error with pos + import chain.
2. **Topologically sort** types by derivation edges (base, union members, list
   item). Build `xsd.SimpleType` / `ComplexType` in that order; a back-edge =
   **cyclic type definition error** (explicit detection as the safety net).
3. As each type is built: compute effective facets, run M3 narrowing +
   intra-facet validation; resolve QName/NOTATION facet values using the schema
   node's captured `NSContext`.
4. Resolution-time consistency checks: `default` only with `use="optional"`,
   `default`/`fixed` exclusivity, `nillable` constraints, substitution-group
   `final`/`block`, override/redefine is a legal derivation of what it replaces.

**Spec families:** `st-props-correct`, `ct-props-correct`, `e-props-correct`,
`a-props-correct`, `au-props-correct`; `cos-ct-extends`, `cos-ct-restricts`,
`derivation-ok-restriction`; `cos-equiv-class` (substitution groups); cyclic-type
via `ct-props-correct.3`.

**Exit:** `Parse` returns a fully linked, validated `[]*xsd.Schema`.

---

## Milestone 7 — Imports, includes, redefine/override, resolver, cyclic imports

- `SchemaResolver` interface: `Resolve(ctx, location, base string) (io.ReadCloser, error)`
  — the URI/Location→bytes seam. Default `FileResolver`.
- Discover loads the transitive closure once; cyclic imports recorded as edges,
  never an error. `import` without `schemaLocation` = namespace dependency
  satisfied by another path or errored in Pass 2.
- `include` merges into same target namespace; `import` brings a different
  namespace; chameleon-include (no target namespace) absorbs the includer's
  namespace.

**Spec families:** `src-import`, `src-include`, `src-include.2` (chameleon),
`src-redefine`, `src-override`, `src-resolve` (missing-namespace failure).

**Tests:** A↔B cyclic import; missing import; chameleon include; resolver
injected via test double.

---

## Milestone 8 — Safe mutation / extension API

Methods on the `xsd` model that let users extend/mutate **without producing an
invalid model**: each mutator re-runs the relevant M3/M6 validation on the
affected subgraph and returns an error instead of corrupting state.

- e.g. `(*SimpleType).AddEnumeration(lexical string) error`,
  `RestrictWith(facets) (*SimpleType, error)`,
  `(*ComplexType).AddElement(...) error`.
- Preserve and allow setting `Extensions` (foreign content), so round-tripping
  doesn't lose unknown content.
- Copy-on-write/builder style so a failed mutation leaves the original untouched.

**Spec families:** re-runs the *same* IDs as M3/M6, so a mutation can never bypass
a clause the parser enforces.

**Tests:** narrowing-violating mutation rejected; valid mutation reflected in
`EffectiveFacets()`; foreign content survives a mutate cycle.

---

## Milestone 9 — Conformance suite

- Wire the **W3C XSD 1.1 test suite** through the resolver (positive + negative).
  The suite is large and **not committed**: it is a git submodule at
  `testdata/xsdtests` pinning a specific upstream revision (check out with
  `git submodule update --init testdata/xsdtests`). Drive the 1.1 subset from
  `XSD1_1TestCategories.xml`. Negative cases assert the expected `SpecRef.ID`.
- Spec-example tests for facets, dates (Appendix E), regex (Appendix F).
- Fuzz `xmltree` and the pattern translator.
- `CONFORMANCE.md` coverage matrix kept green by the declared-vs-referenced test.

### Expectations baseline (regression ratchet)

The full suite will not pass on day one, so we never gate CI on "100% green."
Instead a checked-in **expectations file** records the exact set of cases that
*currently* pass; the harness fails only on **regressions** — a case that was
passing and now fails — and on **unexpected passes** that aren't yet recorded.

- **`testdata/xsd11-expectations.txt`** — sorted, one line per known case:
  `<case-id> <expected-outcome>` where outcome is `pass` (we produce the
  spec-correct result) or `skip:<reason>` (deliberately deferred, e.g. assertions
  / XPath, with a short reason). Committed and reviewed; it is the ratchet.
- **Harness rule.** For each suite case: run it, compare to the suite's declared
  validity *and* to the expectations file.
  - was `pass`, still passes → ok.
  - was `pass`, now fails → **FAIL** (regression).
  - not listed, now passes → **FAIL** ("unexpected pass — add it to the
    baseline"), so coverage only ratchets up and new wins can't silently rot.
  - listed `skip:*` → not run as a hard gate; reported as a reminder.
- **Growing it.** As milestones land, run `go test ./parser -update-expectations`
  (a flag that rewrites the file from the current run) to add newly-passing
  cases. The diff to `xsd11-expectations.txt` is the visible, reviewable record
  of conformance progress.
- **On suite-revision bump** (advancing the `testdata/xsdtests` submodule):
  re-baseline in the same commit so the revision change and the expectations
  delta are reviewed together.

This keeps the count monotonic — it only ever increases — and turns each
milestone's spec work into a concrete, diffable jump in passing cases.

---

## Milestone 10 — Codegen seam (future, out of scope now)

Don't build it; keep the boundary clean. `codegen` consumes `[]*xsd.Schema` only;
`TypeMapper` over `*xsd.SimpleType`; generated code has no runtime dependency on
goxsd5 (except the strict value-space types). Nothing in M1–M9 assumes a
serializer.

---

## Sequencing & parallelism

- **Critical path:** M0 → M1 → M2 → M3 → M4 → M5 → M6 (strictly ordered).
- **Parallelizable:** Appendix-F regex translator and the date/time value types
  (within M3) can be built alongside; M7 can start once M5's registry shape is
  fixed; M8 and M9 trail M6.
