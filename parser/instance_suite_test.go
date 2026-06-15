package parser

// The W3C XSD test-suite *instance* conformance ratchet (PLAN-validate.md). For
// every <instanceTest> whose enclosing <testGroup> has a schema we build cleanly
// and whose schema the suite declares valid under XSD 1.1, the instance document
// is assessed by xsdvalidate and the verdict (valid vs invalid) is compared to
// the suite's <expected validity>. The result is gated against a committed
// expectations file, testdata/instance-expectations.txt, with the exact same
// ratchet semantics as the schema conformance suite:
//
//   - listed "pass", still correct   → ok
//   - listed "pass", now wrong       → FAIL (regression)
//   - not listed, now correct        → FAIL (re-run with -update-instance-expectations)
//   - not listed, still wrong        → ok (known gap)
//   - listed "skip:<reason>"         → not gated
//
// Run
//
//	go test ./parser -run TestInstanceConformance -update-instance-expectations
//
// to rewrite the pass set; curated skip: lines are preserved.

import (
	"bufio"
	"encoding/xml"
	"flag"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/kud360/goxsd5/xsd"
	"github.com/kud360/goxsd5/xsdvalidate"
	"github.com/kud360/goxsd5/xsdvalidate/xmlsrc"
)

var updateInstanceExpectations = flag.Bool("update-instance-expectations", false,
	"rewrite testdata/instance-expectations.txt from the current instance run")

const instanceExpectationsPath = "../testdata/instance-expectations.txt"

// instScanGroup extends the schema scan structs with the group's instance tests.
type instScanSet struct {
	Groups []instScanGroup `xml:"testGroup"`
}
type instScanGroup struct {
	Name      string          `xml:"name,attr"`
	Version   string          `xml:"version,attr"`
	Schema    *scanSchemaTest `xml:"schemaTest"`
	Instances []instScanTest  `xml:"instanceTest"`
}
type instScanTest struct {
	Name     string         `xml:"name,attr"`
	Version  string         `xml:"version,attr"`
	Doc      scanXlinkRef   `xml:"instanceDocument"`
	Expected []scanExpected `xml:"expected"`
}

// instanceCase is one <instanceTest> applicable to XSD 1.1 under a buildable,
// suite-valid schema.
type instanceCase struct {
	id        string // "<testSet-relpath>#<group>#<instance>"
	docPath   string // resolved path to the instance document
	schemas   []*xsd.Schema
	wantValid bool
}

func TestInstanceConformance(t *testing.T) {
	if _, err := os.Stat(suiteRoot); err != nil {
		t.Skipf("W3C suite not checked out (%v); run: git submodule update --init testdata/xsdtests", err)
	}
	cases := collectInstanceCases(t)
	if len(cases) == 0 {
		t.Fatal("no XSD 1.1 instance cases found under " + suiteRoot)
	}
	want := readInstanceExpectations(t)

	correct := map[string]bool{}
	for _, c := range cases {
		ran, ok := runInstanceCase(c)
		if !ran {
			continue
		}
		correct[c.id] = ok
	}

	if *updateInstanceExpectations {
		writeInstanceExpectations(t, want, correct)
		return
	}

	var regressions, unexpected []string
	for id, ok := range correct {
		exp, listed := want[id]
		switch {
		case listed && exp.skip != "":
		case listed && exp.pass && !ok:
			regressions = append(regressions, id)
		case !listed && ok:
			unexpected = append(unexpected, id)
		}
	}
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
	t.Logf("instance conformance: %d cases, %d recorded passes, %d skips", len(cases), passes, skips)
	if len(unexpected) > 0 {
		t.Errorf("%d unexpected pass(es) — coverage improved; re-run with -update-instance-expectations to record:\n  %s",
			len(unexpected), strings.Join(capList(unexpected, 30), "\n  "))
	}
	if len(regressions) > 0 {
		t.Errorf("%d regression(s) — these were recorded as passing and now fail:\n  %s",
			len(regressions), strings.Join(capList(regressions, 30), "\n  "))
	}
}

