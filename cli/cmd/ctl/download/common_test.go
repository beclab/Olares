package download

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

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
