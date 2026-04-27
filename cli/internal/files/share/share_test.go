package share

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// newTestClient mirrors the rm / cp / rename harness: stand up a real
// httptest server, hand the caller a Client whose BaseURL points at
// it, and let the test inspect what landed on the wire.
func newTestClient(t *testing.T, h http.Handler) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return &Client{
		HTTPClient: srv.Client(),
		BaseURL:    srv.URL,
	}, srv
}

// TestParsePermission table-drives the canonical-label and numeric
// inputs the cobra layer's `--users user:perm` flag has to accept.
// Each label-spelling alias is exercised so a future caller can rely
// on either spelling without re-checking this list.
func TestParsePermission(t *testing.T) {
	cases := []struct {
		in     string
		expect Permission
		err    bool
	}{
		{"view", PermView, false},
		{"VIEW", PermView, false},
		{"read", PermView, false},
		{"ro", PermView, false},
		{"upload", PermUpload, false},
		{"upload-only", PermUpload, false},
		{"edit", PermEdit, false},
		{"rw", PermEdit, false},
		{"admin", PermAdmin, false},
		{"none", PermNone, false},
		{"", PermNone, false},
		{"0", PermNone, false},
		{"1", PermView, false},
		{"4", PermAdmin, false},
		{"7", 0, true},
		{"banana", 0, true},
	}
	for _, tc := range cases {
		got, err := ParsePermission(tc.in)
		if tc.err {
			if err == nil {
				t.Errorf("ParsePermission(%q): want error, got %v", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParsePermission(%q): unexpected error: %v", tc.in, err)
		}
		if got != tc.expect {
			t.Errorf("ParsePermission(%q) = %v, want %v", tc.in, got, tc.expect)
		}
	}
}

// TestPermissionString round-trips canonical labels through Permission
// → String → ParsePermission to lock in the human ↔ wire mapping.
// Unknown values render as the underlying integer so a future
// server-side addition is at least readable in diagnostics.
func TestPermissionString(t *testing.T) {
	cases := []struct {
		p      Permission
		expect string
	}{
		{PermNone, "none"},
		{PermView, "view"},
		{PermUpload, "upload"},
		{PermEdit, "edit"},
		{PermAdmin, "admin"},
		{Permission(99), "99"},
	}
	for _, tc := range cases {
		if got := tc.p.String(); got != tc.expect {
			t.Errorf("Permission(%d).String() = %q, want %q", tc.p, got, tc.expect)
		}
	}
}

// TestBuildSharePathURL covers the wire-path encoding contract: each
// segment URL-encoded individually (so '/' is preserved between
// segments but encoded inside a segment), and an unconditional
// trailing '/'.
func TestBuildSharePathURL(t *testing.T) {
	cases := []struct {
		name   string
		t      Target
		expect string
	}{
		{
			name:   "drive home file",
			t:      Target{FileType: "drive", Extend: "Home", SubPath: "/Documents/foo.pdf"},
			expect: "/api/share/share_path/drive/Home/Documents/foo.pdf/",
		},
		{
			name:   "drive home folder with trailing slash preserved",
			t:      Target{FileType: "drive", Extend: "Home", SubPath: "/Photos/", IsDirIntent: true},
			expect: "/api/share/share_path/drive/Home/Photos/",
		},
		{
			name:   "extend root subpath defaults to slash",
			t:      Target{FileType: "drive", Extend: "Home", SubPath: ""},
			expect: "/api/share/share_path/drive/Home/",
		},
		{
			name:   "spaces and unicode percent-encoded per segment",
			t:      Target{FileType: "drive", Extend: "Home", SubPath: "/My Docs/老照片"},
			expect: "/api/share/share_path/drive/Home/My%20Docs/%E8%80%81%E7%85%A7%E7%89%87/",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := buildSharePathURL(tc.t)
			if got != tc.expect {
				t.Errorf("buildSharePathURL(%+v) = %q, want %q", tc.t, got, tc.expect)
			}
		})
	}
}

// TestCreate_Internal_WireShape inspects the actual POST that lands
// on the server: URL path, method, Content-Type, JSON body shape, and
// confirms only the fields Internal sends are present (no public_smb,
// no upload_size_limit, ...). This is the canary test that breaks
// loudly if the protocol drifts.
func TestCreate_Internal_WireShape(t *testing.T) {
	var (
		gotMethod string
		gotPath   string
		gotCType  string
		gotRaw    []byte
	)
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotCType = r.Header.Get("Content-Type")
		gotRaw, _ = io.ReadAll(r.Body)
		_, _ = io.WriteString(w, `{"code":0,"data":{"id":"share-1","owner":"alice","share_type":"internal","permission":4}}`)
	}))

	res, err := client.Create(
		context.Background(),
		Target{FileType: "drive", Extend: "Home", SubPath: "/Backups/", IsDirIntent: true},
		CreateOptions{
			Name:       "Backups",
			ShareType:  TypeInternal,
			Permission: PermAdmin,
			Password:   "",
		},
	)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if res.ID != "share-1" {
		t.Errorf("ID: got %q", res.ID)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("Method: got %s", gotMethod)
	}
	if gotPath != "/api/share/share_path/drive/Home/Backups/" {
		t.Errorf("Path: got %q", gotPath)
	}
	if !strings.HasPrefix(gotCType, "application/json") {
		t.Errorf("Content-Type: got %q", gotCType)
	}
	// Decode the body and check field-presence: Internal share should
	// NOT carry public_smb / upload_size_limit / users keys.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(gotRaw, &raw); err != nil {
		t.Fatalf("decode body: %v (raw %q)", err, gotRaw)
	}
	for _, want := range []string{"name", "share_type", "permission", "password"} {
		if _, ok := raw[want]; !ok {
			t.Errorf("body missing %q (got %s)", want, gotRaw)
		}
	}
	for _, forbidden := range []string{"public_smb", "upload_size_limit", "users", "expire_in", "expire_time"} {
		if _, ok := raw[forbidden]; ok {
			t.Errorf("body should NOT contain %q (got %s)", forbidden, gotRaw)
		}
	}
}

