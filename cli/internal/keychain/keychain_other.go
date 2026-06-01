//go:build linux

package keychain

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// AES constants and crypto helpers (encryptData / decryptData) plus
// safeFileName live in aesgcm.go (build-tag !windows) so darwin and linux
// can't drift on the on-disk envelope.

// dataDirEnv lets users (typically inside containers / nix builds / test
// rigs) relocate the encrypted store. Keeping the env var olares-specific
// avoids accidental clashes with lark-cli on machines that have both.
const dataDirEnv = "OLARES_CLI_DATA_DIR"

// StorageDir returns the absolute directory for service-scoped encrypted
// blobs on Linux. The lookup chain is:
//
//  1. $OLARES_CLI_DATA_DIR if it's an absolute, cleanly-resolved path,
//  2. XDG-style ~/.local/share/<service>,
//  3. an absolute fallback under os.TempDir() when HOME is unresolvable.
//     Earlier versions returned ".local/share/<service>" relative to CWD,
//     which silently writes to wherever the user happened to invoke the
//     CLI from (worst case: cwd=/). Anchoring to TempDir() guarantees the
//     path is absolute even if it's not the canonical XDG location.
func StorageDir(service string) string {
	if dir := os.Getenv(dataDirEnv); dir != "" {
		if safeDir, ok := safeAbsoluteDir(dir); ok {
			return filepath.Join(safeDir, service)
		}
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		fallback := filepath.Join(os.TempDir(), "olares-cli", "keychain", service)
		fmt.Fprintf(os.Stderr,
			"warning: home directory unresolvable (%v); using fallback keychain dir %s\n",
			err, fallback)
		return fallback
	}
	return filepath.Join(home, ".local", "share", service)
}

// safeAbsoluteDir is a deliberately narrow alternative to lark-cli's
// validate.SafeEnvDirPath: we only accept already-absolute paths after a
// Clean (which collapses "..", "//"), so users can't trick us into landing
// next to the binary or inside a relative path that floats with cwd.
func safeAbsoluteDir(p string) (string, bool) {
	cleaned := filepath.Clean(p)
	if !filepath.IsAbs(cleaned) {
		return "", false
	}
	return cleaned, true
}

// getMasterKey reads (or, when allowCreate is true, generates) the per-service
// master key on disk. The temp-file + rename guards against torn writes when
// two processes race the first-time creation.
func getMasterKey(service string, allowCreate bool) ([]byte, error) {
	dir := StorageDir(service)
	keyPath := filepath.Join(dir, "master.key")

	key, err := os.ReadFile(keyPath)
	if err == nil && len(key) == masterKeyBytes {
		return key, nil
	}
	if err == nil && len(key) != masterKeyBytes {
		return nil, errors.New("keychain is corrupted")
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	if !allowCreate {
		return nil, errNotInitialized
	}

	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	key = make([]byte, masterKeyBytes)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}

	tmpKeyPath := filepath.Join(dir, "master.key."+uuid.New().String()+".tmp")
	defer os.Remove(tmpKeyPath)

	if err := os.WriteFile(tmpKeyPath, key, 0o600); err != nil {
		return nil, err
	}
	if err := os.Rename(tmpKeyPath, keyPath); err != nil {
		// Lost the race: another process won, read its key.
		existingKey, readErr := os.ReadFile(keyPath)
		if readErr == nil && len(existingKey) == masterKeyBytes {
			return existingKey, nil
		}
		return nil, err
	}
	return key, nil
}

func platformGet(service, account string) (string, error) {
	path := filepath.Join(StorageDir(service), safeFileName(account))
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	key, err := getMasterKey(service, false)
	if err != nil {
		return "", err
	}
	plaintext, err := decryptData(data, key)
	if err != nil {
		return "", err
	}
	return plaintext, nil
}

func platformSet(service, account, data string) error {
	key, err := getMasterKey(service, true)
	if err != nil {
		return err
	}
	dir := StorageDir(service)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	encrypted, err := encryptData(data, key)
	if err != nil {
		return err
	}

	targetPath := filepath.Join(dir, safeFileName(account))
	tmpPath := filepath.Join(dir, safeFileName(account)+"."+uuid.New().String()+".tmp")
	defer os.Remove(tmpPath)

	if err := os.WriteFile(tmpPath, encrypted, 0o600); err != nil {
		return err
	}
	return os.Rename(tmpPath, targetPath)
}

func platformRemove(service, account string) error {
	err := os.Remove(filepath.Join(StorageDir(service), safeFileName(account)))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// platformBackend on linux is always file-based (no system keychain
// integration in this build). Returned as a constant so callers don't have
// to special-case the OS.
func platformBackend(_ string) string { return "file" }

// platformPurge wipes the entire service-scoped storage dir on linux:
// master.key + every per-account .enc blob. There is no system keychain
// entry to delete on this build.
func platformPurge(service string) error {
	return os.RemoveAll(StorageDir(service))
}
