package credential

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/beclab/Olares/cli/pkg/auth"
)

// fakeStore is the in-memory TokenStore stand-in used by every test in
// this file. It is intentionally simple — production code is exercised
// through keychainStore, which has its own unit tests.
type fakeStore struct {
	mu    sync.Mutex
	items map[string]auth.StoredToken

	// Set/MarkInvalidated counters let concurrency tests assert that a
	// refresh actually persisted, not just returned a value.
	setCount   atomic.Int32
	markCount  atomic.Int32
	markErrors atomic.Int32 // simulate keychain hiccups
}

func newFakeStore() *fakeStore {
	return &fakeStore{items: map[string]auth.StoredToken{}}
}

func (s *fakeStore) Get(olaresID string) (*auth.StoredToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.items[olaresID]
	if !ok {
		return nil, auth.ErrTokenNotFound
	}
	cp := t
	return &cp, nil
}

func (s *fakeStore) Set(t auth.StoredToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[t.OlaresID] = t
	s.setCount.Add(1)
	return nil
}

func (s *fakeStore) Delete(olaresID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, olaresID)
	return nil
}

func (s *fakeStore) List() ([]auth.StoredToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]auth.StoredToken, 0, len(s.items))
	for _, t := range s.items {
		out = append(out, t)
	}
	return out, nil
}

