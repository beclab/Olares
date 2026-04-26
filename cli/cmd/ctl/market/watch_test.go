package market

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/beclab/Olares/cli/pkg/credential"
)

// classifyForTest mirrors the classification waitForTerminal performs on each
// poll: it returns "success" / "failure" / "progressing" for a hypothetical
// row, without invoking the actual poll loop. Keeping the helper local to the
// test file avoids growing the package's exported surface just for unit
// coverage.
func classifyForTest(t watchTarget, row statusRow) string {
	if !t.matchesOpType(row) {
		return "progressing"
	}
	switch {
	case t.successSet[row.State]:
		return "success"
	case t.failureSet[row.State]:
		return "failure"
	default:
		return "progressing"
	}
}

func TestClassifierInstallLifecycle(t *testing.T) {
	target := newWatchTarget(watchInstall, "myapp", "market.olares")

	cases := []struct {
		name   string
		row    statusRow
		expect string
	}{
		{"pending", statusRow{State: "pending", OpType: "install"}, "progressing"},
		{"downloading", statusRow{State: "downloading", OpType: "install"}, "progressing"},
		{"installing", statusRow{State: "installing", OpType: "install"}, "progressing"},
		{"initializing", statusRow{State: "initializing", OpType: "install"}, "progressing"},
		{"running with install op", statusRow{State: "running", OpType: "install"}, "success"},
		{"installFailed", statusRow{State: "installFailed", OpType: "install"}, "failure"},
		{"downloadFailed", statusRow{State: "downloadFailed", OpType: "install"}, "failure"},
		// Cancel during install is also terminal-failure for the install
		// CTA: from the user's perspective the install they asked for did
		// not happen.
		{"installingCanceled", statusRow{State: "installingCanceled", OpType: "install"}, "failure"},
		// Stale OpType from a prior lifecycle must not prematurely
		// classify any state.
		{"running with stale upgrade op", statusRow{State: "running", OpType: "upgrade"}, "progressing"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := classifyForTest(target, c.row); got != c.expect {
				t.Fatalf("classify(%+v) = %s, want %s", c.row, got, c.expect)
			}
		})
	}
}

func TestClassifierUpgradeWaitsForOpTypeFlip(t *testing.T) {
	// Issuing `upgrade` on a row currently `running, op=install` must NOT
	// short-circuit to success on tick zero; only after the backend
	// flips OpType to `upgrade` (plus reaches `running` again) is the
	// upgrade complete.
	target := newWatchTarget(watchUpgrade, "myapp", "market.olares")

	stale := statusRow{State: "running", OpType: "install"}
	if got := classifyForTest(target, stale); got != "progressing" {
		t.Fatalf("stale install OpType should keep progressing, got %s", got)
	}

	mid := statusRow{State: "upgrading", OpType: "upgrade"}
	if got := classifyForTest(target, mid); got != "progressing" {
		t.Fatalf("upgrading should be progressing, got %s", got)
	}

	done := statusRow{State: "running", OpType: "upgrade"}
	if got := classifyForTest(target, done); got != "success" {
		t.Fatalf("running with upgrade op should be success, got %s", got)
	}

	failed := statusRow{State: "upgradeFailed", OpType: "upgrade"}
	if got := classifyForTest(target, failed); got != "failure" {
		t.Fatalf("upgradeFailed should be failure, got %s", got)
	}
}

func TestClassifierUninstall(t *testing.T) {
	target := newWatchTarget(watchUninstall, "myapp", "market.olares")
	if !target.absentMeansSuccess {
		t.Fatalf("uninstall target must set absentMeansSuccess")
	}
	if got := classifyForTest(target, statusRow{State: "uninstalling", OpType: "uninstall"}); got != "progressing" {
		t.Fatalf("uninstalling should be progressing, got %s", got)
	}
	if got := classifyForTest(target, statusRow{State: "uninstalled", OpType: "uninstall"}); got != "success" {
		t.Fatalf("uninstalled should be success, got %s", got)
	}
	if got := classifyForTest(target, statusRow{State: "uninstallFailed", OpType: "uninstall"}); got != "failure" {
		t.Fatalf("uninstallFailed should be failure, got %s", got)
	}
}

