package db_test

import (
	"path/filepath"
	"testing"

	"github.com/dansimau/agentstats/internal/db"
)

func TestOpenCreatesDB(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "agentstats.db")
	database, err := db.Open(path)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer database.Close()

	// Verify schema is idempotent — second open should succeed.
	database.Close()
	database2, err := db.Open(path)
	if err != nil {
		t.Fatalf("second Open() error: %v", err)
	}
	database2.Close()
}

func TestDefaultPath(t *testing.T) {
	p := db.DefaultPath()
	if p == "" {
		t.Fatal("DefaultPath() returned empty string")
	}
	if filepath.Base(p) != "agentstats.db" {
		t.Errorf("expected agentstats.db, got %s", filepath.Base(p))
	}
}
