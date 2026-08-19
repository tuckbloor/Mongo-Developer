package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"mongo-developer/internal/backup"
	"mongo-developer/internal/laravel"
	"mongo-developer/internal/model"
	"mongo-developer/internal/mongoapp"
	"mongo-developer/internal/store"
)

const appVersion = "1.5.0"

type App struct {
	ctx     context.Context
	store   *store.ConnectionStore
	mongo   *mongoapp.Manager
	backups *backup.Store
	appDir  string
}

func NewApp() (*App, error) {
	cfg, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	appDir := filepath.Join(cfg, "MongoDeveloper")
	cs, err := store.NewConnectionStore(appDir)
	if err != nil {
		return nil, err
	}
	bs, err := backup.NewStore(filepath.Join(appDir, "backups"))
	if err != nil {
		return nil, err
	}
	return &App{store: cs, mongo: mongoapp.New(), backups: bs, appDir: appDir}, nil
}

func (a *App) startup(ctx context.Context) { a.ctx = ctx }
func (a *App) shutdown(context.Context)    { a.mongo.CloseAll() }

func (a *App) GetConnections() ([]model.ConnectionProfile, error) { return a.store.List() }
func (a *App) SaveConnection(input model.ConnectionInput) (model.ConnectionProfile, error) {
	return a.store.Save(input)
}
func (a *App) DeleteConnection(id string) error {
	_ = a.mongo.Disconnect(id)
	return a.store.Delete(id)
}

func (a *App) TestConnection(input model.ConnectionInput) error {
	uri, err := a.store.RuntimeURIForDraft(input)
	if err != nil {
		return err
	}
	return a.mongo.Test(a.ctx, uri)
}
func (a *App) Connect(id string) error {
	uri, err := a.store.RuntimeURI(id)
	if err != nil {
		return err
	}
	return a.mongo.Connect(a.ctx, id, uri)
}
func (a *App) Disconnect(id string) error { return a.mongo.Disconnect(id) }

func (a *App) GetConnectionHealth(id string) (model.ConnectionHealth, error) {
	return a.mongo.Health(a.ctx, id)
}

func (a *App) Reconnect(id string) (model.ConnectionHealth, error) {
	uri, err := a.store.RuntimeURI(id)
	if err != nil {
		return model.ConnectionHealth{}, err
	}
	if err := a.mongo.Connect(a.ctx, id, uri); err != nil {
		return model.ConnectionHealth{}, err
	}
	return a.mongo.Health(a.ctx, id)
}

func (a *App) ExplainFind(id, db, collection, filterJSON, sortJSON string) (model.ExplainResult, error) {
	return a.mongo.ExplainFind(a.ctx, id, db, collection, filterJSON, sortJSON)
}

func (a *App) ExplainAggregation(id, db, collection, pipelineJSON string) (model.ExplainResult, error) {
	return a.mongo.ExplainAggregation(a.ctx, id, db, collection, pipelineJSON)
}

func (a *App) GetDiagnostics(id string) (model.Diagnostics, error) {
	d := model.Diagnostics{
		Version: appVersion, OS: goruntime.GOOS, Arch: goruntime.GOARCH, GoVersion: goruntime.Version(),
		ConfigDirectory: a.appDir, BackupDirectory: a.backups.Root(), CheckedAt: time.Now().Format(time.RFC3339),
	}
	if strings.TrimSpace(id) == "" {
		return d, nil
	}
	p, err := a.store.Get(id)
	if err != nil {
		return d, err
	}
	d.ConnectionName = p.Name
	d.ConnectionURI = p.URI
	d.ReadOnly = p.ReadOnly
	health, healthErr := a.mongo.Health(a.ctx, id)
	d.Connected = health.Connected
	d.PingMS = health.PingMS
	d.MongoDBVersion = health.MongoDBVersion
	if healthErr != nil {
		d.ConnectionError = healthErr.Error()
	}
	return d, nil
}

func (a *App) EnsureDummyTestData(id string) (model.DummyDataReport, error) {
	if err := a.requireWritable(id); err != nil {
		return model.DummyDataReport{}, err
	}
	return a.mongo.EnsureDummyData(a.ctx, id)
}

