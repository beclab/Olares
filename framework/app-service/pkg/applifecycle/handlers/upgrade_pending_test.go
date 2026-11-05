package handlers_test

import (
	"context"
	"testing"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/apiserver/api"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/handlers"
)

func TestOnUpgradePendingTick_SlotFreeAdvances(t *testing.T) {
	am := &appv1.ApplicationManager{ObjectMeta: metav1.ObjectMeta{Name: "upgrade-me"}}
	res := handlers.OnUpgradePendingTick(context.Background(), domain.World{
		AM: am,
		ClusterStats: domain.ClusterStats{
			DownloadingCount: 0,
			NextDownloading:  "upgrade-me",
		},
	})
	if res.Err != nil {
		t.Fatal(res.Err)
	}
	if res.Next != appv1.UpgradeDownloading {
		t.Fatalf("Next = %s, want UpgradeDownloading", res.Next)
	}
}

func TestOnUpgradePendingTick_BlockedByDownloading(t *testing.T) {
	am := &appv1.ApplicationManager{ObjectMeta: metav1.ObjectMeta{Name: "upgrade-me"}}
	res := handlers.OnUpgradePendingTick(context.Background(), domain.World{
		AM:           am,
		ClusterStats: domain.ClusterStats{DownloadingCount: 1, NextDownloading: "upgrade-me"},
	})
	if res.Err != nil {
		t.Fatal(res.Err)
	}
	if res.Next != "" {
		t.Fatalf("Next = %s, want stay in UpgradePending", res.Next)
	}
}

func TestOnUpgradePendingCancelingTick_RestoresStopped(t *testing.T) {
	am := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{
			Name: "upgrade-me",
			Annotations: map[string]string{
				handlers.AnnotPreUpgradeState: appv1.Stopped.String(),
				api.AppPreUpgradeConfigKey:    `{"version":"1.0.0"}`,
				api.AppTargetChartVersionKey:  "2.0.0",
				api.AppVersionKey:             "1.0.0",
			},
		},
		Spec:   appv1.ApplicationManagerSpec{Config: `{"version":"2.0.0"}`},
		Status: appv1.ApplicationManagerStatus{State: appv1.UpgradePendingCanceling},
	}
	res := handlers.OnUpgradePendingCancelingTick(context.Background(), domain.World{AM: am})
	if res.Err != nil {
		t.Fatal(res.Err)
	}
	if res.Next != appv1.Stopped {
		t.Fatalf("Next = %s, want Stopped", res.Next)
	}
	spec, ok := findEffect[domain.PatchSpec](res.Effects)
	if !ok {
		t.Fatal("no PatchSpec restoring the pre-upgrade config")
	}
	if spec.Config != `{"version":"1.0.0"}` {
		t.Fatalf("restored config = %q", spec.Config)
	}
	if spec.SetAnnotations[api.AppVersionKey] != "" {
		t.Fatalf("deployed chart version rewritten to %q", spec.SetAnnotations[api.AppVersionKey])
	}
	if len(spec.DeleteAnnotations) != 2 ||
		spec.DeleteAnnotations[0] != api.AppPreUpgradeConfigKey ||
		spec.DeleteAnnotations[1] != api.AppTargetChartVersionKey {
		t.Fatalf("DeleteAnnotations = %v, want snapshot and target version cleared", spec.DeleteAnnotations)
	}
	patch, ok := findEffect[domain.PatchAM](res.Effects)
	if !ok {
		t.Fatal("no PatchAM")
	}
	if patch.Status.Reason != "UpgradePendingCanceled" {
		t.Fatalf("Reason = %q, want UpgradePendingCanceled", patch.Status.Reason)
	}
}
