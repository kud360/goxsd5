# goxsd5 — code conventions

The full style guide. `CLAUDE.md` carries the short always-on subset and points
here; read this before writing code. Mechanical rules (formatting, imports,
dead code) are enforced by `.golangci.yml`, not restated here — this file is for
the judgment a linter can't make.

> Doc map: **root** holds dev/project docs (this file, `NOTES.md`,
> `CONFORMANCE.md`, `MAINTENANCE.md`). **`docs/`** is the spec corpus
> (W3C XSD 1.1 parts 1–2, XPath 2.0) — reference only, not dev docs.

## Go style

- **Happy path on the left.** Handle errors and edge cases with guard-clause
  early returns; keep the main logic at the lowest indentation.
- **Avoid `else`.** An early `return`/`continue`/`break` in the guard removes the
  need for it. (`revive`'s `early-return`/`indent-error-flow` nudge this.)
- **No concurrency.** No goroutines, channels, or `sync` for control flow. The
  parser/validator is a pure transform — keep it single-threaded and simple.
- **Determinism is a contract.** Observable output must not depend on iteration
  order. Maps are fine as internal lookups, but **never `range` a map to produce
  ordered output** — collect keys into a slice and `sort` them first. No reliance
  on map order, goroutine scheduling, or wall-clock time in behaviour.
- **Mind the exported surface.** Export only what callers genuinely need;
  default to unexported. A smaller API is easier to keep conformant and to evolve.
  Exported symbols get doc comments (`revive` `exported`).
- **Match the surrounding code** — naming, comment density, idiom. This is a
  mature codebase with settled patterns; follow them rather than introducing new
  ones.

## Testing

- **Test behaviour, not implementation.** Assert on observable results
  (validation verdicts, error ids, parsed shapes), not on private fields or call
  sequences. Tests should survive a refactor that preserves behaviour.
- **Coverage is a flashlight, not a target.** Run `go test -cover` (or
  `-coverprofile`) to *find untested behaviour*. Do **not** chase 100% — forcing
  coverage of every branch pushes you into testing implementation, which is the
  thing we're avoiding.
- The **W3C conformance suites are the behavioural ground truth.** Prefer adding
  or un-skipping a real spec case over a synthetic unit test when both fit.

## Project architecture invariants

These are load-bearing — a change that breaks one is wrong even if tests pass.

- **Dependency direction / pure leaves.** `xsd` is a leaf package. `xsdtemporal`
  and `xsdregex` are *pure* leaves (stdlib-only imports). Never add an upward or
  cyclic import; never make a leaf depend inward.
- **Capability interfaces over type switches.** The facet engine dispatches on
  small interfaces (`Lengthed`, `DigitCounted`, `TimezoneAware`), not on concrete
  types. Add a capability, don't add a `switch` on a new type.
- **Mutation via free functions.** Schema mutation lives in `xsdedit` as free
  functions (`xsdedit.RestrictWith(t, f)`, `xsdedit.AddElement(...)`).
  `EffectiveFacets()` stays a method. Don't grow mutating methods on core types.
- **Error-id discipline.** Schema errors carry spec ids (`src-resolve`,
  `ag-props-correct`); instance errors carry `cvc-*` ids. Keep the id accurate to
  the spec clause — they're asserted by the suites.

## The gate (run before every PR)

One command runs everything — `tools/gate.sh` (`--quick` skips the conformance
suites for a fast inner loop). It must end `GATE PASSED`. It runs, in order:

```text
gofmt -l .                                       # formatting (must be empty)
go vet ./...                                      # vet
golangci-lint run                                 # lint
go test ./...                                      # unit tests
.claude/skills/run-goxsd5/smoke.sh                # CLI smoke
go test ./parser -run TestConformanceSuite        # schema ratchet  (baseline 5672)
go test ./parser -run TestInstanceConformance     # instance ratchet (baseline 21429)
```

> **Toolchain note:** `golangci-lint`/`staticcheck` must be **built with Go ≥1.26**
> or they refuse to lint this module ("Go language version … lower than the
> targeted Go version"). Reinstall with the current toolchain:
> `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`.

Never lower a conformance baseline. Genuine coverage gains require regenerating
`testdata/*-expectations.txt` via the `-update-*expectations` flags, committed in
the same PR. See `.claude/skills/run-goxsd5/SKILL.md` for details.
