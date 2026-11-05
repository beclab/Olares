// Package mwops implements ports.MiddlewareOps by fanning two
// unrelated code paths into a single port:
//
//  1. KubeBlocks OpsRequest — Start/Stop wrap
//     pkg/kubeblocks.NewOperation(...).Start() / .Stop(), which
//     creates a kbopv1alpha1.OpsRequest CR the KubeBlocks operator
//     watches. Used by suspending_app / resuming_app for
//     Postgres / Redis / MongoDB / Zinc / Nats middleware apps.
//
//  2. Legacy pre-KubeBlocks MongoDB — UninstallLegacyMongo wraps
//     pkg/middlewareinstaller.Uninstall (helm uninstall) PLUS an
//     explicit "delete Helm release Secret" fixup for the mongodb
//     v1 release, matching the legacy oldMongodbUninstall shape
//     verbatim (pkg/appstate/types.go:392-418).
//
// # What this replaces
//
// The port folds three legacy call sites into one interface:
//
//   - pkg/appstate/suspending_app.go:226-227
//     deps.NewMiddlewareOp(Stop).Stop() → Ops.Stop
//   - pkg/appstate/resuming_app.go:230-231
//     deps.NewMiddlewareOp(Start).Start() → Ops.Start
//   - pkg/appstate/suspend_failed_app.go:68-69
//     deps.NewMiddlewareOp(Stop).Stop() (retry) → Ops.Stop
//   - pkg/appstate/uninstalling_app.go:179-181
//     baseStatefulApp.oldMongodbUninstall → Ops.UninstallLegacyMongo
//   - pkg/appstate/types.go:216-218 (same, from stopping path)
//
// # Design of the seams
//
// Two seams are exposed via functional Options at construction:
//
//  1. WithKubeBlocksFactory overrides how a Start/Stop op is built.
//     Production uses defaultKBFactory which returns
//     pkg/kubeblocks.NewOperation. Tests inject a factory returning
//     a stub KubeBlocksOps so the adapter is unit-testable without
//     preseeding a KubeBlocks Cluster CR (which the real operator's
//     validate() method requires).
//
//  2. MongoUninstaller is a REQUIRED constructor parameter (not an
//     option) because it depends on *rest.Config which the adapter
//     itself does not need to hold. Production wires it to a closure
//     over kubeConfig calling middlewareinstaller.Uninstall; tests
//     pass a stub that records the call and returns a canned error.
//
// The client.Client the adapter holds serves TWO purposes:
//
//   - Passed to KubeBlocksOps factories so Start/Stop create their
//     OpsRequest via the same client (informer cache benefits).
//   - Used directly by UninstallLegacyMongo to Get + Delete the
//     Helm release Secret after the helm uninstall completes.
//
// # UninstallLegacyMongo error tolerance
//
// The legacy oldMongodbUninstall tolerates one specific helm error
// verbatim: "failed to delete release: mongodb". This corresponds
// to a race where the release name in the helm store no longer
// matches the release the caller wanted uninstalled; treating it
// as success is intentional (the mongodb release is either already
// gone or belongs to another controller loop that will retry).
//
// The tolerance is preserved byte-for-byte to keep the uninstall
// path idempotent under the same set of edge cases operators have
// encountered in the field.
//
// # When to wire this into main.go
//
// SAFE ANY TIME once the Runner has the corresponding CallWorkload
// dispatchers. Same reasoning as imgpull:
//
//   - Start/Stop write a KubeBlocks OpsRequest CR — the KubeBlocks
//     operator is idempotent on repeated OpsRequests (each has a
//     UUID-suffixed name), so double-dispatch during the
//     shadow-mode window is a benign extra OpsRequest that
//     reconciles to a no-op.
//   - UninstallLegacyMongo is safe under shadow-mode because the
//     legacy uninstalling_app.go WILL delete the app entirely
//     if this adapter has already removed the helm release —
//     the Helm client's IsInstalled check short-circuits.
//
// Wiring is nonetheless deferred to the step-3 cutover PR so all
// live-mode ports flip together under a single review lens.
package mwops
