package k8sstore_test

import (
	"context"
	"testing"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/k8sstore"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/testutil"
)

func mustGet(t *testing.T, c client.Client, name string) *appv1.ApplicationManager {
	t.Helper()
	am := &appv1.ApplicationManager{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: name}, am); err != nil {
		t.Fatalf("Get after write: %v", err)
	}
	return am
}

// -----------------------------------------------------------------
// PatchStatus
// -----------------------------------------------------------------

func TestPatchStatus_ShadowMode_NoWrite(t *testing.T) {
	ctx := context.Background()
	seed := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "am"},
		Status:     appv1.ApplicationManagerStatus{State: appv1.Pending},
	}
	c := testutil.NewFakeClient(seed)
	stored := mustGet(t, c, "am")

	s := k8sstore.New(c, k8sstore.WithShadowMode(true))
	if !s.InShadowMode() {
		t.Fatal("expected shadow mode on")
	}
	if err := s.PatchStatus(ctx, "am", stored.ResourceVersion, ports.StatusPatch{
		State: appv1.Downloading,
	}); err != nil {
		t.Fatalf("PatchStatus (shadow): %v", err)
	}
	after := mustGet(t, c, "am")
	if after.Status.State != appv1.Pending {
		t.Errorf("shadow-mode PatchStatus mutated state: got %q want %q", after.Status.State, appv1.Pending)
	}
	if after.ResourceVersion != stored.ResourceVersion {
		t.Errorf("shadow-mode PatchStatus bumped RV: got %q want %q", after.ResourceVersion, stored.ResourceVersion)
	}
}

func TestPatchStatus_Live_AppliesFieldMerge(t *testing.T) {
	ctx := context.Background()
	seed := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "am"},
		Status: appv1.ApplicationManagerStatus{
			State:    appv1.Pending,
			Message:  "original",
			Reason:   "OriginalReason",
			OpID:     "op-1",
			Progress: "0.00",
		},
	}
	c := testutil.NewFakeClient(seed)
	stored := mustGet(t, c, "am")

	s := k8sstore.New(c)
	if err := s.PatchStatus(ctx, "am", stored.ResourceVersion, ports.StatusPatch{
		State:   appv1.Downloading,
		Message: "downloading now",
	}); err != nil {
		t.Fatalf("PatchStatus: %v", err)
	}
	after := mustGet(t, c, "am")
	if after.Status.State != appv1.Downloading {
		t.Errorf("state: got %q want Downloading", after.Status.State)
	}
	if after.Status.Message != "downloading now" {
		t.Errorf("message: got %q", after.Status.Message)
	}
	// Empty patch fields must be preserved (field-level merge, not
	// wholesale replace).
	if after.Status.Reason != "OriginalReason" {
		t.Errorf("reason mangled: got %q want OriginalReason", after.Status.Reason)
	}
	if after.Status.OpID != "op-1" {
		t.Errorf("opID mangled: got %q want op-1", after.Status.OpID)
	}
	if after.Status.Progress != "0.00" {
		t.Errorf("progress mangled: got %q want 0.00", after.Status.Progress)
	}
	if after.ResourceVersion == stored.ResourceVersion {
		t.Errorf("live PatchStatus did not bump RV: still %q", after.ResourceVersion)
	}
}

func TestPatchStatus_RVConflict(t *testing.T) {
	ctx := context.Background()
	seed := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "am"},
		Status:     appv1.ApplicationManagerStatus{State: appv1.Pending},
	}
	c := testutil.NewFakeClient(seed)
	s := k8sstore.New(c)

	err := s.PatchStatus(ctx, "am", "99999", ports.StatusPatch{State: appv1.Downloading})
	if err == nil {
		t.Fatal("PatchStatus with stale RV: want error, got nil")
	}
	if !apierrors.IsConflict(err) {
		t.Fatalf("want Conflict, got %T %v", err, err)
	}
	// Ensure the stale attempt did not mutate anything.
	after := mustGet(t, c, "am")
	if after.Status.State != appv1.Pending {
		t.Errorf("stale-RV PatchStatus leaked mutation: state=%q", after.Status.State)
	}
}

