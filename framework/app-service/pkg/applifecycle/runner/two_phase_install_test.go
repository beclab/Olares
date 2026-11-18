package runner_test

import (
	"context"
	"errors"
	"testing"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/fake"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/runner"
)

// The Runner-side support the two-phase install needs: a World bit
// telling the handler which route to take, a spec-write Effect for
// the rejection path, and the routing hints the workload adapter
// attaches to a failure.

// TestBuildWorld_HasWorkloadReplicas pins the World bit that decides
// whether validation runs before or after helm.
//
// A malformed Spec.Config must read as false rather than failing the
// Reconcile: false is the conservative single-phase route, and the
// parse error surfaces with a real message once the workload port
// materialises the same blob.
func TestBuildWorld_HasWorkloadReplicas(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config string
		want   bool
	}{
		{name: "absent", config: "", want: false},
		{name: "declared", config: `{"AppName":"app1","workloadReplicas":{"firefox":1}}`, want: true},
		{name: "empty map", config: `{"AppName":"app1","workloadReplicas":{}}`, want: false},
		{name: "unparseable", config: `{ not json`, want: false},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(t)

			var got bool
			h.table.Add(domain.Transition{
				From: appv1.Installing, On: domain.Tick{},
				Handler: func(_ context.Context, w domain.World) domain.Result {
					got = w.HasWorkloadReplicas
					return domain.Result{}
				},
			})
			h.seedAMWith("app1", string(appv1.Installing), tc.config,
				appv1.ApplicationManagerStatus{})

			if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
				t.Fatalf("Reconcile: %v", err)
			}
			if got != tc.want {
				t.Errorf("World.HasWorkloadReplicas = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestPatchSpec_WritesSpecThenStatus pins the bundle the post-helm
// rejection path emits: a spec write (op converted to a stop, plus
// the stop-all flag) followed by the status flip.
//
// The ordering is load-bearing. Both writes carry an optimistic-lock
// precondition on the AM's ResourceVersion and the spec write bumps
// it, so the Runner has to re-read before applying the status patch.
// If it did not, the PatchAM below would fail its precondition and
// the AM would be left with a Stop op but an Installing status.
func TestPatchSpec_WritesSpecThenStatus(t *testing.T) {
	h := newHarness(t)

	h.table.AddReachable(appv1.Installing, appv1.Stopping)
	h.table.Add(domain.Transition{
		From: appv1.Installing, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{
				Next: appv1.Stopping,
				Effects: []domain.Effect{
					domain.PatchSpec{
						OpType:         appv1.StopOp,
						SetAnnotations: map[string]string{"bytetrade.io/stop-all": "true"},
					},
					domain.PatchAM{Status: appv1.ApplicationManagerStatus{
						State:   appv1.Stopping,
						Reason:  "Unschedulable",
						Message: "insufficient GPU",
						OpType:  appv1.StopOp,
					}},
				},
			}
		},
	})
	h.seedAM("app1", string(appv1.Installing))

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	got, err := h.amStore.Get(context.Background(), "app1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Spec.OpType != appv1.StopOp {
		t.Errorf("Spec.OpType = %q, want StopOp", got.Spec.OpType)
	}
	if got.Annotations["bytetrade.io/stop-all"] != "true" {
		t.Errorf("annotations = %v, want stop-all=true", got.Annotations)
	}
	if got.Status.State != appv1.Stopping {
		t.Errorf("Status.State = %q, want Stopping; the status patch after a spec write must use the refreshed ResourceVersion",
			got.Status.State)
	}
	if got.Status.OpType != appv1.StopOp {
		t.Errorf("Status.OpType = %q, want StopOp", got.Status.OpType)
	}

	kinds := mutationKinds(h.amStore.Mutations())
	wantTail := []string{"PatchSpecAndAnnotations", "PatchStatus"}
	if len(kinds) < 2 || !equalStrings(kinds[len(kinds)-2:], wantTail) {
		t.Errorf("mutation tail = %v, want %v", kinds, wantTail)
	}
}

// TestWorkloadFailed_CarriesAdapterRoutingHints pins that
// LandedState and StopAll survive the trip from the workload port to
// the handler. They are the only channel through which "helm failed
// on an unschedulable pod, park the app instead of tearing it down"
// reaches a pure handler.
func TestWorkloadFailed_CarriesAdapterRoutingHints(t *testing.T) {
	h := newHarness(t)

	h.workloadOps.InstallFunc = func(_ context.Context, _ ports.WorkloadRequest) ports.WorkloadResult {
		return ports.WorkloadResult{
			LandedState:  string(appv1.Stopping),
			LandedReason: "Unschedulable",
			Err:          errors.New("pod is pending"),
			StopAll:      true,
		}
	}

	var seen domain.WorkloadFailed
	h.table.Add(domain.Transition{
		From: appv1.Installing, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{Effects: []domain.Effect{
				domain.CallWorkload{
					Kind:      domain.WorkloadInstall,
					OnSuccess: domain.WorkloadReady{},
					OnFailure: domain.WorkloadFailed{Reason: "HelmInstallFailed"},
				},
			}}
		},
	})
	h.table.Add(domain.Transition{
		From: appv1.Installing, On: domain.WorkloadFailed{},
		Handler: func(_ context.Context, w domain.World) domain.Result {
			seen, _ = w.Action.(domain.WorkloadFailed)
			return domain.Result{}
		},
	})
	h.seedAM("app1", string(appv1.Installing))

	// R1 dispatches; the callback lands in the pending queue.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	h.runner.Stop()

	// R2 consumes it.
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("R2: %v", err)
	}

	if seen.LandedState != appv1.Stopping {
		t.Errorf("WorkloadFailed.LandedState = %q, want Stopping", seen.LandedState)
	}
	if !seen.StopAll {
		t.Error("WorkloadFailed.StopAll = false, want true")
	}
	// The adapter's reason wins over the handler's OnFailure
	// fallback — see launchCallWorkload's priority comment.
	if seen.Reason != "Unschedulable" {
		t.Errorf("WorkloadFailed.Reason = %q, want Unschedulable", seen.Reason)
	}
}

