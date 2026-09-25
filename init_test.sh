#!/usr/bin/env bash
# Prove the blueprint works: scaffold a project into a temp directory, then
# build, vet, test and lint it. CI runs this on every push.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

TARGET="$WORK/ledgerctl"
MODULE="github.com/acme/ledgerctl"

fail() { printf 'FAIL: %s\n' "$*" >&2; exit 1; }
pass() { printf 'ok   %s\n' "$*"; }

printf '== scaffolding %s\n' "$MODULE"
"$SCRIPT_DIR/init.sh" "$TARGET" --module "$MODULE" --no-git

cd "$TARGET"

printf '\n== checking the seeds were replaced\n'
[ -f cmd/ledgerctl/main.go ] || fail "cmd directory was not renamed"
pass "cmd/ledgerctl/main.go exists"

[ ! -d cmd/mycli ] || fail "cmd/mycli still exists"
pass "cmd/mycli is gone"

if grep -rn 'mycli' . --exclude-dir=.git >/dev/null 2>&1; then
  grep -rn 'mycli' . --exclude-dir=.git >&2
  fail "the binary seed survived"
fi
pass "no 'mycli' remains"

if grep -rn 'MYCLI_' . --exclude-dir=.git >/dev/null 2>&1; then
  grep -rn 'MYCLI_' . --exclude-dir=.git >&2
  fail "the env prefix seed survived"
fi
pass "no 'MYCLI_' remains"

grep -q "^module $MODULE\$" go.mod || fail "go.mod module path is wrong"
pass "go.mod declares $MODULE"

grep -q 'LEDGERCTL_LOG' internal/cli/root.go || fail "env prefix was not applied"
pass "env prefix is LEDGERCTL_"

[ -f LICENSE ] || fail "LICENSE was not carried over"
grep -q 'Virgin Music' LICENSE || fail "LICENSE lost its copyright holder"
pass "LICENSE is present"

[ -d internal/greeting ] || fail "the domain package was not carried over"
pass "internal/greeting is present"

grep -q 'ledgerctl \[command\]' internal/cli/testdata/golden/root-help.txt \
  || fail "golden help text was not reseeded"
pass "golden files were reseeded"

printf '\n== building and testing the scaffolded project\n'
go vet ./... || fail "go vet"
pass "go vet"

go test ./... >/dev/null || fail "go test"
pass "go test"

make build >/dev/null || fail "make build"
[ -x bin/ledgerctl ] || fail "bin/ledgerctl was not produced"
pass "make build produced bin/ledgerctl"

printf '\n== checking the built binary behaves\n'
out="$(./bin/ledgerctl hello Kit)"
[ "$out" = "Hello, Kit!" ] || fail "hello printed '$out'"
pass "hello prints the greeting"

out="$(./bin/ledgerctl hello Kit --json)"
case "$out" in
  *'"command":"hello"'*) pass "--json emits an envelope" ;;
  *) fail "--json emitted '$out'" ;;
esac

set +e
./bin/ledgerctl hello >/dev/null 2>&1
code=$?
set -e
[ "$code" -eq 4 ] || fail "missing prerequisite should exit 4, got $code"
pass "exit code contract holds (4 on missing prerequisite)"

set +e
./bin/ledgerctl nope >/dev/null 2>&1
code=$?
set -e
[ "$code" -eq 2 ] || fail "unknown command should exit 2, got $code"
pass "exit code contract holds (2 on unknown command)"

printf '\n== checking SIGINT cancels instead of killing\n'
./bin/ledgerctl hello Kit --delay 30s >/dev/null 2>&1 &
pid=$!
sleep 1
kill -INT "$pid" 2>/dev/null || true
set +e
wait "$pid"
code=$?
set -e
[ "$code" -eq 130 ] || fail "SIGINT should exit 130, got $code"
pass "SIGINT unwinds and exits 130"

printf '\n== checking the Dockerfile was reseeded\n'
grep -q 'internal/cli.Version' Dockerfile || fail "Dockerfile lost its ldflags"
grep -q 'github.com/acme/ledgerctl/internal/cli' Dockerfile \
  || fail "Dockerfile ldflags still point at the seed module"
grep -q '/out/ledgerctl' Dockerfile || fail "Dockerfile binary path was not reseeded"
pass "Dockerfile targets the scaffolded module and binary"

if docker info >/dev/null 2>&1; then
  printf '\n== building the container image\n'
  docker build -q -t ledgerctl:selftest . >/dev/null || fail "docker build"
  out="$(docker run --rm ledgerctl:selftest hello Kit)"
  [ "$out" = "Hello, Kit!" ] || fail "container printed '$out'"
  docker image rm -f ledgerctl:selftest >/dev/null 2>&1 || true
  pass "container image builds and runs"
else
  printf '\n-- docker daemon not available, skipping image build\n'
fi

if command -v golangci-lint >/dev/null 2>&1; then
  printf '\n== linting\n'
  golangci-lint run || fail "golangci-lint"
  pass "golangci-lint"
else
  printf '\n-- golangci-lint not installed, skipping\n'
fi

printf '\n== rejecting bad input\n'
set +e
"$SCRIPT_DIR/init.sh" "$TARGET" --module "$MODULE" --no-git >/dev/null 2>&1
code=$?
set -e
[ "$code" -ne 0 ] || fail "init.sh overwrote a non-empty directory"
pass "init.sh refuses a non-empty target"

set +e
"$SCRIPT_DIR/init.sh" "$WORK/x" --no-git >/dev/null 2>&1
code=$?
set -e
[ "$code" -ne 0 ] || fail "init.sh accepted a missing --module"
pass "init.sh requires --module"

printf '\nALL CHECKS PASSED\n'
