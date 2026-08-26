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
		// will_auto_retry says the sweep may pick the row up again; it
		// does not make the row any less failed right now.
		{status: "error", autoRet: true, want: "failure"},
	}
	for _, tc := range cases {
		task := DownloadTask{Status: tc.status, WillAutoRetry: tc.autoRet}
		if got := classifyWaitStatus(task); got != tc.want {
			t.Fatalf("classifyWaitStatus(%q, auto=%v)=%q want %q", tc.status, tc.autoRet, got, tc.want)
		}
	}
}

// TestWaitStopsOnAnAutoRetryingError pins the whole loop, not just the
// classifier: an error row ends wait on the first poll even when the
// server says it will retry. The later responses are there to fail the
// test loudly if the loop ever starts waiting the sweep out again.
func TestWaitStopsOnAnAutoRetryingError(t *testing.T) {
	d := &seqDoer{responses: [][]byte{
		mustEnvTaskRetry(t, 11, "error", true),
		mustEnvTaskRetry(t, 11, "downloading", false),
		mustEnvTask(t, 11, "completed"),
	}}
	pc := &preparedClient{doer: d}
	prev := waitPollInterval
	waitPollInterval = time.Millisecond
	t.Cleanup(func() { waitPollInterval = prev })

	task, err := waitForTerminal(context.Background(), pc, 11, 0)
	if err != nil {
		t.Fatalf("wait: %v", err)
	}
	if task.Status != "error" {
		t.Fatalf("status = %q, want error", task.Status)
	}
	if len(d.paths) != 1 {
		t.Fatalf("polled %d times, want 1", len(d.paths))
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

	last, err := waitForTerminal(context.Background(), pc, 12, 50*time.Millisecond)
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

	task, err := waitForTerminal(context.Background(), pc, 13, 0)
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

	task, err := waitForTerminal(context.Background(), pc, 7, 0)
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
	_, err = waitForTerminal(context.Background(), pc2, 8, 50*time.Millisecond)
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
	task, err := waitForTerminal(context.Background(), pc, 9, 0)
	if err != nil {
		t.Fatalf("waitForTerminal: %v", err)
	}
	if classifyWaitStatus(task) != "failure" {
		t.Fatalf("status %q should be failure", task.Status)
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
	raw, err := json.Marshal(map[string]interface{}{
		"code": 200,
		"data": map[string]interface{}{
			"id":              id,
			"status":          status,
			"will_auto_retry": willAutoRetry,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
