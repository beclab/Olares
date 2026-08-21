package download

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestValidateTaskStatus(t *testing.T) {
	for _, s := range []string{"", "downloading", "waiting_to_move", "moving", "completed", "seeding"} {
		if err := validateTaskStatus(s); err != nil {
			t.Fatalf("validateTaskStatus(%q): %v", s, err)
		}
	}
	err := validateTaskStatus("nope")
	if err == nil || !strings.Contains(err.Error(), "unsupported --status") {
		t.Fatalf("illegal status should fail locally, got %v", err)
	}
	if strings.Contains(strings.ToLower(err.Error()), "http") {
		t.Fatalf("local validation must not mention HTTP: %v", err)
	}
}

func TestClassifyWaitStatus(t *testing.T) {
	cases := []struct {
		status  string
		autoRet bool
		want    string
	}{
		{status: "completed", want: "success"},
		{status: "seeding", want: "success"},
		{status: "error", want: "failure"},
		{status: "cancelled", want: "failure"},
		{status: "removed", want: "failure"},
		{status: "waiting_to_move", want: "pending"},
		{status: "moving", want: "pending"},
		{status: "downloading", want: "pending"},
		{status: "paused", want: "pending"},
		{status: "waiting", want: "pending"},
		{status: "preparing", want: "pending"},
		// will_auto_retry is copy only; error is terminal from status.
		{status: "error", autoRet: true, want: "failure"},
	}
	for _, tc := range cases {
		task := DownloadTask{Status: tc.status, WillAutoRetry: tc.autoRet}
		if got := classifyWaitStatus(task); got != tc.want {
			t.Fatalf("classifyWaitStatus(%q, auto=%v)=%q want %q", tc.status, tc.autoRet, got, tc.want)
		}
	}
}

// TestWaitEndsOnAutoRetryingError pins that error is terminal even when
// the server will retry; wait must not keep polling toward completed.
func TestWaitEndsOnAutoRetryingError(t *testing.T) {
	d := &seqDoer{responses: [][]byte{
		mustEnvTaskRetry(t, 11, "error", true),
		mustEnvTask(t, 11, "completed"),
	}}
	pc := &preparedClient{doer: d}

	task, err := waitForTerminal(context.Background(), pc, 11, 0, nil)
	if err != nil {
		t.Fatalf("wait: %v", err)
	}
	if task.Status != "error" {
		t.Fatalf("status = %q, want error (must not wait for retry)", task.Status)
	}
	if classifyWaitStatus(task) != "failure" {
		t.Fatalf("auto-retry error should be failure, got %q", classifyWaitStatus(task))
	}
	if len(d.paths) != 1 {
		t.Fatalf("polls = %d, want 1 (stop on first error)", len(d.paths))
	}
}

// TestWaitTimeoutCarriesTaskID pins the recovery contract: a caller that
// times out still learns which row was left behind.
func TestWaitTimeoutCarriesTaskID(t *testing.T) {
	d := &seqDoer{responses: [][]byte{mustEnvTask(t, 12, "moving")}}
	pc := &preparedClient{doer: d}
	prev := waitPollInterval
	waitPollInterval = time.Millisecond
	t.Cleanup(func() { waitPollInterval = prev })

	last, err := waitForTerminal(context.Background(), pc, 12, 50*time.Millisecond, nil)
	if err == nil {
		t.Fatal("expected a timeout error")
	}
	if last.ID != 12 {
		t.Fatalf("last.ID = %d, want 12 so create --wait can still print it", last.ID)
	}
}

// TestWaitToleratesTransientPollErrors pins the retry budget: one blip
// must not abandon a download that is still running.
func TestWaitToleratesTransientPollErrors(t *testing.T) {
	d := &seqDoer{
		responses: [][]byte{nil, mustEnvTask(t, 13, "completed")},
		errs:      []error{errors.New("connection reset"), nil},
	}
	pc := &preparedClient{doer: d}
	prev := waitPollInterval
	waitPollInterval = time.Millisecond
	t.Cleanup(func() { waitPollInterval = prev })

	task, err := waitForTerminal(context.Background(), pc, 13, 0, nil)
	if err != nil {
		t.Fatalf("wait: %v", err)
	}
	if task.Status != "completed" {
		t.Fatalf("status = %q, want completed", task.Status)
	}
}

func TestRunListRejectsAllAppsWithApp(t *testing.T) {
	err := runList(context.Background(), nil, "larepass", "", 0, 0, false, true, true, "table")
	if err == nil || !strings.Contains(err.Error(), "--all-apps cannot be combined with --app") {
		t.Fatalf("expected mutual exclusion error, got %v", err)
	}
}

