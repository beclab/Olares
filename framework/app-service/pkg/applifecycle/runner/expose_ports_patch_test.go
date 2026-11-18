package runner

import (
	"context"
	"testing"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/fake"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
)

func TestPatchExposeConfig_UsesLatestResourceVersion(t *testing.T) {
	store := fake.NewAMStore()
	store.Add(&appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "app1"},
		Spec:       appv1.ApplicationManagerSpec{Config: `{"old":true}`},
	})
	snap, err := store.Get(context.Background(), "app1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if err := store.PatchStatus(context.Background(), "app1", snap.ResourceVersion, ports.StatusPatch{
		State:  appv1.Installing,
		Reason: "ExposePortsStarted",
	}); err != nil {
		t.Fatalf("PatchStatus: %v", err)
	}

	r := &Runner{amStore: store}
	if err := r.patchExposeConfig(context.Background(), "app1", []byte(`{"ports":1}`)); err != nil {
		t.Fatalf("patchExposeConfig: %v", err)
	}
	got, err := store.Get(context.Background(), "app1")
	if err != nil {
		t.Fatalf("Get after patch: %v", err)
	}
	if got.Spec.Config != `{"ports":1}` {
		t.Fatalf("Spec.Config = %q, want assigned ports", got.Spec.Config)
	}
	if got.Status.State != appv1.Installing {
		t.Fatalf("Status.State = %q, want Installing", got.Status.State)
	}
}