func TestClassifierCancelIgnoresOpType(t *testing.T) {
	// cancel's terminal row keeps the *underlying* op (install /
	// upgrade / ...), so matchOpType must be false; otherwise we'd never
	// classify the canceled state as terminal.
	target := newWatchTarget(watchCancel, "myapp", "")
	if target.matchOpType {
		t.Fatalf("cancel target must NOT require OpType match")
	}

	row := statusRow{State: "installingCanceled", OpType: "install"}
	if got := classifyForTest(target, row); got != "success" {
		t.Fatalf("installingCanceled under cancel target should be success, got %s", got)
	}

	failed := statusRow{State: "installingCancelFailed", OpType: "install"}
	if got := classifyForTest(target, failed); got != "failure" {
		t.Fatalf("installingCancelFailed should be failure, got %s", got)
	}
}

func TestClassifierStatusOpAgnostic(t *testing.T) {
	target := newWatchTarget(watchStatus, "myapp", "market.olares")
	if target.matchOpType {
		t.Fatalf("status target must not require OpType match")
	}
	if !target.absentMeansSuccess {
		t.Fatalf("status target must opt into absentMeansSuccess so a row vanishing mid-watch is terminal")
	}

	cases := []struct {
		name   string
		row    statusRow
		expect string
	}{
		// Stable resting states for any lifecycle → success.
		{"running after install", statusRow{State: "running", OpType: "install"}, "success"},
		{"running after upgrade", statusRow{State: "running", OpType: "upgrade"}, "success"},
		{"running after resume", statusRow{State: "running", OpType: "resume"}, "success"},
		{"stopped", statusRow{State: "stopped", OpType: "stop"}, "success"},
		{"uninstalled", statusRow{State: "uninstalled", OpType: "uninstall"}, "success"},
		{"installingCanceled", statusRow{State: "installingCanceled", OpType: "install"}, "success"},
		// Any in-flight state keeps polling.
		{"pending", statusRow{State: "pending", OpType: "install"}, "progressing"},
		{"installing", statusRow{State: "installing", OpType: "install"}, "progressing"},
		{"upgrading", statusRow{State: "upgrading", OpType: "upgrade"}, "progressing"},
		{"stopping", statusRow{State: "stopping", OpType: "stop"}, "progressing"},
		// All declared failure states → failure regardless of OpType.
		{"installFailed", statusRow{State: "installFailed", OpType: "install"}, "failure"},
		{"upgradeFailed", statusRow{State: "upgradeFailed", OpType: "upgrade"}, "failure"},
		{"stopFailed", statusRow{State: "stopFailed", OpType: "stop"}, "failure"},
		{"installingCancelFailed", statusRow{State: "installingCancelFailed", OpType: "install"}, "failure"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := classifyForTest(target, c.row); got != c.expect {
				t.Fatalf("classify(%+v) = %s, want %s", c.row, got, c.expect)
			}
		})
	}
}

func TestWaitForTerminalStatusReachesRunning(t *testing.T) {
	// status --watch is fired against an app that's mid-install; the
	// watcher must recognize `running` as terminal even though no
	// specific op was specified by the caller.
	seq := []statusRow{
		{State: "installing", OpType: "install"},
		{State: "initializing", OpType: "install"},
		{State: "running", OpType: "install"},
	}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(5*time.Second, 5*time.Millisecond)

	row, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchStatus, "myapp", "market.olares"))
	if err != nil {
		t.Fatalf("expected status watch success, got %v", err)
	}
	if row.State != "running" {
		t.Fatalf("expected terminal running, got %s", row.State)
	}
}

func TestWaitForTerminalStatusSurfacesFailure(t *testing.T) {
	seq := []statusRow{
		{State: "downloading", OpType: "install"},
		{State: "installFailed", OpType: "install", Message: "image pull error"},
	}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(5*time.Second, 5*time.Millisecond)

	_, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchStatus, "myapp", "market.olares"))
	if err == nil {
		t.Fatalf("expected installFailed surfaced as failure")
	}
	var fail *watchFailureError
	if !errors.As(err, &fail) {
		t.Fatalf("expected watchFailureError, got %T: %v", err, err)
	}
	if fail.row.State != "installFailed" {
		t.Fatalf("expected installFailed, got %s", fail.row.State)
	}
}

