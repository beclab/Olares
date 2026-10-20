package market

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/beclab/Olares/cli/pkg/clusterclient"
	"github.com/beclab/Olares/cli/pkg/credential"
)

const crashingPodList = `{"items":[{
  "metadata":{"name":"myapp-7d9f","namespace":"myapp-tester"},
  "status":{
    "initContainerStatuses":[{"name":"init-permissions","restartCount":0,"state":{"terminated":{"reason":"Completed"}}}],
    "containerStatuses":[
      {"name":"myapp","restartCount":4,"state":{"waiting":{"reason":"CrashLoopBackOff"}}},
      {"name":"sidecar","restartCount":0,"state":{"running":{}}},
      {"name":"worker","restartCount":0,"state":{"waiting":{"reason":"ImagePullBackOff"}}}
    ]}}],"totalItems":1}`

func TestPodSignalsKeepsOnlyRestartingOrStuckContainers(t *testing.T) {
	var list clusterclient.ListResponse[probedPod]
	if err := json.Unmarshal([]byte(crashingPodList), &list); err != nil {
		t.Fatal(err)
	}
	got := podSignals(list.Items)
	want := []podSignal{
		{Namespace: "myapp-tester", Pod: "myapp-7d9f", Container: "myapp", Reason: "CrashLoopBackOff", Restarts: 4},
		{Namespace: "myapp-tester", Pod: "myapp-7d9f", Container: "worker", Reason: "ImagePullBackOff"},
	}
	if len(got) != len(want) {
		t.Fatalf("signals = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("signal %d = %+v, want %+v", i, got[i], want[i])
		}
	}
	if !got[0].fatal() || got[1].fatal() {
		t.Fatalf("CrashLoopBackOff must be fatal and ImagePullBackOff must not: %+v", got)
	}
}

func TestClusterPodProberSelectsTheAppsOwnPods(t *testing.T) {
	var gotPath, gotSelector string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotSelector = r.URL.Query().Get("labelSelector")
		_, _ = w.Write([]byte(crashingPodList))
	}))
	defer srv.Close()

	c := clusterclient.NewClient(http.DefaultClient, &credential.ResolvedProfile{OlaresID: "tester@olares.test", ControlHubURL: srv.URL})
	signals, err := newClusterPodProber(c)(context.Background(), "myapp")
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/kapis/resources.kubesphere.io/v1alpha3/pods" {
		t.Fatalf("path = %q", gotPath)
	}
	if gotSelector != appNameLabel+"=myapp,"+appOwnerLabel+"=tester" {
		t.Fatalf("labelSelector = %q", gotSelector)
	}
	if len(signals) != 2 {
		t.Fatalf("signals = %+v", signals)
	}
}

func crashingProbe(context.Context, string) ([]podSignal, error) {
	return []podSignal{{Namespace: "myapp-tester", Pod: "myapp-7d9f", Container: "myapp", Reason: "CrashLoopBackOff", Restarts: 4}}, nil
}

func TestWaitForTerminalFailOnCrashLoopStopsAnInFlightInstall(t *testing.T) {
	srv := newFakeStateServer(t, "myapp", "market.olares", []statusRow{{State: "installing", OpType: "install"}})
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(5*time.Second, 5*time.Millisecond)
	opts.FailOnCrashLoop = true
	opts.podProbe = crashingProbe

	row, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchInstall, "myapp", "market.olares"))
	var crash *watchCrashLoopError
	if !errors.As(err, &crash) {
		t.Fatalf("expected watchCrashLoopError, got %T: %v", err, err)
	}
	if row.State != "installing" {
		t.Fatalf("row state = %q, want installing", row.State)
	}
	for _, want := range []string{"CrashLoopBackOff", "myapp-tester/myapp-7d9f/myapp --previous", "market uninstall myapp"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q does not mention %q", err, want)
		}
	}
}

func TestWaitForTerminalTimeoutReportsCrashingContainers(t *testing.T) {
	srv := newFakeStateServer(t, "myapp", "market.olares", []statusRow{{State: "installing", OpType: "install"}})
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(50*time.Millisecond, 5*time.Millisecond)
	opts.podProbe = crashingProbe

	_, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchInstall, "myapp", "market.olares"))
	var timeout *watchTimeoutError
	if !errors.As(err, &timeout) {
		t.Fatalf("expected watchTimeoutError, got %T: %v", err, err)
	}
	if !strings.Contains(err.Error(), "app containers: pod myapp-7d9f container myapp: CrashLoopBackOff (restarts=4)") {
		t.Fatalf("timeout error does not carry the pod signal: %v", err)
	}
}

func TestWaitForTerminalPodProbeErrorFallsBackToTheMarketRow(t *testing.T) {
	seq := []statusRow{{State: "installing", OpType: "install"}, {State: "running", OpType: "install"}}
	srv := newFakeStateServer(t, "myapp", "market.olares", seq)
	mc := newTestMarketClient(t, srv.srv.URL)
	opts := quietOpts(5*time.Second, 5*time.Millisecond)
	opts.FailOnCrashLoop = true
	opts.podProbe = func(context.Context, string) ([]podSignal, error) { return nil, errors.New("HTTP 403") }

	row, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchInstall, "myapp", "market.olares"))
	if err != nil || row.State != "running" {
		t.Fatalf("row=%+v err=%v, want running", row, err)
	}
}

func TestWaitForTerminalInstallStoppedByAppServiceIsAFailure(t *testing.T) {
	for _, last := range []statusRow{
		{State: "stopped", OpType: "install", Reason: "StartUpFailed", Message: "app myapp stop completed"},
		{State: "stopped", OpType: "stop", Reason: "Unschedulable"},
	} {
		srv := newFakeStateServer(t, "myapp", "market.olares", []statusRow{
			{State: "installing", OpType: "install"},
			{State: "stopping", OpType: last.OpType, Reason: last.Reason},
			last,
		})
		mc := newTestMarketClient(t, srv.srv.URL)
		opts := quietOpts(5*time.Second, 5*time.Millisecond)
		opts.podProbe = func(context.Context, string) ([]podSignal, error) { return nil, nil }

		_, err := waitForTerminal(context.Background(), mc, opts, newWatchTarget(watchInstall, "myapp", "market.olares"))
		var fail *watchFailureError
		if !errors.As(err, &fail) {
			t.Fatalf("%s: expected watchFailureError, got %T: %v", last.Reason, err, err)
		}
		for _, want := range []string{"reason=" + last.Reason, "still installed", "market resume myapp", "market uninstall myapp"} {
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("error %q does not mention %q", err, want)
			}
		}
	}
}