func TestRunListRejectsIllegalStatusBeforeRequest(t *testing.T) {
	err := runList(context.Background(), nil, "wise", "bogus", 0, 0, false, false, false, "table")
	if err == nil || !strings.Contains(err.Error(), "unsupported --status") {
		t.Fatalf("expected local status error, got %v", err)
	}
}

func TestWaitForTerminalSuccessAndTimeout(t *testing.T) {
	// success after mover phase
	var calls int
	d := &seqDoer{responses: [][]byte{
		mustEnvTask(t, 7, "waiting_to_move"),
		mustEnvTask(t, 7, "moving"),
		mustEnvTask(t, 7, "completed"),
	}, onCall: func() { calls++ }}
	pc := &preparedClient{doer: d}
	prev := waitPollInterval
	waitPollInterval = time.Millisecond
	t.Cleanup(func() { waitPollInterval = prev })

	task, err := waitForTerminal(context.Background(), pc, 7, 0, nil)
	if err != nil {
		t.Fatalf("wait: %v", err)
	}
	if task.Status != "completed" || calls != 3 {
		t.Fatalf("got status=%s calls=%d", task.Status, calls)
	}

	// timeout while still moving
	d2 := &seqDoer{responses: [][]byte{
		mustEnvTask(t, 8, "moving"),
		mustEnvTask(t, 8, "moving"),
		mustEnvTask(t, 8, "moving"),
	}}
	pc2 := &preparedClient{doer: d2}
	_, err = waitForTerminal(context.Background(), pc2, 8, 50*time.Millisecond, nil)
	if err == nil || !strings.Contains(err.Error(), "timed out") || !strings.Contains(err.Error(), "moving") {
		t.Fatalf("expected timeout with status, got %v", err)
	}
}

// TestWaitTerminalFailureSurfacesAsError pins the script contract:
// wait / create --wait must exit non-zero on error/cancelled/removed
// (JSON output still prints the row, then returns an error).
func TestWaitTerminalFailureSurfacesAsError(t *testing.T) {
	d := &seqDoer{responses: [][]byte{mustEnvTask(t, 9, "cancelled")}}
	pc := &preparedClient{doer: d}
	task, err := waitForTerminal(context.Background(), pc, 9, 0, nil)
	if err != nil {
		t.Fatalf("waitForTerminal: %v", err)
	}
	if classifyWaitStatus(task) != "failure" {
		t.Fatalf("status %q should be failure", task.Status)
	}
}

func TestWaitRetryHint(t *testing.T) {
	auto := waitRetryHint(DownloadTask{ID: 42, Status: "error", WillAutoRetry: true})
	if !strings.Contains(auto, "server will auto-retry") || !strings.Contains(auto, "info 42") {
		t.Fatalf("auto-retry hint = %q", auto)
	}
	final := waitRetryHint(DownloadTask{ID: 42, Status: "error", WillAutoRetry: false})
	if !strings.Contains(final, "will_auto_retry=false") {
		t.Fatalf("final-error hint = %q", final)
	}
	if got := waitRetryHint(DownloadTask{ID: 42, Status: "cancelled"}); got != "" {
		t.Fatalf("cancelled should not get a retry hint, got %q", got)
	}
}

func TestWaitFailureErrorIncludesErrMsg(t *testing.T) {
	err := waitFailureError(DownloadTask{ID: 9, Status: "error", ErrMsg: "connection refused"})
	if err == nil || !strings.Contains(err.Error(), `status "error"`) || !strings.Contains(err.Error(), "connection refused") {
		t.Fatalf("got %v", err)
	}
	plain := waitFailureError(DownloadTask{ID: 9, Status: "cancelled"})
	if plain == nil || !strings.Contains(plain.Error(), `status "cancelled"`) || strings.Contains(plain.Error(), ":") {
		t.Fatalf("cancelled without err_msg should stay a short error, got %v", plain)
	}
}

func TestEmitWaitFailureDetailsJSONOnly(t *testing.T) {
	task := DownloadTask{
		ID:            9,
		Status:        "error",
		ErrMsg:        "403 Forbidden",
		ErrCategory:   "authorization_failed",
		WillAutoRetry: true,
	}
	var table strings.Builder
	emitWaitFailureDetails(&table, task, FormatTable)
	if table.Len() != 0 {
		t.Fatalf("table mode must not use the JSON hint helper: %q", table.String())
	}

	var js strings.Builder
	emitWaitFailureDetails(&js, task, FormatJSON)
	jsonOut := js.String()
	if strings.Contains(jsonOut, "Error:") {
		t.Fatalf("json mode must not reprint err_msg on stderr: %q", jsonOut)
	}
	if !strings.Contains(jsonOut, "server will auto-retry") {
		t.Fatalf("json mode still needs the retry hint: %q", jsonOut)
	}
}

