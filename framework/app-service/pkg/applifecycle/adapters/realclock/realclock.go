// Package realclock is the production ports.Clock adapter: a
// direct pass-through to the standard library time package.
//
// Everything the fake Clock simulates (deterministic Advance,
// pending inspection, scheduled-drop coalescing) is redundant in
// production — the real time package provides the ground truth by
// definition, so the adapter is a handful of one-liners.
//
// The only non-trivial concern is that ports.Ticker is an
// interface (so fakes can inject a manually-driven ticker); we
// wrap *time.Ticker in a struct that satisfies it without adding
// any behaviour.
package realclock

import (
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
)

// Clock is the production ports.Clock adapter. Zero value is
// usable; construct via New for symmetry with other adapters.
type Clock struct{}

// New returns a Clock. Kept as a constructor rather than a bare
// struct literal at call sites for consistency with the rest of
// the adapters/* packages.
func New() *Clock { return &Clock{} }

// Now implements ports.Clock.
func (Clock) Now() time.Time { return time.Now() }

// After implements ports.Clock.
func (Clock) After(d time.Duration) <-chan time.Time { return time.After(d) }

// NewTicker implements ports.Clock.
func (Clock) NewTicker(d time.Duration) ports.Ticker {
	return &ticker{t: time.NewTicker(d)}
}

// ticker wraps *time.Ticker as a ports.Ticker so callers can hold
// the interface type. Stop is guarded by a stopped flag because
// ports.Ticker.Stop is documented idempotent and the underlying
// time.Ticker.Stop is NOT (calling Stop twice on a stopped ticker
// is safe but calling Stop after the receiver is nil is not).
type ticker struct {
	t       *time.Ticker
	stopped bool
}

// C implements ports.Ticker.
func (t *ticker) C() <-chan time.Time { return t.t.C }

// Stop implements ports.Ticker. Idempotent, per the port contract.
func (t *ticker) Stop() {
	if t.stopped {
		return
	}
	t.stopped = true
	t.t.Stop()
}

// Compile-time interface conformance guards. Regressions here
// break the build rather than only main.go, giving us a fast
// signal even before wiring.
var (
	_ ports.Clock  = Clock{}
	_ ports.Ticker = (*ticker)(nil)
)
