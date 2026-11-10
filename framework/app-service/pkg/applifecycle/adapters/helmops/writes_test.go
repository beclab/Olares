package helmops_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	"k8s.io/client-go/rest"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/helmops"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/errcode"
	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
)

// -----------------------------------------------------------------
// Method-by-method: happy path, error propagation, factory error
// short-circuit. The dispatch matrix is small enough to test
// exhaustively.
// -----------------------------------------------------------------

func newOpsWithFake(t *testing.T) (*helmops.Ops, *fakeFactory) {
	t.Helper()
	ff := &fakeFactory{ops: &fakeHelmOps{}}
	o := helmops.New(testutil.NewFakeClient(), &rest.Config{},
		helmops.WithHelmOpsFactory(ff.fn()))
	return o, ff
}

func TestInstall_Success(t *testing.T) {
	o, ff := newOpsWithFake(t)
	am := baseAM(t, "app-x")
	res := o.Install(context.Background(), ports.WorkloadRequest{AM: am})
	if res.Err != nil {
		t.Fatalf("Install: %v", res.Err)
	}
	if got := ff.ops.called(); len(got) != 1 || got[0] != "Install" {
		t.Errorf("methods: %v want [Install]", got)
	}
}

func TestInstall_ErrorPropagates(t *testing.T) {
	o, ff := newOpsWithFake(t)
	wantErr := errors.New("chart apply failed")
	ff.ops.installErr = wantErr
	am := baseAM(t, "app-x")
	res := o.Install(context.Background(), ports.WorkloadRequest{AM: am})
	if !errors.Is(res.Err, wantErr) {
		t.Errorf("err: %v want %v", res.Err, wantErr)
	}
	if res.LandedReason != "" {
		t.Errorf("LandedReason: %q want empty (runner backfills)", res.LandedReason)
	}
}

// TestInstall_UnschedulablePodsRouteToStopping pins the one place
// this adapter makes a routing decision instead of passing an error
// through. The errcode sentinels are wrapped inside the helm error
// and unreadable above this layer, so if the classification is
// dropped an unschedulable app gets torn down (InstallFailed)
// instead of parked at replicas=0 for the user to resume.
//
// The truth table is two independent bits: which pods are pending
// (app-side vs server-side, deciding the stop-all flag) and whether
// HAMI reported the unschedulability (refining the user-facing
// reason).
func TestInstall_UnschedulablePodsRouteToStopping(t *testing.T) {
	for _, tc := range []struct {
		name        string
		err         error
		wantReason  string
		wantStopAll bool
	}{
		{
			name:        "app-side pending",
			err:         errcode.ErrPodPending,
			wantReason:  constants.AppUnschedulable,
			wantStopAll: false,
		},
		{
			name:        "app-side pending, hami",
			err:         errors.Join(errcode.ErrPodPending, errcode.ErrHamiUnschedulable),
			wantReason:  constants.AppHamiSchedulable,
			wantStopAll: false,
		},
		{
			name:        "server-side pending",
			err:         errcode.ErrServerSidePodPending,
			wantReason:  constants.AppUnschedulable,
			wantStopAll: true,
		},
		{
			name:        "server-side pending, hami",
			err:         errors.Join(errcode.ErrServerSidePodPending, errcode.ErrHamiUnschedulable),
			wantReason:  constants.AppHamiSchedulable,
			wantStopAll: true,
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			o, ff := newOpsWithFake(t)
			// Wrapped, as the real helm path delivers it — the
			// classification must survive an errors.Is walk, not
			// just an equality check.
			ff.ops.installErr = fmt.Errorf("helm install: %w", tc.err)
			res := o.Install(context.Background(), ports.WorkloadRequest{AM: baseAM(t, "app-x")})

			if res.Err == nil {
				t.Fatal("Err = nil; the failure must still surface")
			}
			if res.LandedState != string(appv1.Stopping) {
				t.Errorf("LandedState = %q, want Stopping", res.LandedState)
			}
			if res.LandedReason != tc.wantReason {
				t.Errorf("LandedReason = %q, want %q", res.LandedReason, tc.wantReason)
			}
			if res.StopAll != tc.wantStopAll {
				t.Errorf("StopAll = %v, want %v", res.StopAll, tc.wantStopAll)
			}
		})
	}
}

