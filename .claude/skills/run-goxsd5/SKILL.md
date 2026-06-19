---
name: run-goxsd5
description: Build, run, and smoke-test goxsd5 — the XSD 1.1 schema parser/validator CLI and Go library. Use to run the goxsd5 command-line tool, parse/summarize an .xsd schema, validate an XML instance document against a schema, drive the validator end-to-end, or run the W3C conformance test suites.
---

# Run goxsd5

goxsd5 is an XSD 1.1 schema parser and instance validator. It ships as a Go
**library** (packages `xsd`, `parser`, `xsdvalidate`, `xpath`, `xsdwalk`,
`xsdtemporal`, …) plus one **CLI** at `cmd/goxsd5`. There is no GUI.

The agent handle is **`.claude/skills/run-goxsd5/smoke.sh`** — it builds the
binary and drives every CLI surface (schema summary, schema-error reporting,
and instance validation of a passing and a failing document), asserting the
documented exit codes (`0` ok / `1` invalid / `2` usage). It uses a
self-contained schema+instance, so it needs **no** submodule.

All paths below are relative to the repo root (the module dir, where `go.mod`
lives). Requires **Go 1.26+** (`go version`). No `apt-get` packages needed —
pure Go, one `golang.org/x/net` dep already vendored in `go.sum`.

## Run (agent path) — the driver

```bash
.claude/skills/run-goxsd5/smoke.sh           # build + drive the CLI (no submodule needed)
.claude/skills/run-goxsd5/smoke.sh --suite   # also run W3C conformance suites (needs submodule)
```

Last line is `ALL SMOKE CHECKS PASSED` on success (exit 0). The driver writes
its demo schema/instances to a fresh temp dir and builds the binary to another
temp dir — it leaves the repo tree untouched.

## Run the CLI directly

```bash
go build -o /tmp/goxsd5 ./cmd/goxsd5

# Parse a schema and print a per-namespace component summary (exit 1 if the schema has errors):
/tmp/goxsd5 path/to/schema.xsd

# Quiet mode — only print errors:
/tmp/goxsd5 -q path/to/schema.xsd

# Validate an XML instance against a schema (exit 0 valid / 1 invalid; errors are cvc-* ids):
/tmp/goxsd5 -q -validate doc.xml schema.xsd
```

Errors print as `uri:line:col: [constraint-id] message` — schema errors use
spec ids like `[src-resolve]`, instance errors use `[cvc-minInclusive-valid]`.

## Direct invocation (library / internal changes)

Most changes here touch internal packages, not the CLI. Run the package tests:

```bash
go test ./...                              # full suite (~5s; parser is the slow one)
go test ./xsd/... ./xpath/... ./xsdtemporal/...   # fast, focused on one area
```

## W3C conformance suites

The real value of this project is conformance against the W3C `xsdtests`
suite (a git submodule). Two gated tests compare the current run against
committed expectation files and fail on any regression **or** unexpected pass:

```bash
git submodule update --init testdata/xsdtests        # once, if testdata/xsdtests/suite.xml is missing

go test ./parser -run TestConformanceSuite           # ~5698 schema-validity cases
go test ./parser -run TestInstanceConformance        # ~21431 instance-validity cases
```

To record improved coverage after a fix, regenerate the baselines:

```bash
go test ./parser -run TestConformanceSuite   -update-expectations
go test ./parser -run TestInstanceConformance -update-instance-expectations
```

To triage which still-failing cases aren't yet recorded as pass-or-skip:

```bash
GOXSD5_CONFORMANCE_GAPS=1 go test ./parser -run TestConformanceSuite -v 2>&1 | grep "unrecorded gaps" -A30
```

## Gotchas

- **The conformance tests fail on *unexpected passes*, not just regressions.**
  If a fix makes more cases pass, the test errors until you re-run with the
  matching `-update-*expectations` flag to record the new baseline. This is
  intentional — it ratchets coverage forward.
- **`go.mod` requires Go 1.26.1.** An older toolchain won't build it.
- **The submodule is large (~15k files).** `TestConformanceSuite` /
  `TestInstanceConformance` **skip cleanly** (not fail) when
  `testdata/xsdtests/suite.xml` is absent — so a clean checkout still passes
  `go test ./...`. Init the submodule only when you need the suites.
- **`TestScanW3CSuiteValidSchemas` is opt-in** via `GOXSD5_SCAN=1` (it walks
  every schema in the suite). It's a fuzz-style "doesn't crash" scan, not gated.
- The CLI's plain (non-`-q`) mode always prints the component summary to
  stdout *before* the validity verdict; use `-q` when you only want errors.

## Troubleshooting

- `W3C suite not checked out` skip message → `git submodule update --init testdata/xsdtests`.
- `N unexpected pass(es) — coverage improved` test failure → re-run that test
  with its `-update-*expectations` flag, then commit the regenerated
  `testdata/*-expectations.txt`.
- Build error mentioning a Go version → check `go version` is ≥ 1.26.1.
