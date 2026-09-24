# CLAUDE.md

## mycli

<!-- Replace this paragraph. State in two or three sentences what the binary
does, what it refuses to do, and who calls it. -->

A single Go binary. Describe the one job it performs.

## Design Principles

- One binary. No daemon, no plugin system, no runtime code generation.
- Deterministic. The same input produces the same output and the same exit code.
- The command tree holds no logic. `internal/cli` parses flags and formats
  output. Every real operation lives in a separate package under `internal/`.
- Fail with a typed error. A handler returns a `*cliError`, never a bare
  `errors.New`, unless the failure is a bug.
- Add a dependency only when the standard library cannot do the job.

## Stack

- Go 1.26
- `spf13/cobra` for the command tree, `spf13/pflag` for flags
- `spf13/viper` for environment binding
- `rogpeppe/go-internal/testscript` for end-to-end command tests

## Repository Layout

```
cmd/mycli/main.go      entry point; calls cli.Execute and exits
internal/cli/          command tree, global flags, exit codes, output contract
  root.go              persistent flags, exit code table, Execute
  errors.go            cliError constructors, one per exit code
  run.go               the RunE wrapper every command uses
  version.go           build metadata injected by -ldflags
  hello.go             worked example; copy it, then delete it
  testdata/script/     testscript .txtar end-to-end scripts
internal/result/       the --json output envelope
```

Add a domain package under `internal/` for each real operation. Keep files
focused. A file that grows past a few hundred lines is doing too much.

## Output Contract (get this exactly right)

- stdout carries the result. stderr carries diagnostics and logs. Never mix
  them.
- Under `--json`, stdout carries exactly one envelope, on one line, and nothing
  else. This holds on success and on failure.
- Without `--json`, a success prints the handler's plaintext. A failure prints
  nothing on stdout.
- The envelope's `ok` field and the process exit code always agree.
- `warnings` always serializes as an array, never as `null`.
- Error `code` values are a public contract. Callers branch on them. Do not
  change what an existing code means.

## Exit Codes (stable contract — verbatim in root `--help`)

| Code | Meaning |
| ---- | ------- |
| 0 | success |
| 1 | internal error (bug or I/O failure) |
| 2 | usage error (unknown command or flag, bad argument) |
| 3 | validation failure (input failed schema or semantic checks) |
| 4 | missing prerequisite |

Add a code only when no existing code fits. Document the new code in
`root.go`, in the root `Long` help text, and in this table.

## Adding a Command

1. Copy `hello.go` to `<command>.go`.
2. Wrap the handler in `run(g, "<command>", ...)`.
3. Return a `*cmdResult`. Never write to stdout from the handler.
4. Return a typed error from `errors.go` for every expected failure.
5. Register the command in `newRootCmdWith`.
6. Write table tests next to the file, and a `.txtar` script for the
   end-to-end behavior.

## Help Text Requirements

Every command's `Long` text ends with three lines:

```
READS    what the command reads
WRITES   what the command writes
NEVER    the boundary the command does not cross
```

## Testing (a requirement, not an aspiration — CI-enforced)

- Table tests live next to the code they cover.
- `.txtar` scripts under `internal/cli/testdata/script/` cover command
  behavior end to end: exit codes, both output modes, and stderr.
- `make cover` enforces 85% statement coverage over `./internal/...`.
- `make verify` (vet + lint + test) is the pre-push gate.

## Build & Distribution

```
make build     # bin/mycli for the host
make cross     # linux/amd64, linux/arm64, darwin/arm64, static
```

Version, commit, and build date are injected at link time. Do not read them
from a file at runtime.

## Conventions

- Document every exported symbol. Say why, not what.
- Wrap errors with context: `fmt.Errorf("read %s: %w", path, err)`.
- No `panic` outside `main`. No `os.Exit` outside `main`.
- Do not log to stdout. Ever.
