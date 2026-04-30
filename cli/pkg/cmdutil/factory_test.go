package cmdutil

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/beclab/Olares/cli/pkg/auth"
	"github.com/beclab/Olares/cli/pkg/credential"
)

// fakeStore is a minimal in-memory TokenStore for transport tests.
// We deliberately re-declare it here (rather than reaching into the
// credential package's test helpers) so the cmdutil tests stay
// self-contained and the credential test fixtures can change shape
// without breaking transport tests.
type fakeStore struct {
	mu    sync.Mutex
	items map[string]auth.StoredToken
}

func (s *fakeStore) Get(id string) (*auth.StoredToken, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.items[id]
	if !ok {
		return nil, auth.ErrTokenNotFound
	}
	cp := t
	return &cp, nil
}
func (s *fakeStore) Set(t auth.StoredToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.items == nil {
		s.items = map[string]auth.StoredToken{}
	}
	s.items[t.OlaresID] = t
	return nil
}
func (s *fakeStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, id)
	return nil
}
func (s *fakeStore) List() ([]auth.StoredToken, error) { return nil, nil }
func (s *fakeStore) MarkInvalidated(id string, at time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.items[id]
	if !ok {
		return auth.ErrTokenNotFound
	}
	t.InvalidatedAt = at.UnixMilli()
	s.items[id] = t
	return nil
}

func newTransport(t *testing.T, store *fakeStore, authURL, accessToken string) (*refreshingTransport, *credential.Refresher) {
	t.Helper()
	t.Setenv("OLARES_CLI_HOME", t.TempDir())
	r := credential.NewRefresherWith(store, time.Now)
	tr := &refreshingTransport{
		base:      http.DefaultTransport,
		olaresID:  "alice@olares.com",
		authURL:   authURL,
		refresher: r,
		token:     &tokenCell{token: accessToken},
	}
	return tr, r
}

