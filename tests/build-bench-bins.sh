#!/bin/bash
# Build bench binaries for hashtabledb mmap vs no-mmap comparison.
#
# Produces:
#   ../bench-bins/mmap/    aergosvr, aergocli, aergoluac (current go.mod, UseMmap)
#   ../bench-bins/nommap/  same tree, older aergo-lib without UseMmap
#   ./benchflood-bin       flood/load tool used by benchmark-hashtabledb-mmap.sh
#
# Usage (from tests/):
#   ./build-bench-bins.sh
#
# Requires: Go toolchain, cmake, make, and a working `make aergosvr` build.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
MMAP_DIR="$ROOT/bench-bins/mmap"
NOMMAP_DIR="$ROOT/bench-bins/nommap"
BENCHFLOOD="$SCRIPT_DIR/benchflood-bin"

# aergo-lib before hashtabledb UseMmap (parent of commit 6012de9e).
AERGO_LIB_NOMMAP="github.com/aergoio/aergo-lib@v1.1.3-0.20260429030355-b6d239f4cb10"

GO_MOD_BACKUP=""
GO_SUM_BACKUP=""

restore_go_mod() {
  if [ -n "$GO_MOD_BACKUP" ] && [ -f "$GO_MOD_BACKUP" ]; then
    mv -f "$GO_MOD_BACKUP" "$ROOT/go.mod"
    if [ -n "$GO_SUM_BACKUP" ] && [ -f "$GO_SUM_BACKUP" ]; then
      mv -f "$GO_SUM_BACKUP" "$ROOT/go.sum"
    fi
    (cd "$ROOT" && go mod download)
    GO_MOD_BACKUP=""
    GO_SUM_BACKUP=""
  fi
}

trap restore_go_mod EXIT

build_aergo_bins() {
  (cd "$ROOT" && make aergosvr aergocli aergoluac)
}

install_bins() {
  local dest=$1
  mkdir -p "$dest"
  cp "$ROOT/bin/aergosvr" "$ROOT/bin/aergocli" "$ROOT/bin/aergoluac" "$dest/"
}

build_benchflood() {
  if [ ! -x "$BENCHFLOOD" ] || [ "$SCRIPT_DIR/benchflood/main.go" -nt "$BENCHFLOOD" ]; then
    echo "building benchflood -> $BENCHFLOOD"
    (cd "$ROOT" && go build -o "$BENCHFLOOD" ./tests/benchflood)
  else
    echo "benchflood up to date: $BENCHFLOOD"
  fi
}

echo "== build mmap binaries (current go.mod) =="
build_aergo_bins
install_bins "$MMAP_DIR"
echo "installed: $MMAP_DIR"

echo ""
echo "== build no-mmap binaries (downgrade aergo-lib) =="
GO_MOD_BACKUP="$(mktemp)"
GO_SUM_BACKUP="$(mktemp)"
cp "$ROOT/go.mod" "$GO_MOD_BACKUP"
cp "$ROOT/go.sum" "$GO_SUM_BACKUP"
(
  cd "$ROOT"
  go get "$AERGO_LIB_NOMMAP"
  go mod tidy
)
build_aergo_bins
install_bins "$NOMMAP_DIR"
restore_go_mod
trap - EXIT
echo "installed: $NOMMAP_DIR"

build_benchflood

echo ""
echo "Done."
echo "  mmap:    $MMAP_DIR"
echo "  nommap:  $NOMMAP_DIR"
echo "  flood:   $BENCHFLOOD"
ls -la "$MMAP_DIR" "$NOMMAP_DIR" "$BENCHFLOOD"
