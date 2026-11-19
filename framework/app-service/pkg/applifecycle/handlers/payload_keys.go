package handlers

import "github.com/beclab/Olares/framework/app-service/pkg/applifecycle/domain"

// Payload keys handlers use to persist per-AM scratch data through
// AM.Status.Payload. Centralised in this file so:
//
//   - reviewers can see the full surface at a glance
//   - name collisions across handlers are caught by string comparison
//     against the existing constants
//   - a follow-up migration step can grep the whole codebase for
//     these strings when it needs to know which fields survived the
//     port
//
// Naming convention: "handler-name.field-name", lowercase, dot-
// separated. This matches what the legacy pkg/appstate code writes
// into the same map today so an old AM in the cluster is readable by
// the new handlers during rollout.
const (
	// PayloadInstallingFailMessage / PayloadInstallingFailReason
	// stash the original install-failure diagnostic while the
	// cleanup CallWorkload runs. Landed on InstallFailed after
	// cleanup finishes (or times out).
	PayloadInstallingFailMessage = "installing.fail-message"
	PayloadInstallingFailReason  = "installing.fail-reason"

	// PayloadInstallFailedCleanup records whether
	// cleanupAfterInstallFailure has already run for this
	// InstallFailed AM. "done" means OnInstallFailedTick must
	// not re-dispatch; "cleanup" means a retry is in flight;
	// "retry" means the last attempt failed and the next one waits
	// for PayloadInstallFailedCleanupRetryAt.
	PayloadInstallFailedCleanup = "install-failed.cleanup"

	// PayloadInstallFailedCleanupRetryAt is the RFC3339 time after
	// which a "retry" cleanup phase may dispatch again.
	PayloadInstallFailedCleanupRetryAt = "install-failed.cleanup-retry-at"

	// PayloadInstallingPhase records which sub-phase of the
	// Installing state is currently in flight. Read by
	// installing.OnWorkloadReady / installing.OnWorkloadFailed to
	// disambiguate "which CallWorkload just completed", since a
	// bare WorkloadReady action carries no Kind hint.
	//
	// Values are one of the PhaseInstalling* constants below.
	// Empty string is the sentinel meaning "installing was just
	// entered from Downloading; no sub-phase started yet".
	PayloadInstallingPhase = "installing.phase"

	// PayloadPendingQueuedAt is the stable FIFO key for the
	// download slot. Written once when an AM first enters Pending
	// and never rewritten by waitForSlot. Runner.computeClusterStats
	// orders Pending AMs by this timestamp.
	PayloadPendingQueuedAt = domain.PayloadPendingQueuedAt

	// PayloadUpgradeDownloadingPhase records the active
	// UpgradeDownloading sub-phase. prepare refreshes the chart so
	// the following pull sees the new image refs; pull then holds
	// the cluster-wide download slot until helm starts.
	PayloadUpgradeDownloadingPhase = "upgrade-downloading.phase"

	// PayloadUpgradeDownloadingCancelingPhase records the
	// image-cancel sub-phase of UpgradeDownloadingCanceling.
	PayloadUpgradeDownloadingCancelingPhase = "upgrade-downloading-canceling.phase"

	// PayloadDownloadingPhase records the active Downloading
	// sub-phase. A process-local workload claim suppresses duplicate
	// dispatches; after restart the persisted phase is intentionally
	// re-dispatched because that claim no longer exists.
	//
	// Values are one of the PhaseDownloading* constants below.
	// Empty string means "Downloading was just entered from
	// Pending; no pull dispatched yet".
	PayloadDownloadingPhase = "downloading.phase"

	// PayloadInitializingPhase records the active Initializing
	// sub-phase. Same-process duplicate dispatch is guarded by the
	// Runner's workload claim; a fresh Runner recovers the persisted
	// phase by dispatching WaitLaunch again.
	PayloadInitializingPhase = "initializing.phase"

	// PayloadUninstallingPhase records which sub-phase of the
	// Uninstalling state is currently in flight. Two phases:
	// helm-uninstall then wait-for-namespace-deleted. See the
	// PhaseUninstalling* constants for individual semantics.
	PayloadUninstallingPhase = "uninstalling.phase"

	// PayloadStoppingPhase records the active Stopping sub-phase.
	// PhaseStoppingScale is redispatched after restart, while the
	// Runner's process-local claim prevents a duplicate Stop during
	// the same process lifetime.
	PayloadStoppingPhase = "stopping.phase"

	// PayloadResumingPhase records the active Resuming sub-phase.
	// PhaseResumingScale is redispatched after restart, while the
	// Runner's process-local claim prevents a same-process duplicate.
	PayloadResumingPhase = "resuming.phase"

	// PayloadUpgradingPhase records the active Upgrading sub-phase.
	// Persisted phases are recovered by a fresh Runner; a live
	// process-local workload claim suppresses duplicate dispatch.
	PayloadUpgradingPhase = "upgrading.phase"

	// PayloadApplyingEnvPhase records the active ApplyingEnv
	// sub-phase. A fresh Runner redispatches PhaseApplyingEnvHelm;
	// a live workload claim suppresses same-process duplicates.
	PayloadApplyingEnvPhase = "applyingenv.phase"

	// PayloadPendingCancelingPhase records which sub-phase of the
	// PendingCanceling cleanup pipeline is currently in flight.
	// PhaseCancelWaitNS is recovered after restart; a process-local
	// claim suppresses duplicate namespace waits while one is active.
	//
	// Naming note: the key mirrors the state name (not the source
	// state) so a cancel from PendingCanceling and one from
	// DownloadingCanceling never collide even if the AM churns
	// between them via *CancelFailed retries.
	PayloadPendingCancelingPhase = "pending-canceling.phase"

	// PayloadDownloadingCancelingPhase records the sub-phase of
	// the DownloadingCanceling pipeline. Two phases today
	// (PhaseCancelImagePull → PhaseCancelWaitNS) — image cancel
	// first so the ImageManager stops any active pull before the
	// namespace it's writing to disappears.
	PayloadDownloadingCancelingPhase = "downloading-canceling.phase"

	// PayloadInstallingCancelingPhase records the sub-phase of
	// the InstallingCanceling pipeline. Two phases today
	// (PhaseCancelHelmUninstall → PhaseCancelWaitNS) — full helm
	// tear-down first, then wait for the k8s GC to reap the ns.
	PayloadInstallingCancelingPhase = "installing-canceling.phase"

	// PayloadInstallingCancelingTerminal records which terminal the
	// InstallingCanceling pipeline must land on once its cleanup
	// finishes. Install failures stamp string(appv1.InstallFailed)
	// before handing off; a user cancel leaves the key unset and
	// lands on InstallingCanceled.
	//
	// Only the success branch reads it — a cleanup that itself fails
	// is an InstallingCancelFailed either way.
	PayloadInstallingCancelingTerminal = "installing-canceling.terminal"
)