// TestRoundTrip_Success: 200 → no refresh attempt; X-Authorization is
// injected on the outbound request.
func TestRoundTrip_Success(t *testing.T) {
	var seenAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenAuth = r.Header.Get("X-Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	store := &fakeStore{}
	tr, _ := newTransport(t, store, "unused", "AT-1")

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	resp.Body.Close()
	if seenAuth != "AT-1" {
		t.Errorf("X-Authorization on first request = %q, want AT-1", seenAuth)
	}
}

// TestRoundTrip_RefreshOn401: 401 → call refresh → retry once with new
// token → final response makes it back to caller. Asserts the retry
// carries the NEW token (the whole point of the change).
func TestRoundTrip_RefreshOn401(t *testing.T) {
	var (
		hits     atomic.Int32
		seenAuth []string
	)
	muAuth := sync.Mutex{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit := hits.Add(1)
		muAuth.Lock()
		seenAuth = append(seenAuth, r.Header.Get("X-Authorization"))
		muAuth.Unlock()
		if hit == 1 {
			http.Error(w, "unauth", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	}))
	defer srv.Close()

	// /api/refresh server, owned by the refresher.
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"status":"OK","data":{"access_token":"AT-NEW","refresh_token":"RT","session_id":"S"}}`)
	}))
	defer authSrv.Close()

	store := &fakeStore{}
	_ = store.Set(auth.StoredToken{
		OlaresID:     "alice@olares.com",
		AccessToken:  "AT-OLD",
		RefreshToken: "RT",
	})
	tr, _ := newTransport(t, store, authSrv.URL, "AT-OLD")

	body := bytes.NewReader([]byte("payload"))
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL, body)
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("final status = %d, want 200", resp.StatusCode)
	}
	if hits.Load() != 2 {
		t.Errorf("upstream hits = %d, want 2 (one 401 + one retry)", hits.Load())
	}
	if len(seenAuth) != 2 || seenAuth[0] != "AT-OLD" || seenAuth[1] != "AT-NEW" {
		t.Errorf("X-Authorization timeline = %v, want [AT-OLD AT-NEW]", seenAuth)
	}
	if got := tr.token.snapshot(); got != "AT-NEW" {
		t.Errorf("token cell = %q, want AT-NEW (next request must use rotated token)", got)
	}
}

// TestRoundTrip_NoSecondRetry: if the retry ALSO returns 401, we must
// surface the second 401 verbatim — never loop. Two upstream hits, no
// refresh on the second 401.
func TestRoundTrip_NoSecondRetry(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		http.Error(w, "unauth", http.StatusUnauthorized)
	}))
	defer srv.Close()

	var refreshHits atomic.Int32
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		refreshHits.Add(1)
		_, _ = io.WriteString(w, `{"status":"OK","data":{"access_token":"AT-NEW","refresh_token":"RT"}}`)
	}))
	defer authSrv.Close()

	store := &fakeStore{}
	_ = store.Set(auth.StoredToken{
		OlaresID:     "alice@olares.com",
		AccessToken:  "AT-OLD",
		RefreshToken: "RT",
	})
	tr, _ := newTransport(t, store, authSrv.URL, "AT-OLD")

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401 surfaced verbatim", resp.StatusCode)
	}
	if hits.Load() != 2 {
		t.Errorf("upstream hits = %d, want exactly 2 (no infinite loop)", hits.Load())
	}
	if refreshHits.Load() != 1 {
		t.Errorf("refresh hits = %d, want 1", refreshHits.Load())
	}
}

// makeJWT builds a minimal "h.<payload>.s" JWT carrying just an exp
// claim — enough for auth.IsExpired/auth.ExpiresAt. Sig and header
// segments are placeholders since the CLI never verifies them client-
// side.
func makeJWT(t *testing.T, expUnix int64) string {
	t.Helper()
	payload := fmt.Sprintf(`{"exp":%d}`, expUnix)
	enc := base64.RawURLEncoding.EncodeToString([]byte(payload))
	return "h." + enc + ".s"
}

// TestRoundTrip_NonReplayableBody: a streaming body (req.GetBody == nil)
// with a non-JWT access_token must NOT trigger refresh. ExpiresAt fails
// to decode "AT" → preflight is skipped; the server's 401 is then
// surfaced verbatim because the body is unrewindable.
func TestRoundTrip_NonReplayableBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauth", http.StatusUnauthorized)
	}))
	defer srv.Close()

	var refreshHits atomic.Int32
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		refreshHits.Add(1)
		_, _ = io.WriteString(w, `{"status":"OK","data":{"access_token":"X","refresh_token":"Y"}}`)
	}))
	defer authSrv.Close()

	store := &fakeStore{}
	_ = store.Set(auth.StoredToken{
		OlaresID:     "alice@olares.com",
		AccessToken:  "AT",
		RefreshToken: "RT",
	})
	tr, _ := newTransport(t, store, authSrv.URL, "AT")

	// nopReader has no GetBody set; http.NewRequest only auto-sets
	// GetBody for known rewindable types (*bytes.Reader / Buffer /
	// strings.Reader). Mirrors the *os.File case in `files upload`.
	r := io.NopCloser(strings.NewReader("data"))
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL, r)
	req.GetBody = nil
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401 (non-replayable body must NOT trigger refresh+retry)", resp.StatusCode)
	}
	if refreshHits.Load() != 0 {
		t.Errorf("refresh hits = %d; refresh must not run for non-replayable bodies", refreshHits.Load())
	}
}

// TestRoundTrip_PreflightStaleNonReplayable: a streaming body whose
// access_token JWT is already past exp must trigger a pre-flight
// refresh BEFORE any byte hits the upstream. The upstream sees exactly
// one request, carrying the NEW token, and the body is read once.
//
// Without this, files-upload chunks backed by *os.File would 401 once
// per stale-token window and fail the whole upload command.
func TestRoundTrip_PreflightStaleNonReplayable(t *testing.T) {
	var (
		hits     atomic.Int32
		seenAuth atomic.Value
	)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		seenAuth.Store(r.Header.Get("X-Authorization"))
		body, _ := io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
	defer upstream.Close()

	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"status":"OK","data":{"access_token":"AT-NEW","refresh_token":"RT-NEW","session_id":"S"}}`)
	}))
	defer authSrv.Close()

	staleJWT := makeJWT(t, time.Now().Add(-1*time.Minute).Unix())
	store := &fakeStore{}
	_ = store.Set(auth.StoredToken{
		OlaresID:     "alice@olares.com",
		AccessToken:  staleJWT,
		RefreshToken: "RT",
	})
	tr, _ := newTransport(t, store, authSrv.URL, staleJWT)

	body := io.NopCloser(strings.NewReader("chunk-payload"))
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, upstream.URL, body)
	req.GetBody = nil
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if hits.Load() != 1 {
		t.Errorf("upstream hits = %d, want 1 (preflight must rotate before sending, not after a 401)", hits.Load())
	}
	if got := seenAuth.Load(); got != "AT-NEW" {
		t.Errorf("upstream X-Authorization = %v, want AT-NEW (preflight token wasn't propagated)", got)
	}
	if got := tr.token.snapshot(); got != "AT-NEW" {
		t.Errorf("token cell = %q, want AT-NEW", got)
	}
}

