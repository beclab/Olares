package credential

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/beclab/Olares/cli/pkg/auth"
)

const managedID = "alice@olares.com"

// A container starting for the first time has a mount and an empty keychain.
// That must produce a token rather than ErrNotLoggedIn.
func TestRefreshWith_ColdStartHasNothingStored(t *testing.T) {
	setupRefresherEnv(t)
	store := newFakeStore()
	srv := newRefreshServer(t)

	r := NewRefresherWith(store, time.Now)
	got, err := r.RefreshWith(context.Background(), managedID, srv.URL, "", "mounted-RT", false)
	if err != nil {
		t.Fatalf("RefreshWith: %v", err)
	}
	if got != "AT1" {
		t.Fatalf("got = %q, want AT1", got)
	}
	stored, err := store.Get(managedID)
	if err != nil {
		t.Fatalf("nothing persisted: %v", err)
	}
	if !stored.Managed {
		t.Error("stored entry should be marked managed")
	}
}

// The refresh token the server hands back is thrown away: the mount is the
// platform's copy and the only one it can revoke.
func TestRefreshWith_DiscardsReturnedRefreshToken(t *testing.T) {
	setupRefresherEnv(t)
	store := newFakeStore()
	srv := newRefreshServer(t)

	r := NewRefresherWith(store, time.Now)
	if _, err := r.RefreshWith(context.Background(), managedID, srv.URL, "", "mounted-RT", false); err != nil {
		t.Fatalf("RefreshWith: %v", err)
	}
	stored, _ := store.Get(managedID)
	if stored.RefreshToken != "" {
		t.Fatalf("stored refresh token = %q, want it discarded", stored.RefreshToken)
	}
}

