#!/usr/bin/env bash
# gate.sh — the single quality gate for goxsd5.
#
# One authoritative command for the checks every change must pass: formatting,
# vet, lint, unit tests, the CLI smoke test, and the W3C conformance ratchets.
# Humans run it before a PR; the auto-maintenance Implementor runs it before
# opening a PR and the Evaluator runs it to judge one (see MAINTENANCE.md).
#
# It runs ALL stages and reports every failure (it does not stop at the first),
# then exits non-zero if any stage failed — so one run surfaces everything.
#
# Usage:
#   tools/gate.sh            # full gate incl. conformance (inits the submodule)
#   tools/gate.sh --quick    # skip the conformance suites (fast inner loop)
#
# Run from anywhere; it locates the repo root itself.
set -uo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root" || { echo "cannot cd to repo root"; exit 1; }

quick=0
[ "${1:-}" = "--quick" ] && quick=1

failures=()
run() { # run <label> <cmd...>
  local label="$1"; shift
  echo "=== $label ==="
  if "$@"; then
    echo "  ok: $label"
  else
    echo "  FAIL: $label"
    failures+=("$label")
  fi
}

# --- formatting: gofmt -l must print nothing ---
echo "=== gofmt ==="
unformatted="$(gofmt -l . | grep -v '^testdata/xsdtests/' || true)"
if [ -n "$unformatted" ]; then
  echo "  FAIL: gofmt — unformatted files:"; echo "$unformatted" | sed 's/^/    /'
  failures+=("gofmt")
else
  echo "  ok: gofmt"
fi

run "go vet" go vet ./...

# --- lint: requires golangci-lint built with Go >= 1.26 ---
echo "=== golangci-lint ==="
if ! command -v golangci-lint >/dev/null 2>&1; then
  echo "  FAIL: golangci-lint not installed —"
  echo "    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
  failures+=("golangci-lint(missing)")
else
  lint_out="$(golangci-lint run 2>&1)"; lint_rc=$?
  if echo "$lint_out" | grep -q "lower than the targeted Go version"; then
    echo "  FAIL: golangci-lint is built with an older Go than this module —"
    echo "    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    failures+=("golangci-lint(stale)")
  elif [ $lint_rc -ne 0 ]; then
    echo "$lint_out" | sed 's/^/    /'
    echo "  FAIL: golangci-lint"; failures+=("golangci-lint")
  else
    echo "  ok: golangci-lint"
  fi
fi

run "go test ./..." go test ./...
run "smoke" "$repo_root/.claude/skills/run-goxsd5/smoke.sh"

# --- conformance ratchets (the behavioural ground truth) ---
if [ "$quick" -eq 1 ]; then
  echo "=== conformance: skipped (--quick) ==="
else
  if [ ! -f testdata/xsdtests/suite.xml ]; then
    echo "=== conformance: initializing submodule ==="
    git submodule update --init testdata/xsdtests || failures+=("submodule-init")
  fi
  if [ -f testdata/xsdtests/suite.xml ]; then
    run "conformance: schema (baseline 5697)"   go test ./parser -run TestConformanceSuite
    run "conformance: instance (baseline 21497)" go test ./parser -run TestInstanceConformance
  else
    echo "  FAIL: conformance — submodule unavailable, cannot run the ratchets"
    failures+=("conformance(no-submodule)")
  fi
fi

echo
if [ ${#failures[@]} -eq 0 ]; then
  echo "GATE PASSED"
  exit 0
fi
echo "GATE FAILED: ${failures[*]}"
exit 1