// Sub-phase values for PayloadInstallingPhase. Kept as typed string
// aliases so handlers cannot accidentally pass an ad-hoc string that
// slips through review.
const (
	// PhaseInstallingExposePorts marks that a
	// CallWorkload{SetExposePorts} is in flight — the first
	// install step, matching installing_app.go's synchronous
	// SetExposePorts + Spec.Config patch before validation.
	PhaseInstallingExposePorts = "expose-ports"

	// PhaseInstallingValidate marks that a
	// CallWorkload{ValidateInstall} is in flight — the pre-helm
	// resource/quota check runs first and only proceeds to
	// PhaseInstallingHelmInstall on success.
	PhaseInstallingValidate = "validate"

	// PhaseInstallingCleanup marks that a
	// CallWorkload{CleanupInstallFailure} is in flight —
	// Uninstall + DeleteAllocationsForApp + WaitForNamespaceDeleted,
	// the cleanupAfterInstallFailure sequence, before the AM is
	// allowed to land in InstallFailed.
	//
	// No longer stamped by a fresh install failure: that tears down
	// through InstallingCanceling with
	// PayloadInstallingCancelingTerminal. The phase is still handled
	// so an AM a pre-handoff binary parked here still reaches
	// InstallFailed after a rollout.
	PhaseInstallingCleanup = "cleanup"

	// PhaseInstallingHelmInstall marks that a CallWorkload{Install}
	// is in flight. For a two-phase app the release helm creates
	// here renders every workload at replicas=0, so reaching this
	// phase's callback means "the release exists", not "the app is
	// up". Success advances to PhaseInstallingPostValidate (two-
	// phase) or PhaseInstallingScaleUp (single-phase); failure
	// lands in InstallFailed, or in Stopping when the adapter
	// classified the error as an unschedulable-pod failure.
	PhaseInstallingHelmInstall = "helm-install"

	// PhaseInstallingPostValidate marks that the SECOND
	// CallWorkload{ValidateInstall} is in flight — the two-phase
	// resource/quota check that only a deployed (but scaled-to-
	// zero) release can answer. Two-phase apps skip the pre-helm
	// check entirely and run this one instead; single-phase apps
	// never enter this phase.
	//
	// A rejection here does not fail the install: legacy converts
	// the operation into a stop and lands the AM in Stopping with
	// the release intact at replicas=0, so the user can free
	// capacity and resume.
	PhaseInstallingPostValidate = "post-validate"

	// PhaseInstallingScaleUp marks that a CallWorkload{Resume}
	// (helm Scale(-1)) is in flight, bringing every workload from
	// the install-time replicas=0 up to its manifest-declared
	// count. This is the step whose absence left two-phase apps
	// deployed-but-never-running.
	//
	// Dispatched unconditionally, including for single-phase apps:
	// HelmOps.Scale short-circuits to a no-op when the app declares
	// no workloadReplicas, and HelmOpsV2.Scale is a no-op outright,
	// so the phase costs one extra callback rather than needing a
	// branch that could drift from those adapters.
	PhaseInstallingScaleUp = "scale-up"

	// PhaseInstallingWaitStartUp marks that a
	// CallWorkload{WaitStartUp} is in flight — the readiness poll
	// over the pods the scale-up just created. Non-middleware apps
	// only; success advances to Initializing, which runs its own
	// WaitForLaunch before Running.
	PhaseInstallingWaitStartUp = "wait-startup"

	// PhaseInstallingWaitLaunch marks that a
	// CallWorkload{WaitLaunch} is in flight for a MIDDLEWARE app.
	// Middleware skips Initializing entirely: HelmOps.Install
	// already skips startup polling for it, and legacy
	// installing_app.go waits for launch inline and lands directly
	// in Running.
	PhaseInstallingWaitLaunch = "wait-launch"

	// PhaseDownloadingPull marks that a
	// CallWorkload{WorkloadPullImage} is in flight — the
	// ImageManager job has been created and the goroutine will
	// either advance to Installing (skeleton happy path, treats
	// "CreateJob returned nil" as pull complete) or land in
	// DownloadFailed. Full poll-until-terminal semantics are
	// tracked in ADAPTERS_TODO.md.
	PhaseDownloadingPull = "pull"

	// PhaseInitializingWait marks that a
	// CallWorkload{WorkloadWaitLaunch} is in flight — helm's
	// post-install readiness wait is running. Success advances
	// to Running; failure (real launch timeout, NOT a caller-
	// driven cancel — see legacy initializing_app.go:82-95 for
	// the disambiguation the Runner already performs) advances
	// to InitializingCanceling per the legacy failure path.
	PhaseInitializingWait = "wait"

	// PhaseUninstallingHelm marks that a
	// CallWorkload{WorkloadUninstall} is in flight. The Uninstall
	// port covers helm uninstall + PVC/permission/middleware/
	// cache cleanup + compute release + namespace-delete trigger
	// (see ports/workload.go Uninstall doc-comment). Success
	// advances the phase to PhaseUninstallingWaitNS.
	PhaseUninstallingHelm = "helm"

	// PhaseUninstallingWaitNS marks that a
	// CallWorkload{WorkloadWaitNamespaceGone} is in flight — the
	// long poll waiting for the k8s garbage collector to fully
	// tear down the AM's namespace after Uninstall triggered
	// deletion. Success advances to Uninstalled.
	//
	// Legacy uses PollImmediate(1s, 30min); the port's adapter
	// owns the timeout so the handler stays clock-free.
	PhaseUninstallingWaitNS = "wait-ns"

	// PhaseStoppingScale marks that a CallWorkload{WorkloadStop}
	// is in flight. The Stop port covers helm scale(0) + optional
	// v2 shared-server stop + middleware .Stop + compute release
	// (see ports/workload.go Stop doc-comment). Success advances
	// to Stopped.
	PhaseStoppingScale = "scale"

	// PhaseResumingScale marks that a CallWorkload{WorkloadResume}
	// is in flight. The Resume port covers helm scale(-1) +
	// middleware .Start. Success advances to Initializing (which
	// then dispatches WaitForLaunch to complete the resume) —
	// matches legacy ResumingApp.waitAsync line 171.
	PhaseResumingScale = "scale"

	// PhaseUpgradingPrepare marks that a
	// CallWorkload{PrepareUpgrade} is in flight — GetDeployedVersion
	// + downgrade guard + GetAppConfig / GetIndexAndDownloadChart.
	PhaseUpgradingPrepare = "prepare"

	// PhaseUpgradingPull marks that a
	// CallWorkload{WorkloadPullImage} is in flight during upgrade,
	// matching upgrading_app.go's ResolveImageRefs + Create +
	// PollDownloadProgress before helm upgrade.
	PhaseUpgradingPull = "pull"

	// PhaseUpgradingHelm marks that a CallWorkload{WorkloadUpgrade}
	// is in flight — ops.Upgrade() only. Version/chart/image work
	// lives in the preceding prepare and pull phases.
	PhaseUpgradingHelm = "helm"

	// PhaseUpgradingWaitStartUp marks that a
	// CallWorkload{WaitStartUp} is in flight after helm upgrade
	// (skipped when the upgrade was requested from Stopped).
	PhaseUpgradingWaitStartUp = "wait-startup"

	// PhaseInstallFailedCleanup marks an InstallFailed-state
	// retry of cleanupAfterInstallFailure.
	PhaseInstallFailedCleanup = "cleanup"

	// PhaseInstallFailedCleanupDone marks that install-failure
	// cleanup has already run for this AM.
	PhaseInstallFailedCleanupDone = "done"

	// PhaseInstallFailedCleanupRetry marks that the last cleanup
	// attempt failed (typically the namespace outlived the wait) and
	// another attempt is scheduled.
	PhaseInstallFailedCleanupRetry = "retry"

	// PhaseApplyingEnvHelm marks that a CallWorkload{WorkloadApplyEnv}
	// is in flight. The ApplyEnv port covers helm upgrade with
	// env-only overrides (helm_utils.SetValues path). Success
	// advances to Initializing.
	PhaseApplyingEnvHelm = "helm"

	// PhaseCancelImagePull marks that a
	// CallWorkload{WorkloadCancelImagePull} is in flight — the
	// DownloadingCanceling pipeline's first phase, stopping the
	// ImageManager pull before namespace cleanup. Success
	// advances to PhaseCancelWaitNS.
	//
	// Shared with cancel_pipeline.go (DownloadingCanceling) and
	// upgrade_download_cancel.go (UpgradeDownloadingCanceling).
	// PendingCanceling has no image pull to cancel (image work
	// starts in Downloading). InstallingCanceling and
	// UpgradingCanceling start after the image is already local.
	PhaseCancelImagePull = "cancel-pull"

	// PhaseCancelHelmUninstall marks that a
	// CallWorkload{WorkloadUninstall} is in flight — the
	// InstallingCanceling pipeline's first phase, tearing down
	// the partial helm release + compute allocation before
	// namespace cleanup. Success advances to PhaseCancelWaitNS.
	//
	// Shared with cancel_pipeline.go across the InstallingCanceling
	// state. Same WorkloadUninstall Kind as the Uninstalling
	// state's helm phase — the adapter's Uninstall implementation
	// is idempotent, so calling it on a partially-installed AM
	// is safe.
	PhaseCancelHelmUninstall = "helm-uninstall"

	// PhaseCancelWaitNS marks that a
	// CallWorkload{WorkloadWaitNamespaceGone} is in flight — the
	// final phase of every install-pipeline cancel (Pending,
	// Downloading, Installing). The adapter deletes the namespace
	// (if not IsProtectedNamespace) and polls until k8s GC has
	// reaped it. Success advances the AM to the state's matching
	// *Canceled terminal.
	//
	// Shared across three cancel_pipeline.go handlers because
	// they converge on the same "wait for the ns to be gone"
	// step regardless of what phase came before it.
	PhaseCancelWaitNS = "wait-ns"
)

