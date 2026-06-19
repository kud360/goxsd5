package parser

// The W3C XSD test-suite conformance ratchet (PLAN.md M9). For every schema
// test applicable to XSD 1.1, the full pipeline (discovery + both passes) is
// run and its verdict (errors vs. no errors) is compared to the suite's
// declared validity. The result is gated against a committed expectations
// file, testdata/xsd11-expectations.txt:
//
//   - listed "pass", still spec-correct       → ok
//   - listed "pass", now wrong                 → FAIL (regression)
//   - not listed, now spec-correct             → FAIL (unexpected pass; re-run
//                                                 with -update-expectations)
//   - not listed, still wrong                  → ok (a known, unrecorded gap)
//   - listed "skip:<reason>"                   → not gated (deferred feature)
//
// So coverage only ratchets up: a newly-passing case must be recorded, and a
// recorded pass may never silently regress. Run
//
//	go test ./parser -run TestConformanceSuite -update-expectations
//
// to rewrite the pass set from the current run; human-curated skip: lines are
// preserved. The suite is a git submodule at testdata/xsdtests (pinned to a
// specific upstream revision); the harness skips if it is not checked out.

import (
	"bufio"
	"encoding/xml"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/kud360/goxsd5/xsd"
)

var updateExpectations = flag.Bool("update-expectations", false,
	"rewrite testdata/xsd11-expectations.txt from the current conformance run")

const (
	suiteRoot        = "../testdata/xsdtests"
	expectationsPath = "../testdata/xsd11-expectations.txt"
)

// suiteCase is one schema test applicable to XSD 1.1.
type suiteCase struct {
	id        string // "<testSet-relpath>#<group>", unique across the suite
	setFile   string // path to the .testSet that declares it
	docs      []scanXlinkRef
	wantValid bool // suite's declared 1.1 validity (valid vs invalid)
}

// expectation is one line of the expectations file.
type expectation struct {
	pass bool   // "pass"
	skip string // non-empty reason for "skip:<reason>" (pass is then false)
}

func TestConformanceSuite(t *testing.T) {
	if _, err := os.Stat(suiteRoot); err != nil {
		t.Skipf("W3C suite not checked out (%v); run: git submodule update --init testdata/xsdtests", err)
	}
	cases := collectSuiteCases(t)
	if len(cases) == 0 {
		t.Fatal("no XSD 1.1 schema cases found under " + suiteRoot)
	}
	want := readExpectations(t)

	// correct[id] records whether we produced the spec-correct verdict for
	// every case we could actually load. Cases whose root document failed to
	// load at the XML level are excluded (neither gated nor recorded).
	correct := map[string]bool{}
	docByID := map[string]suiteCase{}
	for _, c := range cases {
		loaded, ok := runSuiteCase(c)
		if !loaded {
			continue
		}
		correct[c.id] = ok
		docByID[c.id] = c
	}

	// Diagnostic for growing the baseline: list the cases we get wrong that
	// are neither recorded as passing nor curated as skip, so deferred
	// features can be triaged into skip: lines.
	if os.Getenv("GOXSD5_CONFORMANCE_GAPS") != "" {
		var gaps []string
		for id, ok := range correct {
			if ok {
				continue
			}
			if _, listed := want[id]; listed {
				continue
			}
			verb := "accepts (want invalid)"
			if docByID[id].wantValid {
				verb = "rejects (want valid)"
			}
			gaps = append(gaps, id+" — "+verb)
		}
		sort.Strings(gaps)
		t.Logf("%d unrecorded gaps:\n%s", len(gaps), strings.Join(gaps, "\n"))
	}

	if *updateExpectations {
		writeExpectations(t, want, correct)
		return
	}

	var regressions, unexpected []string
	for id, ok := range correct {
		exp, listed := want[id]
		switch {
		case listed && exp.skip != "":
			// deferred feature: not gated.
		case listed && exp.pass && !ok:
			regressions = append(regressions, id)
		case !listed && ok:
			unexpected = append(unexpected, id)
		}
	}
	// A recorded pass that vanished entirely (case removed from the suite, or
	// its root no longer loads) is also a regression.
	for id, exp := range want {
		if exp.pass {
			if _, ran := correct[id]; !ran {
				regressions = append(regressions, id+" (no longer evaluated)")
			}
		}
	}

	sort.Strings(regressions)
	sort.Strings(unexpected)
	passes, skips := 0, 0
	for _, e := range want {
		if e.pass {
			passes++
		} else if e.skip != "" {
			skips++
		}
	}
	t.Logf("conformance: %d cases, %d recorded passes, %d skips", len(cases), passes, skips)

	if len(unexpected) > 0 {
		t.Errorf("%d unexpected pass(es) — coverage improved; re-run with -update-expectations to record:\n  %s",
			len(unexpected), strings.Join(capList(unexpected, 30), "\n  "))
	}
	if len(regressions) > 0 {
		t.Errorf("%d regression(s) — these were recorded as passing and now fail:\n  %s",
			len(regressions), strings.Join(capList(regressions, 30), "\n  "))
	}
}

