#!/usr/bin/env bash
# smoke.sh — build and drive the goxsd5 CLI end-to-end.
#
# goxsd5 has no GUI; it's an XSD 1.1 parser/validator exposed as a Go library
# and one CLI (cmd/goxsd5). This script is the agent's handle on the running
# app: it builds the binary, then drives every CLI surface — schema summary,
# schema-error reporting, and instance validation (a passing and a failing
# document) — asserting the exit codes the CLI documents (0 ok / 1 invalid /
# 2 usage). It uses a self-contained schema+instance written to a temp dir, so
# it does NOT need the W3C submodule.
#
# Usage:
#   .claude/skills/run-goxsd5/smoke.sh           # build + drive the CLI
#   .claude/skills/run-goxsd5/smoke.sh --suite   # also run the conformance suites (needs submodule)
#
# Run from the repo root (the goxsd5 module directory).
set -uo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
cd "$repo_root" || { echo "cannot cd to repo root"; exit 1; }

bin="$(mktemp -d)/goxsd5"
work="$(mktemp -d)"
fail=0
note() { printf '\n=== %s ===\n' "$1"; }
check() { # check <label> <actual-exit> <expected-exit>
  if [ "$2" = "$3" ]; then echo "PASS: $1 (exit $2)";
  else echo "FAIL: $1 (exit $2, want $3)"; fail=1; fi
}

note "build CLI"
go build -o "$bin" ./cmd/goxsd5 || { echo "FAIL: build"; exit 1; }
echo "built $bin"

note "usage (no args -> exit 2)"
"$bin"; check "usage" "$?" 2

# Self-contained schema + instances.
cat > "$work/demo.xsd" <<'EOF'
<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="age" type="xs:positiveInteger"/>
</xs:schema>
EOF
printf '<?xml version="1.0"?>\n<age>42</age>\n' > "$work/good.xml"
printf '<?xml version="1.0"?>\n<age>-5</age>\n'  > "$work/bad.xml"
# A schema with a static error (positiveInteger cannot have a negative minInclusive facet base mismatch).
cat > "$work/broken.xsd" <<'EOF'
<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="x" type="xs:nosuchtype"/>
</xs:schema>
EOF

note "valid schema summary (-> exit 0)"
"$bin" "$work/demo.xsd"; check "schema summary" "$?" 0

note "schema with error (-> exit 1)"
"$bin" -q "$work/broken.xsd"; check "schema error" "$?" 1

note "instance validation: valid doc (-> exit 0)"
"$bin" -q -validate "$work/good.xml" "$work/demo.xsd"; check "valid instance" "$?" 0

note "instance validation: invalid doc (-> exit 1, cvc-* error)"
out=$("$bin" -q -validate "$work/bad.xml" "$work/demo.xsd" 2>&1); rc=$?
echo "$out"
check "invalid instance" "$rc" 1
echo "$out" | grep -q 'cvc-' && echo "PASS: cvc-* error reported" || { echo "FAIL: no cvc-* error"; fail=1; }

# precisionDecimal (XSD 1.1): a maxScale-constrained restriction exercising the
# registration + scale-facet wiring (issue #25). price allows at most 2
# fractional digits, so "3.00" (scale 2) is valid and "3.000" (scale 3) is not.
cat > "$work/price.xsd" <<'EOF'
<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="price">
    <xs:simpleType>
      <xs:restriction base="xs:precisionDecimal">
        <xs:maxScale value="2"/>
      </xs:restriction>
    </xs:simpleType>
  </xs:element>
</xs:schema>
EOF
printf '<?xml version="1.0"?>\n<price>3.00</price>\n'  > "$work/price-good.xml"
printf '<?xml version="1.0"?>\n<price>3.000</price>\n' > "$work/price-bad.xml"

note "precisionDecimal schema summary (-> exit 0)"
"$bin" "$work/price.xsd"; check "precisionDecimal schema" "$?" 0

note "precisionDecimal instance: within maxScale (-> exit 0)"
"$bin" -q -validate "$work/price-good.xml" "$work/price.xsd"; check "precisionDecimal valid instance" "$?" 0

note "precisionDecimal instance: exceeds maxScale (-> exit 1, cvc-maxScale-valid)"
out=$("$bin" -q -validate "$work/price-bad.xml" "$work/price.xsd" 2>&1); rc=$?
echo "$out"
check "precisionDecimal invalid instance" "$rc" 1
echo "$out" | grep -q 'maxScale' && echo "PASS: maxScale error reported" || { echo "FAIL: no maxScale error"; fail=1; }

if [ "${1:-}" = "--suite" ]; then
  note "W3C conformance suites (gated against committed expectations)"
  if [ -e testdata/xsdtests/suite.xml ]; then
    go test ./parser -run 'TestConformanceSuite|TestInstanceConformance' 2>&1 | tail -5
    check "conformance suites" "${PIPESTATUS[0]}" 0
  else
    echo "SKIP: submodule not checked out — run: git submodule update --init testdata/xsdtests"
  fi
fi

note "result"
[ "$fail" = 0 ] && { echo "ALL SMOKE CHECKS PASSED"; exit 0; } || { echo "SOME CHECKS FAILED"; exit 1; }
