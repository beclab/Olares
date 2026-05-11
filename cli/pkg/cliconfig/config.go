package cliconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/beclab/Olares/cli/pkg/olares"
)

// MultiProfileConfig is the on-disk schema of ~/.olares-cli/config.json.
// It tracks all known profiles plus which one is currently active, mirroring
// lark-cli's MultiAppConfig but stripped of brand / strict-mode / multi-app
// concerns (see docs/notes/olares-cli-auth-profile-config.md §11).
type MultiProfileConfig struct {
	CurrentProfile  string          `json:"currentProfile,omitempty"`
	PreviousProfile string          `json:"previousProfile,omitempty"`
	Profiles        []ProfileConfig `json:"profiles,omitempty"`
}

// ProfileConfig is a single profile entry: a target Olares instance + the
// user identity used to talk to it. The primary key is OlaresID; Name is an
// optional alias users can pass to commands like `profile use <name>`.
//
// Tokens are NOT stored in this file — they live in the OS keychain
// (one entry per OlaresID; see cli/internal/keychain and pkg/auth.TokenStore).
type ProfileConfig struct {
	// Name is an optional human-friendly alias. If empty, OlaresID is used as
	// the display name.
	Name string `json:"name,omitempty"`

	// OlaresID is the canonical user identity, e.g. "alice@olares.com".
	// All per-user URLs (auth / vault / desktop) are derived from it.
	OlaresID string `json:"olaresId"`

	// UserUID is optionally populated after login for diagnostics. Never used
	// as an authoritative identity (see §7.5 of the design doc on JWT trust).
	UserUID string `json:"userUid,omitempty"`

	// AuthURLOverride bypasses the standard URL derivation. Used for dev /
	// internal environments. Leave empty in production.
	AuthURLOverride string `json:"authUrlOverride,omitempty"`

	// LocalURLPrefix is inserted between the service subdomain and the
	// terminus name when deriving URLs (e.g. "dev." → "auth.dev.alice.olares.com").
	// Leave empty in production.
	LocalURLPrefix string `json:"localUrlPrefix,omitempty"`

	// InsecureSkipVerify disables TLS verification for HTTP calls under this
	// profile. Dev only.
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`

	// OwnerRole is the role this user has on the target Olares, populated
	// best-effort from /api/backend/v1/user-info on login / import / whoami.
	// One of the BFL wire constants from
	// framework/bfl/pkg/constants/constants.go:
	//
	//	"owner"   — the (single) instance owner; full privileges.
	//	"admin"   — has all privileges except managing other admins and a
	//	            handful of hardware/restart-class operations.
	//	"normal"  — least-privileged user; the SPA labels this "User" in
	//	            the UI. We keep the wire value verbatim and translate
	//	            to the friendly label only in printed output.
	//
	// Empty when unknown — pre-existing profiles created before this field
	// existed, or first-run before whoami has been called. Empty MUST be
	// treated as "skip preflight, let the server decide" so older configs
	// keep working without a forced re-login.
	//
	// Stored on ProfileConfig (rather than a separate identity.json or the
	// keychain) because it's not a secret, it's stable across re-logins for
	// the same olaresId, and it travels naturally with the rest of the
	// profile when users back up / migrate ~/.olares-cli/.
	OwnerRole string `json:"ownerRole,omitempty"`

	// WhoamiRefreshedAt is the unix-second timestamp of the last successful
	// /api/backend/v1/user-info call that wrote OwnerRole. Used by `profile
	// whoami` (and the `settings users me` / `settings me whoami` aliases)
	// to render "last refreshed" hints, and by Phase 1+ preflight code to
	// decide whether the cache is stale enough to refetch silently.
	WhoamiRefreshedAt int64 `json:"whoamiRefreshedAt,omitempty"`
}

// DisplayName returns Name if set, else OlaresID. Used in CLI output where we
// want a stable handle for the profile.
func (p *ProfileConfig) DisplayName() string {
	if p.Name != "" {
		return p.Name
	}
	return p.OlaresID
}

// ResolvedAuthURL returns the auth URL the CLI should hit for this profile,
// honoring AuthURLOverride when set.
func (p *ProfileConfig) ResolvedAuthURL() (string, error) {
	if p.AuthURLOverride != "" {
		return p.AuthURLOverride, nil
	}
	id, err := olares.ParseID(p.OlaresID)
	if err != nil {
		return "", err
	}
	return id.AuthURL(p.LocalURLPrefix), nil
}

// FindProfile looks up a profile by Name first, then OlaresID. Mirrors
// lark-cli's MultiAppConfig.FindApp lookup order so that aliases shadow raw
// IDs in command-line UX. Returns nil if no match.
func (m *MultiProfileConfig) FindProfile(key string) *ProfileConfig {
	if key == "" {
		return nil
	}
	for i := range m.Profiles {
		if m.Profiles[i].Name == key {
			return &m.Profiles[i]
		}
	}
	for i := range m.Profiles {
		if m.Profiles[i].OlaresID == key {
			return &m.Profiles[i]
		}
	}
	return nil
}

// FindByOlaresID is a strict OlaresID lookup. Used by login / import flows
// where we explicitly want to detect "same olaresId already exists".
func (m *MultiProfileConfig) FindByOlaresID(olaresID string) *ProfileConfig {
	if olaresID == "" {
		return nil
	}
	for i := range m.Profiles {
		if m.Profiles[i].OlaresID == olaresID {
			return &m.Profiles[i]
		}
	}
	return nil
}