// TestRoundTrip_PreflightWithinSkew: exp is 30s in the future, well
// inside preflightSkew (60s). Even though the token is technically not
// expired, we rotate eagerly so a multi-second upload can't outlive it.
func TestRoundTrip_PreflightWithinSkew(t *testing.T) {
	var hits atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	var refreshHits atomic.Int32
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		refreshHits.Add(1)
		_, _ = io.WriteString(w, `{"status":"OK","data":{"access_token":"AT-NEW","refresh_token":"RT-NEW"}}`)
	}))
	defer authSrv.Close()

	soonJWT := makeJWT(t, time.Now().Add(30*time.Second).Unix())
	store := &fakeStore{}
	_ = store.Set(auth.StoredToken{
		OlaresID:     "alice@olares.com",
		AccessToken:  soonJWT,
		RefreshToken: "RT",
	})
	tr, _ := newTransport(t, store, authSrv.URL, soonJWT)

	body := io.NopCloser(strings.NewReader("x"))
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, upstream.URL, body)
	req.GetBody = nil
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	resp.Body.Close()
	if hits.Load() != 1 {
		t.Errorf("upstream hits = %d, want 1", hits.Load())
	}
	if refreshHits.Load() != 1 {
		t.Errorf("refresh hits = %d, want 1 (within preflightSkew window must trigger preflight)", refreshHits.Load())
	}
}

// TestRoundTrip_NoPreflightWhenFresh: exp 1h in the future is well
// outside preflightSkew. We must NOT decode + refresh per request; the
// reactive 401 path stays cheaper for the common case.
func TestRoundTrip_NoPreflightWhenFresh(t *testing.T) {
	var seenAuth atomic.Value
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenAuth.Store(r.Header.Get("X-Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	var refreshHits atomic.Int32
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		refreshHits.Add(1)
		_, _ = io.WriteString(w, `{"status":"OK","data":{"access_token":"AT-NEW","refresh_token":"RT-NEW"}}`)
	}))
	defer authSrv.Close()

	freshJWT := makeJWT(t, time.Now().Add(1*time.Hour).Unix())
	store := &fakeStore{}
	_ = store.Set(auth.StoredToken{
		OlaresID:     "alice@olares.com",
		AccessToken:  freshJWT,
		RefreshToken: "RT",
	})
	tr, _ := newTransport(t, store, authSrv.URL, freshJWT)

	body := io.NopCloser(strings.NewReader("x"))
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, upstream.URL, body)
	req.GetBody = nil
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	resp.Body.Close()
	if got := seenAuth.Load(); got != freshJWT {
		t.Errorf("upstream X-Authorization = %v, want the original (fresh) token", got)
	}
	if refreshHits.Load() != 0 {
		t.Errorf("refresh hits = %d, want 0 (fresh token must not preflight)", refreshHits.Load())
	}
}

// TestRoundTrip_NoPreflightOnReplayable: a stale JWT with a replayable
// body (*bytes.Reader) must skip preflight and rely on the reactive 401
// path. This keeps cheap JSON requests off the JWT-decode hot path.
func TestRoundTrip_NoPreflightOnReplayable(t *testing.T) {
	var hits atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit := hits.Add(1)
		if hit == 1 {
			http.Error(w, "unauth", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"status":"OK","data":{"access_token":"AT-NEW","refresh_token":"RT-NEW"}}`)
	}))
	defer authSrv.Close()

	staleJWT := makeJWT(t, time.Now().Add(-1*time.Minute).Unix())
	store := &fakeStore{}
	_ = store.Set(auth.StoredToken{
		OlaresID:     "alice@olares.com",
		AccessToken:  staleJWT,
		RefreshToken: "RT",
	})
	tr, _ := newTransport(t, store, authSrv.URL, staleJWT)

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, upstream.URL, bytes.NewReader([]byte("x")))
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	resp.Body.Close()
	if hits.Load() != 2 {
		t.Errorf("upstream hits = %d, want 2 (must take the reactive 401 path, not preflight)", hits.Load())
	}
}

// TestRoundTrip_PreflightFailureBubbles: when preflight refresh itself
// fails (server says the grant is dead), the transport must NOT consume
// the body — it has to return the typed error with the streaming body
// still intact, so the caller can render the "run profile login" CTA
// against an unread *os.File rather than half a chunk on the wire.
func TestRoundTrip_PreflightFailureBubbles(t *testing.T) {
	var upstreamHits atomic.Int32
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamHits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"status":"FAIL"}`, http.StatusUnauthorized)
	}))
	defer authSrv.Close()

	staleJWT := makeJWT(t, time.Now().Add(-1*time.Minute).Unix())
	store := &fakeStore{}
	_ = store.Set(auth.StoredToken{
		OlaresID:     "alice@olares.com",
		AccessToken:  staleJWT,
		RefreshToken: "RT",
	})
	tr, _ := newTransport(t, store, authSrv.URL, staleJWT)

	body := io.NopCloser(strings.NewReader("chunk"))
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, upstream.URL, body)
	req.GetBody = nil
	_, err := tr.RoundTrip(req)
	if err == nil {
		t.Fatal("expected error")
	}
	var inv *credential.ErrTokenInvalidated
	if !errors.As(err, &inv) {
		t.Errorf("err = %v, want *credential.ErrTokenInvalidated", err)
	}
	if upstreamHits.Load() != 0 {
		t.Errorf("upstream hits = %d, want 0 (preflight failure must not send the body)", upstreamHits.Load())
	}
}

