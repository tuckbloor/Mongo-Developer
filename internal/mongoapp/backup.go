package mongoapp

import (
	"archive/zip"
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"mongo-developer/internal/backup"
	"mongo-developer/internal/model"
)

var systemDatabases = map[string]bool{"admin": true, "config": true, "local": true}

func (m *Manager) CreateBackup(parent context.Context, id, sourceName, scope, database, path string) (model.BackupRecord, error) {
	client, err := m.client(id)
	if err != nil {
		return model.BackupRecord{}, err
	}
	if scope != "all" && scope != "database" {
		return model.BackupRecord{}, fmt.Errorf("backup scope must be all or database")
	}
	if scope == "database" {
		database = strings.TrimSpace(database)
		if database == "" {
			return model.BackupRecord{}, fmt.Errorf("choose a database to back up")
		}
		if systemDatabases[database] {
			return model.BackupRecord{}, fmt.Errorf("MongoDB system database %q is not included in application backups", database)
		}
	}

	ctx, cancel := context.WithTimeout(parent, 30*time.Minute)
	defer cancel()

	dbNames := []string{database}
	if scope == "all" {
		dbNames, err = client.ListDatabaseNames(ctx, bson.D{})
		if err != nil {
			return model.BackupRecord{}, err
		}
		filtered := dbNames[:0]
		for _, name := range dbNames {
			if !systemDatabases[name] {
				filtered = append(filtered, name)
			}
		}
		dbNames = filtered
	}
	if len(dbNames) == 0 {
		return model.BackupRecord{}, fmt.Errorf("there are no user databases to back up")
	}
	sort.Strings(dbNames)

	tempPath := path + ".partial"
	_ = os.Remove(tempPath)
	out, err := os.Create(tempPath)
	if err != nil {
		return model.BackupRecord{}, err
	}
	zw := zip.NewWriter(out)
	ok := false
	defer func() {
		if !ok {
			_ = zw.Close()
			_ = out.Close()
			_ = os.Remove(tempPath)
		}
	}()

	manifest := backup.Manifest{
		FormatVersion: backup.FormatVersion,
		CreatedAt:     time.Now().Format(time.RFC3339),
		Scope:         scope,
		Database:      database,
		SourceName:    strings.TrimSpace(sourceName),
	}

	for _, dbName := range dbNames {
		specs, err := client.Database(dbName).ListCollectionSpecifications(ctx, bson.D{})
		if err != nil {
			return model.BackupRecord{}, fmt.Errorf("list collections in %s: %w", dbName, err)
		}
		sort.Slice(specs, func(i, j int) bool {
			if specs[i].Type != specs[j].Type {
				return specs[i].Type < specs[j].Type
			}
			return specs[i].Name < specs[j].Name
		})
		dbManifest := backup.DatabaseManifest{Name: dbName}
		for _, spec := range specs {
			collectionName := spec.Name
			if strings.HasPrefix(collectionName, "system.") {
				continue
			}
			entry := backup.CollectionManifest{Name: collectionName, Type: spec.Type}
			if entry.Type == "" {
				entry.Type = "collection"
			}
			if len(spec.Options) > 0 {
				optionsJSON, marshalErr := bson.MarshalExtJSON(spec.Options, true, true)
				if marshalErr != nil {
					return model.BackupRecord{}, fmt.Errorf("read options for %s.%s: %w", dbName, collectionName, marshalErr)
				}
				entry.OptionsEJSON = string(optionsJSON)
			}

			if entry.Type == "view" {
				dbManifest.Collections = append(dbManifest.Collections, entry)
				continue
			}

			entry.DataFile = fmt.Sprintf("data/%s/%s.bson", backup.SafeKey(dbName), backup.SafeKey(collectionName))
			entry.IndexesFile = fmt.Sprintf("indexes/%s/%s.ejsonl", backup.SafeKey(dbName), backup.SafeKey(collectionName))
			dataWriter, err := zw.CreateHeader(&zip.FileHeader{Name: entry.DataFile, Method: zip.Deflate})
			if err != nil {
				return model.BackupRecord{}, err
			}
			cursor, err := client.Database(dbName).Collection(collectionName).Find(ctx, bson.D{})
			if err != nil {
				return model.BackupRecord{}, fmt.Errorf("read %s.%s: %w", dbName, collectionName, err)
			}
			for cursor.Next(ctx) {
				raw := cursor.Current
				if _, err := dataWriter.Write(raw); err != nil {
					_ = cursor.Close(ctx)
					return model.BackupRecord{}, err
				}
				entry.DocumentCount++
			}
			if err := cursor.Err(); err != nil {
				_ = cursor.Close(ctx)
				return model.BackupRecord{}, err
			}
			_ = cursor.Close(ctx)

			indexWriter, err := zw.CreateHeader(&zip.FileHeader{Name: entry.IndexesFile, Method: zip.Deflate})
			if err != nil {
				return model.BackupRecord{}, err
			}
			indexCursor, err := client.Database(dbName).Collection(collectionName).Indexes().List(ctx)
			if err != nil {
				return model.BackupRecord{}, fmt.Errorf("read indexes for %s.%s: %w", dbName, collectionName, err)
			}
			for indexCursor.Next(ctx) {
				jsonLine, err := bson.MarshalExtJSON(indexCursor.Current, true, true)
				if err != nil {
					_ = indexCursor.Close(ctx)
					return model.BackupRecord{}, err
				}
				if _, err := indexWriter.Write(append(jsonLine, '\n')); err != nil {
					_ = indexCursor.Close(ctx)
					return model.BackupRecord{}, err
				}
				entry.IndexCount++
			}
			if err := indexCursor.Err(); err != nil {
				_ = indexCursor.Close(ctx)
				return model.BackupRecord{}, err
			}
			_ = indexCursor.Close(ctx)
			dbManifest.Collections = append(dbManifest.Collections, entry)
		}
		manifest.Databases = append(manifest.Databases, dbManifest)
	}

	if err := backup.WriteManifest(zw, manifest); err != nil {
		return model.BackupRecord{}, err
	}
	if err := zw.Close(); err != nil {
		return model.BackupRecord{}, err
	}
	if err := out.Close(); err != nil {
		return model.BackupRecord{}, err
	}
	if err := os.Rename(tempPath, path); err != nil {
		return model.BackupRecord{}, err
	}
	ok = true
	return backup.RecordFromArchive(path)
}

