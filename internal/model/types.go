package model

import "time"

type ConnectionProfile struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	URI       string    `json:"uri"`
	Username  string    `json:"username,omitempty"`
	AuthDB    string    `json:"authDb,omitempty"`
	TLS       bool      `json:"tls"`
	ReadOnly  bool      `json:"readOnly"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ConnectionInput struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name"`
	URI      string `json:"uri"`
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	AuthDB   string `json:"authDb,omitempty"`
	TLS      bool   `json:"tls"`
	ReadOnly bool   `json:"readOnly"`
}

type DatabaseInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size,omitempty"`
}

type CollectionInfo struct {
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}

type FindRequest struct {
	ConnectionID string `json:"connectionId"`
	Database     string `json:"database"`
	Collection   string `json:"collection"`
	FilterJSON   string `json:"filterJson"`
	SortJSON     string `json:"sortJson,omitempty"`
	Skip         int64  `json:"skip"`
	Limit        int64  `json:"limit"`
}

type FindResult struct {
	Documents []string `json:"documents"`
	Total     int64    `json:"total"`
	QueryMS   int64    `json:"queryMs"`
}

type CollectionStats struct {
	Documents     int64   `json:"documents"`
	Indexes       int64   `json:"indexes"`
	AvgDocumentKB float64 `json:"avgDocumentKb"`
	TotalSizeMB   float64 `json:"totalSizeMb"`
	StorageSizeMB float64 `json:"storageSizeMb"`
}

type LaravelModelInfo struct {
	ClassName       string   `json:"className"`
	Namespace       string   `json:"namespace,omitempty"`
	Path            string   `json:"path"`
	Collection      string   `json:"collection"`
	Connection      string   `json:"connection,omitempty"`
	MongoModel      bool     `json:"mongoModel"`
	UsesLegacyField bool     `json:"usesLegacyField"`
	Relationships   []string `json:"relationships,omitempty"`
}

type LaravelProjectInfo struct {
	Path                  string             `json:"path"`
	IsLaravel             bool               `json:"isLaravel"`
	LaravelVersion        string             `json:"laravelVersion,omitempty"`
	MongoPackageInstalled bool               `json:"mongoPackageInstalled"`
	MongoPackageVersion   string             `json:"mongoPackageVersion,omitempty"`
	EnvExists             bool               `json:"envExists"`
	MongoConfigured       bool               `json:"mongoConfigured"`
	ArtisanExists         bool               `json:"artisanExists"`
	Models                []LaravelModelInfo `json:"models,omitempty"`
}

type DummyDataReport struct {
	Database    string `json:"database"`
	Collections int    `json:"collections"`
	Documents   int    `json:"documents"`
	Indexes     int    `json:"indexes"`
	Created     bool   `json:"created"`
}

type BackupRecord struct {
	FileName        string `json:"fileName"`
	CreatedAt       string `json:"createdAt"`
	Scope           string `json:"scope"`
	Database        string `json:"database,omitempty"`
	DatabaseCount   int    `json:"databaseCount"`
	CollectionCount int    `json:"collectionCount"`
	DocumentCount   int64  `json:"documentCount"`
	IndexCount      int    `json:"indexCount"`
	SizeBytes       int64  `json:"sizeBytes"`
	SourceName      string `json:"sourceName,omitempty"`
}

type BackupCreateResult struct {
	Backup        BackupRecord `json:"backup"`
	DownloadPath  string       `json:"downloadPath,omitempty"`
	DownloadError string       `json:"downloadError,omitempty"`
}

type RestoreReport struct {
	FileName        string `json:"fileName"`
	DatabaseCount   int    `json:"databaseCount"`
	CollectionCount int    `json:"collectionCount"`
	DocumentCount   int64  `json:"documentCount"`
	IndexCount      int    `json:"indexCount"`
	RestoredAt      string `json:"restoredAt"`
}

type ConnectionHealth struct {
	Connected      bool   `json:"connected"`
	PingMS         int64  `json:"pingMs"`
	MongoDBVersion string `json:"mongoDbVersion,omitempty"`
	CheckedAt      string `json:"checkedAt"`
	Error          string `json:"error,omitempty"`
}

type ExplainResult struct {
	Mode              string `json:"mode"`
	Database          string `json:"database"`
	Collection        string `json:"collection"`
	Stage             string `json:"stage"`
	IndexName         string `json:"indexName,omitempty"`
	UsesIndex         bool   `json:"usesIndex"`
	CollectionScan    bool   `json:"collectionScan"`
	DocumentsExamined int64  `json:"documentsExamined"`
	KeysExamined      int64  `json:"keysExamined"`
	Returned          int64  `json:"returned"`
	ExecutionMS       int64  `json:"executionMs"`
	SuggestedIndex    string `json:"suggestedIndex,omitempty"`
	Summary           string `json:"summary"`
}

type Diagnostics struct {
	Version         string `json:"version"`
	OS              string `json:"os"`
	Arch            string `json:"arch"`
	GoVersion       string `json:"goVersion"`
	ConfigDirectory string `json:"configDirectory"`
	BackupDirectory string `json:"backupDirectory"`
	ConnectionName  string `json:"connectionName,omitempty"`
	ConnectionURI   string `json:"connectionUri,omitempty"`
	Connected       bool   `json:"connected"`
	PingMS          int64  `json:"pingMs"`
	MongoDBVersion  string `json:"mongoDbVersion,omitempty"`
	ReadOnly        bool   `json:"readOnly"`
	CheckedAt       string `json:"checkedAt"`
	ConnectionError string `json:"connectionError,omitempty"`
}
