package repos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// reposServerOpts plumbs per-test customization into reposServer
// without proliferating constructor variants.
type reposServerOpts struct {
	// statusCode for the GET /api/repos/ response (default 200).
	statusCode int
	// body returned verbatim for /api/repos/ requests; if non-empty,
	// supersedes the per-type slices below. Useful for code != 0
	// envelope tests and malformed-body coverage.
	body string
	// per-type repo slices; selected by the `type` query param.
	mine        []Repo
	shareToMe   []Repo
	sharedOut   []Repo
	// recordQuery captures the query strings the server saw so tests
	// can assert that the right `type=...` was sent.
	recordQuery *[]string
}

// reposServer wires up an httptest.Server emulating /api/repos/ for
// the three filter modes. Returns the server (closed via t.Cleanup)
// and a ready-to-use Client.
func reposServer(t *testing.T, o reposServerOpts) (*Client, *httptest.Server) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/repos/", func(w http.ResponseWriter, r *http.Request) {
		if o.recordQuery != nil {
			*o.recordQuery = append(*o.recordQuery, r.URL.RawQuery)
		}
		if o.statusCode != 0 && o.statusCode != http.StatusOK {
			http.Error(w, o.body, o.statusCode)
			return
		}
		if o.body != "" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, o.body)
			return
		}
		var rows []Repo
		switch r.URL.Query().Get("type") {
		case "":
			rows = o.mine
		case "share_to_me":
			rows = o.shareToMe
		case "shared":
			rows = o.sharedOut
		default:
			rows = nil
		}
		w.Header().Set("Content-Type", "application/json")
		// Match the wire shape: {repos: [...]}; nil rows must
		// serialize as `null` to exercise the "missing/null repos"
		// path, but in the common case we want an empty slice.
		if rows == nil {
			rows = []Repo{}
		}
		_ = json.NewEncoder(w).Encode(struct {
			Repos []Repo `json:"repos"`
		}{rows})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &Client{HTTPClient: srv.Client(), BaseURL: srv.URL}, srv
}

