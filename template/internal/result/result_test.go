package result

import (
	"bytes"
	"strings"
	"testing"
)

func TestSuccessEnvelopeIsCompactSingleLine(t *testing.T) {
	var buf bytes.Buffer
	if err := Success("hello", map[string]any{"items": 42}).Write(&buf); err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if !strings.HasSuffix(out, "\n") {
		t.Fatalf("want trailing newline, got %q", out)
	}
	if strings.Count(strings.TrimRight(out, "\n"), "\n") != 0 {
		t.Fatalf("envelope must be one line, got %q", out)
	}
	if !strings.Contains(out, `"ok":true`) || !strings.Contains(out, `"warnings":[]`) {
		t.Fatalf("missing required fields: %s", out)
	}
}

func TestFailureEnvelopeOmitsDataAndCarriesStableCode(t *testing.T) {
	e := Failure("hello", &Error{
		Code:    "VALIDATION_FAILED",
		Message: "name must not be empty",
		Detail:  map[string]string{"field": "name"},
	})
	var buf bytes.Buffer
	if err := e.Write(&buf); err != nil {
		t.Fatalf("write: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, `"data"`) {
		t.Fatalf("failure envelope must omit data: %s", out)
	}
	if !strings.Contains(out, `"code":"VALIDATION_FAILED"`) {
		t.Fatalf("missing stable error code: %s", out)
	}
	if !strings.Contains(out, `"warnings":[]`) {
		t.Fatalf("warnings must serialize as []: %s", out)
	}
}

func TestWarningsAlwaysSerializeAsEmptyArray(t *testing.T) {
	var buf bytes.Buffer
	if err := Success("hello", nil).Write(&buf); err != nil {
		t.Fatalf("write: %v", err)
	}
	if !strings.Contains(buf.String(), `"warnings":[]`) {
		t.Fatalf("warnings missing or null: %s", buf.String())
	}
}

func TestWriteDoesNotEscapeHTML(t *testing.T) {
	var buf bytes.Buffer
	if err := Success("hello", map[string]string{"q": "a<b>c&d"}).Write(&buf); err != nil {
		t.Fatalf("write: %v", err)
	}
	if !strings.Contains(buf.String(), "a<b>c&d") {
		t.Fatalf("HTML was escaped: %s", buf.String())
	}
}
