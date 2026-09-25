package greeting

import (
	"errors"
	"strings"
	"testing"
)

func TestComposeBuildsTheGreeting(t *testing.T) {
	cases := []struct {
		name string
		in   string
		opts Options
		want string
	}{
		{"plain", "Kit", Options{}, "Hello, Kit!"},
		{"shout", "Kit", Options{Shout: true}, "HELLO, KIT!"},
		{"trims surrounding space", "  Kit  ", Options{}, "Hello, Kit!"},
		{"at the limit", strings.Repeat("a", MaxNameLen), Options{}, "Hello, " + strings.Repeat("a", MaxNameLen) + "!"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Compose(c.in, c.opts)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.want {
				t.Fatalf("want %q, got %q", c.want, got)
			}
		})
	}
}

func TestComposeRejectsAnEmptyName(t *testing.T) {
	for _, in := range []string{"", "   ", "\t\n"} {
		_, err := Compose(in, Options{})
		if !errors.Is(err, ErrNoName) {
			t.Fatalf("Compose(%q): want ErrNoName, got %v", in, err)
		}
	}
}

func TestComposeRejectsAnOverlongName(t *testing.T) {
	in := strings.Repeat("a", MaxNameLen+1)
	_, err := Compose(in, Options{})

	var tooLong *ErrNameTooLong
	if !errors.As(err, &tooLong) {
		t.Fatalf("want *ErrNameTooLong, got %v", err)
	}
	if tooLong.Len != MaxNameLen+1 || tooLong.Limit != MaxNameLen {
		t.Fatalf("bad measurements: %+v", tooLong)
	}
	if !strings.Contains(tooLong.Error(), "limit is 64") {
		t.Fatalf("message must state the limit: %s", tooLong.Error())
	}
}
