package clusterop

import (
	"errors"
	"testing"
)

func TestClaimStorePersistsClaimsAcrossInstances(t *testing.T) {
	dir := t.TempDir()
	first, err := NewClaimStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := first.Claim("alice\x00reboot\x00request-1"); err != nil {
		t.Fatal(err)
	}

	second, err := NewClaimStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := second.Claim("alice\x00reboot\x00request-1"); !errors.Is(err, ErrClaimExists) {
		t.Fatalf("second claim error = %v, want %v", err, ErrClaimExists)
	}
}

func TestReleasedClaimCanBeClaimedAgain(t *testing.T) {
	store, err := NewClaimStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Claim("request"); err != nil {
		t.Fatal(err)
	}
	if err := store.Release("request"); err != nil {
		t.Fatal(err)
	}
	if err := store.Claim("request"); err != nil {
		t.Fatalf("claim after release: %v", err)
	}
}

func TestClaimDistinguishesPendingFromCompleted(t *testing.T) {
	store, err := NewClaimStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Claim("request"); err != nil {
		t.Fatal(err)
	}
	if completed, err := store.Completed("request"); err != nil || completed {
		t.Fatalf("pending claim completed = %v, error = %v", completed, err)
	}
	if err := store.Complete("request"); err != nil {
		t.Fatal(err)
	}
	if completed, err := store.Completed("request"); err != nil || !completed {
		t.Fatalf("completed claim completed = %v, error = %v", completed, err)
	}
}