// TestInstall_GenericErrorHasNoRoutingHint is the negative half of
// the above: an ordinary helm failure must leave LandedState empty
// so the handler applies its default InstallFailed target.
func TestInstall_GenericErrorHasNoRoutingHint(t *testing.T) {
	o, ff := newOpsWithFake(t)
	ff.ops.installErr = errors.New("chart values invalid")
	res := o.Install(context.Background(), ports.WorkloadRequest{AM: baseAM(t, "app-x")})
	if res.LandedState != "" {
		t.Errorf("LandedState = %q, want empty for a generic helm error", res.LandedState)
	}
	if res.StopAll {
		t.Error("StopAll = true, want false")
	}
}

func TestPrepareUpgrade_NoAnnotations_Succeeds(t *testing.T) {
	o, _ := newOpsWithFake(t)
	am := baseAM(t, "app-x")
	res := o.PrepareUpgrade(context.Background(), ports.WorkloadRequest{AM: am})
	if res.Err != nil {
		t.Fatalf("PrepareUpgrade: %v", res.Err)
	}
}

func TestPrepareUpgrade_DowngradeRejected(t *testing.T) {
	o := helmops.New(testutil.NewFakeClient(), &rest.Config{},
		helmops.WithHelmOpsFactory((&fakeFactory{ops: &fakeHelmOps{}}).fn()),
		helmops.WithGetVersion(func(_ context.Context, _ *rest.Config, _, _ string) (string, int, error) {
			return "2.0.0", 1, nil
		}),
	)
	am := baseAM(t, "app-x")
	am.Annotations = map[string]string{
		"bytetrade.io/target-chart-version": "1.0.0",
	}
	res := o.PrepareUpgrade(context.Background(), ports.WorkloadRequest{AM: am})
	if res.Err == nil {
		t.Fatal("expected downgrade rejection")
	}
	if res.LandedReason != "DowngradeNotAllowed" {
		t.Fatalf("LandedReason = %q, want DowngradeNotAllowed", res.LandedReason)
	}
}

func TestPrepareUpgrade_VersionGuardPassesWithoutRepo(t *testing.T) {
	o := helmops.New(testutil.NewFakeClient(), &rest.Config{},
		helmops.WithHelmOpsFactory((&fakeFactory{ops: &fakeHelmOps{}}).fn()),
		helmops.WithGetVersion(func(_ context.Context, _ *rest.Config, _, _ string) (string, int, error) {
			return "1.0.0", 1, nil
		}),
	)
	am := baseAM(t, "app-x")
	am.Annotations = map[string]string{
		"bytetrade.io/target-chart-version": "1.2.0",
	}
	res := o.PrepareUpgrade(context.Background(), ports.WorkloadRequest{AM: am})
	if res.Err != nil {
		t.Fatalf("PrepareUpgrade: %v", res.Err)
	}
}

func TestUpgrade_Success(t *testing.T) {
	o, ff := newOpsWithFake(t)
	am := baseAM(t, "app-x")
	res := o.Upgrade(context.Background(), ports.WorkloadRequest{AM: am})
	if res.Err != nil {
		t.Fatalf("Upgrade: %v", res.Err)
	}
	if got := ff.ops.called(); len(got) != 1 || got[0] != "Upgrade" {
		t.Errorf("methods: %v want [Upgrade]", got)
	}
}

func TestUpgrade_ErrorPropagates(t *testing.T) {
	o, ff := newOpsWithFake(t)
	wantErr := errors.New("upgrade failed")
	ff.ops.upgradeErr = wantErr
	am := baseAM(t, "app-x")
	res := o.Upgrade(context.Background(), ports.WorkloadRequest{AM: am})
	if !errors.Is(res.Err, wantErr) {
		t.Errorf("err: %v want %v", res.Err, wantErr)
	}
}

func TestApplyEnv_Success(t *testing.T) {
	o, ff := newOpsWithFake(t)
	am := baseAM(t, "app-x")
	res := o.ApplyEnv(context.Background(), ports.WorkloadRequest{AM: am})
	if res.Err != nil {
		t.Fatalf("ApplyEnv: %v", res.Err)
	}
	if got := ff.ops.called(); len(got) != 1 || got[0] != "ApplyEnv" {
		t.Errorf("methods: %v want [ApplyEnv]", got)
	}
}

func TestApplyEnv_ErrorPropagates(t *testing.T) {
	o, ff := newOpsWithFake(t)
	wantErr := errors.New("env update failed")
	ff.ops.applyEnvErr = wantErr
	am := baseAM(t, "app-x")
	res := o.ApplyEnv(context.Background(), ports.WorkloadRequest{AM: am})
	if !errors.Is(res.Err, wantErr) {
		t.Errorf("err: %v want %v", res.Err, wantErr)
	}
}

