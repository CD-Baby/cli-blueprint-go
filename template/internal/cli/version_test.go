package cli

import (
	"runtime"
	"strings"
	"testing"
)

func TestVersionPlainOutput(t *testing.T) {
	out, code := runRoot(t, "version")
	if code != ExitOK {
		t.Fatalf("want exit 0, got %d (%s)", code, out)
	}
	for _, want := range []string{"mycli", Version, Commit, runtime.Version()} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q: %s", want, out)
		}
	}
}

func TestVersionJSONEnvelope(t *testing.T) {
	out, code := runRoot(t, "version", "--json")
	if code != ExitOK {
		t.Fatalf("want exit 0, got %d (%s)", code, out)
	}
	for _, want := range []string{`"ok":true`, `"command":"version"`, `"build_date"`, `"go":`} {
		if !strings.Contains(out, want) {
			t.Fatalf("envelope missing %s: %s", want, out)
		}
	}
}

func TestVersionRejectsArguments(t *testing.T) {
	_, code := runRoot(t, "version", "extra")
	if code != ExitUsage {
		t.Fatalf("want exit %d, got %d", ExitUsage, code)
	}
}
