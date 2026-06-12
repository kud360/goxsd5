#!/usr/bin/env bash
# Fetch the W3C XSD 1.1 test suite at a pinned revision into testdata/xsdtests/.
# The suite is large (~230 MB) and gitignored, so it is fetched on demand rather
# than committed. Re-run this to (re)populate it locally or in CI.
set -euo pipefail

REPO="https://github.com/w3c/xsdtests.git"
# Pinned revision — bump deliberately, then re-baseline expectations (see PLAN.md M9).
REV="7bc3365c652a322f3d762021b3879eb92dae7e30"

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/xsdtests"

if [ -d "$DIR/.git" ] && [ "$(git -C "$DIR" rev-parse HEAD 2>/dev/null)" = "$REV" ]; then
  echo "xsdtests already at $REV"
  exit 0
fi

rm -rf "$DIR"
mkdir -p "$DIR"
git -C "$DIR" init -q
git -C "$DIR" remote add origin "$REPO"
git -C "$DIR" fetch -q --depth 1 origin "$REV"
git -C "$DIR" checkout -q FETCH_HEAD
echo "xsdtests fetched at $REV"