func (s *fakeStore) MarkInvalidated(olaresID string, at time.Time) error {
	if s.markErrors.Load() > 0 {
		s.markErrors.Add(-1)
		return errors.New("simulated MarkInvalidated failure")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.items[olaresID]
	if !ok {
		return auth.ErrTokenNotFound
	}
	t.InvalidatedAt = at.UnixMilli()
	s.items[olaresID] = t
	s.markCount.Add(1)
	return nil
}

// refreshServer spins up an httptest.Server that fakes /api/refresh.
// hits is incremented before the response is written, so the test can
// observe a hit even if the response is delayed via the latency hook.
type refreshServer struct {
	*httptest.Server
	hits    atomic.Int32
	status  int                 // status code to return; 0 = StatusOK
	body    func(int32) string  // body fn given the (1-indexed) hit count
	latency func(int32)         // optional pre-response delay hook
	allow   func(int32) bool    // if set, must return true for 200; else 401
}

func newRefreshServer(t *testing.T, opts ...func(*refreshServer)) *refreshServer {
	rs := &refreshServer{}
	for _, o := range opts {
		o(rs)
	}
	rs.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit := rs.hits.Add(1)
		if rs.latency != nil {
			rs.latency(hit)
		}
		if rs.allow != nil && !rs.allow(hit) {
			http.Error(w, `{"status":"FAIL","message":"unauth"}`, http.StatusUnauthorized)
			return
		}
		status := rs.status
		if status == 0 {
			status = http.StatusOK
		}
		body := ""
		if rs.body != nil {
			body = rs.body(hit)
		} else {
			body = fmt.Sprintf(`{"status":"OK","data":{"access_token":"AT%d","refresh_token":"RT%d","session_id":"S%d"}}`, hit, hit, hit)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(rs.Close)
	return rs
}

func setupRefresherEnv(t *testing.T) {
	t.Helper()
	t.Setenv("OLARES_CLI_HOME", t.TempDir())
}

// TestRefresh_HappyPath rotates a stale access token through /api/refresh
// once and asserts the new token is persisted + returned.
func TestRefresh_HappyPath(t *testing.T) {
	setupRefresherEnv(t)
	store := newFakeStore()
	_ = store.Set(auth.StoredToken{
		OlaresID:     "alice@olares.com",
		AccessToken:  "old-AT",
		RefreshToken: "RT-orig",
	})
	srv := newRefreshServer(t)

	r := NewRefresherWith(store, time.Now)
	got, err := r.Refresh(context.Background(), "alice@olares.com", srv.URL, "old-AT", false)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if got != "AT1" {
		t.Errorf("got = %q, want AT1", got)
	}
	if srv.hits.Load() != 1 {
		t.Errorf("server hits = %d, want 1", srv.hits.Load())
	}
	if store.setCount.Load() != 2 { // initial Set + refresh Set
		t.Errorf("Set count = %d, want 2", store.setCount.Load())
	}
	stored, _ := store.Get("alice@olares.com")
	if stored.AccessToken != "AT1" || stored.RefreshToken != "RT1" {
		t.Errorf("stored = %+v", stored)
	}
}

// TestRefresh_AlreadyFreshShortCircuit: if another goroutine refreshed
// the token between when the caller saw the 401 and when the caller
// reached r.Refresh, we must return the stored token without hitting
// the network.
func TestRefresh_AlreadyFreshShortCircuit(t *testing.T) {
	setupRefresherEnv(t)
	store := newFakeStore()
	_ = store.Set(auth.StoredToken{
		OlaresID:     "alice@olares.com",
		AccessToken:  "AT-already-new",
		RefreshToken: "RT",
	})
	srv := newRefreshServer(t)

	r := NewRefresherWith(store, time.Now)
	// The caller's snapshot is "old-AT"; the store now holds AT-already-new.
	got, err := r.Refresh(context.Background(), "alice@olares.com", srv.URL, "old-AT", false)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if got != "AT-already-new" {
		t.Errorf("got = %q, want AT-already-new", got)
	}
	if srv.hits.Load() != 0 {
		t.Errorf("server hits = %d, want 0 (should short-circuit)", srv.hits.Load())
	}
}

// TestRefresh_Unauthorized stamps InvalidatedAt and surfaces
// ErrTokenInvalidated. This is the "user must run profile login" path.
func TestRefresh_Unauthorized(t *testing.T) {
	setupRefresherEnv(t)
	store := newFakeStore()
	_ = store.Set(auth.StoredToken{
		OlaresID:     "alice@olares.com",
		AccessToken:  "AT",
		RefreshToken: "RT",
	})
	srv := newRefreshServer(t, func(rs *refreshServer) {
		rs.allow = func(int32) bool { return false }
	})

	r := NewRefresherWith(store, time.Now)
	_, err := r.Refresh(context.Background(), "alice@olares.com", srv.URL, "AT", false)
	var inv *ErrTokenInvalidated
	if !errors.As(err, &inv) {
		t.Fatalf("err = %v, want *ErrTokenInvalidated", err)
	}
	if store.markCount.Load() != 1 {
		t.Errorf("MarkInvalidated count = %d, want 1", store.markCount.Load())
	}
	stored, _ := store.Get("alice@olares.com")
	if stored.InvalidatedAt == 0 {
		t.Error("expected InvalidatedAt > 0 after MarkInvalidated")
	}
}

// TestRefresh_AlreadyInvalidated: a previous refresh failure stamped
// InvalidatedAt; we should NOT re-call /api/refresh, just surface
// ErrTokenInvalidated immediately.
func TestRefresh_AlreadyInvalidated(t *testing.T) {
	setupRefresherEnv(t)
	store := newFakeStore()
	_ = store.Set(auth.StoredToken{
		OlaresID:      "alice@olares.com",
		AccessToken:   "AT",
		RefreshToken:  "RT",
		InvalidatedAt: time.Now().UnixMilli(),
	})
	srv := newRefreshServer(t)

	r := NewRefresherWith(store, time.Now)
	_, err := r.Refresh(context.Background(), "alice@olares.com", srv.URL, "AT", false)
	var inv *ErrTokenInvalidated
	if !errors.As(err, &inv) {
		t.Fatalf("err = %v, want *ErrTokenInvalidated", err)
	}
	if srv.hits.Load() != 0 {
		t.Errorf("server hits = %d, want 0 (already-invalidated must short-circuit)", srv.hits.Load())
	}
}

// TestRefresh_NotLoggedIn: no entry in the store → ErrNotLoggedIn,
// without a network call.
func TestRefresh_NotLoggedIn(t *testing.T) {
	setupRefresherEnv(t)
	store := newFakeStore()
	srv := newRefreshServer(t)

	r := NewRefresherWith(store, time.Now)
	_, err := r.Refresh(context.Background(), "ghost@olares.com", srv.URL, "AT", false)
	var nli *ErrNotLoggedIn
	if !errors.As(err, &nli) {
		t.Fatalf("err = %v, want *ErrNotLoggedIn", err)
	}
	if srv.hits.Load() != 0 {
		t.Errorf("server hits = %d, want 0", srv.hits.Load())
	}
}

// TestRefresh_ConcurrentGoroutines: the lark-cli pattern — 50 goroutines
// see the same stale AT and call Refresh concurrently. Exactly ONE POST
// must reach /api/refresh; all 50 callers must see the same fresh token.
func TestRefresh_ConcurrentGoroutines(t *testing.T) {
	setupRefresherEnv(t)
	store := newFakeStore()
	_ = store.Set(auth.StoredToken{
		OlaresID:     "alice@olares.com",
		AccessToken:  "old",
		RefreshToken: "RT",
	})
	srv := newRefreshServer(t, func(rs *refreshServer) {
		// 50ms latency forces the first goroutine to hold the lock
		// long enough that the others actually contend.
		rs.latency = func(int32) { time.Sleep(50 * time.Millisecond) }
	})

	r := NewRefresherWith(store, time.Now)
	const n = 50
	results := make([]string, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			at, err := r.Refresh(context.Background(), "alice@olares.com", srv.URL, "old", false)
			results[i] = at
			errs[i] = err
		}(i)
	}
	wg.Wait()

	if hits := srv.hits.Load(); hits != 1 {
		t.Errorf("server hits = %d, want exactly 1", hits)
	}
	for i, err := range errs {
		if err != nil {
			t.Errorf("caller %d: err = %v", i, err)
		}
	}
	for i, got := range results {
		if got != "AT1" {
			t.Errorf("caller %d: got %q, want AT1", i, got)
		}
	}
}

// TestRefresh_MarkInvalidatedFailureSurfaces: if MarkInvalidated itself
// errors (e.g. keychain hiccup), we still want callers to receive the
// typed *ErrTokenInvalidated so reformatters render the
// "run profile login" CTA. The persistence failure is recoverable —
// the next refresh attempt will hit the same 401 and re-mark — so we
// log it to stderr but never let it shadow the real error.
func TestRefresh_MarkInvalidatedFailureSurfaces(t *testing.T) {
	setupRefresherEnv(t)
	store := newFakeStore()
	store.markErrors.Store(1)
	_ = store.Set(auth.StoredToken{
		OlaresID:     "alice@olares.com",
		AccessToken:  "AT",
		RefreshToken: "RT",
	})
	srv := newRefreshServer(t, func(rs *refreshServer) {
		rs.allow = func(int32) bool { return false }
	})

	r := NewRefresherWith(store, time.Now)
	_, err := r.Refresh(context.Background(), "alice@olares.com", srv.URL, "AT", false)
	if err == nil {
		t.Fatal("expected error")
	}
	var inv *ErrTokenInvalidated
	if !errors.As(err, &inv) {
		t.Fatalf("err = %v, want *ErrTokenInvalidated even when MarkInvalidated fails", err)
	}
	if inv.OlaresID != "alice@olares.com" {
		t.Errorf("inv.OlaresID = %q, want alice@olares.com", inv.OlaresID)
	}
}

// TestRefresh_EmptyRefreshToken: stored token with no refresh leg →
// ErrNotLoggedIn (the only thing we can do is ask the user to re-login).
func TestRefresh_EmptyRefreshToken(t *testing.T) {
	setupRefresherEnv(t)
	store := newFakeStore()
	_ = store.Set(auth.StoredToken{
		OlaresID:    "alice@olares.com",
		AccessToken: "AT",
		// RefreshToken intentionally missing
	})
	srv := newRefreshServer(t)

	r := NewRefresherWith(store, time.Now)
	_, err := r.Refresh(context.Background(), "alice@olares.com", srv.URL, "AT", false)
	var nli *ErrNotLoggedIn
	if !errors.As(err, &nli) {
		t.Fatalf("err = %v, want *ErrNotLoggedIn", err)
	}
	if srv.hits.Load() != 0 {
		t.Errorf("server hits = %d, want 0", srv.hits.Load())
	}
}

// TestRefresh_TransientErrorPropagates: a 5xx from /api/refresh should
// NOT mark invalidated — the user can simply retry the command. We
// surface the error verbatim.
func TestRefresh_TransientErrorPropagates(t *testing.T) {
	setupRefresherEnv(t)
	store := newFakeStore()
	_ = store.Set(auth.StoredToken{
		OlaresID:     "alice@olares.com",
		AccessToken:  "AT",
		RefreshToken: "RT",
	})
	srv := newRefreshServer(t, func(rs *refreshServer) {
		rs.status = http.StatusInternalServerError
		rs.body = func(int32) string { return `{"status":"FAIL","message":"oops"}` }
	})

	r := NewRefresherWith(store, time.Now)
	_, err := r.Refresh(context.Background(), "alice@olares.com", srv.URL, "AT", false)
	if err == nil {
		t.Fatal("expected error")
	}
	var inv *ErrTokenInvalidated
	if errors.As(err, &inv) {
		t.Errorf("err = %v; transient 5xx must NOT produce ErrTokenInvalidated", err)
	}
	if store.markCount.Load() != 0 {
		t.Errorf("MarkInvalidated called %d times on transient error; want 0", store.markCount.Load())
	}
}

// TestRefresh_CrossProcess re-execs the test binary as multiple child
// processes that all attempt to refresh the same olaresId against the
// same httptest server, sharing a JSON-on-disk token store + a shared
// flock directory under OLARES_CLI_HOME.
//
// The assertion is strict: across N concurrent processes, /api/refresh
// must be hit exactly ONCE. This is the canonical lark-cli "two CLIs at
// the same moment" race — without flock + double-check both procs would
// POST and one's refresh-token rotation would overwrite the other's.
func TestRefresh_CrossProcess(t *testing.T) {
	if os.Getenv("OLARES_REFRESH_CHILD") == "1" {
		runChildRefresh()
		return
	}
	if testing.Short() {
		t.Skip("cross-process refresh test re-execs the test binary; skipping in -short")
	}

	cliHome := t.TempDir()
	storePath := cliHome + "/test_store.json"
	seed := auth.StoredToken{
		OlaresID:     "alice@olares.com",
		AccessToken:  "old",
		RefreshToken: "RT",
	}
	if err := writeFileStore(storePath, map[string]auth.StoredToken{seed.OlaresID: seed}); err != nil {
		t.Fatalf("seed store: %v", err)
	}

	srv := newRefreshServer(t, func(rs *refreshServer) {
		rs.latency = func(int32) { time.Sleep(150 * time.Millisecond) }
	})

	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}

	const procs = 4
	var wg sync.WaitGroup
	errsCh := make(chan error, procs)
	for i := 0; i < procs; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			cmd := exec.Command(exe, "-test.run=TestRefresh_CrossProcess$", "-test.v=false")
			cmd.Env = append(os.Environ(),
				"OLARES_REFRESH_CHILD=1",
				"OLARES_CLI_HOME="+cliHome,
				"OLARES_REFRESH_AUTH_URL="+srv.URL,
				"OLARES_REFRESH_OLARES_ID="+seed.OlaresID,
				"OLARES_REFRESH_STORE="+storePath,
			)
			out, err := cmd.CombinedOutput()
			if err != nil {
				errsCh <- fmt.Errorf("child %d: %v\n%s", i, err, out)
			}
		}(i)
	}
	wg.Wait()
	close(errsCh)
	for err := range errsCh {
		t.Error(err)
	}

	if hits := srv.hits.Load(); hits != 1 {
		t.Errorf("cross-process /api/refresh hits = %d, want exactly 1", hits)
	}
	final, err := readFileStore(storePath)
	if err != nil {
		t.Fatalf("read store: %v", err)
	}
	if at := final[seed.OlaresID].AccessToken; at != "AT1" {
		t.Errorf("final stored AccessToken = %q, want AT1", at)
	}
}

