package appstate

import (
	"context"
	"testing"

	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

	got := getAM(t, c, "nginx")
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

	got := getAM(t, c, "nginx")
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
	got := getAM(t, c, "nginx")
	if len(got.Status.OpRecords) != 1 {
		t.Errorf("OpRecords len=%d want 1 (unchanged)", len(got.Status.OpRecords))
	}
}

// updateStatus must reject any (from -> to) edge not declared in
// StateTransitions. A jump like Pending -> Uninstalled would clobber the
// state machine's invariants, so the patch must be refused and the stored
// state must remain unchanged.
func TestUpdateStatusRejectsInvalidTransition(t *testing.T) {
	am := testutil.NewAppManager("nginx", testutil.WithState(appsv1.Pending))
	c := testutil.NewFakeClient(am)
	b := &baseStatefulApp{manager: am, client: c}

	err := b.updateStatus(context.TODO(), am, appsv1.Uninstalled, nil, "msg", "")
	if err == nil {
		t.Fatalf("updateStatus(Pending -> Uninstalled) returned nil, want error")
	}

	got := getAM(t, c, "nginx")
	if got.Status.State != appsv1.Pending {
		t.Errorf("state=%s want Pending (rejected write must not mutate)", got.Status.State)
	}
	if got.Status.OpGeneration != 0 {
		t.Errorf("OpGeneration=%d want 0 (rejected write must not bump generation)", got.Status.OpGeneration)
	}
}

// Idempotent same-state writes must still succeed even though the table
// does not list a (X -> X) self-loop. This protects RetryOnConflict
// re-reads, deferred Finally re-assertions, and operator workflows that
// refresh message/reason without changing state.
func TestUpdateStatusAllowsSelfWrite(t *testing.T) {
	am := testutil.NewAppManager("nginx", testutil.WithState(appsv1.InstallFailed))
	c := testutil.NewFakeClient(am)
	b := &baseStatefulApp{manager: am, client: c}

	if err := b.updateStatus(context.TODO(), am, appsv1.InstallFailed, nil, "updated msg", "updated reason"); err != nil {
		t.Fatalf("updateStatus self-write rejected: %v", err)
	}
	got := getAM(t, c, "nginx")
	if got.Status.State != appsv1.InstallFailed {
		t.Errorf("state=%s want InstallFailed", got.Status.State)
	}
	if got.Status.Message != "updated msg" || got.Status.Reason != "updated reason" {
		t.Errorf("message/reason=%q/%q want updated values", got.Status.Message, got.Status.Reason)
	}
	if got.Status.OpGeneration != 1 {
		t.Errorf("OpGeneration=%d want 1 (self-write should still bump generation)", got.Status.OpGeneration)
	}
}
