package mongoapp

import "testing"

func TestDummyCollectionsAreRichAndUnique(t *testing.T) {
	seeds := dummyCollections()
	if len(seeds) < 15 {
		t.Fatalf("expected at least 15 practice collections, got %d", len(seeds))
	}
	seen := map[string]bool{}
	docs := 0
	indexes := 0
	for _, seed := range seeds {
		if seed.Name == "" {
			t.Fatal("practice collection has an empty name")
		}
		if seen[seed.Name] {
			t.Fatalf("duplicate practice collection %q", seed.Name)
		}
		seen[seed.Name] = true
		if len(seed.Docs) == 0 {
			t.Fatalf("practice collection %q has no documents", seed.Name)
		}
		docs += len(seed.Docs)
		indexes += len(seed.Indexes)
	}
	if docs < 300 {
		t.Fatalf("expected at least 300 practice documents, got %d", docs)
	}
	if indexes < 20 {
		t.Fatalf("expected at least 20 custom indexes, got %d", indexes)
	}
}

func TestDummyDatabaseNameIsStable(t *testing.T) {
	if DummyDatabaseName != "dummy_test_data" {
		t.Fatalf("dummy database name changed to %q", DummyDatabaseName)
	}
}
