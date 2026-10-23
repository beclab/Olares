package k8sstore

import (
	"context"
	"fmt"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
)

// PatchStatus implements ports.AMStore.
//
// Live mode:
//
//  1. Get the current AM (single API round-trip; cache-backed
//     when the client is manager-scoped).
//  2. Compare current.ResourceVersion to the caller's curRV. This
//     is the FAST-FAIL path — if the caller is obviously stale
//     (e.g. reconciling against an informer that hasn't seen the
//     latest update yet), we surface Conflict before doing any
//     write. The API server would catch it anyway via the
//     MergeFromWithOptimisticLock precondition below, but bailing
//     locally spares the round-trip.
//  3. DeepCopy current + applyStatusPatch. The DeepCopy is
//     mandatory so the MergeFrom base is the pre-patch object;
//     mutating current directly would produce an empty patch.
//  4. Patch with MergeFromWithOptimisticLock so the API server
//     rejects any last-minute race between our Get and our Patch.
//
// Shadow mode: skip steps 3-4 and log the intended diff at Info
// level (structured, key/value). The current Get + RV check still
// runs so operators can spot "shadow said X but curRV is stale"
// races even without an actual write.
//
// NotFound on the initial Get is returned unchanged so the Runner
// can classify it as a benign no-op (AM deleted between watch
// event and Reconcile).
func (s *Store) PatchStatus(ctx context.Context, name, curRV string, patch ports.StatusPatch) error {
	current := &appv1.ApplicationManager{}
	if err := s.c.Get(ctx, types.NamespacedName{Name: name}, current); err != nil {
		return err
	}
	if current.ResourceVersion != curRV {
		return newConflictAM(name, current.ResourceVersion, curRV)
	}

	if s.InShadowMode() {
		klog.InfoS("k8sstore shadow-mode: skipping PatchStatus",
			"name", name,
			"curRV", curRV,
			"state", string(patch.State),
			"opType", string(patch.OpType),
			"opID", patch.OpID,
			"reason", patch.Reason,
			"message", patch.Message,
			"progress", patch.Progress,
			"hasOpRecord", patch.OpRecord != nil,
			"payloadKeys", len(patch.Payload),
		)
		return nil
	}

	patched := current.DeepCopy()
	applyStatusPatch(&patched.Status, patch)
	return s.c.Patch(ctx, patched,
		client.MergeFromWithOptions(current, client.MergeFromWithOptimisticLock{}))
}

// PatchSpecAndAnnotations implements ports.AMStore. Same
// get-then-patch shape as PatchStatus; the merge semantic is
// documented on ports.SpecPatch + ports.AnnotationPatch (empty
// string = don't touch on Spec; Set/Delete apply on annotations
// with Set winning on key collision).
func (s *Store) PatchSpecAndAnnotations(ctx context.Context, name, curRV string, spec *ports.SpecPatch, ann *ports.AnnotationPatch) error {
	current := &appv1.ApplicationManager{}
	if err := s.c.Get(ctx, types.NamespacedName{Name: name}, current); err != nil {
		return err
	}
	if current.ResourceVersion != curRV {
		return newConflictAM(name, current.ResourceVersion, curRV)
	}

	if s.InShadowMode() {
		specDiff := "<nil>"
		if spec != nil {
			specDiff = fmt.Sprintf("Config-len=%d OpType=%q", len(spec.Config), string(spec.OpType))
		}
		annDiff := "<nil>"
		if ann != nil {
			annDiff = fmt.Sprintf("Set=%d Delete=%d", len(ann.Set), len(ann.Delete))
		}
		klog.InfoS("k8sstore shadow-mode: skipping PatchSpecAndAnnotations",
			"name", name,
			"curRV", curRV,
			"spec", specDiff,
			"annotations", annDiff,
		)
		return nil
	}

	patched := current.DeepCopy()
	applySpecPatch(&patched.Spec, spec)
	applyAnnotationPatch(patched, ann)
	return s.c.Patch(ctx, patched,
		client.MergeFromWithOptions(current, client.MergeFromWithOptimisticLock{}))
}

