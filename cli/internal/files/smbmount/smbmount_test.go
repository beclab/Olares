package smbmount

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestClient mirrors the share / cp / permission harness: stand
// up a real httptest server, hand the caller a Client whose BaseURL
// points at it, and let the test inspect what landed on the wire.
func newTestClient(t *testing.T, h http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &Client{
		HTTPClient: srv.Client(),
		BaseURL:    srv.URL,
	}, srv
}

// TestBuildMountURL pins the query-string + node-segment shapes
// LarePass uses, including the "drop the node segment when empty"
// branch (stores/files.ts L1266-L1272).
func TestBuildMountURL(t *testing.T) {
	cases := []struct {
		name   string
		node   string
		expect string
	}{
		{"with node", "node-a", "/api/mount/node-a/?external_type=smb"},
		{"empty node drops segment", "", "/api/mount/?external_type=smb"},
		{"node with space", "node a", "/api/mount/node%20a/?external_type=smb"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := buildMountURL(c.node)
			if got != c.expect {
				t.Errorf("buildMountURL(%q) = %q, want %q", c.node, got, c.expect)
			}
		})
	}
}

// TestBuildUnmountURL pins the per-segment percent-encoding plus the
// trailing slash + ?external_type=<...> tail. Mirrors the shape
// stores/operation.ts L861-L870 builds via string concatenation.
func TestBuildUnmountURL(t *testing.T) {
	cases := []struct {
		name         string
		fileType     string
		fileExtend   string
		entry        string
		externalType string
		expect       string
	}{
		{
			name:         "smb canonical",
			fileType:     "external",
			fileExtend:   "main",
			entry:        "smb-host-share",
			externalType: "smb",
			expect:       "/api/unmount/external/main/smb-host-share/?external_type=smb",
		},
		{
			name:         "entry with space",
			fileType:     "external",
			fileExtend:   "main",
			entry:        "smb host share",
			externalType: "smb",
			expect:       "/api/unmount/external/main/smb%20host%20share/?external_type=smb",
		},
		{
			name:         "usb pass-through (future-proofing)",
			fileType:     "external",
			fileExtend:   "main",
			entry:        "usb1",
			externalType: "usb",
			expect:       "/api/unmount/external/main/usb1/?external_type=usb",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := buildUnmountURL(c.fileType, c.fileExtend, c.entry, c.externalType)
			if got != c.expect {
				t.Errorf("buildUnmountURL = %q, want %q", got, c.expect)
			}
		})
	}
}

// TestBuildHistoryURL keeps the GET / PUT / DELETE shape consistent
// with apps/.../components/files/smb/ConnectServerStep1.vue
// (CommonFetch.put('/api/smb_history/' + extend + '/', ...)).
func TestBuildHistoryURL(t *testing.T) {
	cases := []struct {
		node, expect string
	}{
		{"node-a", "/api/smb_history/node-a/"},
		{"node a", "/api/smb_history/node%20a/"},
	}
	for _, c := range cases {
		got := buildHistoryURL(c.node)
		if got != c.expect {
			t.Errorf("buildHistoryURL(%q) = %q, want %q", c.node, got, c.expect)
		}
	}
}

// TestFetchNodes mirrors cp.TestFetchNodes_Envelope: the wire
// envelope is `{data:{nodes:[...]}}` and we surface
// `nodes[0].Name` to the cobra layer as the default node.
func TestFetchNodes(t *testing.T) {
	t.Run("envelope happy path", func(t *testing.T) {
		c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/nodes/" {
				t.Errorf("path = %q", r.URL.Path)
			}
			_, _ = io.WriteString(w, `{"data":{"nodes":[{"name":"node-a","master":true},{"name":"node-b"}]}}`)
		}))
		nodes, err := c.FetchNodes(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if len(nodes) != 2 || nodes[0].Name != "node-a" || !nodes[0].Master {
			t.Errorf("got %+v", nodes)
		}
	})
	t.Run("empty list errors out", func(t *testing.T) {
		c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, `{"data":{"nodes":[]}}`)
		}))
		_, err := c.FetchNodes(context.Background())
		if err == nil {
			t.Fatal("expected error on empty nodes list")
		}
	})
	t.Run("non-2xx surfaces as HTTPError", func(t *testing.T) {
		c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		_, err := c.FetchNodes(context.Background())
		var hErr *HTTPError
		if !errors.As(err, &hErr) {
			t.Fatalf("err type = %T, want *HTTPError", err)
		}
		if hErr.Status != http.StatusUnauthorized {
			t.Errorf("Status = %d", hErr.Status)
		}
	})
}

