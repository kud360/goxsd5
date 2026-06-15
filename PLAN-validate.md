# goxsd5 — Instance Validation Plan

Extend the schema processor into an **in-process schema-validity assessor**: given
a compiled `[]*xsd.Schema` and an instance document, decide validity per the XSD
1.1 Part 1 §3 **Validation Rules** (the `cvc-*` constraints). The assessor is a
new top-level package, **`xsdvalidate`**, and is **format-pluggable**: XML is the
first instance source; JSON and BER are future sources behind the same
abstraction.

This plan governs `xsdvalidate` and its source adapters only. It does not modify the
`xsd` core except where a leaf-safe value-space hook is genuinely missing. Every
milestone is bound by the same **Spec Traceability** rules as `PLAN.md` (constraint
ID + section anchor on every check; `SpecRef` on every error; `CONFORMANCE.md`
coverage matrix extended to the `cvc-*` family).

---

## Why this extends cleanly (current assets)

The library already does schema-*component* validation and built a rich, linked
model. The hard prerequisites for instance *assessment* are largely present:

- **Compiled model** — `xsd.Schema` (global `Elements`/`Types`/`Attributes`),
  `ElementDecl` (type, nillable, default/fixed, substitution groups, identity
  constraints), `ComplexType.Content` (simple / element / empty, mixed, open
  content), particles/groups/wildcards, assertions, type alternatives.
- **Value + facet engine — already the simple-type instance validator.**
  `(*SimpleType).ParseValue(lexical, ctx)` (`xsd/facets.go:301`) runs
  whiteSpace → pattern → lexical→value → length/range/enumeration and emits the
  exact `cvc-*` SpecRefs, taking a `ValueContext` for QName/NOTATION namespace
  resolution. The datatype-specific logic it calls is authored in
  `builtin`/`builtin/xsdtype`, next to the type definitions.
- **UPA already enforced** (`parser/upa.go`) → instance content models are
  deterministic; the content matcher needs no ambiguity handling.
- **`xpath`** already parses the selector/field/assert subset (evaluator is the
  missing half).
- **`xmltree`** gives a position-bearing instance tree.
- **Public conformance suite already vendored** — see *Test strategy*.

---

## Decisions (settled)

- **New top-level package `xsdvalidate`.** The `xsd` core stays a **pure leaf**; the
  instance abstraction lives *above* it, in `xsdvalidate`.
- **External visitor engine, not methods on model types.** Assessment is a
  contextual, stateful, cross-cutting traversal; it reads the immutable model and
  owns its run-state. See *Locality boundary* for the precise rule.
- **Format pluggability via an abstract infoset interface.** The engine walks an
  `Element`/`Attribute`/`Node` interface; each format ships an adapter. The engine
  never imports an XML (or JSON/BER) package.
- **Delegate value-space checks to the type.** Builtins validate themselves next
  to their definitions via the on-the-type hooks; the engine adds only the rules
  that are *not* per-datatype.

---

## Architecture

```
xsd/  builtin/  builtin/xsdtype/  xsdregex/      ← unchanged core (pure leaf)
        ↑
xsdwalk/                (NEW) shared model traversal — used by xsdvalidate AND future codegen
  ├─ visitor.go    push/exhaustive visitor over types/particles/facets (codegen, docs, diff)
  ├─ automaton.go  content-model automaton: deterministic states from particles (pull driver)
  └─ query.go      model queries: subst-group acceptance, wildcard match,
                   governing-type (xsi:type/CTA derivation-OK), attribute-use lookup
        ↑
xsdvalidate/               (NEW) format-agnostic assessor — pull walk, instance-guided
  ├─ infoset.go    abstract instance: Element / Attribute / Node interfaces
  ├─ validator.go  New(schemas, opts) → *Validator; Assess(root) → *Result
  ├─ assess.go     drives xsdwalk with instance-checking actions; the cvc-* engine
  ├─ idc.go        identity-constraint evaluation
  ├─ assert.go     assertion / CTA evaluation (staged, optional)
  └─ result.go     outcome + PSVI-lite (assigned type per node)
xsdvalidate/xmlsrc/  (NEW)    xmltree.Node → xsdvalidate.Element
xsdvalidate/jsonsrc/ (FUTURE) JSON → infoset (object key→local name, array→repeats)
xsdvalidate/bersrc/  (FUTURE) BER/ASN.1 tag→element
```

**Two walks, one algebra.** `xsdwalk` serves a **push/exhaustive** walk (schema
only — codegen, docs, diff visit every component) and a **pull/demand-driven**
walk (instance leads, model follows — the validator and any generated deserializer
runtime touch only reachable parts, in instance order). The reusable core is the
*model algebra* — the content automaton + the queries — not the driver; push and
pull drivers stay thin on top. `xsdvalidate` is "the pull driver with cvc-* actions";
codegen is "the push driver with emit-code actions."

### The infoset contract

Assessment is defined over the abstract infoset (a tree of element/attribute/
character items), not over XML syntax. The contract is essentially the XML
`[children]` property:

