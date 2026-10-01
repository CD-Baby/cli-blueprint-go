package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/example/mycli/internal/result"
)

// runRootSplit runs the tree with stdout and stderr captured separately, since
// the property under test is what stdout alone carries.
func runRootSplit(t *testing.T, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	root := newRootCmd()
	var out, errBuf bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errBuf)
	code = execute(context.Background(), root, args)
	return out.String(), errBuf.String(), code
}

// oneEnvelope asserts stdout is exactly one JSON line and decodes it.
func oneEnvelope(t *testing.T, stdout string) result.Envelope {
	t.Helper()
	lines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	if stdout == "" || len(lines) != 1 {
		t.Fatalf("want exactly one envelope line on stdout, got %q", stdout)
	}
	var e result.Envelope
	if err := json.Unmarshal([]byte(lines[0]), &e); err != nil {
		t.Fatalf("stdout is not an envelope: %v (%q)", err, stdout)
	}
	return e
}

func TestUsageFailuresEmitOneEnvelopeUnderJSON(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		command string
	}{
		{"unknown command", []string{"nope", "--json"}, "mycli"},
		{"unknown flag after --json", []string{"hello", "Kit", "--json", "--bogus"}, "hello"},
		// pflag stops at --bogus and never parses --json, so only the raw
		// argument scan can see that JSON was requested.
		{"unknown flag before --json", []string{"hello", "Kit", "--bogus", "--json"}, "hello"},
		{"too many arguments", []string{"hello", "a", "b", "--json"}, "hello"},
		{"bad log level", []string{"hello", "Kit", "--log-level", "loud", "--json"}, "hello"},
		{"bad log level from env", []string{"hello", "Kit", "--json"}, "hello"},
		{"arguments to version", []string{"version", "extra", "--json"}, "version"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.name == "bad log level from env" {
				t.Setenv("MYCLI_LOG", "loud")
			}
			stdout, stderr, code := runRootSplit(t, c.args...)
			if code != ExitUsage {
				t.Fatalf("want exit %d, got %d", ExitUsage, code)
			}
			e := oneEnvelope(t, stdout)
			if e.OK {
				t.Fatalf("envelope must report failure: %s", stdout)
			}
			if e.Error == nil || e.Error.Code != "USAGE_ERROR" {
				t.Fatalf("want USAGE_ERROR, got %s", stdout)
			}
			if e.Command != c.command {
				t.Fatalf("want command %q, got %q", c.command, e.Command)
			}
			if stderr == "" {
				t.Fatal("the diagnostic must still reach stderr")
			}
		})
	}
}

func TestHandlerFailureIsWrittenExactlyOnce(t *testing.T) {
	t.Setenv("MYCLI_NAME", "")
	stdout, _, code := runRootSplit(t, "hello", "--json")
	if code != ExitPrereq {
		t.Fatalf("want exit %d, got %d", ExitPrereq, code)
	}
	if e := oneEnvelope(t, stdout); e.Error.Code != "MISSING_PREREQUISITE" {
		t.Fatalf("want MISSING_PREREQUISITE, got %s", stdout)
	}
}

func TestCanceledFailureIsWrittenExactlyOnce(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	root := newRootCmd()
	var out, errBuf bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errBuf)
	if code := execute(ctx, root, []string{"hello", "Kit", "--delay", "30s", "--json"}); code != ExitCanceled {
		t.Fatalf("want exit %d, got %d", ExitCanceled, code)
	}
	if e := oneEnvelope(t, out.String()); e.Error.Code != "CANCELED" {
		t.Fatalf("want CANCELED, got %s", out.String())
	}
}

func TestUsageFailureWithoutJSONLeavesStdoutEmpty(t *testing.T) {
	for _, args := range [][]string{
		{"nope"},
		{"hello", "Kit", "--bogus"},
		{"nope", "--json=false"},
	} {
		stdout, stderr, code := runRootSplit(t, args...)
		if code != ExitUsage {
			t.Fatalf("%v: want exit %d, got %d", args, ExitUsage, code)
		}
		if stdout != "" {
			t.Fatalf("%v: stdout must stay empty, got %q", args, stdout)
		}
		if stderr == "" {
			t.Fatalf("%v: the diagnostic must reach stderr", args)
		}
	}
}

func TestWantsJSON(t *testing.T) {
	cases := []struct {
		args []string
		want bool
	}{
		{nil, false},
		{[]string{"hello"}, false},
		{[]string{"--json"}, true},
		{[]string{"--json=true"}, true},
		{[]string{"--json=1"}, true},
		{[]string{"--json=false"}, false},
		{[]string{"--json=0"}, false},
		{[]string{"--json=yes"}, true},
		{[]string{"--json", "--json=false"}, false},
		{[]string{"--json=false", "--json"}, true},
		{[]string{"hello", "--", "--json"}, false},
		{[]string{"--jsonx"}, false},
	}
	for _, c := range cases {
		if got := wantsJSON(c.args); got != c.want {
			t.Errorf("wantsJSON(%q) = %v, want %v", c.args, got, c.want)
		}
	}
}