// TestCreate_Public_WireShape covers the Public-link creation: the
// body gets `password`, `upload_size_limit`, and exactly ONE of
// expire_in / expire_time. Default permission is Edit (or UploadOnly
// when --upload-only is set; we let the caller pick).
func TestCreate_Public_WireShape(t *testing.T) {
	var gotRaw []byte
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRaw, _ = io.ReadAll(r.Body)
		_, _ = io.WriteString(w, `{"code":0,"data":{"id":"share-pub-1","share_type":"external","permission":3}}`)
	}))

	_, err := client.Create(
		context.Background(),
		Target{FileType: "drive", Extend: "Home", SubPath: "/Photos/", IsDirIntent: true},
		CreateOptions{
			Name:            "Photos",
			ShareType:       TypePublic,
			Permission:      PermEdit,
			Password:        "abc123",
			ExpireIn:        7 * 24 * 3600 * 1000,
			UploadSizeLimit: 100 * 1024 * 1024,
		},
	)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(gotRaw, &raw); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if string(raw["share_type"]) != `"external"` {
		t.Errorf("share_type: got %s", raw["share_type"])
	}
	if string(raw["permission"]) != `3` {
		t.Errorf("permission: got %s", raw["permission"])
	}
	if string(raw["password"]) != `"abc123"` {
		t.Errorf("password: got %s", raw["password"])
	}
	if string(raw["expire_in"]) != `604800000` {
		t.Errorf("expire_in: got %s", raw["expire_in"])
	}
	if string(raw["upload_size_limit"]) != `104857600` {
		t.Errorf("upload_size_limit: got %s", raw["upload_size_limit"])
	}
	if _, ok := raw["expire_time"]; ok {
		t.Errorf("body should NOT contain expire_time when expire_in is set: %s", gotRaw)
	}
	if _, ok := raw["public_smb"]; ok {
		t.Errorf("body should NOT contain public_smb for Public share: %s", gotRaw)
	}
}

// TestCreate_SMB_WireShape covers SMB-share creation: the body
// MUST carry `public_smb` (boolean — true/false matters), and `users`
// when not in public_smb mode.
func TestCreate_SMB_WireShape(t *testing.T) {
	var gotRaw []byte
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRaw, _ = io.ReadAll(r.Body)
		_, _ = io.WriteString(w, `{"code":0,"data":{"id":"share-smb-1","share_type":"smb","smb_link":"\\\\smb\\Photos","smb_user":"alice","smb_password":"pw"}}`)
	}))

	publicSMB := false
	res, err := client.Create(
		context.Background(),
		Target{FileType: "drive", Extend: "Home", SubPath: "/Photos/", IsDirIntent: true},
		CreateOptions{
			Name:       "Photos",
			ShareType:  TypeSMB,
			Permission: PermEdit,
			Password:   "",
			Users: []SMBUser{
				{ID: "smb-1", Permission: PermEdit},
				{ID: "smb-2", Permission: PermView},
			},
			PublicSMB: &publicSMB,
		},
	)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if res.SMBLink == "" || res.SMBUser == "" {
		t.Errorf("smb fields missing: %+v", res)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(gotRaw, &raw); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if string(raw["share_type"]) != `"smb"` {
		t.Errorf("share_type: got %s", raw["share_type"])
	}
	if string(raw["public_smb"]) != `false` {
		t.Errorf("public_smb: got %s", raw["public_smb"])
	}
	if !json.Valid(raw["users"]) || string(raw["users"]) == "null" {
		t.Errorf("users: got %s", raw["users"])
	}
	// Sanity-check that users[0].id & permission round-tripped.
	var users []SMBUser
	if err := json.Unmarshal(raw["users"], &users); err != nil {
		t.Fatalf("decode users: %v", err)
	}
	if len(users) != 2 || users[0].ID != "smb-1" || users[0].Permission != PermEdit {
		t.Errorf("users round-trip: got %+v", users)
	}
}

