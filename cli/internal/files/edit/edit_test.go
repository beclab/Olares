package edit

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestClient mirrors the rm / cp / rename / mkdir test
// harnesses: stand up a real httptest server, hand the caller a
// Client whose BaseURL points at it, and let the test inspect
// what landed on the wire.
func newTestClient(t *testing.T, h http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &Client{
		HTTPClient: srv.Client(),
		BaseURL:    srv.URL,
	}, srv
}

// TestPlan_FileEdit is the canonical case: edit a file in place.
// Endpoint must be the per-resource path with NO trailing slash
// (a trailing slash would route the PUT through the directory
// handler).
func TestPlan_FileEdit(t *testing.T) {
	tgt := Target{FileType: "drive", Extend: "Home", SubPath: "/Documents/notes.md"}
	op, err := Plan(tgt)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if op.Endpoint != "/api/resources/drive/Home/Documents/notes.md" {
		t.Errorf("Endpoint: got %q", op.Endpoint)
	}
	if strings.HasSuffix(op.Endpoint, "/") {
		t.Errorf("Endpoint must NOT end with '/' for files: %q", op.Endpoint)
	}
	if op.DisplayPath != "drive/Home/Documents/notes.md" {
		t.Errorf("DisplayPath: got %q", op.DisplayPath)
	}
}

// TestPlan_PercentEncoding: the wire URL must use the same
// JS-shaped encoding the rest of the CLI uses (encodepath.EncodeURL).
// Verify a path with spaces and unicode survives unchanged on the
// display side and is percent-encoded on the wire side.
func TestPlan_PercentEncoding(t *testing.T) {
	tgt := Target{FileType: "drive", Extend: "Home", SubPath: "/My Docs/老照片.txt"}
	op, err := Plan(tgt)
	if err != nil {
		t.Fatalf("Plan: %v", err)
	}
	if !strings.Contains(op.Endpoint, "/api/resources/drive/Home/My%20Docs/") {
		t.Errorf("path encoding: got %q", op.Endpoint)
	}
	// Display path keeps the human-readable form (so log lines
	// don't show users a URL-encoded mess).
	if op.DisplayPath != "drive/Home/My Docs/老照片.txt" {
		t.Errorf("DisplayPath: got %q", op.DisplayPath)
	}
}

// TestPlan_RejectsBadInput table-drives the input-validation
// contract. Each must produce an error that points at the
// offending input — error wording is part of the UX so we assert
// on substrings, not exact strings.
func TestPlan_RejectsBadInput(t *testing.T) {
	cases := []struct {
		name   string
		tgt    Target
		expect string // substring that must appear in the error
	}{
		{
			name:   "root of volume",
			tgt:    Target{FileType: "drive", Extend: "Home", SubPath: "/"},
			expect: "root of",
		},
		{
			name:   "root via empty SubPath",
			tgt:    Target{FileType: "drive", Extend: "Home", SubPath: ""},
			expect: "root of",
		},
		{
			name:   "trailing slash (directory path)",
			tgt:    Target{FileType: "drive", Extend: "Home", SubPath: "/Documents/"},
			expect: "directory path",
		},
		{
			name:   "single-dot segment",
			tgt:    Target{FileType: "drive", Extend: "Home", SubPath: "/foo/./bar"},
			expect: "path-traversal",
		},
		{
			name:   "double-dot segment",
			tgt:    Target{FileType: "drive", Extend: "Home", SubPath: "/foo/../bar"},
			expect: "path-traversal",
		},
		{
			name:   "empty fileType",
			tgt:    Target{FileType: "", Extend: "Home", SubPath: "/foo"},
			expect: "empty fileType",
		},
		{
			name:   "empty extend",
			tgt:    Target{FileType: "drive", Extend: "", SubPath: "/foo"},
			expect: "empty fileType",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := Plan(c.tgt)
			if err == nil {
				t.Fatalf("Plan(%+v): want error containing %q, got nil", c.tgt, c.expect)
			}
			if !strings.Contains(err.Error(), c.expect) {
				t.Errorf("Plan(%+v): error %q does not contain %q", c.tgt, err.Error(), c.expect)
			}
		})
	}
}

