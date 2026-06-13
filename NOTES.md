# Working notes (restart context)

Self-notes for resuming work. Keep updated at every checkpoint commit.

## Goal
Implement PLAN.md (XSD 1.1 parser, packages `xsd`, `builtin`, `parser`,
`parser/xmltree`), then run the W3C suite via the M9 ratchet harness and
baseline `testdata/xsd11-expectations.txt`.

## Status — M9 DONE (2026-06-12). All milestones M0–M9 complete.
NOTE ON ORDERING: user chose to do M8 before M9 (numeric order), overriding
the original NOTES checklist that had M9 first.
- [x] M0 foundations (xsd: Pos, QName, SpecRef registry, Error/ErrorList, RefIDs)
- [x] M1 parser/xmltree (NS-scoped tree, line/col, src-qname, foreign content)
- [x] M2 xsd model skeleton (model.go — full Part 1 §3 component shapes)
- [x] M3 value space + facet pipeline + Appendix-G regex
- [x] M4 builtin package (all 1.1 builtins incl. dateTimeStamp/yearMonth/dayTimeDuration)
- [x] M5 parser pass 1 (structural table, walker, registry — see below)
- [x] M6 parser pass 2 (builder + finishComplexTypes post-pass + test suite)
- [x] M7 composition (import/include/redefine/override, SchemaResolver,
  public Parse) — see "M7 shape" below; GOXSD5_SCAN now runs the FULL
  pipeline per suite group: 5231 valid groups → 99 with errors (triage below)
- [x] M8 mutation API (xsd/mutate.go), CONFORMANCE.md fill-in + guard test,
  cmd/goxsd5 — see "M8 shape" below
- [x] M9 W3C harness + expectations baseline (ratchet); pre-M7 scan-triage
  bugs fixed — see "M9 shape" below

## M9 shape (as built)
- parser/conformance_suite_test.go — TestConformanceSuite, the ratchet. Runs
  by default when testdata/xsdtests is present (skips otherwise), gating
  `go test ./parser`. Case id = "<testSet-relpath>#<group>" (unique, 5241
  groups). Per case: full pipeline, our verdict (errs.Empty()) vs suite
  validity; gated against testdata/xsd11-expectations.txt. Rules: recorded
  pass now failing = regression FAIL; unlisted+now-correct = unexpected-pass
  FAIL ("re-run -update-expectations"); unlisted+still-wrong = tolerated;
  skip: = not gated. -update-expectations rewrites the pass set, PRESERVING
  hand-curated skip: lines. Root-doc load failures (XML 1.1/DTD/non-wf) are
  excluded from gating (not a schema verdict). GOXSD5_CONFORMANCE_GAPS=1 logs
  the unrecorded gaps for triage.
- BASELINE: 5709 determinate 1.1 schema cases. Initial M9 baseline was 5537
  pass / 26 skip; post-M9 conformance work (commits after 7ff3a11) raised it
  to 5572 pass (see "Post-M9 conformance fixes" below).

## Post-M9 conformance fixes (deferred-gap triage, 5537 → 5572 pass)
Worked the false positives + small/medium well-defined checks. Each landed
with unit tests and a re-baseline; NO regressions. Commits:
- facet-restriction equality (ParseFacetValue skips base range bounds);
  regex hyphen handling per Part 2 §G.1 ([a-z-+] valid, [!--]/[--z] invalid);
  base {final} blocks complex/simpleContent/simpleType derivation
  (checkFinalAllows). NB: simpleType/@final="extension" stays legal.
- NOTATION must carry an enumeration when used as element/attribute/list-item/
  union-member type, and its enum values must name declared notations;
  anySimpleType/anyAtomicType barred as list item / union member.
- FULL vc:* conditional inclusion (§4.2.2): typeAvailable/Unavailable +
  facetAvailable/Unavailable evaluation + value validity (new SpecCIP);
  a CI-ignored <schema> is emptied (drops its <include>s). validate.go.
- targetNamespace on local element/attr (src-element.4.3 / src-attribute.6.3);
  walker now carries an ancestor stack (w.path).
