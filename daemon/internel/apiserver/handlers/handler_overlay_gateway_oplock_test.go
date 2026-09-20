package handlers

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestClearOverlayGatewayOpLocks(t *testing.T) {
	dir := t.TempDir()
	enable := filepath.Join(dir, "enable.lock")
	disable := filepath.Join(dir, "disable.lock")

	if err := os.WriteFile(enable, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(disable, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	clearOverlayGatewayOpLocks(enable, disable)

	if _, err := os.Stat(enable); !os.IsNotExist(err) {
		t.Fatalf("enable lock still present: %v", err)
	}
	if _, err := os.Stat(disable); !os.IsNotExist(err) {
		t.Fatalf("disable lock still present: %v", err)
	}

	// Idempotent when missing.
	clearOverlayGatewayOpLocks(enable, disable)
}

func TestConsumeOverlayGatewayOpLock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "op.lock")

	inFlight, stale := consumeOverlayGatewayOpLock(path)
	if inFlight || stale != "" {
		t.Fatalf("absent lock: inFlight=%v stale=%q", inFlight, stale)
	}

	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	inFlight, stale = consumeOverlayGatewayOpLock(path)
	if !inFlight || stale != "" {
		t.Fatalf("fresh lock: inFlight=%v stale=%q", inFlight, stale)
	}

	past := time.Now().Add(-OverlayGatewayOpLockTTL - time.Second)
	if err := os.Chtimes(path, past, past); err != nil {
		t.Fatal(err)
	}
	inFlight, stale = consumeOverlayGatewayOpLock(path)
	if inFlight {
		t.Fatal("stale lock still reported in-flight")
	}
	if stale != overlayGatewayOpLockStaleMsg {
		t.Fatalf("stale msg=%q", stale)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("stale lock not removed: %v", err)
	}
}