// TestPlan_NamespaceAllowlist locks in the allow-list from the
// package docstring:
//
//   - drive / sync / cache / external are supported (the LarePass
//     GUI's `onSaveFile` natively wires these to /api/resources
//     PUT).
//   - awss3 / google / dropbox / tencent are NOT supported. The
//     read leg is fine now — the unified /api/raw/ endpoint
//     serves raw bytes on cloud drives too — but the write leg
//     (PUT /api/resources/<cloud-path>) is still unverified per
//     driver: only awss3/utils.ts has a put() helper at all in
//     the GUI; google / dropbox / tencent have no save-related
//     plumbing, so PUT-ing there would hit an endpoint nobody
//     has exercised end-to-end. That gap needs to close before
//     we re-enable cloud-drive editing.
//   - share / internal stay rejected — read-only / cross-user
//     views in the LarePass UX with no save affordance.
//
// Adding a namespace to either side should be an obvious code-
// review signal — that's why this lives next to Plan and not as
// a generic FrontendPath helper.
func TestPlan_NamespaceAllowlist(t *testing.T) {
	supported := []string{"drive", "sync", "cache", "external"}
	for _, ft := range supported {
		t.Run("supported/"+ft, func(t *testing.T) {
			tgt := Target{FileType: ft, Extend: "x", SubPath: "/y.txt"}
			if _, err := Plan(tgt); err != nil {
				t.Errorf("Plan(%s): unexpected error: %v", ft, err)
			}
		})
	}
	rejected := []string{"awss3", "google", "dropbox", "tencent", "share", "internal", "unknown"}
	for _, ft := range rejected {
		t.Run("rejected/"+ft, func(t *testing.T) {
			tgt := Target{FileType: ft, Extend: "x", SubPath: "/y.txt"}
			_, err := Plan(tgt)
			if err == nil {
				t.Fatalf("Plan(%s): want error, got nil", ft)
			}
			if !strings.Contains(err.Error(), "not supported") {
				t.Errorf("Plan(%s): error %q does not say 'not supported'", ft, err.Error())
			}
		})
	}
}

// TestPlan_CloudDriveDedicatedMessage pins the targeted error a
// cloud-drive Plan returns. The GUI / docs both refer to these as
// "supported" because the URL shape is uniform — the actual
// failure is now on the WRITE leg (PUT /api/resources/<cloud-path>
// is only wired in awss3's v2 utils; google / dropbox / tencent
// have no save helper at all, so the PUT shape is unverified
// per cloud driver). The historical fetch-leg risk (preview JSON
// envelopes from /api/raw) was retired with the cloud-bridge
// consolidation — the unified raw endpoint now serves cloud bytes.
// A generic "not supported" message would leave users guessing;
// the dedicated message names the writeback gap and points at
// the proven download → edit-locally → upload alternative.
func TestPlan_CloudDriveDedicatedMessage(t *testing.T) {
	for _, ft := range []string{"awss3", "google", "dropbox", "tencent"} {
		t.Run(ft, func(t *testing.T) {
			_, err := Plan(Target{FileType: ft, Extend: "acct", SubPath: "/x.txt"})
			if err == nil {
				t.Fatalf("Plan(%s): want error, got nil", ft)
			}
			msg := err.Error()
			if !strings.Contains(msg, "cloud-drive") {
				t.Errorf("err should call out 'cloud-drive': %v", msg)
			}
			if !strings.Contains(msg, "/api/resources") {
				t.Errorf("err should name the /api/resources writeback gap: %v", msg)
			}
			if !strings.Contains(msg, "files download") || !strings.Contains(msg, "files upload") {
				t.Errorf("err should suggest the download → upload workaround: %v", msg)
			}
		})
	}
}