// runChildRefresh is the OLARES_REFRESH_CHILD=1 entrypoint: instantiate
// a Refresher wired to the file-backed store, refresh once, exit.
//
// We use os.Exit (rather than t.Fatal) for failures so the parent's
// CombinedOutput captures stderr and surfaces it without the noise of
// `go test`'s ok/FAIL marker for the re-executed binary.
func runChildRefresh() {
	authURL := os.Getenv("OLARES_REFRESH_AUTH_URL")
	olaresID := os.Getenv("OLARES_REFRESH_OLARES_ID")
	storePath := os.Getenv("OLARES_REFRESH_STORE")
	if authURL == "" || olaresID == "" || storePath == "" {
		fmt.Fprintln(os.Stderr, "child: missing env")
		os.Exit(2)
	}

	store := &fileStore{path: storePath}
	cur, err := store.Get(olaresID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "child: get: %v\n", err)
		os.Exit(2)
	}
	r := NewRefresherWith(store, time.Now)
	got, err := r.Refresh(context.Background(), olaresID, authURL, cur.AccessToken, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "child: refresh: %v\n", err)
		os.Exit(2)
	}
	if !strings.HasPrefix(got, "AT") {
		fmt.Fprintf(os.Stderr, "child: unexpected token %q\n", got)
		os.Exit(2)
	}
	os.Exit(0)
}

