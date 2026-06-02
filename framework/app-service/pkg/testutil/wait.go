package testutil

import (
	"testing"
	"time"
)

// WaitClosed blocks until done is closed or the timeout elapses, failing the
// test on timeout. It is used to synchronize on the appstate operation's
// Done() channel without importing the appstate package (avoiding a cycle).
func WaitClosed(t *testing.T, done <-chan struct{}, timeout time.Duration) {
	t.Helper()
	if done == nil {
		t.Fatal("WaitClosed: done channel is nil")
	}
	select {
	case <-done:
	case <-time.After(timeout):
		t.Fatalf("WaitClosed: timed out after %s waiting for operation to finish", timeout)
	}
}

// Eventually polls cond until it returns true or the timeout elapses.
func Eventually(t *testing.T, timeout, interval time.Duration, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		if cond() {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("Eventually: condition not met within %s", timeout)
		}
		time.Sleep(interval)
	}
}
