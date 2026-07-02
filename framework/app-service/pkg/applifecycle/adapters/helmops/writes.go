package helmops

import (
	"context"
	"errors"

	appv1 "github.com/beclab/api/api/app.bytetrade.io/v1alpha1"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
	"github.com/beclab/Olares/framework/app-service/pkg/constants"
	"github.com/beclab/Olares/framework/app-service/pkg/errcode"
)

// The methods in this file share one shape:
//
//   1. buildHelmOps to resolve config + factory a HelmOpsInterface.
//   2. Invoke the matching HelmOpsInterface method.
//   3. Fold any error into WorkloadResult{Err: ...}.
//
// LandedReason is left empty; per runner/effects.go:270-280, the
// Runner will backfill from the handler's OnFailure.Reason when the
// workload does not provide a specific reason. Helm errors are
// opaque strings anyway, so a workload-side classification would be
// mostly guessing.
//
// LandedState is populated in exactly two cases (Upgrade / Resume)
// where the legacy state machine had a documented
// upgrade-of-stopped or post-install-stopped branch. Those cases
// are commented at the call site.

// Install implements ports.WorkloadOps. Legacy: installing_app.go:185.
//
// A helm install that fails because the app's pods cannot be
// scheduled is NOT an install failure: legacy routes it to Stopping,
// leaving the release deployed so the user can free up capacity and
// resume, rather than to InstallFailed which tears the release down.
// classifyInstallErr encodes that routing on the WorkloadResult
// because the errcode sentinel it keys off is wrapped inside the
// helm error and invisible above this layer.
func (o *Ops) Install(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	ops, err := o.buildHelmOps(ctx, req)
	if err != nil {
		return ports.WorkloadResult{Err: err}
	}
	if err := ops.Install(); err != nil {
		return classifyInstallErr(err)
	}
	return ports.WorkloadResult{}
}

// classifyInstallErr maps a helm-install error onto the routing +
// reason pair legacy installing_app.go wrote for it:
//
//	ErrServerSidePodPending → Stopping, stop-all flag set
//	ErrPodPending           → Stopping
//	anything else           → no routing hint; the handler's default
//	                          InstallFailed target applies
//
// The reason is the user-facing string that lands in
// AM.Status.Reason, and ErrHamiUnschedulable refines it for both
// pending flavours — a HAMI-unschedulable GPU request is actionable
// (pick a different compute mode) in a way that generic
// unschedulability is not.
func classifyInstallErr(err error) ports.WorkloadResult {
	serverSide := errors.Is(err, errcode.ErrServerSidePodPending)
	if !serverSide && !errors.Is(err, errcode.ErrPodPending) {
		return ports.WorkloadResult{Err: err}
	}
	reason := constants.AppUnschedulable
	if errors.Is(err, errcode.ErrHamiUnschedulable) {
		reason = constants.AppHamiSchedulable
	}
	return ports.WorkloadResult{
		LandedState:  string(appv1.Stopping),
		LandedReason: reason,
		Err:          err,
		StopAll:      serverSide,
	}
}

// Upgrade implements ports.WorkloadOps. Legacy: upgrading_app.go:294.
//
// LandedState is NOT set here — the legacy Upgrade handler consults
// the AnnotPreUpgradeState annotation stamped by the HTTP handler
// (see handler_installer_upgrade_v2.go: AppPreUpgradeStateKey) to
// decide between Initializing and Stopped after helm succeeds.
// Keeping the routing decision in the handler (which reads AM
// annotations) means the adapter stays pure-passthrough.
func (o *Ops) Upgrade(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	ops, err := o.buildHelmOps(ctx, req)
	if err != nil {
		return ports.WorkloadResult{Err: err}
	}
	if err := ops.Upgrade(); err != nil {
		return ports.WorkloadResult{Err: err}
	}
	return ports.WorkloadResult{}
}

