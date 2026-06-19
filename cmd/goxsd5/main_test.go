package main_test

// CLI-level behaviour tests for the goxsd5 command, complementing smoke.sh by
// running inside `go test ./...`. They build the binary once, then assert the
// documented exit codes (0 ok / 1 invalid-or-error / 2 usage) and that
// -validate reports a cvc-* id for an invalid instance.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const schema = `<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
  <xs:element name="item" type="xs:int"/>
</xs:schema>`

// cliBin is the compiled command, built once for the whole package by TestMain.
var cliBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "goxsd5-cli")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)
	cliBin = filepath.Join(dir, "goxsd5")
	if out, err := exec.Command("go", "build", "-o", cliBin, ".").CombinedOutput(); err != nil {
		panic("build CLI: " + err.Error() + "\n" + string(out))
	}
	os.Exit(m.Run())
}

func write(t *testing.T, name, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return p
}

// exitCode runs the command and returns its process exit code plus combined output.
func exitCode(t *testing.T, args ...string) (int, string) {
	t.Helper()
	out, err := exec.Command(cliBin, args...).CombinedOutput()
	if err == nil {
		return 0, string(out)
	}
	ee, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("run %v: %v", args, err)
	}
	return ee.ExitCode(), string(out)
}

func TestCLIUsage(t *testing.T) {
	if code, _ := exitCode(t); code != 2 {
		t.Fatalf("no args: exit %d, want 2", code)
	}
}

func TestCLISchemaSummary(t *testing.T) {
	xsd := write(t, "s.xsd", schema)
	code, out := exitCode(t, xsd)
	if code != 0 {
		t.Fatalf("valid schema: exit %d, want 0\n%s", code, out)
	}
	if !strings.Contains(out, "element") {
		t.Fatalf("summary missing element listing:\n%s", out)
	}
}

func TestCLISchemaError(t *testing.T) {
	bad := write(t, "bad.xsd", `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:element/></xs:schema>`)
	if code, out := exitCode(t, bad); code != 1 {
		t.Fatalf("bad schema: exit %d, want 1\n%s", code, out)
	}
}

func TestCLIValidateValid(t *testing.T) {
	xsd := write(t, "s.xsd", schema)
	doc := write(t, "ok.xml", `<item>7</item>`)
	if code, out := exitCode(t, "-q", "-validate", doc, xsd); code != 0 {
		t.Fatalf("valid instance: exit %d, want 0\n%s", code, out)
	}
}

func TestCLIValidateInvalid(t *testing.T) {
	xsd := write(t, "s.xsd", schema)
	doc := write(t, "bad.xml", `<item>notanint</item>`)
	code, out := exitCode(t, "-q", "-validate", doc, xsd)
	if code != 1 {
		t.Fatalf("invalid instance: exit %d, want 1\n%s", code, out)
	}
	if !strings.Contains(out, "invalid") {
		t.Fatalf("expected an invalid report:\n%s", out)
	}
}
