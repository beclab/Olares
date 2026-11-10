package domain

import (
	"time"

	appsv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"
)

// Effect is the sum type for every side effect a handler can request. The
// Runner is the SOLE executor of effects; handlers only return them as
// data. See CONTRACT.md rule 1-3.
type Effect interface {
	isEffect()
}

// PatchAM requests that the Runner apply Status to the current
// ApplicationManager using an optimistic-lock Patch (MergeFromWithOptimisticLock).
//
// The Runner is responsible for:
//
//   - preserving unset fields from the current AM (Runner does a
//     field-level merge, not a wholesale replace)
//   - performing DeepEqual against the current AM and skipping the API
//     call when nothing changed (persist-if-changed)
//   - retrying on ResourceVersion conflict (the Reconcile is re-enqueued)
type PatchAM struct {
	Status appsv1.ApplicationManagerStatus

	// DeletePayloadKeys removes selected Status.Payload entries after
	// Status.Payload is merged. Use this when a state transition must
	// retain unrelated scratch data while dropping only the phase /
	// intent keys it owns; omitting a key from Status.Payload does not
	// delete it under the store's merge semantics.
	DeletePayloadKeys []string
}

// isEffect implements Effect.
func (PatchAM) isEffect() {}

// PatchSpec requests that the Runner write AM.Spec fields and/or
// object-meta annotations under optimistic concurrency, via
// AMStore.PatchSpecAndAnnotations.
//
// Handlers should almost never emit this — Spec is the apiserver's
// to write, Status is the state machine's. The legitimate cases are
// the ones the legacy code already had, and each needs reviewer
// sign-off:
//
//   - flipping Spec.OpType when the state machine converts an
//     install into a stop (installing_app.go's post-helm rejection
//     branch)
//   - the stop-all annotation that tells the stop path to tear down
//     a shared server rather than just this app's workloads
//   - persisting expose-port assignment back into Spec.Config
//
// ORDERING: a PatchSpec MUST precede any PatchAM in the same Result.
// Both writes carry an optimistic-lock precondition on the AM's
// ResourceVersion, and the spec write bumps it — the Runner re-reads
// the AM afterwards so the following status patch still matches, but
// it does not (and cannot) rewind a status write that already
// happened. Emitting them in the other order makes the spec write
// fail its precondition and surface as a spurious conflict requeue.
type PatchSpec struct {
	// OpType and Config mirror ports.SpecPatch: empty means "leave
	// the existing value alone".
	OpType OpType
	Config string

	// SetAnnotations / DeleteAnnotations mirror
	// ports.AnnotationPatch. Set wins on key collision.
	SetAnnotations    map[string]string
	DeleteAnnotations []string
}

// isEffect implements Effect.
func (PatchSpec) isEffect() {}

// EmitEvent requests that the Runner publish a domain event via the
// injected EventRecorder port (NATS in production, in-memory
// FakeRecorder in tests). Handlers MUST NOT publish directly. See
// CONTRACT.md rule 1.
type EmitEvent struct {
	// Type mirrors client-go/record.EventRecorder's event type ("Normal"
	// / "Warning").
	Type string

	// Reason is the machine-readable event reason.
	Reason string

	// Message is the human-readable event message.
	Message string

	// Payload carries additional structured fields to be attached to the
	// NATS message. Never nil for effects with structured fields.
	Payload map[string]string
}

// isEffect implements Effect.
func (EmitEvent) isEffect() {}

// ScheduleTick requests that the Runner requeue the current AM for
// Reconcile after the given duration. This is the ONLY way to express a
// "wait N seconds and then re-check" semantic; handlers MUST NOT call
// time.Sleep or wait.PollImmediate. See CONTRACT.md rule 3.
//
// The Runner converts ScheduleTick into reconcile.Result{RequeueAfter: After}
// so controller-runtime's workqueue handles the actual scheduling.
type ScheduleTick struct {
	After time.Duration
}

// isEffect implements Effect.
func (ScheduleTick) isEffect() {}

