package runner_test

import (
	"context"
	"testing"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/fake"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/runner"
)

func TestDispatchUninstall_UsesUninstallAllWhenAnnotated(t *testing.T) {
	h := newHarness(t)
	h.table.Add(domain.Transition{
		From: appv1.Uninstalling, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{Effects: []domain.Effect{
				domain.CallWorkload{
					Kind:      domain.WorkloadUninstall,
					OnSuccess: domain.WorkloadReady{},
					OnFailure: domain.WorkloadFailed{Reason: "HelmUninstallFailed"},
				},
			}}
		},
	})
	am := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{
			Name: "app1",
			Annotations: map[string]string{
				api.AppUninstallAllKey: "true",
			},
		},
		Spec: appv1.ApplicationManagerSpec{
			AppName:      "app1",
			AppOwner:     "alice",
			AppNamespace: "app1-alice",
			Type:         appv1.App,
			Config:       `{"AppName":"app1"}`,
		},
		Status: appv1.ApplicationManagerStatus{State: appv1.Uninstalling},
	}
	h.amStore.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: "app1"}})
	h.amStore.Add(am)

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	h.runner.Stop()

	if got := h.workloadOps.Calls(); len(got) != 1 || got[0].Method != "UninstallAll" {
		t.Fatalf("workload calls = %+v, want [UninstallAll]", got)
	}
}

func TestDispatchStop_DeletesAllocationsForTarget(t *testing.T) {
	compute := &fake.ComputeOps{}
	h := newHarnessWithCompute(t, compute)
	h.table.Add(domain.Transition{
		From: appv1.Stopping, On: domain.Tick{},
		Handler: func(_ context.Context, _ domain.World) domain.Result {
			return domain.Result{Effects: []domain.Effect{
				domain.CallWorkload{
					Kind:      domain.WorkloadStop,
					OnSuccess: domain.WorkloadReady{},
					OnFailure: domain.WorkloadFailed{Reason: "HelmStopFailed"},
				},
			}}
		},
	})
	am := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "app1"},
		Spec: appv1.ApplicationManagerSpec{
			AppName:      "app1",
			AppOwner:     "alice",
			AppNamespace: "app1-alice",
			Type:         appv1.App,
			Config:       `{"AppName":"app1"}`,
		},
		Status: appv1.ApplicationManagerStatus{State: appv1.Stopping},
	}
	h.amStore.AddApp(&appv1.Application{ObjectMeta: metav1.ObjectMeta{Name: "app1"}})
	h.amStore.Add(am)

	if _, err := h.runner.Reconcile(context.Background(), runner.Request{Name: "app1"}); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	h.runner.Stop()

	if got := h.workloadOps.Calls(); len(got) != 1 || got[0].Method != "Stop" {
		t.Fatalf("workload calls = %+v, want [Stop]", got)
	}
	if got := compute.Calls(); len(got) != 1 || got[0].Method != "DeleteForTarget" {
		t.Fatalf("compute calls = %+v, want [DeleteForTarget]", got)
	}
}
