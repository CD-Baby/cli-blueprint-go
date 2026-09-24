// Package result defines the machine-readable output envelope every command
// emits under --json, on success or failure. Exactly one envelope is printed
// per invocation, as a single compact line on stdout.
package result

import (
	"encoding/json"
	"io"
)

// Envelope is the --json result object. stdout carries the envelope and nothing
// else; the exit code and OK always agree.
type Envelope struct {
	OK       bool     `json:"ok"`
	Command  string   `json:"command"`
	Data     any      `json:"data,omitempty"`
	Error    *Error   `json:"error,omitempty"`
	Warnings []string `json:"warnings"`
}

// Error is the stable, agent-branchable failure payload. Code values are part
// of the public contract and must not change meaning across versions.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Detail  any    `json:"detail,omitempty"`
}

// Success builds an ok envelope. warnings is normalized to a non-nil slice so
// it always serializes as [] rather than null.
func Success(command string, data any, warnings ...string) Envelope {
	return Envelope{OK: true, Command: command, Data: data, Warnings: norm(warnings)}
}

// Failure builds an error envelope.
func Failure(command string, err *Error, warnings ...string) Envelope {
	return Envelope{OK: false, Command: command, Error: err, Warnings: norm(warnings)}
}

func norm(w []string) []string {
	if w == nil {
		return []string{}
	}
	return w
}

// Write emits the envelope as a single compact line (trailing newline) to w.
// HTML escaping is disabled so message text and JSON pointers render verbatim.
func (e Envelope) Write(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(e)
}
