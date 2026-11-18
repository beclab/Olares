package fake_test

import (
	"context"
	"testing"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/fake"
	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
)

func newAM(name, state string) *appv1.ApplicationManager {
	return &appv1.ApplicationManager{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status: appv1.ApplicationManagerStatus{
			State: appv1.ApplicationManagerState(state),
		},
	}
}

// TestAMStore_GetNotFound pins that Get on a missing AM returns an
// apierrors.IsNotFound-shaped error — production callers depend on
// that predicate to distinguish "AM gone" from "API down".
func TestAMStore_GetNotFound(t *testing.T) {
	s := fake.NewAMStore()
	_, err := s.Get(context.Background(), "nope")
	if !apierrors.IsNotFound(err) {
		t.Fatalf("Get(missing) err = %v, want IsNotFound", err)
	}
}

// TestAMStore_AddThenGet pins the round-trip: Add gives us an RV;
// Get returns a deep copy with the same RV.
func TestAMStore_AddThenGet(t *testing.T) {
	s := fake.NewAMStore()
	rv := s.Add(newAM("app1", "Pending"))
	if rv == "" {
		t.Fatal("Add returned empty ResourceVersion")
	}

	got, err := s.Get(context.Background(), "app1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ResourceVersion != rv {
		t.Fatalf("Get RV = %q, want %q", got.ResourceVersion, rv)
	}
	if string(got.Status.State) != "Pending" {
		t.Fatalf("Get state = %q, want Pending", got.Status.State)
	}

	// Mutating the returned copy must NOT affect the store.
	got.Status.State = "Hacked"
	fresh, _ := s.Get(context.Background(), "app1")
	if string(fresh.Status.State) != "Pending" {
		t.Fatalf("Get returned a live reference (mutation leaked): got state=%q", fresh.Status.State)
	}
}

// TestAMStore_PatchStatus_OK pins the happy path + RV bump.
func TestAMStore_PatchStatus_OK(t *testing.T) {
	s := fake.NewAMStore()
	rv := s.Add(newAM("app1", "Pending"))

	err := s.PatchStatus(context.Background(), "app1", rv, ports.StatusPatch{
		State:   "Downloading",
		Message: "starting image pull",
	})
	if err != nil {
		t.Fatalf("PatchStatus: %v", err)
	}
	got, _ := s.Get(context.Background(), "app1")
	if string(got.Status.State) != "Downloading" {
		t.Fatalf("state after patch = %q, want Downloading", got.Status.State)
	}
	if got.Status.Message != "starting image pull" {
		t.Fatalf("message after patch = %q, want %q", got.Status.Message, "starting image pull")
	}
	if got.ResourceVersion == rv {
		t.Fatalf("RV did not bump after PatchStatus: still %q", got.ResourceVersion)
	}
}

// TestAMStore_PatchStatus_Conflict pins optimistic concurrency:
// patching with a stale RV must return apierrors.IsConflict.
func TestAMStore_PatchStatus_Conflict(t *testing.T) {
	s := fake.NewAMStore()
	rv := s.Add(newAM("app1", "Pending"))

	// First patch succeeds; second with the ORIGINAL rv must
	// conflict.
	if err := s.PatchStatus(context.Background(), "app1", rv, ports.StatusPatch{State: "Downloading"}); err != nil {
		t.Fatalf("first patch: %v", err)
	}
	err := s.PatchStatus(context.Background(), "app1", rv, ports.StatusPatch{State: "Installing"})
	if !apierrors.IsConflict(err) {
		t.Fatalf("second patch with stale RV err = %v, want IsConflict", err)
	}
}

// TestAMStore_PatchStatus_MergeSemantic pins that empty patch
// fields are preserved (field-level merge). Without this handlers
// would accidentally clear Reason/Message when they only meant to
// touch State.
func TestAMStore_PatchStatus_MergeSemantic(t *testing.T) {
	s := fake.NewAMStore()
	am := newAM("app1", "Installing")
	am.Status.Message = "installing"
	am.Status.Reason = "InProgress"
	rv := s.Add(am)

	// Patch only State; Message + Reason must survive.
	if err := s.PatchStatus(context.Background(), "app1", rv, ports.StatusPatch{State: "Running"}); err != nil {
		t.Fatalf("PatchStatus: %v", err)
	}
	got, _ := s.Get(context.Background(), "app1")
	if got.Status.Message != "installing" {
		t.Fatalf("Message dropped by empty-field patch: got %q, want %q", got.Status.Message, "installing")
	}
	if got.Status.Reason != "InProgress" {
		t.Fatalf("Reason dropped by empty-field patch: got %q, want %q", got.Status.Reason, "InProgress")
	}
}

