package runner_test

import (
	"context"
	"testing"
	"time"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/appcfg"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/runner"
)

// seedAMWith installs an AM carrying a Spec.Config blob and an initial
// Status, for the cases where seedAM's minimal shape is not enough.
func (h *harness) seedAMWith(name, state, config string, status appv1.ApplicationManagerStatus) {
	h.t.Helper()
	status.State = appv1.ApplicationManagerState(state)
	h.amStore.Add(&appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: appv1.ApplicationManagerSpec{
			AppName:      name,
			AppOwner:     "alice",
			AppNamespace: name + "-alice",
			Type:         appv1.App,
			Config:       config,
		},
		Status: status,
	})
	h.amStore.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: name}})
}

// TestCallWorkload_PopulatesAppCfg pins that the Runner materialises
// AM.Spec.Config into WorkloadRequest.AppCfg before invoking a port.
//
// Regression: the field was left nil, so imgpull's strict cast
// rejected every image pull with "expected non-nil
// *appcfg.ApplicationConfig, got <nil>" — no ImageManager CR was ever
// created and no app could install.
func TestCallWorkload_PopulatesAppCfg(t *testing.T) {
	h := newHarness(t)

	var got any
	h.imagePuller.ResolveRefsFunc = func(_ context.Context, _ *appv1.ApplicationManager, appCfg any) ([]appv1.Ref, error) {
		got = appCfg
		return nil, nil
	}
	h.table.Add(domain.Transition{
		From: appv1.Downloading, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{Effects: []domain.Effect{
				domain.CallWorkload{Kind: domain.WorkloadPullImage},
			}}
		},
	})
	h.seedAMWith("app1", string(appv1.Downloading),
		`{"AppName":"app1","Namespace":"app1-alice"}`,
		appv1.ApplicationManagerStatus{})

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	h.runner.Stop()

	cfg, ok := got.(*appcfg.ApplicationConfig)
	if !ok || cfg == nil {
		t.Fatalf("ResolveRefs got AppCfg %#v, want non-nil *appcfg.ApplicationConfig", got)
	}
	if cfg.AppName != "app1" {
		t.Errorf("AppCfg.AppName = %q, want app1", cfg.AppName)
	}
	if calls := h.imagePuller.Calls(); len(calls) != 1 || calls[0].Method != "CreateJob" {
		t.Errorf("image puller calls = %+v, want single CreateJob", calls)
	}
}

// TestCallWorkload_AppCfgAbsent_LeavesRequestNil pins that an empty
// Spec.Config produces an untyped nil rather than a typed nil pointer,
// so adapters that fall back on `req.AppCfg == nil` still take their
// fallback path.
func TestCallWorkload_AppCfgAbsent_LeavesRequestNil(t *testing.T) {
	h := newHarness(t)

	sawNil := false
	h.imagePuller.ResolveRefsFunc = func(_ context.Context, _ *appv1.ApplicationManager, appCfg any) ([]appv1.Ref, error) {
		sawNil = appCfg == nil
		return nil, nil
	}
	h.table.Add(domain.Transition{
		From: appv1.Downloading, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{Effects: []domain.Effect{
				domain.CallWorkload{Kind: domain.WorkloadPullImage},
			}}
		},
	})
	h.seedAM("app1", string(appv1.Downloading))

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	h.runner.Stop()

	if !sawNil {
		t.Error("AppCfg was not an untyped nil for an AM with empty Spec.Config")
	}
}

// pullHarness seeds an AM in Downloading whose Tick dispatches an
// image pull, and returns the harness ready to be stepped.
func pullHarness(t *testing.T) *harness {
	t.Helper()
	h := newHarness(t)
	h.table.Add(domain.Transition{
		From: appv1.Downloading, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{Effects: []domain.Effect{
				domain.CallWorkload{
					Kind:      domain.WorkloadPullImage,
					OnSuccess: domain.WorkloadReady{},
					OnFailure: domain.WorkloadFailed{Reason: "ImagePullFailed"},
				},
			}}
		},
	})
	h.seedAMWith("app1", string(appv1.Downloading),
		`{"AppName":"app1"}`, appv1.ApplicationManagerStatus{
			OpType: appv1.InstallOp,
			OpID:   "op-pull",
		})
	return h
}

