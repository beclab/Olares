package appstate

import (
	"context"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func getAM(t *testing.T, b *baseStatefulApp, name string) *appsv1.ApplicationManager {
	t.Helper()
	var am appsv1.ApplicationManager
	if err := b.client.Get(context.TODO(), types.NamespacedName{Name: name}, &am); err != nil {
		t.Fatalf("get AM %s: %v", name, err)
	}
	return &am
}

func TestUpdateStatusIncrementsGenerationAndPrependsRecord(t *testing.T) {
	am := testutil.NewAppManager("nginx", testutil.WithState(appsv1.Installing))
	am.Status.OpGeneration = 5
	am.Status.OpRecords = []appsv1.OpRecord{{OpID: "old", Message: "old"}}

	c := testutil.NewFakeClient(am)
	b := &baseStatefulApp{manager: am, client: c}

	now := metav1.Now()
	rec := &appsv1.OpRecord{OpID: "new", Message: "new", StateTime: &now}
	if err := b.updateStatus(context.TODO(), am, appsv1.Running, rec, "done", "reason"); err != nil {
		t.Fatalf("updateStatus: %v", err)
	}

	got := getAM(t, b, "nginx")
	if got.Status.State != appsv1.Running {
		t.Errorf("state=%q want Running", got.Status.State)
	}
	if got.Status.Message != "done" || got.Status.Reason != "reason" {
		t.Errorf("message/reason=%q/%q", got.Status.Message, got.Status.Reason)
	}
	if got.Status.OpGeneration != 6 {
		t.Errorf("OpGeneration=%d want 6", got.Status.OpGeneration)
	}
	if len(got.Status.OpRecords) != 2 {
		t.Fatalf("OpRecords len=%d want 2", len(got.Status.OpRecords))
	}
	if got.Status.OpRecords[0].OpID != "new" {
		t.Errorf("newest record not prepended: %q", got.Status.OpRecords[0].OpID)
	}
}

func TestUpdateStatusCapsRecordsAt20(t *testing.T) {
	am := testutil.NewAppManager("nginx", testutil.WithState(appsv1.Installing))
	existing := make([]appsv1.OpRecord, 20)
	for i := range existing {
		existing[i] = appsv1.OpRecord{OpID: "old"}
	}
	am.Status.OpRecords = existing

	c := testutil.NewFakeClient(am)
	b := &baseStatefulApp{manager: am, client: c}

	now := metav1.Now()
	rec := &appsv1.OpRecord{OpID: "new", StateTime: &now}
	if err := b.updateStatus(context.TODO(), am, appsv1.Running, rec, "msg", ""); err != nil {
		t.Fatalf("updateStatus: %v", err)
	}

	got := getAM(t, b, "nginx")
	if len(got.Status.OpRecords) != 20 {
		t.Fatalf("OpRecords len=%d want 20 (capped)", len(got.Status.OpRecords))
	}
	if got.Status.OpRecords[0].OpID != "new" {
		t.Errorf("newest record should be first, got %q", got.Status.OpRecords[0].OpID)
	}
}

func TestUpdateStatusNilRecordKeepsRecords(t *testing.T) {
	am := testutil.NewAppManager("nginx", testutil.WithState(appsv1.Installing))
	am.Status.OpRecords = []appsv1.OpRecord{{OpID: "old"}}
	c := testutil.NewFakeClient(am)
	b := &baseStatefulApp{manager: am, client: c}

	if err := b.updateStatus(context.TODO(), am, appsv1.Running, nil, "msg", ""); err != nil {
		t.Fatalf("updateStatus: %v", err)
	}
	got := getAM(t, b, "nginx")
	if len(got.Status.OpRecords) != 1 {
		t.Errorf("OpRecords len=%d want 1 (unchanged)", len(got.Status.OpRecords))
	}
}
