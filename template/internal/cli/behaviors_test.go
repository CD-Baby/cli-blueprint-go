package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// runRoot drives a fresh command tree in-process with output captured, and
// returns the combined stdout/stderr plus the process exit code.
func runRoot(t *testing.T, args ...string) (string, int) {
	t.Helper()
	root := newRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	code := execute(root)
	return out.String(), code
}

func TestRunWrapsNonCliErrorAsInternal(t *testing.T) {
	g := &globalFlags{json: true}
	cmd := &cobra.Command{Use: "x"}
	var out bytes.Buffer
	cmd.SetOut(&out)

	h := run(g, "demo", func(_ *cobra.Command, _ []string) (*cmdResult, error) {
		return nil, errors.New("boom")
	})
	err := h(cmd, nil)

	var ce *cliError
	if !errors.As(err, &ce) || ce.ExitCode() != ExitInternal {
		t.Fatalf("want internal cliError, got %v", err)
	}
	if !strings.Contains(out.String(), `"code":"INTERNAL_ERROR"`) {
		t.Fatalf("bad envelope: %s", out.String())
	}
}

func TestRunStaysSilentOnFailureWithoutJSON(t *testing.T) {
	g := &globalFlags{}
	cmd := &cobra.Command{Use: "x"}
	var out bytes.Buffer
	cmd.SetOut(&out)

	h := run(g, "demo", func(_ *cobra.Command, _ []string) (*cmdResult, error) {
		return nil, validationError("nope", nil)
	})
	if err := h(cmd, nil); err == nil {
		t.Fatal("want an error")
	}
	if out.String() != "" {
		t.Fatalf("stdout must stay empty without --json, got %q", out.String())
	}
}

func TestRunToleratesNilResult(t *testing.T) {
	g := &globalFlags{}
	cmd := &cobra.Command{Use: "x"}
	var out bytes.Buffer
	cmd.SetOut(&out)

	h := run(g, "demo", func(_ *cobra.Command, _ []string) (*cmdResult, error) {
		return nil, nil
	})
	if err := h(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.String() != "" {
		t.Fatalf("want no output, got %q", out.String())
	}
}

func TestErrorConstructorsCarryTheirExitCode(t *testing.T) {
	cases := []struct {
		name string
		err  *cliError
		code int
		ecod string
	}{
		{"usage", usageError("u", nil), ExitUsage, "USAGE_ERROR"},
		{"validation", validationError("v", nil), ExitValidation, "VALIDATION_FAILED"},
		{"prereq", prereqError("p", nil), ExitPrereq, "MISSING_PREREQUISITE"},
		{"internal", internalError("i", nil), ExitInternal, "INTERNAL_ERROR"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if c.err.ExitCode() != c.code {
				t.Fatalf("exit code: want %d, got %d", c.code, c.err.ExitCode())
			}
			if c.err.Result().Code != c.ecod {
				t.Fatalf("error code: want %s, got %s", c.ecod, c.err.Result().Code)
			}
			if c.err.Error() == "" {
				t.Fatal("Error() must carry the message")
			}
		})
	}
}

func TestExecuteReturnsUsageForUnknownCommand(t *testing.T) {
	_, code := runRoot(t, "no-such-command")
	if code != ExitUsage {
		t.Fatalf("want %d, got %d", ExitUsage, code)
	}
}

func TestExecuteReturnsOKForHelp(t *testing.T) {
	out, code := runRoot(t, "--help")
	if code != ExitOK {
		t.Fatalf("want %d, got %d", ExitOK, code)
	}
	if !strings.Contains(out, "EXIT CODES") {
		t.Fatalf("root help must document the exit code table: %s", out)
	}
}

func TestQuietImpliesErrorLogLevel(t *testing.T) {
	g := &globalFlags{}
	root := newRootCmdWith(g)
	root.SetArgs([]string{"hello", "world", "--quiet"})
	var out bytes.Buffer
	root.SetOut(&out)
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.logLevel != "error" {
		t.Fatalf("want log level error, got %q", g.logLevel)
	}
}

func TestEnvSuppliesFlagDefault(t *testing.T) {
	t.Setenv("MYCLI_LOG", "debug")
	g := &globalFlags{}
	root := newRootCmdWith(g)
	root.SetArgs([]string{"hello", "world"})
	var out bytes.Buffer
	root.SetOut(&out)
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.logLevel != "debug" {
		t.Fatalf("want log level from env, got %q", g.logLevel)
	}
}

func TestExplicitFlagBeatsEnv(t *testing.T) {
	t.Setenv("MYCLI_LOG", "debug")
	g := &globalFlags{}
	root := newRootCmdWith(g)
	root.SetArgs([]string{"hello", "world", "--log-level", "warn"})
	var out bytes.Buffer
	root.SetOut(&out)
	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.logLevel != "warn" {
		t.Fatalf("flag must beat env, got %q", g.logLevel)
	}
}