func TestUninstall_Success(t *testing.T) {
	o, ff := newOpsWithFake(t)
	am := baseAM(t, "app-x")
	res := o.Uninstall(context.Background(), ports.WorkloadRequest{AM: am})
	if res.Err != nil {
		t.Fatalf("Uninstall: %v", res.Err)
	}
	if got := ff.ops.called(); len(got) != 1 || got[0] != "Uninstall" {
		t.Errorf("methods: %v want [Uninstall]", got)
	}
}

func TestUninstall_ErrorPropagates(t *testing.T) {
	o, ff := newOpsWithFake(t)
	wantErr := errors.New("chart delete failed")
	ff.ops.uninstallErr = wantErr
	am := baseAM(t, "app-x")
	res := o.Uninstall(context.Background(), ports.WorkloadRequest{AM: am})
	if !errors.Is(res.Err, wantErr) {
		t.Errorf("err: %v want %v", res.Err, wantErr)
	}
}

// TestUninstall_NotFoundIsTolerated pins the legacy tolerance:
// "not found" substring in the helm error is treated as success.
// Both variants (case-preserved, upper-case) must be tolerated
// because helm wraps errors from different drivers with different
// casing.
func TestUninstall_NotFoundIsTolerated(t *testing.T) {
	cases := []string{
		"release not found",
		"Release Not Found",
		"chart myapp is Not Found in cluster",
	}
	for _, msg := range cases {
		t.Run(msg, func(t *testing.T) {
			o, ff := newOpsWithFake(t)
			ff.ops.uninstallErr = errors.New(msg)
			am := baseAM(t, "app-x")
			res := o.Uninstall(context.Background(), ports.WorkloadRequest{AM: am})
			if res.Err != nil {
				t.Errorf("tolerated %q should not propagate: %v", msg, res.Err)
			}
		})
	}
}

// TestUninstall_NonToleratedError_Propagates guards the tolerance's
// specificity: an error that does NOT contain "not found" must
// propagate.
func TestUninstall_NonToleratedError_Propagates(t *testing.T) {
	o, ff := newOpsWithFake(t)
	wantErr := errors.New("connection refused")
	ff.ops.uninstallErr = wantErr
	am := baseAM(t, "app-x")
	res := o.Uninstall(context.Background(), ports.WorkloadRequest{AM: am})
	if !errors.Is(res.Err, wantErr) {
		t.Errorf("err: %v want %v", res.Err, wantErr)
	}
}

func TestUninstallAll_Success(t *testing.T) {
	o, ff := newOpsWithFake(t)
	am := baseAM(t, "app-x")
	res := o.UninstallAll(context.Background(), ports.WorkloadRequest{AM: am})
	if res.Err != nil {
		t.Fatalf("UninstallAll: %v", res.Err)
	}
	if got := ff.ops.called(); len(got) != 1 || got[0] != "UninstallAll" {
		t.Errorf("methods: %v want [UninstallAll]", got)
	}
}

// TestUninstallAll_NotFoundIsTolerated pins the same substring
// tolerance as Uninstall — the two variants must share exactly one
// classification helper so their behaviour cannot drift.
func TestUninstallAll_NotFoundIsTolerated(t *testing.T) {
	o, ff := newOpsWithFake(t)
	ff.ops.uninstallAllErr = errors.New("release not found")
	am := baseAM(t, "app-x")
	res := o.UninstallAll(context.Background(), ports.WorkloadRequest{AM: am})
	if res.Err != nil {
		t.Errorf("tolerated 'not found' should not propagate: %v", res.Err)
	}
}

func TestStop_ScalesToZero(t *testing.T) {
	o, ff := newOpsWithFake(t)
	am := baseAMWithReplicas(t, "app-x")
	res := o.Stop(context.Background(), ports.WorkloadRequest{AM: am})
	if res.Err != nil {
		t.Fatalf("Stop: %v", res.Err)
	}
	if len(ff.ops.scaleReplicasLog) != 1 || ff.ops.scaleReplicasLog[0] != 0 {
		t.Errorf("Scale replicas: %v want [0]", ff.ops.scaleReplicasLog)
	}
}

