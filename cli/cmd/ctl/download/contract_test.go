package download

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestDownloadTaskDecodesWillAutoRetry(t *testing.T) {
	raw := []byte(`{
		"id": 9,
		"status": "error",
		"app": "wise",
		"url": "https://ex",
		"file_name": "a.mp4",
		"will_auto_retry": true,
		"created_at": "2026-01-01T00:00:00Z",
		"updated_at": "2026-01-01T00:00:00Z"
	}`)
	var task DownloadTask
	if err := json.Unmarshal(raw, &task); err != nil {
		t.Fatal(err)
	}
	if !task.WillAutoRetry {
		t.Fatalf("WillAutoRetry = false; want true")
	}
}

func TestDoMutateListPreservesWillAutoRetry(t *testing.T) {
	d := &fakeDoer{resp: []byte(`{
		"code":200,
		"total":1,
		"list":[{"id":1,"status":"error","app":"wise","file_name":"a","will_auto_retry":true,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:00:00Z"}]
	}`)}
	var result ListResult
	if err := doGet(context.Background(), d, "/api/download/list?app=wise", &result); err != nil {
		t.Fatal(err)
	}
	if len(result.List) != 1 || !result.List[0].WillAutoRetry {
		t.Fatalf("unexpected list: %+v", result.List)
	}
}

func TestDoMutateSyncPreservesWillAutoRetry(t *testing.T) {
	d := &fakeDoer{resp: []byte(`{
		"code":200,
		"has_more":false,
		"list":[{"id":7,"status":"error","app":"wise","file_name":"b","will_auto_retry":true,"created_at":"2026-01-01T00:00:00Z","updated_at":"2026-01-01T00:02:00Z"}]
	}`)}
	var result SyncResult
	if err := doGet(context.Background(), d, "/api/download/sync", &result); err != nil {
		t.Fatal(err)
	}
	if len(result.Items) != 1 || !result.Items[0].WillAutoRetry || result.HasMore {
		t.Fatalf("unexpected sync: %+v", result)
	}
}

func TestTaskErrorRecoveryPrefersErrorCode(t *testing.T) {
	// Message deliberately does NOT look like a mover conflict; the
	// stable error_code must still drive the recovery CTA.
	got := taskErrorRecovery("PUT", "/api/download/cancel/42", 409, "something else", "task_in_mover_phase", nil)
	for _, want := range []string{"wait for the move to finish", "olares-cli knowledge download info 42"} {
		if !strings.Contains(got, want) {
			t.Fatalf("mover error_code recovery %q missing %q", got, want)
		}
	}

	got = taskErrorRecovery("PUT", "/api/download/pause/42", 400, "nope", "illegal_pause_status", nil)
	for _, want := range []string{"pause only works while downloading or waiting", "olares-cli knowledge download info 42"} {
		if !strings.Contains(got, want) {
			t.Fatalf("illegal pause error_code recovery %q missing %q", got, want)
		}
	}

	got = taskErrorRecovery("POST", "/api/download", 503, "yt-dlp unavailable: connection refused", "dependency_unavailable", nil)
	if !strings.Contains(got, ytdlpMarketInstall) {
		t.Fatalf("dependency error_code recovery %q missing install hint", got)
	}

	d := &fakeDoer{resp: []byte(`{"code":409,"message":"opaque","error_code":"task_in_mover_phase"}`)}
	err := doMutate(context.Background(), d, "PUT", "/api/download/resume/9", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	for _, want := range []string{"opaque", "wait for the move to finish", "olares-cli knowledge download info 9"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("doMutate error_code path %q missing %q", err, want)
		}
	}
}

func TestParseErrorBodyExtractsErrorCode(t *testing.T) {
	code, msg, errCode := parseErrorBody(500, `{"code":409,"message":"mover","error_code":"task_in_mover_phase"}`)
	if code != 409 || msg != "mover" || errCode != "task_in_mover_phase" {
		t.Fatalf("got code=%d msg=%q errCode=%q", code, msg, errCode)
	}
}

func TestTaskErrorRecoveryToleratesUnknownErrorCode(t *testing.T) {
	// Unknown codes must not panic or invent a CTA; message fallback
	// still applies when the status/path match legacy heuristics.
	got := taskErrorRecovery("PUT", "/api/download/resume/9", 409, "opaque future", "not_a_known_code", nil)
	for _, want := range []string{"wait for the move to finish", "olares-cli knowledge download info 9"} {
		if !strings.Contains(got, want) {
			t.Fatalf("unknown error_code should fall back to 409 heuristic, got %q missing %q", got, want)
		}
	}

	got = taskErrorRecovery("POST", "/api/download", 400, "bad url", "brand_new_code", nil)
	if !strings.Contains(got, "check --path/--url/--extra") {
		t.Fatalf("unknown error_code should fall back to message/status heuristic, got %q", got)
	}

	got = taskErrorRecovery("GET", "/api/download/list", 500, "boom", "brand_new_code", nil)
	if got != "" {
		t.Fatalf("unknown error_code with no matching fallback must be empty, got %q", got)
	}
}