// collectInstanceCases walks every .testSet, builds the schema for each group
// whose schema the suite declares valid under 1.1 and which we build without
// error, then yields the group's applicable instance tests bound to that schema.
func collectInstanceCases(t *testing.T) []instanceCase {
	var setFiles []string
	filepath.WalkDir(suiteRoot, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".testSet") {
			setFiles = append(setFiles, p)
		}
		return nil
	})
	sort.Strings(setFiles)

	var cases []instanceCase
	for _, sf := range setFiles {
		data, err := os.ReadFile(sf)
		if err != nil {
			t.Fatalf("read %s: %v", sf, err)
		}
		var ts instScanSet
		if err := xml.Unmarshal(data, &ts); err != nil {
			t.Fatalf("unmarshal %s: %v", sf, err)
		}
		rel, _ := filepath.Rel(suiteRoot, sf)
		rel = filepath.ToSlash(rel)
		for _, g := range ts.Groups {
			if len(g.Instances) == 0 || g.Schema == nil {
				continue
			}
			if !scanVersionApplies(g.Version) || !scanVersionApplies(g.Schema.Version) {
				continue
			}
			if scanExpectedValidity(g.Schema.Expected) != "valid" {
				continue // instances are only meaningful against a valid schema
			}
			schemas, okSchema := buildGroupSchema(sf, g.Schema)
			if !okSchema {
				continue // we don't build this schema cleanly; not gated here
			}
			for _, inst := range g.Instances {
				if !scanVersionApplies(inst.Version) {
					continue
				}
				validity := scanExpectedValidity(inst.Expected)
				if validity != "valid" && validity != "invalid" {
					continue
				}
				docPath := filepath.Join(filepath.Dir(sf), filepath.FromSlash(inst.Doc.Href))
				cases = append(cases, instanceCase{
					id:        rel + "#" + g.Name + "#" + inst.Name,
					docPath:   docPath,
					schemas:   schemas,
					wantValid: validity == "valid",
				})
			}
		}
	}
	return cases
}

// buildGroupSchema loads and builds a group's schema document(s) as one schema
// set; ok is false if a root fails to load or the build reports errors.
func buildGroupSchema(setFile string, st *scanSchemaTest) (schemas []*xsd.Schema, ok bool) {
	errs := &xsd.ErrorList{}
	l := newLoader(FileResolver{}, errs)
	for _, d := range st.Docs {
		p := filepath.Join(filepath.Dir(setFile), filepath.FromSlash(d.Href))
		if err := l.loadRoot(p); err != nil {
			return nil, false
		}
	}
	built, err := finish(l, errs)
	if err != nil {
		return nil, false
	}
	return built, true
}

// runInstanceCase assesses one instance. ran is false when the instance file
// cannot be read or parsed as XML at all (suite/infra issue, not a verdict).
func runInstanceCase(c instanceCase) (ran, ok bool) {
	f, err := os.Open(c.docPath)
	if err != nil {
		return false, false
	}
	defer f.Close()
	v := xsdvalidate.New(c.schemas, nil)
	res, err := xmlsrc.Validate(v, f, c.docPath)
	if err != nil {
		return false, false
	}
	return true, res.Valid() == c.wantValid
}

func readInstanceExpectations(t *testing.T) map[string]expectation {
	want := map[string]expectation{}
	f, err := os.Open(instanceExpectationsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return want
		}
		t.Fatalf("open instance expectations: %v", err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		id, outcome, found := strings.Cut(line, "\t")
		if !found {
			t.Fatalf("malformed instance expectations line: %q", line)
		}
		outcome = strings.TrimSpace(outcome)
		switch {
		case outcome == "pass":
			want[id] = expectation{pass: true}
		case strings.HasPrefix(outcome, "skip:"):
			want[id] = expectation{skip: strings.TrimSpace(outcome[len("skip:"):])}
		default:
			t.Fatalf("unknown outcome %q in instance expectations for %q", outcome, id)
		}
	}
	return want
}

func writeInstanceExpectations(t *testing.T, want map[string]expectation, correct map[string]bool) {
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
	b.WriteString("# XSD 1.1 INSTANCE conformance expectations (W3C test suite). Generated by\n")
	b.WriteString("# `go test ./parser -run TestInstanceConformance -update-instance-expectations`.\n")
	b.WriteString("# Each line is `<testSet-relpath>#<group>#<instance>\\tpass|skip:<reason>`.\n")
	b.WriteString("# `pass` lines ratchet up automatically; `skip:` lines are curated by hand.\n")
	b.WriteString("# See PLAN-validate.md.\n")
	passes, skips := 0, 0
	for _, id := range ids {
		b.WriteString(id)
		b.WriteByte('\t')
		b.WriteString(lines[id])
		b.WriteByte('\n')
		if lines[id] == "pass" {
			passes++
		} else {
			skips++
		}
	}
	if err := os.WriteFile(instanceExpectationsPath, []byte(b.String()), 0o644); err != nil {
		t.Fatalf("write instance expectations: %v", err)
	}
	t.Logf("wrote %s: %d pass, %d skip", instanceExpectationsPath, passes, skips)
}
