# Working notes (restart context)

Self-notes for resuming work. Keep updated at every checkpoint commit.

## Goal
Implement PLAN.md (XSD 1.1 parser, packages `xsd`, `builtin`, `parser`,
`parser/xmltree`), then run the W3C suite via the M9 ratchet harness and
baseline `testdata/xsd11-expectations.txt`.

## >>> NEXT SESSION — cos-particle-restrict phase 3 (the hard remainder) <<<
Phase-3 progress so far (5635 → 5640): the RESTRICTION open-content subset
(item 4 below, restriction half) and the all→choice subsumption (item 3) are
DONE. The remaining gaps are the genuinely hard / risky ones. CORE SAFETY RULE
that has held every session: when a construct isn't fully analyzable, GIVE UP
(return, report NOTHING) — the checks are necessary conditions for L(R) ⊆ L(B),
so a violation is always a real error, and giving up never costs a false
positive. The ratchet (GOXSD5_CONFORMANCE_GAPS=1 go test ./parser -run
TestConformanceSuite -v) lists every remaining gap and flags any regression.
Keep the unit-test + re-baseline + checkpoint-commit rhythm.

DONE this phase (commits after 7686f2f):
- [x] item 4 (RESTRICTION half): checkOpenContentRestrict in restrict.go,
  wired in finishComplexTypes. derivation-ok-restriction §3.4.6.4 clause 9:
  kept open content must be same-or-narrower mode (interleave > suffix),
  wildcard namespace subset, identical-or-stronger processContents. Helpers
  effectiveOpenContent / openContentOpenness / processContentsAtLeastAsStrict /
  particleMatchesNonEmpty / particleHasWildcard. TWO false-positive gates
  learned the hard way: (a) skip the implicit restriction of xs:anyType
  (bct.BaseType == nil) — a plain complexType with open content tripped
  open001/005/... as "base has no open content"; (b) the mode-openness check
  only fires when the restriction content model can produce an element (empty
  content ⇒ interleave≡suffix, open020); (c) suppress the "base has no open
  content" error when the base's own content model holds a wildcard that can
  absorb the open elements (open022). +4 (open016.bad/017/018/019).
- [x] item 3 all→choice: checkParticleRestrict refactored — base slots built
  once, per-name bag check extracted to checkRestrictRun with RUN-LOCAL
  accumulation, dispatched over one flat run OR each branch from
  choiceBranches. A choice (occurring once) of flat element runs is checked
  branch-by-branch (each branch is one possible instance ⇒ each must be a valid
  restriction). +1 (all233).

REMAINING (roughly cheapest/safest first):
1. processContents ordering at the restrict.go wildcard-mapping site (cheap,
   isolated, but fixes NO case alone — prerequisite for item 2). NSSubset in
   particle restriction also requires the restricting wildcard's processContents
   be identical-or-STRICTER than the base's. Helper already exists:
   processContentsAtLeastAsStrict(sub, super) in restrict.go. Add it where a
   restriction wildcard maps to slots[baseWC].wc (NSSubset site). NOT inside
   namespaceConstraintSubset (pure namespace relation reused by attributes).
