# mycli

<!-- Replace this paragraph with what the binary does. -->

A single Go binary. One sentence on the job it performs.

See [`CLAUDE.md`](CLAUDE.md) for the working contract: the output envelope, the
exit code table, and the rules for adding a command.

## Install

```
make build      # bin/mycli
```

## Use

```
mycli hello Kit                # Hello, Kit!
mycli hello Kit --json         # {"ok":true,"command":"hello",...}
mycli version
mycli --help

mycli hello Kit --log-level debug   # diagnostics on stderr; stdout stays clean
mycli hello Kit --quiet             # errors only
```

## Exit codes

| Code | Meaning |
| ---- | ------- |
| 0 | success |
| 1 | internal error |
| 2 | usage error |
| 3 | validation failure |
| 4 | missing prerequisite |

## Develop

```
make build     # build bin/mycli
make test      # go test ./...
make race      # tests with the race detector
make cover     # tests with the CI coverage gate (85%)
make verify    # vet + lint + test (pre-push gate)
make lint      # golangci-lint
make cross     # build every release target
make help      # list all targets
```

Run a single test:

```
go test ./internal/cli -run TestHelloJSONEnvelope
```

Run a single script:

```
go test ./internal/cli -run 'TestScripts/hello'
```
