package realclock

import (
	"testing"
	"time"
)

// TestNow_MonotonicallyForward — the least-content test possible
// on a wall-clock adapter, but critical: proves Now actually
// advances (i.e. we didn't accidentally return a captured value).
func TestNow_MonotonicallyForward(t *testing.T) {
	c := New()
	t1 := c.Now()
	time.Sleep(2 * time.Millisecond)
	t2 := c.Now()
	if !t2.After(t1) {
		t.Errorf("Now(t2)=%v not after Now(t1)=%v", t2, t1)
	}
}

// TestAfter_Fires — proves After delivers within a reasonable
// window. Wall-clock timing on shared CI is imprecise; we allow a
// generous upper bound because the assertion is "did it fire at
// all", not "fired within N ms".
func TestAfter_Fires(t *testing.T) {
	c := New()
	select {
	case <-c.After(5 * time.Millisecond):
	case <-time.After(1 * time.Second):
		t.Fatal("After(5ms) did not fire within 1s")
	}
}

// TestNewTicker_FiresAndStops — proves the ticker delivers, AND
// that Stop is idempotent (the port contract requires it — the
// naive `time.Ticker.Stop` is safe to call twice but doesn't
// short-circuit, so we wrap it).
func TestNewTicker_FiresAndStops(t *testing.T) {
	c := New()
	tk := c.NewTicker(5 * time.Millisecond)

	select {
	case <-tk.C():
	case <-time.After(1 * time.Second):
		t.Fatal("Ticker.C did not fire within 1s")
	}

	tk.Stop()
	tk.Stop() // idempotency check — must not panic
}