```go
type Name struct{ Space, Local string }

type Element interface {
    Name() Name
    Attributes() []Attribute             // non-namespace attributes
    Children() []Node                    // ordered element|character items (mixed-aware)
    Lookup(prefix string) (uri string, ok bool)  // in-scope ns bindings
    Pos() xsd.Pos
}
type Attribute interface {
    Name() Name
    Value() string
}
```

The XML adapter resolves prefixes, entities, CDATA, and `xsi:*` into this model.
Non-XML adapters **invent a documented mapping**: the schema language stays XSD,
and "validate JSON/BER against an XSD" means "validate a format-derived infoset
against the XSD" (cf. BadgerFish / JSONx). That mapping lives entirely in the
adapter and never perturbs the engine.

### Locality boundary (the rule)

- **value-space** (lexical / facet / QName-resolution) → **on the type**, in
  `builtin`. The engine calls `decl.Type.ParseValue(text, nsCtx)`; it never
  type-switches on datatype name. New builtin ⇒ its validation ships with its
  definition, zero engine changes.
- **document-scoped** (`xs:ID`/`IDREF`/`IDREFS` uniqueness, keyref resolution,
  assertions) → **engine**. These are intrinsically not per-datatype.

**No `ValidateInstance` method on `xsd` model types.** Such a method would need
the instance node as a parameter, forcing `xsd` to depend on the infoset
abstraction and undoing the pure-leaf invariant; and assessment is contextual
(element-decl-driven, schema-wide substitution/wildcard, stateful ID tables) so a
method would take a fat context and merely delegate. The only "on the type"
granularity that is leaf-safe and welcome is **value-space** (`ParseValue`, and a
possible future `(*SimpleType).ValidateValue(v Value) error`).

### Public API

```go
v := xsdvalidate.New(schemas, nil)  // compile once; immutable, concurrency-safe
res := v.Assess(root)               // root is an xsdvalidate.Element

res, _ := xmlsrc.Validate(v, xmlReader)   // convenience for the XML source
```

`cmd/goxsd5` gains `-validate doc.xml`.

---

## Milestones (each is conformance-measurable on xsts)

### V0 — Harness + skeleton
Extend `suitescan_test.go` to read `<instanceTest>`/`<instanceDocument href>`/
`<expected validity>` (already under the same `testGroup`s the schema ratchet
scans). Add the `xsdvalidate` skeleton + `xmlsrc` adapter over `xmltree`. Add a
`testdata/instance-expectations.txt` ratchet mirroring the schema baseline. No
real validation yet — this establishes the measurement loop.

### V1 — Simple-type & local validity
Char content via `ParseValue` + facets; attribute validation; `xsi:type`
override (+ derivation-OK against the declared type, using existing model
support); `xsi:nil`; fixed/default; simple-content complex types. Add the
document-scoped `cvc-id` (ID uniqueness / IDREF resolution) in the engine.
Leans almost entirely on the existing value/facet code.
*Families:* `cvc-simple-type`, `cvc-datatype-valid`, `cvc-attribute`, `cvc-au`,
`cvc-elt` (basics), `cvc-type`, `cvc-id`.

### V2 — Content-model matching
The particle automaton in `xsdwalk` (built from `Particle`/`ModelGroup`/
`Wildcard`; deterministic thanks to UPA), driven here as a checker. Occurrence ranges, sequence/choice/all,
wildcards (skip/lax/strict + namespace constraint), substitution-group
acceptance + blocking, open content (1.1), mixed content. The bulk of the engine.
*Families:* `cvc-complex-type`, `cvc-particle`, `cvc-elt`, `cvc-wildcard`,
`cvc-model-group`.

### V3 — Identity constraints
An evaluator for the restricted selector/field XPath subset over the infoset
(child/attribute/descendant/union steps). key/keyref/unique node tables;
uniqueness + keyref resolution. Pure tree evaluation → format-agnostic.
*Families:* `cvc-identity-constraint`, `cvc-selector`, `cvc-field`.

### V4 — Assertions & CTA (staged, optional / feature-flagged)
A larger XPath-2.0-subset evaluator over the infoset for `xs:assert` and
`xs:alternative/@test`. CTA selects the element's governing type before V1–V2
run, so the test-evaluation hook is wired in at V1 but may be limited until V4.
*Families:* `cvc-assertion`, type-alternative resolution.

**Cross-cutting:** reuse `xsd.SpecRef` + `errors.go` for `cvc-*` error identity;
PSVI-lite `Result` (per-node assigned type, default-applied) that also feeds IDC/
assertion contexts and future codegen.

---

## Test strategy

The vendored **W3C XML Schema Test Suite** (`testdata/xsdtests` submodule) carries
**~28,000 `<instanceTest>`** entries, each with `<instanceDocument href>` +
`<expected validity="valid|invalid">`, grouped under the same `testGroup`s the
schema ratchet already scans. This is the canonical instance-assessment
conformance corpus; no new corpus to source. Reuse the existing scan + ratchet
machinery with a parallel `instance-expectations.txt`. Per-milestone, the ratchet
records the newly-passing slice; regressions fail the suite.