func (a *App) ResetDummyTestData(id string) (model.DummyDataReport, error) {
	if err := a.requireWritable(id); err != nil {
		return model.DummyDataReport{}, err
	}
	return a.mongo.InstallDummyData(a.ctx, id)
}

func (a *App) DummyTestDataStatus(id string) (model.DummyDataReport, error) {
	return a.mongo.DummyDataStatus(a.ctx, id)
}
func (a *App) ListDatabases(id string) ([]model.DatabaseInfo, error) {
	return a.mongo.ListDatabases(a.ctx, id)
}
func (a *App) ListCollections(id, db string) ([]model.CollectionInfo, error) {
	return a.mongo.ListCollections(a.ctx, id, db)
}

func (a *App) CreateDatabase(id, db, initialCollection string) error {
	if err := a.requireWritable(id); err != nil {
		return err
	}
	return a.mongo.CreateDatabase(a.ctx, id, db, initialCollection)
}
func (a *App) DropDatabase(id, db string) error {
	if err := a.requireWritable(id); err != nil {
		return err
	}
	return a.mongo.DropDatabase(a.ctx, id, db)
}
func (a *App) CreateCollection(id, db, collection string) error {
	if err := a.requireWritable(id); err != nil {
		return err
	}
	return a.mongo.CreateCollection(a.ctx, id, db, collection)
}
func (a *App) DropCollection(id, db, collection string) error {
	if err := a.requireWritable(id); err != nil {
		return err
	}
	return a.mongo.DropCollection(a.ctx, id, db, collection)
}
func (a *App) FindDocuments(req model.FindRequest) (model.FindResult, error) {
	return a.mongo.Find(a.ctx, req)
}
func (a *App) InsertDocument(id, db, collection, json string) (string, error) {
	if err := a.requireWritable(id); err != nil {
		return "", err
	}
	return a.mongo.Insert(a.ctx, id, db, collection, json)
}
func (a *App) ReplaceDocument(id, db, collection, filterJSON, documentJSON string) (int64, error) {
	if err := a.requireWritable(id); err != nil {
		return 0, err
	}
	return a.mongo.Replace(a.ctx, id, db, collection, filterJSON, documentJSON)
}
func (a *App) DeleteDocument(id, db, collection, filterJSON string) (int64, error) {
	if err := a.requireWritable(id); err != nil {
		return 0, err
	}
	return a.mongo.Delete(a.ctx, id, db, collection, filterJSON)
}
func (a *App) GetCollectionStats(id, db, collection string) (model.CollectionStats, error) {
	return a.mongo.Stats(a.ctx, id, db, collection)
}
func (a *App) ListIndexes(id, db, collection string) ([]string, error) {
	return a.mongo.ListIndexes(a.ctx, id, db, collection)
}
func (a *App) CreateIndex(id, db, collection, keysJSON, name string, unique bool) (string, error) {
	if err := a.requireWritable(id); err != nil {
		return "", err
	}
	return a.mongo.CreateIndex(a.ctx, id, db, collection, keysJSON, name, unique)
}
func (a *App) RunAggregation(id, db, collection, pipelineJSON string) ([]string, error) {
	writeStage, err := mongoapp.AggregationWriteStage(pipelineJSON)
	if err != nil {
		return nil, err
	}
	if writeStage != "" {
		if err := a.requireWritable(id); err != nil {
			return nil, fmt.Errorf("aggregation stage %s writes data: %w", writeStage, err)
		}
	}
	return a.mongo.Aggregate(a.ctx, id, db, collection, pipelineJSON, 100)
}

func (a *App) ImportJSON(id, db, collection string) (int, error) {
	if err := a.requireWritable(id); err != nil {
		return 0, err
	}
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Title: "Import JSON/EJSON", Filters: []runtime.FileFilter{{DisplayName: "JSON files", Pattern: "*.json"}}})
	if err != nil || path == "" {
		return 0, err
	}
	return a.mongo.ImportJSON(a.ctx, id, db, collection, path)
}
func (a *App) ExportJSON(id, db, collection string) (string, error) {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{Title: "Export collection", DefaultFilename: collection + ".json", Filters: []runtime.FileFilter{{DisplayName: "JSON files", Pattern: "*.json"}}})
	if err != nil || path == "" {
		return "", err
	}
	if !strings.HasSuffix(strings.ToLower(path), ".json") {
		path += ".json"
	}
	if err := a.mongo.ExportJSON(a.ctx, id, db, collection, path); err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) ListBackups() ([]model.BackupRecord, error) {
	return a.backups.List()
}

