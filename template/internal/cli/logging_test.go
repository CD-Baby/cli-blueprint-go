package cli

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestParseLevelAcceptsTheDocumentedSet(t *testing.T) {
	cases := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
		"WARN":  slog.LevelWarn,
		" info": slog.LevelInfo,
	}
	for in, want := range cases {
		got, err := parseLevel(in)
		if err != nil {
			t.Fatalf("parseLevel(%q): %v", in, err)
		}
		if got != want {
			t.Fatalf("parseLevel(%q): want %v, got %v", in, want, got)
		}
	}
}

func TestParseLevelRejectsUnknownValueAsUsageError(t *testing.T) {
	_, err := parseLevel("loud")
	if err == nil {
		t.Fatal("want an error")
	}
	ce, ok := err.(*cliError)
	if !ok || ce.ExitCode() != ExitUsage {
		t.Fatalf("want a usage cliError, got %#v", err)
	}
	if !strings.Contains(ce.Error(), "loud") {
		t.Fatalf("message must name the bad value: %s", ce.Error())
	}
}

func TestNewLoggerFiltersBelowItsLevel(t *testing.T) {
	var buf bytes.Buffer
	l := newLogger(&buf, slog.LevelWarn)
	l.Debug("hidden")
	l.Info("hidden")
	l.Warn("shown")
	out := buf.String()
	if strings.Contains(out, "hidden") {
		t.Fatalf("below-level records leaked: %s", out)
	}
	if !strings.Contains(out, "shown") {
		t.Fatalf("warn record missing: %s", out)
	}
}

func TestLogFallsBackToDiscardWhenUnset(t *testing.T) {
	g := &globalFlags{}
	if g.log() == nil {
		t.Fatal("log() must never return nil")
	}
	g.log().Error("must not panic")
}

func TestLogsGoToStderrNeverStdout(t *testing.T) {
	root := newRootCmd()
	var out, errBuf bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"hello", "Kit", "--log-level", "debug"})
	if code := execute(context.Background(), root); code != ExitOK {
		t.Fatalf("want exit 0, got %d", code)
	}
	if strings.Contains(out.String(), "resolved name") {
		t.Fatalf("a log record reached stdout: %s", out.String())
	}
	if !strings.Contains(errBuf.String(), "resolved name") {
		t.Fatalf("debug record missing from stderr: %s", errBuf.String())
	}
	if out.String() != "Hello, Kit!\n" {
		t.Fatalf("stdout must carry only the result, got %q", out.String())
	}
}

func TestDefaultLevelHidesDebugRecords(t *testing.T) {
	root := newRootCmd()
	var out, errBuf bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"hello", "Kit"})
	if code := execute(context.Background(), root); code != ExitOK {
		t.Fatalf("want exit 0, got %d", code)
	}
	if strings.Contains(errBuf.String(), "resolved name") {
		t.Fatalf("debug record leaked at the default level: %s", errBuf.String())
	}
}

func TestQuietSuppressesWarnRecords(t *testing.T) {
	root := newRootCmd()
	var out, errBuf bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"hello", "Kit", "--dry-run", "--quiet"})
	if code := execute(context.Background(), root); code != ExitOK {
		t.Fatalf("want exit 0, got %d", code)
	}
	if strings.Contains(errBuf.String(), "dry run") {
		t.Fatalf("--quiet must suppress warn records: %s", errBuf.String())
	}
}

func TestBadLogLevelIsAUsageFailure(t *testing.T) {
	_, code := runRoot(t, "hello", "Kit", "--log-level", "loud")
	if code != ExitUsage {
		t.Fatalf("want exit %d, got %d", ExitUsage, code)
	}
}

func TestBadLogLevelFromEnvIsAlsoRejected(t *testing.T) {
	t.Setenv("MYCLI_LOG", "loud")
	_, code := runRoot(t, "hello", "Kit")
	if code != ExitUsage {
		t.Fatalf("want exit %d, got %d", ExitUsage, code)
	}
}
