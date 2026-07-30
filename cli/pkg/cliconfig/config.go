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

	// ClusterContext caches the most recent /capi/app/detail response from
	// the per-user ControlHub BFF (see pkg/olares/id.go::ControlHubURL).
	// The `cluster context` command (cmd/ctl/cluster/context.go) writes
	// this on first run and on `--refresh`, and reads it for the no-network
	// fast path.
	//
	// IMPORTANT: this cache MUST NOT be consulted to decide whether a
	// `cluster ...` verb is allowed to run — the server is the only
	// authoritative source for "can this user list these pods?". It exists
	// purely so `cluster context` can render the user's identity / role /
	// accessible workspaces without a round-trip every invocation, and so
	// error-wrap helpers can include the cached role in their messages.
	// See skills/olares-cluster/SKILL.md for the security rationale.
	//
	// nil for pre-existing profiles or profiles where `cluster context`
	// has never been called.
	ClusterContext *ClusterContextCache `json:"clusterContext,omitempty"`

	// ClusterContextRefreshedAt is the unix-second timestamp of the last
	// successful /capi/app/detail call that wrote ClusterContext. Mirrors
	// WhoamiRefreshedAt; used by `cluster context` to render "last
	// refreshed" hints.
	ClusterContextRefreshedAt int64 `json:"clusterContextRefreshedAt,omitempty"`

	// BackendVersion caches the Olares OS version of the target instance,
	// read from /api/olares-info's osVersion. Command-side version
	// branching (cmdutil.Factory.OlaresBackendAtLeast) reads it to pick the
	// right version-specific implementation without a network round-trip on
	// every invocation.
	//
	// The cache is populated eagerly on `profile login` / `profile import`
	// and refreshed on demand (`profile whoami --refresh` / `profile list
	// --refresh`, or auto-fetched the first time a command needs the version
	// and the cache is empty). There is deliberately no TTL: a backend
	// upgrade is a rare, explicit event, so a stale value is corrected by the
	// user re-running with --refresh rather than by silently re-fetching on a
	// timer.
	//
	// Empty for pre-existing profiles or before the first version-aware
	// command runs. Treated as "unknown — detect on next use".
	BackendVersion string `json:"backendVersion,omitempty"`

	// BackendVersionRefreshedAt is the unix-second timestamp of the last
	// successful /api/olares-info read that wrote BackendVersion. Surfaced
	// for diagnostics ("last refreshed" hints); it does NOT drive any TTL.
	BackendVersionRefreshedAt int64 `json:"backendVersionRefreshedAt,omitempty"`

	// Location records where the CLI sits relative to this Olares instance:
	// one of pkg/olares.Location's wire values ("external" / "lan" / "host" /
	// "cluster"). It selects the connection method (URL scheme + host + DNS
	// resolver) at runtime — see pkg/olares.ID.Endpoints and pkg/access.
	//
	// Empty means "unknown — probe on next use": pre-existing profiles
	// created before this field existed, or a login/import where every probe
	// failed. An empty value triggers a one-off ProbeLocation backfill the
	// next time a command resolves this profile (cmdutil.Factory).
	Location string `json:"location,omitempty"`

	// LocationProbedAt is the unix-second timestamp of the last successful
	// ProbeLocation that wrote Location.
	LocationProbedAt int64 `json:"locationProbedAt,omitempty"`

	// LocationUnreachableAt is the unix-second timestamp of the most recent
	// "every probe failed" (access.ErrUnreachable) result. It acts as a
	// short cooldown so back-to-back commands during a network outage fail
	// fast instead of re-running the full (slow) probe each time. Cleared to
	// 0 by the next successful request or probe.
	LocationUnreachableAt int64 `json:"locationUnreachableAt,omitempty"`
}