// TestClient_PutBytes_Success: the client sends a PUT against the
// computed endpoint with the supplied body + Content-Type, and
// surfaces a 2xx as nil error. Inspect the captured request to
// confirm the wire shape matches the web app's saveFile call.
func TestClient_PutBytes_Success(t *testing.T) {
	var (
		gotMethod      string
		gotPath        string
		gotContentType string
		gotBody        []byte
	)
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))

	op := Op{Endpoint: "/api/resources/drive/Home/notes.md", DisplayPath: "drive/Home/notes.md"}
	want := []byte("hello world\n")
	if err := c.PutBytes(context.Background(), op, want, ""); err != nil {
		t.Fatalf("PutBytes: %v", err)
	}
	if gotMethod != http.MethodPut {
		t.Errorf("method: got %q, want PUT", gotMethod)
	}
	if gotPath != "/api/resources/drive/Home/notes.md" {
		t.Errorf("path: got %q", gotPath)
	}
	if gotContentType != DefaultContentType {
		t.Errorf("Content-Type: got %q, want %q", gotContentType, DefaultContentType)
	}
	if string(gotBody) != string(want) {
		t.Errorf("body: got %q, want %q", gotBody, want)
	}
}

// TestClient_PutBytes_CustomContentType: when a non-empty content
// type is passed, it must be threaded through verbatim (so a user
// can save JSON / YAML / markdown with the right server-side
// hint).
func TestClient_PutBytes_CustomContentType(t *testing.T) {
	var gotCT string
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	op := Op{Endpoint: "/api/resources/drive/Home/data.json"}
	if err := c.PutBytes(context.Background(), op, []byte(`{"k":1}`), "application/json"); err != nil {
		t.Fatalf("PutBytes: %v", err)
	}
	if gotCT != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", gotCT)
	}
}

// TestClient_PutBytes_HTTPError: a non-2xx surfaces as *HTTPError
// with the status / URL / method preserved, so the cobra layer's
// reformatter can branch on Status without stringly-typed parsing.
func TestClient_PutBytes_HTTPError(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	}))
	op := Op{Endpoint: "/api/resources/drive/Home/x.md"}
	err := c.PutBytes(context.Background(), op, []byte("ignored"), "")
	if err == nil {
		t.Fatalf("want error, got nil")
	}
	if !IsHTTPStatus(err, http.StatusForbidden) {
		t.Errorf("IsHTTPStatus(403): got false; err=%v", err)
	}
	var hErr *HTTPError
	if !errors.As(err, &hErr) {
		t.Fatalf("error is not *HTTPError: %T", err)
	}
	if hErr.Method != http.MethodPut {
		t.Errorf("method: got %q", hErr.Method)
	}
	if !strings.Contains(hErr.Body, "forbidden") {
		t.Errorf("body preserved: got %q", hErr.Body)
	}
}

// TestClient_Fetch_Success: GET /api/raw/<encPath> returns the
// file contents verbatim (no envelope unwrapping). Confirm we
// pass through the bytes the server sent.
func TestClient_Fetch_Success(t *testing.T) {
	want := []byte("# Notes\n\nhello world\n")
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method: got %q, want GET", r.Method)
		}
		if r.URL.Path != "/api/raw/drive/Home/Documents/notes.md" {
			t.Errorf("path: got %q", r.URL.Path)
		}
		_, _ = w.Write(want)
	}))
	got, err := c.Fetch(context.Background(), "drive/Home/Documents/notes.md", 0)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("body: got %q, want %q", got, want)
	}
}

