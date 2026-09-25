package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// checkGolden compares got against testdata/golden/<name>.
//
// Setting UPDATE_GOLDEN rewrites the file instead of comparing. An environment
// variable is used rather than a -update flag so `go test ./...` regenerates
// every package's golden files in one command: a flag would have to be
// registered in every package, and any package that omitted it would fail the
// run.
func checkGolden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", "golden", name)

	if os.Getenv("UPDATE_GOLDEN") != "" {
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatalf("create golden dir: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Logf("updated %s", path)
		return
	}

	want, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("read %s: %v (run `make update-golden` to create it)", path, err)
	}
	if got != string(want) {
		t.Fatalf("output does not match %s\n--- want ---\n%s\n--- got ---\n%s", path, want, got)
	}
}

// The root help text is a published contract: CLAUDE.md requires the exit code
// table to appear in it verbatim. A golden file makes any edit to it visible in
// review instead of silent.
func TestRootHelpMatchesGolden(t *testing.T) {
	out, code := runRoot(t, "--help")
	if code != ExitOK {
		t.Fatalf("want exit 0, got %d", code)
	}
	checkGolden(t, "root-help.txt", out)
}

func TestHelloHelpMatchesGolden(t *testing.T) {
	out, code := runRoot(t, "hello", "--help")
	if code != ExitOK {
		t.Fatalf("want exit 0, got %d", code)
	}
	checkGolden(t, "hello-help.txt", out)
}
