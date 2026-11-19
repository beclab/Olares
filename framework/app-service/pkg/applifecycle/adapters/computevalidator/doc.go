// Package computevalidator implements ports.InstallValidator on top
// of the existing pkg/compute/validation validator framework plus
// pkg/compute.BuildInsufficientPayload for the NATS notification.
//
// # What this replaces
//
// Legacy call sites the port consolidates:
//
//   - pkg/appstate/installing_app.go:157 (deps.RunInstallValidation)
//     — currently wraps validation.Run(ctx, in,
//     InstallRuntimePressureValidators()...). Moved into this
//     adapter as ValidateInstall / ValidatePostHelm.
//
//   - pkg/appstate/installing_app.go:148
//     (compute.PublishComputeInsufficientNotification) — the
//     "compute resource insufficient" NATS notification. Moved
//     here as NotifyInsufficientCompute; the payload shape is
//     built by the shared pkg/compute.BuildInsufficientPayload
//     helper so both this adapter and the legacy path publish
//     byte-identical NATS bodies.
//
// # Design of the seams
//
// Two seams are exposed via functional Options at construction:
//
//  1. WithChainFor overrides which pkg/compute/validation.Validator
//     chain runs for a given ports.ValidationOp. Production uses
//     the built-in mapping (install → InstallRuntimePressure,
//     resume → ResumePressure, upgrade → empty pass-through since
//     legacy upgrade does not use this framework). Tests inject a
//     stub chain of fake Validators to assert the mapping without
//     spinning up kubesphere HTTP + compute allocation.
//
//  2. NotifierFunc is the "publish this payload to this subject"
//     seam NotifyInsufficientCompute delegates to. Production wires
//     it to event.PublishToQueue (or, once EventRecorder is live,
//     to natsrecorder.Recorder.RecordNotification). Tests pass a
//     recording stub.
//
// # Pure-mapping methods
//
// ValidateInstall and ValidatePostHelm run the SAME validator chain
// (InstallRuntimePressureValidators) — the semantic difference is
// entirely on the caller side (pre-helm vs post-helm at
// replicas=0). Two port methods, one adapter body; the difference
// is the call site, not the check.
//
// # ValidationOp → legacy Op mapping
//
//	ports.ValidateOpInstall  → v1alpha1.InstallOp  + InstallRuntimePressureValidators
//	ports.ValidateOpResume   → v1alpha1.ResumeOp   + ResumePressureValidators
//	ports.ValidateOpUpgrade  → v1alpha1.UpgradeOp  + empty chain (returns OK immediately)
//	  ↑ legacy upgrade does not use validation.Run; the chain is
//	  intentionally empty so callers who mistakenly ask "validate
//	  the upgrade" get a benign pass rather than a mismatch error.
//
// # When to wire this into main.go
//
// SAFE ANY TIME. Same reasoning as imgpull:
//
//   - Validators are read-only (kubesphere HTTP + k8s API reads +
//     compute plan). The compute-allocation validator DOES write
//     an Allocation CR when it decides success, but the same write
//     path is what the LEGACY installing_app.go already uses on
//     the same trigger — swapping is a no-op re: race hazard
//     because at most one call site is active per AM at once
//     (the legacy DownloadingApp fires this today; the Runner will
//     fire it in the future, gated by k8sstore shadow-mode until
//     cutover).
//
//   - NotifyInsufficientCompute publishes to a separate subject
//     (os.notification, not os.application.*) so double-publishing
//     during migration would land on a downstream consumer that
//     already de-duplicates (system notification service).
//
// Wiring is nonetheless deferred to the step-3 cutover PR so all
// EventRecorder / notifier plumbing lands together with a single
// review lens.
package computevalidator
