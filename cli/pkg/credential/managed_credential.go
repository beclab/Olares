package credential

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Environment variables set by app-service's cli-credential webhook when an
// application declares `permission.loginOlaresCLI`. Neither is set on a host
// install, which is what makes "am I a managed container?" a cheap check.
//
//   - EnvCredentialsDir points at a read-only mount of the per-install Secret
//     holding credential.json.
//   - EnvCacheDir points at a writable emptyDir the CLI may use for its
//     config + keychain (see cliconfig.Home / keychain.StorageDir).
const (
	EnvCredentialsDir = "OLARES_CLI_CREDENTIALS_DIR"
	EnvCacheDir       = "OLARES_CLI_CACHE_DIR"

	credentialFilename = "credential.json"
)

// ManagedCredential is the wire shape of credential.json, written by
// app-service (framework/app-service/pkg/olarescli/grant.go). The refresh
// token is an LLDAP-derived long-lived grant scoped to the installing user;
// the CLI never persists it anywhere — the mount is its only home.
type ManagedCredential struct {
	RefreshToken string `json:"refreshToken"`
	OlaresID     string `json:"olaresId"`
	AppName      string `json:"appName"`
}

// LoadManagedCredential reads the mounted credential.json, if there is one.
//
// The second return is false for every "this is not a managed container"
// condition as well as every malformed-input condition: env unset, file
// missing, unreadable, not JSON, or any of the three fields empty. A
// half-filled credential is treated as absent rather than as an error
// because a partially-granted install must not stop the CLI from working
// with whatever profiles the user configured by hand.
//
// Diagnosing a grant that should be there but isn't goes through
// OLARES_CLI_DEBUG, which prints the reason to stderr.
func LoadManagedCredential() (*ManagedCredential, bool) {
	dir := strings.TrimSpace(os.Getenv(EnvCredentialsDir))
	if dir == "" {
		return nil, false
	}
	cred, err := loadManagedCredentialFrom(dir)
	if err != nil {
		debugManaged("%v", err)
		return nil, false
	}
	return cred, true
}

// loadManagedCredentialFrom is the testable core: it reports why a directory
// yielded no credential instead of collapsing everything into a bool.
func loadManagedCredentialFrom(dir string) (*ManagedCredential, error) {
	path := filepath.Join(dir, credentialFilename)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("managed credential: %s does not exist", path)
		}
		return nil, fmt.Errorf("managed credential: read %s: %w", path, err)
	}
	cred := &ManagedCredential{}
	if err := json.Unmarshal(data, cred); err != nil {
		return nil, fmt.Errorf("managed credential: parse %s: %w", path, err)
	}
	cred.RefreshToken = strings.TrimSpace(cred.RefreshToken)
	cred.OlaresID = strings.TrimSpace(cred.OlaresID)
	cred.AppName = strings.TrimSpace(cred.AppName)

	// Same judgement as the platform's grantFromSecret: a grant is only a
	// grant when it names a token, a user and an app.
	var missing []string
	if cred.RefreshToken == "" {
		missing = append(missing, "refreshToken")
	}
	if cred.OlaresID == "" {
		missing = append(missing, "olaresId")
	}
	if cred.AppName == "" {
		missing = append(missing, "appName")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("managed credential: %s is missing %s", path, strings.Join(missing, ", "))
	}
	return cred, nil
}

// debugManaged writes to stderr only under OLARES_CLI_DEBUG, matching the
// gating internal/keychain uses for its own hints.
func debugManaged(format string, args ...any) {
	if os.Getenv("OLARES_CLI_DEBUG") == "" {
		return
	}
	fmt.Fprintf(os.Stderr, "[olares-cli managed] "+format+"\n", args...)
}
