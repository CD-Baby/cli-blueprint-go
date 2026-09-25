# Guide

What the skeleton decides for you, and why. Change any of it — but change it
knowingly.

## 1. The binary is a contract, not a program

A CLI that an agent or a script calls is an API. Three things make it one:

- **A stable exit code table.** The caller branches on the code without
  parsing text.
- **A machine-readable output mode.** `--json` emits one envelope, on one
  line, on stdout, and nothing else.
- **Stable error codes.** `VALIDATION_FAILED` means the same thing in
  version 9 as in version 1.

`internal/result` and the exit code constants in `root.go` hold that contract.
Treat both as public API.

## 2. stdout is the result; stderr is the story

| Mode | Success | Failure |
| ---- | ------- | ------- |
| default | handler plaintext on stdout | nothing on stdout, diagnostic on stderr |
| `--json` | one envelope on stdout | one envelope on stdout, diagnostic on stderr |

The envelope's `ok` and the exit code always agree. A caller that reads stdout
never has to strip a log line first. This is why the skeleton forbids logging
to stdout.

## 3. Every command goes through one wrapper

`run(g, "<command>", handler)` is the only way a command produces output. The
handler returns data; the wrapper decides the format. A command that writes to
stdout itself breaks the contract, and the next command written by copying it
breaks it again.

```go
RunE: run(g, "thing", func(cmd *cobra.Command, args []string) (*cmdResult, error) {
    if bad {
        return nil, validationError("why it is bad", detail)
    }
    return &cmdResult{data: d, plain: "human text\n", warnings: w}, nil
}),
```

## 3a. Logging is stderr, and it is real

`--log-level` and `--quiet` are wired to a `log/slog` handler on stderr. Read
the logger through `g.log()`, which falls back to a discard logger so a handler
never needs a nil check.

An unknown level is a usage error, not a fallback to `info`. If someone types
`--log-level debgu`, the right answer is to say so, not to run the command
with the logs they asked for silently dropped.

Log at `debug` for anything the caller did not ask to see. A command run at the
default level should print its result and nothing else.

## 4. Failures are typed

`errors.go` holds one constructor per exit code. A handler returns one of
them. The wrapper turns an untyped error into `INTERNAL_ERROR` with exit 1 —
which is correct, because an untyped error is a bug you did not anticipate.

Pick the code by asking who is at fault:

| Constructor | Exit | Fault |
| ----------- | ---- | ----- |
| `usageError` | 2 | the caller typed it wrong |
| `validationError` | 3 | the input is well-formed but wrong |
| `prereqError` | 4 | something must exist first |
| `internalError` | 1 | we are at fault |

## 5. The command tree holds no logic

`internal/cli` parses flags and formats output. Real work lives in its own
package under `internal/`. This is not architecture for its own sake: it is
what lets you test the work without a `*cobra.Command`, and what keeps
`root.go` readable at command twenty.

The blueprint ships the split rather than only describing it.
`internal/greeting` validates the name and builds the text. It returns
`ErrNoName` and `*ErrNameTooLong`, its own values, and knows nothing about
exit codes. `hello.go` calls it and translates:

```go
func helloFailure(err error) error {
	var tooLong *greeting.ErrNameTooLong
	switch {
	case errors.Is(err, greeting.ErrNoName):
		return prereqError("no name given: ...", nil)
	case errors.As(err, &tooLong):
		return validationError(tooLong.Error(),
			map[string]int{"length": tooLong.Len, "limit": tooLong.Limit})
	default:
		return internalError(err.Error(), nil)
	}
}
```

The domain decides what is wrong. The command decides what that costs the
caller. Keep that seam and both halves stay testable on their own.

## 6. Three test layers, all required

- **Table tests** next to the code. Fast, precise, they cover branches.
- **`.txtar` scripts** under `internal/cli/testdata/script/`. They run the
  real binary and assert on exit codes, stdout, stderr and both output modes.
- **Golden files** under `internal/cli/testdata/golden/`. Help text is a
  published contract, so it is pinned. `make update-golden` rewrites the
  files; the diff is the point, so read it before committing.

A `.txtar` file is a script plus its fixture files in one text file:

```
exec mycli hello Kit --json
stdout '"ok":true'

! exec mycli hello
stderr 'no name given'

-- fixture.json --
{"some": "input"}
```

`make cover` enforces 85% over `./internal/...`. The gate is on library code
only; `cmd/` is one line and does not need covering.

## 7. Build metadata comes from the linker

`Version`, `Commit` and `BuildDate` are `-ldflags` variables. No version file
to forget to bump, no git call at runtime, and the binary reports the truth
about itself when someone hands it to you six months later.

## 8. Environment variables bind through a table

```go
var envBindings = []struct{ flag, env string }{
    {"log-level", "MYCLI_LOG"},
}
```

Precedence is flag > env > default, and an explicitly set flag is never
overwritten. Add a row; do not add a second mechanism.

## What was deliberately left out

A workspace root and its discovery rules, a TOML config and secrets layer,
JSON Schema validation, and a persistent ledger. Each is worth having in the
project that needs it, and none belongs in every CLI. Add the one you need
when you need it, and keep it in its own package under `internal/`.

## Adding your first command

1. `cp -r internal/greeting internal/thing` and rewrite it. This is where
   the logic goes. It must test without cobra.
2. `cp internal/cli/hello.go internal/cli/thing.go`
3. Rename `newHelloCmd` to `newThingCmd`, call `thing`, and rewrite
   `helloFailure` into `thingFailure`.
4. Register it in `newRootCmdWith`.
5. Write the table tests for both packages and a `.txtar` script.
6. Delete `hello.go`, `hello_test.go`, `internal/greeting/`,
   `testdata/script/hello.txtar` and `testdata/golden/hello-help.txt`.
7. `make update-golden`, read the diff, then `make verify`
