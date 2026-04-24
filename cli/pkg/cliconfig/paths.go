// Package cliconfig owns the on-disk profile configuration of olares-cli
// (~/.olares-cli/config.json + ~/.olares-cli/tokens.json).
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

// Filenames inside the config dir.
const (
	configFilename = "config.json"
	tokensFilename = "tokens.json"
)

// Permissions for the config dir & files. tokens.json carries refresh tokens
// in plaintext during Phase 1; both files therefore use 0600 / 0700.
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

// TokensFile returns the absolute path to tokens.json (without creating it).
func TokensFile() (string, error) {
	dir, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, tokensFilename), nil
}