func TestPatchStatus_NotFound(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewFakeClient()
	s := k8sstore.New(c)

	err := s.PatchStatus(ctx, "missing", "1", ports.StatusPatch{State: appv1.Downloading})
	if err == nil {
		t.Fatal("PatchStatus on missing AM: want error, got nil")
	}
	if !apierrors.IsNotFound(err) {
		t.Fatalf("want NotFound, got %T %v", err, err)
	}
}

func TestPatchStatus_PayloadClearSemantic(t *testing.T) {
	ctx := context.Background()
	seed := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "am"},
		Status: appv1.ApplicationManagerStatus{
			State:   appv1.Downloading,
			Payload: map[string]string{"phase": "pulling", "stale": "true"},
		},
	}
	c := testutil.NewFakeClient(seed)
	stored := mustGet(t, c, "am")

	s := k8sstore.New(c)
	// Empty non-nil map = clear the entire Payload map.
	if err := s.PatchStatus(ctx, "am", stored.ResourceVersion, ports.StatusPatch{
		State:   appv1.Initializing,
		Payload: map[string]string{},
	}); err != nil {
		t.Fatalf("PatchStatus: %v", err)
	}
	after := mustGet(t, c, "am")
	if len(after.Status.Payload) != 0 {
		t.Errorf("expected Payload cleared, got %v", after.Status.Payload)
	}
}

func TestPatchStatus_PayloadMergeSemantic(t *testing.T) {
	ctx := context.Background()
	seed := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "am"},
		Status: appv1.ApplicationManagerStatus{
			State:   appv1.Downloading,
			Payload: map[string]string{"phase": "pulling"},
		},
	}
	c := testutil.NewFakeClient(seed)
	stored := mustGet(t, c, "am")

	s := k8sstore.New(c)
	if err := s.PatchStatus(ctx, "am", stored.ResourceVersion, ports.StatusPatch{
		Payload: map[string]string{"progress": "50"},
	}); err != nil {
		t.Fatalf("PatchStatus: %v", err)
	}
	after := mustGet(t, c, "am")
	if got := after.Status.Payload["phase"]; got != "pulling" {
		t.Errorf("phase not preserved: got %q", got)
	}
	if got := after.Status.Payload["progress"]; got != "50" {
		t.Errorf("progress not added: got %q", got)
	}
}

func TestPatchStatus_PayloadDeleteKeysSemantic(t *testing.T) {
	ctx := context.Background()
	seed := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "am"},
		Status: appv1.ApplicationManagerStatus{
			State: appv1.InstallingCanceling,
			Payload: map[string]string{
				"phase":  "wait-ns",
				"intent": "installFailed",
				"keep":   "yes",
			},
		},
	}
	c := testutil.NewFakeClient(seed)
	stored := mustGet(t, c, "am")

	s := k8sstore.New(c)
	if err := s.PatchStatus(ctx, "am", stored.ResourceVersion, ports.StatusPatch{
		Payload:           map[string]string{"cleanup": "done"},
		DeletePayloadKeys: []string{"phase", "intent"},
	}); err != nil {
		t.Fatalf("PatchStatus: %v", err)
	}
	after := mustGet(t, c, "am")
	if _, ok := after.Status.Payload["phase"]; ok {
		t.Errorf("phase key retained: %v", after.Status.Payload)
	}
	if _, ok := after.Status.Payload["intent"]; ok {
		t.Errorf("intent key retained: %v", after.Status.Payload)
	}
	if after.Status.Payload["keep"] != "yes" || after.Status.Payload["cleanup"] != "done" {
		t.Errorf("Payload = %v, want unrelated + merged keys retained", after.Status.Payload)
	}
}

func TestPatchStatus_OpRecordPrepended(t *testing.T) {
	ctx := context.Background()
	seed := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "am"},
		Status: appv1.ApplicationManagerStatus{
			State: appv1.Pending,
			OpRecords: []appv1.OpRecord{
				{OpID: "old-1", Message: "prior"},
			},
		},
	}
	c := testutil.NewFakeClient(seed)
	stored := mustGet(t, c, "am")

	s := k8sstore.New(c)
	if err := s.PatchStatus(ctx, "am", stored.ResourceVersion, ports.StatusPatch{
		State:    appv1.Downloading,
		OpRecord: &appv1.OpRecord{OpID: "new-1", Message: "starting"},
	}); err != nil {
		t.Fatalf("PatchStatus: %v", err)
	}
	after := mustGet(t, c, "am")
	if len(after.Status.OpRecords) != 2 {
		t.Fatalf("OpRecords length: got %d want 2", len(after.Status.OpRecords))
	}
	if after.Status.OpRecords[0].OpID != "new-1" {
		t.Errorf("OpRecord not prepended: [0]=%+v", after.Status.OpRecords[0])
	}
	if after.Status.OpRecords[1].OpID != "old-1" {
		t.Errorf("existing OpRecord dropped: [1]=%+v", after.Status.OpRecords[1])
	}
}