// The mount, not the keychain, supplies the refresh token — even when the
// keychain holds a stale one from an earlier manual login that got taken over.
func TestRefreshWith_UsesMountedTokenNotStored(t *testing.T) {
	setupRefresherEnv(t)
	store := newFakeStore()
	_ = store.Set(auth.StoredToken{OlaresID: managedID, AccessToken: "old-AT", RefreshToken: "stale-RT"})

	var seen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var body struct {
			RefreshToken string `json:"refreshToken"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
		}
		seen = body.RefreshToken
		fmt.Fprint(w, `{"status":"OK","data":{"access_token":"AT1","refresh_token":"RT1"}}`)
	}))
	t.Cleanup(srv.Close)

	r := NewRefresherWith(store, time.Now)
	if _, err := r.RefreshWith(context.Background(), managedID, srv.URL, "old-AT", "mounted-RT", false); err != nil {
		t.Fatalf("RefreshWith: %v", err)
	}
	if seen != "mounted-RT" {
		t.Fatalf("server saw refresh token %q, want mounted-RT", seen)
	}
}

// A grant the platform already re-issued must not be blocked by the marker a
// previous failure left behind.
func TestRefreshWith_RetriesPastInvalidatedMarker(t *testing.T) {
	setupRefresherEnv(t)
	store := newFakeStore()
	_ = store.Set(auth.StoredToken{OlaresID: managedID, AccessToken: "old-AT", Managed: true})
	_ = store.MarkInvalidated(managedID, time.Now())
	srv := newRefreshServer(t)

	r := NewRefresherWith(store, time.Now)
	got, err := r.RefreshWith(context.Background(), managedID, srv.URL, "old-AT", "mounted-RT", false)
	if err != nil {
		t.Fatalf("RefreshWith: %v", err)
	}
	if got != "AT1" {
		t.Fatalf("got = %q, want AT1", got)
	}
	stored, _ := store.Get(managedID)
	if stored.InvalidatedAt != 0 {
		t.Fatalf("InvalidatedAt = %d, want cleared by the successful refresh", stored.InvalidatedAt)
	}
}

// 401 still means the grant is dead; the marker is stamped so `profile list`
// can report it.
func TestRefreshWith_UnauthorizedMarksInvalidated(t *testing.T) {
	setupRefresherEnv(t)
	store := newFakeStore()
	_ = store.Set(auth.StoredToken{OlaresID: managedID, AccessToken: "old-AT", Managed: true})
	srv := newRefreshServer(t, func(rs *refreshServer) {
		rs.allow = func(int32) bool { return false }
	})

	r := NewRefresherWith(store, time.Now)
	_, err := r.RefreshWith(context.Background(), managedID, srv.URL, "old-AT", "mounted-RT", false)
	var invalidated *ErrTokenInvalidated
	if !errors.As(err, &invalidated) {
		t.Fatalf("err = %v, want ErrTokenInvalidated", err)
	}
	if store.markCount.Load() != 1 {
		t.Fatalf("MarkInvalidated count = %d, want 1", store.markCount.Load())
	}
}

// A 5xx is transient: no marker, so the next command tries again.
func TestRefreshWith_ServerErrorDoesNotMarkInvalidated(t *testing.T) {
	setupRefresherEnv(t)
	store := newFakeStore()
	_ = store.Set(auth.StoredToken{OlaresID: managedID, AccessToken: "old-AT", Managed: true})
	srv := newRefreshServer(t, func(rs *refreshServer) {
		rs.status = http.StatusInternalServerError
		rs.body = func(int32) string { return `{"status":"FAIL"}` }
	})

	r := NewRefresherWith(store, time.Now)
	_, err := r.RefreshWith(context.Background(), managedID, srv.URL, "old-AT", "mounted-RT", false)
	if err == nil {
		t.Fatal("want an error")
	}
	var invalidated *ErrTokenInvalidated
	if errors.As(err, &invalidated) {
		t.Fatal("a 5xx must not be reported as an invalidated grant")
	}
	if store.markCount.Load() != 0 {
		t.Fatalf("MarkInvalidated count = %d, want 0", store.markCount.Load())
	}
}

// The platform promises second-factor grants. A first-factor token reaches no
// per-service host, so it is refused instead of persisted.
func TestRefreshWith_RejectsFirstFactorToken(t *testing.T) {
	setupRefresherEnv(t)
	store := newFakeStore()
	srv := newRefreshServer(t, func(rs *refreshServer) {
		rs.body = func(int32) string {
			return `{"status":"OK","data":{"access_token":"AT1","refresh_token":"RT1","fa2":true}}`
		}
	})

	r := NewRefresherWith(store, time.Now)
	_, err := r.RefreshWith(context.Background(), managedID, srv.URL, "", "mounted-RT", false)
	var notSecondFactor *ErrManagedNotSecondFactor
	if !errors.As(err, &notSecondFactor) {
		t.Fatalf("err = %v, want ErrManagedNotSecondFactor", err)
	}
	if _, err := store.Get(managedID); !errors.Is(err, auth.ErrTokenNotFound) {
		t.Fatal("a first-factor token must not be persisted")
	}
}

func TestRefreshWith_RequiresRefreshToken(t *testing.T) {
	setupRefresherEnv(t)
	r := NewRefresherWith(newFakeStore(), time.Now)
	if _, err := r.RefreshWith(context.Background(), managedID, "https://auth.example", "", "", false); err == nil {
		t.Fatal("want an error when no refresh token is supplied")
	}
}

// The HTTP transport calls the ordinary Refresh when a token expires
// mid-command. For a managed entry there is no stored refresh token to rotate
// with, so it has to go back to the mount — otherwise a container works
// exactly until its first access token expires and then reports "not logged
// in" for an account that cannot be logged into.
func TestRefresh_ManagedEntryFallsBackToTheMount(t *testing.T) {
	setupRefresherEnv(t)
	store := newFakeStore()
	_ = store.Set(auth.StoredToken{OlaresID: managedID, AccessToken: "old-AT", Managed: true})

	var seen string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var body struct {
			RefreshToken string `json:"refreshToken"`
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
		}
		seen = body.RefreshToken
		fmt.Fprint(w, `{"status":"OK","data":{"access_token":"AT1","refresh_token":"RT1"}}`)
	}))
	t.Cleanup(srv.Close)

	r := NewRefresherWith(store, time.Now)
	r.loadManaged = func() (*ManagedCredential, bool) {
		return &ManagedCredential{RefreshToken: "mounted-RT", OlaresID: managedID, AppName: "lares"}, true
	}

	got, err := r.Refresh(context.Background(), managedID, srv.URL, "old-AT", false)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if got != "AT1" {
		t.Errorf("got = %q, want AT1", got)
	}
	if seen != "mounted-RT" {
		t.Errorf("server saw refresh token %q, want the mounted one", seen)
	}
	stored, _ := store.Get(managedID)
	if stored.RefreshToken != "" || !stored.Managed {
		t.Errorf("stored = %+v, want it to stay a managed shell", stored)
	}
}

// With the mount gone there is nothing to rotate with, and saying so is more
// use than sending a request that will 401.
func TestRefresh_ManagedEntryWithoutAMount(t *testing.T) {
	setupRefresherEnv(t)
	store := newFakeStore()
	_ = store.Set(auth.StoredToken{OlaresID: managedID, AccessToken: "old-AT", Managed: true})
	srv := newRefreshServer(t)

	r := NewRefresherWith(store, time.Now)
	r.loadManaged = func() (*ManagedCredential, bool) { return nil, false }

	_, err := r.Refresh(context.Background(), managedID, srv.URL, "old-AT", false)
	var notLoggedIn *ErrNotLoggedIn
	if !errors.As(err, &notLoggedIn) {
		t.Fatalf("err = %v, want ErrNotLoggedIn", err)
	}
	if srv.hits.Load() != 0 {
		t.Errorf("server hits = %d, want 0", srv.hits.Load())
	}
}

// The non-managed path keeps rotating and storing the refresh token it is
// given; nothing about the extraction changed it.
func TestRefresh_StillRotatesRefreshTokenAndIsNotManaged(t *testing.T) {
	setupRefresherEnv(t)
	store := newFakeStore()
	_ = store.Set(auth.StoredToken{OlaresID: managedID, AccessToken: "old-AT", RefreshToken: "RT-orig"})
	srv := newRefreshServer(t)

	r := NewRefresherWith(store, time.Now)
	if _, err := r.Refresh(context.Background(), managedID, srv.URL, "old-AT", false); err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	stored, _ := store.Get(managedID)
	if stored.RefreshToken != "RT1" {
		t.Errorf("stored refresh token = %q, want RT1", stored.RefreshToken)
	}
	if stored.Managed {
		t.Error("an ordinary refresh must not mark the entry managed")
	}
}
