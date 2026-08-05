package clusterop

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrClaimExists makes a repeated request idempotent.
var ErrClaimExists = errors.New("power request was already claimed")

const completedClaim = "completed"

// ClaimStore persists one at-most-once claim per signed power request.
type ClaimStore struct {
	dir string
}

// NewClaimStore prepares a claim directory.
func NewClaimStore(dir string) (*ClaimStore, error) {
	if strings.TrimSpace(dir) == "" {
		return nil, errors.New("power claim store: empty directory")
	}
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return nil, fmt.Errorf("create power claim dir: %w", err)
	}
	if err := os.Chmod(dir, dirMode); err != nil {
		return nil, fmt.Errorf("secure power claim dir: %w", err)
	}
	return &ClaimStore{dir: dir}, nil
}

// Claim records key before the corresponding power command is executed.
func (s *ClaimStore) Claim(key string) error {
	path := s.path(key)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, recordMode)
	if errors.Is(err, os.ErrExist) {
		return ErrClaimExists
	}
	if err != nil {
		return fmt.Errorf("claim power request: %w", err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(path)
		return fmt.Errorf("persist power request claim: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(path)
		return fmt.Errorf("close power request claim: %w", err)
	}
	if err := syncDir(s.dir); err != nil {
		_ = os.Remove(path)
		return fmt.Errorf("sync power claim directory: %w", err)
	}
	return nil
}

// Release removes a claim when command execution failed before acceptance.
func (s *ClaimStore) Release(key string) error {
	path := s.path(key)
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("release power request claim: %w", err)
	}
	if err := syncDir(s.dir); err != nil {
		return fmt.Errorf("sync released power claim: %w", err)
	}
	return nil
}

// Complete records that the claimed command was accepted.
func (s *ClaimStore) Complete(key string) error {
	path := s.path(key)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, recordMode)
	if err != nil {
		return fmt.Errorf("open power request claim: %w", err)
	}
	if _, err := f.WriteString(completedClaim); err != nil {
		_ = f.Close()
		return fmt.Errorf("complete power request claim: %w", err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return fmt.Errorf("persist completed power request claim: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close completed power request claim: %w", err)
	}
	return nil
}

// Completed reports whether an existing claim reached command acceptance.
func (s *ClaimStore) Completed(key string) (bool, error) {
	data, err := os.ReadFile(s.path(key))
	if err != nil {
		return false, fmt.Errorf("read power request claim: %w", err)
	}
	return string(data) == completedClaim, nil
}

func (s *ClaimStore) path(key string) string {
	sum := sha256.Sum256([]byte(key))
	return filepath.Join(s.dir, hex.EncodeToString(sum[:])+".claim")
}