func TestStop_ScaleErrorPropagates(t *testing.T) {
	o, ff := newOpsWithFake(t)
	wantErr := errors.New("scale failed")
	ff.ops.scaleErr = wantErr
	am := baseAMWithReplicas(t, "app-x")
	res := o.Stop(context.Background(), ports.WorkloadRequest{AM: am})
	if !errors.Is(res.Err, wantErr) {
		t.Errorf("err: %v want %v", res.Err, wantErr)
	}
}

// TestResume_ScalesToNegativeOne pins the "restore manifest
// replicaCount" semantics: Scale(-1) is the sentinel the underlying
// helm pipeline treats as "use the chart-declared replica count".
func TestResume_ScalesToNegativeOne(t *testing.T) {
	o, ff := newOpsWithFake(t)
	am := baseAMWithReplicas(t, "app-x")
	res := o.Resume(context.Background(), ports.WorkloadRequest{AM: am})
	if res.Err != nil {
		t.Fatalf("Resume: %v", res.Err)
	}
	if len(ff.ops.scaleReplicasLog) != 1 || ff.ops.scaleReplicasLog[0] != -1 {
		t.Errorf("Scale replicas: %v want [-1]", ff.ops.scaleReplicasLog)
	}
}

func TestResume_ScaleErrorPropagates(t *testing.T) {
	o, ff := newOpsWithFake(t)
	wantErr := errors.New("scale-up failed")
	ff.ops.scaleErr = wantErr
	am := baseAMWithReplicas(t, "app-x")
	res := o.Resume(context.Background(), ports.WorkloadRequest{AM: am})
	if !errors.Is(res.Err, wantErr) {
		t.Errorf("err: %v want %v", res.Err, wantErr)
	}
}

func TestStop_NoWorkloadReplicas_UsesPatchPath(t *testing.T) {
	o, ff := newOpsWithFake(t)
	am := baseAM(t, "app-x")
	res := o.Stop(context.Background(), ports.WorkloadRequest{AM: am})
	if res.Err != nil {
		t.Fatalf("Stop patch path: %v", res.Err)
	}
	if len(ff.ops.scaleReplicasLog) != 0 {
		t.Errorf("Scale called on patch-path app: %v", ff.ops.scaleReplicasLog)
	}
}

func TestResume_NoWorkloadReplicas_UsesPatchPath(t *testing.T) {
	o, ff := newOpsWithFake(t)
	am := baseAM(t, "app-x")
	res := o.Resume(context.Background(), ports.WorkloadRequest{AM: am})
	if res.Err != nil {
		t.Fatalf("Resume patch path: %v", res.Err)
	}
	if len(ff.ops.scaleReplicasLog) != 0 {
		t.Errorf("Scale called on patch-path app: %v", ff.ops.scaleReplicasLog)
	}
}

func TestWaitForStartUp_Success(t *testing.T) {
	o, ff := newOpsWithFake(t)
	am := baseAM(t, "app-x")
	res := o.WaitForStartUp(context.Background(), ports.WorkloadRequest{AM: am})
	if res.Err != nil {
		t.Fatalf("WaitForStartUp: %v", res.Err)
	}
	if got := ff.ops.called(); len(got) != 1 || got[0] != "WaitForStartUp" {
		t.Errorf("methods: %v want [WaitForStartUp]", got)
	}
}

func TestWaitForStartUp_ErrorPropagates(t *testing.T) {
	o, ff := newOpsWithFake(t)
	wantErr := errors.New("secret not found")
	ff.ops.waitStartUpErr = wantErr
	am := baseAM(t, "app-x")
	res := o.WaitForStartUp(context.Background(), ports.WorkloadRequest{AM: am})
	if !errors.Is(res.Err, wantErr) {
		t.Errorf("err: %v want %v", res.Err, wantErr)
	}
	if res.LandedState != "" {
		t.Errorf("LandedState = %q, want empty for a startup error that is not unschedulable", res.LandedState)
	}
}

