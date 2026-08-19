package mongoapp

import (
	"bytes"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestReadBSONDocumentsRoundTrip(t *testing.T) {
	one, err := bson.Marshal(bson.D{{Key: "name", Value: "Chris"}, {Key: "active", Value: true}})
	if err != nil {
		t.Fatal(err)
	}
	two, err := bson.Marshal(bson.D{{Key: "name", Value: "Lexi"}, {Key: "score", Value: int32(42)}})
	if err != nil {
		t.Fatal(err)
	}
	stream := append(append([]byte{}, one...), two...)
	var names []string
	count, err := readBSONDocuments(bytes.NewReader(stream), func(doc bson.D) error {
		for _, e := range doc {
			if e.Key == "name" {
				names = append(names, e.Value.(string))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 || len(names) != 2 || names[0] != "Chris" || names[1] != "Lexi" {
		t.Fatalf("unexpected restore stream result: count=%d names=%v", count, names)
	}
}

func TestReadBSONDocumentsRejectsTruncatedStream(t *testing.T) {
	_, err := readBSONDocuments(bytes.NewReader([]byte{10, 0, 0, 0, 1, 2}), func(bson.D) error { return nil })
	if err == nil {
		t.Fatal("expected truncated BSON stream to fail")
	}
}

func TestSanitizeIndexSpec(t *testing.T) {
	if _, ok := sanitizeIndexSpec(bson.D{{Key: "v", Value: int32(2)}, {Key: "key", Value: bson.D{{Key: "_id", Value: int32(1)}}}, {Key: "name", Value: "_id_"}}); ok {
		t.Fatal("_id_ index must not be recreated")
	}
	spec, ok := sanitizeIndexSpec(bson.D{{Key: "v", Value: int32(2)}, {Key: "key", Value: bson.D{{Key: "email", Value: int32(1)}}}, {Key: "name", Value: "email_1"}, {Key: "unique", Value: true}, {Key: "ns", Value: "app.users"}})
	if !ok {
		t.Fatal("expected custom index to be retained")
	}
	for _, e := range spec {
		if e.Key == "v" || e.Key == "ns" {
			t.Fatalf("server-only index field leaked into createIndexes spec: %s", e.Key)
		}
	}
}