// awaitTicker blocks until the pull goroutine has registered its poll
// ticker with the fake clock, so Advance cannot race ahead of it.
func (h *harness) awaitTicker() {
	h.t.Helper()
	for i := 0; i < 1000; i++ {
		if h.clock.Pending() > 0 {
			return
		}
		time.Sleep(time.Millisecond)
	}
	h.t.Fatal("pull goroutine never registered a poll ticker")
}

// TestImagePull_WaitsForCompletion pins that the pull dispatcher
// blocks until the bytes have landed, not merely until CreateJob
// returns.
//
// Regression: the dispatcher treated CreateJob returning nil as a
// finished download, so Downloading fell through to Installing within
// milliseconds — helm ran against images that were still pulling, and
// the one-at-a-time download slot was released before any bytes moved.
func TestImagePull_WaitsForCompletion(t *testing.T) {
	h := pullHarness(t)
	h.imagePuller.SetProgress("app1", ports.ImageProgress{Percent: 40})

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	h.awaitTicker()

	// Mid-pull: a poll tick must not resolve the workload.
	h.clock.Advance(imagePollStep)
	time.Sleep(20 * time.Millisecond)
	if action, ok := h.runner.Pending().Peek("app1"); ok {
		t.Fatalf("pull resolved to %T while still at 40%%", action)
	}

	h.imagePuller.SetProgress("app1", ports.ImageProgress{Percent: 100, Complete: true})
	h.clock.Advance(imagePollStep)
	h.runner.Stop()

	action, ok := h.runner.Pending().Peek("app1")
	if !ok {
		t.Fatal("pending queue is empty; expected WorkloadReady after Complete")
	}
	if _, ok := action.(domain.WorkloadReady); !ok {
		t.Fatalf("pending action = %#v, want WorkloadReady", action)
	}
}

// TestImagePull_PublishesProgressOnPercentChange pins that each
// observed percent change is RecordAppEvent'd (legacy
// ImageManagerClient.updateProgress), carrying the AM's OpID so
// market / bfl can attach the tick to the in-flight install.
func TestImagePull_PublishesProgressOnPercentChange(t *testing.T) {
	h := pullHarness(t)
	h.imagePuller.SetProgress("app1", ports.ImageProgress{Percent: 40.12})

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	h.awaitTicker()

	h.clock.Advance(imagePollStep)
	waitForProgressEvent(t, h, "40.12")

	h.imagePuller.SetProgress("app1", ports.ImageProgress{Percent: 40.12})
	h.clock.Advance(imagePollStep)
	time.Sleep(20 * time.Millisecond)
	if n := countProgressEvents(h, "40.12"); n != 1 {
		t.Fatalf("duplicate 40.12 events = %d, want 1", n)
	}

	h.imagePuller.SetProgress("app1", ports.ImageProgress{Percent: 100, Complete: true})
	h.clock.Advance(imagePollStep)
	waitForProgressEvent(t, h, "100.00")
	h.runner.Stop()

	var saw4012, saw100 bool
	for _, ev := range h.recorder.Events() {
		if ev.Progress == "40.12" {
			saw4012 = true
			if ev.OpID != "op-pull" {
				t.Errorf("40.12 OpID = %q, want op-pull", ev.OpID)
			}
			if ev.OpType != string(appv1.InstallOp) {
				t.Errorf("40.12 OpType = %q, want install", ev.OpType)
			}
			if ev.State != string(appv1.Downloading) {
				t.Errorf("40.12 State = %q, want downloading", ev.State)
			}
			if ev.User != "alice" || ev.Name != "app1" {
				t.Errorf("40.12 User/Name = %q/%q", ev.User, ev.Name)
			}
			if ev.RawAppName != "app1" {
				t.Errorf("40.12 RawAppName = %q, want app1 (AppName fallback)", ev.RawAppName)
			}
		}
		if ev.Progress == "100.00" {
			saw100 = true
			if ev.OpID != "op-pull" {
				t.Errorf("100%% OpID = %q, want op-pull", ev.OpID)
			}
		}
	}
	if !saw4012 || !saw100 {
		t.Fatalf("progress events missing 40.12=%v 100=%v; got %+v", saw4012, saw100, h.recorder.Events())
	}
}