func TestWaitForStartUp_UnschedulableRoutesToStopping(t *testing.T) {
	for _, tc := range []struct {
		name        string
		err         error
		wantReason  string
		wantStopAll bool
	}{
		{name: "app-side pending", err: errcode.ErrPodPending, wantReason: constants.AppUnschedulable},
		{name: "app-side pending, hami", err: errors.Join(errcode.ErrPodPending, errcode.ErrHamiUnschedulable), wantReason: constants.AppHamiSchedulable},
		{name: "server-side pending", err: errcode.ErrServerSidePodPending, wantReason: constants.AppUnschedulable, wantStopAll: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o, ff := newOpsWithFake(t)
			ff.ops.waitStartUpErr = tc.err
			res := o.WaitForStartUp(context.Background(), ports.WorkloadRequest{AM: baseAM(t, "app-x")})
			if res.LandedState != string(appv1.Stopping) {
				t.Errorf("LandedState = %q, want Stopping", res.LandedState)
			}
			if res.LandedReason != tc.wantReason {
				t.Errorf("LandedReason = %q, want %q", res.LandedReason, tc.wantReason)
			}
			if res.StopAll != tc.wantStopAll {
				t.Errorf("StopAll = %v, want %v", res.StopAll, tc.wantStopAll)
			}
		})
	}
}

func TestWaitForLaunch_Success(t *testing.T) {
	o, ff := newOpsWithFake(t)
	am := baseAM(t, "app-x")
	res := o.WaitForLaunch(context.Background(), ports.WorkloadRequest{AM: am})
	if res.Err != nil {
		t.Fatalf("WaitForLaunch: %v", res.Err)
	}
	if got := ff.ops.called(); len(got) != 1 || got[0] != "WaitForLaunch" {
		t.Errorf("methods: %v want [WaitForLaunch]", got)
	}
}

func TestWaitForLaunch_ErrorPropagates(t *testing.T) {
	o, ff := newOpsWithFake(t)
	wantErr := errors.New("wait for launch timed out")
	ff.ops.waitLaunchErr = wantErr
	am := baseAM(t, "app-x")
	res := o.WaitForLaunch(context.Background(), ports.WorkloadRequest{AM: am})
	if !errors.Is(res.Err, wantErr) {
		t.Errorf("err: %v want %v", res.Err, wantErr)
	}
}

// TestAllMethods_FactoryErrorShortCircuits pins that a
// factory-construction failure prevents the underlying helm call
// (no leakage of half-initialised state).
func TestAllMethods_FactoryErrorShortCircuits(t *testing.T) {
	tests := []struct {
		name string
		call func(o *helmops.Ops, req ports.WorkloadRequest) ports.WorkloadResult
	}{
		{"Install", func(o *helmops.Ops, r ports.WorkloadRequest) ports.WorkloadResult {
			return o.Install(context.Background(), r)
		}},
		{"Upgrade", func(o *helmops.Ops, r ports.WorkloadRequest) ports.WorkloadResult {
			return o.Upgrade(context.Background(), r)
		}},
		{"ApplyEnv", func(o *helmops.Ops, r ports.WorkloadRequest) ports.WorkloadResult {
			return o.ApplyEnv(context.Background(), r)
		}},
		{"Uninstall", func(o *helmops.Ops, r ports.WorkloadRequest) ports.WorkloadResult {
			return o.Uninstall(context.Background(), r)
		}},
		{"UninstallAll", func(o *helmops.Ops, r ports.WorkloadRequest) ports.WorkloadResult {
			return o.UninstallAll(context.Background(), r)
		}},
		{"Stop", func(o *helmops.Ops, r ports.WorkloadRequest) ports.WorkloadResult {
			return o.Stop(context.Background(), r)
		}},
		{"Resume", func(o *helmops.Ops, r ports.WorkloadRequest) ports.WorkloadResult {
			return o.Resume(context.Background(), r)
		}},
		{"WaitForStartUp", func(o *helmops.Ops, r ports.WorkloadRequest) ports.WorkloadResult {
			return o.WaitForStartUp(context.Background(), r)
		}},
		{"WaitForLaunch", func(o *helmops.Ops, r ports.WorkloadRequest) ports.WorkloadResult {
			return o.WaitForLaunch(context.Background(), r)
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ff := &fakeFactory{ops: &fakeHelmOps{}, err: errors.New("boom")}
			o := helmops.New(testutil.NewFakeClient(), &rest.Config{},
				helmops.WithHelmOpsFactory(ff.fn()))
			am := baseAM(t, "app-x")
			if tc.name == "Stop" || tc.name == "Resume" {
				am = baseAMWithReplicas(t, "app-x")
			}
			res := tc.call(o, ports.WorkloadRequest{AM: am})
			if res.Err == nil {
				t.Fatal("expected error from factory")
			}
			if len(ff.ops.called()) != 0 {
				t.Errorf("underlying method invoked despite factory error: %v", ff.ops.called())
			}
		})
	}
}
