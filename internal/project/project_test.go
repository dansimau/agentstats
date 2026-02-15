package project_test

import (
	"path/filepath"
	"testing"

	"github.com/dansimau/agentstats/internal/db"
	"github.com/dansimau/agentstats/internal/project"
)

func openTestDB(t *testing.T) interface{ Close() error } {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	d, err := db.Open(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

func TestUpsertAndFind(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	database, err := db.Open(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer database.Close()

	dir := t.TempDir()

	// First upsert creates the project.
	p1, err := project.Upsert(database, dir)
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}
	if p1.ID == "" {
		t.Error("expected non-empty ID")
	}

	// Second upsert returns the same project.
	p2, err := project.Upsert(database, dir)
	if err != nil {
		t.Fatalf("second Upsert: %v", err)
	}
	if p1.ID != p2.ID {
		t.Errorf("expected same project ID, got %q vs %q", p1.ID, p2.ID)
	}

	// Find returns the same project.
	p3, err := project.Find(database, dir)
	if err != nil {
		t.Fatalf("Find: %v", err)
	}
	if p3 == nil {
		t.Fatal("Find returned nil")
	}
	if p3.ID != p1.ID {
		t.Errorf("Find ID mismatch: %q vs %q", p3.ID, p1.ID)
	}
}

func TestFind_NotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	database, err := db.Open(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer database.Close()

	p, err := project.Find(database, t.TempDir())
	if err != nil {
		t.Fatalf("Find error: %v", err)
	}
	if p != nil {
		t.Error("expected nil for unknown project")
	}
}