func (a *App) GetBackupDirectory() string {
	return a.backups.Root()
}

func (a *App) CreateBackup(id, scope, database string) (model.BackupCreateResult, error) {
	profile, err := a.store.Get(id)
	if err != nil {
		return model.BackupCreateResult{}, err
	}
	path, fileName, err := a.backups.NewPath(scope, database, time.Now())
	if err != nil {
		return model.BackupCreateResult{}, err
	}
	record, err := a.mongo.CreateBackup(a.ctx, id, profile.Name, scope, database, path)
	if err != nil {
		return model.BackupCreateResult{}, err
	}

	result := model.BackupCreateResult{Backup: record}
	downloadPath, dialogErr := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save backup copy",
		DefaultFilename: fileName,
		Filters:         []runtime.FileFilter{{DisplayName: "Mongo Developer backup", Pattern: "*.zip"}},
	})
	if dialogErr != nil {
		result.DownloadError = dialogErr.Error()
	} else if downloadPath != "" {
		if !strings.HasSuffix(strings.ToLower(downloadPath), ".zip") {
			downloadPath += ".zip"
		}
		if err := a.backups.Copy(fileName, downloadPath); err != nil {
			result.DownloadError = err.Error()
		} else {
			result.DownloadPath = downloadPath
		}
	}
	return result, nil
}

func (a *App) DownloadBackup(fileName string) (string, error) {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Save backup copy",
		DefaultFilename: fileName,
		Filters:         []runtime.FileFilter{{DisplayName: "Mongo Developer backup", Pattern: "*.zip"}},
	})
	if err != nil || path == "" {
		return "", err
	}
	if !strings.HasSuffix(strings.ToLower(path), ".zip") {
		path += ".zip"
	}
	if err := a.backups.Copy(fileName, path); err != nil {
		return "", err
	}
	return path, nil
}

func (a *App) DeleteBackup(fileName string) error {
	return a.backups.Delete(fileName)
}

func (a *App) RestoreBackup(id, fileName string) (model.RestoreReport, error) {
	if err := a.requireWritable(id); err != nil {
		return model.RestoreReport{}, err
	}
	path, err := a.backups.Path(fileName)
	if err != nil {
		return model.RestoreReport{}, err
	}
	return a.mongo.RestoreBackup(a.ctx, id, path)
}

func (a *App) RestoreBackupFromFile(id string) (model.RestoreReport, error) {
	if err := a.requireWritable(id); err != nil {
		return model.RestoreReport{}, err
	}
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "Restore Mongo Developer backup",
		Filters: []runtime.FileFilter{{DisplayName: "Mongo Developer backup", Pattern: "*.zip"}},
	})
	if err != nil || path == "" {
		return model.RestoreReport{}, err
	}
	return a.mongo.RestoreBackup(a.ctx, id, path)
}

func (a *App) SelectLaravelProject() (model.LaravelProjectInfo, error) {
	path, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: "Select Laravel project"})
	if err != nil || path == "" {
		return model.LaravelProjectInfo{}, err
	}
	return laravel.Inspect(path)
}
func (a *App) InspectLaravelProject(path string) (model.LaravelProjectInfo, error) {
	return laravel.Inspect(path)
}
func (a *App) GenerateLaravelModel(path, collection, className, sampleJSON string) (string, error) {
	return laravel.GenerateModel(path, collection, className, sampleJSON)
}
func (a *App) LaravelSetupHelp(id, database string) (string, error) {
	p, err := a.store.Get(id)
	if err != nil {
		return "", err
	}
	// Deliberately use the persisted password-free URI in help text.
	return laravel.Instructions(p.URI, database), nil
}

func (a *App) requireWritable(id string) error {
	p, err := a.store.Get(id)
	if err != nil {
		return err
	}
	if p.ReadOnly {
		return fmt.Errorf("%s is read-only; destructive operations are disabled", p.Name)
	}
	return nil
}
