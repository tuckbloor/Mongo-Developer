package mongoapp

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"mongo-developer/internal/model"
)

type Manager struct {
	mu      sync.RWMutex
	clients map[string]*mongo.Client
}

func New() *Manager { return &Manager{clients: map[string]*mongo.Client{}} }

func (m *Manager) Connect(parent context.Context, id, uri string) error {
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()
	opts := options.Client().ApplyURI(uri).
		SetAppName("Mongo Developer").
		SetConnectTimeout(7 * time.Second).
		SetServerSelectionTimeout(7 * time.Second)
	client, err := mongo.Connect(opts)
	if err != nil {
		return err
	}
	if err := client.Ping(ctx, readpref.PrimaryPreferred()); err != nil {
		_ = client.Disconnect(context.Background())
		return err
	}
	m.mu.Lock()
	old := m.clients[id]
	m.clients[id] = client
	m.mu.Unlock()
	if old != nil {
		_ = old.Disconnect(context.Background())
	}
	return nil
}

func (m *Manager) Test(parent context.Context, uri string) error {
	id := "__test__"
	if err := m.Connect(parent, id, uri); err != nil {
		return err
	}
	return m.Disconnect(id)
}

func (m *Manager) Disconnect(id string) error {
	m.mu.Lock()
	c := m.clients[id]
	delete(m.clients, id)
	m.mu.Unlock()
	if c == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.Disconnect(ctx)
}

func (m *Manager) CloseAll() {
	m.mu.Lock()
	clients := m.clients
	m.clients = map[string]*mongo.Client{}
	m.mu.Unlock()
	for _, c := range clients {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_ = c.Disconnect(ctx)
		cancel()
	}
}

func (m *Manager) client(id string) (*mongo.Client, error) {
	m.mu.RLock()
	c := m.clients[id]
	m.mu.RUnlock()
	if c == nil {
		return nil, fmt.Errorf("connection is not active")
	}
	return c, nil
}

func opContext(parent context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, 30*time.Second)
}

func (m *Manager) ListDatabases(parent context.Context, id string) ([]model.DatabaseInfo, error) {
	c, err := m.client(id)
	if err != nil {
		return nil, err
	}
	ctx, cancel := opContext(parent)
	defer cancel()
	names, err := c.ListDatabaseNames(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	out := make([]model.DatabaseInfo, 0, len(names))
	for _, n := range names {
		out = append(out, model.DatabaseInfo{Name: n})
	}
	return out, nil
}

func (m *Manager) ListCollections(parent context.Context, id, db string) ([]model.CollectionInfo, error) {
	c, err := m.client(id)
	if err != nil {
		return nil, err
	}
	ctx, cancel := opContext(parent)
	defer cancel()
	names, err := c.Database(db).ListCollectionNames(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	out := make([]model.CollectionInfo, 0, len(names))
	for _, n := range names {
		out = append(out, model.CollectionInfo{Name: n, Type: "collection"})
	}
	return out, nil
}

func (m *Manager) CreateDatabase(parent context.Context, id, db, initialCollection string) error {
	if strings.TrimSpace(db) == "" {
		return fmt.Errorf("database name is required")
	}
	if strings.TrimSpace(initialCollection) == "" {
		initialCollection = "initial"
	}
	return m.CreateCollection(parent, id, db, initialCollection)
}

func (m *Manager) DropDatabase(parent context.Context, id, db string) error {
	c, err := m.client(id)
	if err != nil {
		return err
	}
	ctx, cancel := opContext(parent)
	defer cancel()
	return c.Database(db).Drop(ctx)
}

func (m *Manager) CreateCollection(parent context.Context, id, db, collection string) error {
	c, err := m.client(id)
	if err != nil {
		return err
	}
	if strings.TrimSpace(collection) == "" {
		return fmt.Errorf("collection name is required")
	}
	ctx, cancel := opContext(parent)
	defer cancel()
	return c.Database(db).CreateCollection(ctx, collection)
}

func (m *Manager) DropCollection(parent context.Context, id, db, collection string) error {
	c, err := m.client(id)
	if err != nil {
		return err
	}
	ctx, cancel := opContext(parent)
	defer cancel()
	return c.Database(db).Collection(collection).Drop(ctx)
}

func ParseDocument(input string) (bson.D, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		input = "{}"
	}
	var doc bson.D
	if err := bson.UnmarshalExtJSON([]byte(input), false, &doc); err != nil {
		return nil, fmt.Errorf("invalid JSON/EJSON document: %w", err)
	}
	return doc, nil
}

func (m *Manager) Find(parent context.Context, req model.FindRequest) (model.FindResult, error) {
	c, err := m.client(req.ConnectionID)
	if err != nil {
		return model.FindResult{}, err
	}
	filter, err := ParseDocument(req.FilterJSON)
	if err != nil {
		return model.FindResult{}, err
	}
	if req.Limit <= 0 || req.Limit > 500 {
		req.Limit = 50
	}
	if req.Skip < 0 {
		req.Skip = 0
	}
	findOpts := options.Find().SetLimit(req.Limit).SetSkip(req.Skip)
	if strings.TrimSpace(req.SortJSON) != "" && strings.TrimSpace(req.SortJSON) != "{}" {
		sortDoc, err := ParseDocument(req.SortJSON)
		if err != nil {
			return model.FindResult{}, err
		}
		findOpts.SetSort(sortDoc)
	}
	ctx, cancel := opContext(parent)
	defer cancel()
	coll := c.Database(req.Database).Collection(req.Collection)
	start := time.Now()
	cursor, err := coll.Find(ctx, filter, findOpts)
	if err != nil {
		return model.FindResult{}, err
	}
	defer cursor.Close(ctx)
	docs := []string{}
	for cursor.Next(ctx) {
		b, err := bson.MarshalExtJSON(cursor.Current, true, false)
		if err != nil {
			return model.FindResult{}, err
		}
		docs = append(docs, string(b))
	}
	if err := cursor.Err(); err != nil {
		return model.FindResult{}, err
	}
	total, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return model.FindResult{}, err
	}
	return model.FindResult{Documents: docs, Total: total, QueryMS: time.Since(start).Milliseconds()}, nil
}

