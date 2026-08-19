package mongoapp

import "testing"

func TestParseDocument(t *testing.T) {
	if _, err := ParseDocument(`{"position":"ST","rating":{"$gte":80}}`); err != nil {
		t.Fatal(err)
	}
}

func TestParseDocumentRejectsArray(t *testing.T) {
	if _, err := ParseDocument(`[1,2,3]`); err == nil {
		t.Fatal("expected document parser to reject array")
	}
}