- Element Declarations Consistent (cos-element-consistent) in
  finishComplexTypes, incl. substitution-group "implicitly contains".
REMAINING gaps (~117), all the genuinely hard / large features — NONE done:
- UPA / cos-nonambig (all240-243, subsgroup902/903, sg-abstract-upa*): needs
  a particle automaton. ~8 cases.
- Particle restriction "subsumption" / cos-particle-restrict (saxon All
  all2xx, ~21 cases): the hardest XSD algorithm.
- xs:all extension rules cos-ct-extends (all302-313, ~9); nested group-ref in
  all (all008-011).
- Wildcards cos-aw-* intersection/union/subset (Wild ~17).
- Open content (Open ~8, openContent ~4).
- CTA / assertions XPath (saxon CTA 6, ibm typeAlternatives 5): needs an XPath
  engine — effectively out of scope; candidates for skip: lines.
- 2 override false positives (over009 double-override dup; over030 false
  mg-props-correct cycle — override-internals bugs) + iri-001 (custom DTD
  entities &URI; — encoding/xml limitation, skip candidate).
Triage tip: GOXSD5_CONFORMANCE_GAPS=1 go test ./parser -run TestConformanceSuite -v
- TRIAGE FIXES (landed before baselining; were false positives in the M7 scan):
  - Dropped the three 1.0 ID rules XSD 1.1 relaxed: a-props-correct.3,
    e-props-correct.5, ct-props-correct.5 (value-constraint VALIDITY checks
    kept). Removed isIDDerived + the three negative build_test cases.
  - xs:error builtin (builtin.ErrorType, §3.16.7.3) — union, no members,
    final=#all, Parse always errors; registered in newRegistry.
  - Four XSI attribute decls (builtin.XSIAttributes, §3.2.7) — resolvable
    with NO import (lookupRef treats XSINS as reachable like XSDNS); decl
    grew a builtinAttr field, buildAttrUse uses it.
  - xml namespace: builtinResolver (parser/builtinschemas.go) wraps the
    loader's resolver and serves a bundled xml.xsd for the well-known w3.org
    URLs (2001/2007/2009 revs, http+https). Wired in newLoader so ALL paths
    (Parse, harness, tests) get it.
  - parser/builtinschemas_test.go unit-tests all three so they hold without
    the suite (which is gitignored).
- suitescan_test.go (GOXSD5_SCAN, M5/M7 informational error histogram) kept
  alongside the real harness; still useful for triage.