// ClusterContextCache is the per-profile snapshot of /capi/app/detail
// (apps/packages/app/src/apps/controlPanelCommon/network/network.ts:222 —
// `AppDetailResponse`). Field names keep the wire JSON tags so callers can
// dump the cached struct verbatim when --output json without an extra
// mapping layer. Per-namespace ACL details (the SPA's ksConfig / config /
// globalRules maps) are intentionally NOT mirrored here: they're megabytes
// per profile and the CLI never needs to evaluate them locally.
type ClusterContextCache struct {
	// Username is the BFL username on the target Olares (usually equal to
	// the local part of OlaresID, but not enforced).
	Username string `json:"username,omitempty"`

	// GlobalRole is the KubeSphere global role this user holds, e.g.
	// "platform-admin" / "platform-self-provisioner" /
	// "platform-regular". Used purely for display in `cluster context`
	// output and in error-wrap hints — verb gating MUST defer to the
	// server (see ProfileConfig.ClusterContext doc).
	GlobalRole string `json:"globalrole,omitempty"`

	// Email is surfaced from /capi/app/detail's user.email. Populated
	// only when the BFL user record carries one; empty otherwise.
	Email string `json:"email,omitempty"`

	// Workspaces is the list of KubeSphere workspaces the user can see
	// according to the server. Used by `cluster context` listing only;
	// `cluster pod list -n <ns>` etc. still call the server directly.
	Workspaces []string `json:"workspaces,omitempty"`

	// SystemNamespaces is the list of system-owned namespaces the user
	// has visibility into (kube-system, os-system, ...). Surfaced
	// alongside Workspaces in `cluster context` output.
	SystemNamespaces []string `json:"systemNamespaces,omitempty"`

	// GrantedClusters is the list of multi-cluster identifiers the user
	// has grants on. Single-cluster Olares installs typically return
	// `["host"]`; we keep it here for forward-compat with multi-cluster
	// Olares deployments.
	GrantedClusters []string `json:"grantedClusters,omitempty"`

	// ClusterRole is the ks-installer-derived overall cluster role
	// (apps/packages/app/src/apps/controlPanelCommon/network/network.ts
	// `AppDetailResponse.clusterRole`). Surfaced for display only.
	ClusterRole string `json:"clusterRole,omitempty"`
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
	err = m.updateLockedInto(func(c *MultiProfileConfig) error {
		ch, e := c.applyOwnerRole(olaresID, role, refreshedAt)
		changed = ch
		return e
	})
	return changed, err
}

// applyOwnerRole mutates the in-memory role + timestamp for the profile keyed
// by olaresID WITHOUT persisting. It returns the same "changed" semantics as
// SetOwnerRole. Used under the config lock by SetOwnerRole and the batched
// SetDetectResults.
func (m *MultiProfileConfig) applyOwnerRole(olaresID, role string, refreshedAt int64) (bool, error) {
	target := m.FindByOlaresID(olaresID)
	if target == nil {
		return false, fmt.Errorf("profile %q not found", olaresID)
	}
	prev := target.OwnerRole
	target.OwnerRole = role
	target.WhoamiRefreshedAt = refreshedAt
	return role != "" && prev != role, nil
}

// SetClusterContext atomically updates the ClusterContext +
// ClusterContextRefreshedAt fields for the profile keyed by olaresID, then
// persists config.json. Mirrors SetOwnerRole's contract.
//
// Returns:
//   - changed: true iff GlobalRole transitioned to a different non-empty
//     value (used by `cluster context` to print a "role changed" notice).
//     A first-time write (cache was nil) also reports changed=true.
//   - err:     any I/O / serialization error from SaveMultiProfileConfig.
//
// `ctx` is the freshly-decoded snapshot to persist; passing nil is treated
// as an explicit "clear the cache" (used by tests / future eviction).
//
// IMPORTANT: this writes a snapshot of identity/role/visibility metadata
// only — the cache MUST NOT be consulted to decide whether a verb is
// allowed to run. See ProfileConfig.ClusterContext doc.
func (m *MultiProfileConfig) SetClusterContext(olaresID string, ctx *ClusterContextCache, refreshedAt int64) (changed bool, err error) {
	err = m.updateLockedInto(func(c *MultiProfileConfig) error {
		ch, e := c.applyClusterContext(olaresID, ctx, refreshedAt)
		changed = ch
		return e
	})
	return changed, err
}

// applyClusterContext mutates the in-memory cluster-context snapshot for the
// profile keyed by olaresID WITHOUT persisting. Same "changed" semantics as
// SetClusterContext. Used under the config lock.
func (m *MultiProfileConfig) applyClusterContext(olaresID string, ctx *ClusterContextCache, refreshedAt int64) (bool, error) {
	target := m.FindByOlaresID(olaresID)
	if target == nil {
		return false, fmt.Errorf("profile %q not found", olaresID)
	}
	prevRole := ""
	if target.ClusterContext != nil {
		prevRole = target.ClusterContext.GlobalRole
	}
	target.ClusterContext = ctx
	target.ClusterContextRefreshedAt = refreshedAt
	newRole := ""
	if ctx != nil {
		newRole = ctx.GlobalRole
	}
	return newRole != "" && prevRole != newRole, nil
}

