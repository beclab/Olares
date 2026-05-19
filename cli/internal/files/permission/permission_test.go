package permission

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestClient mirrors the share / rm / cp harness: stand up a
// real httptest server, hand the caller a Client whose BaseURL
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

// TestBuildPermissionURL covers the wire-path encoding contract:
// segments percent-encoded individually (so '/' is preserved between
// segments but encoded inside a segment), with an unconditional
// trailing '/' for backend routing parity with the LarePass GUI.
func TestBuildPermissionURL(t *testing.T) {
	cases := []struct {
		name   string
		t      Target
		expect string
	}{
		{
			name:   "drive home file",
			t:      Target{FileType: "drive", Extend: "Home", SubPath: "/Documents/foo.pdf"},
			expect: "/api/permission/drive/Home/Documents/foo.pdf/",
		},
		{
			name:   "drive home dir with trailing slash preserved",
			t:      Target{FileType: "drive", Extend: "Home", SubPath: "/Photos/", IsDirIntent: true},
			expect: "/api/permission/drive/Home/Photos/",
		},
		{
			name:   "drive data file",
			t:      Target{FileType: "drive", Extend: "Data", SubPath: "/builds/out.bin"},
			expect: "/api/permission/drive/Data/builds/out.bin/",
		},
		{
			name:   "cache node deep path",
			t:      Target{FileType: "cache", Extend: "node-a", SubPath: "/scratch/sub/"},
			expect: "/api/permission/cache/node-a/scratch/sub/",
		},
		{
			name:   "extend root with empty subpath defaults to slash",
			t:      Target{FileType: "drive", Extend: "Home", SubPath: ""},
			expect: "/api/permission/drive/Home/",
		},
		{
			name:   "subpath without leading slash gets one prepended",
			t:      Target{FileType: "drive", Extend: "Home", SubPath: "Documents/foo.txt"},
			expect: "/api/permission/drive/Home/Documents/foo.txt/",
		},
		{
			name:   "spaces are percent-encoded inside a segment",
			t:      Target{FileType: "drive", Extend: "Home", SubPath: "/My Folder/"},
			expect: "/api/permission/drive/Home/My%20Folder/",
		},
		{
			name:   "unicode and reserved chars",
			t:      Target{FileType: "drive", Extend: "Home", SubPath: "/à & b/"},
			expect: "/api/permission/drive/Home/%C3%A0%20%26%20b/",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := buildPermissionURL(tc.t)
			if got != tc.expect {
				t.Errorf("buildPermissionURL = %q, want %q", got, tc.expect)
			}
		})
	}
}

// TestIsSupported covers the LarePass `permissionInDriveType` mirror.
// drive / cache: yes. Everything else (sync / external / cloud): no.
// Locking the membership down at unit-test level catches a
// regression where someone widens SupportedFileTypes without first
// confirming the LarePass GUI surfaces those namespaces.
func TestIsSupported(t *testing.T) {
	allow := []string{"drive", "cache"}
	deny := []string{"sync", "external", "awss3", "dropbox", "google", "tencent", "share", "internal", "drives" /* typo */}
	for _, ft := range allow {
		if !IsSupported(ft) {
			t.Errorf("IsSupported(%q) = false, want true", ft)
		}
	}
	for _, ft := range deny {
		if IsSupported(ft) {
			t.Errorf("IsSupported(%q) = true, want false", ft)
		}
	}
}

// TestSupportedFileTypesList renders the allow-list deterministically
// so error messages don't churn from map iteration order.
func TestSupportedFileTypesList(t *testing.T) {
	got := SupportedFileTypesList()
	want := "cache, drive"
	if got != want {
		t.Errorf("SupportedFileTypesList() = %q, want %q", got, want)
	}
}