func TestEmitWaitOutcomeSuccessAndFailure(t *testing.T) {
	ok := DownloadTask{ID: 6, Status: "completed", DownloadProvider: "yt-dlp", FileName: "clip.mp4", Path: "Home/Downloads"}
	var success strings.Builder
	emitWaitOutcome(&success, ok, "success")
	got := success.String()
	for _, want := range []string{"download '6': completed (status=completed)", "provider: yt-dlp", "name: clip.mp4", "path: Home/Downloads"} {
		if !strings.Contains(got, want) {
			t.Fatalf("success outcome missing %q in %q", want, got)
		}
	}

	fail := DownloadTask{ID: 6, Status: "error", ErrMsg: "http error 412", ErrCategory: "network_error", WillAutoRetry: true, FileName: "clip.mp4"}
	var failed strings.Builder
	emitWaitOutcome(&failed, fail, "failure")
	out := failed.String()
	for _, want := range []string{"download '6' failed: http error 412", "err_category: network_error", "server will auto-retry"} {
		if !strings.Contains(out, want) {
			t.Fatalf("failure outcome missing %q in %q", want, out)
		}
	}
}

func TestWaitProgressPrintsOnChangeNotEveryPoll(t *testing.T) {
	d := &seqDoer{responses: [][]byte{
		mustEnvTaskFields(t, map[string]interface{}{"id": 7, "status": "downloading", "percent": 10.0, "download_provider": "yt-dlp", "file_name": "clip.mp4", "will_auto_retry": false}),
		mustEnvTaskFields(t, map[string]interface{}{"id": 7, "status": "downloading", "percent": 10.2, "download_provider": "yt-dlp", "file_name": "clip.mp4", "will_auto_retry": false}),
		mustEnvTaskFields(t, map[string]interface{}{"id": 7, "status": "downloading", "percent": 16.0, "download_provider": "yt-dlp", "file_name": "clip.mp4", "will_auto_retry": false}),
		mustEnvTaskFields(t, map[string]interface{}{"id": 7, "status": "completed", "percent": 100.0, "download_provider": "yt-dlp", "file_name": "clip.mp4", "will_auto_retry": false}),
	}}
	pc := &preparedClient{doer: d}
	prev := waitPollInterval
	waitPollInterval = time.Millisecond
	t.Cleanup(func() { waitPollInterval = prev })

	var buf strings.Builder
	task, err := waitForTerminal(context.Background(), pc, 7, 0, &buf)
	if err != nil {
		t.Fatalf("wait: %v", err)
	}
	if task.Status != "completed" {
		t.Fatalf("status = %q", task.Status)
	}
	out := buf.String()
	if strings.Contains(out, "provider=") || strings.Contains(out, "name=") || strings.Contains(out, "status=") {
		t.Fatalf("ticks must stay compact, got %q", out)
	}
	if !strings.Contains(out, "[7] downloading 10.0%") {
		t.Fatalf("first tick missing: %q", out)
	}
	if strings.Contains(out, "10.2%") {
		t.Fatalf("sub-5%% change must not reprint: %q", out)
	}
	if !strings.Contains(out, "[7] downloading 16.0%") {
		t.Fatalf("5%% jump should reprint: %q", out)
	}
	if strings.Contains(out, "completed") {
		t.Fatalf("terminal status belongs in the summary, not ticks: %q", out)
	}
}

