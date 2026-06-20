# Working notes (restart context)

Self-notes for resuming work. Keep updated at every checkpoint commit.

## Goal
Implement PLAN.md (XSD 1.1 parser, packages `xsd`, `builtin`, `parser`,
`parser/xmltree`), then run the W3C suite via the M9 ratchet harness and
baseline `testdata/xsd11-expectations.txt`.

## >>> DONE 2026-06-20 — particleEmptiable: implement §3.9.6.3 and wire at cos-valid-default.2.2 / src-ct §3.4.2.2 — 5672/21429 held <<<
Implemented the Particle Emptiable relation (XSD 1.1 §3.9.6.3) in
parser/restrict.go: `particleEmptiable` recursively decides whether a
particle can match the empty sequence, and a new `mgEmptiable` helper
consolidates the compositor switch (eliminating duplication between the
`*xsd.ModelGroup` and `*xsd.GroupRef` arms). An empty `choice` group
now correctly returns `true` (ETR minimum = 0, §3.8.6.6). Wired at two
call sites previously guarded only by a TODO:
 - buildterms.go (cos-valid-default §3.3.6.3): element with mixed
   content and a non-emptiable particle now rejected at schema build time.
 - buildcomplex.go (src-ct §3.4.2.2 clause 2): simpleContent restriction
   of a mixed-content base with a non-emptiable particle now rejected.
The deferred sub-phrase "mixed-emptiable check for value constraints +
simpleContent rest of mixed CT (emptiable part)" is removed from the
Deliberate deferrals list in the M3 section below. Ratchets held 5672 / 21429.

## >>> DONE 2026-06-17 — new leaf pkg xsdtemporal: one home for date/time/duration parsing — 5672/21429 held <<<
The date/time + duration value space and its lexical parsers were duplicated in
spirit across two places: xsdtype/datetime.go+duration.go (rigorous, the value
layer) and xpath/eval.go (a sloppy parseISODate/parseISOTime/stripTimezone that
fed fn:*-from-* and "failed open"). Unified into ONE pure leaf package
`xsdtemporal` (stdlib-only imports) that both reach into.
 - xsdtemporal owns: DateTime (7-property model) + DateTimeKind, Duration, ALL
   calendar math (daysInMonth/isLeap/dayNumber/civilFromDays/AddDuration),
   timeline + the two partial-order Compares, and every lexical parser
   (ParseDateTime/Date/Time/GYear/GYearMonth/GMonth/GDay/GMonthDay → (*DateTime,
   error); ParseDuration). It can't import xsd, so it defines its OWN
   `Order` (int, -1/0/1 mirroring xsd.Order); Compare returns it. HasTimezone()
   stays a method — *DateTime satisfies xsd.TimezoneAware structurally.
 - builtin/xsdtype is now a THIN adapter (temporal.go): type aliases
   (DateTime/Duration/DateTimeKind = xsdtemporal.X + Kind* const re-exports),
   ParseDuration passthrough, and the eight ParseX(s, xsd.ValueContext)
   (xsd.Value, error) wrappers that lift *DateTime via liftDT (guards the
   typed-nil-in-interface trap on error). value.go CompareValues converts the
   leaf Order with xsd.Order(o). builtin.go UNCHANGED (still calls
   xsdtype.ParseDateTime etc).
 - xpath/eval.go dateComponent now calls xsdtemporal.ParseDate/Time/DateTime
   (rigorous, shares the value layer's reader) and reads dt.Year/.Month/... ;
   seconds via dt.Second.Float64(). Deleted parseISODate/parseISOTime/
   stripTimezone. xpath gains ONE internal import (a pure leaf), so it stays
   "depends only on a leaf". Strictly more correct: same components on valid
   input, still fails open on parse error.
 - Temporal tests moved with the code: xsdtemporal/temporal_test.go (parse,
   compare, day-pinning, FuzzParseDuration/DateTime, using local Order);
   xsdtype keeps only decimal/CompareValues tests. git mv preserved history on
   datetime.go/duration.go. Full suite green; ratchets held 5672 / 21429.

## >>> DONE 2026-06-15 — architecture review refactors (2 of 2) — 5672/21429 held <<<
Post-conformance structural cleanups from an architecture review (no behaviour
change; both pinned by the ratchets).
 1. WILDCARD ADMISSION unified into ONE canonical impl in package `xsd`
    (`xsd/wildcard.go`): `(*Wildcard).AllowsNamespace` / `AllowsName` (static:
    namespace constraint + literal notQName) and `.Allows(q, WildcardContext)`
    (adds ##defined/##definedSibling via Defined/DefinedSibling lookups). Was
    implemented 5× (xsdwalk.WildcardAllows+namespaceAllowed, parser.namespaceAllowed
    [literal dup], parser.wildcardAllowsName, automaton.wildcardOK, assess.
    attrWildcardAllows). Callers wire the context they have (matcher: global-elem +
    sibling; attr assessor: global-attr, DefinedSibling nil). parser keyword consts
    now alias xsd.WildcardDefined/Sibling. Commit 7941fa8.
 2. XPATH: ONE parser + ONE lexer. The strict parser (parse.go/lexer.go/kindtest.go)
    now BUILDS the evaluator's AST (the exprNode types in eval.go) while keeping its
    strict syntax errors + TypeRef recording; the second lenient lexer+parser
    (lexExpr/exprParser in eval.go) is DELETED. Single `parseTree()` (xpath.go) feeds
    both Parse (schema-time: TypeRefs+syntax err) and evalExpr (instance-time: walks
    AST). KEY preservation trick: the old eval parser failed as a UNIT (first
    out-of-subset construct → whole expr fail-open), so a single parser-level
    `p.unsupported` flag reproduces it exactly — treat-as, intersect/except, node
    comparison (is/<</>>), non-atomic sequence types, kind-test node tests are parsed
    for validity (Parse sees no false error) but flag the expr; evalExpr fails open on
    it. -438 LOC in pkg; 8.3M-exec fuzz of Parse+EvalBool, no panic. Commit 44d2020.

## >>> DONE 2026-06-15 — greedy matcher: explicit content beats open-content wildcard (+1 open025) — INSTANCE SUITE NOW CLEAN <<<
Instance ratchet 21428 → 21429 (+1: saxon open025); schema held 5672. THE LAST
verdict-wrong instance case. The instance suite is now FULLY conformant: 21429
pass + 2 skip (assert011 doc()); the remaining 3 of 21434 error before producing a
verdict (infra-excluded, ungated).
Root cause: the matcher's occurrence loop was reluctant (tried exiting the loop
before consuming another occurrence), so after one `i` it would exit `i+` and let
the interleave open-content wildcard absorb an invalid `<i>42.3</i>` (skip →
unvalidated) instead of the explicit `i` (which validates it as integer → fails).
Fix: make the loop GREEDY — try to consume one more occurrence BEFORE exiting, so
the explicit content model takes every child it can and the wildcard only fills
gaps (XSD 1.1 §3.4.4.2). UPA-determinism makes the explicit model unambiguous, so
greedy never needs to backtrack a match. This is STATE-sensitive (unlike the
name-based guard tried+reverted earlier): in open025.v1 a later `<i>` after a `<d>`
still goes to the wildcard because `i+` is closed by then. KEY RESULT: clean across
all 21434 instance cases AND the 5672 schema suite — no regressions despite this
being the core content-model matcher used everywhere.

## >>> DONE 2026-06-15 — empty content type forbids whitespace (+1 open012) <<<
Instance ratchet 21427 → 21428 (+1: saxon open012); schema held 5672. A complex
type whose particle can never match an element (e.g. <xs:sequence/>) and has no
open content is an EMPTY content type (§3.4.2), and cvc-complex-type.2.1 admits NO
character content — not even whitespace (vs element-only clause 2.3, which allows
whitespace). assessElementContent: when !mixed && no open content &&
!particleCanMatchElement(particle), reject any non-"" charContent. New
particleCanMatchElement helper (recurses groups/group-refs; nil or maxOccurs=0 →
empty). Feared regressions did NOT materialise — no other instance has whitespace
in an empty-particle element-only type expecting valid; only open012 moved.

## >>> DONE 2026-06-15 — mutual/circular override no longer drops the replacement (+1 over023) <<<
Instance ratchet 21426 → 21427 (+1: saxon over023); schema held 5672. A mutual
override (over023 overrides over023a; over023a overrides over023 back, empty body)
was losing the `doc` element entirely → "no global element declaration". Root
cause in loader.register: the transitive suppress() of a replacement's overridden
original walks the target's composition closure (target.targets), which in a CYCLE
leads back to the OWNING document and wrongly suppressed the owner's OWN
replacement. Fix: seed suppress's visited set with rep.owner — a replacement
suppresses originals through the target closure but never the component in its own
document. One-line change (loader.go:238); schema ratchet held (this touches
override/redefine suppression broadly), so no regression.

## >>> DONE 2026-06-15 — schema-level defaults skip xs:override-declared types (+3) <<<
Instance ratchet 21423 → 21426 (+3: saxon open045/open043, ibm s3_4_2_4ii08);
schema held 5672. A complex type declared inside <xs:override> belongs to the
OVERRIDDEN document (§4.2.4), so THAT document's schema-level defaults apply — not
the overriding schema's. BOTH defaultAttributes AND defaultOpenContent follow this
(open043's schema comment "applies to types within xs:override" is the test THEME;
the instance annotations confirm the default does NOT apply — beta's open content /
dag come from the overridden doc, which has none).
 - builder.overrideTarget map[*xmltree.Node]*schemaDoc, populated in buildSchemas
   from l.reps (kind=="override": each top-level override child node → rep.target).
 - New b.defaultsDoc(n, doc) returns the override target for an override-child node,
   else doc. Used by applyDefaultAttributes (defaultAttributes) AND
   fillElementOnlyContent (defaultOpenContent) — only the DEFAULTS source swaps;
   namespace/reference resolution still uses the overriding doc. open045 (<b
   xml:lang> via dag now rejected), open043 (<b><extra/></b> via open content now
   rejected), s3_4_2_4ii08.

## >>> DONE 2026-06-15 — EDC against base-type local decls (+1 wild068) <<<
Instance ratchet 21422 → 21423 (+1); schema held 5672. Extends the dynamic EDC
check (checkDynamicEDC) to the BASE-type chain: when a restriction drops an
element the base declared and lets a wildcard absorb it, the element keeps the
base's locally declared type, so a wildcard-matched <e> must be validly derived
from that type. zang restricts zing dropping zing's local `e`=union(date|time)
onto a ##local lax wildcard; the global `e` is xs:duration, so `<e>PT12H</e>`
(duration, NOT date/time) is now correctly INVALID (was accepted via the global).
 - xsdvalidate/assess.go: new baseLocalDeclTypes(ct) walks ct + its BaseType
   chain, unioning each level's localDeclTypes (most-derived name wins).
   assessElementContent now takes ct and uses it; checkDynamicEDC unchanged.
   Safe for extensions (base elements already in the merged particle) — only
   restriction-absorbed names gain a constraint. Pinned by the instance ratchet.

## >>> DONE 2026-06-15 — union pattern vs member whiteSpace (+1 simple085) <<<
Instance ratchet 21421 → 21422 (+1); schema held 5672. A pattern facet on a union
(restriction) is matched against the value as normalized by the VALIDATING
MEMBER's whiteSpace, not the raw lexical — the union itself has no whiteSpace, so
the member determines normalization (cvc-pattern-valid; Saxon issue 2247). This
makes unions consistent with atomics (which already do whiteSpace→pattern).
 - xsd/facets.go parseValue: Stage-2 pattern check now SKIPS unions (deferred).
 - buildValue VarietyUnion: after a DirectMember validates the value, the union's
   EffectiveFacets().PatternGroups are checked against m's whiteSpace-normalized
   value; a member that validates but fails the pattern is passed over. New
   patternsMatch() helper (shared with Stage 2). simple085: union member is
   xs:string whiteSpace=collapse, so "  Hello   world" → "Hello world" matches
   pattern "Hello world". Low risk: differs from before only when a value has
   excess whitespace AND a collapsing member — pinned by the instance ratchet.

## >>> DONE 2026-06-15 — xsi:schemaLocation hint loading (+2 targetNS) <<<
Instance ratchet 21419 → 21421 (+2); schema held 5672. Closes the two multi-schema
false-REJECTs (sun targetNS00101m1_p, ST_targetNS00101m2_p) where the instance's
root element (and its xsi:type) is declared only in a 2nd schema document the
testGroup's schemaTest does NOT list — the instance names it via
xsi:schemaLocation. A conforming processor follows that hint (§4.3.2).
 - NEW reusable primitive parser.SchemaLocationHints(root) (parser/instancehints.go,
   non-test): returns the locations from the root's xsi:schemaLocation (the
   location half of each (ns, loc) pair) + xsi:noNamespaceSchemaLocation, in doc
   order. Unit test parser/instancehints_test.go.
 - HARNESS wiring (instance_suite_test.go): buildGroupSchema now also returns the
   resolved group doc PATHS; new buildSchemaFromDocs(paths) (the loader dedupes by
   URI, so a doc named in both the group and the instance hint loads once); new
   instanceSchemas(docPath, groupSchemas, groupPaths) parses the instance, and when
   xsi:schemaLocation adds docs beyond the group's, builds an augmented per-instance
   schema — otherwise reuses the shared group schema (build only re-runs when hints
   genuinely extend the set, so negligible perf cost). KEY: hint paths resolve
   relative to the INSTANCE dir, group paths relative to the testSet dir; filepath.
   Join cleans both to the same string for a shared file, so dedup works.
 - NOT yet wired into cmd/goxsd5 -validate (would need a multi-root parser.Parse);
   the SchemaLocationHints primitive is ready for it. Remaining instance wrong: 8.

