#!/usr/bin/env bash
# Scaffold a new Go CLI project from template/.
#
# Usage:
#   ./init.sh <target-dir> --module <module-path> [--binary <name>] [options]
#
# Example:
#   ./init.sh ~/Code/ledgerctl --module github.com/acme/ledgerctl
#
# The template is a real, compiling, tested Go module. This script copies it and
# replaces three seed strings:
#
#   github.com/example/mycli  ->  --module
#   mycli                      ->  --binary (default: basename of --module)
#   MYCLI_                     ->  --env-prefix (default: binary, upper-cased)

set -euo pipefail

readonly SEED_MODULE="github.com/example/mycli"
readonly SEED_BINARY="mycli"
readonly SEED_PREFIX="MYCLI_"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly TEMPLATE_DIR="$SCRIPT_DIR/template"

die() { printf 'error: %s\n' "$*" >&2; exit 1; }
info() { printf '  %s\n' "$*"; }

usage() {
  sed -n '2,20p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'
  cat <<'EOF'

Options:
  --module PATH       Go module path (required)
  --binary NAME       binary name (default: basename of --module)
  --env-prefix NAME   environment variable prefix (default: BINARY_)
  --no-git            do not run git init
  --no-tidy           do not run go mod tidy
  --no-verify         do not build and test the result
  -h, --help          show this message
EOF
}

target=""
module=""
binary=""
prefix=""
do_git=1
do_tidy=1
do_verify=1

while [ $# -gt 0 ]; do
  case "$1" in
    -h|--help) usage; exit 0 ;;
    --module) module="${2:-}"; shift 2 ;;
    --binary) binary="${2:-}"; shift 2 ;;
    --env-prefix) prefix="${2:-}"; shift 2 ;;
    --no-git) do_git=0; shift ;;
    --no-tidy) do_tidy=0; shift ;;
    --no-verify) do_verify=0; shift ;;
    -*) die "unknown flag: $1" ;;
    *)
      [ -z "$target" ] || die "unexpected argument: $1"
      target="$1"; shift ;;
  esac
done

[ -n "$target" ] || { usage >&2; die "target directory is required"; }
[ -n "$module" ] || { usage >&2; die "--module is required"; }
[ -d "$TEMPLATE_DIR" ] || die "template directory not found: $TEMPLATE_DIR"

# Derive the binary name and the environment prefix when they were not given.
[ -n "$binary" ] || binary="$(basename "$module")"
if [ -z "$prefix" ]; then
  prefix="$(printf '%s' "$binary" | tr '[:lower:]-' '[:upper:]_')_"
fi

case "$module" in
  */*) ;;
  *) die "--module must be a full module path, for example github.com/you/$binary" ;;
esac
case "$binary" in
  *[!a-zA-Z0-9_-]*) die "--binary may contain only letters, digits, hyphen and underscore" ;;
esac

# Refuse to write into a directory that already holds files.
if [ -e "$target" ]; then
  [ -d "$target" ] || die "target exists and is not a directory: $target"
  if [ -n "$(ls -A "$target" 2>/dev/null)" ]; then
    die "target directory is not empty: $target"
  fi
fi

printf 'scaffolding %s\n' "$target"
info "module      $module"
info "binary      $binary"
info "env prefix  $prefix"

mkdir -p "$target"
# Copy the template, dotfiles included, minus build artifacts and any local git.
( cd "$TEMPLATE_DIR" && tar --exclude='./bin' --exclude='./.git' \
    --exclude='./coverage.out' --exclude='*.out' -cf - . ) | ( cd "$target" && tar -xf - )

# Replace the seeds. The module path is replaced first: it contains the binary
# seed, so doing it first keeps the second pass from touching it twice.
find "$target" -type f -print0 | while IFS= read -r -d '' f; do
  LC_ALL=C sed -i.bak \
    -e "s|$SEED_MODULE|$module|g" \
    -e "s|$SEED_PREFIX|$prefix|g" \
    -e "s|$SEED_BINARY|$binary|g" \
    "$f"
  rm -f "$f.bak"
done

# Rename the command directory to match the binary.
if [ "$binary" != "$SEED_BINARY" ]; then
  mv "$target/cmd/$SEED_BINARY" "$target/cmd/$binary"
fi

cd "$target"

if [ "$do_tidy" -eq 1 ]; then
  info "go mod tidy"
  go mod tidy
fi

if [ "$do_git" -eq 1 ] && [ ! -d .git ]; then
  info "git init"
  git init -q
  git add -A
  git -c user.email=init@local -c user.name=init commit -qm "Scaffold $binary from cli-blueprint-go"
fi

if [ "$do_verify" -eq 1 ]; then
  info "go build ./..."
  go build ./...
  info "go test ./..."
  go test ./... >/dev/null
fi

cat <<EOF

done: $target

Next:
  cd $target
  \$EDITOR CLAUDE.md README.md        # replace the placeholder descriptions
  \$EDITOR internal/cli/hello.go      # copy it into your first real command, then delete it
  make verify
EOF
