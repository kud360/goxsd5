# testdata

## xsdtests — W3C XSD 1.1 Test Suite

`testdata/xsdtests/` is the official **W3C XML Schema (XSD) 1.1 test suite**,
<https://github.com/w3c/xsdtests>. It is the conformance suite that `parser` is
validated against: positive and negative schema cases, with negative cases
asserted to fail with the expected `SpecRef.ID`.

**It is a git submodule, not committed content.** The suite is large (~230 MB),
so only a gitlink pinning a specific upstream revision lives in this repo's
history. Check it out (locally and in CI) with:

```sh
git submodule update --init testdata/xsdtests
```

A fresh `git clone --recurse-submodules` of this repo populates it automatically.

To bump the pinned revision, advance the submodule and re-baseline the
expectations file in the **same commit** so the revision change
doesn't silently alter the regression set:

```sh
git -C testdata/xsdtests fetch origin && git -C testdata/xsdtests checkout <new-rev>
git add testdata/xsdtests
go test ./parser -run TestConformanceSuite -update-expectations  # re-baseline
git add testdata/xsd11-expectations.txt
```

Key entry points within the suite:

- `XSD1_1TestCategories.xml` / `.xhtml` — the XSD 1.1 test category index.
- `suite.xml` / `extra-suite.xml` — top-level test set manifests.
- `msMeta/`, `sunMeta/`, `boeingMeta/`, `nistMeta/` — per-vendor metadata
  describing each test case (schema document(s), instance, and expected
  validity outcome).
