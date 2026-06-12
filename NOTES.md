# Working notes (restart context)

Self-notes for resuming work. Keep updated at every checkpoint commit.

## Goal
Implement PLAN.md (XSD 1.1 parser, packages `xsd`, `builtin`, `parser`,
`parser/xmltree`), then run the W3C suite via the M9 ratchet harness and
baseline `testdata/xsd11-expectations.txt`.

## Status
- [x] M0 foundations (xsd: Pos, QName, SpecRef, errors)
- [x] M1 parser/xmltree
- [ ] M2 xsd model skeleton
- [ ] M3 value space + facet pipeline + Appendix-F regex
- [ ] M4 builtin package
- [ ] M5 parser pass 1
- [ ] M6 parser pass 2
- [ ] M7 imports/includes/override/redefine + resolver
- [ ] M9 W3C harness + expectations baseline
- [ ] M8 mutation API, CONFORMANCE.md fill-in, cmd/goxsd5

## Key facts discovered
- W3C suite is already fetched at `testdata/xsdtests/` (gitignored, pinned in
  `testdata/fetch-xsdtests.sh`).
- Suite layout: `suite.xml` → `<ts:testSetRef xlink:href>` → testSet files
  (ns `http://www.w3.org/XML/2004/xml-schema-test-suite/`) → `testGroup` →
  `schemaTest` → `schemaDocument xlink:href` + `expected validity=valid|invalid`
  (sometimes several `expected` with `version="1.0"`/`"1.1"`; no version attr =
  applies to all). testSet/testGroup/schemaTest may also carry `version` attrs.
  `validity="indeterminate"` exists (16 cases) — treat as skip.
- Harness: schema tests only (we are a schema parser, not instance validator).
- Spec refs: cite stable constraint IDs + anchors into docs/clean/*.md.

## Conventions
- Every enforced constraint: `// spec: <id> — XSD 1.1 Part N §x (anchor)` at the
  enforcement site + SpecRef constant in xsd/specref.go + row in CONFORMANCE.md.
- Errors: `uri:line:col: [id] message`.
- Commit at each milestone boundary; update this file in the same commit.