## M8 shape (as built)
- xsd/mutate.go — safe mutation API, all copy-on-write (validate a candidate,
  commit only on success; a rejected mutation leaves the receiver untouched):
  - (*SimpleType).RestrictWith(declared *Facets) (*SimpleType, error) —
    derives a new anonymous restriction subtype, mirroring the parser's
    applyRestriction exactly (CheckFacetRestriction + MergeFacets +
    ValidateFacetSet + fundamental-facet recompute). The single trustworthy
    primitive; the others build on the same checks.
  - (*SimpleType).AddEnumeration(...lexicals) error — in place; values
    accumulate onto THIS step's enumeration (repeated calls grow one set, not
    a chain), each parsed against t.BaseType (enumeration-valid-restriction).
    Refuses to mutate IsBuiltin types (they're shared) — derive with
    RestrictWith first.
  - (*SimpleType).AddPattern(p) error — appends a new single-pattern group
    (AND across groups); compiles via CompileRegex. Builtin-guarded.
  - (*SimpleType).EffectiveFacets() *Facets — read-only view of t.Facets.
  - (*ComplexType).AddElement(elem, min, max) error — element/mixed content
    only; wraps existing content + new element in a sequence. Rejects simple
    content and bad occurrence ranges. (Light: UPA/particle checks are
    deferred parser-wide anyway.)
  - DEVIATION from PLAN: AddEnumeration takes ...lexicals and is in-place
    (PLAN wrote `AddEnumeration(lexical string) error`); RestrictWith returns
    a new type (PLAN signature kept). In-place accumulation is the correct
    XSD semantics — a chain of single-value enum restrictions would make each
    prior value invalid.
  - Tests in xsd/mutate_test.go are package xsd_test so they can use real
    builtins as bases (builtin imports xsd → internal test would cycle).
- CONFORMANCE.md — Impl column filled (file:line of the first // spec:
  annotation) for 60 enforced rows; 21 previously-untracked SpecRefs added in
  a "Constraints reconciled in M8" section (6 enforced, 15 deferred/N-A).
- xsd/conformance_test.go — the matrix guard (PLAN's "machine-checkable, not
  just comments"). Four tests over the source tree (walks .., skips _test.go
  and testdata):
  - every xsd.Refs ID has a matrix row (base-ID match, so a row labelled
    p-props-correct.2.1 covers Ref p-props-correct);
  - done/wip rows name a real SpecRef + any Impl file exists (rot-free: file
    presence, NOT line — line numbers in Impl are documentation, not gated);
  - every SpecRef constant is referenced in code UNLESS its matrix row is
    deferred/N-A (declared-but-unreferenced guard, the PLAN's core check);
  - every // spec: <id> annotation maps to a declared SpecRef.
  Caught two real annotation-text bugs (parser/loader.go said "override"
  not "src-override"; parser/validate.go's src-cip note isn't an enforced
  constraint → de-annotated) and one mislabelled row (cos-equiv-class was
  "wip" but enforcement is reported under e-props-correct.4 → marked
  deferred).
- cmd/goxsd5/main.go — CLI: parser.Parse(path, nil) → per-namespace component
  summary on stdout (suppress with -q), errors as
  uri:line:col: [id] message on stderr, exit 1 on any error.

## M7 scan triage (full-pipeline scan over 1.1-valid-expected groups)
85/5231 groups report errors; NONE are composition bugs:
- regex-valid:1 — was 16; the `[-]` lone-hyphen class bug is FIXED
  (post-M7 polish commit). The 1 leftover is a different pattern → triage
  in M9.
- src-resolve:42 — mostly xml:lang & co after the import of
  http://www.w3.org/2001/xml.xsd fails (FileResolver can't fetch URLs;
  unresolvable import is legal, the later reference then errors). M9
  decision: ship built-in xml-namespace attribute decls or map the URL in
  the harness resolver; suite has a local common/xml.xsd? (check).
- a-props-correct:15 + e-props-correct:8 + ct-props-correct:16 — ALL the
  1.0 ID rules (ID-derived w/ value constraint; >1 ID attribute). XSD 1.1
  RELAXED these (suite marks them valid) — verify against docs/clean and
  drop/gate the three checks (a-props-correct.3, e-props-correct.5,
  ct-props-correct.5 as implemented in M6).
- src-restriction-base-or-simpleType:8 — the known precisionDecimal
  optional-feature family → M9 skip list.
- leftovers (sch-props-correct:4, mg-props-correct:1, src-include:1,
  min/max facet:5, parseFail:9 XML-1.1/DTD docs) → M9 skip list / triage.

## M7 shape (as built)
- resolver.go: SchemaResolver{Resolve(location, base) (io.ReadCloser,
  error)}; FileResolver; resolveLocation joins URL-aware (RFC 3986) with
  filepath fallback and doubles as the loader's dedup key.
- loader.go: trees cached per URI (treeEntry distinguishes resolveErr —
  tolerated for import, fatal for include/redefine/override — from
  parseErr, always fatal); schemaDoc INSTANCES cached per (uri,
  effectiveTNS) — chameleon include (src-include.2.2) re-instances the
  same tree under the includer's TNS with chameleonNS set; cyclic
  import/include terminate on the instance cache. Pass-1 validation runs
  once per URI (second instance discards duplicate diagnostics).
  src-import.1.1/.1.2/.3 and src-include.2.1 enforced in compose().
- Redefine/override: replacement children register as THE globals; the
  originals are suppressed from global registration (suppression
  propagates transitively through the target's own compositions via
  doc.targets) and parked in doc.originals. Redefine children get a
  per-child pseudo-doc whose scoped registry maps ONLY their own name to
  the original (self-derivation per src-redefine.5, checked for types);
  override children with no matching original are IGNORED per the
  override transformation (not added). Unmatched redefine = src-redefine
  error. Group/attrGroup redefine occurrence checks (exactly-one self-ref
  min=max=1 / superset) deferred.
- Chameleon reference remapping: qnameAttr is doc-aware; chameleonQName
  maps a reference resolved to "" → absorbed TNS. Token-list sites
  (memberTypes, substitutionGroup, notQName) mapped too;
  doc.defaultAttributes mapped at instance creation.
- lookupRef(builder): every registry resolution now checks namespace
  reachability first (target NS, XSD NS always, doc.importedNS — recorded
  from import compositions regardless of resolution success) →
  src-resolve "not imported here", else "not declared". xml namespace is
  NOT special-cased (strict; see scan triage).
- buildschema.go: buildSchemas groups l.order docs by TNS (root's first),
  one xsd.Schema per namespace; newSchemaShell takes doc-level props from
  the first doc (form defaults etc. keep applying PER DOC during build —
  they ride on decl.doc); addComponent guards every kind (now incl.
  notation) with the registered-decl check, so suppressed originals and
  dups don't build into the maps; replacement children added via
  addReplacementComponents. checkTypeCycles per schema, finishComplexTypes
  once. Single-doc buildSchema kept for the M5/M6 tests.
- parser.go: Parse(location, *Options{Resolver}) → ([]*xsd.Schema, error);
  error is errs.Err() aggregate; unloadable ROOT returns a plain error.
  finish(l, errs) is the register+build tail, shared with the suite scan
  (multi-root: loadRoot per schemaDocument of a group, then finish).
- Known limitations (accepted): builder memoizes per NODE, so one
  document chameleon-included into TWO different namespaces in one schema
  set would share components (rare; revisit if the suite hits it);
  within a redefine child, ALL references to its own name resolve to the
  original (spec limits this to the self-derivation reference); ICs of
  suppressed (overridden) elements are skipped wholesale, ICs inside
  override children register normally.
- Tests: parser_test.go (mapResolver double keyed by resolveLocation;
  import incl. cyclic + not-imported + mismatches; include incl.
  chameleon remapping, cyclic, relative paths; redefine incl. pervasive
  replacement, base-doc-sees-redefinition, group self-ref to original,
  existence + self-derivation negatives; override incl. pervasive +
  transitive-through-include + unmatched-ignored); testdata/m7 fixtures
  for the default FileResolver path.

## M6 shape (as built)
Files (all under parser/, tests in build_test.go + TestBuildSmoke):
- builder.go: builder w/ per-node memo; ST cycles caught eagerly via
  building-marks (st-props-correct.2); CTs memoize their SHELL before
  content is built (content refs back into an unfinished type are legal!)
  and derivation-chain cycles are found by checkTypeCycles in
  buildschema.go (post-pass, breaks the cycle → model stays walkable).
  derivationMethods has a visited guard for not-yet-broken cycles.
- buildsimple.go: restriction (applyRestriction shared w/ simpleContent),
  list, union (member-union flattening), facet construction parsing
  enum/bounds with the BASE type (nsContext adapter), applicability table
  (cos-applicable-facets), src-single-facet-value, fundamental facets.
  xsd.SimpleType.EffectiveCompare() was exported for this.
- buildcomplex.go: simpleContent rest/ext (incl. inline-simpleType
  effective base), complexContent + abbreviated form, extension particle =
  sequence(base, own), attr-use merging (extension=union w/
  ct-props-correct.4, restriction=override+prohibited), ct-props-correct.5
  (≤1 ID attr), openContent + defaultOpenContent application,
  defaultAttributes group.
- buildterms.go: particles (maxOccurs=0 → nil), local/global elements
  (type fallback to subst-head type, value-constraint validation
  cos-valid-default / e-props-correct.5 ID rule, subst-group final
  exclusion via derivationMethods — unreachable chains deliberately NOT
  errored, deferred to derivation-ok), attributes (a-props-correct.2/.3,
  au-props-correct.2 fixed consistency), groups/attrGroups w/ cycle
  detection, wildcards (##other → {tns,absent}; notQName ##defined kept as
  sentinel QName), ICs (ref= form shares component, category match
  src-identity-constraint.5, keyref refer + field-count c-props-correct).
- annot.go: Annotation + Extensions (foreign attrs/nodes) capture.
- buildschema.go: buildSchema(reg, doc, errs) → linked *xsd.Schema; skips
  dup-named nodes (only the registered decl builds); redefine/override
  children NOT assembled (M7). checkTypeCycles + finishComplexTypes
  post-passes (see design note below).

DESIGN NOTE — finishComplexTypes (found by the M6 test suite, NOT by the
smoke test): a base CT can still be MID-BUILD when a derived type is
constructed, because the base's own content may legally reach back into the
derived type (kitchen sink: base contains <element ref="doc"/>, doc's type
is derived, derived extends base). At that moment base.Content/.AttributeUses
are nil/empty, so anything reading base properties at build time silently
loses data. Therefore ALL base-dependent merging is deferred to
finishComplexTypes, a topological post-pass over b.types (runs after
checkTypeCycles, so derivation chains are acyclic): (a) extension effective
particle = sequence(base particle, own) — the base Particle component is
SHARED, not copied; (b) attribute-use merging — each CT's declared material
is parked in builder.pendingAttrs{own, wc, prohibited, override, wcFallback}
during construction and merged with the completed base in the post-pass,
followed by applyDefaultAttributes and checkAttrUses (ct-props-correct.4/5).
Wildcard fallback to the base applies to extensions and simpleContent
(wcFallback), NOT to complexContent restrictions. The simpleContent
early-return path (extension of a simple-content base adding nothing) keeps
its build-time shortcut; finishExtensionParticle has a defensive
SimpleContent branch for the in-progress-base case.

Deliberate deferrals (do NOT implement now): UPA (cos-nonambig), EDC,
cos-particle-restrict, cos-ct-extends/restricts particle checks, wildcard
union/intersection (cos-aw-*; first-wildcard-wins approximation in
buildAttrUses/extendAttrUses), mixed-emptiable check for value
constraints + simpleContent rest of mixed CT (emptiable part), derived
vs base mixed consistency (cos-ct-extends), NOTATION enum values resolving
to declared notations.
Gotcha found by smoke test: kitchen sink had GENUINELY invalid XSD
(final="list union" on a type used as list item/union member;
finalDefault=restriction outlawing the schema's own restrictions; unique
ref'ing a key). The builder was right; the fixtures were wrong.

## M5 shape (as built — parser package)
- `elemtable.go`: per-context table (variants `element@global` vs
  `element@local`, `restriction@simple|simpleContent|complexContent`, …) of
  allowed attrs (+ value checks) and a content model; `content.go` is a tiny
  set-of-positions NFA matcher over child local names with
  furthest-failure error reporting; `attrcheck.go` value checkers
  (parse helpers `parseDerivationSet`/`parseMaxOccurs`/`parseNonNegInt`
  are reused by pass 2). `refError` lets a checker override the reported
  SpecRef (QName checks report src-qname).
- `validate.go`: walker; also enforces src-id, p-props min≤max, and does
  vc:minVersion/maxVersion pruning (pruned nodes recorded in
  `schemaDoc.pruned` — pass 2 must skip them). Other vc:* conditions
  (typeAvailable…) are NOT evaluated (elements retained).
- `schemadoc.go`: `loadDoc` → `schemaDoc` (doc-level attrs, compositions
  list for M7, defaultOpenContent node). `registry.go`: symbol spaces
  (types/elements/attrs/groups/attrGroups/notations/ICs), global registry
  seeded with builtins + AnyType, dup = sch-props-correct.2; redefine/
  override children land in `schemaDoc.scoped` (chained registry), M7 wires
  cross-doc semantics. Named identity constraints are registered globally.
- Suite smoke scan (`suitescan_test.go`, opt-in `GOXSD5_SCAN=1`): 5231
  1.1-valid-expected groups → only 8 pass-1 false positives, ALL from the
  optional precisionDecimal feature (xs:minScale/maxScale, saxon PDecimal +
  ibm D3_3_4) → M9 skip list, do NOT add to the facet table. Remaining
  suite .xsd parse failures are out of scope: XML 1.1 docs (encoding/xml
  can't), DTD entity refs, intentionally non-wf files → M9 skip list.
- xmltree fix: strip U+FEFF after transcode (UTF-16 BOM survived as
  CharData and broke ~30 suite docs).

## Still-deferred (carry into M9 expectations)
- UPA (cos-nonambig), EDC, complex-type restriction particle checks
  (cos-particle-restrict): defer; expectations ratchet tolerates.
- Assertions/CTA (xs:assert/xs:alternative XPath): stored in model, never
  evaluated → suite cases needing XPath evaluation listed as skip.

## M9 harness facts (verified against the checked-out suite)
- Suite already fetched at `testdata/xsdtests/` (gitignored; pinned rev in
  `testdata/fetch-xsdtests.sh`).
- `suite.xml` → `<ts:testSetRef xlink:href>` → testSet files (namespace
  `http://www.w3.org/XML/2004/xml-schema-test-suite/`) → `testGroup` →
  `schemaTest` → `schemaDocument xlink:href` (possibly several) +
  `expected validity="valid|invalid|indeterminate"`.
- `expected` may carry `version` ("1.0", "1.1", "1.0 1.1"…); no version
  attr = all versions. testSet/testGroup/schemaTest can carry `version`
  too (e.g. saxonMeta/* are version="1.1"). Pick the 1.1-applicable
  expectation; skip cases not applicable to 1.1; treat indeterminate as skip.
- Schema tests only (we don't validate instances).
- Harness rule (PLAN M9): expectations file is the ratchet — regression
  (was pass, now fail) = FAIL; unexpected pass (not listed, now passes) =
  FAIL telling you to re-run with -update-expectations.

## Design deviations from PLAN.md (deliberate)
- SimpleType uses Variety + facet fields rather than a typed
  `Derivation = *Restriction|*Extension|*List|*Union` union; ComplexType has
  DerivationMethod. Equivalent information, less indirection.
- M4 "bootstrapping check" test (parse W3C datatypes.xsd and diff builtin
  facet tables) deferred until the parser exists (needs M5/M6).

## Implementation gotchas already handled (don't re-litigate)
- encoding/xml leaves unbound prefixes as raw prefix in Name.Space —
  xmltree.checkBound detects via ':'/'/'-heuristic + in-scope check.
- strings.Builder must not be copied → stack of *strings.Builder in xmltree.
- Go RE2 caps counted repeats at 1000 → applyQuant decomposes (exact for
  fixed/lower bounds; approximate for very wide {n,m} ranges).
- \p{IsUnknownBlock} = match-anything per spec note (Postel default), both
  for \p and \P. Unknown *categories* are hard errors.
- Go unicode.Categories lacks Cn and the C group omits it: categorySet("C")
  unions computed Cn (complement of all assigned).
- xs:base64Binary validated by the spec's regex (single embedded spaces
  legal) before StdEncoding decode of the de-spaced string.
- DateTime: year 0 allowed (1.1), 24:00:00 rolls over, TZ range ±14:00,
  timeline compare with ±840-minute shifts for the no-TZ partial order.
- Duration compare: 4 reference dateTimes + AddDuration with day pinning.
- Decimal = unscaled big.Int + scale (normalized); TotalDigits/FractionDigits
  read straight off it.
- Facet merge: declaring either min-bound clears both inherited min-bounds
  (narrowing checks make that sound); pattern groups accumulate (AND across
  steps, OR within); enumeration replaces (base membership enforced at
  construction time by parsing enums with the base type).
- whiteSpace facet: WSUnset (union types) = no normalization. Patterns are
  matched against the *normalized* lexical.

## Conventions
- Every enforced constraint: `// spec: <id> — XSD 1.1 Part N §x (anchor)` at
  the enforcement site + SpecRef constant in xsd/specref.go + row in
  CONFORMANCE.md (fill Impl column when wiring the conformance test in M8/M9).
- Errors: `uri:line:col: [id] message`; negative tests assert xsd.RefIDs().
- Commit at each milestone boundary; update this file in the same commit.