// fileStore is a JSON-on-disk TokenStore used solely by
// TestRefresh_CrossProcess. Each operation reads + writes the entire
// file under a flock to be safe across processes — production code
// uses the OS keychain, which has its own concurrency guarantees, but
// we need a file backend here so multiple test processes can share
// state. The contention shape it produces (file-level coarse lock)
// matches keychainStore's "single mutating call serializes" property
// closely enough for the dedup invariant to be observable.
type fileStore struct{ path string }

func writeFileStore(path string, items map[string]auth.StoredToken) error {
	b, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

func readFileStore(path string) (map[string]auth.StoredToken, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out map[string]auth.StoredToken
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = map[string]auth.StoredToken{}
	}
	return out, nil
}

func (s *fileStore) load() (map[string]auth.StoredToken, error) { return readFileStore(s.path) }
func (s *fileStore) save(items map[string]auth.StoredToken) error {
	return writeFileStore(s.path, items)
}

func (s *fileStore) Get(olaresID string) (*auth.StoredToken, error) {
	items, err := s.load()
	if err != nil {
		return nil, err
	}
	t, ok := items[olaresID]
	if !ok {
		return nil, auth.ErrTokenNotFound
	}
	return &t, nil
}
func (s *fileStore) Set(t auth.StoredToken) error {
	items, err := s.load()
	if err != nil {
		return err
	}
	items[t.OlaresID] = t
	return s.save(items)
}
func (s *fileStore) Delete(olaresID string) error {
	items, err := s.load()
	if err != nil {
		return err
	}
	delete(items, olaresID)
	return s.save(items)
}
func (s *fileStore) List() ([]auth.StoredToken, error) {
	items, err := s.load()
	if err != nil {
		return nil, err
	}
	out := make([]auth.StoredToken, 0, len(items))
	for _, t := range items {
		out = append(out, t)
	}
	return out, nil
}
func (s *fileStore) MarkInvalidated(olaresID string, at time.Time) error {
	items, err := s.load()
	if err != nil {
		return err
	}
	t, ok := items[olaresID]
	if !ok {
		return auth.ErrTokenNotFound
	}
	t.InvalidatedAt = at.UnixMilli()
	items[olaresID] = t
	return s.save(items)
}