// TestValidateInstall_RejectionSetsStopAll pins that a validation
// rejection consults ComputeOps for the stop's blast radius, and that
// a lookup error scopes the stop narrowly rather than tearing down a
// shared server other apps depend on.
func TestValidateInstall_RejectionSetsStopAll(t *testing.T) {
	for _, tc := range []struct {
		name    string
		manages bool
		lookErr error
		want    bool
	}{
		{name: "manages shared server", manages: true, want: true},
		{name: "app-scoped", manages: false, want: false},
		{name: "lookup error degrades to app-scoped", lookErr: errors.New("configmap unreadable"), want: false},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			computeOps := &fake.ComputeOps{
				ManagesSharedServerFunc: func(_ context.Context, _ any) (bool, error) {
					return tc.manages, tc.lookErr
				},
			}
			h := newHarnessWithCompute(t, computeOps)

			h.validator.ValidateInstallFunc = func(_ context.Context, _ ports.ValidationInput) (ports.ValidationDecision, error) {
				return ports.ValidationDecision{
					OK:      false,
					Reason:  "Unschedulable",
					Message: "insufficient GPU",
				}, nil
			}

			var seen domain.WorkloadFailed
			h.table.Add(domain.Transition{
				From: appv1.Installing, On: domain.Tick{},
				Handler: func(_ context.Context, _ domain.World) domain.Result {
					return domain.Result{Effects: []domain.Effect{
						domain.CallWorkload{
							Kind:      domain.WorkloadValidateInstall,
							OnSuccess: domain.WorkloadReady{},
							OnFailure: domain.WorkloadFailed{Reason: "InstallValidationRejected"},
						},
					}}
				},
			})
			h.table.Add(domain.Transition{
				From: appv1.Installing, On: domain.WorkloadFailed{},
				Handler: func(_ context.Context, w domain.World) domain.Result {
					seen, _ = w.Action.(domain.WorkloadFailed)
					return domain.Result{}
				},
			})
			h.seedAMWith("app1", string(appv1.Installing),
				`{"AppName":"app1"}`, appv1.ApplicationManagerStatus{})

			if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
				t.Fatalf("R1: %v", err)
			}
			h.runner.Stop()
			if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
				t.Fatalf("R2: %v", err)
			}

			if seen.StopAll != tc.want {
				t.Errorf("WorkloadFailed.StopAll = %v, want %v", seen.StopAll, tc.want)
			}
		})
	}
}

