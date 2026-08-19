package mongoapp

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"mongo-developer/internal/model"
)

func (m *Manager) Health(parent context.Context, id string) (model.ConnectionHealth, error) {
	client, err := m.client(id)
	if err != nil {
		return model.ConnectionHealth{Connected: false, CheckedAt: time.Now().Format(time.RFC3339), Error: err.Error()}, err
	}
	ctx, cancel := context.WithTimeout(parent, 7*time.Second)
	defer cancel()
	start := time.Now()
	if err := client.Ping(ctx, readpref.PrimaryPreferred()); err != nil {
		return model.ConnectionHealth{Connected: false, CheckedAt: time.Now().Format(time.RFC3339), Error: err.Error()}, err
	}
	health := model.ConnectionHealth{Connected: true, PingMS: time.Since(start).Milliseconds(), CheckedAt: time.Now().Format(time.RFC3339)}
	var buildInfo bson.M
	if err := client.Database("admin").RunCommand(ctx, bson.D{{Key: "buildInfo", Value: 1}}).Decode(&buildInfo); err == nil {
		health.MongoDBVersion = stringValue(buildInfo["version"])
	}
	return health, nil
}

func (m *Manager) ExplainFind(parent context.Context, id, db, collection, filterJSON, sortJSON string) (model.ExplainResult, error) {
	client, err := m.client(id)
	if err != nil {
		return model.ExplainResult{}, err
	}
	filter, err := ParseDocument(filterJSON)
	if err != nil {
		return model.ExplainResult{}, err
	}
	findCommand := bson.D{{Key: "find", Value: collection}, {Key: "filter", Value: filter}}
	var sortDoc bson.D
	if strings.TrimSpace(sortJSON) != "" && strings.TrimSpace(sortJSON) != "{}" {
		sortDoc, err = ParseDocument(sortJSON)
		if err != nil {
			return model.ExplainResult{}, err
		}
		findCommand = append(findCommand, bson.E{Key: "sort", Value: sortDoc})
	}
	ctx, cancel := opContext(parent)
	defer cancel()
	var raw bson.M
	cmd := bson.D{{Key: "explain", Value: findCommand}, {Key: "verbosity", Value: "executionStats"}}
	if err := client.Database(db).RunCommand(ctx, cmd).Decode(&raw); err != nil {
		return model.ExplainResult{}, err
	}
	return buildExplainResult("find", db, collection, raw, suggestedIndexFromFind(filter, sortDoc)), nil
}

func AggregationWriteStage(pipelineJSON string) (string, error) {
	var pipeline mongo.Pipeline
	if err := bson.UnmarshalExtJSON([]byte(strings.TrimSpace(pipelineJSON)), false, &pipeline); err != nil {
		return "", fmt.Errorf("invalid aggregation pipeline: %w", err)
	}
	for _, stage := range pipeline {
		for _, item := range stage {
			if item.Key == "$out" || item.Key == "$merge" {
				return item.Key, nil
			}
		}
	}
	return "", nil
}

func (m *Manager) ExplainAggregation(parent context.Context, id, db, collection, pipelineJSON string) (model.ExplainResult, error) {
	client, err := m.client(id)
	if err != nil {
		return model.ExplainResult{}, err
	}
	var pipeline mongo.Pipeline
	if err := bson.UnmarshalExtJSON([]byte(strings.TrimSpace(pipelineJSON)), false, &pipeline); err != nil {
		return model.ExplainResult{}, fmt.Errorf("invalid aggregation pipeline: %w", err)
	}
	ctx, cancel := opContext(parent)
	defer cancel()
	var raw bson.M
	aggregateCommand := bson.D{
		{Key: "aggregate", Value: collection},
		{Key: "pipeline", Value: pipeline},
		{Key: "cursor", Value: bson.D{}},
	}
	cmd := bson.D{{Key: "explain", Value: aggregateCommand}, {Key: "verbosity", Value: "executionStats"}}
	if err := client.Database(db).RunCommand(ctx, cmd).Decode(&raw); err != nil {
		return model.ExplainResult{}, err
	}
	return buildExplainResult("aggregation", db, collection, raw, suggestedIndexFromPipeline(pipeline)), nil
}

