//go:build linux

package keychain

import (
	"os"
	"path/filepath"
	"testing"
)

// TestStorageDir_UsesValidatedDataDirEnv verifies the absolute-path branch
// of $OLARES_CLI_DATA_DIR after Clean: the supplied path normalizes ".."
// segments, but service isolation is preserved by the trailing join.
func TestStorageDir_UsesValidatedDataDirEnv(t *testing.T) {
	base := t.TempDir()
	base, _ = filepath.EvalSymlinks(base)
	t.Setenv("OLARES_CLI_DATA_DIR", filepath.Join(base, "data", "..", "store"))

	got := StorageDir("svc")
	want := filepath.Join(base, "store", "svc")
	if got != want {
		t.Fatalf("StorageDir() = %q, want %q", got, want)
	}
}

// TestStorageDir_InvalidDataDirFallsBackToDefault verifies that a non-absolute
// $OLARES_CLI_DATA_DIR is rejected and we fall back to the XDG default. We
// never want a relative dir floating with cwd.
func TestStorageDir_InvalidDataDirFallsBackToDefault(t *testing.T) {
	home := t.TempDir()
	home, _ = filepath.EvalSymlinks(home)
	t.Setenv("OLARES_CLI_DATA_DIR", "relative-data")
	t.Setenv("HOME", home)

	got := StorageDir("svc")
	want := filepath.Join(home, ".local", "share", "svc")
	if got != want {
		t.Fatalf("StorageDir() = %q, want %q", got, want)
	}
}

// TestPlatformRoundTrip exercises the full Get/Set/Remove cycle on the file
// backend (which is what's actually shipped on Linux). This is the closest
// we can get to a smoke test without mocking keychain bits.
func TestPlatformRoundTrip(t *testing.T) {
	home := t.TempDir()
	home, _ = filepath.EvalSymlinks(home)
	t.Setenv("OLARES_CLI_DATA_DIR", "")
	t.Setenv("HOME", home)

	const (
		service = "olares-cli-test"
		account = "alice@olares.com"
		secret  = `{"olaresId":"alice@olares.com","accessToken":"abc"}`
	)

	if got, err := platformGet(service, account); err != nil || got != "" {
		t.Fatalf("platformGet on empty store = (%q, %v); want (\"\", nil)", got, err)
	}

	if err := platformSet(service, account, secret); err != nil {
		t.Fatalf("platformSet() error = %v", err)
	}
	got, err := platformGet(service, account)
	if err != nil {
		t.Fatalf("platformGet() error = %v", err)
	}
	if got != secret {
		t.Fatalf("platformGet() = %q, want %q", got, secret)
	}

	// File perms must stay 0600 — the encrypted blob is plaintext to anyone
	// with read access to the master key sitting next to it.
	encPath := filepath.Join(StorageDir(service), safeFileName(account))
	st, err := os.Stat(encPath)
	if err != nil {
		t.Fatalf("stat encrypted file: %v", err)
	}
	if mode := st.Mode().Perm(); mode != 0o600 {
		t.Fatalf("encrypted file perms = %v, want 0600", mode)
	}

	if err := platformRemove(service, account); err != nil {
		t.Fatalf("platformRemove() error = %v", err)
	}
	if got, err := platformGet(service, account); err != nil || got != "" {
		t.Fatalf("platformGet after Remove = (%q, %v); want (\"\", nil)", got, err)
	}
	if err := platformRemove(service, account); err != nil {
		t.Fatalf("platformRemove(missing) = %v; want nil", err)
	}
}

// TestPlatformGet_CorruptedEncryptedBlob ensures that a tampered .enc file
// surfaces a decryption error rather than a silent empty return.
func TestPlatformGet_CorruptedEncryptedBlob(t *testing.T) {
	home := t.TempDir()
	home, _ = filepath.EvalSymlinks(home)
	t.Setenv("OLARES_CLI_DATA_DIR", "")
	t.Setenv("HOME", home)

	service := "olares-cli-test"
	account := "alice@olares.com"

	if err := platformSet(service, account, "secret"); err != nil {
		t.Fatalf("seed platformSet() error = %v", err)
	}
	encPath := filepath.Join(StorageDir(service), safeFileName(account))
	if err := os.WriteFile(encPath, []byte("corrupt"), 0o600); err != nil {
		t.Fatalf("corrupt blob write: %v", err)
	}

	if _, err := platformGet(service, account); err == nil {
		t.Fatal("platformGet(corrupted) returned nil; want error")
	}
}