// SetBackendVersion atomically updates the BackendVersion +
// BackendVersionRefreshedAt fields for the profile keyed by olaresID, then
// persists config.json. Mirrors SetOwnerRole's contract.
//
// Returns:
//   - changed: true iff BackendVersion transitioned to a different non-empty
//     value (callers can use this to notice "the backend was upgraded"). A
//     first-time write (empty → version) also reports changed=true.
//   - err:     any I/O / serialization error from SaveMultiProfileConfig.
//
// refreshedAt is the wall-clock the caller observed the /api/olares-info
// success at; passed in (rather than re-read here) so test code can pin it
// deterministically.
func (m *MultiProfileConfig) SetBackendVersion(olaresID, version string, refreshedAt int64) (changed bool, err error) {
	err = m.updateLockedInto(func(c *MultiProfileConfig) error {
		ch, e := c.applyBackendVersion(olaresID, version, refreshedAt)
		changed = ch
		return e
	})
	return changed, err
}

// applyBackendVersion mutates the in-memory version + timestamp for the
// profile keyed by olaresID WITHOUT persisting. Same "changed" semantics as
// SetBackendVersion. Used under the config lock by SetBackendVersion and the
// batched SetDetectResults.
func (m *MultiProfileConfig) applyBackendVersion(olaresID, version string, refreshedAt int64) (bool, error) {
	target := m.FindByOlaresID(olaresID)
	if target == nil {
		return false, fmt.Errorf("profile %q not found", olaresID)
	}
	prev := target.BackendVersion
	target.BackendVersion = version
	target.BackendVersionRefreshedAt = refreshedAt
	return version != "" && prev != version, nil
}

// SetLocation atomically updates the Location + LocationProbedAt fields for
// the profile keyed by olaresID and persists config.json. A successful probe
// implies the instance is reachable, so this also clears
// LocationUnreachableAt (resetting any outage cooldown).
//
// probedAt is the wall-clock the caller observed the probe success at;
// passed in (rather than re-read here) so tests can pin it deterministically.
func (m *MultiProfileConfig) SetLocation(olaresID, location string, probedAt int64) error {
	return m.updateLockedInto(func(c *MultiProfileConfig) error {
		return c.applyLocation(olaresID, location, probedAt)
	})
}

// applyLocation mutates the in-memory location fields (and clears the outage
// cooldown) for the profile keyed by olaresID WITHOUT persisting. Used under
// the config lock by SetLocation and the batched SetDetectResults.
func (m *MultiProfileConfig) applyLocation(olaresID, location string, probedAt int64) error {
	target := m.FindByOlaresID(olaresID)
	if target == nil {
		return fmt.Errorf("profile %q not found", olaresID)
	}
	target.Location = location
	target.LocationProbedAt = probedAt
	target.LocationUnreachableAt = 0
	return nil
}

// SetDetectResults atomically persists the outcome of a unified detect pass in
// a SINGLE locked config write: the location (always), plus the role and
// backend version when they were fetched this pass (empty = "not fetched",
// left untouched). Collapsing the three writes detect used to do into one
// avoids redundant load+save cycles and keeps the trio consistent on disk.
func (m *MultiProfileConfig) SetDetectResults(
	olaresID, location string, probedAt int64,
	role string, roleAt int64,
	version string, versionAt int64,
) error {
	return m.updateLockedInto(func(c *MultiProfileConfig) error {
		if err := c.applyLocation(olaresID, location, probedAt); err != nil {
			return err
		}
		if role != "" {
			if _, err := c.applyOwnerRole(olaresID, role, roleAt); err != nil {
				return err
			}
		}
		if version != "" {
			if _, err := c.applyBackendVersion(olaresID, version, versionAt); err != nil {
				return err
			}
		}
		return nil
	})
}

// SetLocationUnreachable stamps LocationUnreachableAt for the profile keyed
// by olaresID and persists config.json. The Location field is intentionally
// left unchanged — we keep the last-known-good value so a transient outage
// doesn't lose the previously-detected position (see the reprobe design).
func (m *MultiProfileConfig) SetLocationUnreachable(olaresID string, at int64) error {
	return m.updateLockedInto(func(c *MultiProfileConfig) error {
		target := c.FindByOlaresID(olaresID)
		if target == nil {
			return fmt.Errorf("profile %q not found", olaresID)
		}
		target.LocationUnreachableAt = at
		return nil
	})
}

// ClearLocationUnreachable resets LocationUnreachableAt to 0 (lifting the
// outage cooldown) and persists config.json. It is a no-op — including no
// disk write — when the field is already 0, so the common "request
// succeeded, nothing was in cooldown" path stays cheap.
func (m *MultiProfileConfig) ClearLocationUnreachable(olaresID string) error {
	return m.updateLockedInto(func(c *MultiProfileConfig) error {
		target := c.FindByOlaresID(olaresID)
		if target == nil {
			return fmt.Errorf("profile %q not found", olaresID)
		}
		if target.LocationUnreachableAt == 0 {
			return errNoConfigChange
		}
		target.LocationUnreachableAt = 0
		return nil
	})
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