func (m *Manager) Insert(parent context.Context, id, db, collection, documentJSON string) (string, error) {
	c, err := m.client(id)
	if err != nil {
		return "", err
	}
	doc, err := ParseDocument(documentJSON)
	if err != nil {
		return "", err
	}
	ctx, cancel := opContext(parent)
	defer cancel()
	res, err := c.Database(db).Collection(collection).InsertOne(ctx, doc)
	if err != nil {
		return "", err
	}
	b, err := bson.MarshalExtJSON(bson.M{"_id": res.InsertedID}, true, false)
	return string(b), err
}

func (m *Manager) Replace(parent context.Context, id, db, collection, filterJSON, documentJSON string) (int64, error) {
	c, err := m.client(id)
	if err != nil {
		return 0, err
	}
	filter, err := ParseDocument(filterJSON)
	if err != nil {
		return 0, err
	}
	doc, err := ParseDocument(documentJSON)
	if err != nil {
		return 0, err
	}
	ctx, cancel := opContext(parent)
	defer cancel()
	res, err := c.Database(db).Collection(collection).ReplaceOne(ctx, filter, doc)
	if err != nil {
		return 0, err
	}
	return res.ModifiedCount, nil
}

func (m *Manager) Delete(parent context.Context, id, db, collection, filterJSON string) (int64, error) {
	c, err := m.client(id)
	if err != nil {
		return 0, err
	}
	filter, err := ParseDocument(filterJSON)
	if err != nil {
		return 0, err
	}
	ctx, cancel := opContext(parent)
	defer cancel()
	res, err := c.Database(db).Collection(collection).DeleteOne(ctx, filter)
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}

func (m *Manager) Stats(parent context.Context, id, db, collection string) (model.CollectionStats, error) {
	c, err := m.client(id)
	if err != nil {
		return model.CollectionStats{}, err
	}
	ctx, cancel := opContext(parent)
	defer cancel()
	coll := c.Database(db).Collection(collection)
	count, err := coll.EstimatedDocumentCount(ctx)
	if err != nil {
		return model.CollectionStats{}, err
	}
	stats := model.CollectionStats{Documents: count}
	idxCursor, err := coll.Indexes().List(ctx)
	if err == nil {
		for idxCursor.Next(ctx) {
			stats.Indexes++
		}
		_ = idxCursor.Close(ctx)
	}
	var raw bson.M
	if err := c.Database(db).RunCommand(ctx, bson.D{{Key: "collStats", Value: collection}}).Decode(&raw); err == nil {
		size := asFloat(raw["size"])
		storage := asFloat(raw["storageSize"])
		avg := asFloat(raw["avgObjSize"])
		stats.TotalSizeMB = size / 1024 / 1024
		stats.StorageSizeMB = storage / 1024 / 1024
		stats.AvgDocumentKB = avg / 1024
	}
	return stats, nil
}