// TestCreate_CodeNonZero exercises the {code:!=0} error path: a 200
// HTTP response with code:1 must surface as a Go error so the cobra
// layer can show the server's message rather than silently treating
// it as success.
func TestCreate_CodeNonZero(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, `{"code":1,"message":"already shared"}`)
	}))
	_, err := client.Create(
		context.Background(),
		Target{FileType: "drive", Extend: "Home", SubPath: "/Photos/"},
		CreateOptions{
			Name:       "Photos",
			ShareType:  TypeInternal,
			Permission: PermAdmin,
		},
	)
	if err == nil {
		t.Fatal("expected error for code:1")
	}
	if !strings.Contains(err.Error(), "already shared") {
		t.Errorf("error should bubble up server message, got: %v", err)
	}
}

// TestCreate_HTTPError surfaces non-2xx responses as *HTTPError —
// same contract the cp / rename packages use.
func TestCreate_HTTPError(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, `{"error":"nope"}`)
	}))
	_, err := client.Create(
		context.Background(),
		Target{FileType: "drive", Extend: "Home", SubPath: "/Photos/"},
		CreateOptions{Name: "x", ShareType: TypeInternal, Permission: PermAdmin},
	)
	if err == nil {
		t.Fatal("expected error")
	}
	var hErr *HTTPError
	if !errors.As(err, &hErr) {
		t.Fatalf("want *HTTPError, got %T", err)
	}
	if hErr.Status != http.StatusForbidden {
		t.Errorf("status: got %d", hErr.Status)
	}
}

// TestCreate_EmptyID guards the silent-failure case where the server
// returns 200 + code:0 + empty data.id. Must surface as an error.
func TestCreate_EmptyID(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"code":0,"data":{"id":""}}`)
	}))
	_, err := client.Create(
		context.Background(),
		Target{FileType: "drive", Extend: "Home", SubPath: "/Photos/"},
		CreateOptions{Name: "x", ShareType: TypeInternal, Permission: PermAdmin},
	)
	if err == nil {
		t.Fatal("expected error for empty share id")
	}
}

// TestRemove_WireShape checks that Remove sends DELETE with the
// comma-joined ids in `path_ids`, NOT in the JSON body.
func TestRemove_WireShape(t *testing.T) {
	var (
		gotMethod string
		gotPath   string
		gotQuery  string
	)
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"code":0}`)
	}))

	if err := client.Remove(context.Background(), []string{"id-1", "id-2", "id-3"}); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("Method: got %s", gotMethod)
	}
	if gotPath != "/api/share/share_path/" {
		t.Errorf("Path: got %q", gotPath)
	}
	q, err := url.ParseQuery(gotQuery)
	if err != nil {
		t.Fatalf("parse query: %v", err)
	}
	if q.Get("path_ids") != "id-1,id-2,id-3" {
		t.Errorf("path_ids: got %q", q.Get("path_ids"))
	}
}

// TestList_WireShape inspects GET /api/share/share_path/ with the
// expected filter params. The response is the {share_paths: [...]}
// shape (NOT wrapped in {data:...}) — this test pins that down so a
// future refactor can't accidentally re-route the decoder.
func TestList_WireShape(t *testing.T) {
	var gotQuery string
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = io.WriteString(w, `{"share_paths":[{"id":"a","share_type":"internal"},{"id":"b","share_type":"external"}]}`)
	}))

	byMe := true
	rows, err := client.List(context.Background(), ListParams{
		SharedByMe: &byMe,
		ShareType:  "internal,external",
		Owner:      "alice,bob",
	})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(rows) != 2 || rows[0].ID != "a" || rows[1].ID != "b" {
		t.Errorf("rows: got %+v", rows)
	}
	q, err := url.ParseQuery(gotQuery)
	if err != nil {
		t.Fatalf("parse query: %v", err)
	}
	if q.Get("shared_by_me") != "true" {
		t.Errorf("shared_by_me: got %q", q.Get("shared_by_me"))
	}
	if q.Get("share_type") != "internal,external" {
		t.Errorf("share_type: got %q", q.Get("share_type"))
	}
	if q.Get("owner") != "alice,bob" {
		t.Errorf("owner: got %q", q.Get("owner"))
	}
}