// Annotation keys handlers read off AM.Annotations. Values mirror
// the legacy pkg/apiserver/api/types.go constants verbatim so an
// AM annotated by the legacy stack (or vice versa) is understood
// without a re-write.
//
// Kept in this file (not a separate one) because there's a small
// number of them; if the surface grows, split.
const (
	// AnnotPreUpgradeState is the annotation the upgrade
	// controller stamps onto an AM *before* transitioning it into
	// Upgrading, capturing the state the upgrade was requested
	// from. Read by the Upgrading handler's success branch to
	// pick between two landing states:
	//
	//   value == "stopped" → land in Stopped (upgrade-from-Stopped
	//                        fast-path; helm rendered replicas=0
	//                        so there's nothing to wait on).
	//   anything else      → land in Initializing (normal path;
	//                        Initializing dispatches WaitForLaunch).
	//
	// Mirrors pkg/apiserver/api/types.go AppPreUpgradeStateKey.
	AnnotPreUpgradeState = "bytetrade.io/pre-upgrade-state"

	// AnnotStopAll tells the stop path to tear down the whole
	// shared server backing the app rather than just this app's
	// workloads. The Installing handler stamps it when helm
	// reported that SERVER-SIDE pods could not be scheduled — the
	// app's own workloads are fine, the server they depend on is
	// the thing that has to come down.
	//
	// Mirrors pkg/apiserver/api/types.go AppStopAllKey.
	AnnotStopAll = "bytetrade.io/stop-all"

	// AnnotPendingPod is stamped by the pod-abnormal-suspend
	// controller when it wants a mid-resume abort. Resuming's
	// success branch reads it and hands off to Stopping instead
	// of Initializing. Mirrors
	// pkg/apiserver/api/types.go AppStopByControllerDuePendingPod.
	AnnotPendingPod = "bytetrade.io/pending-pod"
)
