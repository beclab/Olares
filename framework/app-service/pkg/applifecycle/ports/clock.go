package ports

import "time"

// Clock is the injectable time source. It exists so:
//
//  1. Handlers can compute timeouts without importing "time" (which
//     also makes handlerpurity rule 3 easier to enforce — time.Sleep
//     stays a hard reject even in hypothetical helper code).
//  2. Tests can advance time deterministically via
//     adapters/fake.Clock.Advance(d) instead of racing wall-clock
//     sleeps. This is what makes TimeoutFired action assertions
//     reproducible.
//  3. The Runner has a single seam to swap in a controllable ticker
//     for ReconcileAll periodic sweeps in integration tests.
//
// Legacy usage this replaces:
//   - pkg/appstate/types.go:303      time.Now() in IsTimeout
//   - pkg/appstate/upgrading_app.go:342-349 custom TTL vs time.Now()
//   - controllers/appmgr_controller.go:77   wait.Until every 2m
//   - controllers/appmgr_controller.go:104  time.Sleep(1s) between AMs
//   - pkg/images/client.go:121              1s ticker download poll
//   - pkg/event/event.go:                   1s retry ticker
//
// All of the above collapse into Now/After/NewTicker on this port.
type Clock interface {
	// Now returns the current wall time. HANDLER-SAFE — the Runner
	// records World.Now = Clock.Now() before Engine.Step so handlers
	// can compare TTLs deterministically.
	Now() time.Time

	// After returns a channel that fires once after d elapses.
	// RUNNER-ONLY — handlers never wait on channels (CONTRACT.md
	// rule 3); they emit ScheduleTick effects instead.
	After(d time.Duration) <-chan time.Time

	// NewTicker returns a Ticker firing every d. RUNNER-ONLY (used
	// by the periodic ReconcileAll loop and by long-running port
	// implementations).
	NewTicker(d time.Duration) Ticker
}

// Ticker is a thin wrapper around time.Ticker with an interface
// signature so adapters/fake can inject a manually-driven version.
type Ticker interface {
	// C returns the receive-only channel the ticker fires on.
	C() <-chan time.Time

	// Stop halts the ticker. Idempotent; safe to call multiple times.
	Stop()
}
