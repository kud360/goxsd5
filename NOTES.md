# Working notes (restart context)

Self-notes for resuming work. Keep updated at every checkpoint commit.

## Goal
Implement PLAN.md (XSD 1.1 parser, packages `xsd`, `builtin`, `parser`,
`parser/xmltree`), then run the W3C suite via the M9 ratchet harness and
baseline `testdata/xsd11-expectations.txt`.

## Status — M5 done (2026-06-12). Resume at M6.
- [x] M0 foundations (xsd: Pos, QName, SpecRef registry, Error/ErrorList, RefIDs)
- [x] M1 parser/xmltree (NS-scoped tree, line/col, src-qname, foreign content)
- [x] M2 xsd model skeleton (model.go — full Part 1 §3 component shapes)
- [x] M3 value space + facet pipeline + Appendix-G regex
- [x] M4 builtin package (all 1.1 builtins incl. dateTimeStamp/yearMonth/dayTimeDuration)
- [x] M5 parser pass 1 (structural table, walker, registry — see below)
- [ ] M6 parser pass 2 ← NEXT
- [ ] M7 imports/includes/override/redefine + resolver
- [ ] M9 W3C harness + expectations baseline
- [ ] M8 mutation API, CONFORMANCE.md fill-in, cmd/goxsd5

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

## Plan for M6–M7 (decided, not yet built)
- Pass 2 = build-on-demand recursion with memoization + in-progress marks
  (equivalent to topo sort; back-edge = cyclic type error per
  ct-props-correct.3 / st-props-correct.2). Each built SimpleType:
  facet construction parses enum/bounds lexicals **with the base type**
  (gives enumeration-valid-restriction for free), then ValidateFacetSet +
  CheckFacetRestriction (xsd/facets_check.go), then MergeFacets.
  QName-typed facet values resolve via the node's captured NSContext
  (xmltree.Node.NS / ResolveQName).
- Override/redefine: per-schema scoped registry shadowing the global one.
- Discovery: SchemaResolver interface {Resolve(location, base) (io.ReadCloser, error)},
  default FileResolver; load transitive closure once; cyclic imports fine.
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
