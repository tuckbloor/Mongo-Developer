package mongoapp

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestBuildExplainResultFindsCollectionScan(t *testing.T) {
	raw := bson.M{
		"queryPlanner":   bson.M{"winningPlan": bson.M{"stage": "COLLSCAN"}},
		"executionStats": bson.M{"totalDocsExamined": int64(1200), "totalKeysExamined": int64(0), "nReturned": int64(12), "executionTimeMillis": int64(7)},
	}
	got := buildExplainResult("find", "demo", "users", raw, `{"email":1}`)
	if !got.CollectionScan || got.UsesIndex || got.DocumentsExamined != 1200 || got.SuggestedIndex == "" {
		t.Fatalf("unexpected explain result: %#v", got)
	}
}

func TestBuildExplainResultFindsIndex(t *testing.T) {
	raw := bson.M{
		"queryPlanner":   bson.M{"winningPlan": bson.M{"inputStage": bson.M{"stage": "IXSCAN", "indexName": "email_1"}}},
		"executionStats": bson.M{"totalDocsExamined": int64(12), "totalKeysExamined": int64(12), "nReturned": int64(12)},
	}
	got := buildExplainResult("find", "demo", "users", raw, `{"email":1}`)
	if !got.UsesIndex || got.CollectionScan || got.IndexName != "email_1" || got.SuggestedIndex != "" {
		t.Fatalf("unexpected explain result: %#v", got)
	}
}

func TestSuggestedIndexFromFind(t *testing.T) {
	filter := bson.D{{Key: "active", Value: true}, {Key: "department_id", Value: 4}}
	sortDoc := bson.D{{Key: "name", Value: int32(-1)}}
	got := suggestedIndexFromFind(filter, sortDoc)
	if got != `{"active":1,"department_id":1,"name":-1}` {
		t.Fatalf("unexpected suggested index: %s", got)
	}
}

func TestAggregationWriteStage(t *testing.T) {
	stage, err := AggregationWriteStage(`[ { "$match": {"active": true} }, { "$out": "archive" } ]`)
	if err != nil || stage != "$out" {
		t.Fatalf("expected $out write stage, got %q err=%v", stage, err)
	}
	stage, err = AggregationWriteStage(`[ { "$match": {"active": true} } ]`)
	if err != nil || stage != "" {
		t.Fatalf("expected read-only pipeline, got %q err=%v", stage, err)
	}
}
