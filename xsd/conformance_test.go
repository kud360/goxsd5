package xsd

// The conformance matrix guard. It keeps three things in sync so that
// "strict to spec" stays verifiable as the code grows:
//
//   - the SpecRef registry (every declared constraint),
//   - CONFORMANCE.md (the human-readable coverage matrix),
//   - the source (// spec: annotations and SpecRef constant usages).
//
// It fails if a constraint is declared but unreferenced, referenced but
// undeclared, enforced but untracked, or tracked with a stale Impl pointer.

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// repoFiles returns every non-test Go source file in the module.
func repoFiles(t *testing.T) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir("..", func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "testdata" || d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(p, ".go") && !strings.HasSuffix(p, "_test.go") {
			files = append(files, p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	return files
}

// baseID returns the registered SpecRef ID that is label, or of which label
// is a sub-clause (e.g. "p-props-correct.2.1" -> "p-props-correct"). The
// longest match wins; "" when none is registered.
func baseID(label string) string {
	best := ""
	for id := range Refs {
		if label == id || strings.HasPrefix(label, id+".") {
			if len(id) > len(best) {
				best = id
			}
		}
	}
	return best
}

type docRow struct {
	label, status, impl string
}

var rowRe = regexp.MustCompile(`^\|\s*([A-Za-z][^|]*?)\s*\|[^|]*\|[^|]*\|\s*([^|]*?)\s*\|\s*([^|]*?)\s*\|\s*$`)

func readMatrix(t *testing.T) []docRow {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "CONFORMANCE.md"))
	if err != nil {
		t.Fatalf("read CONFORMANCE.md: %v", err)
	}
	var rows []docRow
	for line := range strings.SplitSeq(string(data), "\n") {
		m := rowRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		label := strings.Fields(m[1])[0] // drop trailing "(annotation)"
		if label == "Constraint" {
			continue // header
		}
		rows = append(rows, docRow{label: label, status: m[2], impl: m[3]})
	}
	return rows
}

func enforced(status string) bool {
	s := strings.ToLower(status)
	return strings.HasPrefix(s, "done") || strings.HasPrefix(s, "wip")
}

// TestConformanceMatrixCoverage: every registered SpecRef has a matrix row.
func TestConformanceMatrixCoverage(t *testing.T) {
	covered := map[string]bool{}
	for _, r := range readMatrix(t) {
		if b := baseID(r.label); b != "" {
			covered[b] = true
		}
	}
	for id := range Refs {
		if !covered[id] {
			t.Errorf("SpecRef %q has no row in CONFORMANCE.md", id)
		}
	}
}

// TestConformanceMatrixRowsValid: every done/wip row names a registered
// constraint, and any Impl pointer names a file that exists.
func TestConformanceMatrixRowsValid(t *testing.T) {
	for _, r := range readMatrix(t) {
		if enforced(r.status) && baseID(r.label) == "" {
			t.Errorf("row %q is marked %q but matches no registered SpecRef", r.label, r.status)
		}
		if r.impl == "" {
			continue
		}
		file := strings.SplitN(r.impl, ":", 2)[0]
		if _, err := os.Stat(filepath.Join("..", file)); err != nil {
			t.Errorf("row %q Impl points to missing file %q", r.label, file)
		}
	}
}

// TestSpecRefConstantsReferenced: every SpecRef constant is either used to
// construct an error somewhere in the source, or its matrix row marks it
// deferred/N/A (a constant reserved ahead of enforcement). Anything else is
// a declared-but-unreferenced constant — dead weight or a forgotten wiring.
func TestSpecRefConstantsReferenced(t *testing.T) {
	src, err := os.ReadFile("specref.go")
	if err != nil {
		t.Fatalf("read specref.go: %v", err)
	}
	declRe := regexp.MustCompile(`(Spec[A-Za-z0-9]+)\s*=\s*ref\([12],\s*"([^"]+)"`)
	type constDecl struct{ name, id string }
	var decls []constDecl
	for _, m := range declRe.FindAllStringSubmatch(string(src), -1) {
		decls = append(decls, constDecl{name: m[1], id: m[2]})
	}
	if len(decls) != len(Refs) {
		t.Fatalf("found %d SpecRef constants but %d registry entries", len(decls), len(Refs))
	}

	// A constraint may go unreferenced only if every matrix row for it is
	// deferred/N-A (none claims enforcement).
	enforcedInMatrix := map[string]bool{}
	for _, r := range readMatrix(t) {
		if b := baseID(r.label); b != "" && enforced(r.status) {
			enforcedInMatrix[b] = true
		}
	}

	corpus := map[string]string{}
	for _, f := range repoFiles(t) {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		if !strings.HasSuffix(f, "specref.go") {
			corpus[f] = string(b) // the declaration itself doesn't count as use
		}
	}
	for _, d := range decls {
		used := false
		for _, body := range corpus {
			if wordRe(d.name).MatchString(body) {
				used = true
				break
			}
		}
		if used {
			continue
		}
		if enforcedInMatrix[d.id] {
			t.Errorf("SpecRef %s (%s) is marked enforced in CONFORMANCE.md but its constant is never referenced", d.name, d.id)
		}
		// else: deferred/N-A constant reserved ahead of enforcement — allowed.
	}
}

// TestSpecAnnotationsDeclared: every `// spec: <id>` annotation names a
// registered constraint (the comment text can't drift from the registry).
func TestSpecAnnotationsDeclared(t *testing.T) {
	annRe := regexp.MustCompile(`// spec:\s*([A-Za-z][A-Za-z0-9_.\-]*)`)
	for _, f := range repoFiles(t) {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		for _, m := range annRe.FindAllStringSubmatch(string(b), -1) {
			if baseID(m[1]) == "" {
				t.Errorf("%s: // spec: %s does not match any registered SpecRef", f, m[1])
			}
		}
	}
}

func wordRe(name string) *regexp.Regexp {
	return regexp.MustCompile(`\b` + regexp.QuoteMeta(name) + `\b`)
}
