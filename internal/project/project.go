package project

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dansimau/agentstats/internal/gitx"
	"github.com/google/uuid"
)

// Project represents a tracked project.
type Project struct {
	ID        string
	GitOrigin string // empty if no remote
	Directory string
}

// Resolve returns the canonical project directory and git origin for a
// given working directory.
func Resolve(cwd string) (dir string, origin string) {
	abs, err := filepath.Abs(cwd)
	if err != nil {
		abs = cwd
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		resolved = abs
	}

	if gitx.IsRepo(resolved) {
		root := gitx.RepoRoot(resolved)
		if root != "" {
			resolved = root
		}
		origin = gitx.GetOriginURL(resolved)
	}
	return resolved, origin
}

// Upsert finds or creates a project for the given directory and optional
// git origin. Origin-first matching handles re-clones to new paths.
func Upsert(db *sql.DB, cwd string) (*Project, error) {
	dir, origin := Resolve(cwd)

	// Try match by git_origin first (handles re-clones).
	if origin != "" {
		p, err := findByOrigin(db, origin)
		if err != nil {
			return nil, err
		}
		if p != nil {
			// Update directory if it has changed.
			if p.Directory != dir {
				if _, err := db.Exec(
					`UPDATE projects SET directory=? WHERE id=?`, dir, p.ID,
				); err != nil {
					return nil, fmt.Errorf("update project dir: %w", err)
				}
				p.Directory = dir
			}
			return p, nil
		}
	}

	// Try match by directory.
	p, err := findByDir(db, dir)
	if err != nil {
		return nil, err
	}
	if p != nil {
		return p, nil
	}

	// Create new project.
	id := uuid.New().String()
	var originVal interface{}
	if origin != "" {
		originVal = origin
	}
	if _, err := db.Exec(
		`INSERT INTO projects (id, git_origin, directory) VALUES (?, ?, ?)`,
		id, originVal, dir,
	); err != nil {
		return nil, fmt.Errorf("insert project: %w", err)
	}
	return &Project{ID: id, GitOrigin: origin, Directory: dir}, nil
}

// Find looks up a project for the given directory (read-only).
func Find(db *sql.DB, cwd string) (*Project, error) {
	dir, origin := Resolve(cwd)

	if origin != "" {
		p, err := findByOrigin(db, origin)
		if err != nil {
			return nil, err
		}
		if p != nil {
			return p, nil
		}
	}
	return findByDir(db, dir)
}

// FindByID looks up a project by its ID.
func FindByID(db *sql.DB, id string) (*Project, error) {
	row := db.QueryRow(`SELECT id, COALESCE(git_origin,''), directory FROM projects WHERE id=?`, id)
	p := &Project{}
	if err := row.Scan(&p.ID, &p.GitOrigin, &p.Directory); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return p, nil
}

func findByOrigin(db *sql.DB, origin string) (*Project, error) {
	row := db.QueryRow(
		`SELECT id, COALESCE(git_origin,''), directory FROM projects WHERE git_origin=?`,
		origin,
	)
	p := &Project{}
	if err := row.Scan(&p.ID, &p.GitOrigin, &p.Directory); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query by origin: %w", err)
	}
	return p, nil
}

func findByDir(db *sql.DB, dir string) (*Project, error) {
	row := db.QueryRow(
		`SELECT id, COALESCE(git_origin,''), directory FROM projects WHERE directory=?`,
		dir,
	)
	p := &Project{}
	if err := row.Scan(&p.ID, &p.GitOrigin, &p.Directory); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query by dir: %w", err)
	}
	return p, nil
}

// ShortName returns a human-readable name for the project.
func (p *Project) ShortName() string {
	return filepath.Base(p.Directory)
}

// DisplayOrigin returns a cleaned-up display version of the git origin.
func (p *Project) DisplayOrigin() string {
	if p.GitOrigin == "" {
		return ""
	}
	// Strip common git URL noise for display.
	s := p.GitOrigin
	// ssh: git@github.com:user/repo.git -> github.com/user/repo
	if len(s) > 4 && s[:4] == "git@" {
		s = s[4:]
		s = replaceFirst(s, ":", "/")
	}
	// https://github.com/user/repo.git -> github.com/user/repo
	for _, prefix := range []string{"https://", "http://"} {
		if len(s) > len(prefix) && s[:len(prefix)] == prefix {
			s = s[len(prefix):]
		}
	}
	// Strip trailing .git
	if len(s) > 4 && s[len(s)-4:] == ".git" {
		s = s[:len(s)-4]
	}
	return s
}

func replaceFirst(s, old, new string) string {
	i := indexOf(s, old)
	if i < 0 {
		return s
	}
	return s[:i] + new + s[i+len(old):]
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// cwdExists is used in tests; exported for test packages.
func init() {
	_ = os.Getenv // keep os import used
}