func (m *Manager) RestoreBackup(parent context.Context, id, path string) (model.RestoreReport, error) {
	client, err := m.client(id)
	if err != nil {
		return model.RestoreReport{}, err
	}
	manifest, err := backup.ReadManifest(path)
	if err != nil {
		return model.RestoreReport{}, err
	}
	for _, db := range manifest.Databases {
		if systemDatabases[db.Name] {
			return model.RestoreReport{}, fmt.Errorf("backup contains protected MongoDB system database %q", db.Name)
		}
	}

	ctx, cancel := context.WithTimeout(parent, 30*time.Minute)
	defer cancel()
	zr, err := zip.OpenReader(path)
	if err != nil {
		return model.RestoreReport{}, err
	}
	defer zr.Close()

	report := model.RestoreReport{FileName: filepathBase(path), RestoredAt: time.Now().Format(time.RFC3339)}
	for _, dbEntry := range manifest.Databases {
		if err := client.Database(dbEntry.Name).Drop(ctx); err != nil {
			return model.RestoreReport{}, fmt.Errorf("drop database %s before restore: %w", dbEntry.Name, err)
		}
		report.DatabaseCount++
		for _, collEntry := range dbEntry.Collections {
			database := client.Database(dbEntry.Name)
			if err := restoreCollectionDefinition(ctx, database, collEntry); err != nil {
				return model.RestoreReport{}, fmt.Errorf("create %s.%s: %w", dbEntry.Name, collEntry.Name, err)
			}
			report.CollectionCount++
			if collEntry.Type == "view" {
				continue
			}
			collection := database.Collection(collEntry.Name)
			dataFile, err := backup.FindEntry(zr, collEntry.DataFile)
			if err != nil {
				return model.RestoreReport{}, err
			}
			dataReader, err := dataFile.Open()
			if err != nil {
				return model.RestoreReport{}, err
			}
			count, restoreErr := restoreBSONStream(ctx, collection, dataReader)
			_ = dataReader.Close()
			if restoreErr != nil {
				return model.RestoreReport{}, fmt.Errorf("restore %s.%s: %w", dbEntry.Name, collEntry.Name, restoreErr)
			}
			report.DocumentCount += count

			if collEntry.IndexesFile != "" {
				indexFile, err := backup.FindEntry(zr, collEntry.IndexesFile)
				if err != nil {
					return model.RestoreReport{}, err
				}
				indexReader, err := indexFile.Open()
				if err != nil {
					return model.RestoreReport{}, err
				}
				indexCount, restoreErr := restoreIndexes(ctx, client.Database(dbEntry.Name), collEntry.Name, indexReader)
				_ = indexReader.Close()
				if restoreErr != nil {
					return model.RestoreReport{}, fmt.Errorf("restore indexes for %s.%s: %w", dbEntry.Name, collEntry.Name, restoreErr)
				}
				report.IndexCount += indexCount
			}
		}
	}
	return report, nil
}