// -----------------------------------------------------------------
// PatchSpecAndAnnotations
// -----------------------------------------------------------------

func TestPatchSpecAndAnnotations_ShadowMode(t *testing.T) {
	ctx := context.Background()
	seed := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "am",
			Annotations: map[string]string{"keep": "yes"},
		},
		Spec: appv1.ApplicationManagerSpec{OpType: appv1.InstallOp},
	}
	c := testutil.NewFakeClient(seed)
	stored := mustGet(t, c, "am")
	s := k8sstore.New(c, k8sstore.WithShadowMode(true))

	if err := s.PatchSpecAndAnnotations(ctx, "am", stored.ResourceVersion,
		&ports.SpecPatch{OpType: appv1.CancelOp},
		&ports.AnnotationPatch{Set: map[string]string{"new": "value"}},
	); err != nil {
		t.Fatalf("PatchSpecAndAnnotations (shadow): %v", err)
	}
	after := mustGet(t, c, "am")
	if after.Spec.OpType != appv1.InstallOp {
		t.Errorf("shadow leaked spec write: %q", after.Spec.OpType)
	}
	if _, ok := after.Annotations["new"]; ok {
		t.Errorf("shadow leaked annotation write: %v", after.Annotations)
	}
	if after.ResourceVersion != stored.ResourceVersion {
		t.Errorf("shadow bumped RV: %q -> %q", stored.ResourceVersion, after.ResourceVersion)
	}
}

func TestPatchSpecAndAnnotations_Live(t *testing.T) {
	ctx := context.Background()
	seed := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "am",
			Annotations: map[string]string{"keep": "yes", "drop": "me"},
		},
		Spec: appv1.ApplicationManagerSpec{OpType: appv1.InstallOp, Config: "old-config"},
	}
	c := testutil.NewFakeClient(seed)
	stored := mustGet(t, c, "am")
	s := k8sstore.New(c)

	if err := s.PatchSpecAndAnnotations(ctx, "am", stored.ResourceVersion,
		&ports.SpecPatch{OpType: appv1.CancelOp, Config: "new-config"},
		&ports.AnnotationPatch{
			Set:    map[string]string{"new": "value"},
			Delete: []string{"drop"},
		},
	); err != nil {
		t.Fatalf("PatchSpecAndAnnotations: %v", err)
	}
	after := mustGet(t, c, "am")
	if after.Spec.OpType != appv1.CancelOp {
		t.Errorf("OpType: %q want CancelOp", after.Spec.OpType)
	}
	if after.Spec.Config != "new-config" {
		t.Errorf("Config: %q want new-config", after.Spec.Config)
	}
	if v, ok := after.Annotations["keep"]; !ok || v != "yes" {
		t.Errorf("keep annotation lost: %v", after.Annotations)
	}
	if _, ok := after.Annotations["drop"]; ok {
		t.Errorf("drop annotation not deleted: %v", after.Annotations)
	}
	if v := after.Annotations["new"]; v != "value" {
		t.Errorf("new annotation missing: %v", after.Annotations)
	}
}

func TestPatchSpecAndAnnotations_NilSubPatchesAllowed(t *testing.T) {
	ctx := context.Background()
	seed := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "am"},
	}
	c := testutil.NewFakeClient(seed)
	stored := mustGet(t, c, "am")
	s := k8sstore.New(c)

	// spec=nil + ann=nil is a legal (if useless) call — must not
	// crash. Behaviour is a no-op patch that still bumps RV via
	// the API server's monotonic counter.
	if err := s.PatchSpecAndAnnotations(ctx, "am", stored.ResourceVersion, nil, nil); err != nil {
		t.Fatalf("PatchSpecAndAnnotations(nil, nil): %v", err)
	}
}

