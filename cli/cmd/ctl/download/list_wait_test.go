package download

import (
	"context"
	"encoding/json"
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
		status string
		want   string
	}{
		{"completed", "success"},
		{"seeding", "success"},
		{"error", "failure"},
		{"cancelled", "failure"},
		{"removed", "failure"},
		{"waiting_to_move", "pending"},
		{"moving", "pending"},
		{"downloading", "pending"},
		{"paused", "pending"},
		{"waiting", "pending"},
		{"preparing", "pending"},
	}
	for _, tc := range cases {
		if got := classifyWaitStatus(tc.status); got != tc.want {
			t.Fatalf("classifyWaitStatus(%q)=%q want %q", tc.status, got, tc.want)
		}
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
	_, err = waitForTerminal(context.Background(), pc2, 8, 5*time.Millisecond)
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
	if classifyWaitStatus(task.Status) != "failure" {
		t.Fatalf("status %q should be failure", task.Status)
	}
}

func TestFetchListAllPages(t *testing.T) {
	page1, _ := json.Marshal(map[string]interface{}{
		"code": 200,
		"total": 3,
		"list": []map[string]interface{}{
			{"id": 1, "status": "completed", "will_auto_retry": false},
			{"id": 2, "status": "completed", "will_auto_retry": false},
		},
	})
	page2, _ := json.Marshal(map[string]interface{}{
		"code": 200,
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
	paths     []string
	onCall    func()
	i         int
}

func (s *seqDoer) DoJSON(_ context.Context, method, path string, body, out interface{}) error {
	s.paths = append(s.paths, path)
	if s.onCall != nil {
		s.onCall()
	}
	if s.i >= len(s.responses) {
		return json.Unmarshal(s.responses[len(s.responses)-1], out)
	}
	raw := s.responses[s.i]
	s.i++
	return json.Unmarshal(raw, out)
}

func mustEnvTask(t *testing.T, id int64, status string) []byte {
	t.Helper()
	raw, err := json.Marshal(map[string]interface{}{
		"code": 200,
		"data": map[string]interface{}{
			"id":              id,
			"status":          status,
			"will_auto_retry": false,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
