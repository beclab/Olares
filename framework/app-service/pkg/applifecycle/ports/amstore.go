package ports

import (
	"context"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"
)

// AMStore is the sole write path for ApplicationManager CRs. It also
// exposes the AM/Application reads the Runner materializes into
// domain.World before invoking Engine.Step.
//
// This port unifies three legacy write paths that historically drifted:
//
//   - pkg/appstate/types.go:71-120 `updateStatus`         (controller-runtime Patch)
//   - pkg/utils/app/app.go:167-229 `UpdateAppMgrStatus`   (typed clientset Update)
//   - direct client.Update in pending_app.go / installing_app.go /
//     suspending_app.go                                    (raw Update)
//
// All three collapse into PatchStatus + PatchSpecAndAnnotations here;
// UpdateFull is retained as an escape hatch for the pending_app.go
// "download concurrency gate" write, which must overwrite the whole
// AM to be safe under the historical mutex regime (Runner will
// eventually replace it with a Patch-with-ResourceVersion once the
// gate moves to World.ClusterStats + Guard, per plan §1.5).
type AMStore interface {
	// Get returns the AM by name. HANDLER-SAFE READ (materialized into
	// domain.World.AM). Returns a NotFound error the caller can test
	// via apierrors.IsNotFound.
	Get(ctx context.Context, name string) (*appv1.ApplicationManager, error)

	// List enumerates all AMs. HANDLER-SAFE READ used exclusively by
	// the Runner to compute domain.World.ClusterStats (e.g.
	// DownloadingCount for the "at most one Downloading" guard).
	// Handlers never see this directly.
	List(ctx context.Context) ([]appv1.ApplicationManager, error)

	// GetApplication returns the paired Application CR by name.
	// HANDLER-SAFE READ (materialized into domain.World.AppExists).
	// A NotFound error means the Application has been GC'd and the AM
	// is orphaned — a signal several handlers branch on
	// (pkg/appstate/{pending,running,uninstalled}_app.go).
	GetApplication(ctx context.Context, name string) (*appv1.Application, error)

	// PatchStatus applies the given StatusPatch under optimistic
	// concurrency control using curResourceVersion as the precondition.
	// RUNNER-ONLY WRITE: this is where every `domain.PatchAM` effect
	// lands.
	//
	// Semantics:
	//   * empty StatusPatch fields are preserved (field-level merge,
	//     not a wholesale replace)
	//   * OpRecord != nil is prepended to Status.OpRecords atomically
	//     with the same Patch (matches makeRecord + updateStatus in
	//     the legacy code)
	//   * a ResourceVersion mismatch returns a Conflict error that the
	//     Runner surfaces as `reconcile.Result{Requeue: true}` — the
	//     next Reconcile re-materializes World and re-runs the handler
	//   * a DeepEqual-with-current short-circuit is enforced by the
	//     adapter, not the caller
	PatchStatus(ctx context.Context, name, curResourceVersion string, patch StatusPatch) error

	// PatchSpecAndAnnotations updates spec fields and object-meta
	// annotations under optimistic concurrency. RUNNER-ONLY WRITE.
	//
	// Split from PatchStatus because K8s treats status as a
	// sub-resource and the two writes take different code paths in
	// the API server. Handlers should almost never touch spec — the
	// two legitimate cases are:
	//   * installing_app.go:76 persisting expose-port assignment
	//   * suspending_app.go stop-all annotation flip
	// Any new spec write requires reviewer sign-off.
	PatchSpecAndAnnotations(ctx context.Context, name, curResourceVersion string,
		spec *SpecPatch, ann *AnnotationPatch) error

	// UpdateFull replaces the entire AM object using an Update RPC.
	// RUNNER-ONLY WRITE. Kept as an escape hatch for the
	// pending_app.go download-concurrency gate; new code should use
	// PatchStatus / PatchSpecAndAnnotations instead. Retirement is
	// tracked as part of the cutover step (plan §8.9).
	UpdateFull(ctx context.Context, am *appv1.ApplicationManager) error
}

// StatusPatch is the field-level merge diff PatchStatus applies to
// ApplicationManager.Status. Empty values mean "leave the existing
// field untouched"; explicit zero requires a pointer field (see
// StatusTime / UpdateTime).
type StatusPatch struct {
	State    domain.State
	Message  string
	Reason   string
	OpType   domain.OpType
	OpID     string
	Progress string

	// Payload merges into Status.Payload key-by-key. Set nil to leave
	// the map untouched; set an empty non-nil map to clear it.
	Payload map[string]string

	// DeletePayloadKeys removes selected keys after Payload is
	// merged. This is the targeted counterpart to an empty Payload
	// map, which clears every key.
	DeletePayloadKeys []string

	// OpRecord, when non-nil, is prepended to Status.OpRecords in the
	// same Patch. This preserves the atomicity guarantee the legacy
	// `makeRecord` helper achieved by holding the Status write in one
	// call.
	OpRecord *appv1.OpRecord

	// StatusTime and UpdateTime are pointers so a caller can set an
	// explicit zero (rare) versus "don't touch" (default nil).
	StatusTime *metav1.Time
	UpdateTime *metav1.Time
}

// SpecPatch carries the ApplicationManager.Spec fields Runner may
// write. Only the two known legitimate spec writes are exposed:
//
//   - Config: helm-values persistence for expose-port assignment
//   - OpType: stop-all / resume-all peer-op flipping
//
// Empty strings mean "don't touch".
type SpecPatch struct {
	Config string
	OpType domain.OpType
}

// AnnotationPatch specifies object-meta annotation edits. Set entries
// take precedence over Delete entries; both are applied in the same
// Patch RPC.
type AnnotationPatch struct {
	Set    map[string]string
	Delete []string
}