// TestRoundTrip_RefreshFailureBubbles: refresh server returns 401 →
// refresher returns ErrTokenInvalidated → transport surfaces it (caller
// sees the "run profile login" CTA). The original upstream response is
// already drained; only the typed error reaches the caller.
func TestRoundTrip_RefreshFailureBubbles(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauth", http.StatusUnauthorized)
	}))
	defer srv.Close()

	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"status":"FAIL"}`, http.StatusUnauthorized)
	}))
	defer authSrv.Close()

	store := &fakeStore{}
	_ = store.Set(auth.StoredToken{
		OlaresID:     "alice@olares.com",
		AccessToken:  "AT",
		RefreshToken: "RT",
	})
	tr, _ := newTransport(t, store, authSrv.URL, "AT")

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	_, err := tr.RoundTrip(req)
	if err == nil {
		t.Fatal("expected error")
	}
	var inv *credential.ErrTokenInvalidated
	if !errors.As(err, &inv) {
		t.Errorf("err = %v, want *credential.ErrTokenInvalidated", err)
	}
}

// TestRoundTrip_403SameAs401: 403 must trigger refresh just like 401.
// Some Olares deployments emit 403 instead of 401 for an expired but
// otherwise-valid signature.
func TestRoundTrip_403SameAs401(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hits.Add(1) == 1 {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"status":"OK","data":{"access_token":"AT-NEW","refresh_token":"RT"}}`)
	}))
	defer authSrv.Close()

	store := &fakeStore{}
	_ = store.Set(auth.StoredToken{
		OlaresID:     "alice@olares.com",
		AccessToken:  "AT-OLD",
		RefreshToken: "RT",
	})
	tr, _ := newTransport(t, store, authSrv.URL, "AT-OLD")

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	resp, err := tr.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200 after 403→refresh→retry", resp.StatusCode)
	}
}

// TestCanRetry guards the canRetry table — these are the body-shape
// invariants the rest of the file relies on. If an http stdlib upgrade
// changes how GetBody is auto-populated, this test is the alarm.
func TestCanRetry(t *testing.T) {
	for _, tc := range []struct {
		name string
		body io.Reader
		want bool
	}{
		{"nil body", nil, true},
		{"bytes.Reader", bytes.NewReader([]byte("a")), true},
		{"bytes.Buffer", bytes.NewBuffer([]byte("a")), true},
		{"strings.Reader", strings.NewReader("a"), true},
		// io.NopCloser hides the underlying type from net/http, so
		// http.NewRequest can't set GetBody → not retryable.
		{"opaque reader", io.NopCloser(strings.NewReader("a")), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "http://x", tc.body)
			if err != nil {
				t.Fatal(err)
			}
			if got := canRetry(req); got != tc.want {
				t.Errorf("canRetry = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestTokenCell_RaceFree exercises the read+write paths under -race.
// Without the RWMutex this would flag immediately.
func TestTokenCell_RaceFree(t *testing.T) {
	c := &tokenCell{token: "init"}
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				if i%2 == 0 {
					c.update(fmt.Sprintf("v-%d-%d", i, j))
				} else {
					_ = c.snapshot()
				}
			}
		}(i)
	}
	wg.Wait()
}
