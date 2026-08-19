package backup

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"mongo-developer/internal/model"
)

const FormatVersion = 1
const ManifestName = "manifest.json"

type Manifest struct {
	FormatVersion int                `json:"formatVersion"`
	CreatedAt     string             `json:"createdAt"`
	Scope         string             `json:"scope"`
	Database      string             `json:"database,omitempty"`
	SourceName    string             `json:"sourceName,omitempty"`
	Databases     []DatabaseManifest `json:"databases"`
}

type DatabaseManifest struct {
	Name        string               `json:"name"`
	Collections []CollectionManifest `json:"collections"`
}

type CollectionManifest struct {
	Name          string `json:"name"`
	Type          string `json:"type,omitempty"`
	OptionsEJSON  string `json:"optionsEjson,omitempty"`
	DataFile      string `json:"dataFile,omitempty"`
	IndexesFile   string `json:"indexesFile,omitempty"`
	DocumentCount int64  `json:"documentCount"`
	IndexCount    int    `json:"indexCount"`
}

type Store struct {
	root string
}

func NewStore(root string) (*Store, error) {
	if strings.TrimSpace(root) == "" {
		return nil, fmt.Errorf("backup directory is required")
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, err
	}
	return &Store{root: root}, nil
}

func (s *Store) Root() string { return s.root }

func (s *Store) NewPath(scope, database string, now time.Time) (string, string, error) {
	if scope != "all" && scope != "database" {
		return "", "", fmt.Errorf("backup scope must be all or database")
	}
	label := "all"
	if scope == "database" {
		label = safeLabel(database)
		if label == "" {
			return "", "", fmt.Errorf("database name is required")
		}
	}
	stamp := now.Local().Format("20060102-150405")
	base := fmt.Sprintf("mongo-developer-%s-%s.zip", label, stamp)
	path := filepath.Join(s.root, base)
	for i := 2; ; i++ {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path, filepath.Base(path), nil
		} else if err != nil {
			return "", "", err
		}
		base = fmt.Sprintf("mongo-developer-%s-%s-%d.zip", label, stamp, i)
		path = filepath.Join(s.root, base)
	}
}

func (s *Store) Path(fileName string) (string, error) {
	name := filepath.Base(strings.TrimSpace(fileName))
	if name == "." || name == "" || name != fileName || !strings.HasSuffix(strings.ToLower(name), ".zip") {
		return "", fmt.Errorf("invalid backup file name")
	}
	path := filepath.Join(s.root, name)
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	return path, nil
}

func (s *Store) List() ([]model.BackupRecord, error) {
	entries, err := os.ReadDir(s.root)
	if err != nil {
		return nil, err
	}
	out := make([]model.BackupRecord, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".zip") {
			continue
		}
		path := filepath.Join(s.root, entry.Name())
		record, err := RecordFromArchive(path)
		if err != nil {
			continue
		}
		out = append(out, record)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt > out[j].CreatedAt })
	return out, nil
}

func (s *Store) Delete(fileName string) error {
	path, err := s.Path(fileName)
	if err != nil {
		return err
	}
	return os.Remove(path)
}

func (s *Store) Copy(fileName, destination string) error {
	path, err := s.Path(fileName)
	if err != nil {
		return err
	}
	return copyFile(path, destination)
}