func TestPatchSpecAndAnnotations_RVConflict(t *testing.T) {
	ctx := context.Background()
	seed := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "am"},
	}
	c := testutil.NewFakeClient(seed)
	s := k8sstore.New(c)

	err := s.PatchSpecAndAnnotations(ctx, "am", "99999",
		&ports.SpecPatch{OpType: appv1.CancelOp}, nil)
	if err == nil {
		t.Fatal("PatchSpecAndAnnotations stale RV: want error, got nil")
	}
	if !apierrors.IsConflict(err) {
		t.Fatalf("want Conflict, got %T %v", err, err)
	}
}

// -----------------------------------------------------------------
// UpdateFull
// -----------------------------------------------------------------

func TestUpdateFull_ShadowMode(t *testing.T) {
	ctx := context.Background()
	seed := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "am"},
		Status:     appv1.ApplicationManagerStatus{State: appv1.Pending},
	}
	c := testutil.NewFakeClient(seed)
	stored := mustGet(t, c, "am")

	s := k8sstore.New(c, k8sstore.WithShadowMode(true))
	mutated := stored.DeepCopy()
	mutated.Status.State = appv1.Downloading

	if err := s.UpdateFull(ctx, mutated); err != nil {
		t.Fatalf("UpdateFull (shadow): %v", err)
	}
	after := mustGet(t, c, "am")
	if after.Status.State != appv1.Pending {
		t.Errorf("shadow UpdateFull leaked: state=%q", after.Status.State)
	}
	if after.ResourceVersion != stored.ResourceVersion {
		t.Errorf("shadow UpdateFull bumped RV: %q -> %q", stored.ResourceVersion, after.ResourceVersion)
	}
}

func TestUpdateFull_Live(t *testing.T) {
	ctx := context.Background()
	seed := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "am"},
		Status:     appv1.ApplicationManagerStatus{State: appv1.Pending},
	}
	c := testutil.NewFakeClient(seed)
	stored := mustGet(t, c, "am")

	s := k8sstore.New(c)
	mutated := stored.DeepCopy()
	mutated.Status.State = appv1.Downloading

	if err := s.UpdateFull(ctx, mutated); err != nil {
		t.Fatalf("UpdateFull: %v", err)
	}
	after := mustGet(t, c, "am")
	if after.Status.State != appv1.Downloading {
		t.Errorf("state: %q want Downloading", after.Status.State)
	}
}

func TestUpdateFull_NilError(t *testing.T) {
	ctx := context.Background()
	c := testutil.NewFakeClient()
	s := k8sstore.New(c)

	if err := s.UpdateFull(ctx, nil); err == nil {
		t.Fatal("UpdateFull(nil): want error, got nil")
	}
}

// -----------------------------------------------------------------
// SetShadowMode runtime toggle
// -----------------------------------------------------------------

func TestSetShadowMode_RuntimeToggle(t *testing.T) {
	ctx := context.Background()
	seed := &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: "am"},
		Status:     appv1.ApplicationManagerStatus{State: appv1.Pending},
	}
	c := testutil.NewFakeClient(seed)
	s := k8sstore.New(c, k8sstore.WithShadowMode(true))

	// First write with shadow ON — no effect.
	stored := mustGet(t, c, "am")
	if err := s.PatchStatus(ctx, "am", stored.ResourceVersion, ports.StatusPatch{
		State: appv1.Downloading,
	}); err != nil {
		t.Fatalf("shadow PatchStatus: %v", err)
	}
	if mustGet(t, c, "am").Status.State != appv1.Pending {
		t.Fatal("shadow ON but state changed")
	}

	// Flip shadow OFF and retry — write must land.
	s.SetShadowMode(false)
	if s.InShadowMode() {
		t.Fatal("SetShadowMode(false) had no effect")
	}
	stored = mustGet(t, c, "am")
	if err := s.PatchStatus(ctx, "am", stored.ResourceVersion, ports.StatusPatch{
		State: appv1.Downloading,
	}); err != nil {
		t.Fatalf("live PatchStatus: %v", err)
	}
	if mustGet(t, c, "am").Status.State != appv1.Downloading {
		t.Fatal("shadow OFF but write did not land")
	}
}
