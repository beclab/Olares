package wizard

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Storage abstracts a small key/value store keyed by (kind, id).
// CLI uses DirKVStorage to persist AppState/Vault objects under
// ~/.olares/<did>/{kind}-{id}.json.
type Storage interface {
	Get(kind, id string, out any) error
	Put(kind, id string, in any) error
	Delete(kind, id string) error
	List(kind string) ([]string, error)
}

// DirKVStorage is a filesystem-backed Storage implementation that
// stores each (kind, id) entry as a JSON file under Root.
type DirKVStorage struct {
	Root string
	mu   sync.Mutex
}

// NewDirKVStorage creates a new DirKVStorage at the given root.
// Creates the root directory if it does not exist.
func NewDirKVStorage(root string) (*DirKVStorage, error) {
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, fmt.Errorf("failed to create storage root %s: %w", root, err)
	}
	return &DirKVStorage{Root: root}, nil
}

func (s *DirKVStorage) path(kind, id string) string {
	safeKind := sanitize(kind)
	safeID := sanitize(id)
	return filepath.Join(s.Root, fmt.Sprintf("%s-%s.json", safeKind, safeID))
}

func (s *DirKVStorage) Get(kind, id string, out any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.path(kind, id))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

func (s *DirKVStorage) Put(kind, id string, in any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := json.MarshalIndent(in, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s/%s: %w", kind, id, err)
	}
	target := s.path(kind, id)
	tmp := target + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("failed to write %s: %w", tmp, err)
	}
	return os.Rename(tmp, target)
}

func (s *DirKVStorage) Delete(kind, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := os.Remove(s.path(kind, id))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *DirKVStorage) List(kind string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entries, err := os.ReadDir(s.Root)
	if err != nil {
		return nil, err
	}
	prefix := sanitize(kind) + "-"
	var ids []string
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ".json") {
			continue
		}
		id := strings.TrimSuffix(strings.TrimPrefix(name, prefix), ".json")
		ids = append(ids, id)
	}
	return ids, nil
}

// sanitize replaces characters that are unsafe in filenames.
func sanitize(s string) string {
	r := strings.NewReplacer("/", "_", "\\", "_", ":", "_", " ", "_")
	return r.Replace(s)
}

// DefaultStorageRoot returns ~/.olares/<did>; falls back to current dir
// if the user home cannot be resolved.
func DefaultStorageRoot(did string) (string, error) {
	if did == "" {
		return "", fmt.Errorf("did is required to determine storage root")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve user home dir: %w", err)
	}
	return filepath.Join(home, ".olares", sanitize(did)), nil
}