// TestClient_Fetch_NotFound: a 404 surfaces as *HTTPError and
// IsNotFound returns true, so the cobra layer can branch on
// `--create` to start with an empty buffer instead of failing.
func TestClient_Fetch_NotFound(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	_, err := c.Fetch(context.Background(), "drive/Home/missing.md", 0)
	if err == nil {
		t.Fatalf("want error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("IsNotFound: got false; err=%v", err)
	}
}

// TestClient_Fetch_MaxBytesGuard pins Bug 5's fix: when maxBytes
// > 0, the bounded read in Fetch wraps the body in a
// LimitReader(_, maxBytes+1) and returns *TooLargeError if the
// server delivered more than maxBytes bytes. The test deliberately
// drives a server that ignores any client-side hint (no
// Content-Length quirks, no Range support) so the assertion
// targets the LimitReader path itself rather than a server-side
// optimisation.
func TestClient_Fetch_MaxBytesGuard(t *testing.T) {
	t.Run("body within cap passes through", func(t *testing.T) {
		want := []byte("hello\n")
		c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write(want)
		}))
		got, err := c.Fetch(context.Background(), "drive/Home/x.md", 1024)
		if err != nil {
			t.Fatalf("Fetch: %v", err)
		}
		if string(got) != string(want) {
			t.Errorf("body: got %q, want %q", got, want)
		}
	})

	t.Run("body exactly at cap passes through", func(t *testing.T) {
		// Boundary case: maxBytes bytes is OK; the LimitReader
		// reads up to maxBytes+1 to detect overflow but a body
		// of exactly maxBytes shouldn't trip the check.
		const cap = 16
		want := bytes.Repeat([]byte{'a'}, cap)
		c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write(want)
		}))
		got, err := c.Fetch(context.Background(), "drive/Home/x.md", cap)
		if err != nil {
			t.Fatalf("Fetch: %v", err)
		}
		if len(got) != cap {
			t.Errorf("len: got %d, want %d", len(got), cap)
		}
	})

	t.Run("body over cap returns *TooLargeError without unbounded read", func(t *testing.T) {
		// Server tries to deliver 64 KiB but the cap is 1 KiB:
		// the LimitReader caps the buffer at 1024+1 bytes, and
		// Fetch returns *TooLargeError. This is the exact path
		// Bug 5 was about — a Stat.Size==0 listing followed by
		// a real 64 KiB body must NOT pull the whole 64 KiB
		// into memory.
		const cap = 1024
		const serverWants = 64 * 1024
		c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write(bytes.Repeat([]byte{'x'}, serverWants))
		}))
		_, err := c.Fetch(context.Background(), "drive/Home/big.bin", cap)
		if err == nil {
			t.Fatal("want *TooLargeError, got nil")
		}
		var tle *TooLargeError
		if !errors.As(err, &tle) {
			t.Fatalf("err type: got %T (%v), want *TooLargeError", err, err)
		}
		if tle.Limit != cap {
			t.Errorf("Limit: got %d, want %d", tle.Limit, cap)
		}
		// Read should be Limit+1 (LimitReader bounded the buffer
		// at maxBytes+1 bytes regardless of what the server tried
		// to send) — defending the "bounded memory" property.
		if tle.Read > cap+1 {
			t.Errorf("Read: got %d (Cap+1=%d); LimitReader-bounded read leaked", tle.Read, cap+1)
		}
	})

	t.Run("maxBytes=0 disables the cap entirely", func(t *testing.T) {
		// 32 KiB body, no cap → must return all bytes, no
		// TooLargeError.
		body := bytes.Repeat([]byte{'y'}, 32*1024)
		c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write(body)
		}))
		got, err := c.Fetch(context.Background(), "drive/Home/big.bin", 0)
		if err != nil {
			t.Fatalf("Fetch: %v", err)
		}
		if len(got) != len(body) {
			t.Errorf("len: got %d, want %d", len(got), len(body))
		}
	})
}

// TestClient_Put_EmptyEndpoint: defensive guard. Plan should never
// emit an empty Endpoint, but if a future caller forgets to call
// Plan we want a typed error rather than a silent malformed URL.
func TestClient_Put_EmptyEndpoint(t *testing.T) {
	c := &Client{HTTPClient: http.DefaultClient, BaseURL: "http://x"}
	err := c.PutBytes(context.Background(), Op{}, []byte("x"), "")
	if err == nil {
		t.Fatal("want error, got nil")
	}
	if !strings.Contains(err.Error(), "empty Endpoint") {
		t.Errorf("error: got %q", err.Error())
	}
}
