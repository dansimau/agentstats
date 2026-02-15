#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BIN="$ROOT/bin"

TARGETS=(
  "darwin/arm64"
  "darwin/amd64"
  "linux/amd64"
  "linux/arm64"
)

mkdir -p "$BIN"

for target in "${TARGETS[@]}"; do
  os="${target%%/*}"
  arch="${target##*/}"
  out="$BIN/agentstats-${os}-${arch}"
  echo "Building $out..."
  GOOS="$os" GOARCH="$arch" go build -trimpath -ldflags="-s -w" \
    -o "$out" ./cmd/agentstats
done

chmod +x "$BIN"/agentstats-*
chmod +x "$BIN/agentstats"

echo "Done. Binaries in $BIN/"
