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
- `log/slog` for stderr diagnostics
- `rogpeppe/go-internal/testscript` for end-to-end command tests

## Repository Layout

```
cmd/mycli/main.go      entry point; calls cli.Execute and exits
internal/cli/          command tree, global flags, exit codes, output contract
  root.go              persistent flags, exit code table, Execute
  errors.go            cliError constructors, one per exit code
  run.go               the RunE wrapper every command uses
  logging.go           the stderr logger and its level parsing
  version.go           build metadata injected by -ldflags
  hello.go             worked example; copy it, then delete it
  testdata/script/     testscript .txtar end-to-end scripts
  testdata/golden/     golden files for help text
internal/greeting/     worked example of a domain package; delete it with hello
internal/result/       the --json output envelope
```

Add a domain package under `internal/` for each real operation. `internal/cli`
must stay free of logic: it parses flags, calls a domain package, and maps the
domain's errors onto the exit code contract. `internal/greeting` and
`hello.go` are that split in miniature. Keep files focused. A file that grows
past a few hundred lines is doing too much.

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

## Signals and Cancellation

- `Execute` installs `signal.NotifyContext` for SIGINT and SIGTERM. A signal
  cancels the command's context; it does not kill the process, so deferred
  cleanup runs. A second signal hits the default handler and terminates at
  once, so a command that ignores the first can never become unkillable.
- Pass `cmd.Context()` into everything that blocks:
  `http.NewRequestWithContext`, `db.QueryContext`, `exec.CommandContext`, and
  any `select` that waits. A handler that blocks without the context is a bug.
- Return `ctx.Err()` when the wait loses. `run` maps `context.Canceled` and
  `context.DeadlineExceeded` onto exit 130, so handlers never build that error
  themselves.
- `hello --delay` is the worked example.

## Logging

- Logs go to stderr through `g.log()`. Never to stdout, under any flag.
- `--log-level` takes `debug`, `info`, `warn` or `error`. The default is
  `info`. `--quiet` raises the floor to `error`. `$MYCLI_LOG` supplies the
  level when the flag is absent.
- An unknown level is a usage error (exit 2), not a silent fallback. A typo
  must not quietly discard the logs the caller asked for.
- Use `debug` for anything a caller did not ask to see. A command at the
  default level prints its result and nothing more.

## Exit Codes (stable contract — verbatim in root `--help`)

| Code | Meaning |
| ---- | ------- |
| 0 | success |
| 1 | internal error (bug or I/O failure) |
| 2 | usage error (unknown command or flag, bad argument) |
| 3 | validation failure (input failed schema or semantic checks) |
| 4 | missing prerequisite |
| 130 | canceled by SIGINT or SIGTERM (128 + signal number) |

Add a code only when no existing code fits. Document the new code in
`root.go`, in the root `Long` help text, and in this table.

## Adding a Command

1. Put the work in a package under `internal/`. It must compile and test
   without cobra, and must return its own error values, not CLI errors.
2. Copy `hello.go` to `<command>.go`.
3. Wrap the handler in `run(g, "<command>", ...)`.
4. Call the domain package. Keep the handler free of logic.
5. Map the domain's errors onto the exit code contract, as `helloFailure`
   does. `errors.Is` for sentinels, `errors.As` for typed errors.
6. Return a `*cmdResult`. Never write to stdout from the handler.
7. Register the command in `newRootCmdWith`.
8. Write table tests beside both files, and a `.txtar` script for the
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
- Golden files under `internal/cli/testdata/golden/` pin the help text, so a
  change to a published contract shows up in review. Regenerate with
  `make update-golden`, and read the diff before you commit it.
- `make cover` enforces 85% statement coverage over `./internal/...`.
- `make verify` (vet + lint + test) is the pre-push gate.

## Build & Distribution

```
make build     # bin/mycli for the host
make cross     # linux/amd64, linux/arm64, darwin/arm64, static
make docker    # container image: scratch, static, non-root
```

Version, commit, and build date are injected at link time. Do not read them
from a file at runtime. The Dockerfile takes the same three values as build
args, so an image reports the same metadata a host build does.

The runtime stage is `scratch`: the binary, a CA bundle, and nothing else. Do
not add a shell or a package manager to debug a container. Reproduce the
problem with the host binary instead. If the image ever needs a writable path,
mount it; do not build one in.

## Conventions

- Document every exported symbol. Say why, not what.
- Wrap errors with context: `fmt.Errorf("read %s: %w", path, err)`.
- No `panic` outside `main`. No `os.Exit` outside `main`.
- Do not log to stdout. Ever. Use `g.log()`, which writes to stderr.

## License

`LICENSE` is proprietary: copyright Virgin Music, all rights reserved, with a
grant to Virgin Music and CD Baby staff for company work. Keep the file in
every project scaffolded from the blueprint, and do not publish this code
outside the company without approval.
