package ports

import "context"

// InstallValidator runs the pre-install and post-helm resource /
// quota checks currently implemented in pkg/compute/validation.
//
// Legacy call sites:
//   - pkg/appstate/installing_app.go:157   deps.RunInstallValidation
//   - pkg/appstate/installing_app.go:148   PublishComputeInsufficientNotification
//
// The validator chain (cluster-capacity via kubesphere HTTP,
// user-quota, k8s-request, compute-mode, node-pressure, compute
// allocation) is stitched together inside the adapter; this port
// exposes only the two decision points the state machine actually
// branches on.
type InstallValidator interface {
	// ValidateInstall runs the pre-helm validators for the
	// installing state. HANDLER-SAFE READ (called by the Runner
	// before Engine.Step; the Decision is stashed on
	// domain.World.InstallDecision — field TBD when writing the
	// Installing handler).
	//
	// A OK==false Decision must be surfaced by the handler as a
	// PatchAM effect that transitions to InstallFailed, plus an
	// EmitEvent effect carrying the Reason/Resource metadata.
	ValidateInstall(ctx context.Context, in ValidationInput) (ValidationDecision, error)

	// ValidatePostHelm runs the second phase of validation for
	// two-phase installs (helm-first, then resource verification).
	// HANDLER-SAFE READ.
	ValidatePostHelm(ctx context.Context, in ValidationInput) (ValidationDecision, error)

	// NotifyInsufficientCompute is the specialised NATS notification
	// the legacy code emits when allocation fails (compute mode /
	// GPU pressure). RUNNER-ONLY WRITE.
	//
	// Kept as an explicit port method (rather than a generic
	// EventRecorder call) because the payload shape and NATS
	// subject differ from the app-lifecycle stream — matching the
	// legacy compute/notification.go behaviour.
	//
	// `appCfg` — concrete type: *pkg/appcfg.ApplicationConfig.
	NotifyInsufficientCompute(ctx context.Context, appCfg any, reason error) error
}

// ValidationInput mirrors pkg/compute/validation.Input at the port
// boundary. `AppCfg` is the concrete *pkg/appcfg.ApplicationConfig
// (cast by the adapter). `Op` uses a local string alias rather than
// pkg/compute/validation.Op so ports stays free of the compute
// package's transitive imports.
type ValidationInput struct {
	AppCfg any
	Op     ValidationOp
	Token  string
}

// ValidationOp identifies which validation chain the adapter
// selects. The set matches pkg/compute/validation.Op constants;
// mismatch is a bug caught in the adapter.
type ValidationOp string

const (
	// ValidateOpInstall selects the pre-install validation chain
	// (cluster capacity, user quota, k8s requests, compute mode,
	// node pressure, compute allocation).
	ValidateOpInstall ValidationOp = "install"

	// ValidateOpUpgrade selects the pre-upgrade chain (subset of
	// install — chart-version deltas and resource deltas only).
	ValidateOpUpgrade ValidationOp = "upgrade"

	// ValidateOpResume selects the resume chain (cluster pressure
	// only; the app's own quota is already accounted for by its
	// existing allocation).
	ValidateOpResume ValidationOp = "resume"
)

// ValidationDecision mirrors pkg/compute/validation.Decision at the
// port boundary. Field names and semantics are preserved so the
// adapter is a 1:1 mapping.
type ValidationDecision struct {
	// OK is true when every validator in the chain passed. When
	// false the remaining fields describe which validator rejected
	// and why.
	OK bool

	// Resource is the resource bucket that failed (as reported by
	// pkg/constants.ResourceType). Empty when OK==true.
	Resource string

	// Reason is the failure code (pkg/constants.ResourceConditionType).
	Reason string

	// Message is the human-readable failure description.
	Message string

	// Validator is the name of the specific validator that
	// short-circuited the chain, for observability. Empty when
	// OK==true.
	Validator string
}