// TestClient_Get_HappyPath confirms the GET wire shape:
//   - method GET
//   - URL /api/permission/<fileType>/<extend><subPath>/
//   - response body {uid:<int>} decodes into the int return
func TestClient_Get_HappyPath(t *testing.T) {
	var gotMethod, gotPath string
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"uid":1000}`)
	}))

	uid, err := c.Get(context.Background(), Target{
		FileType: "drive", Extend: "Home", SubPath: "/foo.txt",
	})
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if uid != 1000 {
		t.Errorf("uid = %d, want 1000", uid)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/api/permission/drive/Home/foo.txt/" {
		t.Errorf("path = %q, want /api/permission/drive/Home/foo.txt/", gotPath)
	}
}

// TestClient_Set_HappyPath confirms the PUT wire shape:
//   - method PUT
//   - URL /api/permission/<fileType>/<extend><subPath>/?uid=<int>[&recursive=1]
//   - body == `{}` (the LarePass GUI sends an explicit empty object;
//     keep it byte-for-byte)
//   - recursive=true emits literal "1" (not "true") to match GUI
func TestClient_Set_HappyPath(t *testing.T) {
	cases := []struct {
		name      string
		uid       int
		recursive bool
		wantQ     string
	}{
		{name: "non-recursive", uid: 0, recursive: false, wantQ: "uid=0"},
		{name: "recursive user", uid: 1000, recursive: true, wantQ: "recursive=1&uid=1000"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var gotMethod, gotPath, gotRawQ string
			var gotBody []byte
			c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotPath = r.URL.Path
				gotRawQ = r.URL.RawQuery
				gotBody, _ = io.ReadAll(r.Body)
				w.WriteHeader(http.StatusOK)
			}))
			err := c.Set(context.Background(), Target{
				FileType: "cache", Extend: "node-a", SubPath: "/x/",
			}, tc.uid, tc.recursive)
			if err != nil {
				t.Fatalf("Set: %v", err)
			}
			if gotMethod != http.MethodPut {
				t.Errorf("method = %q, want PUT", gotMethod)
			}
			if gotPath != "/api/permission/cache/node-a/x/" {
				t.Errorf("path = %q", gotPath)
			}
			if gotRawQ != tc.wantQ {
				t.Errorf("query = %q, want %q", gotRawQ, tc.wantQ)
			}
			if string(gotBody) != "{}" {
				t.Errorf("body = %q, want %q", string(gotBody), "{}")
			}
		})
	}
}

// TestClient_HTTPError surfaces a non-2xx response as *HTTPError so
// the cobra layer's errors.As branches can reformat by status. We
// hit GET (the simpler path); the same `do` helper handles PUT.
func TestClient_HTTPError(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error":"unauthorized"}`)
	}))
	_, err := c.Get(context.Background(), Target{
		FileType: "drive", Extend: "Home", SubPath: "/x.txt",
	})
	var hErr *HTTPError
	if !errors.As(err, &hErr) {
		t.Fatalf("err type = %T, want *HTTPError", err)
	}
	if hErr.Status != http.StatusUnauthorized {
		t.Errorf("Status = %d, want 401", hErr.Status)
	}
	if !strings.Contains(hErr.Body, "unauthorized") {
		t.Errorf("Body = %q, expected to contain 'unauthorized'", hErr.Body)
	}
	if hErr.Method != http.MethodGet {
		t.Errorf("Method = %q, want GET", hErr.Method)
	}
}

// TestClient_Get_RejectsEmptyTarget locks in the defense-in-depth
// guard against callers that bypass the FrontendPath parser. The
// error must surface client-side without a wire call.
func TestClient_Get_RejectsEmptyTarget(t *testing.T) {
	c, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected wire call: %s %s", r.Method, r.URL)
	}))
	_, err := c.Get(context.Background(), Target{FileType: "", Extend: "Home"})
	if err == nil {
		t.Fatal("Get with empty fileType: expected error, got nil")
	}
	_, err = c.Get(context.Background(), Target{FileType: "drive", Extend: ""})
	if err == nil {
		t.Fatal("Get with empty extend: expected error, got nil")
	}
}

// TestTargetString covers the human-readable display path used in
// error messages and progress lines. Keeps the same shape as
// FrontendPath.String().
func TestTargetString(t *testing.T) {
	cases := []struct {
		t      Target
		expect string
	}{
		{Target{FileType: "drive", Extend: "Home", SubPath: "/foo.txt"}, "drive/Home/foo.txt"},
		{Target{FileType: "drive", Extend: "Data", SubPath: "/dir/"}, "drive/Data/dir/"},
		{Target{FileType: "cache", Extend: "n", SubPath: ""}, "cache/n/"},
	}
	for _, tc := range cases {
		if got := tc.t.String(); got != tc.expect {
			t.Errorf("Target%+v.String() = %q, want %q", tc.t, got, tc.expect)
		}
	}
}
