package fake

import (
	"sort"
	"sync"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/ports"
)

// Clock is a manually-advanced Clock. It exists so tests can drive
// TimeoutFired / ScheduleTick / periodic-tick behaviours
// deterministically without race-y wall-clock sleeps.
//
// Zero value is legal and starts at time.Unix(0, 0).UTC(); use
// NewClock to start at an arbitrary anchor. All methods are safe for
// concurrent use.
type Clock struct {
	mu sync.Mutex

	// now is the "current" wall time reported by Now(). Advanced
	// only via Advance / Set.
	now time.Time

	// pending is the ordered heap of scheduled deliveries that
	// After() and NewTicker() have registered but not yet fired.
	// The slice is kept sorted by fireAt ascending so Advance only
	// walks the prefix that fires.
	pending []scheduled
}

// scheduled is one pending delivery on either an After channel or a
// Ticker channel. The kind field lets Advance decide whether to
// re-schedule after firing (Ticker: yes, After: no).
type scheduled struct {
	fireAt   time.Time
	interval time.Duration // 0 for After, non-zero for NewTicker
	ch       chan time.Time
	stopped  *bool // ticker only; set to true by Ticker.Stop
}

// NewClock returns a Clock anchored at `start`. The anchor is
// arbitrary — tests typically use `time.Date(2026, 1, 1, ...)` for
// readable log timestamps.
func NewClock(start time.Time) *Clock {
	return &Clock{now: start}
}

// Now implements ports.Clock. Returns the current fake time.
func (c *Clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

// After implements ports.Clock. Registers a one-shot delivery at
// `now + d` and returns the receive channel. The returned channel
// is unbuffered; Advance may block briefly at delivery time waiting
// for a reader — tests should typically use `select ... default` or
// spawn a small goroutine to consume.
//
// Callers usually don't invoke After directly; the Runner does. A
// test asserting a ScheduleTick effect can just call Advance(d) and
// then inspect the resulting Reconcile.
func (c *Clock) After(d time.Duration) <-chan time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	ch := make(chan time.Time, 1) // buffered so Advance never blocks
	c.schedule(scheduled{fireAt: c.now.Add(d), ch: ch})
	return ch
}

// NewTicker implements ports.Clock. Returns a Ticker whose C fires
// every d fake seconds after each Advance that spans a firing.
func (c *Clock) NewTicker(d time.Duration) ports.Ticker {
	if d <= 0 {
		panic("fake.Clock.NewTicker: non-positive interval")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	stopped := false
	ch := make(chan time.Time, 1)
	c.schedule(scheduled{fireAt: c.now.Add(d), interval: d, ch: ch, stopped: &stopped})
	return &fakeTicker{clock: c, ch: ch, stopped: &stopped}
}

// Advance moves the clock forward by d, firing every After/Ticker
// whose fireAt falls within [now, now+d]. The scheduler is stable:
// deliveries fire in fireAt order and, on ties, in registration
// order. Tickers with elapsed intervals fire once per Advance call
// (they do NOT catch up multiple missed periods — matching
// time.Ticker's actual behaviour under back-pressure).
//
// Returns the number of deliveries dispatched, for test convenience.
func (c *Clock) Advance(d time.Duration) int {
	if d < 0 {
		panic("fake.Clock.Advance: negative duration")
	}
	c.mu.Lock()
	target := c.now.Add(d)
	c.now = target

	fired := 0
	// walk-and-collect: we release the mutex before sending on the
	// channels so a reader that calls back into the clock (e.g. a
	// Runner delegating to Reconcile) doesn't deadlock.
	var toDeliver []scheduled
	remaining := c.pending[:0]
	for _, s := range c.pending {
		if s.stopped != nil && *s.stopped {
			continue // ticker was stopped; drop it silently
		}
		if !s.fireAt.After(target) {
			toDeliver = append(toDeliver, s)
			if s.interval > 0 {
				// re-arm the ticker at fireAt + interval (not
				// target + interval) so long Advances don't drift
				// the phase.
				next := s.fireAt.Add(s.interval)
				for !next.After(target) {
					next = next.Add(s.interval) // skip missed periods, per time.Ticker semantics
				}
				s.fireAt = next
				remaining = append(remaining, s)
			}
		} else {
			remaining = append(remaining, s)
		}
	}
	c.pending = remaining
	sort.SliceStable(c.pending, func(i, j int) bool {
		return c.pending[i].fireAt.Before(c.pending[j].fireAt)
	})
	c.mu.Unlock()

	for _, s := range toDeliver {
		// Non-blocking send: channels are buffered=1; if the
		// reader hasn't drained yet we drop, matching time.Ticker
		// coalescing behaviour.
		select {
		case s.ch <- target:
			fired++
		default:
			// coalesced drop; count-as-not-fired for observability
		}
	}
	return fired
}

// Set jumps the clock to `t`. Must be >= current now (time only
// moves forward — Set is Advance-with-a-timestamp).
func (c *Clock) Set(t time.Time) int {
	c.mu.Lock()
	d := t.Sub(c.now)
	c.mu.Unlock()
	return c.Advance(d)
}

// Pending returns the number of un-fired scheduled deliveries.
// Useful in tests to assert "handler emitted exactly one
// ScheduleTick effect".
func (c *Clock) Pending() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.pending)
}

// schedule inserts s into c.pending in fireAt order. Caller must
// hold c.mu.
func (c *Clock) schedule(s scheduled) {
	c.pending = append(c.pending, s)
	sort.SliceStable(c.pending, func(i, j int) bool {
		return c.pending[i].fireAt.Before(c.pending[j].fireAt)
	})
}

// fakeTicker implements ports.Ticker on top of a Clock.
type fakeTicker struct {
	clock   *Clock
	ch      chan time.Time
	stopped *bool
}

// C implements ports.Ticker.
func (t *fakeTicker) C() <-chan time.Time { return t.ch }

// Stop implements ports.Ticker. Safe to call multiple times.
func (t *fakeTicker) Stop() {
	t.clock.mu.Lock()
	defer t.clock.mu.Unlock()
	*t.stopped = true
}

// Interface conformance guard.
var _ ports.Clock = (*Clock)(nil)
var _ ports.Ticker = (*fakeTicker)(nil)
