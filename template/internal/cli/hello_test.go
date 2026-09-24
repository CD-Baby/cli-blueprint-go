package cli

import (
	"strings"
	"testing"
)

func TestHelloPlainOutput(t *testing.T) {
	out, code := runRoot(t, "hello", "Kit")
	if code != ExitOK {
		t.Fatalf("want exit 0, got %d (%s)", code, out)
	}
	if out != "Hello, Kit!\n" {
		t.Fatalf("got %q", out)
	}
}

func TestHelloJSONEnvelope(t *testing.T) {
	out, code := runRoot(t, "hello", "Kit", "--json")
	if code != ExitOK {
		t.Fatalf("want exit 0, got %d (%s)", code, out)
	}
	for _, want := range []string{`"ok":true`, `"command":"hello"`, `"greeting":"Hello, Kit!"`, `"warnings":[]`} {
		if !strings.Contains(out, want) {
			t.Fatalf("envelope missing %s: %s", want, out)
		}
	}
}

func TestHelloShoutAddsWarning(t *testing.T) {
	out, code := runRoot(t, "hello", "Kit", "--shout", "--json")
	if code != ExitOK {
		t.Fatalf("want exit 0, got %d (%s)", code, out)
	}
	if !strings.Contains(out, "HELLO, KIT!") {
		t.Fatalf("want upper-cased greeting: %s", out)
	}
	if !strings.Contains(out, "shout mode") {
		t.Fatalf("want a warning in the envelope: %s", out)
	}
}

func TestHelloDryRunAddsWarning(t *testing.T) {
	out, _ := runRoot(t, "hello", "Kit", "--dry-run", "--json")
	if !strings.Contains(out, "dry run") {
		t.Fatalf("want a dry-run warning: %s", out)
	}
}

func TestHelloReadsNameFromEnv(t *testing.T) {
	t.Setenv("MYCLI_NAME", "Ada")
	out, code := runRoot(t, "hello")
	if code != ExitOK {
		t.Fatalf("want exit 0, got %d (%s)", code, out)
	}
	if out != "Hello, Ada!\n" {
		t.Fatalf("got %q", out)
	}
}

func TestHelloWithoutNameIsPrereqFailure(t *testing.T) {
	t.Setenv("MYCLI_NAME", "")
	out, code := runRoot(t, "hello", "--json")
	if code != ExitPrereq {
		t.Fatalf("want exit %d, got %d (%s)", ExitPrereq, code, out)
	}
	if !strings.Contains(out, `"code":"MISSING_PREREQUISITE"`) {
		t.Fatalf("bad envelope: %s", out)
	}
}

func TestHelloOverlongNameIsValidationFailure(t *testing.T) {
	long := strings.Repeat("a", maxNameLen+1)
	out, code := runRoot(t, "hello", long, "--json")
	if code != ExitValidation {
		t.Fatalf("want exit %d, got %d (%s)", ExitValidation, code, out)
	}
	if !strings.Contains(out, `"code":"VALIDATION_FAILED"`) || !strings.Contains(out, `"limit":64`) {
		t.Fatalf("bad envelope: %s", out)
	}
}

func TestHelloRejectsTwoArguments(t *testing.T) {
	_, code := runRoot(t, "hello", "a", "b")
	if code != ExitUsage {
		t.Fatalf("want exit %d, got %d", ExitUsage, code)
	}
}
