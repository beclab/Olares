// Package cliconfig owns the on-disk profile configuration of olares-cli
// (~/.olares-cli/config.json). Token secrets are NOT stored here — they
// live in the OS keychain via cli/internal/keychain.
//
// The package is named cliconfig (not "config") to avoid clashing with the
// pre-existing cmd/config package, which serves a different purpose
// (per-command flag wiring).
package cliconfig

import (
	"fmt"
	"os"
	"path/filepath"
)

// homeEnv is the environment variable used to override the config dir, mirroring
// lark-cli's $LARK_CLI_HOME convention.
const homeEnv = "OLARES_CLI_HOME"

// defaultDir is the directory name used under $HOME when $OLARES_CLI_HOME is
// unset.
const defaultDir = ".olares-cli"

// configFilename is the only file this package owns. Token secrets used to
// live next to it as tokens.json (Phase 1, plaintext); Phase 2 moved them
// into the OS keychain — see cli/internal/keychain and
// cli/pkg/auth/token_store_keychain.go.
const configFilename = "config.json"

// Permissions for the config dir & file. config.json holds the profile index
// (no secrets) but we still keep it 0600 because it does carry the
// `currentProfile` selection and any auth-URL overrides.
const (
	dirPerm  os.FileMode = 0o700
	filePerm os.FileMode = 0o600
)

// Home returns the resolved olares-cli config directory. It honors the
// $OLARES_CLI_HOME override and falls back to $HOME/.olares-cli. The directory
// is NOT created here — callers that intend to write should call EnsureHome
// instead.
func Home() (string, error) {
	if v := os.Getenv(homeEnv); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home: %w", err)
	}
	return filepath.Join(home, defaultDir), nil
}

// EnsureHome resolves Home() and ensures the directory exists with 0700 perms.
func EnsureHome() (string, error) {
	dir, err := Home()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, dirPerm); err != nil {
		return "", fmt.Errorf("create %s: %w", dir, err)
	}
	return dir, nil
}

// ConfigFile returns the absolute path to config.json (without creating it).
func ConfigFile() (string, error) {
	dir, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFilename), nil
}
