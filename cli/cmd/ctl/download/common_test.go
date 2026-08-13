package download

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"

	"github.com/beclab/Olares/cli/pkg/cmdutil"
	"github.com/beclab/Olares/cli/pkg/credential"
	"github.com/beclab/Olares/cli/pkg/whoami"
)

type fakeDoer struct {
	lastMethod string
	lastPath   string
	lastBody   interface{}
	resp       []byte
	err        error
}

func TestPrepareVersionGate(t *testing.T) {
	prev := viper.GetString(cmdutil.FlagOlaresVersion)
	t.Cleanup(func() { viper.Set(cmdutil.FlagOlaresVersion, prev) })

	t.Run("1.12.6 is rejected before profile preparation", func(t *testing.T) {
		viper.Set(cmdutil.FlagOlaresVersion, "1.12.6")
		pc, err := prepare(context.Background(), cmdutil.NewFactory())
		if err == nil || pc != nil {
			t.Fatalf("expected version-gate failure, got client=%v err=%v", pc, err)
		}
		if !strings.Contains(err.Error(), "knowledge download") ||
			!strings.Contains(err.Error(), "requires Olares >= 1.12.7") {
			t.Fatalf("unexpected gate error: %v", err)
		}
	})

	for _, version := range []string{"1.12.7", "1.12.7-20260723"} {
		t.Run(version+" passes the gate", func(t *testing.T) {
			viper.Set(cmdutil.FlagOlaresVersion, version)
			_, err := prepare(context.Background(), cmdutil.NewFactory())
			if err != nil && (strings.Contains(err.Error(), "requires Olares") ||
				strings.Contains(err.Error(), "backend version could not be determined")) {
				t.Fatalf("version %s should pass the gate: %v", version, err)
			}
		})
	}

	t.Run("undetectable version fails closed with profile refresh", func(t *testing.T) {
		viper.Set(cmdutil.FlagOlaresVersion, "not-a-version")
		pc, err := prepare(context.Background(), cmdutil.NewFactory())
		if err == nil || pc != nil {
			t.Fatalf("expected fail-closed version error, got client=%v err=%v", pc, err)
		}
		for _, want := range []string{"could not be determined", "profile login", "profile list --refresh-version"} {
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("error %q missing %q", err, want)
			}
		}
	})
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

func TestDoMutateAddsRecoveryForKnownTaskErrors(t *testing.T) {
	d := &fakeDoer{resp: []byte(`{"code":409,"message":"task already exists"}`)}
	err := doMutate(context.Background(), d, "POST", "/api/download", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	for _, want := range []string{"task already exists", "olares-cli knowledge download list"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q missing %q", err, want)
		}
	}

	d = &fakeDoer{resp: []byte(`{"code":409,"message":"preference already exists"}`)}
	err = doMutate(context.Background(), d, "PUT", "/api/user/preferences", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "knowledge download list") {
		t.Fatalf("unrelated conflict received task recovery: %q", err)
	}

	d = &fakeDoer{resp: []byte(`{"code":404,"message":"task not found"}`)}
	err = doMutate(context.Background(), d, "PUT", "/api/download/pause/42", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "olares-cli knowledge download list") {
		t.Fatalf("task 404 should refresh IDs, got %v", err)
	}

	d = &fakeDoer{resp: []byte(`{"code":409,"message":"GID xxxx isalready registered"}`)}
	err = doMutate(context.Background(), d, "POST", "/api/download", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	for _, want := range []string{"isalready registered", "olares-cli knowledge download list"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("duplicate torrent %q missing %q", err, want)
		}
	}

	d = &fakeDoer{resp: []byte(`{"code":500,"message":"yt-dlp unavailable: connection refused"}`)}
	err = doMutate(context.Background(), d, "POST", "/api/download", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	for _, want := range []string{"yt-dlp unavailable", ytdlpMarketInstall} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("yt-dlp create failure %q missing %q", err, want)
		}
	}

	d = &fakeDoer{resp: []byte(`{"code":500,"message":"timeout of 3000ms exceeded"}`)}
	err = doGet(context.Background(), d, "/api/url/inspect?url=https://example.com", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	for _, want := range []string{"timeout of 3000ms", "channel/RSS probes", "create the task directly"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("inspect timeout %q missing %q", err, want)
		}
	}

	d = &fakeDoer{resp: []byte(`{"code":409,"message":"task is moving to its destination"}`)}
	err = doMutate(context.Background(), d, "PUT", "/api/download/cancel/42", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	for _, want := range []string{"wait for the move to finish", "olares-cli knowledge download info 42"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("mid-move conflict %q missing %q", err, want)
		}
	}
}

func TestDoMutateAddsRecoveryForHTTPTaskErrors(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		path    string
		want    []string
		notWant string
	}{
		{
			name:    "mid-move conflict",
			status:  http.StatusConflict,
			path:    "/api/download/cancel/42",
			want:    []string{"wait for the move to finish", "olares-cli knowledge download info 42"},
			notWant: "knowledge download status",
		},
		{
			name:   "missing task",
			status: http.StatusNotFound,
			path:   "/api/download/pause/42",
			want:   []string{"refresh task IDs", "olares-cli knowledge download list"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(test.status)
			}))
			t.Cleanup(server.Close)

			client := whoami.NewHTTPClient(server.Client(), server.URL, "")
			err := doMutate(context.Background(), client, "POST", test.path, nil, nil)
			if err == nil {
				t.Fatal("expected error")
			}
			for _, want := range test.want {
				if !strings.Contains(err.Error(), want) {
					t.Fatalf("error %q missing %q", err, want)
				}
			}
			if test.notWant != "" && strings.Contains(err.Error(), test.notWant) {
				t.Fatalf("error %q contains obsolete CTA %q", err, test.notWant)
			}
		})
	}
}

