#!/usr/bin/env bash
# Dev: build for current platform and configure Claude Code hooks.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BIN="$ROOT/bin"

# Detect current platform.
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')"
PLATFORM="${OS}-${ARCH}"
BINARY="$BIN/agentstats-${PLATFORM}"

echo "Building agentstats for ${PLATFORM}..."
mkdir -p "$BIN"
go build -o "$BINARY" "$ROOT/cmd/agentstats"
chmod +x "$BINARY"
chmod +x "$BIN/agentstats"

echo "Binary: $BINARY"
echo ""
echo "To enable Claude Code hooks, add the following to ~/.claude/settings.json"
echo "(merge into the existing 'hooks' object if present):"
echo ""
cat <<EOF
{
  "hooks": {
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "${BINARY} hook prompt-start --agent claude-code",
            "async": true
          }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "${BINARY} hook prompt-end --agent claude-code",
            "async": true
          }
        ]
      }
    ]
  }
}
EOF