// TestAMStore_PatchStatus_OpRecordPrepend pins that OpRecord is
// PREPENDED, matching legacy makeRecord + updateStatus semantics.
func TestAMStore_PatchStatus_OpRecordPrepend(t *testing.T) {
	s := fake.NewAMStore()
	am := newAM("app1", "Installing")
	am.Status.OpRecords = []appv1.OpRecord{{OpID: "old"}}
	rv := s.Add(am)

	if err := s.PatchStatus(context.Background(), "app1", rv, ports.StatusPatch{
		OpRecord: &appv1.OpRecord{OpID: "new"},
	}); err != nil {
		t.Fatalf("PatchStatus: %v", err)
	}
	got, _ := s.Get(context.Background(), "app1")
	if len(got.Status.OpRecords) != 2 {
		t.Fatalf("OpRecords len = %d, want 2", len(got.Status.OpRecords))
	}
	if got.Status.OpRecords[0].OpID != "new" {
		t.Fatalf("OpRecords[0].OpID = %q, want %q (must be prepended)", got.Status.OpRecords[0].OpID, "new")
	}
	if got.Status.OpRecords[1].OpID != "old" {
		t.Fatalf("OpRecords[1].OpID = %q, want %q", got.Status.OpRecords[1].OpID, "old")
	}
}

// TestAMStore_PatchStatus_PayloadClearVsMerge pins the Payload
// semantics documented on ports.StatusPatch:
//
//   - nil            → leave the existing map untouched
//   - empty non-nil  → clear the map entirely
//   - non-empty      → merge key-by-key (existing keys preserved)
//   - delete keys    → remove only named keys after the merge
//
// State-leaving handlers (installing.finishInstall,
// downloading.handoffToInstalling) rely on the middle rule to
// prevent stale phase Payload from bleeding into the next state.
// A regression here silently breaks every such handler.
func TestAMStore_PatchStatus_PayloadClearVsMerge(t *testing.T) {
	s := fake.NewAMStore()
	am := newAM("app1", "Installing")
	am.Status.Payload = map[string]string{"phase": "helm-install", "keep": "yes"}
	rv := s.Add(am)

	t.Run("nil leaves map untouched", func(t *testing.T) {
		if err := s.PatchStatus(context.Background(), "app1", rv, ports.StatusPatch{
			Message: "unrelated update",
		}); err != nil {
			t.Fatalf("PatchStatus: %v", err)
		}
		got, _ := s.Get(context.Background(), "app1")
		if len(got.Status.Payload) != 2 || got.Status.Payload["phase"] != "helm-install" {
			t.Fatalf("Payload = %+v, want unchanged", got.Status.Payload)
		}
	})

	t.Run("non-empty merges key-by-key", func(t *testing.T) {
		got, _ := s.Get(context.Background(), "app1")
		if err := s.PatchStatus(context.Background(), "app1", got.ResourceVersion, ports.StatusPatch{
			Payload: map[string]string{"phase": "done"},
		}); err != nil {
			t.Fatalf("PatchStatus: %v", err)
		}
		got, _ = s.Get(context.Background(), "app1")
		if got.Status.Payload["phase"] != "done" {
			t.Fatalf("phase = %q, want done", got.Status.Payload["phase"])
		}
		if got.Status.Payload["keep"] != "yes" {
			t.Fatalf("keep = %q, want yes (merge must preserve untouched keys)", got.Status.Payload["keep"])
		}
	})

	t.Run("delete keys removes only named entries", func(t *testing.T) {
		got, _ := s.Get(context.Background(), "app1")
		if err := s.PatchStatus(context.Background(), "app1", got.ResourceVersion, ports.StatusPatch{
			Payload:           map[string]string{"cleanup": "done"},
			DeletePayloadKeys: []string{"phase"},
		}); err != nil {
			t.Fatalf("PatchStatus: %v", err)
		}
		got, _ = s.Get(context.Background(), "app1")
		if _, ok := got.Status.Payload["phase"]; ok {
			t.Fatalf("phase key retained: %+v", got.Status.Payload)
		}
		if got.Status.Payload["keep"] != "yes" || got.Status.Payload["cleanup"] != "done" {
			t.Fatalf("Payload = %+v, want unrelated + merged keys retained", got.Status.Payload)
		}
	})

	t.Run("empty non-nil clears the entire map", func(t *testing.T) {
		got, _ := s.Get(context.Background(), "app1")
		if err := s.PatchStatus(context.Background(), "app1", got.ResourceVersion, ports.StatusPatch{
			Payload: map[string]string{},
		}); err != nil {
			t.Fatalf("PatchStatus: %v", err)
		}
		got, _ = s.Get(context.Background(), "app1")
		if len(got.Status.Payload) != 0 {
			t.Fatalf("Payload = %+v, want empty (empty non-nil map must clear all keys)", got.Status.Payload)
		}
	})
}

// TestAMStore_List pins List returning independent copies.
func TestAMStore_List(t *testing.T) {
	s := fake.NewAMStore()
	s.Add(newAM("a", "Pending"))
	s.Add(newAM("b", "Running"))

	got, err := s.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("List len = %d, want 2", len(got))
	}
}

// TestAMStore_MutationsLog pins that the log records every write
// in call order, so tests can assert interaction shape.
func TestAMStore_MutationsLog(t *testing.T) {
	s := fake.NewAMStore()
	rv := s.Add(newAM("app1", "Pending"))
	_ = s.PatchStatus(context.Background(), "app1", rv, ports.StatusPatch{State: "Downloading"})

	got := s.Mutations()
	if len(got) != 2 {
		t.Fatalf("Mutations len = %d, want 2", len(got))
	}
	if got[0].Kind != "AddAM" || got[1].Kind != "PatchStatus" {
		t.Fatalf("kinds = [%q %q], want [AddAM PatchStatus]", got[0].Kind, got[1].Kind)
	}
}