func TestWaitProgressSkipsInitialWaitingZero(t *testing.T) {
	d := &seqDoer{responses: [][]byte{
		mustEnvTaskFields(t, map[string]interface{}{"id": 8, "status": "waiting", "percent": 0.0, "will_auto_retry": false}),
		mustEnvTaskFields(t, map[string]interface{}{"id": 8, "status": "preparing", "percent": 0.0, "will_auto_retry": false}),
		mustEnvTaskFields(t, map[string]interface{}{"id": 8, "status": "completed", "percent": 100.0, "file_name": "clip.mp4", "will_auto_retry": false}),
	}}
	pc := &preparedClient{doer: d}
	prev := waitPollInterval
	waitPollInterval = time.Millisecond
	t.Cleanup(func() { waitPollInterval = prev })

	var buf strings.Builder
	if _, err := waitForTerminal(context.Background(), pc, 8, 0, &buf); err != nil {
		t.Fatalf("wait: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "waiting") {
		t.Fatalf("initial waiting 0%% is covered by the header, got %q", out)
	}
	if !strings.Contains(out, "[8] preparing 0.0%") {
		t.Fatalf("preparing should still print: %q", out)
	}
}

func TestWaitProgressHeartbeatWhenStalled(t *testing.T) {
	d := &seqDoer{responses: [][]byte{
		mustEnvTaskFields(t, map[string]interface{}{"id": 9, "status": "downloading", "percent": 50.5, "will_auto_retry": false}),
		mustEnvTaskFields(t, map[string]interface{}{"id": 9, "status": "downloading", "percent": 50.5, "will_auto_retry": false}),
		mustEnvTaskFields(t, map[string]interface{}{"id": 9, "status": "downloading", "percent": 50.5, "will_auto_retry": false}),
		mustEnvTaskFields(t, map[string]interface{}{"id": 9, "status": "completed", "percent": 100.0, "will_auto_retry": false}),
	}}
	pc := &preparedClient{doer: d}
	prevInterval := waitPollInterval
	prevBeat := waitProgressHeartbeat
	waitPollInterval = time.Millisecond
	waitProgressHeartbeat = time.Millisecond
	t.Cleanup(func() {
		waitPollInterval = prevInterval
		waitProgressHeartbeat = prevBeat
	})

	var buf strings.Builder
	if _, err := waitForTerminal(context.Background(), pc, 9, 0, &buf); err != nil {
		t.Fatalf("wait: %v", err)
	}
	if strings.Count(buf.String(), "[9] downloading 50.5%") < 2 {
		t.Fatalf("stalled polls should heartbeat, got %q", buf.String())
	}
}

func TestRenderInfoIncludesWillAutoRetry(t *testing.T) {
	var buf strings.Builder
	if err := renderInfo(&buf, DownloadTask{ID: 1, Status: "error", WillAutoRetry: true}); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "WillAutoRetry") || !strings.Contains(got, "true") {
		t.Fatalf("info table should show WillAutoRetry, got %q", got)
	}
}

func TestFetchListAllPages(t *testing.T) {
	page1, _ := json.Marshal(map[string]interface{}{
		"code":  200,
		"total": 3,
		"list": []map[string]interface{}{
			{"id": 1, "status": "completed", "will_auto_retry": false},
			{"id": 2, "status": "completed", "will_auto_retry": false},
		},
	})
	page2, _ := json.Marshal(map[string]interface{}{
		"code":  200,
		"total": 3,
		"list": []map[string]interface{}{
			{"id": 3, "status": "completed", "will_auto_retry": false},
		},
	})
	d := &seqDoer{responses: [][]byte{page1, page2}}
	pc := &preparedClient{doer: d}
	res, err := fetchListAll(context.Background(), pc, "wise", "", 2)
	if err != nil {
		t.Fatalf("fetchListAll: %v", err)
	}
	if len(res.List) != 3 || res.Total != 3 {
		t.Fatalf("got %d of %d", len(res.List), res.Total)
	}
	if !strings.Contains(d.paths[0], "page=1") || !strings.Contains(d.paths[1], "page=2") {
		t.Fatalf("paths = %v", d.paths)
	}
}

type seqDoer struct {
	responses [][]byte
	// errs, when set, is indexed alongside responses; a non-nil entry
	// fails that call instead of decoding it.
	errs   []error
	paths  []string
	onCall func()
	i      int
}

func (s *seqDoer) DoJSON(_ context.Context, method, path string, body, out interface{}) error {
	s.paths = append(s.paths, path)
	if s.onCall != nil {
		s.onCall()
	}
	if s.i >= len(s.responses) {
		return json.Unmarshal(s.responses[len(s.responses)-1], out)
	}
	idx := s.i
	s.i++
	if idx < len(s.errs) && s.errs[idx] != nil {
		return s.errs[idx]
	}
	return json.Unmarshal(s.responses[idx], out)
}

func mustEnvTask(t *testing.T, id int64, status string) []byte {
	t.Helper()
	return mustEnvTaskRetry(t, id, status, false)
}

func mustEnvTaskRetry(t *testing.T, id int64, status string, willAutoRetry bool) []byte {
	t.Helper()
	return mustEnvTaskFields(t, map[string]interface{}{
		"id":              id,
		"status":          status,
		"will_auto_retry": willAutoRetry,
	})
}

func mustEnvTaskFields(t *testing.T, fields map[string]interface{}) []byte {
	t.Helper()
	raw, err := json.Marshal(map[string]interface{}{
		"code": 200,
		"data": fields,
	})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
