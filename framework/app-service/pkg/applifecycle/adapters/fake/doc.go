// Package fake provides in-memory implementations of the Port
// interfaces declared in pkg/applifecycle/ports. They exist to make
// handler and Runner unit tests trivially fast: no envtest, no live
// cluster, no NATS server, no wall clock, no filesystem.
//
// # Design principles
//
//  1. **Zero external I/O.** Every fake is a pure Go struct with a
//     mutex-guarded internal map / slice. Any test that fails on a
//     fake fails deterministically.
//
//  2. **Programmable via public fields.** Rather than hiding state
//     behind method-chained builders, each fake exposes hook fields
//     (e.g. `InstallFunc func(...) WorkloadResult`) that tests can
//     assign to steer behaviour per case. Defaults are no-op success.
//
//  3. **Observable via getter helpers.** Every fake records every
//     call it received (`Calls()`, `Events()`, `PatchStatuses()`) so
//     tests can assert the interaction shape, not just the return
//     value.
//
//  4. **Never spawn a goroutine.** Timer/ticker fakes are advanced
//     synchronously via `Advance(d)`. This keeps tests race-free
//     without a testing.Short-style skip.
//
//  5. **Thread-safe.** Fakes are safe under concurrent use so the
//     Runner (which invokes ports from multiple goroutines when
//     dispatching CallWorkload effects) can be tested end-to-end.
//
// # Compile-time interface conformance
//
// Each fake file ends with a `var _ ports.X = (*FakeX)(nil)` guard
// so a signature drift between the interface and the fake is caught
// at build time, not at test time.
//
// # File layout
//
//	clock.go      — FakeClock: manually-advanced time source
//	recorder.go   — FakeEventRecorder: in-memory event slice
//	amstore.go    — FakeAMStore: in-memory AM/Application store with
//	                ResourceVersion optimistic concurrency
//	workload.go   — FakeWorkloadOps: no-op successful helm ops with
//	                per-method hooks
//	image.go      — FakeImagePuller: programmable ImageProgress
//	validator.go  — FakeInstallValidator: default OK, hook-overridable
//	middleware.go — FakeMiddlewareOps: default success, hook-overridable
//	identity.go   — FakeIdentity: default { admin: false, owner: "" }
//	compute.go    — FakeComputeOps: no-op success + call log
//	market.go     — FakeMarketChartRepo: returns pre-loaded configs
package fake