// ApplyEnv implements ports.WorkloadOps. Legacy:
// applying_env_app.go:111. This is the env-only helm-upgrade — no
// chart re-download, just a values override.
func (o *Ops) ApplyEnv(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	ops, err := o.buildHelmOps(ctx, req)
	if err != nil {
		return ports.WorkloadResult{Err: err}
	}
	if err := ops.ApplyEnv(); err != nil {
		return ports.WorkloadResult{Err: err}
	}
	return ports.WorkloadResult{}
}

// Uninstall implements ports.WorkloadOps. Legacy:
// uninstalling_app.go:226 (and installing_canceling_app.go +
// install_failure_cleanup.go via the same code path).
//
// The legacy call site tolerates one specific helm error:
// strings.Contains(err.Error(), "not found"), corresponding to a
// release that was already deleted (usually by a parallel cancel).
// That tolerance is preserved here so the WorkloadResult reports
// success in the same shape as legacy, avoiding a spurious
// UninstallFailed transition when the state machine and legacy
// controller both race to uninstall the same release.
func (o *Ops) Uninstall(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	ops, err := o.buildHelmOps(ctx, req)
	if err != nil {
		return ports.WorkloadResult{Err: err}
	}
	if err := ops.Uninstall(); err != nil {
		if isReleaseNotFoundErr(err) {
			return ports.WorkloadResult{}
		}
		return ports.WorkloadResult{Err: err}
	}
	return ports.WorkloadResult{}
}

// UninstallAll implements ports.WorkloadOps. Legacy:
// uninstalling_app.go:190. Same behaviour as Uninstall but also
// tears down the shared-namespace resources (deployment labels,
// namespace deletion) — the shared-NS logic is baked into the
// legacy UninstallAll method, so no adapter-side handling needed.
//
// Same "not found" tolerance as Uninstall.
func (o *Ops) UninstallAll(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	ops, err := o.buildHelmOps(ctx, req)
	if err != nil {
		return ports.WorkloadResult{Err: err}
	}
	if err := ops.UninstallAll(); err != nil {
		if isReleaseNotFoundErr(err) {
			return ports.WorkloadResult{}
		}
		return ports.WorkloadResult{Err: err}
	}
	return ports.WorkloadResult{}
}

// Stop implements ports.WorkloadOps. Legacy:
// suspending_app.go:scaleOrPatchSuspend — helm Scale(0) when the
// app declares workloadReplicas, otherwise the original
// SuspendV1/V2 patch helpers. Retryable webhook errors poll the
// same way the legacy exec loop did (5s / 30m).
//
// Wait-for-pods / middleware / compute / stop-all cascade stay
// in the Runner so those ports remain independently fakeable.
func (o *Ops) Stop(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	stopServer := stopAllRequested(req.AM)
	appName := ""
	if req.AM != nil {
		appName = req.AM.Spec.AppName
	}
	if err := retryScaleOrPatch(ctx, appName, func() error {
		return o.scaleOrPatchStop(ctx, req, stopServer)
	}); err != nil {
		return ports.WorkloadResult{Err: err}
	}
	return ports.WorkloadResult{}
}

// Resume implements ports.WorkloadOps. Legacy:
// resuming_app.go:scaleOrPatchResume — helm Scale(-1) when the
// app declares workloadReplicas, otherwise the original
// ResumeV1/V2 patch helpers.
func (o *Ops) Resume(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	resumeServer := resumeAllRequested(req.AM)
	if err := o.scaleOrPatchResume(ctx, req, resumeServer); err != nil {
		return ports.WorkloadResult{Err: err}
	}
	return ports.WorkloadResult{}
}