// TestImagePull_ProgressEventUsesCatalogIdentity pins the fields
// ImageManagerClient.updateProgress used to publish. The market card
// is keyed by the catalog name, which for middleware is Spec.RawAppName
// and not the AM object name ({namespace}-{app}). An upgrade pull
// must advertise state=upgrading even when the goroutine's status
// snapshot still says downloading.
func TestImagePull_ProgressEventUsesCatalogIdentity(t *testing.T) {
	h := newHarness(t)
	h.table.Add(domain.Transition{
		From: appv1.Downloading, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{Effects: []domain.Effect{
				domain.CallWorkload{
					Kind:      domain.WorkloadPullImage,
					OnSuccess: domain.WorkloadReady{},
					OnFailure: domain.WorkloadFailed{Reason: "UpgradeImagePullFailed"},
				},
			}}
		},
	})
	const amName = "citus-middleware-citus"
	h.amStore.Add(&appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{
			Name: amName,
			Annotations: map[string]string{
				"bytetrade.io/market-source": "market.olares",
			},
		},
		Spec: appv1.ApplicationManagerSpec{
			AppName:      "citus",
			RawAppName:   "citus",
			AppOwner:     "alice",
			AppNamespace: "citus-middleware",
			Type:         appv1.Middleware,
			OpType:       appv1.UpgradeOp,
			Source:       "market",
			Config:       `{"AppName":"citus","title":"Citus","icon":"citus.png"}`,
		},
		Status: appv1.ApplicationManagerStatus{
			State:  appv1.Downloading,
			OpType: appv1.UpgradeOp,
			OpID:   "op-upgrade",
		},
	})
	h.amStore.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: "citus"}})
	h.imagePuller.SetProgress(amName, ports.ImageProgress{Percent: 12.5})

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: amName}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	h.awaitTicker()
	h.clock.Advance(imagePollStep)
	waitForProgressEvent(t, h, "12.50")
	h.runner.Stop()

	var got bool
	for _, ev := range h.recorder.Events() {
		if ev.Progress != "12.50" {
			continue
		}
		got = true
		if ev.Name != "citus" || ev.RawAppName != "citus" {
			t.Errorf("Name/RawAppName = %q/%q, want citus/citus", ev.Name, ev.RawAppName)
		}
		if ev.RawAppName == amName {
			t.Errorf("RawAppName fell back to the AM object name %q", amName)
		}
		if ev.State != string(appv1.UpgradeDownloading) {
			t.Errorf("State = %q, want upgradeDownloading", ev.State)
		}
		if ev.Type != string(appv1.Middleware) {
			t.Errorf("Type = %q, want middleware", ev.Type)
		}
		if ev.Title != "Citus" || ev.Icon != "citus.png" {
			t.Errorf("Title/Icon = %q/%q, want Citus/citus.png", ev.Title, ev.Icon)
		}
		if ev.MarketSource != "market.olares" {
			t.Errorf("MarketSource = %q, want market.olares", ev.MarketSource)
		}
		if ev.ChartOwner != "alice" {
			t.Errorf("ChartOwner = %q, want alice", ev.ChartOwner)
		}
		if ev.OpID != "op-upgrade" || ev.OpType != string(appv1.UpgradeOp) {
			t.Errorf("OpID/OpType = %q/%q, want op-upgrade/upgrade", ev.OpID, ev.OpType)
		}
	}
	if !got {
		t.Fatalf("progress event 12.50 missing; got %+v", h.recorder.Events())
	}
}

