package fake_test

import (
	"testing"
	"time"

	"github.com/beclab/Olares/framework/app-service/pkg/applifecycle/adapters/fake"
)

// anchor is a fixed, arbitrary-but-readable start point used across
// tests to make log timestamps skimmable. It is not "now" — Clock
// starts here regardless of wall time.
var anchor = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func TestClock_NowAndAdvance_Basic(t *testing.T) {
	c := fake.NewClock(anchor)

	if got := c.Now(); !got.Equal(anchor) {
		t.Fatalf("Now() = %v, want %v", got, anchor)
	}

	c.Advance(5 * time.Second)
	if got := c.Now(); !got.Equal(anchor.Add(5 * time.Second)) {
		t.Fatalf("after Advance(5s), Now() = %v, want %v", got, anchor.Add(5*time.Second))
	}
}

// TestClock_After_FiresAtAdvance pins the single most important
// behaviour: an After(d) channel receives EXACTLY when Advance
// spans d, and NOT before.
func TestClock_After_FiresAtAdvance(t *testing.T) {
	c := fake.NewClock(anchor)

	ch := c.After(3 * time.Second)

	// Nothing has advanced yet — channel must be empty.
	select {
	case <-ch:
		t.Fatal("After(3s) fired before any Advance")
	default:
	}

	// Advance by less than d — still nothing.
	c.Advance(2 * time.Second)
	select {
	case <-ch:
		t.Fatal("After(3s) fired at t=2s")
	default:
	}

	// Advance to exactly d — must fire.
	c.Advance(1 * time.Second)
	select {
	case got := <-ch:
		if !got.Equal(anchor.Add(3 * time.Second)) {
			t.Fatalf("After delivered %v, want %v", got, anchor.Add(3*time.Second))
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("After(3s) did not fire after reaching t=3s")
	}
}

// TestClock_After_ReturnedFireCount pins the Advance return value —
// tests use it as a shorthand to assert "exactly N ScheduleTick
// effects landed in this reconcile".
func TestClock_After_ReturnedFireCount(t *testing.T) {
	c := fake.NewClock(anchor)
	_ = c.After(1 * time.Second)
	_ = c.After(2 * time.Second)
	_ = c.After(5 * time.Second)

	// One Advance that spans two of the three.
	if got := c.Advance(2 * time.Second); got != 2 {
		t.Fatalf("Advance(2s) fired %d, want 2", got)
	}
	if got := c.Pending(); got != 1 {
		t.Fatalf("Pending() = %d, want 1", got)
	}

	// Second Advance fires the last one.
	if got := c.Advance(10 * time.Second); got != 1 {
		t.Fatalf("Advance(10s) fired %d, want 1", got)
	}
	if got := c.Pending(); got != 0 {
		t.Fatalf("Pending() = %d after all fired, want 0", got)
	}
}

// TestClock_NewTicker_FiresRepeatedly pins that a ticker re-arms
// after each firing without needing user re-registration.
func TestClock_NewTicker_FiresRepeatedly(t *testing.T) {
	c := fake.NewClock(anchor)
	ticker := c.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// t=1s: no fire.
	c.Advance(1 * time.Second)
	select {
	case <-ticker.C():
		t.Fatal("ticker fired at t=1s (interval=2s)")
	default:
	}

	// t=2s: one fire.
	c.Advance(1 * time.Second)
	if !recvWithin(t, ticker.C(), 50*time.Millisecond) {
		t.Fatal("ticker did not fire at t=2s")
	}

	// t=4s: one more fire.
	c.Advance(2 * time.Second)
	if !recvWithin(t, ticker.C(), 50*time.Millisecond) {
		t.Fatal("ticker did not fire at t=4s")
	}
}

// TestClock_NewTicker_Coalesce pins time.Ticker parity: a single
// Advance spanning multiple periods delivers AT MOST one tick per
// call (dropped ticks are silently coalesced, matching the real
// stdlib behaviour on a slow reader).
func TestClock_NewTicker_Coalesce(t *testing.T) {
	c := fake.NewClock(anchor)
	ticker := c.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Advance 5s in one shot; only one tick lands on the buffered
	// channel (buffer=1, subsequent sends drop).
	fired := c.Advance(5 * time.Second)
	if fired != 1 {
		t.Fatalf("Advance(5s) fired=%d, want 1 (coalesce)", fired)
	}
	if !recvWithin(t, ticker.C(), 50*time.Millisecond) {
		t.Fatal("ticker did not deliver the coalesced tick")
	}
}

// TestClock_NewTicker_Stop pins Stop idempotency and the guarantee
// that a stopped ticker never fires again.
func TestClock_NewTicker_Stop(t *testing.T) {
	c := fake.NewClock(anchor)
	ticker := c.NewTicker(1 * time.Second)

	ticker.Stop()
	ticker.Stop() // idempotent

	c.Advance(10 * time.Second)
	select {
	case <-ticker.C():
		t.Fatal("stopped ticker fired")
	default:
	}
}

// TestClock_Advance_NegativePanics pins the "time only moves
// forward" invariant, which the Runner relies on when computing
// timeout deadlines.
func TestClock_Advance_NegativePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Advance(-1s) did not panic")
		}
	}()
	c := fake.NewClock(anchor)
	c.Advance(-1 * time.Second)
}

// TestClock_Set_JumpsForward exercises the timestamp-based advance
// convenience.
func TestClock_Set_JumpsForward(t *testing.T) {
	c := fake.NewClock(anchor)
	ch := c.After(5 * time.Second)

	c.Set(anchor.Add(10 * time.Second))
	if !recvWithin(t, ch, 50*time.Millisecond) {
		t.Fatal("Set past deadline did not fire After channel")
	}
	if got := c.Now(); !got.Equal(anchor.Add(10 * time.Second)) {
		t.Fatalf("Now() after Set = %v, want %v", got, anchor.Add(10*time.Second))
	}
}

// recvWithin reports true if ch delivers within d. Used to avoid
// hangs while still allowing goroutine scheduling headroom.
func recvWithin(t *testing.T, ch <-chan time.Time, d time.Duration) bool {
	t.Helper()
	select {
	case <-ch:
		return true
	case <-time.After(d):
		return false
	}
}