func restoreCollectionDefinition(ctx context.Context, database *mongo.Database, entry backup.CollectionManifest) error {
	cmd := bson.D{{Key: "create", Value: entry.Name}}
	if strings.TrimSpace(entry.OptionsEJSON) != "" {
		var optionsDoc bson.D
		if err := bson.UnmarshalExtJSON([]byte(entry.OptionsEJSON), false, &optionsDoc); err != nil {
			return fmt.Errorf("invalid collection options: %w", err)
		}
		cmd = append(cmd, optionsDoc...)
	}
	return database.RunCommand(ctx, cmd).Err()
}

func restoreBSONStream(ctx context.Context, collection *mongo.Collection, r io.Reader) (int64, error) {
	batch := make([]any, 0, 500)
	var inserted int64
	flush := func() error {
		if len(batch) == 0 {
			return nil
		}
		res, err := collection.InsertMany(ctx, batch)
		if err != nil {
			return err
		}
		inserted += int64(len(res.InsertedIDs))
		batch = batch[:0]
		return nil
	}
	_, err := readBSONDocuments(r, func(doc bson.D) error {
		batch = append(batch, doc)
		if len(batch) >= 500 {
			return flush()
		}
		return nil
	})
	if err != nil {
		return inserted, err
	}
	if err := flush(); err != nil {
		return inserted, err
	}
	return inserted, nil
}

func readBSONDocuments(r io.Reader, emit func(bson.D) error) (int64, error) {
	var count int64
	for {
		header := make([]byte, 4)
		_, err := io.ReadFull(r, header)
		if err == io.EOF {
			return count, nil
		}
		if err == io.ErrUnexpectedEOF {
			return count, fmt.Errorf("truncated BSON document length")
		}
		if err != nil {
			return count, err
		}
		length := int(binary.LittleEndian.Uint32(header))
		if length < 5 || length > 20*1024*1024 {
			return count, fmt.Errorf("invalid BSON document length %d", length)
		}
		raw := make([]byte, length)
		copy(raw, header)
		if _, err := io.ReadFull(r, raw[4:]); err != nil {
			return count, fmt.Errorf("truncated BSON document: %w", err)
		}
		var doc bson.D
		if err := bson.Unmarshal(raw, &doc); err != nil {
			return count, fmt.Errorf("invalid BSON document: %w", err)
		}
		if err := emit(doc); err != nil {
			return count, err
		}
		count++
	}
}

func restoreIndexes(ctx context.Context, database *mongo.Database, collectionName string, r io.Reader) (int, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 20*1024*1024)
	restored := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var raw bson.D
		if err := bson.UnmarshalExtJSON([]byte(line), false, &raw); err != nil {
			return restored, fmt.Errorf("invalid index metadata: %w", err)
		}
		spec, ok := sanitizeIndexSpec(raw)
		if !ok {
			continue
		}
		cmd := bson.D{
			{Key: "createIndexes", Value: collectionName},
			{Key: "indexes", Value: bson.A{spec}},
		}
		if err := database.RunCommand(ctx, cmd).Err(); err != nil {
			return restored, err
		}
		restored++
	}
	return restored, scanner.Err()
}

func sanitizeIndexSpec(doc bson.D) (bson.D, bool) {
	name := ""
	for _, e := range doc {
		if e.Key == "name" {
			name, _ = e.Value.(string)
		}
	}
	if name == "_id_" {
		return nil, false
	}
	allowed := map[string]bool{
		"key": true, "name": true, "unique": true, "sparse": true,
		"expireAfterSeconds": true, "partialFilterExpression": true, "collation": true,
		"wildcardProjection": true, "hidden": true, "storageEngine": true,
		"weights": true, "default_language": true, "language_override": true,
		"textIndexVersion": true, "2dsphereIndexVersion": true, "bits": true,
		"min": true, "max": true, "bucketSize": true,
	}
	out := bson.D{}
	for _, e := range doc {
		if allowed[e.Key] {
			out = append(out, e)
		}
	}
	return out, len(out) > 0
}

func filepathBase(path string) string {
	path = strings.ReplaceAll(path, "\\", "/")
	if i := strings.LastIndex(path, "/"); i >= 0 {
		return path[i+1:]
	}
	return path
}