// CallWorkload requests that the Runner execute a long-running side
// effect on an injected workload port (helm ops, image puller, namespace
// deleter, etc.) in a Runner-owned goroutine. Handlers MUST NOT spawn
// goroutines themselves. See CONTRACT.md rule 2.
//
// When the workload terminates the Runner feeds OnSuccess or OnFailure
// back into Engine.Step, closing the loop through the state machine. See
// CONTRACT.md rule 5.
type CallWorkload struct {
	// Kind identifies the workload op (e.g. "helm.Install", "image.Pull",
	// "namespace.WaitDeleted"). The Runner routes based on Kind.
	Kind WorkloadKind

	// Params carries the workload's input parameters. Runner passes it
	// verbatim to the port adapter.
	Params map[string]string

	// OnSuccess is the Action fed back into the state machine when the
	// workload reaches its success terminal. Typically WorkloadReady{}.
	OnSuccess Action

	// OnFailure is the Action fed back on failure. Typically
	// WorkloadFailed{Reason, Message}.
	OnFailure Action

	// Preempts marks this workload as an interruption of whatever is
	// already running for the AM, rather than the usual "one workload
	// at a time" step. The Runner abandons the in-flight goroutine's
	// claim — and discards its eventual result — before dispatching
	// this one.
	//
	// Cancel pipelines set it on the first CallWorkload that has to
	// run while the source state's work is still in flight:
	// CancelImagePull (download / upgrade), Uninstall (install),
	// and Stop (Initializing / Resuming / ApplyingEnv handoff).
	Preempts bool
}

// isEffect implements Effect.
func (CallWorkload) isEffect() {}

// CancelInFlight cancels the context of the AM's in-flight workload
// and keeps the claim. The goroutine still delivers OnSuccess or
// OnFailure. Resume cancel uses this so the in-flight helm upgrade
// can return before Stopping starts another one.
type CancelInFlight struct{}

// isEffect implements Effect.
func (CancelInFlight) isEffect() {}

// AppendOpRecord requests that the Runner prepend a new OpRecord to the
// AM's Status.OpRecords slice (matches makeRecord + updateStatus behavior
// in the legacy pkg/appstate). The Runner merges the record into the
// next PatchAM effect for the same reconcile so the write is atomic.
type AppendOpRecord struct {
	Record appsv1.OpRecord
}

// isEffect implements Effect.
func (AppendOpRecord) isEffect() {}

// WorkloadKind identifies a long-running side effect handled by an
// adapter port. Values are declared here rather than as free-form strings
// so the compiler catches typos in handler code.
type WorkloadKind string

// Kinds enumerated here match the legacy operations in
// pkg/appstate/*_app.go so table_default.go transitions can be filled in
// incrementally without changing this enum.
const (
	WorkloadInstall               WorkloadKind = "helm.Install"
	WorkloadUpgrade               WorkloadKind = "helm.Upgrade"
	WorkloadUninstall             WorkloadKind = "helm.Uninstall"
	WorkloadStop                  WorkloadKind = "helm.Stop"
	WorkloadResume                WorkloadKind = "helm.Resume"
	WorkloadApplyEnv              WorkloadKind = "helm.ApplyEnv"
	WorkloadPullImage             WorkloadKind = "image.Pull"
	WorkloadCancelImagePull       WorkloadKind = "image.CancelPull"
	WorkloadWaitStartUp           WorkloadKind = "workload.WaitStartUp"
	WorkloadWaitLaunch            WorkloadKind = "workload.WaitLaunch"
	WorkloadWaitNamespaceGone     WorkloadKind = "namespace.WaitDeleted"
	WorkloadValidateInstall       WorkloadKind = "validate.Install"
	WorkloadSetExposePorts        WorkloadKind = "app.SetExposePorts"
	WorkloadCleanupInstallFailure WorkloadKind = "install.CleanupFailure"
	WorkloadPrepareUpgrade        WorkloadKind = "upgrade.Prepare"
	WorkloadDeleteAllocations     WorkloadKind = "compute.DeleteAllocations"
)