func TestParseType(t *testing.T) {
	cases := []struct {
		in   string
		want Type
		err  bool
	}{
		{"", TypeMine, false},
		{"mine", TypeMine, false},
		{"share-to-me", TypeSharedToMe, false},
		{"share_to_me", TypeSharedToMe, false},
		{"shared-to-me", TypeSharedToMe, false},
		{"share-with-me", TypeSharedToMe, false},
		{"shared", TypeShared, false},
		{"shared-by-me", TypeShared, false},
		{"unknown", "", true},
		{"all", "", true}, // "all" is a CLI-level alias, not a wire value
	}
	for _, tc := range cases {
		got, err := ParseType(tc.in)
		if tc.err {
			if err == nil {
				t.Errorf("ParseType(%q): expected error, got %q", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseType(%q): unexpected error %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseType(%q): got %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestList_Mine(t *testing.T) {
	mine := []Repo{
		{RepoID: "r1", RepoName: "Library 1", Permission: "rw", OwnerEmail: "alice@olares.com"},
		{RepoID: "r2", RepoName: "Library 2", Permission: "rw", OwnerEmail: "alice@olares.com"},
	}
	var seen []string
	c, _ := reposServer(t, reposServerOpts{mine: mine, recordQuery: &seen})

	got, err := c.List(context.Background(), TypeMine)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != len(mine) {
		t.Fatalf("got %d rows, want %d", len(got), len(mine))
	}
	for i := range got {
		if got[i].RepoID != mine[i].RepoID {
			t.Errorf("row %d: got id %q, want %q", i, got[i].RepoID, mine[i].RepoID)
		}
	}
	// TypeMine must NOT include `type=` in the query string —
	// matches the web app's fetchMineRepo behavior.
	if len(seen) != 1 {
		t.Fatalf("expected 1 server call, got %d", len(seen))
	}
	if seen[0] != "" {
		t.Errorf("TypeMine should send no query string, got %q", seen[0])
	}
}

func TestList_SharedToMe(t *testing.T) {
	rows := []Repo{
		{RepoID: "r3", RepoName: "Bob's library",
			SharePermission: "rw", ShareType: "personal",
			UserEmail: "bob@olares.com", UserName: "Bob"},
	}
	var seen []string
	c, _ := reposServer(t, reposServerOpts{shareToMe: rows, recordQuery: &seen})

	got, err := c.List(context.Background(), TypeSharedToMe)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 1 || got[0].SharePermission != "rw" {
		t.Fatalf("unexpected rows: %+v", got)
	}
	q, _ := url.ParseQuery(seen[0])
	if q.Get("type") != "share_to_me" {
		t.Errorf("expected ?type=share_to_me, got %q", seen[0])
	}
}

func TestList_Shared(t *testing.T) {
	rows := []Repo{{RepoID: "r4", RepoName: "Shared Out", SharePermission: "r"}}
	var seen []string
	c, _ := reposServer(t, reposServerOpts{sharedOut: rows, recordQuery: &seen})

	got, err := c.List(context.Background(), TypeShared)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 1 || got[0].RepoID != "r4" {
		t.Fatalf("unexpected rows: %+v", got)
	}
	q, _ := url.ParseQuery(seen[0])
	if q.Get("type") != "shared" {
		t.Errorf("expected ?type=shared, got %q", seen[0])
	}
}

func TestList_EmptySlice(t *testing.T) {
	c, _ := reposServer(t, reposServerOpts{}) // every type returns nil → encoded as []
	got, err := c.List(context.Background(), TypeMine)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if got == nil {
		t.Fatal("got nil; want non-nil empty slice")
	}
	if len(got) != 0 {
		t.Errorf("got %d rows, want 0", len(got))
	}
}

func TestList_NullRepos(t *testing.T) {
	// Server returns `{"repos": null}`. Some response shapes might
	// drop the field altogether; both should yield an empty
	// (non-nil) slice from List.
	c, _ := reposServer(t, reposServerOpts{body: `{"repos":null}`})
	got, err := c.List(context.Background(), TypeMine)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("expected empty slice, got %#v", got)
	}
}

func TestList_CodeNonZero(t *testing.T) {
	c, _ := reposServer(t, reposServerOpts{
		body: `{"code":-1,"message":"server says no","repos":null}`,
	})
	_, err := c.List(context.Background(), TypeMine)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "server says no") || !strings.Contains(err.Error(), "code -1") {
		t.Errorf("error %q lacks expected substrings", err)
	}
}

func TestList_HTTPError(t *testing.T) {
	c, _ := reposServer(t, reposServerOpts{
		statusCode: http.StatusUnauthorized,
		body:       `unauthorized`,
	})
	_, err := c.List(context.Background(), TypeMine)
	if err == nil {
		t.Fatal("expected error")
	}
	var hErr *HTTPError
	if !errors.As(err, &hErr) {
		t.Fatalf("expected *HTTPError, got %T (%v)", err, err)
	}
	if hErr.Status != http.StatusUnauthorized {
		t.Errorf("status: got %d, want 401", hErr.Status)
	}
}

func TestListAll_Concatenates(t *testing.T) {
	c, _ := reposServer(t, reposServerOpts{
		mine:      []Repo{{RepoID: "m1"}},
		shareToMe: []Repo{{RepoID: "s1"}, {RepoID: "s2"}},
		sharedOut: []Repo{{RepoID: "o1"}},
	})
	got, err := c.ListAll(context.Background())
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	want := []string{"m1", "s1", "s2", "o1"}
	if len(got) != len(want) {
		t.Fatalf("got %d rows, want %d", len(got), len(want))
	}
	for i, id := range want {
		if got[i].RepoID != id {
			t.Errorf("row %d: got %q, want %q", i, got[i].RepoID, id)
		}
	}
}

func TestGet_HitInMine(t *testing.T) {
	c, _ := reposServer(t, reposServerOpts{
		mine:      []Repo{{RepoID: "m1", RepoName: "Library M"}},
		shareToMe: []Repo{{RepoID: "s1"}},
		sharedOut: []Repo{{RepoID: "o1"}},
	})
	got, err := c.Get(context.Background(), "m1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got == nil {
		t.Fatal("expected hit, got nil")
	}
	if got.RepoName != "Library M" {
		t.Errorf("name: got %q, want %q", got.RepoName, "Library M")
	}
}

func TestGet_HitInShared(t *testing.T) {
	c, _ := reposServer(t, reposServerOpts{
		shareToMe: []Repo{{RepoID: "s1", RepoName: "Bob's"}},
	})
	got, err := c.Get(context.Background(), "s1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got == nil || got.RepoName != "Bob's" {
		t.Fatalf("unexpected result: %+v", got)
	}
}

func TestGet_Miss(t *testing.T) {
	c, _ := reposServer(t, reposServerOpts{
		mine: []Repo{{RepoID: "m1"}},
	})
	got, err := c.Get(context.Background(), "missing-id")
	if err != nil {
		t.Fatalf("Get: unexpected error %v", err)
	}
	if got != nil {
		t.Errorf("expected nil result, got %+v", got)
	}
}

func TestGet_EmptyID(t *testing.T) {
	c := &Client{}
	if _, err := c.Get(context.Background(), ""); err == nil {
		t.Error("expected error for empty repoID")
	}
}

func TestRepo_DecodesSyncFields(t *testing.T) {
	// Verify the shared-only fields decode without dropping any
	// keys; this catches off-by-one tag typos and missing fields.
	const wire = `{
        "repo_id": "abc",
        "repo_name": "demo",
        "encrypted": false,
        "last_modified": "2026-04-27T00:00:00Z",
        "permission": "rw",
        "size": 12345,
        "type": "mine",
        "status": "normal",
        "owner_email": "alice@olares.com",
        "owner_name": "Alice",
        "owner_contact_email": "alice@olares.com",
        "modifier_email": "alice@olares.com",
        "modifier_name": "Alice",
        "modifier_contact_email": "alice@olares.com",
        "monitored": true,
        "starred": false,
        "salt": "",
        "share_permission": "rw",
        "share_type": "personal",
        "user_email": "bob@olares.com",
        "user_name": "Bob",
        "contact_email": "bob@olares.com",
        "is_admin": true
    }`
	var got Repo
	if err := json.Unmarshal([]byte(wire), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.RepoID != "abc" || got.RepoName != "demo" {
		t.Fatalf("common fields: %+v", got)
	}
	if got.Permission != "rw" || int64(got.Size) != 12345 {
		t.Errorf("permission/size: %+v", got)
	}
	if bool(got.Encrypted) {
		t.Errorf("expected encrypted false, got %+v", got)
	}
	if !got.Monitored || got.Starred {
		t.Errorf("flags: %+v", got)
	}
	if got.SharePermission != "rw" || got.UserEmail != "bob@olares.com" || !got.IsAdmin {
		t.Errorf("shared fields: %+v", got)
	}
}

// Some Seahub / gateway nodes encode encrypted and size as JSON
// strings instead of bool / number — GET /api/repos/ must not fail
// (regression: "cannot unmarshal string into bool").
func TestRepo_DecodesStringEncodedBoolAndSize(t *testing.T) {
	const wire = `{
		"encrypted": "false",
		"is_virtual": false,
		"last_modified": "2026-04-24T08:03:45.000Z",
		"repo_id": "b7ffab7f-3ceb-4e36-aeb7-74d958ad0a7a",
		"repo_name": "My Library",
		"size": "0",
		"status": "normal",
		"type": "mine"
	}`
	var got Repo
	if err := json.Unmarshal([]byte(wire), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if bool(got.Encrypted) {
		t.Errorf("encrypted: want false, got %+v", got.Encrypted)
	}
	if int64(got.Size) != 0 {
		t.Errorf("size: want 0, got %d", int64(got.Size))
	}
	if got.IsVirtual {
		t.Errorf("is_virtual: want false, got true")
	}
}

// recordedReq captures method + raw query + parsed query of an
// inbound /api/repos/ call so the mutation tests can assert wire
// shape without having to read the request twice.
type recordedReq struct {
	method   string
	rawQuery string
	query    url.Values
}

// mutateServerOpts is the per-test knob set for mutateServer. We
// keep it small and explicit (one field per scenario) so each test
// can read the fixture without cross-referencing flags.
type mutateServerOpts struct {
	// status to return; default 200.
	status int
	// body returned verbatim. When empty for non-error paths the
	// server emits a synthetic envelope mirroring real Seahub:
	//   - Create: {"code":0, "repo_id":"<from-name>", "repo_name":"<name>"}
	//   - Rename / Delete: {"code":0, "message":"OK"}
	// This default lets the success-path tests stay focused on the
	// request side without re-spelling the response in every case.
	body string
}

// mutateServer wires up an httptest.Server that records the inbound
// call and returns either the canned body or a synthetic one. Used
// by the Create / Rename / Delete tests to assert wire shape +
// response handling in isolation. Distinct from reposServer because
// the GET fixture is GET-only by design.
func mutateServer(t *testing.T, o mutateServerOpts) (*Client, *recordedReq) {
	t.Helper()
	rec := &recordedReq{}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/repos/", func(w http.ResponseWriter, r *http.Request) {
		rec.method = r.Method
		rec.rawQuery = r.URL.RawQuery
		rec.query = r.URL.Query()
		if o.status != 0 && o.status != http.StatusOK {
			http.Error(w, o.body, o.status)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if o.body != "" {
			fmt.Fprint(w, o.body)
			return
		}
		switch r.Method {
		case http.MethodPost:
			fmt.Fprintf(w, `{"code":0,"repo_id":"new-%s","repo_name":%q}`,
				r.URL.Query().Get("repoName"), r.URL.Query().Get("repoName"))
		default:
			fmt.Fprint(w, `{"code":0,"message":"OK"}`)
		}
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &Client{HTTPClient: srv.Client(), BaseURL: srv.URL}, rec
}

func TestCreate_OK(t *testing.T) {
	c, rec := mutateServer(t, mutateServerOpts{})
	repo, err := c.Create(context.Background(), "Project Alpha")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if repo == nil {
		t.Fatal("expected repo, got nil")
	}
	if rec.method != http.MethodPost {
		t.Errorf("method: got %q, want POST", rec.method)
	}
	// repoName is the only param expected on the wire and must be
	// URL-encoded (the synthetic server's mux will reject anything
	// else with a 404). The Go side uses url.Values{}.Encode() which
	// emits %20 for spaces; that's fine — Axios + Seahub accept the
	// same encoding.
	if got := rec.query.Get("repoName"); got != "Project Alpha" {
		t.Errorf("repoName: got %q, want %q", got, "Project Alpha")
	}
	if !strings.Contains(rec.rawQuery, "repoName=") {
		t.Errorf("raw query missing repoName: %q", rec.rawQuery)
	}
	if repo.RepoID != "new-Project Alpha" {
		t.Errorf("repo.RepoID: got %q, want %q", repo.RepoID, "new-Project Alpha")
	}
	if repo.RepoName != "Project Alpha" {
		t.Errorf("repo.RepoName: got %q, want %q", repo.RepoName, "Project Alpha")
	}
}

func TestCreate_EmptyName(t *testing.T) {
	c := &Client{}
	if _, err := c.Create(context.Background(), ""); err == nil {
		t.Error("expected error for empty name")
	}
}

func TestCreate_CodeNonZero(t *testing.T) {
	c, _ := mutateServer(t, mutateServerOpts{
		body: `{"code":-1,"message":"name already used"}`,
	})
	_, err := c.Create(context.Background(), "dup")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "name already used") || !strings.Contains(err.Error(), "code -1") {
		t.Errorf("error %q lacks expected substrings", err)
	}
}

func TestCreate_NoRepoIDReturned(t *testing.T) {
	// Server returns success but no repo metadata — the helper
	// should refuse to silently succeed (otherwise the caller would
	// get a Repo with empty fields and propagate that downstream).
	c, _ := mutateServer(t, mutateServerOpts{
		body: `{"code":0}`,
	})
	_, err := c.Create(context.Background(), "x")
	if err == nil {
		t.Fatal("expected error for missing repo_id, got nil")
	}
	if !strings.Contains(err.Error(), "did not return a repo_id") {
		t.Errorf("error %q lacks expected hint", err)
	}
}

func TestCreate_HTTPError(t *testing.T) {
	c, _ := mutateServer(t, mutateServerOpts{
		status: http.StatusForbidden,
		body:   `forbidden`,
	})
	_, err := c.Create(context.Background(), "x")
	if err == nil {
		t.Fatal("expected error")
	}
	var hErr *HTTPError
	if !errors.As(err, &hErr) || hErr.Status != http.StatusForbidden {
		t.Errorf("expected HTTP 403 wrapper, got %T (%v)", err, err)
	}
}

func TestRename_OK(t *testing.T) {
	c, rec := mutateServer(t, mutateServerOpts{})
	if err := c.Rename(context.Background(), "abc-123", "New Name"); err != nil {
		t.Fatalf("Rename: %v", err)
	}
	if rec.method != http.MethodPatch {
		t.Errorf("method: got %q, want PATCH", rec.method)
	}
	if got := rec.query.Get("repoId"); got != "abc-123" {
		t.Errorf("repoId: got %q, want %q", got, "abc-123")
	}
	if got := rec.query.Get("destination"); got != "New Name" {
		t.Errorf("destination: got %q, want %q", got, "New Name")
	}
}

func TestRename_PlainTextSuccessBody(t *testing.T) {
	// Some gateways return HTTP 200 and the bare word "success" instead
	// of a JSON object for PATCH /api/repos/.
	c, _ := mutateServer(t, mutateServerOpts{body: "success"})
	if err := c.Rename(context.Background(), "fad67b90-4641-4d76-a05a-3c84198fffef", "test22"); err != nil {
		t.Fatalf("Rename: %v", err)
	}
}

func TestRename_EmptyArgs(t *testing.T) {
	c := &Client{}
	if err := c.Rename(context.Background(), "", "x"); err == nil {
		t.Error("expected error for empty repoID")
	}
	if err := c.Rename(context.Background(), "abc", ""); err == nil {
		t.Error("expected error for empty newName")
	}
}

func TestRename_CodeNonZero(t *testing.T) {
	c, _ := mutateServer(t, mutateServerOpts{
		body: `{"code":403,"message":"permission denied"}`,
	})
	err := c.Rename(context.Background(), "abc", "X")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "permission denied") || !strings.Contains(err.Error(), "code 403") {
		t.Errorf("error %q lacks expected substrings", err)
	}
}

func TestRename_HTTPError(t *testing.T) {
	c, _ := mutateServer(t, mutateServerOpts{
		status: http.StatusNotFound,
		body:   `not found`,
	})
	err := c.Rename(context.Background(), "missing", "X")
	if err == nil {
		t.Fatal("expected error")
	}
	var hErr *HTTPError
	if !errors.As(err, &hErr) || hErr.Status != http.StatusNotFound {
		t.Errorf("expected HTTP 404 wrapper, got %T (%v)", err, err)
	}
}

func TestDelete_OK(t *testing.T) {
	c, rec := mutateServer(t, mutateServerOpts{})
	if err := c.Delete(context.Background(), "abc-123"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if rec.method != http.MethodDelete {
		t.Errorf("method: got %q, want DELETE", rec.method)
	}
	if got := rec.query.Get("repoId"); got != "abc-123" {
		t.Errorf("repoId: got %q, want %q", got, "abc-123")
	}
	// Sanity-check no extra params leak into the wire shape — the
	// front-end's deleteRepo sends ONLY repoId.
	if len(rec.query) != 1 {
		t.Errorf("expected exactly 1 query param, got %d (%q)", len(rec.query), rec.rawQuery)
	}
}

func TestDelete_EmptyID(t *testing.T) {
	c := &Client{}
	if err := c.Delete(context.Background(), ""); err == nil {
		t.Error("expected error for empty repoID")
	}
}

func TestDelete_CodeNonZero(t *testing.T) {
	c, _ := mutateServer(t, mutateServerOpts{
		body: `{"code":-1,"message":"repo locked"}`,
	})
	err := c.Delete(context.Background(), "abc")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "repo locked") || !strings.Contains(err.Error(), "code -1") {
		t.Errorf("error %q lacks expected substrings", err)
	}
}

func TestDelete_HTTPError(t *testing.T) {
	c, _ := mutateServer(t, mutateServerOpts{
		status: http.StatusUnauthorized,
		body:   `auth`,
	})
	err := c.Delete(context.Background(), "abc")
	if err == nil {
		t.Fatal("expected error")
	}
	var hErr *HTTPError
	if !errors.As(err, &hErr) || hErr.Status != http.StatusUnauthorized {
		t.Errorf("expected HTTP 401 wrapper, got %T (%v)", err, err)
	}
}

// TestMutate_NilOutTolerated guards the do() change that lets out
// be nil. Not a public API surface, but easy to assert via a direct
// call so we don't regress when refactoring do().
func TestMutate_NilOutTolerated(t *testing.T) {
	c, _ := mutateServer(t, mutateServerOpts{
		body: `{"code":0,"message":"OK"}`,
	})
	if err := c.do(context.Background(), http.MethodDelete, c.BaseURL+"/api/repos/?repoId=x", nil); err != nil {
		t.Errorf("nil out should be tolerated, got: %v", err)
	}
}