// Current returns the active profile, or nil if there isn't one (no profiles
// at all, or CurrentProfile pointing at a stale entry).
func (m *MultiProfileConfig) Current() *ProfileConfig {
	if m.CurrentProfile == "" {
		if len(m.Profiles) == 0 {
			return nil
		}
		return &m.Profiles[0]
	}
	return m.FindProfile(m.CurrentProfile)
}

// Upsert inserts or replaces a profile by OlaresID. If a profile with the
// same OlaresID exists its slot is overwritten in place (preserving order);
// otherwise the new profile is appended. Returns the (possibly newly inserted)
// profile in its persisted slot.
func (m *MultiProfileConfig) Upsert(p ProfileConfig) *ProfileConfig {
	for i := range m.Profiles {
		if m.Profiles[i].OlaresID == p.OlaresID {
			m.Profiles[i] = p
			return &m.Profiles[i]
		}
	}
	m.Profiles = append(m.Profiles, p)
	return &m.Profiles[len(m.Profiles)-1]
}

// SetOwnerRole atomically updates the OwnerRole + WhoamiRefreshedAt fields
// for the profile keyed by olaresID, then persists config.json.
//
// Returns:
//   - changed: true iff OwnerRole transitioned to a different non-empty
//     value (used by callers to decide whether to print a "role changed"
//     notice). A first-time write (empty → role) also reports changed=true
//     because that's a new piece of information from the user's perspective.
//   - err:     any I/O / serialization error from SaveMultiProfileConfig.
//
// If no profile matches olaresID we return (false, error) — callers should
// surface this rather than silently writing nothing, because every code
// path here only runs after a successful API call against that olaresID.
//
// refreshedAt is the wall-clock the caller observed the API success at;
// passed in (rather than re-read here) so test code and replay-style
// flows can pin it deterministically.
func (m *MultiProfileConfig) SetOwnerRole(olaresID, role string, refreshedAt int64) (changed bool, err error) {
	target := m.FindByOlaresID(olaresID)
	if target == nil {
		return false, fmt.Errorf("profile %q not found", olaresID)
	}
	prev := target.OwnerRole
	target.OwnerRole = role
	target.WhoamiRefreshedAt = refreshedAt
	if err := SaveMultiProfileConfig(m); err != nil {
		return false, fmt.Errorf("save config: %w", err)
	}
	return role != "" && prev != role, nil
}

// Remove deletes a profile by Name or OlaresID. If the removed profile was
// the current one, CurrentProfile is repointed to PreviousProfile (if still
// valid) or to the first remaining profile. Returns the removed profile and a
// boolean indicating whether anything was deleted.
func (m *MultiProfileConfig) Remove(key string) (*ProfileConfig, bool) {
	idx := -1
	for i := range m.Profiles {
		if m.Profiles[i].Name == key || m.Profiles[i].OlaresID == key {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil, false
	}
	removed := m.Profiles[idx]
	m.Profiles = append(m.Profiles[:idx], m.Profiles[idx+1:]...)

	wasCurrent := m.CurrentProfile == removed.Name || m.CurrentProfile == removed.OlaresID
	wasPrevious := m.PreviousProfile == removed.Name || m.PreviousProfile == removed.OlaresID
	if wasPrevious {
		m.PreviousProfile = ""
	}
	if wasCurrent {
		// Prefer falling back to PreviousProfile if it's still a valid entry,
		// otherwise to whatever ended up first in the slice.
		switch {
		case m.PreviousProfile != "" && m.FindProfile(m.PreviousProfile) != nil:
			m.CurrentProfile = m.PreviousProfile
			m.PreviousProfile = ""
		case len(m.Profiles) > 0:
			m.CurrentProfile = m.Profiles[0].DisplayName()
		default:
			m.CurrentProfile = ""
		}
	}
	return &removed, true
}

// SetCurrent flips CurrentProfile / PreviousProfile, resolving "-" to the
// previous profile (a la `cd -`). Returns the newly-current profile.
func (m *MultiProfileConfig) SetCurrent(key string) (*ProfileConfig, error) {
	if key == "-" {
		if m.PreviousProfile == "" {
			return nil, errors.New("no previous profile to switch back to")
		}
		key = m.PreviousProfile
	}
	target := m.FindProfile(key)
	if target == nil {
		return nil, fmt.Errorf("profile %q not found", key)
	}
	newCurrent := target.DisplayName()
	if m.CurrentProfile != newCurrent {
		m.PreviousProfile = m.CurrentProfile
		m.CurrentProfile = newCurrent
	}
	return target, nil
}

// LoadMultiProfileConfig reads config.json from disk. A missing file yields
// an empty config (not an error) so first-run UX works.
func LoadMultiProfileConfig() (*MultiProfileConfig, error) {
	path, err := ConfigFile()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &MultiProfileConfig{}, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if len(data) == 0 {
		return &MultiProfileConfig{}, nil
	}
	cfg := &MultiProfileConfig{}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return cfg, nil
}

// SaveMultiProfileConfig writes config.json atomically with 0600 perms,
// creating the parent directory if needed.
func SaveMultiProfileConfig(cfg *MultiProfileConfig) error {
	dir, err := EnsureHome()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, configFilename)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return atomicWriteFile(path, data, filePerm)
}

// atomicWriteFile writes data to path via a temp file + rename, mirroring the
// safety pattern in lark-cli's core.SaveMultiAppConfig.
func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file in %s: %w", dir, err)
	}
	tmpName := tmp.Name()
	cleanup := func() {
		_ = os.Remove(tmpName)
	}
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