// TestValidateInstall_RejectionWithoutComputeOps pins the optional-
// port default: no ComputeOps means no shared-server knowledge, so
// the stop stays scoped to the app.
func TestValidateInstall_RejectionWithoutComputeOps(t *testing.T) {
	h := newHarness(t)
	h.validator.ValidateInstallFunc = func(_ context.Context, _ ports.ValidationInput) (ports.ValidationDecision, error) {
		return ports.ValidationDecision{OK: false, Reason: "Unschedulable"}, nil
	}

	var seen domain.WorkloadFailed
	h.table.Add(domain.Transition{
		From: appv1.Installing, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{Effects: []domain.Effect{
				domain.CallWorkload{
					Kind:      domain.WorkloadValidateInstall,
					OnFailure: domain.WorkloadFailed{},
				},
			}}
		},
	})
	h.table.Add(domain.Transition{
		From: appv1.Installing, On: domain.WorkloadFailed{},
		Handler: func(_ context.Context, w domain.World) domain.Result {
			seen, _ = w.Action.(domain.WorkloadFailed)
			return domain.Result{}
		},
	})
	h.seedAMWith("app1", string(appv1.Installing),
		`{"AppName":"app1"}`, appv1.ApplicationManagerStatus{})

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("R1: %v", err)
	}
	h.runner.Stop()
	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("R2: %v", err)
	}

	if seen.StopAll {
		t.Error("StopAll = true with no ComputeOps wired, want false")
	}
}

// TestDispatchWorkloadKind_WaitStartUp pins the routing for the
// readiness poll the scale-up phase depends on. Without it the kind
// would fall through to the unrouted branch and every install would
// fail right after scaling up.
func TestDispatchWorkloadKind_WaitStartUp(t *testing.T) {
	h := newHarness(t)

	h.table.Add(domain.Transition{
		From: appv1.Installing, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{Effects: []domain.Effect{
				domain.CallWorkload{Kind: domain.WorkloadWaitStartUp},
			}}
		},
	})
	h.seedAM("app1", string(appv1.Installing))

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	h.runner.Stop()

	calls := h.workloadOps.Calls()
	if len(calls) != 1 || calls[0].Method != "WaitForStartUp" {
		t.Fatalf("workload calls = %+v, want single WaitForStartUp", calls)
	}
}

// newHarnessWithCompute is newHarness plus a ComputeOps port. Kept
// here rather than folded into newHarness so the default harness
// keeps exercising the nil-ComputeOps path that most tests care
// about.
func newHarnessWithCompute(t *testing.T, computeOps ports.ComputeOps) *harness {
	t.Helper()
	h := newHarness(t)
	r, err := runner.New(runner.Config{
		Engine: h.engine, AMStore: h.amStore, WorkloadOps: h.workloadOps,
		Validator: h.validator, ImagePuller: h.imagePuller,
		Recorder: h.recorder, Clock: h.clock, Requeuer: h.requeuer,
		Pending: h.runner.Pending(), ComputeOps: computeOps,
	})
	if err != nil {
		t.Fatalf("runner.New: %v", err)
	}
	h.runner.Stop()
	h.runner = r
	t.Cleanup(r.Stop)
	return h
}

func mutationKinds(muts []fake.Mutation) []string {
	out := make([]string, len(muts))
	for i, m := range muts {
		out[i] = m.Kind
	}
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