2. MULTI-base-wildcard cardinality (all238, all244, wild048). ≥2 base wildcards
   with overlapping namespace constraints, so checkParticleRestrict bails
   ("more than one base wildcard: give up", baseWC tracks at most one). Need a
   real solver: each restriction particle may be coverable by SEVERAL base
   wildcards; the question is whether the summed restriction cardinalities pack
   into the base wildcards' ranges so every restriction-allowed multiset is
   base-allowed. all244's comment ("invalid because r allows (one,one,one,
   three,three) whereas b does not") is the canonical case — a small flow/
   packing problem. all238 ALSO needs item 1 (its invalidity is the
   processContents mismatch: R2 lax overlaps W1 strict on namespace "two").
   HARD; may be left tolerated (spec §3.9.6 permits provisional acceptance of
   <all>-group derivations a processor can't decide).
3. EXTENSION open content (open030/033/046, complex018 stays tolerated as a
   spec bug — bug 16786). DELIBERATELY LEFT this session: the extension cases
   need the open-content COMBINATION/inheritance semantics our builder does NOT
   model (open046: an empty extension with appliesToEmpty=false should INHERIT
   the base's open content; our buildOpenContent leaves ec.OpenContent=nil).
   A naive "both present ⇒ B.wc ⊆ R.wc and openness(R) ≥ openness(B)" check
   would catch open030/033 (+2) but risks false positives on valid extension
   cases and is wrong for open046. Do this ONLY after teaching the builder the
   real {open content} mapping for extensions. open048 is a SEPARATE pass-1
   bug: maxOccurs on an <openContent>/<defaultOpenContent>'s <any> must be
   rejected (saxon bug 15618) — that's an elemtable fix, not restriction.
4. EDC type-table / wildcard cases wild069/078/079/081 + wild041 (xsi: in
   notQName). Element Declarations Consistent with type tables + a notQName
   xsi:* edge. Separate from particle restriction; lower priority — read each
   schema first.
SKIP/leave tolerated: all308 (xs:all extension of mixed empty content, spec bug
6202); complex018 (spec bug 16786). Triage each gap by reading its schema in
testdata/xsdtests/saxonData (or ibmData) BEFORE coding; the comment at the top
of each .xsd states why it is invalid.

## Status — M9 DONE (2026-06-12). All milestones M0–M9 complete. Post-M9 at 5640 pass.
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
  to 5635 pass (see "Post-M9 conformance fixes" below).

## Post-M9 conformance fixes (deferred-gap triage, 5537 → 5640 pass)
- SESSION 2026-06-13 phase 3 (5635 → 5640, +5): RESTRICTION open-content subset
  (+4 open016.bad/017/018/019) and all→choice particle restriction (+1 all233).
  See the phase-3 NEXT SESSION block at the top for the design, the
  false-positive gates (anyType base, empty-content mode, base-wildcard
  absorption), and what remains (multi-base-wildcard packing, extension open
  content + its inheritance prerequisite, EDC).
Worked the false positives + small/medium well-defined checks. Each landed
with unit tests and a re-baseline; NO regressions. Commits:
- SESSION 2026-06-13 part 5 (5628 → 5635, +7): ELEMENT WILDCARD particle
  subsumption. checkParticleRestrict (restrict.go) now handles wildcards in the
  flat all/sequence bag check, but only when the BASE group has at most ONE
  wildcard (≥2 → give up: a restriction particle could map to either). Base
  slots are element-or-wildcard; a restriction element maps to a base element
  by name/substitution, else to the base wildcard via NSCompat
  (wildcardAllowsName); a restriction wildcard maps to the base wildcard via
  NSSubset (namespaceConstraintSubset); all particles mapping to a slot sum
  their occurrence ranges, which must stay within the slot's. processContents
  ordering is NOT checked (lenient → no false positive; loses all238). +7
  (all229/235/236 wildcard-in-all NSSubset+cardinality, ibm allGroup si02/03,
  restrictionOfComplexTypes si02, wild051 multi-wildcard notQName gap). flatGroup
  replaced flatElementGroup.
- SESSION 2026-06-13 part 4 (5624 → 5628, +4): ATTRIBUTE WILDCARD SUBSET. New
  parser/wildcard.go implements the Wildcard Subset relation (cos-ns-subset,
  §3.10.6.2: namespaceConstraintSubset) + Wildcard allows Expanded Name
  (cvc-wildcard-name: wildcardAllowsName) — variety cases 1-4 plus the
  disallowed-names (notQName, ##defined/##definedSibling) conditions. Wired in
  finishComplexTypes: a complexContent/simpleContent RESTRICTION whose declared
  attribute wildcard is not a subset of the base's (or that adds a wildcard the
  base lacks) is a derivation-ok-restriction error (checkAttrWildcardRestriction
  in restrict.go). +4 (wild020-022 attr ns subset, wild057 drops base
  ##defined). Element-wildcard particle subsumption still deferred (part 5).
- SESSION 2026-06-13 part 3 (5613 → 5624, +11): PARTICLE RESTRICTION
  (cos-particle-restrict, §3.4.6.4 / §3.9.6) — first cut. parser/restrict.go:
  checkParticleRestrict runs in finishComplexTypes for each complexContent
  restriction whose OWN content model and whose BASE content model are both a
  flat all/sequence of element particles (flatElementGroup; a wildcard,
  choice, group ref, nested group, or open content makes it give up and report
  NOTHING — those wildcard-subset/choice/open cases stay tolerated, no false
  positives). For flat element models the per-name count an instance may carry
  ranges exactly over [Σ minOccurs, Σ maxOccurs], so these NECESSARY conditions
  for L(R) ⊆ L(B) are checked: every restriction element maps to a base
  particle (exact name OR substitution-group membership via accepted() closure;
  ambiguous → give up); each base particle's summed restriction occurrence
  range ⊆ its own range; required (min>0) base particles must remain present;
  the restricting type validly derives from the base's; nillability not
  widened. Type check is LENIENT on unions (involvesUnion → accept) because
  member-union flattening discarded intervening unions, so cos-st-derived-ok
  clause 2.2.4 can't be decided cleanly — simple011/014/015 stay tolerated.
  +11 cases (saxon All all202-205, all212-215 [seq restricting all], all223/224
  [substitution-group split occurrence], all227 [child type not derived]).
  STILL DEFERRED (give-up paths): wildcard-in-all (all229/235/236/238/244),
  attribute/element wildcard subset (wild020-022/048/051/057, ibm allGroup
  si02/03, restrictionOfComplexTypes si02), all→choice (all233), open content
  subset (open016-019/033). These need the full subsumption algorithm.
- SESSION 2026-06-13 part 2 (5597 → 5613, +16): UNIQUE PARTICLE ATTRIBUTION
  (cos-nonambig) done in full. (a) <all> groups: pairwise competition test
  (checkAllUPA in buildschema.go) — two element particles whose accepted-name
  sets, incl. transitive substitution-group closure, overlap; two wildcard
  particles with overlapping namespace constraints; element-vs-wildcard never
  compete. (b) sequence/choice: Glushkov position automaton (parser/upa.go) —
  first/follow sets; any first/follow set with two competing positions is a
  violation; <all> groups make it bail (handled by checkAllUPA). +7 cases
  (all240-243, subsgroup902/903, sg-abstract-upa). Then ALL-GROUP EXTENSION
  semantics (§3.4.2.3.3): clause 4.2.3.2 merges all+all into one <all> (was a
  sequence) so re-added elements trip UPA; clause 4.2.3.3 + cos-all-limited.1
  reject all-vs-sequence extension; cos-particle-extend.3.1 requires the
  extension <all>'s minOccurs == base's. +9 cases (all302/303/305/309-313, ibm
  openContent s3_4_1si04). all308 (mixed-empty, spec bug 6202) left tolerated.
  Helpers: wildcardsOverlap/namespacesIntersect/namesOverlap/allGroupTerm.
- SESSION 2026-06-13 part 2 (5597 → 5613, +16): UNIQUE PARTICLE ATTRIBUTION
  (cos-nonambig) done in full. (a) <all> groups: pairwise competition test
  (checkAllUPA in buildschema.go) — two element particles whose accepted-name
  sets, incl. transitive substitution-group closure, overlap; two wildcard
  particles with overlapping namespace constraints; element-vs-wildcard never
  compete. (b) sequence/choice: Glushkov position automaton (parser/upa.go) —
  first/follow sets; any first/follow set with two competing positions is a
  violation; <all> groups make it bail (handled by checkAllUPA). +7 cases
  (all240-243, subsgroup902/903, sg-abstract-upa). Then ALL-GROUP EXTENSION
  semantics (§3.4.2.3.3): clause 4.2.3.2 merges all+all into one <all> (was a
  sequence) so re-added elements trip UPA; clause 4.2.3.3 + cos-all-limited.1
  reject all-vs-sequence extension; cos-particle-extend.3.1 requires the
  extension <all>'s minOccurs == base's. +9 cases (all302/303/305/309-313, ibm
  openContent s3_4_1si04). all308 (mixed-empty, spec bug 6202) left tolerated.
  Helpers: wildcardsOverlap/namespacesIntersect/namesOverlap/allGroupTerm.
- SESSION 2026-06-13 part 1 (5572 → 5597, +25): explicitTimezone-valid-restriction
  (§4.3.16.5, required/prohibited can't widen); circular substitution groups
  (e-props-correct.6); substitution-group type-derivation (e-props-correct.4 —
  member type must be validly derived from head, not only un-excluded);
  Wildcard Properties Correct rule 4 (w-props-correct §3.10.6 — notQName names
  must lie in a namespace the constraint allows); defaultAttributes name
  collision is a duplicate (ct-props-correct.4, was silently deduped);
  extension mixed-consistency (cos-ct-extends.1.4.3.2.2.1) + empty-extension
  copies base content type incl. mixed per §3.4.2.3.3 (fixed a latent
  kitchenSink fixture that extended a mixed base with element-only content);
  cos-all-limited clauses 2 + 1.3 (nested model groups in <all> must be <all>
  and occur exactly once); reject malformed QNames with empty prefix/local in
  src-qname (wild039).
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
REMAINING gaps (~60 after the 2026-06-13 session), all the genuinely hard /
large features. UPA is DONE (part 2); cos-particle-restrict is PARTLY done
(part 3) — remaining cases below:
- Particle restriction "subsumption" / cos-particle-restrict: DONE for flat
  all/sequence element models (part 3), single-base-wildcard element/wildcard
  models (part 5), and attribute wildcard subset (part 4). REMAINING (give-up
  paths, still tolerated): MULTI-base-wildcard combination reasoning (all238,
  all244, wild048 — ≥2 base wildcards with overlapping namespaces, the
  pathological cases), all→choice subsumption (all233 — choice compositor
  gives up), processContents ordering in NSSubset (part of all238). These need
  proper choice MapAndSum/RecurseLax and a multi-wildcard cardinality solver.
- EDC type-table / wildcard cases wild069/078/079/081 + wild041 (xsi: in
  notQName).
- Open content extension/restriction subset (open016-019/030/033/046/048,
  complex018) — open-content particle subsumption, deferred.
- all308 (xs:all extension of mixed empty content — spec bug 6202, borderline;
  leave tolerated).
- Wildcards cos-aw-* intersection/union/subset (Wild restriction subset, EDC
  type-table cases wild069/078/079/081) + wild041 (xsi: in notQName).
- Open content extension/restriction subset (open030/046 extension,
  remaining ibm openContent si05/06 extension).
- CTA / assertions XPath (saxon CTA 6, ibm typeAlternatives 5): needs an XPath
  engine — effectively out of scope; candidates for skip: lines.
- 2 override false positives (over009 double-override dup; over030 false
  mg-props-correct cycle — override-internals bugs) + over014 + iri-001
  (custom DTD entities &URI; — encoding/xml limitation, skip candidate).
- simple011/014/015 (union derived-by-restriction substitutability —
  particle/type-derivation), complex018 (open content restriction subset).
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
