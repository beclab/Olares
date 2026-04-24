package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/beclab/Olares/cli/pkg/cliconfig"
)

// StoredToken is the per-olaresId record persisted to ~/.olares-cli/tokens.json
// during Phase 1.
//
// There is intentionally NO `ExpiresAt` field: AccessToken is a JWT and the
// only authoritative expiry comes from decoding its `exp` claim via
// auth.ExpiresAt. Mirroring the server's `expires_in` here would just create
// a second source of truth that can drift.
//
// RefreshToken is stored verbatim. It is not necessarily a JWT, so we never
// attempt to decode it.
//
// InvalidatedAt encodes server-side grant invalidation discovered by the
// client (e.g. /api/refresh returning 401/403). 0 means valid (or expiry
// has not yet been "discovered"); any value > 0 marks the entire grant
// (access_token + refresh_token) as unusable, even if the JWT's `exp`
// is still in the future. Phase 1 only DEFINES this field — no code path
// writes it. Phase 2's refreshWithLock is the writer. The only way to
// clear it back to 0 is a successful `profile login` / `profile import`
// (Set() defensively zeroes it).
type StoredToken struct {
	OlaresID      string `json:"olaresId"`
	AccessToken   string `json:"accessToken"`
	RefreshToken  string `json:"refreshToken,omitempty"`
	SessionID     string `json:"sessionId,omitempty"`
	GrantedAt     int64  `json:"grantedAt,omitempty"`     // unix milliseconds, audit-only
	InvalidatedAt int64  `json:"invalidatedAt,omitempty"` // unix milliseconds; 0 = valid
}

// tokensFile is the on-disk schema. Keyed by OlaresID for O(1) lookup; the
// nested OlaresID field on StoredToken is redundant but kept for self-describing
// dumps.
type tokensFile struct {
	Tokens map[string]StoredToken `json:"tokens"`
}

// TokenStore is the Phase 1 plaintext-JSON token backend. It is intentionally
// a tiny interface (Get/Set/Delete/List/MarkInvalidated) so that Phase 2 can
// swap in an OS keychain implementation behind the same surface.
//
// MarkInvalidated stamps an existing entry's InvalidatedAt without touching
// other fields. Returns ErrTokenNotFound if no entry exists for olaresID.
// Phase 2's refreshWithLock calls this when /api/refresh returns 401/403.
type TokenStore interface {
	Get(olaresID string) (*StoredToken, error)
	Set(token StoredToken) error
	Delete(olaresID string) error
	List() ([]StoredToken, error)
	MarkInvalidated(olaresID string, at time.Time) error
}

// ErrTokenNotFound is returned when no token is stored for a given olaresId.
var ErrTokenNotFound = errors.New("token not found")

// fileStore is the default plaintext-JSON implementation of TokenStore. Reads
// and writes are sequential (no concurrency); Phase 2 will add flock for
// cross-process safety together with keychain.
type fileStore struct {
	path string
}

// NewFileStore creates a TokenStore backed by ~/.olares-cli/tokens.json (or
// the override resolved by cliconfig.TokensFile()).
func NewFileStore() (TokenStore, error) {
	path, err := cliconfig.TokensFile()
	if err != nil {
		return nil, err
	}
	return &fileStore{path: path}, nil
}

// NewFileStoreAt is exposed for tests; production code should call NewFileStore.
func NewFileStoreAt(path string) TokenStore {
	return &fileStore{path: path}
}

func (s *fileStore) load() (*tokensFile, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &tokensFile{Tokens: map[string]StoredToken{}}, nil
		}
		return nil, fmt.Errorf("read %s: %w", s.path, err)
	}
	if len(data) == 0 {
		return &tokensFile{Tokens: map[string]StoredToken{}}, nil
	}
	tf := &tokensFile{}
	if err := json.Unmarshal(data, tf); err != nil {
		return nil, fmt.Errorf("parse %s: %w", s.path, err)
	}
	if tf.Tokens == nil {
		tf.Tokens = map[string]StoredToken{}
	}
	return tf, nil
}

func (s *fileStore) save(tf *tokensFile) error {
	if _, err := cliconfig.EnsureHome(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(tf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal tokens: %w", err)
	}
	return atomicWriteFile(s.path, data, 0o600)
}

func (s *fileStore) Get(olaresID string) (*StoredToken, error) {
	tf, err := s.load()
	if err != nil {
		return nil, err
	}
	tok, ok := tf.Tokens[olaresID]
	if !ok {
		return nil, ErrTokenNotFound
	}
	return &tok, nil
}

func (s *fileStore) Set(token StoredToken) error {
	if token.OlaresID == "" {
		return errors.New("StoredToken.OlaresID is required")
	}
	tf, err := s.load()
	if err != nil {
		return err
	}
	// Defensive: a fresh grant always supersedes any prior invalidation
	// stamp. Callers shouldn't be passing InvalidatedAt > 0 here, but if
	// they do (or if they forget to clear it when overwriting), normalize.
	token.InvalidatedAt = 0
	tf.Tokens[token.OlaresID] = token
	return s.save(tf)
}

func (s *fileStore) MarkInvalidated(olaresID string, at time.Time) error {
	tf, err := s.load()
	if err != nil {
		return err
	}
	tok, ok := tf.Tokens[olaresID]
	if !ok {
		return ErrTokenNotFound
	}
	tok.InvalidatedAt = at.UnixMilli()
	tf.Tokens[olaresID] = tok
	return s.save(tf)
}

func (s *fileStore) Delete(olaresID string) error {
	tf, err := s.load()
	if err != nil {
		return err
	}
	if _, ok := tf.Tokens[olaresID]; !ok {
		return ErrTokenNotFound
	}
	delete(tf.Tokens, olaresID)
	return s.save(tf)
}

func (s *fileStore) List() ([]StoredToken, error) {
	tf, err := s.load()
	if err != nil {
		return nil, err
	}
	out := make([]StoredToken, 0, len(tf.Tokens))
	for _, t := range tf.Tokens {
		out = append(out, t)
	}
	return out, nil
}

// atomicWriteFile mirrors cliconfig.atomicWriteFile but is duplicated here to
// avoid an exported helper just for cross-package use. Both implementations
// must stay in sync.
func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file in %s: %w", dir, err)
	}
	tmpName := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpName) }
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		cleanup()
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		cleanup()
		return fmt.Errorf("rename %s -> %s: %w", tmpName, path, err)
	}
	return nil
}
