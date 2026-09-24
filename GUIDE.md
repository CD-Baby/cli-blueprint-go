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

## 6. Two test layers, both required

- **Table tests** next to the code. Fast, precise, they cover branches.
- **`.txtar` scripts** under `internal/cli/testdata/script/`. They run the
  real binary and assert on exit codes, stdout, stderr and both output modes.

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

`org-pulse-cli` also has workspace discovery, a TOML config and secrets layer,
JSON Schema validation, and a ledger. They are real and they work — but they
are that project's domain, not every CLI's. Copy them from `org-pulse-cli`
when a project needs them.

## Adding your first command

1. `cp internal/cli/hello.go internal/cli/thing.go`
2. Rename `newHelloCmd` to `newThingCmd` and rewrite the handler.
3. Register it in `newRootCmdWith`.
4. Write the table tests and a `.txtar` script.
5. Delete `hello.go`, `hello_test.go` and `testdata/script/hello.txtar`.
6. `make verify`