## >>> DONE 2026-06-15 — broad instance push: date fns, cast-as, empty-choice, nested-all (+6) <<<
Six general fixes across xpath + the xsdwalk matcher, after the assertion cluster.
Instance ratchet 21413 → 21419 (+6); schema held 5672. Diagnostic: throwaway
parser/zz_gapdiag_test.go (GAPDIAG=1) listed every verdict-wrong instance with its
errors; deleted after. Started from 16 actionable wrong (19 total − 3 that error
out before a verdict); ended at 10.
 - fn:*-from-date/dateTime/time FAMILY (xpath/eval.go: year/month/day-from-date·
   DateTime, hours/minutes/seconds-from-time·DateTime). dateComponent() parses the
   ISO lexical (stripTimezone + parseISODate/parseISOTime). +2 saxon vc002.n1 /
   vc007.n1 (year-from-date($value) eq 2008 now evaluates vs fail-open accept).
 - "cast as T" OPERATOR (was only castable-as/instance-of): evaluates the cast
   VALUE — xs:boolean → a real bool (so its effective boolean is the value, not
   "non-empty string"); numeric kinds → number; else the lexical. A non-castable
   value is a dynamic error. +2 ibm typeAlternatives_004 (@a cast as xs:boolean CTA
   selecting an asserting alternative) / vc_007. cast-as is used 12× in the suite.
 - EMPTY <xs:choice/> matches the empty LANGUAGE, not ε (automaton.go matchGroup):
   was `return cont(pos)` (treated as nullable), now `return false`. A required
   occurrence is unsatisfiable; emptiness only via a minOccurs=0 wrapper (handled
   in matchParticle before the group). +1 saxon complex022 (`<z/>` vs `<choice/>`).
 - NESTED xs:all FLATTENING (automaton.go matchAll + flattenAll): an xs:all may
   ref a named group whose content is itself an xs:all (cos-all-limited's one
   nesting). allAccept only handled element/wildcard terms, so a group-ref-to-all
   never matched. flattenAll expands such members into the outer all's list
   (propagating an optional wrapper's minOccurs=0). +1 saxon all007.
Unit tests: xpath TestEvalTypedAtoms extended (date fns + cast-as); xsdwalk
TestMatcher gained empty-choice + nested-all rows.
REMAINING 10 instance wrong (all specialised hard cases, triaged — NOT general):
 - xsi:schemaLocation loading (sun targetNS00101m1_p, ST_targetNS00101m2_p): the
   instance names its root's schema only via xsi:schemaLocation (a 2nd doc the
   testGroup doesn't list). Needs the validator/harness to load instance-hinted
   schemas — crosses the xsdvalidate↔parser layer (xsdvalidate has no loader) and
   the harness builds one schema per GROUP not per instance. Clearest "next"
   general fix but structurally invasive; deferred.
 - open-content GREEDY precedence (saxon open025): the open-content wildcard must
   not absorb a child the EXPLICIT model can consume AT THE CURRENT STATE. After
   one `i`, our backtracking matcher exits `i+` and lets the interleave wildcard
   grab an invalid `<i>42.3</i>` (skip→unvalidated). A name-based "wildcard never
   matches a sibling name" guard was TRIED and REVERTED: it broke open025.v1 /
   open047.v3 where a sibling-named element legitimately goes to the wildcard once
   the explicit model can't place it (state-sensitive, not name-based). Needs a
   real greedy/deterministic automaton — substantial matcher rework.
 - EMPTY content type forbids whitespace (saxon open012.n3): `<xs:sequence/>` must
   reduce to the EMPTY content type (cvc-complex-type.2.1 — no char content, not
   even whitespace), but we build it as element-only (2.3 allows whitespace).
   Risky: could regress many whitespace-in-empty-element instances. Deferred.
 - override SCOPING of defaultOpenContent/defaultAttributes (saxon open043/open045)
   + CIRCULAR override (over023) + defaultAttributes-in-override (ibm s3_4_2_4ii08):
   xs:override interaction with schema-level defaults; specialised.
 - dynamic EDC vs BASE local decl (saxon wild068): a wildcard-matched element whose
   name has a local decl in the restriction's BASE type must validate against that
   base-local type (zang drops zing's local `e`=union(date|time); `<e>PT12H</e>`
   duration must fail). Extends the existing EDC infra to base-local decls.
 - pattern-on-union vs member whiteSpace (saxon simple085): a pattern on a union
   restriction applies to the value AFTER the validating member's whiteSpace
   normalisation. Needs reordering whiteSpace/pattern for unions; debatable
   (Saxon issue 2247) and risky for other union+pattern cases.

## >>> DONE 2026-06-15 — typed $value + recursive assertions: ALL assertion gaps closed (+12) <<<
Closed the entire remaining Assertions/XPath instance cluster. Instance ratchet
21401 → 21413 (+12); schema held 5672. The ONLY remaining assertion miss is
assert011 (uses doc() — external document access, disallowed by spec+Saxon), now
a curated skip (first two `skip:` lines in instance-expectations.txt). After this
the instance suite is at 21413 pass / 2 skip / 19 wrong, and NONE of the 19 are
assertion/XPath (they're open-content, VC version-control, multi-schema, etc.).
Two-package change (xpath + xsdvalidate); unit test xpath TestEvalTypedAtoms +
the W3C instance ratchet pin every case.
 - TYPED $value (the crux): the xpath evaluator gained a typed-atom layer so a
   schema-typed $value compares with the right semantics. EvalContext.TypedVars
   (map name→[]TypedAtom{Lexical,Kind}); AtomKind ∈ {Untyped,Number,String}.
   Internal typedItem; asNumber() refuses a KindString atom even when the lexical
   looks numeric. Fixes the ibm whitespace cases (assert_021/023): $value of
   xs:string is KindString so "\n   100\n" eq '100' is a STRING compare → false →
   INVALID (was: both atomized to the number 100 → wrongly equal). valueKind() in
   xsdvalidate picks Number iff Fundamentals().Numeric, else String (dates compare
   as ISO strings, which orders chronologically when forms match — no date type
   needed). KEY GOTCHA: do NOT force String on numeric types (10>5 would break as
   "10"<"5"); classify by Fundamentals.Numeric.
 - RECURSIVE assertion check by variety (xsdvalidate assess.go assertSatisfied,
   replaces the flat checkSimpleAssertions): LIST → each item validated against
   ItemType incl. its assertions, then the list's own assertions over the item
   sequence; UNION → the value must satisfy SOME DirectMember *including that
   member's assertions* before the union's own apply, and if no member accepts it
   → INVALID. This is what was missing: assertion facets live on the member/item
   types, not the union/list, so they were never evaluated (fail-open accept).
   +4: assert_030 (list-of-MYINT, odd item), assert_031 (union MYINT|date, "3"
   fails MYINT assert and isn't a date), assert_032 (union MYINT|MYDATE, far-future
   date fails MYDATE), assert_034 (list-of-MYDATE on an ATTRIBUTE — also wired
   checkSimpleAssertions into validateAttrValue). typeHasAssertions() guards so
   assertion-free types are untouched (zero behaviour change).
 - SIMPLE-TYPE assertions have NO context item (§ datatypes): only $value is
   defined, so `.`, a relative path, position(), last() raise a DYNAMIC error.
   New errDynamic (distinct from errUnsupported): EvalBool maps errDynamic →
   (false,TRUE) — a definite "assertion unsatisfied" — vs errUnsupported →
   (false,false) fail-open. EvalContext.NoContextItem drives evaluator.ctxAbsent.
   Verified SAFE: a suite-wide grep found ONLY assert-simple008/009/010 use
   `.`/position/last in a simpleType xs:assertion (nothing we pass relies on a
   context item there). +3 (assert-simple008 `. castable`, 009 position(), 010
   last() — all want INVALID via the run-time error).
 - FAILED CONSTRUCTOR CAST is errDynamic too: xs:date("2001-01-01!!!") → assertion
   false (was errUnsupported→fail-open). +1 assert-simple007.
 - data() FUNCTION + complex-assert $value: checkAssertions now binds $value to
   the simple-content value sequence (typed); data() atomises a bound sequence.
   GOTCHA/regression-and-fix: a general data() broke assert014/017/022 (valids)
   whose `data(.) instance of xs:date` probes typed-ness of NODES we don't model —
   so data() is restricted to already-atomic operands (data($value) ok; data(node)
   stays unsupported→fail-open). +2: assert010 ($value gt xs:date(@startDate),
   simpleContent extension), assert_035 (count($value) le 5 / every $x in
   data($value) satisfies $x mod 2=0).
REMAINING assertion/XPath false-accepts: NONE (only assert011/doc(), skipped).

## >>> DONE 2026-06-15 — XPATH assertion evaluator push (+31 instance cases) <<<
Eight commits enriching the XPath subset (xpath/eval.go) + two engine rules, all
ratcheting the instance suite up with the schema suite held at 5672 throughout.
Instance ratchet 21370 → 21401 (+31). Each landed as its own commit + re-baseline.
Diagnostic: throwaway parser/zz_instdiag_test.go (INSTDIAG=1) listed every
verdict-wrong instance with its schema's assert/alternative/simple-assertion
exprs; deleted after.
 - "STAY IN SUBTREE" (§3.13.4.1): the assertion XDM root is the PARENTLESS
   element being assessed, so a leading "/" or "//" is rooted at a non-existent
   document node → EMPTY sequence (was: leading "//" walked the element's own
   subtree). A relative "//" step now expands descendant-OR-self (was strict
   descendants, dropping the step node's own children/attrs), and the attribute
   step honours "descend". +5 (ibm d4_3_15ii12/17/31/32 count(//x) eq N; v15
   count(.//@attr1)). Key: ALL absolute-// suite instance tests expect INVALID
   (StayInSubtree category), so "absolute → empty" is safe.
 - SEQUENCE "(a,b,c)" + RANGE "to" + UNION "|"/"union" (precedence: Range
   between Comparison/Additive, Union between Multiplicative/Unary). +4 (ibm ii03
   count(@a|@b); saxon assert007/008/008a .n3 — "result = (seq)").
 - COMPLEX {assertions} ACCUMULATE down the derivation chain (ENGINE fix,
   buildcomplex.go mergeComplexType): {assertions} = base's followed by own
   (§3.4.2.3.2/3, BOTH extension+restriction). Was storing own-only, so a base
   assertion was dropped on the derived type. +4 (saxon assert006.n1, 007/008/
   008a .n2). Unit TestComplexAssertionChain.
 - SIMPLE-TYPE xs:assertion FACETS now evaluated (assess.go checkSimpleAssertions
   wired into validateSimpleContent): $value bound to the whiteSpace-normalized
   value (evaluator gained "$name" var refs, EvalContext.Vars map[string][]string),
   element as context node so "." atomizes to the same value. ATOMIC scope only —
   list $value, position()/last() stay fail-open. +10 (ibm ii20/22/24, vc_001/
   003/005; saxon assert-simple003/004/006, over010). Unit TestEvalVars.
 - current-date()/current-dateTime()/current-time() return today's ISO string;
   ISO dates of equal form compare chronologically via string cmp (suite values
   are far past/future so verdict is run-date-independent). +3 (ibm ii11/ii13
   "current-date() le @date" date 2000-12-12; saxon assert-simple001 value 2080).
 - POSITIONED-NODE rewrite of the path machinery (xpath/eval.go, the big one):
   tree navigation is NOT exposed by the Node interface or infoset (downward-only
   walk preserved). The evaluator threads *nodeCtx{node,parent,index} built as it
   descends; in a tree each node has one parent and downward nav always reaches it
   via that parent, so the synthesized parent IS the real one — and the parentless
   assertion root makes reverse/sibling axes of the context empty (stay-in-subtree
   for free). Added: ALL 12 AXES (child/descendant[-or-self]/attribute/self/parent/
   ancestor[-or-self]/following[-preceding]-sibling/following/preceding; reverse
   axes emit reverse-doc-order so positional preds work); explicit "axis::test" +
   node()/text() (lexer splits "::" and ".." while keeping single ":" in QNames);
   PRIMARY-LED paths ($w/following-sibling::*, (e)/step, fn()/step); QUANTIFIED
   some/every...satisfies + for...return with a runtime node-valued var scope
   (shadows EvalContext.Vars); position()/last() (focus threaded); distinct-values;
   path steps de-dup by node identity + per-context-node predicates. +4 (saxon
   assert007/008/008a .n1 quantified+following-sibling; assert005.n1 preceding::).
   This CLOSED the navigation "blocker" from the same-day earlier note. Unit
   TestEvalAxes. All prior xpath tests passed unchanged through the rewrite.
 - LIST-type $value bound as the item SEQUENCE (assess.go checkSimpleAssertions,
   strings.Fields when Variety==VarietyList). +1 (saxon assert-simple005
   count($value) eq count(distinct-values($value))).
REMAINING XPath false-ACCEPTS (hard / out of scope): date assertions on list-union
types (ibm assert_032/034); simple-type position()/last() where the focus is the
VALUE sequence not nodes (saxon assert-simple009/010); xs:date()-cast comparisons
(assert-simple007/008). The ibm assert_021..035 groups bundle many (schema,
instance) pairs in ONE testGroup — a HARNESS-structural limitation, not XPath.
The ~17 other false-ACCEPTS are non-XPath features (open content / VC version-
control / wildcard / defaultAttributes / Complex022), out of scope here.

## >>> DONE 2026-06-15 — scattered false-ACCEPT fixes: substitution type-block, fixed-mixed, fixed-ID, union facets, nilled ws <<<
Five small spec fixes after the EDC cluster. Instance ratchet 21360 → 21370
(schema held 5672). Each landed as its own commit + ratchet re-baseline.
 - SUBSTITUTION honours head TYPE's block (xsdwalk substChain): exclude set now
   folds in typeBlock(head's type) — a member derived from the head type by a
   method the head TYPE blocks (complexType block=) is not substitutable
   (§3.3.6.3). +3 (sun ElemDecl disallowedSubst00503).
 - FIXED on MIXED content (assess.go cvc-elt.5.2.2.1): a fixed value on a mixed
   content type forbids element children and requires char content == fixed. +2
   (sun valueConstraint00701/00801).
 - FIXED attribute ID harvest (assess.go): an absent optional attr contributes
   its value constraint whether default OR fixed; new attrValueConstraint. A
   fixed xs:ID absent on two elements → duplicate ID. +2 (saxon Id011/013).
 - UNION validation via DirectMembers (xsd/facets.go buildValue): was flattening
   through an intervening restriction-of-union (BasicMembers), bypassing its
   pattern/enum. Now tries DIRECT members, each validating through its own facets.
   +2 (ibm union ii02/ii04). (simple085 — pattern-vs-collapsed-whitespace on a
   union — is a SEPARATE whiteSpace-ordering issue, still open.)
 - NILLED element strictness (assess.go cvc-elt.3.2.1): a nilled element must
   have NO character children incl. whitespace (was hasContent = non-whitespace
   only). +1 (saxon All004.n02).

## >>> DONE 2026-06-15 — attribute ##defined + tighter EDC for wildcard-matched elements <<<
Two more false-ACCEPT clusters (Wild). Instance ratchet 21344 → 21360 (schema held
5672). Both in xsdvalidate/assess.go.
 - ATTRIBUTE wildcard ##defined (cvc-wildcard clause 2.2): an attr wildcard with
   notQName="##defined" must not admit a globally-declared attribute. New
   attrWildcardAllows wraps WildcardAllows + a.v.attrs lookup. +6 (saxon
   Wild054/055/056/058/059/060). (##definedSibling is illegal on attr wildcards.)
 - DYNAMIC EDC for wildcard-matched elements (XSD 1.1 §3.9 "locally declared type"
   / cvc-assess-elt; test category TighterMatchingRuleForEDC): when a lax/strict
   wildcard matches an element whose name ALSO names an element decl in the content
   model, the element's ACTUAL governing type (after xsi:type/subst, = res.Types[el])
   must be validly derived from that locally declared type; a name with NO global
   decl is governed by the local type directly (xsi:type-overridable). New
   localDeclTypes + assessLocallyTyped + checkDynamicEDC; resolveXSIType made
   nil-decl-safe. +10 (saxon Wild061-064/067/075/076, ibm edcWildcard). KEY GOTCHA
   (two wrong cuts before the right one): it's NOT "local type replaces global"
   (broke wild063: -12 valid as local integer but must fail global positiveInteger)
   and NOT "global DECLARED type derived from local" (broke wild064.v2: global
   decimal not derived from local integer, but xsi:type=int IS) — it's the
   POST-xsi:type governing type (res.Types[el]) derived from the local type.
REMAINING false-ACCEPTS after this (~64) are dominated by fail-open assertions
(Assert/assertion/assert ≈40, need richer XPath) + scattered hard cases. The 7
false-REJECTS are unchanged (XPath axis / multi-schema / override / union-pattern-ws).

## >>> DONE 2026-06-15 — open-content inheritance + ##defined/##definedSibling wildcard exclusions <<<
Two more clusters. Instance ratchet 21315 → 21344 (schema held 5672). After this,
only 7 instance false-REJECTS remain (see list at end); the ~80 false-ACCEPTS are
dominated by fail-open assertions (Assert/assertion ≈40) + a few hard cases.
 - EXTENSION open-content inheritance (parser, §3.4.2.3.3): an extension's
   explicit content type carries the BASE's {open content} (clause 4.2.3); the
   own <openContent>/<defaultOpenContent> layers per clause 6.1 (absent or
   mode='none' ⇒ keep the base's — an extension can NOT suppress inherited open
   content, mode='none' is not removal) / 6.2 (authored ⇒ wildcard UNION of own +
   base, own's mode). New inheritExtensionOpenContent (restrict.go) wired into
   mergeComplexType before the narrowing check. +6 (saxon Open027/031/047, incl.
   the mode='none' Open031). Unit test TestExtensionOpenContentInheritance.
 - ##defined / ##definedSibling matcher exclusions (xsdwalk/automaton.go,
   cvc-wildcard clauses 2 & 3): the matcher ignored these context keywords so a
   negative wildcard never excluded anything. New wildcardOK wraps WildcardAllows:
   ##defined ⇒ name must not resolve to a global element (LookupGlobal); 
   ##definedSibling ⇒ name must not match any element decl in the content model,
   directly OR implicitly via substitution groups (collectSiblings gathers the
   model's decls per Match; matchesSibling uses SubstitutableFor). All five
   matcher WildcardAllows sites now call m.wildcardOK. +23 (wg substitution-groups
   sg-and-defined-Sibling-*, saxon Wild negatives). Unit test
   TestMatcherDefinedSiblingWildcard.
REMAINING 7 instance false-REJECTS (2026-06-15): (1) assertion count(.//@attr1)
— needs descendant-attribute XPath axis; (2) vc_007 — vc:* version-control attr;
(3) All007 — all-group cvc-particle; (4) Override over023 — xs:override element
not surfaced as global; (5) simple085 — pattern vs collapsed whitespace value;
(6) ElemDecl targetNS00101m + (7) SType ST_targetNS00101m2_p — multi-schema /
aux-namespace type not loaded (likely needs the group's 2nd schemaDocument or
xsi:schemaLocation; verify harness loads both docs).

## >>> DONE 2026-06-15 — instance false-REJECT clusters: root xsi:type, union derivation, cos-aw-union <<<
Three follow-on commits after the cvc-id batch, each a clean spec rule. Instance
ratchet 21302 → 21315 (schema held 5672 throughout). Same INSTDIAG diagnostic
method (throwaway parser/zz_instdiag_test.go, deleted).
 - ROOT xsi:type with no element decl (assess.go assessRoot): a validation root
   lacking a governing element declaration but carrying xsi:type has a governing
   TYPE definition (cvc-assess-elt clause 8) and is assessed against it instead of
   being rejected "no global element declaration". New resolveRootXSIType; the
   no-decl-no-xsitype path still errors. +4 (sun CType/SType targetNS, ST_name).
 - DerivationOK union transitive membership (xsdwalk/query.go, cos-st-derived-ok
   clause 2.2.4): D is validly derived from a union B if derived from a type in
   B's transitive membership AND B + every intervening union have EMPTY facets
   (2.2.4.3). New derivedFromUnionMember + unionFacetsEmpty (checks pattern/enum/
   assertion on EffectiveFacets). So xsi:type may name a union member, but not
   reach through a restriction-narrowed union. +4 (ibm union v05/v07, saxon
   Simple012/016 valids); simple016.n01 STAYS invalid (xs:date only reachable
   through the pattern-restricted intervening union dt). FIRST cut used
   BasicMembers (flattened) and over-accepted — DirectMembers + facet gate is key.
 - cos-aw-union (parser/wildcard.go + buildcomplex.go mergeComplexType): an
   EXTENSION's {attribute wildcard} is the UNION of base + own wildcard (was: own,
   or base as fallback — the long-deferred TODO). New wildcardUnion with §3.10.6.3
   namespace clauses (Not∪Not=Not(∩)→any if empty; Enum∪Not=Not(diff)→any if
   empty; Enum∪Enum=Enum(∪)) + setUnionNot helper. {disallowed names} via
   notQNameUnionDisallowed: a name disallowed by one survives only if the OTHER
   does not allow it (wildcardAllowsName); ##defined/##definedSibling keyword
   survives only if in BOTH. +9 (saxon Wild013-016/045/046/083, sun suntest
   test008); wild046.n2 STAYS invalid (xml:lang excluded). GOTCHA: first cut used
   notQName intersection and wrongly accepted xml:lang — the spec rule is
   "not allowed by the other", not "in both". Unit test TestBuildAttrWildcardUnion.

## >>> DONE 2026-06-15 — instance cvc-id rework + value-constraint defaults + xs:ENTITY <<<
Closed the biggest instance false-REJECT/false-ACCEPT cluster (cvc-id family).
Instance ratchet 21241 → 21298 (+57); schema held at 5672. All changes in
`xsdvalidate/assess.go` unless noted. Diagnostic method: a throwaway
`parser/zz_instdiag_test.go` (INSTDIAG=1) listed every verdict-wrong case by
testSet/kind; deleted after. Wrong-count 190 → 133 over this session.
 - cvc-id HARVESTING now recurses variety (collectID): list → per item, union →
   first member that ParseValue-validates the lexical (the actual {member type}).
   Fixes list-of-ID (s3_3_4v10/13/14/15/16/30), union-of-ID (v17–v21), IDREFS
   subsumed. Previously only atomic ID/IDREF + list-of-IDREF were seen.
 - cvc-id UNIQUENESS is now per-binding-ELEMENT, not per-value (XSD 1.1 §3.3.4.5).
   `ids` is `map[string]Element` (value → bound element). Same value on the SAME
   element (two ID attrs, repeated list items) is allowed; only a clash between
   DISTINCT elements errors (saxon id001/id004). The binding owner differs by
   source: an ID ATTRIBUTE binds to its element; an ID in ELEMENT SIMPLE CONTENT
   binds to the element's PARENT. Threaded a `parent Element` param through
   assessElement→assessType→assessComplexType→validateSimpleContent (and
   assessElementContent passes `el` as the children's parent, assessChild relays).
   collectID(t, lexical, pos, ctx, owner): ctx = ns-context element, owner =
   binding target. cvc-id.1: an element-content ID whose owner is nil (the
   validation ROOT's own content, parent out of scope) → "binds to no element in
   scope" error (ibm instance_invalid ii26/ii27 now correctly INVALID).
 - VALUE-CONSTRAINT DEFAULTS on empty content now applied (cvc-elt.5.1.1):
   validateSimpleContent, if text=="" && !hasElementChildren && decl has Default
   (else Fixed), substitutes that value before validating + ID-harvesting. Also
   absent ATTRIBUTE uses with a default now harvest their ID (assessAttributes
   tail loop, owner=el). Fixes idIDREF v11/v26/v28/v29 + sun valueConstraint
   00201/00301/00501/00601/01101 + saxon Id pattern/minLength-on-"" cases.
 - xs:ENTITY/ENTITIES REFERENTIAL VALIDITY: new plumbing. xmltree now collects
   unparsed (NDATA) general-entity names from the internal DTD subset
   (parseDTDUnparsedEntities + entityDeclEnd/unparsedEntityName helpers in
   xmltree.go), stores them on the ROOT `Node.UnparsedEntities map[string]bool`.
   xsdvalidate gained an optional `DocumentInfo{ UnparsedEntities() }` capability
   (infoset.go); xmlsrc.element implements it; assessRoot reads it (haveEntities
   flag → fail-open if a source omits DTD info). collectID ENTITY branch errors
   (SpecDatatypeValid) when a value names no declared unparsed entity. Cleanly
   separates saxon Id id020/id021 valid (entity1 declared) from invalid (only
   entity2 declared) — and was REQUIRED because applying the element default
   "entity1" exposed the missing check (those two .n01 cases had been passing for
   the wrong reason: empty "" failed the NCName pattern).
   Unit tests: xmltree_test.go TestUnparsedEntities / TestNoUnparsedEntities.
   No bespoke xsdvalidate unit test — the instance ratchet pins these exact W3C
   cases (id001/id004/ii26/idIDREF v*/id020) directly.
REMAINING instance gaps (133 wrong, mostly false-ACCEPT by design): assertions/CTA
fail-open needing richer XPath (saxonMeta/ibmMeta assert*), attribute-wildcard
negative cases (Wild ##other/notQName), open-content suffix/interleave matching
(Open027/031/047), xsi:type union-member derivation (union v05/v07, Simple012/016),
"no global element declaration" for some sun targetNS/Override groups. Next
tractable clusters: xsi:type union derivation, attribute-wildcard, open content.

## >>> DONE 2026-06-15 — cos-aw-intersect + defaultOpenContent appliesToEmpty <<<
Both were identified as "obvious next parser TODOs" after the V0–V4 instance
validation session (NOTES block above, lines ~67-75): the matcher bug fix
exposed that ~10 instance negative cases were passing for the wrong reason.
 - cos-aw-intersect (§3.10.6.3): when multiple `<attributeGroup ref>` children
   contribute wildcards, the effective {attribute wildcard} is their INTERSECTION
   (not the first one). Also applies to the type's own `<anyAttribute>` when
   it appears alongside group refs. New `wildcardIntersect(w1, w2 *xsd.Wildcard)`
   in parser/wildcard.go (plus helpers `notQNameUnion`, `stringsUnion`,
   `stringsDifference`, `stringsIntersection`). The intersection per the five
   spec clauses: both-Any=Any; one-Any+other=other; Not∩Not=Not(union);
   Not∩Enum=Enum(difference); Enum∩Enum=Enum(intersection). processContents uses
   `min()` (strict=0 < lax=1 < skip=2). NotQName takes the union (either
   wildcard's exclusions apply to the intersection). Fixes: wild023.n1, wild024.n1,
   wild025.n3, wild026.n1/.n4, wild043.n3, suntest test007.2.n/.7.n/.8.n.
   WIRING: buildAttrUses (buildterms.go) now calls `wildcardIntersect(wc, g.Wildcard)`
   for each `<attributeGroup ref>` (was: `if wc==nil { wc=g.Wildcard }`) and
   `wildcardIntersect(wc, b.buildWildcard(c, doc))` for the `<anyAttribute>` case
   (was: `wc = b.buildWildcard(c, doc)` which overwrote group wildcards).
   cos-aw-union (extension's wildcard union with base) remains deferred (the
   comment in mergeComplexType is unchanged).
 - defaultOpenContent appliesToEmpty (§3.11.4.2): a bare `<xs:sequence/>` with no
   child elements has an EMPTY content type (particleMatchesNonEmpty=false), not a
   non-empty element-only type. Changed the condition in fillElementOnlyContent
   (buildcomplex.go) from `particle != nil` to `mixed || particleMatchesNonEmpty(particle)`.
   `mixed=true` means the content type is always non-empty (mixed variety), so it
   gets defaultOpenContent regardless of `appliesToEmpty`. Fixes: open012.n1.
   open012.n3 remains a gap — the instance validator allows whitespace in content
   type "empty" (a separate xsdvalidate issue, not the parser's model).
CONFORMANCE: schema suite held at 5672; instance suite +10 (21231 → 21241).
Unit tests: TestBuildAttrWildcardIntersect (4 sub-cases: Not∩Not, Any∩Not,
Enum∩Not, Enum∩Enum via group+anyAttribute) and TestBuildDefaultOpenContentAppliesToEmpty
(4 sub-cases: empty-false=no-OC, empty-true=OC, mixed-false=OC, non-empty-false=OC).

## >>> DONE 2026-06-14 — xmltree.Parse streams (no unbounded io.ReadAll) <<<
`parser/xmltree.Parse` used `io.ReadAll` twice (raw + transcoded) and held the
whole document in `treeParser.data` — an unbounded-memory vector on untrusted
input. Now streaming, keeping the *Node DOM (callers need random tree access):
 - `ParseLimit(r, uri, maxBytes)` added; `Parse` calls it with `MaxDocumentBytes`
   (1 GiB default). `maxBytes <= 0` = unlimited. Limit enforced mid-stream by
   `boundedReader` (errors past max), not read-all-then-check.
 - Transcoding streams (`charset.NewReader(limitReader(r))`); no raw/data copies.
 - Only the PROLOG is buffered: `readProlog` consumes XML decl/comments/PIs/DOCTYPE
   (incl. internal subset) into a small buffer, leaves root `<` unread, then
   `io.MultiReader(prolog, br)` feeds the decoder. `parseDTDEntities` runs on just
   the prolog. `consumeDoctype` tracks quotes + `[ ]` depth + `<!-- -->` so a
   `]`/`>` inside an entity value or comment doesn't end the subset.
 - Line/col: `lineReader` records newline offsets as bytes flow to the decoder
   (O(lines) index, not O(doc) byte slice); correct because the decoder reads in
   document order so the index always covers any offset pos() queries.
 - CAVEAT (unchanged): peak memory is still O(doc) — dominated by the returned
   DOM, not byte buffers. The size cap is the real memory bound. Sublinear would
   need a SAX API + reworking loader/xmlsrc (declined).
 - Tests added in xmltree_test.go: DTD entities, tricky entity value (]/>/comment),
   prolog misc, BOM strip, UTF-16 transcode, ParseLimit. Full suite + 8s fuzz green.

## >>> DONE 2026-06-14 — instance validation V0–V4 implemented (PLAN-validate.md) <<<
The whole instance-validation plan landed in one session (3 milestone commits +
1 refactor commit). New packages:
 - `xsdwalk` (pure-leaf-ish, imports only xsd): the shared model algebra —
   content-model `Matcher` (automaton.go: occurrence/sequence/choice/all,
   wildcards, substitution groups, open/mixed; CPS-backtracking) + queries
   (query.go: DerivationOK, SubstitutableFor, WildcardAllows, AttributeUse).
 - `xsdvalidate`: the cvc-* assessor over an abstract infoset
   (`Element`/`Attribute`/`Node` — Node is an OPEN marker `any`, NOT a sealed
   interface, so format adapters in other packages can implement it). assess.go
   is the engine; idc.go identity constraints; assert.go assertions+CTA;
   result.go PSVI-lite. `xsdvalidate/xmlsrc` adapts parser/xmltree.
 - XPath evaluator lives in `xpath` (eval.go, the "evaluator half" the plan
   named) over an abstract `Node`/`NodeAttr`/`EvalContext.Castable` — xpath
   stays a PURE LEAF (no xsd/builtin dep); xsdvalidate/xpathadapt.go adapts.
KEY DECISIONS / GOTCHAS (so a resume doesn't relitigate):
 - Value-space delegated to the type via `(*SimpleType).ParseValue` (engine never
   type-switches on datatype). Document-scoped rules (cvc-id ID/IDREF, IDC,
   assertions) live in the engine.
 - New cvc-* SpecRefs in xsd/specref.go (cvc-elt/type/complex-type/attribute/au/
   particle/wildcard/id/identity-constraint/assertion) + CONFORMANCE.md rows
   (matrix guard enforces row+ref+impl-file existence — keep in sync).
 - IDC: selector/field matched by LOCAL NAME (IdentityConstraint carries no ns
   context); field values compared in the VALUE SPACE via each node's type with
   a collapsed-string fallback (so 5==5.0 but string"3.0"!=decimal3.0); singleton
   list == atomic; skip-wildcard subtrees excluded from constraint scope.
 - Assertions/CTA FAIL OPEN: any unsupported XPath construct → assertion
   satisfied / type-table alternative unmatched → never a false rejection. This
   is what makes the ratchet safe to grow. XSD 1.1 inheritable attributes are
   threaded for CTA; added `xsd.AttributeUse.Inheritable` (use-level overrides
   decl; parser buildAttrUse populates).
 - Matcher bug fixed: an empty/nullable REQUIRED particle (`<xs:sequence/>`
   minOccurs=1) now matches the empty sequence (was failing cvc-particle). This
   fix exposed pre-existing DEFERRED parser gaps on ~22 negative cases that had
   been passing for the wrong reason: cos-aw-intersect (attribute-wildcard
   intersection, buildterms.go ~348 "first wildcard stands in for it") and
   defaultOpenContent appliesToEmpty. Those are the obvious next parser TODOs.
CONFORMANCE: parallel `testdata/instance-expectations.txt` ratchet (TAB-separated,
mirrors schema ratchet; `go test ./parser -run TestInstanceConformance
-update-instance-expectations`). 21434 instance cases, 21231 verdict-correct.
Schema ratchet unchanged (5672). cmd/goxsd5 gained `-validate doc.xml`.
REMAINING GAPS (future, not blocking): richer XPath (parent axis `..`, more fns)
for more assert/CTA cases; cross-scope keyref; simple-type-level assertions; IDC
true namespace resolution; whitespace in empty content type (xsdvalidate gap —
open012.n3); override + defaultOpenContent/defaultAttributes inheritance for types
defined within xs:override (open043.n2, open045.n2). cos-aw-intersect +
defaultOpenContent appliesToEmpty are DONE (2026-06-15, +10 instance passes).

## >>> PLAN 2026-06-14 — instance validation plan authored (PLAN-validate.md), DONE (see block above) <<<
Wrote `PLAN-validate.md`: extend the schema processor
into an in-process, **format-pluggable** schema-validity assessor (XSD 1.1 Part 1
§3 `cvc-*` rules). Two NEW top-level packages, named to the `xsd*` theme:
 - `xsdwalk` — shared model algebra (content-model automaton + model queries:
   subst-group acceptance, wildcard match, governing-type after xsi:type/CTA,
   attribute-use lookup). Serves a PUSH/exhaustive walk (codegen/docs/diff) AND a
   PULL/demand-driven walk (validator, generated-deserializer runtime). The
   reusable core is the algebra, not the driver.
 - `xsdvalidate` — the assessor: pull-driver + cvc-* actions over an abstract
   infoset interface (`Element`/`Attribute`/`Node`); XML is just one source
   (`xsdvalidate/xmlsrc` over `xmltree`; `jsonsrc`/`bersrc` future, each inventing
   a documented format→infoset mapping). Engine never imports an XML package.
KEY DESIGN RULINGS (settled in plan):
 - `xsd` stays a PURE LEAF — infoset abstraction lives ABOVE it. **No
   `ValidateInstance` method on model types** (would force xsd→infoset dep + is
   contextual/stateful/cross-cutting). Locality rule: **value-space** (lexical/
   facet/QName) stays on the type via existing `(*SimpleType).ParseValue` (already
   emits cvc-* SpecRefs, `xsd/facets.go:301`) — engine never type-switches on
   datatype; **document-scoped** (cvc-id ID/IDREF, keyref, assertions) in engine.
 - Milestones V0 harness → V1 simple/local → V2 content-model → V3 identity
   constraints → V4 assertions/CTA (optional). Conformance via the SAME ratchet
   harness: the vendored W3C suite already has ~28k `<instanceTest>` entries;
   parallel `testdata/instance-expectations.txt`.
 - Future codegen (M10) = compiler back-end on the SAME front-end; shares `xsd`
   model + `builtin` value-space + `xsdwalk`; separate memory/execution model +
   infoset interface. Value-space/structure-space symmetry: builtin may get
   `GenXMLMarshaller`/`GenXMLUnmarshaller` (compile-time analog of `ParseValue`).
Nothing built yet — no code, no NOTES checkpoint to follow. Resume by reading
`PLAN-validate.md` then starting V0.

## >>> DONE 2026-06-14 — testdata/xsdtests converted to a git submodule <<<
Replaced the bespoke `testdata/fetch-xsdtests.sh` fetch-on-demand script with a
proper git submodule at `testdata/xsdtests` → https://github.com/w3c/xsdtests.git,
pinned at the same revision (gitlink 7bc3365c652a322f3d762021b3879eb92dae7e30).
The local checkout was already that exact SHA with the upstream remote, so
`git submodule add` staged it in place (no re-clone). The ~230 MB suite still
stays out of this repo's history — only the gitlink + .gitmodules are committed.
Changes: removed `/testdata/xsdtests/` from .gitignore (submodule add refuses an
ignored path); added .gitmodules; `git rm testdata/fetch-xsdtests.sh`. Doc/test
refs updated to `git submodule update --init testdata/xsdtests`:
conformance_suite_test.go (header comment + skip message), testdata/README.md
(new checkout + revision-bump recipe), PLAN.md M9 (×2). NOT marked `shallow` in
.gitmodules — a recorded non-tip SHA + shallow fetch is fragile; standard init
reliably checks out the pinned commit. Conformance held: 5709 cases / 5672
recorded passes / 26 skips against the submodule. To bump the suite: checkout the
new SHA inside the submodule, `git add testdata/xsdtests`, re-baseline
expectations with `-update-expectations`, all in one commit.

## >>> DONE 2026-06-14 — comprehensive Go native fuzz testing added <<<
Added 9 `func Fuzz*` targets (Go 1.18+ native fuzzing) across all four
input-bearing surfaces, each with a curated seed corpus and invariants beyond
mere no-panic. NEW FILES (all `*_test.go`, no production code touched):
 - xsdregex/fuzz_test.go — FuzzTranslateRegex (output always RE2-compilable +
   deterministic), FuzzCompileRegex (matcher non-nil & safe on any subject).
 - builtin/xsdtype/fuzz_test.go — FuzzParseDecimal (self-eq, canonical String()
   round-trip, non-negative digit counts), FuzzParseDuration, FuzzParseDateTime
   (drives all 7 Kind*), FuzzCompareValues (antisymmetry a?b == -(b?a) +
   Equal/Compare agreement on two parsed decimals).
 - parser/xmltree/fuzz_test.go — FuzzParse (non-nil root, non-negative Pos,
   walkable tree, no nil children), FuzzIsNCName (true ⇒ no colon).
 - parser/fuzz_test.go — FuzzParseSchema (full public Parse via in-memory
   mapResolver; no nil schema elements), FuzzParseValue (every builtin.AllBuiltins
   type × arbitrary lexical, nil ctx is handled at builtin.go:270, never panics).
FINDING (no real bug): FuzzParseSchema flagged that an over-strong "schemas slice
never nil" assertion was wrong — a root doc that fails to load returns nil,error
by design (parser.go:32 loadRoot path). Relaxed the test to the true invariant
(no panic + no nil *elements*) rather than touch correct code. go vet + gofmt
clean; full normal suite green (seed corpora double as unit tests under `go
test`); ~10M+ execs across extended 30s runs of the deepest targets (schema
build, regex translate, value parse), zero crashers, no testdata/fuzz corpus
generated. Run one with e.g. `go test ./xsdregex/ -run '^$' -fuzz FuzzTranslateRegex`.

## >>> DONE 2026-06-14 — primitiveFundamentals map → authored on the type <<<
Same anti-pattern the package-factoring refactor already killed for the parser's
name-keyed primitiveFacetMask, but it had survived in core xsd: a `map[string]
Fundamentals` keyed by primitive local name ("string"/"decimal"/"dateTime"/…),
hard-coding XSD datatype identity inside the supposed-to-be datatype-agnostic
PURE LEAF, and silently handing a custom primitive the zero value
{OrderedFalse,false}. Removed it. The authored base case now lives as a single
field xsd.SimpleType.FundamentalBase *Fundamentals, set ONLY on primitives in
builtin (next to Applicable) via the fund* presets (fundUnordered/fundDecimal/
fundFloating/fundTemporal, shared *xsd.Fundamentals); the primitive() ctor takes
it. WHY the full Fundamentals struct and not just {ordered}/{numeric}: a zero
Fundamentals{} is exactly a string-like primitive, and EVERY primitive is
unbounded + countably infinite, so the stored {bounded}=false /
{cardinality}=CountablyInfinite are TRUE of the primitive itself (not stale
derived state) — it's the more complete, self-consistent description.
Fundamentals() now does `f = *prim.FundamentalBase` off PrimitiveType() then
recomputes {bounded}/{cardinality} from the effective facets (up-the-chain
derivation already there). Shared preset pointers are safe: Fundamentals() copies
by value, nothing mutates through them. Field is named FundamentalBase (not
Fundamentals) to avoid colliding with the method. nil for list/union + atomic
ur-types (no primitive) → §F defaults. Conformance held 5672 / 26; build + vet +
full suite + gofmt clean.

## >>> DONE 2026-06-14 — derived-state cleanup: 4 redundant fields removed <<<
Follow-on to the package-factoring refactor: stripped stored state that was
either dead or derivable from the authored source of truth, so divergence
between authored and derived data is now structurally impossible. Conformance
held at 5672 / 26 skips at every step; build + full suite + gofmt clean.
 1. xsd.Facets.HasEnumeration (bool field) → method HasEnumeration() returning
    len(Enumeration) > 0. The field only ever distinguished "present but empty"
    from "absent", but a present-but-empty enumeration arises ONLY from error
    recovery (every <enumeration> child either appends a value or emits
    enumeration-valid-restriction and continues), so by the time it occurs ≥1
    error is already reported — pure suppression state, deleted. Only behavioral
    delta: a NOTATION type whose enum values all fail to parse now also gets the
    "must constrain with enumeration" error (cascade on an already-invalid schema).
 2. SimpleType.Facets (the memoized effective-facet cache) DELETED. EffectiveFacets()
    now derives on demand: MergeFacets(base.EffectiveFacets(), &DeclaredFacets) up
    the chain, plus whiteSpace=collapse/fixed injected for VarietyList (Part 2
    §4.3.6) from the variety rather than stored. DeclaredFacets is the sole authored
    facet state; every build/clone/mutate site stopped writing the cache (parser
    applyRestriction/buildSTList, builtin primitive/restrict/list, xsdedit
    RestrictWith/AddEnumeration/AddPattern). parseValue reads t.EffectiveFacets().
    The old "hot path" justification for the cache is moot: ParseValue is called
    only during schema construction (validating facet/enum values) and from xsdedit,
    never in an instance-validation loop, and there are no benchmarks.
 3. SimpleType.MemberTypes DELETED — fully dead (declared, never read or written
    anywhere). The flattened basic members are derived by BasicMembers() from the
    canonical un-flattened DirectMembers, which every caller already uses.
 4. SimpleType.IsBuiltin DELETED + PrimitiveType() reworked. PrimitiveType() now
    detects the primitive boundary by "nearest atomic ancestor with Applicable != 0"
    (Applicable is authored ONLY on primitives, single write site builtin.go) instead
    of the structural "IsBuiltin && base==anyAtomicType" match — cheaper (no per-call
    QName), principled, and future-correct for custom primitives (they'll carry an
    authored Applicable without needing to masquerade as builtin). IsBuiltin's other
    role (mutation protection) moved to xsdedit.reservedType(t) = Name.Namespace is
    XSDNS || XSINS; NOTE the XSINS arm is load-bearing — xsiSchemaLocationType lives
    in XSINS, so a bare XSDNS check would have silently stopped protecting it.
 FUTURE (deferred per decision this session): xsdedit.NewPrimitive(name, base,
 parse, compare, applicable, ws) constructor + a Validate rule (primitive needs
 Applicable!=0 and non-nil Parse) — only worth adding when builtin/gotype or another
 real custom-primitive consumer lands; YAGNI until then.

## >>> DONE 2026-06-14 — deferred-SpecRef triage items A + B implemented <<<
Both genuine-but-unexercised static holes (A: redefine group/attributeGroup
self-reference occurrence; B: ag-props-correct clause 2 at the attribute-group
definition level) are now wired, zero-ratchet as predicted (5672 held). NONE of
the umbrella/N-A/vacuous refs below were touched — they remain correctly deferred.
 - [A] parser/loader.go checkRedefineSelfBase now dispatches group →
   checkRedefineGroupSelfRef and attributeGroup → checkRedefineAttrGroupSelfRef
   (the old default:return TODO is gone). Group: recursive descendant scan for
   <group ref=self> with NO <element> ancestor (clause 6.1 gate); >1 ⇒ SpecSrcRedefine
   (6.1.1), each must have occurs()==(1,1) (6.1.2). attributeGroup: direct-child
   scan for <attributeGroup ref=self>; >1 ⇒ SpecSrcRedefine (7.1). IMPORTANT
   correction to the original plan text: count==0 is VALID (the 6.2/7.2 restriction
   case), so we do NOT "require count==1" — only count>=2 and bad occurrence are
   errors; the no-self-ref existence requirement is already covered by the orig==nil
   check in registerReplacement. Stays appealed to src-redefine (SpecSrcExpredef
   unchanged / umbrella). Tests: TestRedefine +5 (2 valid: no-self-ref restriction,
   one attrGroup self-ref; 3 invalid: two group self-refs, group self-ref
   maxOccurs=2, two attrGroup self-refs).
 - [B] parser/buildterms.go buildAttributeGroup now calls checkAttrGroupDefDups
   (emits SpecAGPropsCorrect) — a pass over the definition's DIRECT <attribute>
   children only (nested <attributeGroup ref> checked where resolved; prohibited
   uses skipped), using attrUseName so the expanded names match the real
   {attribute uses}. Covers the never-referenced-definition case the post-merge
   ct-props-correct.4 check misses. Tests: TestBuilderNegatives +1 (dup name) +
   TestBuildAttributeGroupDefValid (distinct names + ref repeating a name = clean).
CONFORMANCE.md: src-redefine row updated (occurrence checks done; restriction-
subset still deferred); ag-props-correct deferred → wip.

## >>> DONE 2026-06-14 — package re-factoring: model made datatype-agnostic <<<
Decoupled the built-in datatype value-spaces (and the regex engine, and the
mutation API) from the core `xsd` package so the model no longer hard-codes any
specific datatype. `xsd` is now a PURE LEAF (go list -deps shows no internal
deps). Conformance held at 5672 / 26 skips at every step; full suite + vet clean.

NEW PACKAGE LAYOUT:
 - xsd (core)          — model + facet engine + value abstractions only. 7 files.
 - builtin/xsdtype  → xsd     — concrete value spaces (String/Boolean/Decimal/
     DateTime/Duration/…), CompareValues/Equal, ParseDecimal/ParseDateTime/etc.
 - xsdregex (leaf)            — CompileRegex/TranslateRegex + ucblocks (pure, zero deps).
 - xsdedit → xsd, xsdregex    — RestrictWith/AddEnumeration/AddPattern/AddElement/Validate.
 - builtin → xsd, xsdtype, xsdregex ; parser → +builtin.

WHAT CHANGED & WHY:
 1. Value is now `type Value any` (was sealed `interface{ isValue() }`). The seal
    made user-defined value spaces impossible; opening it is the enabler for
    custom primitives. The facet engine no longer type-switches on concrete
    value types — it discovers capabilities through interfaces in xsd/value.go:
      Lengthed{ Len() int }  (stdlib-style; presence == "length applies"),
      DigitCounted{ TotalDigits()/FractionDigits() }, TimezoneAware{ HasTimezone() }.
    The 4 old switch sites in facets.go (length/digits/timezone) now use these;
    enumeration now compares with t.compareFunc() not the global Equal (lets a
    custom value space's own equality drive enumeration — also a latent fix).
 2. Applicability of constraining facets is now a first-class model property:
    SimpleType.Applicable (FacetSet, authored ONLY on primitives in builtin) +
    SimpleType.ApplicableFacets() method. Replaces parser's name-keyed
    primitiveFacetMask (which gave unknown/custom primitives a wrong allFacets
    pass). NOTE: Applicable ≠ DeclaredFacets — Applicable is which facet KINDS
    may be declared (a uint bitset); DeclaredFacets is the facet VALUES set.
 3. Default comparator: CompareValues moved to xsdtype, wired onto
    xs:anySimpleType.Compare in builtin; every type roots there, so compareFunc()
    resolves it through the chain. Core's compareFunc() default is `incomparable`
    (0,false). buildValue's old `String(norm)` fallback is gone (returns an error;
    anySimpleType's identity Parse covers every real atomic type).
 4. regex KEPT IN CORE INITIALLY then moved out: the only core→regex caller was
    mutate.AddPattern, so moving the mutation API out first freed regex to become
    the leaf xsdregex package with no cycle. xsd no longer references regex at all.
 5. Mutation API: methods → FREE FUNCTIONS in xsdedit (Go can't define methods on
    xsd types from another pkg). API CHANGE: t.RestrictWith(f) → xsdedit.RestrictWith(t,f),
    ct.AddElement(...) → xsdedit.AddElement(ct,...), t.Validate() → xsdedit.Validate(t).
    EffectiveFacets() stayed a method on xsd.SimpleType (core Fundamentals() uses it).
    The two old private deps (parseFunc/describe) are inlined over the exported surface.
 6. NEW: xsdedit.Validate(t) — registration-time intrinsic checks for
    programmatically-built types (which skip the parser's schema-level checks):
    atomic ⇒ Parse resolvable; order/enum facets ⇒ a Compare resolvable; declared
    facets ⊆ base.ApplicableFacets(); list/union structural sanity. Tests in
    xsdedit/edit_test.go (TestValidate*).
FUTURE (xsdtype/gotype): a sibling builtin/gotype package was discussed for
Go-native (lax) value semantics vs xsdtype's strict ones — not built yet.

## >>> NEXT SESSION — phase-9 iri-001 pushed 5671 → 5672 (+1). Remainder below <<<
REMAINING 2 GAPS after phase 9 (GOXSD5_CONFORMANCE_GAPS=1 to list); the hard
tail is now EXHAUSTED — what is left is 2 spec bugs (correctly tolerated). No
more genuinely-fixable gaps known:
 - SPEC BUGS, leave tolerated: saxon all308 (xs:all extension of mixed empty
   content, bug 6202); saxon complex018 (open content restriction subset, bug
   16786).

SESSION 2026-06-13 phase 9 (5671 → 5672, +1): the former "encoding/xml
limitation" (wg IRI iri-001) is FIXED, not out of scope. The schema uses an
internal DTD subset (`<!DOCTYPE xs:schema [ <!ENTITY URI "…"> … ]>`) and
references those custom general entities inside xs:pattern values (`&URI;`).
Go's encoding/xml does not process the internal subset, so the references were
hard errors. FIX (parser/xmltree/xmltree.go): pre-scan the internal subset in
Parse, collect `<!ENTITY name "value">` decls (first-wins; skip parameter/
external entities and comments via a quote-aware markup-decl scanner), fully
resolve each replacement text (char refs, the 5 predefined entities, nested
general-entity refs, with a cycle guard), and hand the map to
xml.Decoder.Entity — the decoder then substitutes &name; literally wherever
referenced. Verified encoding/xml substitutes the map value verbatim (an `&`
in the value is NOT re-parsed), so resolving `&amp;`→`&` etc. up front is
correct. Each type also carries an equivalent literal pattern, so the expanded
regex compiles and the schema builds → validity=valid as expected.
 - DONE phase 8 (cos-particle-restrict, wildcard shadows named in <all>):
   wild069. See phase-8 block below.

## >>> FUTURE WORK — deferred-SpecRef triage (the 7 never-emitted refs) <<<
INVESTIGATION 2026-06-14. The conformance test classes the registry's
never-referenced constants into "deferred/N-A" (allowed unreferenced, see
xsd/conformance_test.go TestSpecRefConstantsReferenced). Seven Spec* constants
register but no call site emits them: SpecCosCTRestricts, SpecCosCTDerivedOK,
SpecAGPropsCorrect, SpecMGDPropsCorrect, SpecSTRestrictsFacets, SpecSrcExpredef,
SpecAssertionsValid. Investigated each for real, un-implemented static work.
HEADLINE: conformance is EXHAUSTED (5672, only 2 spec bugs remain), so NONE of
these has a failing suite case — enforcing ANY of them moves the ratchet ZERO.
This is robustness/attribution work, not conformance work. Most need NOTHING.

OUT OF SCOPE — do NOT schedule (instance-validation rule, not schema-time):
 - cvc-assertions-valid (N/A): xs:assert/xs:assertion are STORED UNEVALUATED
   (xsd/model.go:317-318, 580-584; parser/attrcheck.go:234). cvc-* are instance
   Validation Rules; this is a schema-only validator. Becomes live only if a
   future M10+ adds instance validation. The constant exists so the matrix row
   stays complete. Leave N/A.

UMBRELLA refs — every ENFORCEABLE sub-clause already emits under a more specific
ref; wiring the umbrella too would only risk double-reporting. Annotate & leave:
 - cos-ct-derived-ok (§3.4.6): the Type Derivation OK (Complex) relation. Its
   decisions already emit — extension validity → cos-ct-extends; restriction →
   derivation-ok-restriction / cos-particle-restrict; substitutability gating →
   e-props-correct / cos-equiv-class paths (validlySubstitutable,
   derivationMethods in restrict.go). No distinct error site of its own.
 - cos-ct-restricts (§3.4.6): same shape — attribute-use-subset + content-type
   restriction conditions emit under derivation-ok-restriction (restrict.go) +
   cos-particle-restrict. Umbrella over already-wired checks.
 - st-restrict-facets (§3.16.6): facet-restriction recursion companion to
   cos-st-restricts; facet-restriction validity already emits under the specific
   *-valid-restriction facet refs (length-valid-restriction, … in
   xsd/facets_check.go) and cos-st-restricts. Umbrella.
 - src-expredef (§4.2.4): its one ENFORCEABLE clause ("in all cases there must
   be a top-level definition of the appropriate name+kind in the redefined
   document") ALREADY emits under src-redefine at parser/loader.go:327. The
   self-reference occurrence rule everyone associates with it is actually
   src-redefine clauses 6.1.1/6.1.2/7.1 (see item A). Nothing distinct to wire.

GENUINE but UNEXERCISED static holes — real missing checks, but ZERO suite cases
exercise them (so unit-test-only; LOW priority; success gate is "5672 holds +
new unit negatives pass", NOT a ratchet bump):
 - [A] Redefined <group>/<attributeGroup> self-reference OCCURRENCE constraint
   (src-redefine 6.1.1/6.1.2 group: exactly one self-ref, its minOccurs ==
   maxOccurs == 1; 7.1 attributeGroup: exactly one self-ref). This is the
   explicit TODO at parser/loader.go:361 (checkRedefineSelfBase bails on the
   group/attributeGroup case). SCOPE: extend checkRedefineSelfBase — for a
   redefined <group>/<attributeGroup>, scan its descendants for a ref= whose
   resolved QName == the redefined name+tns; require count==1 and (group only)
   minOccurs/maxOccurs both 1-or-absent; emit SpecSrcRedefine. Necessary
   condition ⇒ no false positive. Add TestOverride/redefine negatives (zero/two
   self-refs; group self-ref with maxOccurs=2). Note: appealed to src-redefine,
   so this does NOT clear SpecSrcExpredef — that one stays umbrella.
 - [B] ag-props-correct clause 2 at the attribute-group DEFINITION level (no two
   {attribute uses} share an expanded name). Today duplicate-attribute detection
   runs only AFTER merge into a complex type (parser/buildcomplex.go:370/417 →
   ct-props-correct; au-props-correct on uses), so a standalone <attributeGroup>
   that itself declares two same-named attributes and is NEVER referenced escapes
   unflagged. SCOPE: a small pass over each attribute-group definition's direct
   attribute children (NOT its nested group refs — those are checked where
   resolved) emitting SpecAGPropsCorrect on a same-expanded-name collision.
   Marginal: any USED group already trips the post-merge check; this only covers
   the unused-definition case. Lowest value of the three actionable items.

NOT actionable: mgd-props-correct (§3.7.6) is near-vacuous ({name} NCName,
{model group} present) — structurally guaranteed by pass-1; the only real
constraint, circular model groups, already emits under mg-props-correct
(parser/buildschema.go:696-726). No work; leave deferred.

RECOMMENDATION: items A and B are the only genuine code work and both are
zero-ratchet robustness with unit tests only; do them as one small commit if/
when touching the redefine + attribute-group paths, else leave. Everything else
is correctly deferred/umbrella and should be annotated, not enforced. If A/B are
declined, this block can be deleted — the registry+matrix are already self-
consistent and the tests pass.

## >>> FUTURE WORK — architecture refactor + duplicate-state removal <<<
Conformance is exhausted (only 2 tolerated spec bugs remain), so the next body
of work is structural, NOT new spec features. Retro (2026-06-13) identified that
the foundation was built in the right order but the hard semantics (UPA/EDC/
particle-restriction/open-content/CTA) were sequenced too late and accreted
around a model that was already frozen. Three tracks, each independently
shippable behind the conformance ratchet (zero pass-count change is the success
gate for C and A; B is a behavior-changing rewrite guarded by differential
testing — see Track B). Do them in C → A → B order: C is the cheapest and
de-risks A; B goes last so it reasons over already-deduplicated state.

### Track C — minimize duplicate / derivable state — DONE (2026-06-13)
All five inventory items landed behind the ratchet (5672 held throughout); each
field→method conversion is its own commit. Summary of what shipped:
 - item 5: ComplexType.Mixed → (*ComplexType).IsMixed() (ElementContent.Mixed is
   now the single canonical source; the finishExtensionParticle drift is gone).
 - item 2: SimpleType.Primitive → (*SimpleType).PrimitiveType() (walks BaseType;
   the primitive = nearest atomic builtin directly restricting xs:anyAtomicType).
 - item 1: SimpleType.MemberTypes → (*SimpleType).BasicMembers() (flattens the
   canonical DirectMembers on demand).
 - item 3 (highest value): the four fundamental facets → (*SimpleType).
   Fundamentals(). Deleted the stored Ordered/Bounded/Cardinality/Numeric fields
   and BOTH copies of the bounded/cardinality derivation (buildsimple + mutate).
   {ordered}/{numeric} now come from a §F.1 table in xsd (sole authored source,
   keyed off PrimitiveType()); {bounded}/{cardinality} compute from the effective
   facets. Nothing read the fields for logic, so conformance was unaffected;
   values are now more correct (xs:byte is {bounded}=true). +TestFundamentalsDerivation.
 - item 4: SimpleType.Facets (effective) left as an explicitly-named memoized
   cache (hot path), now DOCUMENTED as derived-not-authored with DeclaredFacets
   canonical. Per plan, a deeper rework waits for Track A.
NEXT: Track A (topo build, retire finishComplexTypes/pendingAttrs), then B
(unify restrict.go). Both still pending; see blocks below.

### Track C (original plan) — minimize duplicate / derivable state
PRINCIPLE: store a fact ONCE at its canonical source; compute every derivable
view on demand. Every stored-derived field is currently copied by hand at each
clone/build site (proof: the fundamental-facet derivation is duplicated verbatim
in parser/buildsimple.go:91-98 AND xsd/mutate.go:42-51; RestrictWith also hand-
copies Primitive + MemberTypes). That hand-copying IS the bug class — a new
clone site that forgets a field silently desyncs. INVENTORY (canonical → derived
view to replace with a method):

 1. SimpleType.MemberTypes (flattened basic members) is a PURE FUNCTION of
    DirectMembers (un-flattened {member type definitions}) — flatten member
    unions on demand. Keep DirectMembers as canonical; replace the field with
    `(*SimpleType).BasicMembers() []*SimpleType` (memoize lazily if the union-
    parse hot path in facets.go:393 buildValue shows up in a profile — schema-
    only conformance never hits instance validation hard, so probably needn't).
    Readers: facets.go:395, buildsimple.go:121/180/188, restrict.go:1039,
    mutate.go:37. This pair only exists because flattening was chosen in M6 and
    the un-flattened need surfaced 6 phases later (phase 7); collapse it.
 2. SimpleType.Primitive is derivable by walking BaseType to the self-primitive
    (the primitive's own Primitive == itself). Replace with `(*SimpleType).
    PrimitiveType()`. Readers: buildsimple.go:74/254/258/379, buildterms.go:220/
    223, mutate.go:35.
 3. Fundamental facets Ordered/Bounded/Cardinality/Numeric are pure functions of
    (base fundamentals, variety, length/bounds/enum facets). The derivation is
    ALREADY duplicated (buildsimple.go:91-98 vs mutate.go:42-51). Move it to ONE
    method `(*SimpleType).Fundamentals()` (or compute the four lazily) and delete
    the stored fields + both copies. This is the highest-value item: it kills the
    only place the codebase computes the same thing two ways.
 4. SimpleType.Facets (EFFECTIVE) = MergeFacets(base-chain, DeclaredFacets).
    DeclaredFacets is canonical. Effective is read on the value-validation hot
    path (facets.go parseValue reads t.Facets directly), so DO NOT naively drop
    it — keep it as an explicitly-named memoized cache (`effective Facets` +
    `(*SimpleType).EffectiveFacets()` already exists at the API). Document it as
    derived, not authored. Lowest priority; touch only if A makes it free.
 5. ComplexType.Mixed DUPLICATES ElementContent.Mixed (both set to `mixed` at
    buildcomplex.go:200 & :227) and the two can DRIFT — finishComplexTypes
    inherits ec.Mixed independently (buildschema.go:626) without updating
    ct.Mixed. Pick ElementContent.Mixed as canonical for element content; make
    `(*ComplexType).IsMixed()` compute it (false for SimpleContent/empty). Reader
    audit: buildcomplex.go:92, buildschema.go:613/625/626/633/635, buildterms.go:
    110, restrict.go:782/789.
SCOPE NOTE: the xsd package IS the public interface (PLAN), so field→method is a
deliberate API change — fine pre-1.0, and it makes "derived" un-spoofable. Each
removal: replace field with method, update the readers above, delete the build/
clone-site assignments, re-run the ratchet (must stay at current count), commit.

### Track A — topo-ordered type build, retire finishComplexTypes/pendingAttrs — DONE (2026-06-14)
Full shell/finish split landed; 5672 held throughout. Key finding that shaped
the design: the kitchen-sink case (a base's content references an element typed
by a DERIVED type) is a genuine cycle in the dependency graph (base content ↔
derived merge), NOT just a build-order quirk — eager recursive build handled all
FORWARD refs for free but could not linearize this cycle, which is exactly why
the merge was deferred. The split breaks the cycle by decoupling content from
derivation: content references resolve to SHELLS, so the finish recursion follows
derivation edges ONLY and terminates even when a base reaches back into a derived
type. Two commits:
 - STEP 1 (commit 2989597): extracted the three type-internal-reading element
   checks (checkNotationEnum, cos-valid-default value space + mixed,
   substitution-group derivation) out of buildElementDecl into a post-pass
   checkElementDecls (buildterms.go, driven off the b.elements map). This lets
   element refs resolve to type SHELLS without forcing a full type build.
 - STEP 2: buildComplexType now returns a SHELL (header + resolveCTBase: base +
   derivation method + final/src-ct.1 checks), enqueued on b.ctOrder. resolveType
   returns shells. finishComplexTypes drains b.ctOrder by index (anon types are
   discovered as content fills), each via finishComplexType which finishes its
   base FIRST then fills content (fillSimpleContent/fillComplexContent/
   fillElementOnlyContent) and merges INLINE (mergeComplexType: extension particle
   + attr merge, base guaranteed finished). pendingAttrs (the side-table) is gone,
   replaced by a per-type local attrMaterial. The old finishComplexTypes merge
   half is gone; its cross-type STATIC checks (EDC/UPA/restrict/open-content/
   type-alternatives) moved verbatim into runStaticTypeChecks, now called at the
   tail of finishComplexTypes after the drain (checkElementDecls runs there too,
   so cos-valid-default reads MERGED ec.Mixed — no conformance change observed).
   finishExtensionParticle's SimpleContent branch is now dead/defensive (the base
   is always finished, so fillElementOnlyContent resolves simple-content bases
   directly). NEXT: Track B (unify restrict.go).
PLAN.md M6 step 2 SAID "topologically sort types by derivation edges, build so
each base is fully built before its derivatives." The implementation instead did
per-node memo + checkTypeCycles + a finishComplexTypes post-pass, parking each
type's attribute material in builder.pendingAttrs (builder.go:39-58) because the
base can still be mid-build when a derived type is constructed. That post-pass is
a WORKAROUND for not building in topo order. Refactor: after checkTypeCycles
(chains are acyclic there), build types in topological order of the derivation
edge so a base's Content/AttributeUses are complete before any derived type reads
them; then extension-particle assembly and attribute-use merging happen inline at
build time and pendingAttrs + finishComplexTypes's merge half disappear. KEEP the
genuinely cross-type STATIC checks currently riding in finishComplexTypes (EDC,
particle-restriction, open-content, type-alternative) as an explicit post-pass —
those are validation, not construction, and SHOULD run after all types exist.
RISK: the "base content reaches back into a derived type" case (the kitchen-sink
reason the post-pass exists) — verify the shell-memoization still lets content
refs resolve to an in-progress derived type; the topo order is over DERIVATION
edges only, and element/ref recursion is NOT a derivation edge (same insight as
buildGroup phase-4d), so it terminates.

### Track B — DONE (2026-06-14). REWRITE restrict.go to one unified relation
TRACK B COMPLETE: the §3.9.6 particle-restriction relation now lives in
subsumption.go as one region/representative-name relation (particleRestrictUnified
→ slotRun for ≤1 base wildcard, regionRun for ≥2, shared nameTypeOK + unifiedShadow);
the legacy fragments are deleted and restrict.go shrank 1138 → 705 LOC (now: open-
content + attribute restriction checks, type-derivation relations, and the flat-
particle support helpers the relation is built on). 5672 held across all three
steps. STEP 3 (this step): flipped checkParticleRestrict to emit the unified
findings, deleted checkParticleRestrictLegacy + checkRestrictRun +
checkWildcardShadowsNamed + checkMultiWildcardRestrict + checkMultiWildcardRun,
and retired the differential harness (subsumption_test.go) — its legacy oracle is
gone, so the unified relation is now guarded by the conformance suite + build_test
particle-restrict positives/negatives (the message-asserting negatives passed the
flip unchanged, confirming verdict AND wording parity). Whole structural refactor
(Tracks C, A, B) is now DONE; nothing left in this FUTURE WORK block.

STEP 1 DONE (2026-06-14): strangler scaffold + differential harness landed,
5672 held. checkParticleRestrict (now in subsumption.go) is the production entry:
it runs the legacy fragments (renamed checkParticleRestrictLegacy in restrict.go,
unchanged) via a scratch-ErrorList capture (collectLegacyRestrict → restrictViolation
values) and EMITS those, and when the test hook restrictDiff is installed also
computes particleRestrictUnified and hands both to it. particleRestrictUnified is
the new §3.9.6 recursion; it natively decides the WILDCARD-FREE flat element-run
+ choice slice (unifiedRestrictRun = the Recurse case: NameAndTypeOK map + occ
subsumption + required-retained) and DELEGATES every wildcard-bearing shape back
to the legacy fragments, so output is identical until each case is ported. Diff
oracle: TestRestrictDifferentialSuite (rebuilds the whole corpus with the hook;
2602 restrictions, exact multiset agreement on ref+pos+msg) + TestRestrictDifferentialFuzz
(20000 random flat element-run pairs over a ct0<-ct1<-ct2 chain). Found+fixed a
latent nondeterminism: collectReps iterated maps, so the multi-wildcard "allows
an element (N)" diagnostic picked a random offending name; now sorted. NEXT
(step 2): port the single-wildcard NSSubset + wildcard-shadows-named, then the
multi-wildcard rep-name engine (collectReps/baseRegion), removing each delegation
while the diff stays green; preserve all238/wild048 supra-spec coverage. Then
step 3: flip emission to unified, delete the fragments + the capture/hook.

STEP 2 DONE (2026-06-14): the unified relation is fully implemented and
differentially verified; 5672 held, production STILL emits legacy (unified runs
only under the test hook — the flip is step 3). particleRestrictUnified now
natively decides ALL shapes (no delegation): it gates as before, counts base
wildcards, and dispatches the §3.9.6 cardinality rule — ≤1 base wildcard →
slotRun (NSRecurseCheckCardinality: map each restriction particle to the one base
particle it restricts, bound the occurrence sum, retain required content) + the
xs:all unifiedShadow; ≥2 base wildcards → regionRun (the multi-wildcard
rep-name/region packing, name-admissibility + disjoint-gated count + type). The
shared NameAndTypeOK triple is factored into b.nameTypeOK (the slot path checks
the type table, the region path doesn't — matching the fragments). KEY
TRANSCRIPTION BUG the fuzz caught: regionRun must RETURN after name-admissibility
when regions overlap, skipping BOTH count AND the type check (legacy
checkMultiWildcardRun does); running the type check on overlap made unified
stricter (329 fuzz divergences, all legacy-valid/unified-invalid) until gated.
Diff relaxed from exact-message to VERDICT+ref-set parity (a genuine unification
phrases diagnostics differently); green on 2602 suite restrictions + 100k fuzz
pairs (now wildcard-bearing). NEXT (step 3): flip checkParticleRestrict to emit
particleRestrictUnified, delete checkParticleRestrictLegacy + the fragment
functions (checkRestrictRun, checkWildcardShadowsNamed, checkMultiWildcardRestrict,
checkMultiWildcardRun, baseSlot, and now-unused helpers) from restrict.go, and
retire the differential harness (its legacy oracle is gone).

restrict.go accreted phase-by-phase as N necessary-condition fragments
(checkRestrictRun, checkMultiWildcardRestrict, checkWildcardShadowsNamed,
checkOpenContentRestrict, checkAttrWildcardRestriction, …), each a slice of
cos-particle-restrict. The spec describes ONE recursive relation (Particle Valid
(Restriction) §3.9.6: NSRecurseCheckCardinality / NSSubset / Recurse /
RecurseUnordered / RecurseLax / MapAndSum). DECISION (2026-06-13, user): this is
a new lib with no downstream consumers and a strong ratchet — do not fear the
risk; collapse the fragments into the unified recursion. This is allowed to
CHANGE BEHAVIOR (it decides cases the fragments give up on), unlike a pure
reorg. The complexity win is real: one recursion with shared helpers replaces 8
overlapping checks that each reinvent rep-name/region reasoning.
TWO HARD CONSTRAINTS so "simpler" doesn't mean "loses coverage":
 (1) DIFFERENTIAL TEST during the transition. Keep the OLD fragments compiled as
     a reference oracle; run both old + new on every suite restriction pair AND
     on fuzzed/generated particle pairs; assert they agree. Divergence = a bug in
     exactly one — surfaced automatically. This is what makes "we have tests to
     fall back on" load-bearing for the out-of-suite residual (the suite already
     catches in-suite false positives as ratchet regressions, but not shapes it
     never exercises). Delete the fragments ONLY once differential parity holds
     on suite + fuzz.
 (2) PRESERVE the supra-spec coverage. The fragments deliberately exceed the
     NAIVE §3.9.6 recursion: the multi-wildcard cases (all238, wild048) are
     caught by the collectReps/baseRegion representative-name solver because the
     spec's literal relation does NOT catch them. A faithful transcription would
     REGRESS those. So the unified relation must carry the rep-name engine as the
     implementation of its wildcard cases (NSRecurseCheckCardinality / NSSubset),
     not a naive transcription. Keep the "give up ⇒ sound" fallback for anything
     still unanalyzable.
Sequence B LAST (after C and A) so the model state it reasons over is already
deduplicated and the build order is clean.

SUCCESS GATE: GOXSD5_CONFORMANCE_GAPS=1 go test ./parser -run TestConformanceSuite
stays at the current pass count for ALL three tracks (C and A are behavior-
preserving; B may newly DECIDE give-up cases but must never regress a passing
case and must keep differential parity with the fragments it replaces). Keep the
unit-test + re-baseline + checkpoint-commit rhythm. Update this block as tracks
land.

SESSION 2026-06-13 phase 8 (5670 → 5671, +1): cos-particle-restrict — the
WILDCARD-SHADOWS-NAMED case in an <all> base (saxon wild069). In an xs:all
group matching is order-independent, so a named element particle for N
unconditionally takes UPA precedence over an overlapping wildcard: B always
types an <N> by that named declaration. wild069's base zing = all{ e:union(date,
time)?, f:integer, any ##local lax }, with a GLOBAL e:xs:duration; the
restriction zang = all{ f:integer, any ##local lax } drops the named e, so it
routes <e> to its lax wildcard, which binds the global e (duration). duration is
not derived from union(date,time), so <e>P1Y</e> is valid in zang but not zing →
invalid restriction. The plain NSSubset wildcard-vs-wildcard check (W_R ⊆ W_B)
misses this entirely because the wildcards are identical. ELEGANT FIX, no new
subsystem: a focused pass checkWildcardShadowsNamed (parser/restrict.go) run
right after the existing checkRestrictRun, GATED to <all> bases (baseIsAll).
For each base NAMED slot N that the restriction run drops (no R element particle
accepts N — over-counted via accepted() so it only ever suppresses, never
invents) yet some R wildcard matches: it computes the wildcard's effective bound
type via wildcardBoundType (skip → unconstrained; lax → global N's type or
unconstrained if no global; strict → global N's type, or NO valid instance if no
global) and reuses the SAME validlyDerivedByRestriction relation the named-vs-
named branch uses. Unconstrained binding errors unless the base named type is
xs:anyType. SOUNDNESS — the gate is the whole game: the shadowing-precedence
reasoning holds ONLY for <all> (order-independent). In a SEQUENCE the precedence
is positional, so an <e> past the named slot legitimately routes to the wildcard
in BOTH base and restriction — that is exactly saxon wild068 (the sequence twin,
VALID), which regressed on the first un-gated attempt and is now a positive
guard in TestBuildParticleRestrictValidModels. derivationFailureRef still picks
the appealed id: cos-st-derived-ok when the base named type is a union (wild069),
else cos-particle-restrict (the skip-unconstrained negative). Tests: +2
TestBuilderNegatives (union-base duration bind → cos-st-derived-ok; skip
unconstrained → cos-particle-restrict) and +3 valid-models (wild068 sequence
twin; <all> with a wildcard binding a VALIDLY-derived global xs:date; <all>
shadowing an anyType base element). Zero conformance regressions.

SESSION 2026-06-13 phase 7 (5667 → 5670, +3): cos-st-derived-ok CLAUSE 2.2.4
(§3.16.6.3) — UNION derived-by-restriction substitutability, the simple01x
cluster the prior NOTES flagged as "invasive, high regression risk" because the
builder flattens union members. ELEGANT FIX: instead of restructuring the whole
pipeline, ADD an un-flattened view alongside the flattened one. New field
xsd.SimpleType.DirectMembers = the spec {member type definitions} (members
exactly as declared, member-unions NOT flattened); the existing MemberTypes
stays the flattened *basic members* so nothing downstream changes. Populated in
buildSTUnion (DirectMembers = resolved members pre-flatten) and applyRestriction
(a restriction inherits base.DirectMembers per Part 2 §4.1.1 case 2). That
preserves exactly the transitive-membership + intervening-union structure clause
2.2.4 needs. validlyDerivedByRestriction (restrict.go) is no longer lenient on a
UNION BASE: when bType.{variety}=union and the ordinary base chain misses, it
calls stDerivedFromUnion, an EXACT decision of clause 2.2.4 — walk DirectMembers,
a member reached by an ordinary chain is the endpoint M (its facets free,
clause 2.2.4.2), a member union we pass THROUGH is an intervening union whose
{facets} must be empty (clause 2.2.4.3, checked at the top of the recursion via
unionFacetsEmpty); the base union's own facets are likewise required empty.
SOUNDNESS: clause 2.2.4 is the ONLY way to be validly derived from a union base
once the chain walk fails, so a false result is decisive — never a false
positive. When only the RESTRICTION type involves a union but the base does not,
still give up (accept) — the flattened model can't decide that cleanly. Catches:
 - simple011: union(date,time) is not a member of union(date,dateTime,time)
   (fails 2.2.4.2 — a smaller union is not derived from the larger).
 - simple014: xs:date IS in chap's transitive membership but chap is a
   pattern-restricted union (fails 2.2.4.3 — base union carries a facet).
 - simple015: xs:date reached through intervening union dt which carries a
   pattern facet (fails 2.2.4.3 on the intervening union).
DIAGNOSTIC: the type-derivation failure now reports under cos-st-derived-ok (its
actual appealed rule) when the base element's type is a union, else the generic
cos-particle-restrict — derivationFailureRef() picks; references the previously
unwired SpecCosSTDerivedOK so the xsd matrix guard is satisfied. Unit tests:
TestBuilderNegatives +3 (2.2.4.2 smaller-union, 2.2.4.3 facet-restricted base,
2.2.4.3 facet-restricted intervening union) and TestBuildParticleRestrictValid
Models +1 (member reached through a FACET-FREE intervening union stays valid —
guards the recursion's accept path) + repurposed the old union-tolerated case
into a real 2.2.4-holds positive. CONFORMANCE.md cos-st-derived-ok row → wip.
Zero conformance regressions.

SESSION 2026-06-13 phase 6 (5663 → 5667, +4): XPath 2.0 PARSER. New top-level
package `xpath` (xpath/lexer.go, xpath.go, parse.go, kindtest.go) — a faithful
recursive-descent parser for the FULL XPath 2.0 grammar (REC-xpath20-20101214
Appendix A, spec downloaded to docs/raw/xpath20.html, cleaned to
docs/clean/xpath20.md via docscleaner). Clean, xsd-decoupled API:
`xpath.Parse(src) (*Expr, error)` returns a typed *xpath.Error on any static
syntax error, and Expr.TypeRefs lists every cast/castable/treat/instance-of
ATOMIC type target (prefix+local, unresolved) so callers apply schema checks.
KEY DESIGN: the operator-keyword vs name-test ambiguity (and/or/div/eq/… vs an
element named "and") is resolved purely by GRAMMAR POSITION in the recursive
descent — the lexer emits bare tokName for all names and never pre-classifies
keywords; a binary level only tests for its keyword AFTER parsing a left
operand, and operand parsing (down to StepExpr) stops at any non-/ token, so the
keyword is never swallowed as a name test. Lexer greedily absorbs '-'/'.' into
names (XML NCName chars) per A.2.4.1; tokenises the whole input up front for
free lookahead (function-call vs kind-test vs name-test; leading-lone-slash;
for/some/every disambiguation). Type names inside KindTests (element(*, T),
attribute(*, T)) are NOT collected as TypeRefs — they are element/attribute type
annotations, not atomic cast targets, so a valid `instance of element(*,
xs:untyped)` is never mistaken for a non-atomic cast.
INTEGRATION (parser/xpathcheck.go, checkXPathTest, wired at the <alternative>
build site in buildterms.go): reports src-ta (§3.12.3 clause 1: the {test} must
contain no XPath 2.0 static errors) when (a) the test does not PARSE — catches
si04 (malformed: "12 5 2", "((7>=6)", "3 cast as 3", ")(", ">", …), si05 ("AND"
is not the lowercase operator "and" ⇒ leftover token), si06 ("xs1::double" ⇒
"::" is not valid inside a cast-target QName); or (b) a cast/instance-of/etc.
target resolves (quiet registry lookup, prefix via node NS scope, unprefixed via
xpathDefaultNS) to a COMPLEX type — catches ii06 ("cast as messageTypeString",
messageTypeString being a complexType). SOUNDNESS: parse-failure and
cast-to-complex are both necessary static errors ⇒ never a false positive;
anything unresolvable (unbound prefix, unknown type) is left alone per the
give-up discipline. Only <alternative> is wired, NOT <assert> — no assert
conformance case needs it, so wiring it would be pure regression risk; the xpath
package itself is proven against the FULL suite corpus regardless.
REGRESSION GUARD: xpath/xpath_test.go embeds a representative valid set (one per
production) + the complete malformed set; before integration the parser was run
against ALL 224 distinct suite <alternative>/<assert> test exprs (206 valid all
PARSE, 18 malformed all ERROR). Builder unit tests: TestBuilderNegatives +3
(malformed, uppercase-AND, cast-to-complex) and TestBuildTypeAlternativeAndAttr
RestrictValid +2 (well-formed cast/instance-of; cast to a user SIMPLE type is
fine). Zero conformance regressions; the 7 remaining gaps are unchanged (the
old tolerated tail: all308, complex018, simple011/014/015, wild069, iri-001).

SESSION 2026-06-13 phase 5e (5662 → 5663, +1): PARTICLE-RESTRICTION type-table
equivalence (cta0043). derivation-ok-restriction §3.4.6.3 clause 3 "subsumes"
clause 4.6: when a restriction element maps to a base element of the same name,
their {type table}s must be equivalent (you can't change conditional type
assignment when restricting). Added a typeTablesEqual(term, base.decl) check at
the named-element match in checkRestrictRun (restrict.go), right beside the
existing type-derivation + nillability checks. typeTablesEqual (reused from EDC)
is conservative — anonymous alternative types compare by zero name — so it only
ever MISSES a violation, never invents one. +1 (cta0043: appendixType restricts
chapType, redeclaring <stamp> with a dateTime alternative where the base had
dateTimeStamp). Unit tests: TestBuilderNegatives "restriction changes a matched
element's type table" + valid-models "restriction reproduces the type table".

SESSION 2026-06-13 phase 5d (5661 → 5662, +1): SELF-RESTRICTION cycle (over014).
"Override a complex type by self-restriction": the <override> replacement does
NOT get the redefine-style self→original scoped mapping (only redefine does, in
registerReplacement), so its base="structuredDate" resolves through the global
registry to the REPLACEMENT itself ⇒ BaseType == self. checkTypeCycles silently
broke out of its walk on `next == cur` (to tolerate xs:anyType's self-loop —
though anyType's base is actually nil, not self), so a genuine one-step
self-derivation escaped. Fix: in that branch, if cur is a *xsd.ComplexType whose
BaseType == itself, report ct-props-correct and sever to anyType. Also catches a
plain direct `<restriction base="self">`. Unit test TestBuilderNegatives
"complexType restricts itself".

SESSION 2026-06-13 phase 5c (5656 → 5661, +5): EXTENSION OPEN CONTENT
(cos-ct-extends §3.4.6.2 clause 1.4.3.2.2.3) — the cluster NOTES flagged as
needing the real {open content} mapping. checkExtensionOpenContent (restrict.go),
wired in finishComplexTypes extension branch. The full clause reduces elegantly:
clause 1.4.3.2.2.4 (BOT.wc ⊆ EOT.wc) is AUTOMATICALLY satisfied — the {content
type} mapping §3.4.2.2 clause 6.2 makes EOT's wildcard the UNION of the
extension's own and the base's, so the base's is always a subset — leaving only
clause 1.4.3.2.2.3: an extension may not narrow the base's INTERLEAVE open
content to SUFFIX. EOT mode per mapping clauses 5-6: explicit <openContent>
child wins; else <defaultOpenContent> applies when the EXPLICIT (post-merge)
content type is not empty (contentEmpty checks the merged particle, so a
non-empty base makes an empty-particle extension non-empty — THIS is W3C bug
13459 / open046); else inherit base (no narrowing). Violation iff BOT.mode==
interleave && EOT.mode==suffix. Necessary condition ⇒ no false positive. New
fields: pendingAttrs.contentNode (the <extension> node, for finding <openContent>
in the finish pass); helpers openContentModeAttr, contentEmpty. +5 (open030
default-suffix-over-base-interleave, open033 explicit suffix, open046 empty
extension + appliesToEmpty=false default-applies-post-merge, ibm openContent
s3_4_1si05/si06). KITCHEN-SINK FIXTURE was itself spec-invalid under this rule
(base interleave + defaultOpenContent suffix on the `derived` extension = the
open030 pattern) — changed pass1_test.go defaultOpenContent to mode=interleave
and updated the TestBuildComplexContentAssembly assertion. Unit tests:
TestBuilderNegatives +2 (explicit + default-via-empty-extension), valid-models
+2 (interleave kept; suffix→interleave widening allowed).

SESSION 2026-06-13 phase 5b (5655 → 5656, +1): no-xsi for GLOBAL attributes
(§3.2.6.4, wild041). The pass-1 walker already caught a LOCAL attribute with an
explicit targetNamespace="…XMLSchema-instance", but a top-level attribute
inherits its {target namespace} from the schema, so a schema whose own
targetNamespace IS the xsi namespace declares an illegal attribute with no
targetNamespace attr to catch structurally. Added a global-only check in
buildAttributeDecl (buildterms.go): global && a.Name.Namespace == XSINS ⇒
SpecNoXsi. Disjoint from the pass-1 local-explicit check (top-level attrs can't
carry targetNamespace), so no double-report. Unit test TestBuildNoXsiGlobalAttribute.

SESSION 2026-06-13 phase 5 (5651 → 5655, +4): TYPE-ALTERNATIVE substitutability
(e-props-correct.7) + ATTRIBUTE-RESTRICTION inheritability (derivation-ok-restriction).
Two purely-STATIC checks carved out of the CTA/typeAlternatives cluster that
were previously assumed to need an XPath engine — the {test} XPath is never
evaluated:
 (1) e-props-correct.7 (§3.3.6.1): every type named in an element's {type table}
     — each <alternative>'s type AND the default type — must be validly
     substitutable for the element's declared {type definition}, subject to the
     element's {disallowed substitutions}, unless it is xs:error (clause 7.2).
     checkTypeAlternatives (buildschema.go, run in finishComplexTypes tail over
     b.elements) + validlySubstitutable (restrict.go). +2 (s3_12si03 string/decimal
     alt of integer; cta9008err anon complexType alt not derived from docType).
     SOUNDNESS: validlySubstitutable is key-val-sub-type (§3.16.6.5/§3.4.6.5):
     derivationMethods(s,t) found AND methods & (block | t.{prohibited subs}) == 0;
     PLUS cos-st-derived-ok clause 2.2.4 (s derived from a flattened union member
     of t) to avoid false-positives on union-typed elements. CRITICAL gotcha: a
     plain complex type's {base type definition} is left implicit (nil), NOT the
     xs:anyType singleton, so derivationMethods can't walk to anyType — short-circuit
     `t == builtin.AnyType ⇒ true` (every type derives from anyType; also why the
     kitchen-sink "anything" element with anyType decl + tns:base alt is valid).
     The old build_test fixtures used <alternative type="xs:int"> on type="xs:string"
     elements — genuinely spec-invalid; rewritten to type="xs:token" (token <: string)
     so EDC stays the only thing under test.
 (2) derivation-ok-restriction §3.4.6.3 clause 3 via "subsumes" clause 5.3
     (G.{inheritable} = S.{inheritable}): a restriction that redeclares a base
     attribute use of the same name may not flip its inheritability.
     checkAttrRestriction (restrict.go), wired in finishComplexTypes restriction
     branch. +2 (cta9004err true→false; cta9005err absent(false)→true). Compares
     only matched name pairs ⇒ necessary condition ⇒ no false positive.
Unit tests: TestBuilderNegatives +5 (3 e-props-correct.7, 2 inheritability) and a
new TestBuildTypeAlternativeAndAttrRestrictValid (token-restricts-string, anyType
decl, union-member alt, xs:error alt, same-inheritable restriction). STILL gap in
this cluster: cta0043 (Conditional Type Substitutable — a restriction redeclaring
a stamp element with a narrower-but-not-derived alternative type; that's particle
restriction × type-table comparison, derivation-ok-restriction clause 4, harder);
the rest (s3_12si04/05/06, s3_12ii06) are malformed/ill-typed XPath needing an
XPath PARSER — out of scope.

SESSION 2026-06-13 phase 4e (5648 → 5651, +3): MULTI-BASE-WILDCARD particle
restriction solver (all238, all244, wild048 — NOTES remainder item 2). When the
base content model is a flat all/sequence holding ≥2 wildcards, the old code
gave up ("more than one base wildcard"): a single per-base-particle slot can't
say which wildcard a restriction particle maps to, and a restriction wildcard may
straddle several. New checkMultiWildcardRestrict / checkMultiWildcardRun in
restrict.go model each base particle as a baseRegion (a name-membership predicate
+ occurrence range) and reason about the whole packing via a finite set of
REPRESENTATIVE names (collectReps: one generic name per mentioned namespace + one
in a never-mentioned namespace + every explicitly mentioned QName — exhaustive of
all wildcard/element predicate behaviours). Three checks, each backed by a
concrete witness (restriction-valid but base-invalid), so NEVER a false positive:
 (1) OUTSIDE-NAME: a restriction particle that can produce a name no base region
     accepts ⇒ base can't match it. ALWAYS sound (overlap-independent) — this is
     what catches wild048 (R's ##any notQName="a b c..." admits absent "c", which
     base W1=##local notQName="a b c" excludes and W2=notNamespace ##local can't
     take). Runs unconditionally.
 (2) PER-REGION COUNT: minB_i = Σ rmin over restriction particles TRAPPED wholly
     in region i (can't escape it); maxB_i = Σ rmax over particles that can reach
     it. Both achievable simultaneously ⇒ EXACT min/max region count over
     restriction-valid instances. minB_i<base min or maxB_i>base max ⇒ violation.
     Catches all238/all244 (r forces only 3 into base W1's [5,∞] "one/two" region).
 (3) NameAndTypeOK type/nillability for a restriction named element landing in a
     base named region.
CRITICAL SOUNDNESS GATE learned via regressions (wild047/wild049 broke on the
first cut): checks (2)+(3) require the base regions be PAIRWISE DISJOINT. In an
<all> an element particle and a wildcard particle do NOT compete (UPA allows it),
so a base <element name="nm"> can overlap a sibling ##local wildcard — then an
element can be absorbed by EITHER and base validity is a flow problem, not a
per-region count. regionsDisjoint(regions,reps) gates (2)+(3); when false only the
always-sound check (1) runs. wild047/049 are valid exactly because of that
element/wildcard overlap (R's wildcard ⊆ the base wildcard, admits nothing new) —
disjointness gate keeps them valid. Bails entirely on ##defined/##definedSibling
sentinels (wildcardAllowsName ignores them ⇒ would over-accept ⇒ unsound witness).
Routing: checkParticleRestrict counts base wildcards; ≥2 ⇒ checkMultiWildcardRestrict
(handles flat run OR choice branches like the single-WC path), else the existing
slot logic. Unit tests: TestBuilderNegatives +4 (underflow/overflow/escape/notQName
outside), TestBuildParticleRestrictValidModels +2 (disjoint straddle within range;
wild047-shape overlap). Item 1 (processContents ordering) NOT needed after all —
all238's invalidity is also a cardinality underflow, so the solver catches it
without the processContents check. REMAINING multi-WC gap: none of the saxon/ibm
suite cases left for this; wild048 done. (over030 etc. below unchanged.)

SESSION 2026-06-13 phase 4d (5647 → 5648, +1): over030 — FALSE mg-props-correct
cycle fixed. buildGroup conflated group-build recursion with element/type
recursion: building group G → its element ref → that element's type → a group
ref back to G tripped the building-mark while G was still on the stack, even
though recursion THROUGH an element declaration is not a model-group cycle.
Fix: buildGroup now memoizes its SHELL before building the model group (mirrors
the complex-type pattern), so re-entry returns the shell and the recursion is
finite with the structure intact; the building-mark + inline error are gone. A
new dedicated pass checkGroupCycles (buildschema.go, run per schema right after
checkTypeCycles) detects REAL cycles by following only group→nested-group and
group→GroupRef edges (NOT elements), reports mg-props-correct once, and severs
the back-edge (t.Ref=nil; upa.go/EDC already guard nil Ref) so downstream walks
terminate. Unit tests: TestBuildGroupRecursionThroughElement (valid), plus a
nested-model-group cycle negative.

SESSION 2026-06-13 phase 4c (5646 → 5647, +1): over009 — two-level override
chain double-registration FALSE POSITIVE fixed. When top overrides mid and mid
overrides deep, all replacing the same name, both top's and mid's replacement
children registered → spurious sch-props-correct.2 duplicate. Fix in
registerReplacement (loader.go): skip a replacement child when rep.owner.suppressed[k]
— its own document is in turn overridden by a higher composition that replaces
the same name, and the override transformation is applied outermost-last, so the
higher replacement supersedes it. The suppress phase already marks the lower
owner (suppress walks DOWN through targets from the higher override). Unit test:
TestOverride "override of an override registers only the outermost replacement".
over030 (false mg-props-correct cycle) still tolerated — separate bug.

SESSION 2026-06-13 phase 4b (5645 → 5646, +1): open048 — minOccurs/maxOccurs on
the <any> of <openContent>/<defaultOpenContent> is rejected (saxon bug 15618;
that wildcard is an Open Content component, not a Particle). checkOpenContentAnyOccurs
in elemtable.go, wired into both elements' extra. Unit tests in
TestStructuralNegatives (src-ct + src-schema).

SESSION 2026-06-13 phase 4 (5640 → 5645, +5): Element Declarations Consistent
{type table} comparison (cos-element-consistent §3.8.6). checkElementConsistent
in buildschema.go now (a) requires like-named co-occurring declarations to
share a {type table} as well as a top-level type (typeTablesEqual: same length,
same Test strings, same named alternative types — anonymous types compare by
zero name, only ever a false negative), and (b) for a strict/lax wildcard that
binds a like-named GLOBAL element, requires equal {type table}s ONLY (NOT equal
types). CRITICAL false-positive lesson re-learned: a differing TYPE between a
local particle and a wildcard-bound global is a DYNAMIC check (cvc-complex-type.5)
that leaves the SCHEMA valid (wild061/062/066/075 are valid!); only a differing
TYPE TABLE is a static EDC violation. First naive cut applied addElem+full type
check to wildcard globals → 12 regressions (wild061-068/075/076, all006, ibm
edcWildcard) — fixed by splitting into a type-table-only wildcard pass that only
consults a global already colliding with a present local name. +5
(cta9009err/cta9010err two-local-decl type-table mismatch; wild078/079/081
wildcard binds global with mismatched type table). wild069 correctly NO LONGER
caught (it is a restriction/all-wildcard case, not EDC — stays tolerated). Unit
tests: TestBuilderNegatives 3 new + TestBuildElementConsistentValidModels 4 new.

## >>> cos-particle-restrict phase 3 (the hard remainder) <<<
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
1. [MOOT] processContents ordering at the restrict.go wildcard-mapping site.
   Was billed as a prerequisite for item 2 (all238). Item 2 is now done and
   catches all238 via the cardinality underflow, so this fixes no remaining
   case. Helper processContentsAtLeastAsStrict still exists if a future case
   needs it at the single-WC NSSubset site.
2. [DONE phase 4e] MULTI-base-wildcard cardinality (all238, all244, wild048).
   Solver in restrict.go (checkMultiWildcardRestrict). See phase-4e block above.
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
- CTA / assertions XPath: the STATIC-error subset (malformed XPath, bad cast
  QName, cast-to-complex) is now caught by the `xpath` package on <alternative>
  (phase 6, ibm typeAlternatives s3_12si04/05/06/ii06 DONE). What remains needs
  XPath EVALUATION (a runtime engine over an instance), which is out of scope.
- 2 override false positives (over009 double-override dup; over030 false
  mg-props-correct cycle — override-internals bugs) + over014. (iri-001 custom
  DTD entities &URI; — FIXED in phase 9, see NEXT SESSION block.)
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
buildAttrUses/extendAttrUses), derived vs base mixed consistency
(cos-ct-extends), NOTATION enum values resolving to declared notations.
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
