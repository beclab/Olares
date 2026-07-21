package download

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/beclab/Olares/cli/pkg/credential"
)

type fakeDoer struct {
	lastMethod string
	lastPath   string
	lastBody   interface{}
	resp       []byte
	err        error
}

func (f *fakeDoer) DoJSON(_ context.Context, method, path string, body, out interface{}) error {
	f.lastMethod = method
	f.lastPath = path
	f.lastBody = body
	if f.err != nil {
		return f.err
	}
	if out == nil || len(f.resp) == 0 {
		return nil
	}
	return json.Unmarshal(f.resp, out)
}

func TestEdgeBase(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"https://settings.alice.olares.com", "https://settings.alice.olares.com/download"},
		{"https://settings.alice.olares.com/", "https://settings.alice.olares.com/download"},
		{"", "/download"},
	}
	for _, tc := range cases {
		got := edgeBase(&credential.ResolvedProfile{SettingsURL: tc.in})
		if got != tc.want {
			t.Fatalf("edgeBase(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
	if edgeBase(nil) != "" {
		t.Fatal("edgeBase(nil) should be empty")
	}
}

func TestDoMutateDataEnvelope(t *testing.T) {
	d := &fakeDoer{resp: []byte(`{"code":200,"data":{"id":42,"status":"waiting","app":"wise","url":"https://ex","file_name":"a.mp4"}}`)}
	var task DownloadTask
	if err := doMutate(context.Background(), d, "POST", "/api/download", NewDownloadReq{URL: "https://ex", App: "wise"}, &task); err != nil {
		t.Fatal(err)
	}
	if task.ID != 42 || task.Status != "waiting" || task.FileName != "a.mp4" {
		t.Fatalf("unexpected task: %+v", task)
	}
	if d.lastMethod != "POST" || d.lastPath != "/api/download" {
		t.Fatalf("unexpected call %s %s", d.lastMethod, d.lastPath)
	}
}

func TestDoMutateListEnvelope(t *testing.T) {
	d := &fakeDoer{resp: []byte(`{"code":200,"total":2,"list":[{"id":1,"status":"downloading","percent":10.5,"app":"wise","file_name":"a"},{"id":2,"status":"paused","percent":0,"app":"wise","file_name":"b"}]}`)}
	var result ListResult
	if err := doGet(context.Background(), d, "/api/download/list?app=wise", &result); err != nil {
		t.Fatal(err)
	}
	if result.Total != 2 || len(result.List) != 2 || result.List[0].ID != 1 {
		t.Fatalf("unexpected list: %+v", result)
	}
}

func TestDoMutateErrorCode(t *testing.T) {
	d := &fakeDoer{resp: []byte(`{"code":400,"message":"bad url"}`)}
	err := doMutate(context.Background(), d, "POST", "/api/download", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "code 400") || !strings.Contains(err.Error(), "bad url") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoMutateCodeZeroOK(t *testing.T) {
	d := &fakeDoer{resp: []byte(`{"code":0,"data":{"provider":"aria2"}}`)}
	var data InspectData
	if err := doGet(context.Background(), d, "/api/url/inspect?url=x", &data); err != nil {
		t.Fatal(err)
	}
	if data.Provider != "aria2" {
		t.Fatalf("got %+v", data)
	}
}

func TestDoMutateTransportError(t *testing.T) {
	d := &fakeDoer{err: fmt.Errorf("boom")}
	if err := doGet(context.Background(), d, "/api/download/list", nil); err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected boom, got %v", err)
	}
}

func TestDoMutateFileCheckEnvelope(t *testing.T) {
	d := &fakeDoer{resp: []byte(`{"code":200,"exist":true}`)}
	var res FileCheckResult
	if err := doGet(context.Background(), d, "/api/download/file_check/none?user=alice&path=drive/Home/x", &res); err != nil {
		t.Fatal(err)
	}
	if !res.Exist {
		t.Fatalf("expected Exist=true, got %+v", res)
	}
}

func TestDoMutateTorrentInspectEnvelope(t *testing.T) {
	d := &fakeDoer{resp: []byte(`{"code":200,"data":{"name":"x","files":[{"index":1,"path":"a","length":5}]}}`)}
	var res TorrentInspectResult
	if err := doMutate(context.Background(), d, "POST", "/api/download/torrent/inspect", TorrentInspectReq{TorrentFileB64: "AA=="}, &res); err != nil {
		t.Fatal(err)
	}
	if res.Name != "x" || len(res.Files) != 1 || res.Files[0].Index != 1 || res.Files[0].Length != 5 {
		t.Fatalf("unexpected inspect result: %+v", res)
	}
}

func TestDoMutateTorrentFilesBody(t *testing.T) {
	d := &fakeDoer{resp: []byte(`{"code":200,"data":{"task_id":7,"selected":[1,3]}}`)}
	var res SetTorrentFilesResult
	if err := doMutate(context.Background(), d, "PUT", "/api/download/7/torrent/files", SetTorrentFilesReq{Selected: []int{1, 3}}, &res); err != nil {
		t.Fatal(err)
	}
	if d.lastMethod != "PUT" || d.lastPath != "/api/download/7/torrent/files" {
		t.Fatalf("unexpected call %s %s", d.lastMethod, d.lastPath)
	}
	body, ok := d.lastBody.(SetTorrentFilesReq)
	if !ok {
		t.Fatalf("unexpected body type %T", d.lastBody)
	}
	if len(body.Selected) != 2 || body.Selected[0] != 1 || body.Selected[1] != 3 {
		t.Fatalf("unexpected body selected: %+v", body.Selected)
	}
	if res.TaskID != 7 || len(res.Selected) != 2 {
		t.Fatalf("unexpected result: %+v", res)
	}
}

func TestParseSelectedIndices(t *testing.T) {
	cases := []struct {
		in      string
		want    []int
		wantErr bool
	}{
		{"1,3,5", []int{1, 3, 5}, false},
		{" 2 , 4 ", []int{2, 4}, false},
		{"all", []int{}, false},
		{"ALL", []int{}, false},
		{"", nil, true},
		{"0", nil, true},
		{"x", nil, true},
		{"1,x", nil, true},
		{"-1", nil, true},
	}
	for _, tc := range cases {
		got, err := parseSelectedIndices(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("parseSelectedIndices(%q) expected error, got %v", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Fatalf("parseSelectedIndices(%q) unexpected error: %v", tc.in, err)
		}
		if len(got) != len(tc.want) {
			t.Fatalf("parseSelectedIndices(%q)=%v want %v", tc.in, got, tc.want)
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Fatalf("parseSelectedIndices(%q)=%v want %v", tc.in, got, tc.want)
			}
		}
	}
}

func TestDoMutateSyncEnvelope(t *testing.T) {
	// The manager returns the top-level {code, list, has_more} envelope
	// (same "list" slot as the list endpoint), NOT {data:{items,next_cursor}}.
	// The composite cursor is derived client-side from the last row.
	d := &fakeDoer{resp: []byte(`{"code":200,"has_more":true,"list":[{"id":7,"status":"downloading","app":"wise","updated_at":"2026-07-20T14:00:00Z"}]}`)}
	var res SyncResult
	if err := doGet(context.Background(), d, "/api/download/sync?limit=100", &res); err != nil {
		t.Fatal(err)
	}
	if len(res.Items) != 1 || res.Items[0].ID != 7 || !res.HasMore {
		t.Fatalf("unexpected sync result: %+v", res)
	}
	gotSince, gotID := res.NextCursor()
	if gotID != 7 || !gotSince.Equal(time.Date(2026, 7, 20, 14, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected next cursor: %s / %d", gotSince, gotID)
	}
}

func TestDoMutateCookieListEnvelope(t *testing.T) {
	d := &fakeDoer{resp: []byte(`{"code":200,"total":1,"list":[{"domain":"youtube.com","provider":"yt-dlp","has_cookie":true,"updated_at":1700000000}]}`)}
	var res CookieListResult
	if err := doGet(context.Background(), d, "/api/integration/cookies", &res); err != nil {
		t.Fatal(err)
	}
	if res.Total != 1 || len(res.List) != 1 || res.List[0].Domain != "youtube.com" || !res.List[0].HasCookie {
		t.Fatalf("unexpected cookie list: %+v", res)
	}
}

func TestSettingsUpdateKeyValueBody(t *testing.T) {
	// PUT /api/system/settings is a single {key,value} pair (not a
	// whole-object patch); the manager rejects a missing key with 400
	// "key is required". Assert the CLI sends the manager-expected shape
	// and decodes the echoed snapshot.
	d := &fakeDoer{resp: []byte(`{"code":200,"data":{"aria2_max_concurrent":3}}`)}
	req := SystemSettingUpdateReq{Key: systemSettingAria2MaxConcurrent, Value: 3}
	var s SystemSettings
	if err := doMutate(context.Background(), d, "PUT", "/api/system/settings", req, &s); err != nil {
		t.Fatal(err)
	}
	if d.lastMethod != "PUT" || d.lastPath != "/api/system/settings" {
		t.Fatalf("unexpected call %s %s", d.lastMethod, d.lastPath)
	}
	body, ok := d.lastBody.(SystemSettingUpdateReq)
	if !ok {
		t.Fatalf("unexpected body type %T", d.lastBody)
	}
	if body.Key != "aria2_max_concurrent" {
		t.Fatalf("unexpected key: %q", body.Key)
	}
	if s.Aria2MaxConcurrent != 3 {
		t.Fatalf("unexpected settings: %+v", s)
	}
}

func TestParseTaskID(t *testing.T) {
	id, err := parseTaskID("99")
	if err != nil || id != 99 {
		t.Fatalf("got %d %v", id, err)
	}
	if _, err := parseTaskID("0"); err == nil {
		t.Fatal("expected error for 0")
	}
	if _, err := parseTaskID("x"); err == nil {
		t.Fatal("expected error for non-int")
	}
}