func TestClassifierStopResume(t *testing.T) {
	stopT := newWatchTarget(watchStop, "myapp", "")
	if got := classifyForTest(stopT, statusRow{State: "stopping", OpType: "stop"}); got != "progressing" {
		t.Fatalf("stopping should be progressing, got %s", got)
	}
	if got := classifyForTest(stopT, statusRow{State: "stopped", OpType: "stop"}); got != "success" {
		t.Fatalf("stopped should be success, got %s", got)
	}
	if got := classifyForTest(stopT, statusRow{State: "stopFailed", OpType: "stop"}); got != "failure" {
		t.Fatalf("stopFailed should be failure, got %s", got)
	}

	resumeT := newWatchTarget(watchResume, "myapp", "")
	if got := classifyForTest(resumeT, statusRow{State: "running", OpType: "resume"}); got != "success" {
		t.Fatalf("running under resume target should be success, got %s", got)
	}
	if got := classifyForTest(resumeT, statusRow{State: "resumeFailed", OpType: "resume"}); got != "failure" {
		t.Fatalf("resumeFailed should be failure, got %s", got)
	}
}

// fakeStateServer serves /app-store/api/v2/market/state with a configurable
// queue of states so we can drive waitForTerminal end-to-end without a real
// cluster. It models exactly the response shape parseStatusRows expects.
type fakeStateServer struct {
	mu       sync.Mutex
	idx      int32
	app      string
	source   string
	sequence []statusRow
	missing  bool
	srv      *httptest.Server
}

func newFakeStateServer(t *testing.T, app, source string, seq []statusRow) *fakeStateServer {
	t.Helper()
	f := &fakeStateServer{app: app, source: source, sequence: seq}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/market/state") {
			http.NotFound(w, r)
			return
		}
		row := f.next()
		body := f.envelope(row)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeStateServer) next() (row statusRow) {
	i := atomic.AddInt32(&f.idx, 1) - 1
	if i >= int32(len(f.sequence)) {
		// Stay on the last state forever once the queue is exhausted —
		// makes timeout assertions deterministic.
		return f.sequence[len(f.sequence)-1]
	}
	return f.sequence[i]
}

func (f *fakeStateServer) markMissing() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.missing = true
}

func (f *fakeStateServer) envelope(row statusRow) []byte {
	f.mu.Lock()
	missing := f.missing
	f.mu.Unlock()

	apps := []map[string]interface{}{}
	if !missing {
		apps = append(apps, map[string]interface{}{
			"status": map[string]interface{}{
				"name":     f.app,
				"state":    row.State,
				"opType":   row.OpType,
				"progress": row.Progress,
				"message":  row.Message,
			},
		})
	}
	// Mirror the real /market/state shape parseStatusRows expects:
	// the v2 envelope unmarshals resp.Data into MarketStateResponse,
	// whose `user_data.sources[<source>].app_state_latest[].status`
	// path holds the per-app records.
	envelope := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"user_data": map[string]interface{}{
				"sources": map[string]interface{}{
					f.source: map[string]interface{}{
						"type":             "market",
						"app_state_latest": apps,
					},
				},
			},
		},
	}
	b, _ := json.Marshal(envelope)
	return b
}

func newTestMarketClient(t *testing.T, baseURL string) *MarketClient {
	t.Helper()
	rp := &credential.ResolvedProfile{
		Name:        "test",
		OlaresID:    "tester@olares.test",
		AccessToken: "test-token",
		MarketURL:   baseURL,
	}
	return NewMarketClient(http.DefaultClient, http.DefaultClient, rp, "market.olares")
}

// drain swallows any output runWithWatch / waitForTerminal would emit so
// `go test` output isn't polluted; we still inspect the OperationResult /
// error returned by the API.
func quietOpts(timeout, interval time.Duration) *MarketOptions {
	return &MarketOptions{
		Source:        "market.olares",
		Output:        "json", // suppresses opts.info
		Quiet:         true,
		Watch:         true,
		WatchTimeout:  timeout,
		WatchInterval: interval,
	}
}

func TestWaitForTerminalInstallSuccess(t *testing.T) {
	seq := []statusRow{
		{State: "pending", OpType: "install"},
		{State: "downloading", OpType: "install"},
		{State: "installing", OpType: "install"},
		{State: "running", OpType: "install"},
	}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(5*time.Second, 5*time.Millisecond)

	row, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchInstall, "myapp", "market.olares"))
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if row.State != "running" {
		t.Fatalf("expected terminal state running, got %s", row.State)
	}
}

func TestWaitForTerminalInstallFailure(t *testing.T) {
	seq := []statusRow{
		{State: "pending", OpType: "install"},
		{State: "installFailed", OpType: "install", Message: "image pull error"},
	}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(5*time.Second, 5*time.Millisecond)

	_, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchInstall, "myapp", "market.olares"))
	if err == nil {
		t.Fatalf("expected failure, got nil")
	}
	var fail *watchFailureError
	if !errors.As(err, &fail) {
		t.Fatalf("expected watchFailureError, got %T: %v", err, err)
	}
	if fail.row.State != "installFailed" {
		t.Fatalf("expected installFailed in error row, got %s", fail.row.State)
	}
	if !strings.Contains(err.Error(), "installFailed") {
		t.Fatalf("error message should mention installFailed, got %q", err.Error())
	}
}

