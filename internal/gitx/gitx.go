package gitx

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
)

// IsRepo reports whether dir is inside a git repository.
func IsRepo(dir string) bool {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

// RepoRoot returns the canonical top-level directory of the git repo
// containing dir. Returns "" if dir is not in a git repo.
func RepoRoot(dir string) string {
	out, err := run(dir, "git", "rev-parse", "--show-toplevel")
	if err != nil {
		return ""
	}
	root, err := filepath.EvalSymlinks(out)
	if err != nil {
		return out
	}
	return root
}

// HeadHash returns the current HEAD commit hash, or "" if there are no
// commits yet or dir is not a git repo.
func HeadHash(dir string) string {
	out, err := run(dir, "git", "rev-parse", "HEAD")
	if err != nil {
		return ""
	}
	return out
}

// GetOriginURL returns the URL of the "origin" remote, or "" if none exists.
func GetOriginURL(dir string) string {
	out, err := run(dir, "git", "remote", "get-url", "origin")
	if err != nil {
		return ""
	}
	return out
}

func run(dir string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return strings.TrimSpace(stdout.String()), nil
}
