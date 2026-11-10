// Package ports declares the outbound interfaces of the application-
// lifecycle state machine. It is the "P" of Hexagonal Architecture
// (Ports & Adapters) applied to app-service.
//
// # Position in the layering
//
//	handlers (pure)  ─────returns─────►  domain.Result{Effects}
//	          │                                     │
//	          │ reads                               │ executed by
//	          ▼                                     ▼
//	     domain.World  ◄──materializes──   Runner  ──calls──►  ports.*
//	                          │                                   │
//	                          └─────────────────adapters/         │
//	                                            (real & fake)  ◄──┘
//
// Handlers NEVER call ports directly (that would violate CONTRACT.md
// rules 1-3). Handlers read from `domain.World` (pre-materialized by
// the Runner via port reads) and return `domain.Effect` values; the
// Runner is the sole caller of every port method.
//
// The Runner routes each `domain.Effect` to a specific port method
// (see the Effect→Port routing table in the Runner package). This
// package intentionally does not import the Runner; the Runner
// imports this package.
//
// # Handler-safe reads vs Runner-only writes
//
// Every port groups its methods into two facets:
//
//   - HANDLER-SAFE READS materialize into `domain.World.*` before
//     `Engine.Step`. They are cheap, deterministic queries the Runner
//     performs synchronously (informer cache reads, config lookups,
//     `Now()`). Handlers never wait on them because they've already
//     happened.
//
//   - RUNNER-ONLY WRITES are the long-running or mutating operations
//     handlers request via `domain.CallWorkload` / `domain.PatchAM` /
//     `domain.EmitEvent`. The Runner executes them AFTER the handler
//     returns, in Runner-owned goroutines with declared OnSuccess /
//     OnFailure callback Actions (CONTRACT.md rule 5).
//
// Each port's doc comments label individual methods with one of these
// tags so the boundary stays legible at read time.
//
// # Port inventory (10 ports)
//
// One file per port. Each file corresponds to a distinct outbound
// concern discovered during the legacy `pkg/appstate/` I/O audit:
//
//	amstore.go       — ApplicationManager + related CR reads/writes
//	workload.go      — helm install/upgrade/uninstall/stop/resume +
//	                   workload probes (Deployment/StatefulSet/Pod)
//	image.go         — ImageManager CR + image-pull lifecycle
//	validator.go     — pre-install / post-helm resource validation
//	recorder.go      — user-facing app events (replaces AppEventQueue
//	                   global and legacy NATS enqueue helpers)
//	clock.go         — Now() + tickers so TimeoutFired / ScheduleTick
//	                   are deterministic in tests
//	middleware.go    — KubeBlocks + legacy MongoDB middleware ops
//	identity.go      — admin / cluster-owner lookups
//	compute.go       — GPU/CPU allocation lifecycle
//	market.go        — chart repository HTTP fetch + local cache
//
// # Why interfaces (and not concrete types)
//
// Three reasons, in order of importance:
//
//  1. Testability. Handlers unit-tests inject `adapters/fake/*` (see
//     the sibling package) and assert on `Effects` returned + fake
//     ports invoked. No `envtest`, no live cluster, no NATS server.
//  2. Refactor safety. The Runner and every handler compile against
//     these interfaces only; a change to helm/K8s SDK version bumps
//     the adapters but not the domain or handlers.
//  3. Contract enforcement. The `handlerpurity` analyzer flags any
//     handler that reaches for a client.* / helm.* / NATS symbol
//     directly, forcing all outbound work through this package.
//
// # Where the DTOs live
//
// Request/response DTOs referenced by a port method live in the
// same file as the port. This keeps the compile boundary tight —
// adapters and the Runner import `ports` only, never legacy
// `pkg/appstate` types. DTO fields reuse existing CRD types from
// `github.com/beclab/api/api/app.bytetrade.io/v1alpha1` and
// `pkg/appcfg` where doing so does not import `pkg/appstate`; new
// value types are declared here.
package ports
