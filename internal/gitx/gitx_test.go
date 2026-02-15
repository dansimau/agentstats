package gitx_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/dansimau/agentstats/internal/gitx"
)

// initRepo creates a temporary git repo for testing.
func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
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
	return dir
}

func TestIsRepo(t *testing.T) {
	dir := initRepo(t)
	if !gitx.IsRepo(dir) {
		t.Error("IsRepo() should return true for git repo")
	}
	if gitx.IsRepo(t.TempDir()) {
		t.Error("IsRepo() should return false for non-git dir")
	}
}

func TestRepoRoot(t *testing.T) {
	dir := initRepo(t)
	// Create a subdir and check root is still the repo root.
	sub := filepath.Join(dir, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	root := gitx.RepoRoot(sub)
	if root == "" {
		t.Fatal("RepoRoot() returned empty for subdir of git repo")
	}
}

func TestHeadHash_Empty(t *testing.T) {
	// No commits yet → should return "".
	dir := initRepo(t)
	hash := gitx.HeadHash(dir)
	if hash != "" {
		t.Errorf("HeadHash() on empty repo should return '', got %q", hash)
	}
}

func TestHeadHash_WithCommit(t *testing.T) {
	dir := initRepo(t)

	// Create a commit.
	f := filepath.Join(dir, "hello.txt")
	if err := os.WriteFile(f, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
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
	run("add", ".")
	run("commit", "-m", "init")

	hash := gitx.HeadHash(dir)
	if len(hash) != 40 {
		t.Errorf("HeadHash() expected 40-char hash, got %q", hash)
	}
}

func TestGetOriginURL_NoRemote(t *testing.T) {
	dir := initRepo(t)
	url := gitx.GetOriginURL(dir)
	if url != "" {
		t.Errorf("GetOriginURL() with no remote should return '', got %q", url)
	}
}