func RecordFromArchive(path string) (model.BackupRecord, error) {
	manifest, err := ReadManifest(path)
	if err != nil {
		return model.BackupRecord{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return model.BackupRecord{}, err
	}
	record := model.BackupRecord{
		FileName:   filepath.Base(path),
		CreatedAt:  manifest.CreatedAt,
		Scope:      manifest.Scope,
		Database:   manifest.Database,
		SizeBytes:  info.Size(),
		SourceName: manifest.SourceName,
	}
	for _, db := range manifest.Databases {
		record.DatabaseCount++
		for _, coll := range db.Collections {
			record.CollectionCount++
			record.DocumentCount += coll.DocumentCount
			record.IndexCount += coll.IndexCount
		}
	}
	return record, nil
}

func ReadManifest(path string) (Manifest, error) {
	zr, err := zip.OpenReader(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("invalid backup ZIP: %w", err)
	}
	defer zr.Close()
	for _, f := range zr.File {
		if f.Name != ManifestName {
			continue
		}
		r, err := f.Open()
		if err != nil {
			return Manifest{}, err
		}
		data, readErr := io.ReadAll(io.LimitReader(r, 4*1024*1024))
		_ = r.Close()
		if readErr != nil {
			return Manifest{}, readErr
		}
		var m Manifest
		if err := json.Unmarshal(data, &m); err != nil {
			return Manifest{}, fmt.Errorf("invalid backup manifest: %w", err)
		}
		if err := ValidateManifest(m); err != nil {
			return Manifest{}, err
		}
		return m, nil
	}
	return Manifest{}, fmt.Errorf("backup is missing %s", ManifestName)
}

func ValidateManifest(m Manifest) error {
	if m.FormatVersion != FormatVersion {
		return fmt.Errorf("unsupported backup format version %d", m.FormatVersion)
	}
	if m.Scope != "all" && m.Scope != "database" {
		return fmt.Errorf("invalid backup scope %q", m.Scope)
	}
	if len(m.Databases) == 0 {
		return fmt.Errorf("backup contains no databases")
	}
	seenDB := map[string]bool{}
	for _, db := range m.Databases {
		if strings.TrimSpace(db.Name) == "" || seenDB[db.Name] {
			return fmt.Errorf("backup contains an invalid or duplicate database name")
		}
		seenDB[db.Name] = true
		seenColl := map[string]bool{}
		for _, coll := range db.Collections {
			if strings.TrimSpace(coll.Name) == "" || seenColl[coll.Name] {
				return fmt.Errorf("backup contains an invalid or duplicate collection name in %s", db.Name)
			}
			seenColl[coll.Name] = true
			if coll.DataFile != "" && !safeArchiveEntry(coll.DataFile) {
				return fmt.Errorf("backup contains an unsafe data archive path")
			}
			if coll.IndexesFile != "" && !safeArchiveEntry(coll.IndexesFile) {
				return fmt.Errorf("backup contains an unsafe index archive path")
			}
			if coll.Type != "" && coll.Type != "collection" && coll.Type != "view" {
				return fmt.Errorf("backup contains unsupported collection type %q", coll.Type)
			}
		}
	}
	return nil
}

func WriteManifest(zw *zip.Writer, manifest Manifest) error {
	manifest.FormatVersion = FormatVersion
	if err := ValidateManifest(manifest); err != nil {
		return err
	}
	w, err := zw.CreateHeader(&zip.FileHeader{Name: ManifestName, Method: zip.Deflate})
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(manifest)
}

func FindEntry(zr *zip.ReadCloser, name string) (*zip.File, error) {
	if !safeArchiveEntry(name) {
		return nil, fmt.Errorf("unsafe archive entry path")
	}
	for _, f := range zr.File {
		if f.Name == name {
			return f, nil
		}
	}
	return nil, fmt.Errorf("backup entry %s is missing", name)
}

func SafeKey(value string) string {
	const hex = "0123456789abcdef"
	data := []byte(value)
	var b strings.Builder
	b.Grow(len(data) * 2)
	for _, v := range data {
		b.WriteByte(hex[v>>4])
		b.WriteByte(hex[v&15])
	}
	return b.String()
}

func safeLabel(value string) string {
	value = strings.TrimSpace(value)
	var b strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	return strings.Trim(b.String(), "_")
}

func safeArchiveEntry(name string) bool {
	if name == "" || strings.HasPrefix(name, "/") || strings.Contains(name, "\\") {
		return false
	}
	clean := filepath.ToSlash(filepath.Clean(name))
	return clean == name && clean != "." && !strings.HasPrefix(clean, "../") && !strings.Contains(clean, "/../")
}

func copyFile(source, destination string) error {
	if strings.TrimSpace(destination) == "" {
		return fmt.Errorf("destination path is required")
	}
	sourceAbs, sourceErr := filepath.Abs(source)
	destinationAbs, destinationErr := filepath.Abs(destination)
	if sourceErr == nil && destinationErr == nil && strings.EqualFold(filepath.Clean(sourceAbs), filepath.Clean(destinationAbs)) {
		return nil
	}
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}