// WaitForStartUp implements ports.WorkloadOps. Legacy:
// helm_ops_install.go:626, invoked from installing_app.go:322 and
// upgrading_app.go:311. Blocks (bounded by req context) until the
// app's server-side pods reach Ready or the pipeline detects an
// unrecoverable pod pending error.
//
// The underlying HelmOpsInterface.WaitForStartUp used to write
// appv1.Initializing to the AM in-line via apputils.UpdateAppMgrState.
// That write is gone (helm_ops_install.go:643) — it raced the
// WorkloadReady this call produces, flipping the AM out of Installing
// before the engine could route it. Landing Initializing belongs to
// the handler that owns the phase.
func (o *Ops) WaitForStartUp(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	ops, err := o.buildHelmOps(ctx, req)
	if err != nil {
		return ports.WorkloadResult{Err: err}
	}
	if _, err := ops.WaitForStartUp(); err != nil {
		// Same sentinels as helm install: ErrPodPending and
		// ErrServerSidePodPending are unschedulable and route to
		// Stopping. ErrHamiUnschedulable only refines the reason.
		// Any other startup error stays unclassified so the
		// handler can cancel the install.
		return classifyInstallErr(err)
	}
	return ports.WorkloadResult{}
}

// WaitForLaunch implements ports.WorkloadOps. Legacy:
// initializing_app.go:82. Middleware / init-container readiness
// poll. Returns (ready bool, err error); the bool is currently
// discarded by both legacy call sites and by this adapter — a
// success is any non-error return.
//
// The context passed here is honored by the underlying poll.
// WaitForStartUp returns the same context error when its poll is
// canceled, so an installing deadline is a timeout rather than a
// successful startup.
func (o *Ops) WaitForLaunch(ctx context.Context, req ports.WorkloadRequest) ports.WorkloadResult {
	ops, err := o.buildHelmOps(ctx, req)
	if err != nil {
		return ports.WorkloadResult{Err: err}
	}
	if _, err := ops.WaitForLaunch(); err != nil {
		return ports.WorkloadResult{Err: err}
	}
	return ports.WorkloadResult{}
}

// isReleaseNotFoundErr matches the exact string check the legacy
// uninstall call sites perform. Kept as a helper so the two
// Uninstall variants agree on tolerance semantics.
//
// The legacy match is intentionally loose (strings.Contains) —
// helm's driver.ErrReleaseNotFound is wrapped into different error
// shapes depending on the release-store backend, and the human-
// readable "not found" substring is the lowest-common-denominator
// tolerated by every backend. Preserving the loose match keeps the
// adapter tolerance byte-identical to legacy.
func isReleaseNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	// Using a substring check rather than errors.Is because helm
	// wraps ErrReleaseNotFound into a plain string in several
	// places (driver-specific), and the legacy code (pkg/appstate/types.go:228)
	// used the same substring check.
	msg := err.Error()
	return containsFold(msg, "not found")
}

// containsFold is a case-insensitive strings.Contains without the
// ASCII bias of strings.EqualFold-per-position — it lowercases both
// and calls strings.Contains once. Micro-optimisation over a manual
// loop; called at most twice per uninstall (never in a hot path).
func containsFold(s, substr string) bool {
	return indexFold(s, substr) >= 0
}

// indexFold is strings.Index but case-insensitive. Local rather
// than pulling in strings.EqualFold in a loop.
func indexFold(s, substr string) int {
	// Fast path: exact match.
	for i := 0; i+len(substr) <= len(s); i++ {
		if equalASCIIFoldPrefix(s[i:], substr) {
			return i
		}
	}
	return -1
}

// equalASCIIFoldPrefix returns true when s[:len(prefix)] matches
// prefix case-insensitively over ASCII. Non-ASCII bytes compare
// byte-for-byte — helm errors are ASCII in practice.
func equalASCIIFoldPrefix(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		a, b := s[i], prefix[i]
		if 'A' <= a && a <= 'Z' {
			a += 'a' - 'A'
		}
		if 'A' <= b && b <= 'Z' {
			b += 'a' - 'A'
		}
		if a != b {
			return false
		}
	}
	return true
}
