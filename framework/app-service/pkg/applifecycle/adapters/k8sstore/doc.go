// Package k8sstore implements ports.AMStore on top of a
// controller-runtime client.Client, with an operator-visible
// "shadow-mode" write gate.
//
// # What this replaces
//
// ADAPTERS_TODO.md maps three legacy write paths that historically
// drifted into a single AMStore port:
//
//   - pkg/appstate/types.go:71-120       — updateStatus (Patch)
//   - pkg/utils/app/app.go:167-229       — UpdateAppMgrStatus (Update)
//   - pkg/appstate/pending_app.go:*      — raw Update for the
//     download-concurrency gate
//
// This adapter is where those three collapse. Every method takes
// exactly the arguments ports.AMStore declares — no ambient client
// or informer state; the client is injected at construction, so
// tests can substitute testutil.NewFakeClient without envtest.
//
// # Shadow-mode
//
// The migration cannot atomically swap the legacy
// ApplicationManagerController for the new Runner: both write
// AM.Status via different code paths, and running both writers
// concurrently would produce last-writer-wins chaos.
//
// Shadow-mode lets us run the Runner in production with REAL reads
// (Get / List / GetApplication go straight through) but SUPPRESSED
// writes (PatchStatus / PatchSpecAndAnnotations / UpdateFull are
// no-ops that emit a structured log describing the diff they would
// have applied). This gives us:
//
//   - Confidence that the Runner's world materialization works
//     against live cluster state.
//   - Log-diff observability against the legacy controller's
//     actual writes — any divergence surfaces as a "shadow said X,
//     legacy did Y" mismatch we can triage before flipping the
//     switch.
//   - Zero production impact — the legacy controller remains the
//     sole writer, matching today's behaviour byte-for-byte.
//
// Shadow-mode is toggled per Store instance at construction
// (WithShadowMode) and can be flipped at runtime via SetShadowMode
// without recreating the store — useful for staged rollout where
// an operator flips the switch after a soak period.
//
// # Why the field-level merge is duplicated here
//
// The status-field merge logic (empty scalar = "don't touch", nil
// map = "don't touch", empty non-nil map = "clear") is the CONTRACT
// documented on ports.StatusPatch. fake/amstore.go has one copy,
// this package has another. The alternative — a shared helper in
// ports or a third package — would create a compile-time
// dependency this package doesn't need. Both copies are locked
// down by the same set of ports-contract tests once envtest
// coverage lands; drift risk is low and the local copy keeps the
// adapter self-contained.
package k8sstore
