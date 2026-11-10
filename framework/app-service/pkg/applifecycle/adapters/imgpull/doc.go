// Package imgpull implements ports.ImagePuller on top of the
// existing pkg/images.ImageManagerClient plus a fresh pure
// Progress() reader.
//
// # Position in the flow
//
//	handler (Downloading, Upgrading, *Canceling)
//	   │  returns domain.CallWorkload{Kind: WorkloadPullImage / …}
//	   ▼
//	Runner effects dispatcher (dispatchImagePull / dispatchCancelImagePull)
//	   │  calls port method
//	   ▼
//	imgpull.Puller
//	   ├─ ResolveRefs      ── RefResolver seam ── (helm dry-run in main.go)
//	   ├─ CreateJob        ── images.ImageManagerClient.Create        ── ImageManager CR
//	   ├─ CancelJob        ── images.ImageManagerClient.UpdateStatus   ── ImageManager CR
//	   ├─ UpdateJobStatus  ── images.ImageManagerClient.UpdateStatus   ── ImageManager CR
//	   └─ Progress         ── client.Get + images.DownloadPercent     ── ImageManager CR
//	                                    │
//	                                    ▼
//	                          ImageManagerController (separate reconciler)
//	                                    │
//	                                    ▼
//	                                containerd
//
// The Runner never talks to containerd; it only manipulates the
// ImageManager CR (an intent record) and reads its Status. The
// existing ImageManagerController turns intents into pulls.
//
// # What this package composes vs re-implements
//
//   - CreateJob / CancelJob / UpdateJobStatus wrap
//     `pkg/images.NewImageManager(client)` verbatim. That client
//     already has the "delete stale IM CR, snapshot node list,
//     apply schedulable filter, then Create" logic proven in
//     production. Duplicating it here would be pointless drift risk.
//
//   - Progress is a snapshot reader (no polling, no event
//     publishing) because the legacy `PollDownloadProgress` is a
//     blocking ticker + NATS publisher baked into one function.
//     The numeric percent is images.DownloadPercent — the same
//     function GetDownloadProgress / PollDownloadProgress use —
//     so the Runner and the UI cannot disagree on "how far".
//     The Runner owns the poll cadence; EventRecorder emits events.
//
// # RefResolver seam
//
// ResolveRefs needs a full helm dry-run to enumerate the image refs
// the AM wants to pull — that means importing pkg/appinstaller
// (heavy) and pkg/appstate.GetRefsForImageManager (which lives in
// the legacy state machine we're actively deleting). Neither is a
// natural dependency for a small adapter package.
//
// Instead the adapter takes a RefResolver func at construction
// time. main.go wires the production resolver (which is free to
// import whichever legacy helper it needs); tests inject a stub.
// This keeps the adapter unit-testable without any legacy pkg
// import and gives us a clean seam to eventually delete when
// GetRefsForImageManager moves to a neutral home during cleanup
// (step 5).
//
// # When to wire this into main.go
//
// SAFE TO WIRE ANY TIME. Unlike the natsrecorder decision:
//
//   - CreateJob is idempotent at the CR level (legacy already
//     deletes stale IM CR before creating a fresh one) so a Runner
//     that races the legacy DownloadingApp writer produces at most
//     one extra Get+Delete+Create round trip — no user-visible
//     effect.
//   - UpdateJobStatus / CancelJob are Patch operations that only
//     touch IM.Status, not AM.Status — no race with the legacy
//     ApplicationManagerController that k8sstore shadow-mode is
//     protecting.
//   - Progress is a pure read.
//   - The ImageManagerController (which does the actual pulls) is
//     unaffected by whether the IM CR was created by the legacy
//     DownloadingApp handler or by the Runner via this adapter —
//     it just watches the CR and pulls.
//
// Runner.dispatchImagePull calls CreateJob and then blocks in
// awaitImagePull, polling Progress until Percent>=100 or
// Failed, so the WorkloadReady callback means "the bytes are on
// disk". Unpack into containerd (the IM CR's later `completed`
// flip) is left to Installing. Because Progress treats a missing
// IM CR as non-terminal, the window between CreateJob and the
// informer observing the new CR is waited through rather than
// mistaken for completion.
//
// computeProgress maps the DownloadingCanceled IM state to
// Failed=true. That is what stops the poll loop when a cancel lands
// mid-pull; without it the loop would spin on a CR that will never
// reach 100%.
package imgpull