func TestDoMutateAddsRecoveryForRemoveConflict(t *testing.T) {
	d := &fakeDoer{resp: []byte(`{"code":409}`)}
	err := doMutate(
		context.Background(),
		d,
		http.MethodDelete,
		"/api/download/remove",
		RemoveReq{TaskID: 42},
		nil,
	)
	if err == nil {
		t.Fatal("expected error")
	}
	for _, want := range []string{"wait for the move to finish", "olares-cli knowledge download info 42"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("remove conflict %q missing %q", err, want)
		}
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

func TestDoMutateBatchEnvelope(t *testing.T) {
	d := &fakeDoer{resp: []byte(`{"code":200,"succeeded":[1,2],"failed":[{"task_id":3,"error":"not found"}]}`)}
	var res BatchResult
	if err := doMutate(context.Background(), d, "PUT", "/api/download/batch/pause", BatchReq{TaskIDs: []int64{1, 2, 3}}, &res); err != nil {
		t.Fatal(err)
	}
	if len(res.Succeeded) != 2 || res.Succeeded[0] != 1 || res.Succeeded[1] != 2 {
		t.Fatalf("succeeded: %+v", res.Succeeded)
	}
	if len(res.Failed) != 1 || res.Failed[0].TaskID != 3 || res.Failed[0].Error != "not found" {
		t.Fatalf("failed: %+v", res.Failed)
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

func TestValidateApp(t *testing.T) {
	for _, valid := range []string{"", "wise", " larepass ", "larepass"} {
		got, err := validateApp(valid)
		if err != nil {
			t.Fatalf("validateApp(%q) unexpected error: %v", valid, err)
		}
		switch strings.TrimSpace(valid) {
		case "", "wise":
			if got != "wise" {
				t.Fatalf("validateApp(%q)=%q want wise", valid, got)
			}
		default:
			if got != "larepass" {
				t.Fatalf("validateApp(%q)=%q want larepass", valid, got)
			}
		}
	}
	_, err := validateApp("namespace")
	if err == nil {
		t.Fatal("unknown app should fail")
	}
	for _, want := range []string{"unsupported --app", allowedApps} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("validateApp error %q missing %q", err, want)
		}
	}
}

func TestValidateNonNegativeFlag(t *testing.T) {
	if err := validateNonNegativeFlag("--page", 0); err != nil {
		t.Fatal(err)
	}
	if err := validateNonNegativeFlag("--page", 2); err != nil {
		t.Fatal(err)
	}
	err := validateNonNegativeFlag("--page", -1)
	if err == nil || !strings.Contains(err.Error(), "unsupported --page") {
		t.Fatalf("negative page should fail, got %v", err)
	}
}

func TestValidateLimit(t *testing.T) {
	if err := validateLimit(0); err != nil {
		t.Fatal(err)
	}
	if err := validateLimit(100); err != nil {
		t.Fatal(err)
	}
	if err := validateLimit(-1); err == nil || !strings.Contains(err.Error(), "unsupported --limit") {
		t.Fatalf("negative limit should fail, got %v", err)
	}
	if err := validateLimit(101); err == nil || !strings.Contains(err.Error(), "max 100") {
		t.Fatalf("over-max limit should fail, got %v", err)
	}
}

func TestValidateSinceID(t *testing.T) {
	if err := validateSinceID(0); err != nil {
		t.Fatal(err)
	}
	if err := validateSinceID(-1); err == nil || !strings.Contains(err.Error(), "unsupported --since-id") {
		t.Fatalf("negative since-id should fail, got %v", err)
	}
}

func TestValidateDrivePath(t *testing.T) {
	for _, valid := range []string{"drive/Home/Downloads/x", "drive/Data/cache/y"} {
		if err := validateDrivePath(valid); err != nil {
			t.Fatalf("validateDrivePath(%q): %v", valid, err)
		}
	}
	for _, bad := range []string{"", "Home/Downloads/x", "drive/home/x", "/Files/Home/x"} {
		if err := validateDrivePath(bad); err == nil {
			t.Fatalf("validateDrivePath(%q) should fail", bad)
		}
	}
}

func TestInspectYTDLPHint(t *testing.T) {
	falseVal := false
	trueVal := true
	if hint := inspectYTDLPHint(InspectData{Provider: "yt-dlp", Available: &falseVal}); !strings.Contains(hint, ytdlpMarketInstall) {
		t.Fatalf("Available=false should hint install, got %q", hint)
	}
	if hint := inspectYTDLPHint(InspectData{Provider: "aria2", Available: &falseVal}); hint != "" {
		t.Fatalf("aria2 Available=false should not hint yt-dlp, got %q", hint)
	}
	if hint := inspectYTDLPHint(InspectData{Provider: "yt-dlp", Available: &trueVal}); hint != "" {
		t.Fatalf("Available=true should not hint, got %q", hint)
	}
	if hint := inspectYTDLPHint(InspectData{Error: "yt-dlp daemon unreachable"}); !strings.Contains(hint, ytdlpMarketInstall) {
		t.Fatalf("error text should hint install, got %q", hint)
	}
}

func TestShouldHintInspectTimeout(t *testing.T) {
	if !shouldHintInspectTimeout("timeout of 3000ms exceeded") {
		t.Fatal("expected timeout hint")
	}
	if shouldHintInspectTimeout("bad url") {
		t.Fatal("non-timeout should not hint")
	}
}
