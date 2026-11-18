package cursor

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"
)

func testStore(t *testing.T) *PositionStore {
	t.Helper()
	dir := t.TempDir()
	return &PositionStore{cacheDir: dir}
}

func TestSetAndGetPosition(t *testing.T) {
	ps := testStore(t)

	cfg := FileConfig{CursorLine: 5, CursorCol: 10}
	if err := ps.SetPosition("/path/to/file.md", cfg); err != nil {
		t.Fatalf("SetPosition() error = %v", err)
	}

	got, ok := ps.GetPosition("/path/to/file.md")
	if !ok {
		t.Fatal("GetPosition() returned false")
	}
	if got.CursorLine != 5 || got.CursorCol != 10 {
		t.Errorf("got %+v, want {CursorLine:5 CursorCol:10}", got)
	}
}

func TestGetPositionNotFound(t *testing.T) {
	ps := testStore(t)
	_, ok := ps.GetPosition("/nonexistent.md")
	if ok {
		t.Error("GetPosition() should return false for nonexistent file")
	}
}

func TestRemovePosition(t *testing.T) {
	ps := testStore(t)

	cfg := FileConfig{CursorLine: 1, CursorCol: 0}
	if err := ps.SetPosition("/file.md", cfg); err != nil {
		t.Fatal(err)
	}

	if err := ps.RemovePosition("/file.md"); err != nil {
		t.Fatalf("RemovePosition() error = %v", err)
	}

	_, ok := ps.GetPosition("/file.md")
	if ok {
		t.Error("GetPosition() should return false after remove")
	}
}

func TestConfigPath(t *testing.T) {
	ps := &PositionStore{cacheDir: "/tmp/test"}
	got := ps.configPath("/path/to/file.md")
	if filepath.Dir(got) != "/tmp/test" {
		t.Errorf("configPath() dir = %q, want /tmp/test", filepath.Dir(got))
	}
	if filepath.Ext(got) != ".json" {
		t.Errorf("configPath() ext = %q, want .json", filepath.Ext(got))
	}
	if len(filepath.Base(got)) != sha256.Size*2+len(".json") {
		t.Errorf("configPath() base = %q, want %d hex chars", filepath.Base(got), sha256.Size*2)
	}
}

func TestKeyCollisionFree(t *testing.T) {
	if key("/tmp/a/b.md") == key("/_tmp/a/b.md") {
		t.Error("distinct paths must not share a key")
	}
	if key("a/../b.md") != key("b.md") {
		t.Error("cleaned paths must share a key")
	}
}

func TestSetPositionAtomicWrite(t *testing.T) {
	ps := testStore(t)
	if err := ps.SetPosition("/atomic.md", FileConfig{CursorLine: 1, CursorCol: 2}); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(ps.cacheDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("cache dir has %d entries, want 1", len(entries))
	}
}

func TestGetPositionCorruptJSON(t *testing.T) {
	ps := testStore(t)

	// Write invalid JSON
	path := ps.configPath("/bad.json")
	os.WriteFile(path, []byte("not json"), 0644)

	_, ok := ps.GetPosition("/bad.json")
	if ok {
		t.Error("GetPosition() should return false for corrupt JSON")
	}
}

func TestNewPositionStore(t *testing.T) {
	ps, err := NewPositionStore()
	if err != nil {
		t.Fatalf("NewPositionStore() error = %v", err)
	}
	if ps == nil {
		t.Fatal("NewPositionStore() returned nil")
	}
}
