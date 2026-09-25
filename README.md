# cli-blueprint-go

**A production-ready Go CLI skeleton.**

One command scaffolds a new project that already builds, tests, lints and
cross-compiles. You write the first real command, not the eighth copy of the
same wiring.

---

## Quick start

```
git clone <this repo> ~/Code/cli-blueprint-go
~/Code/cli-blueprint-go/init.sh ~/Code/ledgerctl --module github.com/acme/ledgerctl
```

That produces a project that passes `make verify` before you write a line.

```
Options:
  --module PATH       Go module path (required)
  --binary NAME       binary name (default: basename of --module)
  --env-prefix NAME   environment variable prefix (default: BINARY_)
  --no-git            do not run git init
  --no-tidy           do not run go mod tidy
  --no-verify         do not build and test the result
```

## What you get

| Piece | What it does |
| ----- | ------------ |
| `cmd/<bin>/main.go` | entry point; holds no logic |
| `internal/cli/root.go` | command tree, persistent flags, exit code table |
| `internal/cli/run.go` | one output contract shared by every command |
| `internal/cli/errors.go` | typed errors, one constructor per exit code |
| `internal/cli/logging.go` | `log/slog` on stderr, driven by `--log-level` |
| `internal/cli/version.go` | build metadata injected by `-ldflags` |
| `internal/cli/hello.go` | worked example; copy it, then delete it |
| `internal/result/` | the `--json` output envelope |
| `Makefile` | build, test, race, cover gate, verify, lint, cross |
| `.golangci.yml` | golangci-lint v2 config |
| `.github/workflows/ci.yml` | vet, lint, race tests, coverage gate, cross-compile |
| `CLAUDE.md` | the working contract, ready for an agent to read |
| `testdata/script/*.txtar` | end-to-end command tests via `testscript` |

Read [`GUIDE.md`](GUIDE.md) for the contract the skeleton enforces and why.

## The template is real code

`template/` is not a directory of `{{PLACEHOLDER}}` tokens. It is a compiling,
tested Go module named `github.com/example/mycli`. CI builds it, tests it and
lints it on every push, so the skeleton cannot rot.

`init.sh` replaces three seed strings:

| Seed | Replaced with |
| ---- | ------------- |
| `github.com/example/mycli` | `--module` |
| `mycli` | `--binary` |
| `MYCLI_` | `--env-prefix` |

## Develop the blueprint

```
cd template && make verify    # the skeleton must stay green
./init_test.sh                # scaffold to a temp dir and check the result
```

`init_test.sh` is the real test: it scaffolds a project, asserts no seed
survived, then builds it, tests it, lints it, and checks the exit code
contract against the built binary.