func buildExplainResult(mode, db, collection string, raw bson.M, suggested string) model.ExplainResult {
	stages := map[string]bool{}
	indexNames := map[string]bool{}
	walkExplain(raw, stages, indexNames)
	stage := "UNKNOWN"
	for _, candidate := range []string{"IXSCAN", "DISTINCT_SCAN", "EXPRESS_IXSCAN", "COLLSCAN"} {
		if stages[candidate] {
			stage = candidate
			break
		}
	}
	if stage == "UNKNOWN" {
		keys := make([]string, 0, len(stages))
		for k := range stages {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		if len(keys) > 0 {
			stage = keys[0]
		}
	}
	indexName := ""
	for name := range indexNames {
		if name != "" {
			indexName = name
			break
		}
	}
	docsExamined := nestedInt64(raw, "executionStats", "totalDocsExamined")
	keysExamined := nestedInt64(raw, "executionStats", "totalKeysExamined")
	returned := nestedInt64(raw, "executionStats", "nReturned")
	executionMS := nestedInt64(raw, "executionStats", "executionTimeMillis")
	usesIndex := stages["IXSCAN"] || stages["DISTINCT_SCAN"] || stages["EXPRESS_IXSCAN"] || indexName != ""
	collectionScan := stages["COLLSCAN"]
	if usesIndex {
		suggested = ""
	}
	summary := "MongoDB returned an execution plan."
	if usesIndex {
		summary = "Indexed plan — MongoDB is using an index for this query."
	} else if collectionScan {
		if suggested != "" {
			summary = "Collection scan — consider the suggested index if this query is frequent and selective."
		} else {
			summary = "Collection scan — review whether an index would help this query."
		}
	}
	return model.ExplainResult{
		Mode: mode, Database: db, Collection: collection, Stage: stage, IndexName: indexName,
		UsesIndex: usesIndex, CollectionScan: collectionScan, DocumentsExamined: docsExamined,
		KeysExamined: keysExamined, Returned: returned, ExecutionMS: executionMS,
		SuggestedIndex: suggested, Summary: summary,
	}
}

func walkExplain(value any, stages, indexes map[string]bool) {
	switch v := value.(type) {
	case bson.M:
		for key, child := range v {
			if key == "stage" {
				if s := stringValue(child); s != "" {
					stages[strings.ToUpper(s)] = true
				}
			}
			if key == "indexName" {
				if s := stringValue(child); s != "" {
					indexes[s] = true
				}
			}
			walkExplain(child, stages, indexes)
		}
	case bson.D:
		for _, item := range v {
			if item.Key == "stage" {
				if s := stringValue(item.Value); s != "" {
					stages[strings.ToUpper(s)] = true
				}
			}
			if item.Key == "indexName" {
				if s := stringValue(item.Value); s != "" {
					indexes[s] = true
				}
			}
			walkExplain(item.Value, stages, indexes)
		}
	case bson.A:
		for _, item := range v {
			walkExplain(item, stages, indexes)
		}
	case []any:
		for _, item := range v {
			walkExplain(item, stages, indexes)
		}
	case map[string]any:
		for key, child := range v {
			if key == "stage" {
				if s := stringValue(child); s != "" {
					stages[strings.ToUpper(s)] = true
				}
			}
			if key == "indexName" {
				if s := stringValue(child); s != "" {
					indexes[s] = true
				}
			}
			walkExplain(child, stages, indexes)
		}
	}
}

func nestedInt64(root bson.M, parentKey, key string) int64 {
	parent, ok := root[parentKey]
	if !ok {
		return 0
	}
	switch p := parent.(type) {
	case bson.M:
		return asInt64(p[key])
	case map[string]any:
		return asInt64(p[key])
	case bson.D:
		for _, item := range p {
			if item.Key == key {
				return asInt64(item.Value)
			}
		}
	}
	return 0
}

func asInt64(v any) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int32:
		return int64(n)
	case int64:
		return n
	case float64:
		return int64(n)
	}
	return 0
}

func stringValue(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func suggestedIndexFromFind(filter, sortDoc bson.D) string {
	fields := make([]bson.E, 0)
	seen := map[string]bool{}
	for _, item := range filter {
		if strings.HasPrefix(item.Key, "$") || seen[item.Key] {
			continue
		}
		fields = append(fields, bson.E{Key: item.Key, Value: 1})
		seen[item.Key] = true
	}
	for _, item := range sortDoc {
		if strings.HasPrefix(item.Key, "$") || seen[item.Key] {
			continue
		}
		direction := int32(1)
		if asInt64(item.Value) < 0 {
			direction = -1
		}
		fields = append(fields, bson.E{Key: item.Key, Value: direction})
		seen[item.Key] = true
	}
	return indexJSON(fields)
}

func suggestedIndexFromPipeline(pipeline mongo.Pipeline) string {
	fields := make([]bson.E, 0)
	seen := map[string]bool{}
	for _, stage := range pipeline {
		for _, item := range stage {
			switch item.Key {
			case "$match":
				if doc, ok := item.Value.(bson.D); ok {
					for _, f := range doc {
						if !strings.HasPrefix(f.Key, "$") && !seen[f.Key] {
							fields = append(fields, bson.E{Key: f.Key, Value: 1})
							seen[f.Key] = true
						}
					}
				}
			case "$sort":
				if doc, ok := item.Value.(bson.D); ok {
					for _, f := range doc {
						if !strings.HasPrefix(f.Key, "$") && !seen[f.Key] {
							direction := int32(1)
							if asInt64(f.Value) < 0 {
								direction = -1
							}
							fields = append(fields, bson.E{Key: f.Key, Value: direction})
							seen[f.Key] = true
						}
					}
				}
			}
		}
	}
	return indexJSON(fields)
}

func indexJSON(fields bson.D) string {
	if len(fields) == 0 {
		return ""
	}
	b, err := bson.MarshalExtJSON(fields, false, false)
	if err != nil {
		return ""
	}
	return string(b)
}
