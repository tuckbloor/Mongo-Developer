package securestore

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type Store struct {
	mu   sync.Mutex
	path string
}

func New(configDir string) (*Store, error) {
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return nil, err
	}
	return &Store{path: filepath.Join(configDir, "secrets.json")}, nil
}

func (s *Store) Set(id, secret string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	values, err := s.load()
	if err != nil {
		return err
	}
	if secret == "" {
		delete(values, id)
	} else {
		encrypted, err := protect([]byte(secret), filepath.Dir(s.path))
		if err != nil {
			return err
		}
		values[id] = base64.StdEncoding.EncodeToString(encrypted)
	}
	return s.save(values)
}

func (s *Store) Get(id string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	values, err := s.load()
	if err != nil {
		return "", err
	}
	encoded, ok := values[id]
	if !ok {
		return "", nil
	}
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	plaintext, err := unprotect(ciphertext, filepath.Dir(s.path))
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func (s *Store) Delete(id string) error {
	return s.Set(id, "")
}

func (s *Store) load() (map[string]string, error) {
	b, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, err
	}
	values := map[string]string{}
	if len(b) == 0 {
		return values, nil
	}
	if err := json.Unmarshal(b, &values); err != nil {
		return nil, err
	}
	return values, nil
}

func (s *Store) save(values map[string]string) error {
	b, err := json.MarshalIndent(values, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