func asFloat(v any) float64 {
	switch n := v.(type) {
	case int32:
		return float64(n)
	case int64:
		return float64(n)
	case float64:
		return n
	case float32:
		return float64(n)
	default:
		return 0
	}
}

func (m *Manager) ListIndexes(parent context.Context, id, db, collection string) ([]string, error) {
	c, err := m.client(id)
	if err != nil {
		return nil, err
	}
	ctx, cancel := opContext(parent)
	defer cancel()
	cursor, err := c.Database(db).Collection(collection).Indexes().List(ctx)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	out := []string{}
	for cursor.Next(ctx) {
		b, err := bson.MarshalExtJSON(cursor.Current, true, false)
		if err != nil {
			return nil, err
		}
		out = append(out, string(b))
	}
	return out, cursor.Err()
}

func (m *Manager) CreateIndex(parent context.Context, id, db, collection, keysJSON, name string, unique bool) (string, error) {
	c, err := m.client(id)
	if err != nil {
		return "", err
	}
	keys, err := ParseDocument(keysJSON)
	if err != nil {
		return "", err
	}
	opts := options.Index().SetUnique(unique)
	if strings.TrimSpace(name) != "" {
		opts.SetName(strings.TrimSpace(name))
	}
	ctx, cancel := opContext(parent)
	defer cancel()
	return c.Database(db).Collection(collection).Indexes().CreateOne(ctx, mongo.IndexModel{Keys: keys, Options: opts})
}

func (m *Manager) Aggregate(parent context.Context, id, db, collection, pipelineJSON string, limit int64) ([]string, error) {
	c, err := m.client(id)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	var pipeline mongo.Pipeline
	if err := bson.UnmarshalExtJSON([]byte(strings.TrimSpace(pipelineJSON)), false, &pipeline); err != nil {
		return nil, fmt.Errorf("invalid aggregation pipeline: %w", err)
	}
	ctx, cancel := opContext(parent)
	defer cancel()
	cursor, err := c.Database(db).Collection(collection).Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	out := []string{}
	for cursor.Next(ctx) && int64(len(out)) < limit {
		b, err := bson.MarshalExtJSON(cursor.Current, true, false)
		if err != nil {
			return nil, err
		}
		out = append(out, string(b))
	}
	return out, cursor.Err()
}

func (m *Manager) ExportJSON(parent context.Context, id, db, collection, path string) error {
	c, err := m.client(id)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(parent, 10*time.Minute)
	defer cancel()
	cursor, err := c.Database(db).Collection(collection).Find(ctx, bson.D{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	if _, err := w.WriteString("[\n"); err != nil {
		return err
	}
	first := true
	for cursor.Next(ctx) {
		b, err := bson.MarshalExtJSON(cursor.Current, true, true)
		if err != nil {
			return err
		}
		if !first {
			if _, err := w.WriteString(",\n"); err != nil {
				return err
			}
		}
		first = false
		if _, err := w.WriteString("  " + string(b)); err != nil {
			return err
		}
	}
	if err := cursor.Err(); err != nil {
		return err
	}
	_, err = w.WriteString("\n]\n")
	return err
}

func (m *Manager) ImportJSON(parent context.Context, id, db, collection, path string) (int, error) {
	c, err := m.client(id)
	if err != nil {
		return 0, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	const maxImportBytes int64 = 100 * 1024 * 1024
	if info.Size() > maxImportBytes {
		return 0, fmt.Errorf("import file is %.1f MB; V1 limits JSON imports to 100 MB", float64(info.Size())/1024/1024)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	var docs []bson.D
	if err := bson.UnmarshalExtJSON(data, false, &docs); err != nil {
		return 0, fmt.Errorf("import file must be a JSON/EJSON array of documents: %w", err)
	}
	if len(docs) == 0 {
		return 0, nil
	}
	values := make([]any, len(docs))
	for i := range docs {
		values[i] = docs[i]
	}
	ctx, cancel := context.WithTimeout(parent, 10*time.Minute)
	defer cancel()
	res, err := c.Database(db).Collection(collection).InsertMany(ctx, values)
	if err != nil {
		return 0, err
	}
	return len(res.InsertedIDs), nil
}
