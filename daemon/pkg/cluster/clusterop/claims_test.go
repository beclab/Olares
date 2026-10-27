package clusterop

import (
	"errors"
	"os"
	"testing"
	"time"
)

func TestReplayGuardPersistsConsumedBindingAcrossInstances(t *testing.T) {
	dir := t.TempDir()
	expiresAt := time.Now().Add(time.Minute)
	first, err := NewReplayGuard(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := first.Consume("signature/request-1", expiresAt); err != nil {
		t.Fatal(err)
	}

	second, err := NewReplayGuard(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := second.Consume("signature/request-1", expiresAt); !errors.Is(err, ErrReplayConflict) {
		t.Fatalf("second consume error = %v, want %v", err, ErrReplayConflict)
	}
}

func TestReplayGuardDeletesExpiredMarkerBeforeConsume(t *testing.T) {
	guard, err := NewReplayGuard(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := guard.Consume("request", time.Now().Add(-time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := guard.Consume("request", time.Now().Add(time.Minute)); err != nil {
		t.Fatalf("consume after expiry: %v", err)
	}
}

func TestReplayGuardConsumeCleansExpiredMarkersForOtherKeys(t *testing.T) {
	guard, err := NewReplayGuard(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := guard.Consume("expired-request", time.Now().Add(-time.Second)); err != nil {
		t.Fatal(err)
	}
	expiredPath := guard.path("expired-request")
	if err := guard.Consume("new-request", time.Now().Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(expiredPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expired marker remains after another consume: %v", err)
	}
}

func TestReplayGuardCleansExpiredMarkersAcrossInstances(t *testing.T) {
	dir := t.TempDir()
	first, err := NewReplayGuard(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := first.Consume("expired-request", time.Now().Add(-time.Second)); err != nil {
		t.Fatal(err)
	}

	second, err := NewReplayGuard(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := second.Cleanup(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(second.path("expired-request")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expired marker remains: %v", err)
	}
}

func TestReplayGuardCleanupLeavesDamagedMarkersConsumed(t *testing.T) {
	guard, err := NewReplayGuard(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	path := guard.path("damaged-request")
	if err := os.WriteFile(path, []byte("{"), recordMode); err != nil {
		t.Fatal(err)
	}
	if err := guard.Cleanup(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("damaged marker was not retained safely: %v", err)
	}
}

func TestReplayGuardCanForgetRejectedCommand(t *testing.T) {
	guard, err := NewReplayGuard(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := guard.Consume("request", time.Now().Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := guard.Forget("request"); err != nil {
		t.Fatal(err)
	}
	if err := guard.Consume("request", time.Now().Add(time.Minute)); err != nil {
		t.Fatalf("consume after rejected command: %v", err)
	}
}