func TestWaitForTerminalUninstallAbsent(t *testing.T) {
	seq := []statusRow{
		{State: "uninstalling", OpType: "uninstall"},
	}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(5*time.Second, 5*time.Millisecond)

	// After the first poll, drop the row entirely -> simulates the
	// backend having finished cleanup.
	go func() {
		time.Sleep(20 * time.Millisecond)
		srv.markMissing()
	}()

	row, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchUninstall, "myapp", "market.olares"))
	if err != nil {
		t.Fatalf("expected success on row absence, got %v", err)
	}
	if row.State != "uninstalled" {
		t.Fatalf("expected synthesized uninstalled row, got %s", row.State)
	}
}

func TestWaitForTerminalUpgradeWaitsForOpTypeFlip(t *testing.T) {
	// Tick 0 sees the legacy `running, op=install` row from the previous
	// install; only after the backend flips to op=upgrade and reaches
	// running again should we declare success.
	seq := []statusRow{
		{State: "running", OpType: "install"},
		{State: "upgrading", OpType: "upgrade"},
		{State: "running", OpType: "upgrade"},
	}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(5*time.Second, 5*time.Millisecond)

	row, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchUpgrade, "myapp", "market.olares"))
	if err != nil {
		t.Fatalf("expected success once upgrade lifecycle completes, got %v", err)
	}
	if row.OpType != "upgrade" || row.State != "running" {
		t.Fatalf("expected running/upgrade, got %s/%s", row.State, row.OpType)
	}
}

func TestWaitForTerminalCancelLifecycle(t *testing.T) {
	seq := []statusRow{
		{State: "installing", OpType: "install"},
		{State: "installingCanceling", OpType: "install"},
		{State: "installingCanceled", OpType: "install"},
	}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(5*time.Second, 5*time.Millisecond)

	row, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchCancel, "myapp", "market.olares"))
	if err != nil {
		t.Fatalf("expected cancel success, got %v", err)
	}
	if row.State != "installingCanceled" {
		t.Fatalf("expected installingCanceled, got %s", row.State)
	}
}

func TestWaitForTerminalTimeoutSurfacesLastState(t *testing.T) {
	// Stuck in `installing` forever: classifier never reaches a terminal
	// set, so the deadline must fire and the error must carry the last
	// observed state.
	seq := []statusRow{{State: "installing", OpType: "install"}}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(80*time.Millisecond, 5*time.Millisecond)

	_, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchInstall, "myapp", "market.olares"))
	if err == nil {
		t.Fatalf("expected timeout error, got nil")
	}
	var to *watchTimeoutError
	if !errors.As(err, &to) {
		t.Fatalf("expected watchTimeoutError, got %T: %v", err, err)
	}
	if to.last == nil || to.last.State != "installing" {
		t.Fatalf("expected last state installing, got %+v", to.last)
	}
	if !strings.Contains(err.Error(), "installing") {
		t.Fatalf("timeout error must surface last state, got %q", err.Error())
	}
}

// Sanity-guard that the JSON tags on OperationResult haven't drifted: with
// FinalState/FinalOpType empty, the JSON output must NOT contain those keys
// (so existing scripted consumers keep their byte-identical output).
func TestOperationResultJSONOmitsFinalFieldsWhenUnset(t *testing.T) {
	r := OperationResult{App: "a", Operation: "install", Status: "accepted"}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(b), "finalState") || strings.Contains(string(b), "finalOpType") {
		t.Fatalf("non-watch JSON must omit finalState/finalOpType; got %s", b)
	}

	r.FinalState = "running"
	r.FinalOpType = "install"
	b, err = json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(b), `"finalState":"running"`) {
		t.Fatalf("watch JSON must include finalState; got %s", b)
	}
}

// helper that builds a printable description of a watchTarget for error
// messages — keeps the test helpers self-contained.
func describeTarget(t watchTarget) string {
	return fmt.Sprintf("op=%s app=%s source=%s matchOpType=%v absentMeansSuccess=%v",
		t.op, t.appName, t.source, t.matchOpType, t.absentMeansSuccess)
}

var _ = describeTarget // referenced by future tests; suppress lint