func waitForProgressEvent(t *testing.T, h *harness, progress string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if countProgressEvents(h, progress) > 0 {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("timed out waiting for progress event %s; got %+v", progress, h.recorder.Events())
}

func countProgressEvents(h *harness, progress string) int {
	n := 0
	for _, ev := range h.recorder.Events() {
		if ev.Progress == progress {
			n++
		}
	}
	return n
}

// TestImagePull_BytesDoneWithoutCompleteResolves pins that the
// download slot is released as soon as every node's offset/total
// reaches 100%, even while ImageManager is still unpacking layers
// into containerd (Complete stays false until that finishes).
//
// The unpack window used to keep the AM in Downloading at 100%
// until the CR flipped to completed; that wait now belongs to
// Installing (helm install + WaitForStartUp).
func TestImagePull_BytesDoneWithoutCompleteResolves(t *testing.T) {
	h := pullHarness(t)
	h.imagePuller.SetProgress("app1", ports.ImageProgress{Percent: 99})

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	h.awaitTicker()

	h.clock.Advance(imagePollStep)
	time.Sleep(20 * time.Millisecond)
	if action, ok := h.runner.Pending().Peek("app1"); ok {
		t.Fatalf("pull resolved to %T while still at 99%%", action)
	}

	h.imagePuller.SetProgress("app1", ports.ImageProgress{Percent: 100, Complete: false})
	h.clock.Advance(imagePollStep)
	h.runner.Stop()

	action, ok := h.runner.Pending().Peek("app1")
	if !ok {
		t.Fatal("pending queue is empty; expected WorkloadReady at 100% bytes (unpack still running)")
	}
	if _, ok := action.(domain.WorkloadReady); !ok {
		t.Fatalf("pending action = %#v, want WorkloadReady", action)
	}
}

// TestImagePull_FailedProgressSurfaces pins that a Failed
// ImageProgress ends the poll with the ImageManager's own message.
func TestImagePull_FailedProgressSurfaces(t *testing.T) {
	h := pullHarness(t)
	h.imagePuller.SetProgress("app1", ports.ImageProgress{
		Failed:  true,
		Reason:  "failed",
		Message: "pull access denied",
	})

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	h.awaitTicker()
	h.clock.Advance(imagePollStep)
	h.runner.Stop()

	action, ok := h.runner.Pending().Peek("app1")
	if !ok {
		t.Fatal("pending queue is empty; expected WorkloadFailed")
	}
	wf, ok := action.(domain.WorkloadFailed)
	if !ok {
		t.Fatalf("pending action = %#v, want WorkloadFailed", action)
	}
	if wf.Reason != "ImagePullFailed" {
		t.Errorf("Reason = %q, want ImagePullFailed", wf.Reason)
	}
	if wf.Message != "pull access denied" {
		t.Errorf("Message = %q, want the ImageManager message", wf.Message)
	}
}

// imagePollStep is comfortably longer than the Runner's poll interval
// so a single Advance always fires exactly one tick.
const imagePollStep = 2 * time.Second

// TestReconcile_ConflictRestoresPendingAction pins that a popped
// Action survives a dispatch that fails partway.
//
// Regression: Pop is destructive, so an optimistic-concurrency
// conflict during effect dispatch dropped the callback. Phase recovery
// can redispatch after restart, but restoring the callback in the same
// Runner avoids repeating work that already completed.
func TestReconcile_ConflictRestoresPendingAction(t *testing.T) {
	h := newHarness(t)
	// A long timeout must not turn the conflict retry into a delayed
	// RequeueAfter. StatusTime is unset, so the deadline is the full
	// timeout measured from now.
	h.table.SetTimeout(appv1.Downloading, 24*time.Hour)

	// Two PatchAM effects in one Result: both patch under the same
	// World ResourceVersion, so the first bumps the stored RV and the
	// second conflicts. That is the cheapest faithful stand-in for a
	// racing writer.
	h.table.Add(domain.Transition{
		From: appv1.Downloading, On: domain.WorkloadReady{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{Effects: []domain.Effect{
				domain.PatchAM{Status: appv1.ApplicationManagerStatus{
					State: appv1.Downloading, Message: "first",
				}},
				domain.PatchAM{Status: appv1.ApplicationManagerStatus{
					State: appv1.Installing, Message: "second",
				}},
			}}
		},
	})
	h.seedAM("app1", string(appv1.Downloading))
	h.runner.Pending().Deliver("app1", domain.WorkloadReady{})

	out, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"})
	if err != nil {
		t.Fatalf("Reconcile err = %v, want nil (conflict folds into Requeue)", err)
	}
	if !out.Requeue || out.RequeueAfter != 0 {
		t.Fatalf("outcome = %+v, want Requeue=true and RequeueAfter=0 (a long state timeout must not delay the conflict retry)", out)
	}

	restored, ok := h.runner.Pending().Peek("app1")
	if !ok {
		t.Fatal("pending queue is empty; the WorkloadReady callback was lost")
	}
	if _, ok := restored.(domain.WorkloadReady); !ok {
		t.Fatalf("restored action = %T, want WorkloadReady", restored)
	}
}

// TestReconcile_RestoreDoesNotClobberFresherAction pins the
// "most recent wins" half of the restore contract: a callback that
// landed while the failing Reconcile was running must not be
// overwritten by the action being unwound.
func TestReconcile_RestoreDoesNotClobberFresherAction(t *testing.T) {
	h := newHarness(t)

	h.table.Add(domain.Transition{
		From: appv1.Downloading, On: domain.WorkloadReady{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			// Deliver a fresher action mid-flight, then conflict.
			h.runner.Pending().Deliver("app1", domain.WorkloadFailed{Reason: "Fresher"})
			return domain.Result{Effects: []domain.Effect{
				domain.PatchAM{Status: appv1.ApplicationManagerStatus{
					State: appv1.Downloading, Message: "first",
				}},
				domain.PatchAM{Status: appv1.ApplicationManagerStatus{
					State: appv1.Installing, Message: "second",
				}},
			}}
		},
	})
	h.seedAM("app1", string(appv1.Downloading))
	h.runner.Pending().Deliver("app1", domain.WorkloadReady{})

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	pending, ok := h.runner.Pending().Peek("app1")
	if !ok {
		t.Fatal("pending queue is empty")
	}
	wf, ok := pending.(domain.WorkloadFailed)
	if !ok || wf.Reason != "Fresher" {
		t.Fatalf("pending action = %#v, want the fresher WorkloadFailed", pending)
	}
}

// TestEmitEvent_ReportsPatchedStatus pins that an event emitted after
// a PatchAM reports the status that patch wrote.
//
// Regression: applyEmitEvent read World.AM, the pre-Reconcile
// snapshot, so the "download started" event carried the previous
// operation's leftover Progress — the UI jumped straight to 100%.
func TestEmitEvent_ReportsPatchedStatus(t *testing.T) {
	h := newHarness(t)

	h.table.Add(domain.Transition{
		From: appv1.Pending, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{Effects: []domain.Effect{
				domain.PatchAM{Status: appv1.ApplicationManagerStatus{
					State:    appv1.Downloading,
					Progress: "0.00",
				}},
				domain.EmitEvent{Type: "Normal", Reason: "DownloadStarted"},
			}}
		},
	})
	// Progress left at 100.00 by the previous (uninstall) operation.
	h.seedAMWith("app1", string(appv1.Pending), "",
		appv1.ApplicationManagerStatus{Progress: "100.00"})

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	events := h.recorder.Events()
	if len(events) != 1 {
		t.Fatalf("events = %+v, want exactly one", events)
	}
	if events[0].Progress != "0.00" {
		t.Errorf("event Progress = %q, want 0.00", events[0].Progress)
	}
	if events[0].State != string(appv1.Downloading) {
		t.Errorf("event State = %q, want %q", events[0].State, appv1.Downloading)
	}
}