---

## Relationship to future codegen (M10) — shared vs separate

Future codegen (typed structs + minimal-allocation marshal/unmarshal, optionally
with inline validation) is a **different execution strategy for the same
front-end**: the runtime assessor is an *interpreter* over the schema model; the
codegen is a *compiler* from it. They are **mostly complementary**, with a shared
front-end and a few shared building-block libraries, and **separate back-ends**.

**Shared (build these as reusable libraries, not entangled in the walker):**
- The **`xsd` model** — common front-end; both consume it.
- The **datatype value-space runtime** (`builtin`/`builtin/xsdtype`: parse /
  format / compare / canonical form). Generated code links it for the hard
  datatypes (dateTime, duration, decimal) instead of re-emitting parsers.
- **`xsdwalk`** — the shared model algebra (content automaton + queries). The
  assessor *checks* the element sequence with the automaton; a generated
  deserializer *drives* parsing into struct fields with the same states; codegen
  itself uses the push visitor to *emit* that deserializer. Standalone package
  from day one so codegen lowers each state to a generated `switch`.
- The **`cvc-*` rule + error identity** (`xsd.SpecRef`), so interpreted and
  generated validation report *identically*.

**Value-space vs structure-space — the symmetry.** The locality rule (value-space
on the builtin, structure-space in the shared walk) holds for *both* execution
strategies. A future per-datatype `GenXMLMarshaller`/`GenXMLUnmarshaller` on a
builtin is the *compile-time* analog of runtime `ParseValue` — same locality, free
to call `ParseValue` or emit something bespoke; that is the builtin's own story,
orthogonal to the walk:

| | interpret (runtime) | compile (codegen) |
|---|---|---|
| **value-space** (per datatype, on builtin) | `ParseValue` | `GenXMLMarshaller` / `GenXMLUnmarshaller` |
| **structure-space** (per schema, shared) | instance validator | struct + dispatch emitter |

Both columns sit on `xsdwalk` for structure and on `builtin` for values.

**Separate (back-end specific — do not try to unify):**
- **Execution / memory model.** Assessor: generic interface tree, allocation is
  fine, polymorphic. Codegen: concrete structs, minimal allocation, monomorphic,
  inlined. Opposite optimization targets — one shared imperative path would be a
  slow validator *and* non-minimal codegen.
- **The infoset `Node` interface + adapters** belong to the runtime assessor.
  Codegen emits per-format marshalers against concrete types and deliberately
  avoids the generic interface (that interface is exactly the allocation it is
  trying to eliminate). Pluggability mechanism differs: runtime = interface
  dispatch; codegen = per-format emitter templates.
- **IDC / assertions.** Hard to inline efficiently; codegen likely calls back
  into the runtime checker for these document-scoped rules rather than emitting
  them inline.

**The principle that keeps them shareable:** keep validation logic represented as
**inspectable data on the model** (`Facets` is data, `Particle` is data), with two
back-ends — `xsdvalidate` interprets it, codegen lowers it. This is the *same*
conclusion as "no `ValidateInstance` method": imperative methods have exactly one
execution strategy and cannot be lowered to generated source, so avoiding them now
is precisely what keeps codegen able to reuse the logic later.

**Concrete constraints on V0–V4 to preserve the seam:** (1) the model algebra
lives in `xsdwalk`, not inside the validator's walker; (2) all value-space
checks via the `xsd`/`builtin` hooks; (3) `cvc-*` identity in `xsd.SpecRef`;
(4) no run-state stashed on the shared model.

### Other future consumers of `xsdwalk`

The walk is worth extracting precisely because it has plausible consumers beyond
the validator and codegen. Push (model-only) and pull (instance-guided) drivers
both reuse the same automaton + queries:

- **Push / exhaustive:** documentation generation; schema diff / wire-compat
  checking (high value given the serialization goals); sample-instance generation
  (fixtures + fuzz seed corpus — pairs with the existing fuzz targets);
  schema→schema translation (XSD→JSON Schema / Protobuf / Avro / ASN.1 — the same
  mapping work the JSON/BER *adapters* need); coverage/complexity metrics.
- **Pull / demand-driven:** streaming (SAX-style) validation, same automaton fed
  by events; editor content-assist / LSP ("what elements/attributes are valid
  here?" is exactly the automaton's frontier query); reflective data-binding into
  generic maps.

The frontier query ("what can come next") is a single high-leverage primitive
shared across sample generation, content-assist, *and* validator error messages
("expected one of …").

**Scoping caution:** build `xsdwalk` to serve V0–V4 cleanly with a neutral,
consumer-agnostic boundary — validated by being able to *articulate* codegen as a
second consumer, not by implementing for eleven up front. Two real consumers is
enough to find the seam; extract/firm it up when codegen actually arrives, seeded
from the validator's usage.
