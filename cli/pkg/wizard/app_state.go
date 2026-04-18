package wizard

import (
	"fmt"
	"log"
	"sync"
)

// AppState mirrors the relevant subset of the TS AppState
// (apps/packages/sdk/src/core/app.ts: class AppState). Persisted to
// the local Storage as kind="appstate", id="app-state-<did>".
//
// Only fields that we currently need on the CLI are tracked. Runtime-only
// state (storage handle, unlocked secrets) is annotated with json:"-".
type AppState struct {
	ID       string      `json:"id"` // "app-state-<did>"
	Device   *DeviceInfo `json:"device,omitempty"`
	Account  *Account    `json:"account,omitempty"`
	AuthInfo *AuthInfo   `json:"authInfo,omitempty"`
	Orgs     []Org       `json:"orgs,omitempty"`
	Vaults   []Vault     `json:"vaults,omitempty"`
	LastSync string      `json:"lastSync,omitempty"`

	// Session and unlocked secrets are intentionally NOT persisted.
	// A Session is bound to a specific server-side record (with its
	// HMAC key) and becomes invalid as soon as the process exits or
	// the server restarts. Persisting it would cause subsequent runs
	// to sign requests with a dead session id, which the server
	// rejects with [invalid_session]. They live in memory only and
	// must be re-established by Login() on every process start.
	Session  *Session         `json:"-"`
	storage  Storage          `json:"-"`
	unlocked *UnlockedAccount `json:"-"`
	mu       sync.Mutex       `json:"-"`
}

// AppStateKind is the kind used by the local Storage layer.
const AppStateKind = "appstate"

// NewAppState returns a fresh AppState bound to the given storage and DID.
// It does NOT load anything from disk — use LoadAppState for that.
func NewAppState(storage Storage, did string) *AppState {
	id := "app-state"
	if did != "" {
		id = "app-state-" + did
	}
	return &AppState{
		ID:      id,
		Device:  &DeviceInfo{ID: "cli-device-" + generateUUID(), Platform: "go-cli"},
		storage: storage,
	}
}

// LoadAppState attempts to load an existing state from storage; if none
// exists, returns a freshly-initialized AppState.
func LoadAppState(storage Storage, did string) (*AppState, error) {
	state := NewAppState(storage, did)
	if storage == nil {
		return state, nil
	}
	if err := storage.Get(AppStateKind, state.ID, state); err != nil {
		log.Printf("AppState %s not found in storage (creating new): %v", state.ID, err)
		// Reattach storage since Unmarshal does not touch unexported fields.
		state.storage = storage
		return state, nil
	}
	state.storage = storage
	if state.Device == nil {
		state.Device = &DeviceInfo{ID: "cli-device-" + generateUUID(), Platform: "go-cli"}
	}
	return state, nil
}

// Save persists the AppState to the storage. Vaults are serialized inline
// as part of the AppState (see the `Vaults` field above and TS
// AppState.vaults / App.saveState in apps/packages/sdk/src/core/app.ts),
// so we do NOT write per-vault files separately.
func (s *AppState) Save() error {
	if s.storage == nil {
		return nil
	}
	if err := s.storage.Put(AppStateKind, s.ID, s); err != nil {
		return fmt.Errorf("failed to save app state %s: %w", s.ID, err)
	}
	return nil
}

// ClientState interface implementation.

func (s *AppState) GetSession() *Session {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Session
}

func (s *AppState) SetSession(session *Session) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Session = session
}

func (s *AppState) GetAccount() *Account {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.Account
}

func (s *AppState) SetAccount(account *Account) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Account = account
}

func (s *AppState) GetDevice() *DeviceInfo {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Device == nil {
		s.Device = &DeviceInfo{ID: "cli-device-" + generateUUID(), Platform: "go-cli"}
	}
	return s.Device
}

// Unlocked returns the currently in-memory UnlockedAccount (or nil).
func (s *AppState) Unlocked() *UnlockedAccount {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.unlocked
}

// SetUnlocked stores the unlocked account in memory only.
func (s *AppState) SetUnlocked(u *UnlockedAccount) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.unlocked = u
}

// GetVault returns a pointer to the vault with the given id, or nil.
func (s *AppState) GetVault(id string) *Vault {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.Vaults {
		if s.Vaults[i].ID == id {
			return &s.Vaults[i]
		}
	}
	return nil
}

// PutVault inserts or replaces the vault with the same id.
func (s *AppState) PutVault(v Vault) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.Vaults {
		if s.Vaults[i].ID == v.ID {
			s.Vaults[i] = v
			return
		}
	}
	s.Vaults = append(s.Vaults, v)
}

// RemoveVault deletes the vault with the given id from in-memory state.
// The change is persisted next time Save() is called (vaults live inline
// inside the AppState record, so there is no separate file to delete).
func (s *AppState) RemoveVault(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := s.Vaults[:0]
	for _, v := range s.Vaults {
		if v.ID != id {
			out = append(out, v)
		}
	}
	s.Vaults = out
}
