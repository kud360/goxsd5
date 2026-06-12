# testdata

## xsdtests — W3C XSD 1.1 Test Suite

`testdata/xsdtests/` is the official **W3C XML Schema (XSD) 1.1 test suite**,
<https://github.com/w3c/xsdtests>. It is the conformance suite referenced by
Milestone 9 in [`../PLAN.md`](../PLAN.md): positive and negative schema cases that
`parser` is validated against, with negative cases asserted to fail with the
expected `SpecRef.ID`.

**It is not committed.** The suite is large (~230 MB) and gitignored. Fetch it on
demand (locally and in CI) at the pinned revision:

```sh
testdata/fetch-xsdtests.sh
```

Bump the pinned `REV` in that script deliberately; after a bump, re-baseline the
expectations file (see PLAN.md M9) so the revision change doesn't silently alter
the regression set.

Key entry points within the suite:

- `XSD1_1TestCategories.xml` / `.xhtml` — the XSD 1.1 test category index.
- `suite.xml` / `extra-suite.xml` — top-level test set manifests.
- `msMeta/`, `sunMeta/`, `boeingMeta/`, `nistMeta/` — per-vendor metadata
  describing each test case (schema document(s), instance, and expected
  validity outcome).

It carries its own `.git` directory (full upstream clone), so it is an embedded
repository, not part of this module's history.