func capList(s []string, n int) []string {
	if len(s) <= n {
		return s
	}
	return append(s[:n:n], fmt.Sprintf("... and %d more", len(s)-n))
}

// collectSuiteCases walks every .testSet and returns the schema tests that
// apply to XSD 1.1 with a determinate (valid/invalid) expectation.
func collectSuiteCases(t *testing.T) []suiteCase {
	var setFiles []string
	if err := filepath.WalkDir(suiteRoot, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".testSet") {
			setFiles = append(setFiles, p)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	sort.Strings(setFiles)

	var cases []suiteCase
	for _, sf := range setFiles {
		data, err := os.ReadFile(sf)
		if err != nil {
			t.Fatalf("read %s: %v", sf, err)
		}
		var ts scanTestSet
		if err := xml.Unmarshal(data, &ts); err != nil {
			t.Fatalf("unmarshal %s: %v", sf, err)
		}
		rel, _ := filepath.Rel(suiteRoot, sf)
		rel = filepath.ToSlash(rel)
		for _, g := range ts.Groups {
			if g.Schema == nil || !scanVersionApplies(g.Version) || !scanVersionApplies(g.Schema.Version) {
				continue
			}
			validity := scanExpectedValidity(g.Schema.Expected)
			if validity != "valid" && validity != "invalid" {
				continue // indeterminate / not-applicable-to-1.1
			}
			cases = append(cases, suiteCase{
				id:        rel + "#" + g.Name,
				setFile:   sf,
				docs:      g.Schema.Docs,
				wantValid: validity == "valid",
			})
		}
	}
	return cases
}

// runSuiteCase loads the case's schema document(s) as roots of one loader and
// builds the schema set. loaded is false when a root could not be read or
// parsed at the XML level (a suite/infra issue, not a schema verdict); in that
// case the case is excluded from gating. Otherwise ok reports whether our
// verdict (errors vs. none) matches the suite's declared validity.
func runSuiteCase(c suiteCase) (loaded, ok bool) {
	errs := &xsd.ErrorList{}
	l := newLoader(FileResolver{}, errs)
	for _, d := range c.docs {
		p := filepath.Join(filepath.Dir(c.setFile), filepath.FromSlash(d.Href))
		if err := l.loadRoot(p); err != nil {
			return false, false
		}
	}
	_, _ = finish(l, errs)
	gotValid := errs.Empty()
	return true, gotValid == c.wantValid
}

func readExpectations(t *testing.T) map[string]expectation {
	want := map[string]expectation{}
	f, err := os.Open(expectationsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return want // first run, before -update-expectations
		}
		t.Fatalf("open expectations: %v", err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		id, outcome, found := strings.Cut(line, " ")
		if !found {
			t.Fatalf("malformed expectations line: %q", line)
		}
		outcome = strings.TrimSpace(outcome)
		switch {
		case outcome == "pass":
			want[id] = expectation{pass: true}
		case strings.HasPrefix(outcome, "skip:"):
			want[id] = expectation{skip: strings.TrimSpace(outcome[len("skip:"):])}
		default:
			t.Fatalf("unknown outcome %q in expectations for %q", outcome, id)
		}
	}
	return want
}

// writeExpectations rewrites the file: every human-curated skip: line is kept,
// and the pass set becomes every currently spec-correct case not marked skip.
func writeExpectations(t *testing.T, want map[string]expectation, correct map[string]bool) {
	lines := map[string]string{}
	for id, exp := range want {
		if exp.skip != "" {
			lines[id] = "skip:" + exp.skip
		}
	}
	for id, ok := range correct {
		if _, skipped := lines[id]; skipped {
			continue
		}
		if ok {
			lines[id] = "pass"
		}
	}
	ids := make([]string, 0, len(lines))
	for id := range lines {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	var b strings.Builder
	b.WriteString("# XSD 1.1 conformance expectations (W3C test suite). Generated by\n")
	b.WriteString("# `go test ./parser -run TestConformanceSuite -update-expectations`.\n")
	b.WriteString("# Each line is `<testSet-relpath>#<group> pass|skip:<reason>`.\n")
	b.WriteString("# `pass` lines ratchet up automatically; `skip:` lines are curated by hand\n")
	b.WriteString("# for deferred features and are preserved across updates. See M9 in PLAN.md.\n")
	passes, skips := 0, 0
	for _, id := range ids {
		fmt.Fprintf(&b, "%s %s\n", id, lines[id])
		if lines[id] == "pass" {
			passes++
		} else {
			skips++
		}
	}
	if err := os.WriteFile(expectationsPath, []byte(b.String()), 0o644); err != nil {
		t.Fatalf("write expectations: %v", err)
	}
	t.Logf("wrote %s: %d pass, %d skip", expectationsPath, passes, skips)
}
