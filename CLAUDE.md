# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`agentstats` is a Go CLI tool and Claude Code plugin that tracks how much active working time AI coding agents spend on your projects. It hooks into Claude Code's `UserPromptSubmit` and `Stop` events, recording timing data into a local SQLite database. Working time = sum of `(completed_at - submitted_at)` per prompt; idle time between prompts is never counted.

## Commands

```bash
make build        # Build for current platform → bin/agentstats-[platform]
make build-all    # Build for all platforms (darwin/linux × amd64/arm64)
make install      # Build and install to $GOPATH/bin
make test         # Run all tests with race detector (go test -race ./...)
make clean        # Remove binaries
```

Run a single test package:
```bash
go test -race ./internal/hook/...
```

For local dev setup (builds binary and prints hook config instructions):
```bash
bash scripts/install-local.sh
```

## Architecture

### Data Flow

1. **`UserPromptSubmit`** → `agentstats hook prompt-start`
   - Parses Claude Code JSON payload from stdin (session_id, cwd, prompt text)
   - Detects git repo root and HEAD hash
   - Creates/updates `projects` record (origin-first matching)
   - Inserts `prompts` row with `submitted_at`

2. **`Stop`** → `agentstats hook prompt-end`
   - Marks the most recent open prompt as `completed_at`
   - Captures ending git HEAD hash

### Package Structure

| Package | Purpose |
|---|---|
| `internal/hook` | Hook command handlers, Claude payload parser, recording logic |
| `internal/db` | SQLite init, schema, XDG path resolution |
| `internal/project` | Project lifecycle; origin-first dedup, git root detection |
| `internal/gitx` | Git utilities (repo detection, HEAD hash, remote URL) |
| `internal/cli` | `stats` and `history` user-facing commands |
| `cmd/agentstats` | Cobra root command entry point |

### Database Schema

Three tables: `projects` → `sessions` → `prompts`. Key fields on `prompts`: `submitted_at`, `completed_at`, `git_hash_start`, `git_hash_end`. Duration is computed in SQL via `julianday()` arithmetic. DB stored at `~/.local/share/agentstats/agentstats.db` (XDG-aware).

### Extensibility

`internal/hook/interface.go` defines a `Parser` interface for adding support for agents other than Claude Code. Register new parsers in `ParserForAgent()`.

## Key Design Decisions

- **Origin-first project matching**: projects are matched by `git_origin` URL first, then directory path. This ensures re-clones of the same repo are tracked as one project.
- **Async hooks**: both hooks run with `"async": true` so they don't block Claude Code's UI.
- **WAL mode + foreign keys**: SQLite is configured with WAL journaling, foreign keys enabled, and a 5s busy timeout.
