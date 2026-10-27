package clusterop

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ErrReplayConflict means the binding was already consumed.
var ErrReplayConflict = errors.New("power request binding was already consumed")

// ReplayGuard persists the one fact needed to prevent a signed command from
// being executed more than once: its binding was consumed until expiry.
type ReplayGuard struct {
	dir string
}

type replayMarker struct {
	ExpiresAt int64 `json:"expiresAt"`
}

// NewReplayGuard prepares a persistent replay guard directory.
func NewReplayGuard(dir string) (*ReplayGuard, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, errors.New("power replay guard: empty directory")
	}
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return nil, fmt.Errorf("create power replay guard dir: %w", err)
	}
	if err := os.Chmod(dir, dirMode); err != nil {
		return nil, fmt.Errorf("secure power replay guard dir: %w", err)
	}
	guard := &ReplayGuard{dir: dir}
	if err := guard.Cleanup(); err != nil {
		return nil, err
	}
	return guard, nil
}

// Cleanup removes every readable marker that has expired. Damaged markers stay
// consumed: deleting an unreadable record could turn a corrupted disk write
// into permission to replay a power command.
func (s *ReplayGuard) Cleanup() error {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return fmt.Errorf("list power replay markers: %w", err)
	}
	removed := false
	now := time.Now()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".replay") {
			continue
		}
		path := filepath.Join(s.dir, entry.Name())
		marker, err := s.read(path)
		if err != nil || now.Before(time.UnixMilli(marker.ExpiresAt)) {
			continue
		}
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove expired power replay marker: %w", err)
		}
		removed = true
	}
	if removed {
		if err := syncDir(s.dir); err != nil {
			return fmt.Errorf("sync power replay directory: %w", err)
		}
	}
	return nil
}

// Consume atomically records key until expiresAt. A pre-existing unexpired
// marker is a replay conflict; expired markers are removed before retrying.
func (s *ReplayGuard) Consume(key string, expiresAt time.Time) error {
	if err := s.Cleanup(); err != nil {
		return err
	}
	path := s.path(key)
	if marker, err := s.read(path); err == nil && !time.Now().Before(time.UnixMilli(marker.ExpiresAt)) {
		if err := s.Forget(key); err != nil {
			return err
		}
	} else if err == nil {
		return ErrReplayConflict
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read power replay marker: %w", err)
	}

	data, err := json.Marshal(replayMarker{ExpiresAt: expiresAt.UnixMilli()})
	if err != nil {
		return fmt.Errorf("encode power replay marker: %w", err)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, recordMode)
	if errors.Is(err, os.ErrExist) {
		return ErrReplayConflict
	}
	if err != nil {
		return fmt.Errorf("consume power request binding: %w", err)
	}
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(path)
		return fmt.Errorf("write power replay marker: %w", err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(path)
		return fmt.Errorf("persist power replay marker: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(path)
		return fmt.Errorf("close power replay marker: %w", err)
	}
	if err := syncDir(s.dir); err != nil {
		_ = os.Remove(path)
		return fmt.Errorf("sync power replay directory: %w", err)
	}
	return nil
}

// Forget removes the marker only when the underlying command synchronously
// refused to start, allowing the same still-valid authorization to be retried.
func (s *ReplayGuard) Forget(key string) error {
	path := s.path(key)
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("forget power replay marker: %w", err)
	}
	if err := syncDir(s.dir); err != nil {
		return fmt.Errorf("sync power replay directory: %w", err)
	}
	return nil
}

func (s *ReplayGuard) read(path string) (replayMarker, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return replayMarker{}, err
	}
	var marker replayMarker
	if err := json.Unmarshal(data, &marker); err != nil {
		return replayMarker{}, err
	}
	return marker, nil
}

func (s *ReplayGuard) path(key string) string {
	sum := sha256.Sum256([]byte(key))
	return filepath.Join(s.dir, hex.EncodeToString(sum[:])+".replay")
}
