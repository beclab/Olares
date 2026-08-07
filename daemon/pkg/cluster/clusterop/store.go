package clusterop

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"k8s.io/klog/v2"
)

const (
	recordSuffix = ".json"
	recordMode   = 0o600
	dirMode      = 0o700
)

// ErrInvalidID rejects an operation ID that is not usable as a single file
// name. IDs arrive from a URL path, so this is the boundary that keeps a
// request from naming a file outside the state directory.
var ErrInvalidID = errors.New("invalid operation id")

// Store keeps operation records as one JSON file each under a directory in the
// daemon's state area.
//
// A record is replaced by writing a sibling temp file and renaming over the
// old one, so a poll that lands mid-write reads the previous record rather
// than half of the next. There is no database: an operation is a handful of
// small documents, and the one thing that must survive is a reboot of the
// machine the database would have run on.
type Store struct {
	dir string

	// mu serializes writers. Readers do not take it: rename is atomic, so a
	// reader either sees the old file or the new one.
	mu sync.Mutex
}

// NewStore prepares the directory and returns a store rooted at it.
func NewStore(dir string) (*Store, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, errors.New("cluster operation store: empty directory")
	}
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return nil, fmt.Errorf("create cluster operation dir: %w", err)
	}
	// MkdirAll respects the umask and leaves an existing directory alone, so
	// the mode above is a request rather than a guarantee.
	if err := os.Chmod(dir, dirMode); err != nil {
		return nil, fmt.Errorf("secure cluster operation dir: %w", err)
	}
	return &Store{dir: dir}, nil
}

// Dir is where the records live.
func (s *Store) Dir() string { return s.dir }

func (s *Store) path(id string) (string, error) {
	if id == "" || id == "." || id == ".." ||
		strings.ContainsRune(id, filepath.Separator) || strings.ContainsRune(id, '/') {
		return "", fmt.Errorf("%w: %q", ErrInvalidID, id)
	}
	return filepath.Join(s.dir, id+recordSuffix), nil
}

// Save writes the record, replacing any previous version atomically.
func (s *Store) Save(op Operation) error {
	final, err := s.path(op.ID)
	if err != nil {
		return err
	}
	data, err := json.Marshal(op)
	if err != nil {
		return fmt.Errorf("encode operation %s: %w", op.ID, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tmp, err := os.CreateTemp(s.dir, "."+op.ID+".*.tmp")
	if err != nil {
		return fmt.Errorf("stage operation %s: %w", op.ID, err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if err := writeAndSync(tmp, data); err != nil {
		return fmt.Errorf("write operation %s: %w", op.ID, err)
	}
	if err := os.Chmod(tmpName, recordMode); err != nil {
		return fmt.Errorf("secure operation %s: %w", op.ID, err)
	}
	if err := os.Rename(tmpName, final); err != nil {
		return fmt.Errorf("commit operation %s: %w", op.ID, err)
	}
	// The rename itself is what a reader sees; syncing the directory is what
	// makes it survive losing power a moment later, which is the whole point
	// of a record describing a power operation.
	if err := syncDir(s.dir); err != nil {
		return fmt.Errorf("sync operation directory: %w", err)
	}
	return nil
}

func writeAndSync(f *os.File, data []byte) error {
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return err
	}
	return f.Sync()
}

func syncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	if err := d.Sync(); err != nil {
		return err
	}
	return nil
}

// Load reads one record. A record that was never written is not an error.
func (s *Store) Load(id string) (Operation, bool, error) {
	path, err := s.path(id)
	if err != nil {
		return Operation{}, false, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Operation{}, false, nil
		}
		return Operation{}, false, err
	}
	var op Operation
	if err := json.Unmarshal(data, &op); err != nil {
		return Operation{}, false, fmt.Errorf("decode operation %s: %w", id, err)
	}
	return op, true, nil
}

// List reads every record, oldest first. A record that cannot be decoded is
// skipped and logged: one lost operation must not hide the rest.
func (s *Store) List() ([]Operation, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	ops := make([]Operation, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), recordSuffix) {
			continue
		}
		id := strings.TrimSuffix(e.Name(), recordSuffix)
		op, ok, err := s.Load(id)
		if err != nil {
			klog.Warningf("clusterop: skipping unreadable operation %s: %v", id, err)
			continue
		}
		if ok {
			ops = append(ops, op)
		}
	}
	sort.Slice(ops, func(i, j int) bool { return ops[i].CreatedAt.Before(ops[j].CreatedAt) })
	return ops, nil
}

// Delete removes a record. Deleting one that is already gone is a no-op.
func (s *Store) Delete(id string) error {
	path, err := s.path(id)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
