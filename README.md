# agentstats

Track how much time AI coding agents spend actively working on your projects.

Records prompt timing, session data, and git state into a local SQLite database. Works as a Claude Code plugin (via hooks) and provides CLI commands to inspect the data.

## How it works

**Working time** = sum of `(completed_at - submitted_at)` for completed prompts.
Time between prompts (reading output, thinking, approving plans) is never counted.

## Installation

### As a Claude Code plugin

```
/plugin marketplace add github.com/dansimau/agentstats
```

Pre-compiled binaries for all platforms are included. No install step required.

### Local development

```bash
git clone https://github.com/dansimau/agentstats
cd agentstats
make build
bash scripts/install-local.sh
```

Then follow the instructions printed by `install-local.sh` to add the hooks to `~/.claude/settings.json`.

## CLI Commands

### `agentstats stats [--project <dir>]`

Show AI working time statistics for a project. Defaults to the current directory.

```
Project:               myapp (github.com/user/myapp)
Git origin:            git@github.com:user/myapp.git
Total prompts:         42
Total AI working time: 3h 24m 15s
Average per prompt:    4m 52s
Time period:           2024-01-01 to 2024-02-15
```

### `agentstats history [--project <dir>] [--limit N]`

Show recent prompt history. Defaults to current directory, limit 50.

```
#      Time                 Duration    Prompt
-----  -------------------  ----------  -----------------------------------------------
1      2024-02-15 10:23:01  4m 32s      Create a new Go web server with authentication...
2      2024-02-15 10:27:45  -           Add middleware for rate limiting
```

A `-` duration means the prompt is still in flight.

## Database

Data is stored at `~/.local/share/agentstats/agentstats.db` (XDG-aware).

Override with the `--db` flag on any command.

## Smoke test

```bash
# Record a prompt start
echo '{"session_id":"test-001","cwd":"/tmp","prompt":"hello world","hook_event_name":"UserPromptSubmit","transcript_path":"","permission_mode":"default"}' | \
  ./bin/agentstats hook prompt-start

# Record the end
echo '{"session_id":"test-001","cwd":"/tmp","hook_event_name":"Stop","transcript_path":"","permission_mode":"default"}' | \
  ./bin/agentstats hook prompt-end

# Inspect
./bin/agentstats stats --project /tmp
./bin/agentstats history --project /tmp
```

## Build

```bash
make build        # current platform only
make build-all    # all platforms (darwin/linux × amd64/arm64)
make test         # run tests with race detector
```

## Adding support for other agents

1. Implement `hook.Parser` in `internal/hook/youragent.go`
2. Register it in `ParserForAgent()` in `internal/hook/interface.go`
3. Add hooks for your agent in `hooks/hooks.json`
