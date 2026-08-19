package backup

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStoreCreatesTimestampedPathAndListsManifest(t *testing.T) {
	root := t.TempDir()
	store, err := NewStore(root)
	if err != nil {
		t.Fatal(err)
	}
	path, name, err := store.NewPath("database", "dummy_test_data", time.Date(2026, 8, 19, 21, 5, 7, 0, time.Local))
	if err != nil {
		t.Fatal(err)
	}
	if name != "mongo-developer-dummy_test_data-20260819-210507.zip" {
		t.Fatalf("unexpected name: %s", name)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	m := Manifest{FormatVersion: FormatVersion, CreatedAt: "2026-08-19T21:05:07+01:00", Scope: "database", Database: "dummy_test_data", Databases: []DatabaseManifest{{Name: "dummy_test_data", Collections: []CollectionManifest{{Name: "developers", DataFile: "data/a/b.bson", DocumentCount: 42, IndexCount: 2}}}}}
	if err := WriteManifest(zw, m); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	items, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].DocumentCount != 42 || items[0].CollectionCount != 1 {
		t.Fatalf("unexpected items: %#v", items)
	}
}

func TestDeleteRefusesTraversalAndDeletesKnownBackup(t *testing.T) {
	root := t.TempDir()
	store, err := NewStore(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Delete("../anything.zip"); err == nil {
		t.Fatal("expected traversal rejection")
	}
	path := filepath.Join(root, "test.zip")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := store.Delete("test.zip"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("backup was not deleted")
	}
}

func TestNewPathAvoidsTimestampCollision(t *testing.T) {
	root := t.TempDir()
	store, err := NewStore(root)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 19, 21, 5, 7, 0, time.Local)
	first, _, err := store.NewPath("all", "", now)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(first, []byte("placeholder"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, secondName, err := store.NewPath("all", "", now)
	if err != nil {
		t.Fatal(err)
	}
	if secondName != "mongo-developer-all-20260819-210507-2.zip" {
		t.Fatalf("unexpected collision name: %s", secondName)
	}
}

func TestValidateManifestRejectsUnsafeArchivePath(t *testing.T) {
	m := Manifest{FormatVersion: FormatVersion, CreatedAt: "2026-08-19T21:05:07+01:00", Scope: "database", Database: "app", Databases: []DatabaseManifest{{Name: "app", Collections: []CollectionManifest{{Name: "users", Type: "collection", DataFile: "../escape.bson"}}}}}
	if err := ValidateManifest(m); err == nil {
		t.Fatal("expected unsafe path to be rejected")
	}
}