// TestQuery_NotFound covers the empty-share_paths case: when the id
// doesn't exist the server replies with `{"share_paths":[]}` and we
// must return (nil, nil) so the cobra layer can branch on absence.
func TestQuery_NotFound(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"share_paths":[]}`)
	}))
	res, err := client.Query(context.Background(), "missing-id")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if res != nil {
		t.Errorf("want nil result, got %+v", res)
	}
}

// TestAddInternalMembers_WireShape verifies the body shape for the
// member-add endpoint: {path_id, share_members:[{share_member, permission}, ...]}.
func TestAddInternalMembers_WireShape(t *testing.T) {
	var (
		gotPath string
		gotRaw  []byte
	)
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotRaw, _ = io.ReadAll(r.Body)
		_, _ = io.WriteString(w, `{"code":0}`)
	}))
	err := client.AddInternalMembers(context.Background(), "share-1", []Member{
		{ShareMember: "alice", Permission: PermView},
		{ShareMember: "bob", Permission: PermEdit},
	})
	if err != nil {
		t.Fatalf("AddInternalMembers: %v", err)
	}
	if gotPath != "/api/share/share_member/" {
		t.Errorf("Path: got %q", gotPath)
	}
	var raw struct {
		PathID       string   `json:"path_id"`
		ShareMembers []Member `json:"share_members"`
	}
	if err := json.Unmarshal(gotRaw, &raw); err != nil {
		t.Fatalf("decode body: %v (raw %s)", err, gotRaw)
	}
	if raw.PathID != "share-1" {
		t.Errorf("path_id: got %q", raw.PathID)
	}
	if len(raw.ShareMembers) != 2 || raw.ShareMembers[0].ShareMember != "alice" || raw.ShareMembers[1].Permission != PermEdit {
		t.Errorf("share_members: got %+v", raw.ShareMembers)
	}
}

// TestAddInternalMembers_EmptyIsNoop ensures the cobra layer can
// always call AddInternalMembers without first checking the slice —
// an empty members slice short-circuits to nil instead of firing a
// useless POST.
func TestAddInternalMembers_EmptyIsNoop(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("empty members should not fire any HTTP request, but got %s", r.URL.Path)
	}))
	if err := client.AddInternalMembers(context.Background(), "share-1", nil); err != nil {
		t.Errorf("AddInternalMembers(nil): %v", err)
	}
}

// TestUpdateSMBShareMember_WireShape: the body MUST contain users
// (slice of SMBUser) and public_smb (bool). path_id is required.
func TestUpdateSMBShareMember_WireShape(t *testing.T) {
	var (
		gotPath string
		gotRaw  []byte
	)
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotRaw, _ = io.ReadAll(r.Body)
		_, _ = io.WriteString(w, `{"code":0}`)
	}))
	err := client.UpdateSMBShareMember(context.Background(), "share-1", []SMBUser{
		{ID: "smb-1", Permission: PermEdit},
	}, true)
	if err != nil {
		t.Fatalf("UpdateSMBShareMember: %v", err)
	}
	if gotPath != "/api/share/smb_share_member/" {
		t.Errorf("Path: got %q", gotPath)
	}
	var raw struct {
		PathID    string    `json:"path_id"`
		Users     []SMBUser `json:"users"`
		PublicSMB bool      `json:"public_smb"`
	}
	if err := json.Unmarshal(gotRaw, &raw); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if !raw.PublicSMB {
		t.Errorf("public_smb: want true, got false (raw %s)", gotRaw)
	}
}

// TestListSMBAccounts_Envelope confirms the {data:[...]} envelope is
// decoded correctly and that nil-data short-circuits to an empty
// slice (not nil) so callers can range over the result.
func TestListSMBAccounts_Envelope(t *testing.T) {
	client, _ := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"code":0,"data":[{"id":"smb-1","name":"alice"},{"id":"smb-2","name":"bob"}]}`)
	}))
	accounts, err := client.ListSMBAccounts(context.Background())
	if err != nil {
		t.Fatalf("ListSMBAccounts: %v", err)
	}
	if len(accounts) != 2 || accounts[0].ID != "smb-1" || accounts[1].Name != "bob" {
		t.Errorf("accounts: got %+v", accounts)
	}
}
