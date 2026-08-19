package store

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"mongo-developer/internal/model"
	"mongo-developer/internal/securestore"
)

type ConnectionStore struct {
	mu      sync.Mutex
	path    string
	secrets *securestore.Store
}

func NewConnectionStore(configDir string) (*ConnectionStore, error) {
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return nil, err
	}
	secrets, err := securestore.New(configDir)
	if err != nil {
		return nil, err
	}
	return &ConnectionStore{path: filepath.Join(configDir, "connections.json"), secrets: secrets}, nil
}

func (s *ConnectionStore) List() ([]model.ConnectionProfile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	items, err := s.load()
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(i, j int) bool { return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name) })
	return items, nil
}

func (s *ConnectionStore) Get(id string) (model.ConnectionProfile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	items, err := s.load()
	if err != nil {
		return model.ConnectionProfile{}, err
	}
	for _, p := range items {
		if p.ID == id {
			return p, nil
		}
	}
	return model.ConnectionProfile{}, os.ErrNotExist
}

func (s *ConnectionStore) Save(input model.ConnectionInput) (model.ConnectionProfile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return model.ConnectionProfile{}, fmt.Errorf("connection name is required")
	}

	sanitizedURI, username, password, err := normalizeURI(input)
	if err != nil {
		return model.ConnectionProfile{}, err
	}
	if input.Password != "" {
		password = input.Password
	}
	if input.Username != "" {
		username = input.Username
	}

	items, err := s.load()
	if err != nil {
		return model.ConnectionProfile{}, err
	}
	now := time.Now().UTC()
	p := model.ConnectionProfile{
		ID: input.ID, Name: name, URI: sanitizedURI, Username: username,
		AuthDB: strings.TrimSpace(input.AuthDB), TLS: input.TLS, ReadOnly: input.ReadOnly,
		CreatedAt: now, UpdatedAt: now,
	}
	if p.ID == "" {
		p.ID = randomID()
	}
	found := false
	for i := range items {
		if items[i].ID == p.ID {
			p.CreatedAt = items[i].CreatedAt
			items[i] = p
			found = true
			break
		}
	}
	if !found {
		items = append(items, p)
	}
	if err := s.save(items); err != nil {
		return model.ConnectionProfile{}, err
	}
	if password != "" {
		if err := s.secrets.Set(p.ID, password); err != nil {
			return model.ConnectionProfile{}, err
		}
	}
	return p, nil
}

func (s *ConnectionStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	items, err := s.load()
	if err != nil {
		return err
	}
	out := items[:0]
	for _, p := range items {
		if p.ID != id {
			out = append(out, p)
		}
	}
	if err := s.save(out); err != nil {
		return err
	}
	return s.secrets.Delete(id)
}

func (s *ConnectionStore) RuntimeURI(id string) (string, error) {
	p, err := s.Get(id)
	if err != nil {
		return "", err
	}
	password, err := s.secrets.Get(id)
	if err != nil {
		return "", err
	}
	return injectCredentials(p, password)
}

func (s *ConnectionStore) RuntimeURIForDraft(input model.ConnectionInput) (string, error) {
	uri, username, password, err := normalizeURI(input)
	if err != nil {
		return "", err
	}
	p := model.ConnectionProfile{URI: uri, Username: username, AuthDB: input.AuthDB, TLS: input.TLS}
	if input.Username != "" {
		p.Username = input.Username
	}
	if input.Password != "" {
		password = input.Password
	}
	return injectCredentials(p, password)
}

func normalizeURI(input model.ConnectionInput) (sanitized, username, password string, err error) {
	raw := strings.TrimSpace(input.URI)
	if raw == "" {
		host := strings.TrimSpace(input.Host)
		if host == "" {
			host = "localhost"
		}
		port := input.Port
		if port == 0 {
			port = 27017
		}
		raw = fmt.Sprintf("mongodb://%s:%d", host, port)
	}
	if !strings.HasPrefix(raw, "mongodb://") && !strings.HasPrefix(raw, "mongodb+srv://") {
		return "", "", "", fmt.Errorf("URI must start with mongodb:// or mongodb+srv://")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid MongoDB URI: %w", err)
	}
	if u.Host == "" {
		return "", "", "", fmt.Errorf("MongoDB URI has no host")
	}
	if u.User != nil {
		username = u.User.Username()
		password, _ = u.User.Password()
		if username != "" {
			u.User = url.User(username)
		} else {
			u.User = nil
		}
	}
	return u.String(), username, password, nil
}

func injectCredentials(p model.ConnectionProfile, password string) (string, error) {
	u, err := url.Parse(p.URI)
	if err != nil {
		return "", err
	}
	username := p.Username
	if username == "" && u.User != nil {
		username = u.User.Username()
	}
	if username != "" {
		if password != "" {
			u.User = url.UserPassword(username, password)
		} else {
			u.User = url.User(username)
		}
	}
	q := u.Query()
	if p.AuthDB != "" && q.Get("authSource") == "" {
		q.Set("authSource", p.AuthDB)
	}
	if p.TLS && q.Get("tls") == "" && q.Get("ssl") == "" {
		q.Set("tls", "true")
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (s *ConnectionStore) load() ([]model.ConnectionProfile, error) {
	b, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return []model.ConnectionProfile{}, nil
	}
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return []model.ConnectionProfile{}, nil
	}
	var items []model.ConnectionProfile
	if err := json.Unmarshal(b, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *ConnectionStore) save(items []model.ConnectionProfile) error {
	b, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func randomID() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("conn-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
