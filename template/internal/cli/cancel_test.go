package cli

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestCanceledContextExitsWith130(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled when the command starts

	out, code := runRootCtx(ctx, t, "hello", "Kit", "--delay", "30s", "--json")
	if code != ExitCanceled {
		t.Fatalf("want exit %d, got %d (%s)", ExitCanceled, code, out)
	}
	if !strings.Contains(out, `"code":"CANCELED"`) {
		t.Fatalf("bad envelope: %s", out)
	}
	if !strings.Contains(out, `"ok":false`) {
		t.Fatalf("envelope must report failure: %s", out)
	}
}

func TestCancelDuringTheWaitStopsTheCommand(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	_, code := runRootCtx(ctx, t, "hello", "Kit", "--delay", "30s")
	elapsed := time.Since(start)

	if code != ExitCanceled {
		t.Fatalf("want exit %d, got %d", ExitCanceled, code)
	}
	if elapsed > 5*time.Second {
		t.Fatalf("command did not unwind promptly: took %s", elapsed)
	}
}

func TestDelayCompletesWhenNothingCancels(t *testing.T) {
	out, code := runRoot(t, "hello", "Kit", "--delay", "1ms")
	if code != ExitOK {
		t.Fatalf("want exit 0, got %d (%s)", code, out)
	}
	if out != "Hello, Kit!\n" {
		t.Fatalf("got %q", out)
	}
}

func TestCanceledErrorCarriesTheDocumentedCode(t *testing.T) {
	e := canceledError()
	if e.ExitCode() != ExitCanceled {
		t.Fatalf("want %d, got %d", ExitCanceled, e.ExitCode())
	}
	if e.Result().Code != "CANCELED" {
		t.Fatalf("want CANCELED, got %s", e.Result().Code)
	}
}

func TestDeadlineExceededIsAlsoTreatedAsCanceled(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	_, code := runRootCtx(ctx, t, "hello", "Kit", "--delay", "30s")
	if code != ExitCanceled {
		t.Fatalf("want exit %d, got %d", ExitCanceled, code)
	}
}
