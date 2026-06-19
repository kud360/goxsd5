# goxsd5

XSD 1.1 schema parser + instance validator: a Go **library** (`xsd`, `parser`,
`xsdvalidate`, `xpath`, `xsdwalk`, `xsdtemporal`, `xsdregex`, `xsdedit`,
`builtin/xsdtype`, …) plus one CLI at `cmd/goxsd5`. Requires Go 1.26+.

- **Working notes & design history:** `NOTES.md` (read it on resume).
- **How to build/run/test:** `.claude/skills/run-goxsd5/SKILL.md`.
- **Auto-maintenance loop** (Planner / Ship / Implementor / Evaluator): `MAINTENANCE.md`.
- **Full code style & conventions:** `CONVENTIONS.md` — **read it before writing code.**
- **Spec corpus** (XSD 1.1, XPath 2.0): `docs/` — reference only.

## Non-negotiables (always apply)

- **Conformance ratchet.** Two gated suites pin behaviour and fail on regression
  *and* on unrecorded improvement: `go test ./parser -run TestConformanceSuite`
  (baseline **5672**) and `-run TestInstanceConformance` (baseline **21429**).
  Never lower a baseline; regenerate `testdata/*-expectations.txt` for genuine gains.
- **Dependency direction.** `xsd` is a leaf; `xsdtemporal`/`xsdregex` are *pure*
  leaves (stdlib-only). Never add an upward import.
- **Determinism.** No goroutines/concurrency. Never range a map to produce ordered
  output — collect keys and sort.
- **The gate must pass** before any PR: `tools/gate.sh` (fmt, vet, lint, tests,
  smoke, conformance ratchets) ends `GATE PASSED`.

## Go style (see docs/conventions.md for the full rationale)

- Happy path on the left; guard-clause early returns; **avoid `else`**.
- Minimal exported surface — export only what callers need.
- Test **behaviour, not implementation**. Use coverage to find untested behaviour,
  not to chase 100%.
- Match the surrounding code's naming, comment density, and idiom.
