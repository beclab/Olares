//go:build darwin

package keychain

import (
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/zalando/go-keyring"
)

// TestPlatformSetFallsBackToFileMasterKey verifies writes fall back to the
// on-disk master key when both reads (`keyringGet -> ErrNotFound`) and writes
// (`keyringSet -> blocked`) against the system keychain fail. This mirrors
// what happens inside a sandbox / CI runner where the user can't grant
// keychain access at all.
func TestPlatformSetFallsBackToFileMasterKey(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	origGet := keyringGet
	origSet := keyringSet
	keyringGet = func(service, user string) (string, error) {
		return "", keyring.ErrNotFound
	}
	keyringSet = func(service, user, password string) error {
		return errors.New("blocked")
	}
	t.Cleanup(func() {
		keyringGet = origGet
		keyringSet = origSet
	})

	service := "test-service"
	account := "alice@olares.com"
	secret := "secret-value"

	if err := platformSet(service, account, secret); err != nil {
		t.Fatalf("platformSet() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(StorageDir(service), fileMasterKeyName)); err != nil {
		t.Fatalf("file master key not created: %v", err)
	}
	got, err := platformGet(service, account)
	if err != nil {
		t.Fatalf("platformGet() error = %v", err)
	}
	if got != secret {
		t.Fatalf("platformGet() = %q, want %q", got, secret)
	}
}

// TestPlatformGetPrefersFileMasterKey verifies that when both a file master
// key and a (different) system-keychain master key exist, decryption tries
// the file one first. This is the reason a sandboxed-then-unsandboxed CLI
// keeps reading its own writes instead of trying to decrypt them with the
// "newer" system key.
func TestPlatformGetPrefersFileMasterKey(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	fileKey := make([]byte, masterKeyBytes)
	for i := range fileKey {
		fileKey[i] = byte(i + 1)
	}
	keychainKey := make([]byte, masterKeyBytes)
	for i := range keychainKey {
		keychainKey[i] = byte(i + 33)
	}

	origGet := keyringGet
	origSet := keyringSet
	keyringGet = func(service, user string) (string, error) {
		return base64.StdEncoding.EncodeToString(keychainKey), nil
	}
	keyringSet = func(service, user, password string) error { return nil }
	t.Cleanup(func() {
		keyringGet = origGet
		keyringSet = origSet
	})

	service := "test-service"
	account := "alice@olares.com"
	secret := "secret-value"

	dir := StorageDir(service)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, fileMasterKeyName), fileKey, 0o600); err != nil {
		t.Fatalf("write master key: %v", err)
	}
	encrypted, err := encryptData(secret, fileKey)
	if err != nil {
		t.Fatalf("encryptData() error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, safeFileName(account)), encrypted, 0o600); err != nil {
		t.Fatalf("write secret: %v", err)
	}

	got, err := platformGet(service, account)
	if err != nil {
		t.Fatalf("platformGet() error = %v", err)
	}
	if got != secret {
		t.Fatalf("platformGet() = %q, want %q", got, secret)
	}
}

// TestPlatformSetPrefersExistingFileMasterKey verifies that once the file
// master key exists, subsequent writes never touch the system keychain. This
// guarantees no surprise prompts for users who first used the CLI in a
// sandboxed environment.
func TestPlatformSetPrefersExistingFileMasterKey(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	origGet := keyringGet
	origSet := keyringSet
	keyringGet = func(service, user string) (string, error) {
		t.Fatalf("keyringGet should not be called when file master key exists")
		return "", nil
	}
	keyringSet = func(service, user, password string) error {
		t.Fatalf("keyringSet should not be called when file master key exists")
		return nil
	}
	t.Cleanup(func() {
		keyringGet = origGet
		keyringSet = origSet
	})

	service := "test-service"
	account := "alice@olares.com"
	secret := "secret-value"

	dir := StorageDir(service)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	fileKey := make([]byte, masterKeyBytes)
	for i := range fileKey {
		fileKey[i] = byte(i + 1)
	}
	if err := os.WriteFile(filepath.Join(dir, fileMasterKeyName), fileKey, 0o600); err != nil {
		t.Fatalf("write master key: %v", err)
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
}

// TestPlatformRemove_NotPresentIsNoop verifies removing a missing entry does
// not error — the contract Remove() advertises to callers (so `profile remove`
// can be idempotent on machines where login was never run).
func TestPlatformRemove_NotPresentIsNoop(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := platformRemove("svc", "no-such-account"); err != nil {
		t.Fatalf("platformRemove(missing) error = %v", err)
	}
}