// UpdateFull implements ports.AMStore. Overwrites the entire AM.
// The port doc calls this an escape hatch for the pending_app.go
// download-concurrency gate; new code should prefer PatchStatus /
// PatchSpecAndAnnotations.
//
// UpdateFull skips the Get + local RV compare — the caller
// supplies the intended object (including ResourceVersion), so a
// bad RV is caught by the API server's Update precondition
// directly. In shadow-mode the Update is skipped and the intended
// object's key metadata is logged.
func (s *Store) UpdateFull(ctx context.Context, am *appv1.ApplicationManager) error {
	if am == nil {
		return fmt.Errorf("k8sstore.UpdateFull: am is nil")
	}
	if s.InShadowMode() {
		klog.InfoS("k8sstore shadow-mode: skipping UpdateFull",
			"name", am.Name,
			"resourceVersion", am.ResourceVersion,
			"state", string(am.Status.State),
			"opType", string(am.Status.OpType),
		)
		return nil
	}
	return s.c.Update(ctx, am)
}

// -----------------------------------------------------------------
// Field-level merge helpers. These are the CONTRACT documented on
// ports.StatusPatch / ports.SpecPatch / ports.AnnotationPatch;
// duplicated from fake/amstore.go so this adapter stays self-
// contained (see doc.go for the rationale).
// -----------------------------------------------------------------

// applyStatusPatch merges patch into status field-by-field. Empty
// scalar strings are preserved; nil pointer / nil map fields are
// preserved; an empty non-nil Payload clears the map.
func applyStatusPatch(status *appv1.ApplicationManagerStatus, patch ports.StatusPatch) {
	if patch.State != "" {
		status.State = patch.State
	}
	if patch.Message != "" {
		status.Message = patch.Message
	}
	if patch.Reason != "" {
		status.Reason = patch.Reason
	}
	if patch.OpType != "" {
		status.OpType = patch.OpType
	}
	if patch.OpID != "" {
		status.OpID = patch.OpID
	}
	if patch.Progress != "" {
		status.Progress = patch.Progress
	}
	if patch.Payload != nil {
		// Contract (ports/amstore.go StatusPatch.Payload):
		//   nil           → leave the current map untouched
		//   empty non-nil → clear the entire map (state-leaving
		//                   handlers use this to drop stale phase
		//                   keys before entering the next state)
		//   non-empty     → merge key-by-key
		if len(patch.Payload) == 0 {
			status.Payload = nil
		} else {
			if status.Payload == nil {
				status.Payload = make(map[string]string, len(patch.Payload))
			}
			for k, v := range patch.Payload {
				status.Payload[k] = v
			}
		}
	}
	for _, key := range patch.DeletePayloadKeys {
		delete(status.Payload, key)
	}
	if patch.OpRecord != nil {
		status.OpRecords = append([]appv1.OpRecord{*patch.OpRecord}, status.OpRecords...)
	}
	if patch.StatusTime != nil {
		status.StatusTime = patch.StatusTime
	}
	if patch.UpdateTime != nil {
		status.UpdateTime = patch.UpdateTime
	}
}

// applySpecPatch is the tiny counterpart for SpecPatch — only two
// fields are legitimate write targets per the port contract.
func applySpecPatch(spec *appv1.ApplicationManagerSpec, patch *ports.SpecPatch) {
	if patch == nil {
		return
	}
	if patch.Config != "" {
		spec.Config = patch.Config
	}
	if patch.OpType != "" {
		spec.OpType = patch.OpType
	}
}

// applyAnnotationPatch mutates am.Annotations per the patch.
// Set entries win over Delete on key collision, matching the
// documented contract on ports.AnnotationPatch.
func applyAnnotationPatch(am *appv1.ApplicationManager, patch *ports.AnnotationPatch) {
	if patch == nil {
		return
	}
	if am.Annotations == nil && len(patch.Set) > 0 {
		am.Annotations = make(map[string]string, len(patch.Set))
	}
	for _, k := range patch.Delete {
		delete(am.Annotations, k)
	}
	for k, v := range patch.Set {
		am.Annotations[k] = v
	}
}

// newConflictAM returns the same shape apierrors that a real API
// server would emit for a ResourceVersion mismatch, so callers
// that gate on apierrors.IsConflict work identically against this
// adapter, the fake AMStore, and a live cluster.
func newConflictAM(name, stored, given string) error {
	return apierrors.NewConflict(schema.GroupResource{
		Group: appv1.GroupVersion.Group, Resource: "applicationmanagers",
	}, name, fmt.Errorf("resource version conflict: stored=%s given=%s", stored, given))
}