// TestMount_Code200 covers the happy-path mount: server returns
// `{code:200, ...}`, the wire client surfaces a MountResult with
// Code==200 and an empty Paths slice.
func TestMount_Code200(t *testing.T) {
	var gotMethod, gotPath, gotRawQ string
	var gotBody []byte
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotRawQ = r.Method, r.URL.Path, r.URL.RawQuery
		gotBody, _ = io.ReadAll(r.Body)
		_, _ = io.WriteString(w, `{"code":200,"message":"ok"}`)
	}))
	res, err := c.Mount(context.Background(), "node-a", MountOptions{
		SMBPath: "//host/share", User: "alice", Password: "s3cret",
	})
	if err != nil {
		t.Fatalf("Mount: %v", err)
	}
	if res.Code != 200 {
		t.Errorf("Code = %d, want 200", res.Code)
	}
	if len(res.Paths) != 0 {
		t.Errorf("Paths = %v, want empty on code 200", res.Paths)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q", gotMethod)
	}
	if gotPath != "/api/mount/node-a/" {
		t.Errorf("path = %q", gotPath)
	}
	if gotRawQ != "external_type=smb" {
		t.Errorf("query = %q", gotRawQ)
	}
	// Body shape — keep it pinned to the LarePass app's exact
	// {smbPath, user, password} keys.
	var sent map[string]string
	if err := json.Unmarshal(gotBody, &sent); err != nil {
		t.Fatalf("decode body %q: %v", string(gotBody), err)
	}
	if sent["smbPath"] != "//host/share" || sent["user"] != "alice" || sent["password"] != "s3cret" {
		t.Errorf("body = %+v", sent)
	}
}

// TestMount_Code300 confirms the multi-share branch: code 300 →
// MountResult.Paths is populated with the server-supplied paths.
func TestMount_Code300(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"code":300,"message":"choose","data":[{"path":"//h/A"},{"path":"//h/B"},{"path":"//h/C"}]}`)
	}))
	res, err := c.Mount(context.Background(), "node-a", MountOptions{SMBPath: "//host"})
	if err != nil {
		t.Fatalf("Mount: %v", err)
	}
	if res.Code != 300 {
		t.Errorf("Code = %d, want 300", res.Code)
	}
	want := []string{"//h/A", "//h/B", "//h/C"}
	if len(res.Paths) != len(want) {
		t.Fatalf("Paths len = %d, want %d (%+v)", len(res.Paths), len(want), res.Paths)
	}
	for i, p := range want {
		if res.Paths[i] != p {
			t.Errorf("Paths[%d] = %q, want %q", i, res.Paths[i], p)
		}
	}
}

// TestMount_OtherCode_Errors confirms a non-200/300 envelope code
// is surfaced as a regular error (not *HTTPError) with the wire
// message preserved.
func TestMount_OtherCode_Errors(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"code":401,"message":"bad creds"}`)
	}))
	_, err := c.Mount(context.Background(), "node-a", MountOptions{SMBPath: "//host/share"})
	if err == nil {
		t.Fatal("expected error on non-200/300 code")
	}
	if !strings.Contains(err.Error(), "code 401") || !strings.Contains(err.Error(), "bad creds") {
		t.Errorf("error = %v, expected to mention code + message", err)
	}
	// MUST NOT be *HTTPError — server returned 200 OK + envelope error.
	var hErr *HTTPError
	if errors.As(err, &hErr) {
		t.Errorf("error should NOT be *HTTPError on a 200/envelope-failure (got %v)", hErr)
	}
}

// TestMount_RejectsEmptySMBPath locks the defense-in-depth guard
// for callers that bypass the cobra-layer validation.
func TestMount_RejectsEmptySMBPath(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected wire call: %s %s", r.Method, r.URL)
	}))
	_, err := c.Mount(context.Background(), "node-a", MountOptions{SMBPath: ""})
	if err == nil {
		t.Fatal("expected error for empty smbPath")
	}
}

// TestUnmount_HappyPath confirms POST + URL + empty-body shape and
// that an empty 2xx response is treated as success.
func TestUnmount_HappyPath(t *testing.T) {
	var gotMethod, gotPath, gotRawQ string
	var gotBody []byte
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath, gotRawQ = r.Method, r.URL.Path, r.URL.RawQuery
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	if err := c.Unmount(context.Background(), "external", "main", "smb-host-share", "smb"); err != nil {
		t.Fatalf("Unmount: %v", err)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/api/unmount/external/main/smb-host-share/" {
		t.Errorf("path = %q", gotPath)
	}
	if gotRawQ != "external_type=smb" {
		t.Errorf("query = %q", gotRawQ)
	}
	if string(gotBody) != "{}" {
		t.Errorf("body = %q, want {}", string(gotBody))
	}
}

