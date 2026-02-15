package hook_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/dansimau/agentstats/internal/db"
	"github.com/dansimau/agentstats/internal/hook"
)

// makeCommit sets up a git repo with one commit and returns its root dir.
func makeCommit(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=Test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init")
	run("config", "user.email", "test@test.com")
	run("config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-m", "init")
	return dir
}

func TestRoundTrip(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer database.Close()

	repoDir := makeCommit(t)

	sessionID := "session-round-trip-001"

	startInput := &hook.HookInput{
		SessionID:  sessionID,
		Cwd:        repoDir,
		PromptText: "Write some code",
		AgentType:  "claude-code",
		EventType:  hook.EventPromptStart,
	}

	if err := hook.RecordPromptStart(database, startInput); err != nil {
		t.Fatalf("RecordPromptStart: %v", err)
	}

	// Verify prompt was inserted with no completed_at.
	var count int
	row := database.QueryRow(
		`SELECT COUNT(*) FROM prompts WHERE session_id=? AND completed_at IS NULL`,
		sessionID,
	)
	if err := row.Scan(&count); err != nil {
		t.Fatalf("query: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 open prompt, got %d", count)
	}

	endInput := &hook.HookInput{
		SessionID: sessionID,
		Cwd:       repoDir,
		AgentType: "claude-code",
		EventType: hook.EventPromptEnd,
	}
	if err := hook.RecordPromptEnd(database, endInput); err != nil {
		t.Fatalf("RecordPromptEnd: %v", err)
	}

	// Verify prompt now has completed_at.
	row = database.QueryRow(
		`SELECT COUNT(*) FROM prompts WHERE session_id=? AND completed_at IS NOT NULL`,
		sessionID,
	)
	if err := row.Scan(&count); err != nil {
		t.Fatalf("query after end: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 completed prompt, got %d", count)
	}
}

func TestPromptEnd_NoOp(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer database.Close()

	// Stop event with no preceding prompt — should be a silent no-op.
	endInput := &hook.HookInput{
		SessionID: "ghost-session",
		Cwd:       t.TempDir(),
		AgentType: "claude-code",
		EventType: hook.EventPromptEnd,
	}
	if err := hook.RecordPromptEnd(database, endInput); err != nil {
		t.Errorf("RecordPromptEnd with no preceding prompt should not error: %v", err)
	}
}
