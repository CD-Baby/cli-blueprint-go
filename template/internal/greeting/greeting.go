// Package greeting produces the text this CLI exists to emit.
//
// It is the worked example of the rule that the command tree holds no logic.
// Everything here is testable without a *cobra.Command, knows nothing about
// flags, exit codes or output formats, and returns domain errors that the cli
// layer maps onto the exit code contract.
package greeting

import (
	"errors"
	"fmt"
	"strings"
)

// MaxNameLen bounds the name so the package has a real rule to enforce.
const MaxNameLen = 64

// ErrNoName reports that no name was supplied. The cli layer turns this into a
// missing-prerequisite failure.
var ErrNoName = errors.New("no name given")

// ErrNameTooLong reports a name over the limit, and carries the numbers so the
// caller can put them in a structured error payload without re-measuring.
type ErrNameTooLong struct {
	Len   int
	Limit int
}

func (e *ErrNameTooLong) Error() string {
	return fmt.Sprintf("name is %d bytes; the limit is %d", e.Len, e.Limit)
}

// Options controls how Compose renders the greeting.
type Options struct {
	// Shout upper-cases the result.
	Shout bool
}

// Compose validates name and builds the greeting. It returns ErrNoName or
// *ErrNameTooLong, never a formatted CLI message: choosing the exit code and
// the wording is the command layer's job.
func Compose(name string, opts Options) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", ErrNoName
	}
	if len(name) > MaxNameLen {
		return "", &ErrNameTooLong{Len: len(name), Limit: MaxNameLen}
	}

	greeting := "Hello, " + name + "!"
	if opts.Shout {
		greeting = strings.ToUpper(greeting)
	}
	return greeting, nil
}
