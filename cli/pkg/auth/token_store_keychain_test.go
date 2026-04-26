package auth

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/beclab/Olares/cli/internal/keychain"
	"github.com/beclab/Olares/cli/internal/keychain/keychainfake"
)

type staticLister []string

func (s staticLister) ListOlaresIDs() ([]string, error) { return []string(s), nil }

// TestKeychainStore_RoundTrip locks down the basic Get/Set/Delete contract:
// Set persists, Get round-trips, and Delete makes a subsequent Get return
// ErrTokenNotFound (not a backend error).
func TestKeychainStore_RoundTrip(t *testing.T) {
	kc := keychainfake.New()
	store := NewTokenStoreWith(kc, staticLister{"alice@olares.com"})

	if _, err := store.Get("alice@olares.com"); !errors.Is(err, ErrTokenNotFound) {
		t.Fatalf("Get on empty store = %v; want ErrTokenNotFound", err)
	}

	in := StoredToken{
		OlaresID:     "alice@olares.com",
		AccessToken:  "access-tok",
		RefreshToken: "refresh-tok",
		SessionID:    "sess",
		GrantedAt:    time.Now().UnixMilli(),
	}
	if err := store.Set(in); err != nil {
		t.Fatalf("Set: %v", err)
	}

	out, err := store.Get("alice@olares.com")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if out.AccessToken != in.AccessToken || out.RefreshToken != in.RefreshToken {
		t.Errorf("round-trip mismatch: got %+v, want %+v", out, in)
	}

	if err := store.Delete("alice@olares.com"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := store.Get("alice@olares.com"); !errors.Is(err, ErrTokenNotFound) {
		t.Fatalf("Get after Delete = %v; want ErrTokenNotFound", err)
	}
}

// TestKeychainStore_SetClearsInvalidatedAt locks down the invariant the
// Phase 1 fileStore had: a fresh Set() of a token with a non-zero
// InvalidatedAt must NOT carry that stamp through to the keychain entry,
// because Set() represents a successful fresh grant.
func TestKeychainStore_SetClearsInvalidatedAt(t *testing.T) {
	kc := keychainfake.New()
	store := NewTokenStoreWith(kc, staticLister{"alice@olares.com"})

	in := StoredToken{
		OlaresID:      "alice@olares.com",
		AccessToken:   "tok",
		InvalidatedAt: time.Now().UnixMilli(),
	}
	if err := store.Set(in); err != nil {
		t.Fatalf("Set: %v", err)
	}
	out, err := store.Get("alice@olares.com")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if out.InvalidatedAt != 0 {
		t.Errorf("InvalidatedAt = %d, want 0 (Set must clear)", out.InvalidatedAt)
	}
}

// TestKeychainStore_MarkInvalidated stamps the entry without touching other
// fields, and is preserved across a subsequent Get.
func TestKeychainStore_MarkInvalidated(t *testing.T) {
	kc := keychainfake.New()
	store := NewTokenStoreWith(kc, staticLister{"alice@olares.com"})

	if err := store.Set(StoredToken{
		OlaresID:    "alice@olares.com",
		AccessToken: "tok",
	}); err != nil {
		t.Fatalf("Set: %v", err)
	}

	at := time.Now()
	if err := store.MarkInvalidated("alice@olares.com", at); err != nil {
		t.Fatalf("MarkInvalidated: %v", err)
	}
	out, err := store.Get("alice@olares.com")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if out.AccessToken != "tok" {
		t.Errorf("AccessToken changed unexpectedly: %q", out.AccessToken)
	}
	if out.InvalidatedAt != at.UnixMilli() {
		t.Errorf("InvalidatedAt = %d, want %d", out.InvalidatedAt, at.UnixMilli())
	}

	// Marking a missing entry should report ErrTokenNotFound (NOT silently
	// create one), so callers can distinguish "no grant ever existed" from
	// "grant existed and was just stamped".
	if err := store.MarkInvalidated("ghost@olares.com", at); !errors.Is(err, ErrTokenNotFound) {
		t.Errorf("MarkInvalidated(missing) = %v; want ErrTokenNotFound", err)
	}
}

// TestKeychainStore_List_SkipsMissing exercises the contract that List()
// uses the ProfileLister as the index but tolerates missing keychain entries
// (a profile may exist before login). It also confirms profiles whose token
// IS present round-trip.
func TestKeychainStore_List_SkipsMissing(t *testing.T) {
	kc := keychainfake.New()
	store := NewTokenStoreWith(kc, staticLister{
		"alice@olares.com",
		"bob@olares.com",
	})

	if err := store.Set(StoredToken{OlaresID: "alice@olares.com", AccessToken: "a"}); err != nil {
		t.Fatalf("seed Set: %v", err)
	}

	got, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 1 || got[0].OlaresID != "alice@olares.com" {
		t.Fatalf("List = %+v, want exactly alice", got)
	}
}

// TestKeychainStore_List_TolerantToBadEntry verifies the resilience contract:
// a single profile whose stored blob is unreadable (decode error, transient
// keychain failure) must not abort `profile list` for everyone else. We
// inject one corrupted JSON blob and one per-account get error and expect
// the remaining healthy entry to come back, with one warning per casualty.
func TestKeychainStore_List_TolerantToBadEntry(t *testing.T) {
	prevSink := listWarnSink
	var warnings bytes.Buffer
	listWarnSink = &warnings
	defer func() { listWarnSink = prevSink }()

	kc := keychainfake.New()
	store := NewTokenStoreWith(kc, staticLister{
		"alice@olares.com",
		"bob@olares.com",
		"carol@olares.com",
	})

	if err := store.Set(StoredToken{OlaresID: "alice@olares.com", AccessToken: "a"}); err != nil {
		t.Fatalf("seed alice: %v", err)
	}
	// Bob's blob is present but undecodable JSON. This goes through Get's
	// json.Unmarshal path — exactly the corruption mode we care about for
	// the resilience invariant.
	kc.Data[kc.Key(keychain.OlaresCliService, "bob@olares.com")] = "{not-json"
	// Carol's read is gated by a transient keychain access denial (e.g.
	// the OS keychain refused to unlock that one slot). Still must not
	// abort the whole list.
	kc.PerAccountGetErr = map[string]error{
		"carol@olares.com": errors.New("keychain access denied"),
	}

	got, err := store.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 1 || got[0].OlaresID != "alice@olares.com" {
		t.Fatalf("List = %+v, want exactly alice", got)
	}
	out := warnings.String()
	if !strings.Contains(out, "bob@olares.com") || !strings.Contains(out, "carol@olares.com") {
		t.Errorf("expected warnings naming bob and carol, got:\n%s", out)
	}
	if strings.Count(out, "warning:") != 2 {
		t.Errorf("expected exactly 2 warnings, got:\n%s", out)
	}
}

// TestKeychainStore_UsesOlaresCliService asserts the (service, account) pair
// is the contractual one — service is fixed, account is the bare olaresId.
// If this ever drifts, every existing user re-loses their stored token.
func TestKeychainStore_UsesOlaresCliService(t *testing.T) {
	kc := keychainfake.New()
	store := NewTokenStoreWith(kc, staticLister{"alice@olares.com"})

	if err := store.Set(StoredToken{
		OlaresID:    "alice@olares.com",
		AccessToken: "tok",
	}); err != nil {
		t.Fatalf("Set: %v", err)
	}
	wantKey := keychain.OlaresCliService + "\x00alice@olares.com"
	if _, ok := kc.Data[wantKey]; !ok {
		t.Fatalf("expected keychain entry at %q; have %+v", wantKey, kc.Data)
	}
}
