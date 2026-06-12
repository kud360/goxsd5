package parser

// An informational scan of the W3C test suite: pass 1 must not report
// errors on schema documents the suite expects to be valid under XSD 1.1
// (false positives would sink the M9 ratchet before it starts). Opt-in via
// GOXSD5_SCAN=1 because it walks ~15k files; the real harness is M9.

import (
	"encoding/xml"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kud360/goxsd5/parser/xmltree"
	"github.com/kud360/goxsd5/xsd"
)

type scanTestSet struct {
	Groups []scanTestGroup `xml:"testGroup"`
}
type scanTestGroup struct {
	Name    string          `xml:"name,attr"`
	Version string          `xml:"version,attr"`
	Schema  *scanSchemaTest `xml:"schemaTest"`
}
type scanSchemaTest struct {
	Version  string         `xml:"version,attr"`
	Docs     []scanXlinkRef `xml:"schemaDocument"`
	Expected []scanExpected `xml:"expected"`
}
type scanXlinkRef struct {
	Href string `xml:"http://www.w3.org/1999/xlink href,attr"`
}
type scanExpected struct {
	Validity string `xml:"validity,attr"`
	Version  string `xml:"version,attr"`
}

func scanVersionApplies(v string) bool {
	return v == "" || strings.Contains(v, "1.1")
}

// scanExpectedValidity picks the 1.1-applicable expectation: a versionless
// expected applies to all versions; an explicit 1.1 one wins.
func scanExpectedValidity(exps []scanExpected) string {
	out := ""
	for _, e := range exps {
		if e.Version == "" && out == "" {
			out = e.Validity
		}
		if strings.Contains(e.Version, "1.1") {
			out = e.Validity
		}
	}
	return out
}

func TestScanW3CSuiteValidSchemas(t *testing.T) {
	if os.Getenv("GOXSD5_SCAN") == "" {
		t.Skip("set GOXSD5_SCAN=1 to scan the W3C suite")
	}
	root := filepath.Join("..", "testdata", "xsdtests")
	if _, err := os.Stat(root); err != nil {
		t.Skipf("W3C suite not checked out: %v", err)
	}
	var setFiles []string
	filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".testSet") {
			setFiles = append(setFiles, p)
		}
		return nil
	})

	total, failed, parseFail := 0, 0, 0
	counts := map[string]int{}
	var samples []string
	for _, sf := range setFiles {
		data, err := os.ReadFile(sf)
		if err != nil {
			t.Fatalf("read %s: %v", sf, err)
		}
		var ts scanTestSet
		if err := xml.Unmarshal(data, &ts); err != nil {
			t.Fatalf("unmarshal %s: %v", sf, err)
		}
		for _, g := range ts.Groups {
			if g.Schema == nil || !scanVersionApplies(g.Version) || !scanVersionApplies(g.Schema.Version) {
				continue
			}
			if scanExpectedValidity(g.Schema.Expected) != "valid" {
				continue
			}
			errs := &xsd.ErrorList{}
			docOK := true
			for _, d := range g.Schema.Docs {
				p := filepath.Join(filepath.Dir(sf), filepath.FromSlash(d.Href))
				f, err := os.Open(p)
				if err != nil {
					docOK = false
					continue
				}
				node, err := xmltree.Parse(f, p)
				f.Close()
				if err != nil {
					parseFail++
					docOK = false
					continue
				}
				loadDoc(node, p, errs)
			}
			if !docOK {
				continue
			}
			total++
			if !errs.Empty() {
				failed++
				for _, id := range xsd.RefIDs(errs.Err()) {
					counts[id]++
				}
				if len(samples) < 40 {
					samples = append(samples, g.Name+": "+fmt.Sprintf("%v", xsd.AllErrors(errs.Err())[0]))
				}
			}
		}
	}
	t.Logf("valid-expected groups: %d, with pass-1 errors: %d, xml parse failures: %d", total, failed, parseFail)
	t.Logf("error id histogram: %v", counts)
	for _, s := range samples {
		t.Logf("  %s", s)
	}
}
