package users

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

// fakeStatusDoer satisfies the Doer interface that decodeObjectResult uses.
// It drains `responses` one entry per call; once exhausted it returns the
// last entry forever (so tests for "stable terminal state" don't have to
// pad the slice).
type fakeStatusDoer struct {
	calls     int
	responses []statusResponse
}

type statusResponse struct {
	body string
	err  error
}

func (f *fakeStatusDoer) DoJSON(ctx context.Context, method, path string, body, out interface{}) error {
	defer func() { f.calls++ }()
	idx := f.calls
	if idx >= len(f.responses) {
		idx = len(f.responses) - 1
	}
	r := f.responses[idx]
	if r.err != nil {
		return r.err
	}
	if raw, ok := out.(*json.RawMessage); ok {
		*raw = json.RawMessage(r.body)
		return nil
	}
	return json.Unmarshal([]byte(r.body), out)
}

func statusBody(t *testing.T, st accountModifyStatus) string {
	t.Helper()
	envelope := map[string]any{"code": 200, "data": st}
	raw, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal status: %v", err)
	}
	return string(raw)
}

func TestWaitForUserState_DeleteSuccess(t *testing.T) {
	d := &fakeStatusDoer{responses: []statusResponse{
		{body: statusBody(t, accountModifyStatus{Name: "bob", Status: "Deleting"})},
		{body: statusBody(t, accountModifyStatus{Name: "bob", Status: "Deleted"})},
	}}
	st, err := waitForUserState(context.Background(), d, userWatchOptions{
		Timeout:  2 * time.Second,
		Interval: 5 * time.Millisecond,
	}, newUserWatchTarget(userWatchDelete, "bob"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st == nil || strings.TrimSpace(st.Status) != "Deleted" {
		t.Fatalf("expected Deleted, got %+v", st)
	}
}

func TestWaitForUserState_DeleteAbsentMeansSuccess(t *testing.T) {
	d := &fakeStatusDoer{responses: []statusResponse{
		{err: fmt.Errorf("backend error: HTTP 404 not found")},
	}}
	st, err := waitForUserState(context.Background(), d, userWatchOptions{
		Timeout:  2 * time.Second,
		Interval: 5 * time.Millisecond,
	}, newUserWatchTarget(userWatchDelete, "ghost"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st == nil || st.Status != "Deleted" || st.Name != "ghost" {
		t.Fatalf("expected synthesized Deleted/ghost, got %+v", st)
	}
}

func TestWaitForUserState_CreateFailureRaisesFailureError(t *testing.T) {
	d := &fakeStatusDoer{responses: []statusResponse{
		{body: statusBody(t, accountModifyStatus{Name: "alice", Status: "Creating"})},
		{body: statusBody(t, accountModifyStatus{Name: "alice", Status: "Failed", Message: "ldap sync error"})},
	}}
	_, err := waitForUserState(context.Background(), d, userWatchOptions{
		Timeout:  2 * time.Second,
		Interval: 5 * time.Millisecond,
	}, newUserWatchTarget(userWatchCreate, "alice"))
	if err == nil {
		t.Fatal("expected failure error")
	}
	var fail *userWatchFailureError
	if !errors.As(err, &fail) {
		t.Fatalf("expected userWatchFailureError, got %T: %v", err, err)
	}
	if fail.status.Status != "Failed" {
		t.Fatalf("expected status=Failed in failure, got %+v", fail.status)
	}
	if !strings.Contains(fail.Error(), "ldap sync error") {
		t.Fatalf("expected message to surface reason; got %q", fail.Error())
	}
}

func TestWaitForUserState_CreateDisappearedTreatedAsFailure(t *testing.T) {
	// "Deleted" while waiting for Created means the row vanished mid-watch.
	d := &fakeStatusDoer{responses: []statusResponse{
		{body: statusBody(t, accountModifyStatus{Name: "alice", Status: "Deleted"})},
	}}
	_, err := waitForUserState(context.Background(), d, userWatchOptions{
		Timeout:  2 * time.Second,
		Interval: 5 * time.Millisecond,
	}, newUserWatchTarget(userWatchCreate, "alice"))
	if err == nil {
		t.Fatal("expected failure")
	}
	var fail *userWatchFailureError
	if !errors.As(err, &fail) {
		t.Fatalf("expected userWatchFailureError, got %T: %v", err, err)
	}
	if fail.status.Status != "Deleted" {
		t.Fatalf("expected status=Deleted in failure, got %+v", fail.status)
	}
}

func TestWaitForUserState_ConsecutiveErrorsAbort(t *testing.T) {
	// 5 transient transport errors (not 404, not auth) → aborted.
	d := &fakeStatusDoer{responses: []statusResponse{
		{err: fmt.Errorf("backend error: HTTP 502 bad gateway")},
	}}
	_, err := waitForUserState(context.Background(), d, userWatchOptions{
		Timeout:  10 * time.Second,
		Interval: 1 * time.Millisecond,
	}, newUserWatchTarget(userWatchCreate, "alice"))
	if err == nil {
		t.Fatal("expected aborted error")
	}
	if !strings.Contains(err.Error(), "watch aborted after") {
		t.Fatalf("expected 'watch aborted after' in %q", err.Error())
	}
	if d.calls < 5 {
		t.Fatalf("expected at least 5 attempts before abort, got %d", d.calls)
	}
}

func TestWaitForUserState_TimeoutCarriesLastStatus(t *testing.T) {
	d := &fakeStatusDoer{responses: []statusResponse{
		{body: statusBody(t, accountModifyStatus{Name: "bob", Status: "Deleting"})},
	}}
	_, err := waitForUserState(context.Background(), d, userWatchOptions{
		Timeout:  20 * time.Millisecond,
		Interval: 5 * time.Millisecond,
	}, newUserWatchTarget(userWatchDelete, "bob"))
	if err == nil {
		t.Fatal("expected timeout error")
	}
	var to *userWatchTimeoutError
	if !errors.As(err, &to) {
		t.Fatalf("expected userWatchTimeoutError, got %T: %v", err, err)
	}
	if to.last == nil || to.last.Status != "Deleting" {
		t.Fatalf("expected last status=Deleting, got %+v", to.last)
	}
	if !strings.Contains(to.Error(), "Deleting") {
		t.Fatalf("expected last status to surface in message, got %q", to.Error())
	}
}

func TestWaitForUserState_AuthErrorShortCircuits(t *testing.T) {
	d := &fakeStatusDoer{responses: []statusResponse{
		{err: fmt.Errorf("server rejected the access token (HTTP 401)")},
	}}
	_, err := waitForUserState(context.Background(), d, userWatchOptions{
		Timeout:  2 * time.Second,
		Interval: 5 * time.Millisecond,
	}, newUserWatchTarget(userWatchDelete, "bob"))
	if err == nil {
		t.Fatal("expected auth error to surface")
	}
	if d.calls != 1 {
		t.Fatalf("auth error should bail on first attempt, got %d calls", d.calls)
	}
	if !strings.Contains(err.Error(), "401") {
		t.Fatalf("expected 401 in error, got %q", err.Error())
	}
}

func TestNewUserWatchTarget_OpSets(t *testing.T) {
	create := newUserWatchTarget(userWatchCreate, "x")
	if !create.successSet["Created"] {
		t.Fatal("create.successSet missing Created")
	}
	if !create.failureSet["Failed"] || !create.failureSet["Deleted"] {
		t.Fatal("create.failureSet must include Failed and Deleted")
	}
	if create.absentMeansSuccess {
		t.Fatal("create.absentMeansSuccess should be false")
	}

	del := newUserWatchTarget(userWatchDelete, "x")
	if !del.successSet["Deleted"] {
		t.Fatal("delete.successSet missing Deleted")
	}
	if len(del.failureSet) != 0 {
		t.Fatal("delete should have no terminal failure states (timeout only)")
	}
	if !del.absentMeansSuccess {
		t.Fatal("delete.absentMeansSuccess should be true")
	}
}
