package store

import (
	"strings"
	"testing"

	"mongo-developer/internal/model"
)

func TestSaveStripsPasswordFromURI(t *testing.T) {
	s, err := NewConnectionStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	p, err := s.Save(model.ConnectionInput{Name: "local", URI: "mongodb://alice:secret@localhost:27017/?authSource=admin"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(p.URI, "secret") {
		t.Fatalf("stored URI leaked password: %s", p.URI)
	}
	runtimeURI, err := s.RuntimeURI(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(runtimeURI, "secret") {
		t.Fatalf("runtime URI did not restore password: %s", runtimeURI)
	}
}

func TestReadOnlyPersists(t *testing.T) {
	s, err := NewConnectionStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	p, err := s.Save(model.ConnectionInput{Name: "prod", URI: "mongodb://localhost:27017", ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !got.ReadOnly {
		t.Fatal("read-only flag was not persisted")
	}
}