// TestUnmount_PlaintextBody confirms that a 2xx with a plaintext
// (non-JSON) body — the shape the live files-backend actually uses,
// e.g. "Successfully unmounted ..." — is treated as success. Mirrors
// LarePass's `removeFavorite` / `unmount` GUI handlers, which never
// inspect the body at all.
func TestUnmount_PlaintextBody(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "Successfully unmounted SMB share")
	}))
	if err := c.Unmount(context.Background(), "external", "main", "smb-host-share", "smb"); err != nil {
		t.Errorf("expected nil for plaintext 2xx body, got %v", err)
	}
}

// TestUnmount_EnvelopeError exercises the "server returned 2xx with
// {code: N != 0/200}" branch — the only remaining decode path after
// we relaxed plaintext bodies.
func TestUnmount_EnvelopeError(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"code":500,"message":"mount kernel busy"}`)
	}))
	err := c.Unmount(context.Background(), "external", "main", "smb-host-share", "smb")
	if err == nil || !strings.Contains(err.Error(), "code 500") || !strings.Contains(err.Error(), "mount kernel busy") {
		t.Errorf("err = %v, want server-side rejection with code+message", err)
	}
}

// TestUnmount_RejectsEmpty exercises every required-arg guard.
func TestUnmount_RejectsEmpty(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected wire call: %s %s", r.Method, r.URL)
	}))
	cases := []struct {
		name                                   string
		fileType, fileExtend, entry, extType   string
	}{
		{"empty fileType", "", "main", "x", "smb"},
		{"empty fileExtend", "external", "", "x", "smb"},
		{"empty name", "external", "main", "", "smb"},
		{"empty externalType", "external", "main", "x", ""},
	}
	for _, c2 := range cases {
		t.Run(c2.name, func(t *testing.T) {
			err := c.Unmount(context.Background(), c2.fileType, c2.fileExtend, c2.entry, c2.extType)
			if err == nil {
				t.Errorf("expected error for %+v", c2)
			}
		})
	}
}

// TestHistoryList covers both wire shapes — the GUI-observed bare
// array AND the defensive {code,message,data:[]} envelope decoder
// — plus the empty-body branch.
func TestHistoryList(t *testing.T) {
	t.Run("bare array body", func(t *testing.T) {
		c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, `[{"url":"//a/x","username":"u","password":"p","timestamp":1},{"url":"//b/y"}]`)
		}))
		entries, err := c.HistoryList(context.Background(), "node-a")
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 2 {
			t.Fatalf("len = %d", len(entries))
		}
		if entries[0].URL != "//a/x" || entries[0].Username != "u" || entries[0].Password != "p" || entries[0].Timestamp != 1 {
			t.Errorf("entries[0] = %+v", entries[0])
		}
		if entries[1].URL != "//b/y" || entries[1].Password != "" {
			t.Errorf("entries[1] = %+v", entries[1])
		}
	})
	t.Run("envelope body", func(t *testing.T) {
		c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, `{"code":0,"data":[{"url":"//a/x"}]}`)
		}))
		entries, err := c.HistoryList(context.Background(), "node-a")
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 1 || entries[0].URL != "//a/x" {
			t.Errorf("entries = %+v", entries)
		}
	})
	t.Run("empty body", func(t *testing.T) {
		c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		entries, err := c.HistoryList(context.Background(), "node-a")
		if err != nil {
			t.Fatal(err)
		}
		if entries == nil {
			t.Errorf("entries = nil; want empty slice (caller-friendly invariant)")
		}
		if len(entries) != 0 {
			t.Errorf("len = %d", len(entries))
		}
	})
	t.Run("envelope error code", func(t *testing.T) {
		c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, `{"code":500,"message":"boom"}`)
		}))
		_, err := c.HistoryList(context.Background(), "node-a")
		if err == nil || !strings.Contains(err.Error(), "boom") {
			t.Errorf("err = %v, want server message", err)
		}
	})
	t.Run("empty node rejected", func(t *testing.T) {
		c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatalf("unexpected wire call: %s %s", r.Method, r.URL)
		}))
		if _, err := c.HistoryList(context.Background(), ""); err == nil {
			t.Fatal("expected error for empty node")
		}
	})
}

// TestHistoryUpsert pins the PUT shape (method, URL, JSON body).
func TestHistoryUpsert(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		var gotMethod, gotPath string
		var gotBody []byte
		c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod, gotPath = r.Method, r.URL.Path
			gotBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
		}))
		err := c.HistoryUpsert(context.Background(), "node-a", []HistoryEntry{
			{URL: "//a/x", Username: "u", Password: "p"},
			{URL: "//b/y"},
		})
		if err != nil {
			t.Fatal(err)
		}
		if gotMethod != http.MethodPut {
			t.Errorf("method = %q, want PUT", gotMethod)
		}
		if gotPath != "/api/smb_history/node-a/" {
			t.Errorf("path = %q", gotPath)
		}
		// Round-trip the body so a single-vs-double-quote / key-order
		// regression doesn't make this brittle.
		var sent []HistoryEntry
		if err := json.Unmarshal(gotBody, &sent); err != nil {
			t.Fatalf("decode body %q: %v", string(gotBody), err)
		}
		if len(sent) != 2 || sent[0].URL != "//a/x" || sent[1].Password != "" {
			t.Errorf("sent = %+v", sent)
		}
	})
	t.Run("empty entries rejected", func(t *testing.T) {
		c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatalf("unexpected wire call: %s %s", r.Method, r.URL)
		}))
		if err := c.HistoryUpsert(context.Background(), "node-a", nil); err == nil {
			t.Fatal("expected error for empty entries")
		}
	})
	t.Run("entry with empty url rejected", func(t *testing.T) {
		c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatalf("unexpected wire call: %s %s", r.Method, r.URL)
		}))
		err := c.HistoryUpsert(context.Background(), "node-a", []HistoryEntry{{URL: ""}})
		if err == nil {
			t.Fatal("expected error for empty url")
		}
	})
	t.Run("plaintext success body tolerated", func(t *testing.T) {
		c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, "Successfully saved SMB history")
		}))
		err := c.HistoryUpsert(context.Background(), "node-a", []HistoryEntry{{URL: "//a/x"}})
		if err != nil {
			t.Errorf("expected nil for plaintext 2xx body, got %v", err)
		}
	})
}

// TestHistoryRemove pins the DELETE shape: method DELETE, URL same
// as PUT, body is `[{url}]` (an array of objects, NOT a query
// string).
func TestHistoryRemove(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		var gotMethod, gotPath string
		var gotBody []byte
		c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod, gotPath = r.Method, r.URL.Path
			gotBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
		}))
		if err := c.HistoryRemove(context.Background(), "node-a", []string{"//a/x", "//b/y"}); err != nil {
			t.Fatal(err)
		}
		if gotMethod != http.MethodDelete {
			t.Errorf("method = %q, want DELETE", gotMethod)
		}
		if gotPath != "/api/smb_history/node-a/" {
			t.Errorf("path = %q", gotPath)
		}
		var sent []map[string]string
		if err := json.Unmarshal(gotBody, &sent); err != nil {
			t.Fatalf("decode body %q: %v", string(gotBody), err)
		}
		if len(sent) != 2 || sent[0]["url"] != "//a/x" || sent[1]["url"] != "//b/y" {
			t.Errorf("body = %+v", sent)
		}
	})
	t.Run("empty urls rejected", func(t *testing.T) {
		c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatalf("unexpected wire call: %s %s", r.Method, r.URL)
		}))
		if err := c.HistoryRemove(context.Background(), "node-a", nil); err == nil {
			t.Fatal("expected error for empty urls")
		}
	})
	// Regression: live backend replies with the plaintext "Successfully
	// deleted SMB history" — the original strict JSON decoder would
	// fail with `invalid character 'S' looking for beginning of value`.
	// Lock that path down so a future tightening doesn't reintroduce
	// the same break.
	t.Run("plaintext success body tolerated", func(t *testing.T) {
		c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, "Successfully deleted SMB history")
		}))
		err := c.HistoryRemove(context.Background(), "node-a", []string{"//a/x"})
		if err != nil {
			t.Errorf("expected nil for plaintext 2xx body, got %v", err)
		}
	})
}

// TestHTTPError_Truncation makes sure the error message stays
// readable when the server replies with a giant body.
func TestHTTPError_Truncation(t *testing.T) {
	huge := strings.Repeat("X", 1024)
	hErr := &HTTPError{Status: 500, Body: huge, URL: "/x", Method: "POST"}
	msg := hErr.Error()
	if !strings.Contains(msg, "...(truncated)") {
		t.Errorf("expected truncation marker, got %q", msg)
	}
}
