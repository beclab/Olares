package keychain

import (
	"errors"
	"fmt"
	"os"
)

var (
	// ErrNotFound is returned when the requested credential is not found.
	// platformGet implementations return ("", nil) for the common
	// "not present" case; ErrNotFound is reserved for callers that want to
	// turn that empty-string signal into a typed error and is also the
	// sentinel that wrapError refuses to mask.
	ErrNotFound = errors.New("keychain: item not found")

	// errNotInitialized is the internal sentinel used when the master key is
	// missing or invalid. It triggers a more specific operator hint in
	// wrapError so users know to reset / reconfigure rather than blaming
	// permissions.
	errNotInitialized = errors.New("keychain not initialized")
)

// OlaresCliService is the unified keychain service name for all olares-cli
// secrets. Per-secret records are distinguished by their account name, which
// today is always the bare olaresId (e.g. "alice@olares.com"). Mirrors
// lark-cli's `LarkCliService = "lark-cli"` design.
const OlaresCliService = "olares-cli"

// debugEnv toggles the long, multi-line operator hint that wrapError used
// to attach unconditionally. Default-on hints made every error message
// mushroom past 200 chars; gating them behind this env var keeps everyday
// failures grep-friendly and reserves the verbose version for users who
// actually want it.
const debugEnv = "OLARES_CLI_DEBUG"

// debugLookup is a package-level seam so tests can flip the hint on/off
// without writing to process env (which would race other tests in the
// same package).
var debugLookup = func() bool { return os.Getenv(debugEnv) != "" }

// wrapError is the single funnel that turns underlying backend errors into
// user-facing messages. Returning ErrNotFound (or nil) is preserved
// verbatim so callers can use errors.Is on it.
//
// Default output format:
//
//	keychain <op> failed for <service>/<account>: <cause>
//
// With OLARES_CLI_DEBUG set, an actionable English hint is appended in
// parentheses. The two-tier hint (generic vs errNotInitialized) is
// preserved — only the gating changes. Including the (service, account)
// pair on every line means logs are grep-able to a specific keychain
// slot without needing the surrounding context.
func wrapError(op, service, account string, err error) error {
	if err == nil || errors.Is(err, ErrNotFound) {
		return err
	}

	base := fmt.Errorf("keychain %s failed for %s/%s: %w", op, service, account, err)
	if !debugLookup() {
		return base
	}

	hint := "Check whether the OS keychain / credential manager is locked or accessible. " +
		"If you are running inside a sandbox or CI environment, ensure the process has " +
		"permission to use the keychain — running outside the sandbox usually fixes it."
	if errors.Is(err, errNotInitialized) {
		hint = "The keychain master key may have been deleted or corrupted. " +
			"Re-run `olares-cli profile login` (or `profile import`) to re-issue credentials. " +
			"In sandboxed / CI environments, ensure the process can access the OS keychain."
	}
	return fmt.Errorf("%w (%s)", base, hint)
}

// KeychainAccess abstracts Get/Set/Remove for dependency injection. Production
// code wires the package-level functions through Default(); tests can pass a
// fake to assert call patterns without touching the real OS keychain.
type KeychainAccess interface {
	Get(service, account string) (string, error)
	Set(service, account, value string) error
	Remove(service, account string) error
}

// Get retrieves a value from the keychain. Returns ("", nil) when the entry
// does not exist (mirrors lark-cli's contract; callers that prefer a typed
// "not found" should check len(value)==0 or wrap with ErrNotFound).
func Get(service, account string) (string, error) {
	val, err := platformGet(service, account)
	return val, wrapError("Get", service, account, err)
}

// Set stores a value in the keychain, overwriting any existing entry.
func Set(service, account, data string) error {
	return wrapError("Set", service, account, platformSet(service, account, data))
}

// Remove deletes an entry from the keychain. Removing a non-existent entry is
// a no-op and returns nil, matching lark-cli's behavior.
func Remove(service, account string) error {
	return wrapError("Remove", service, account, platformRemove(service, account))
}

// Backend returns a short, machine-friendly identifier of the platform
// backend currently in effect for service. Values are stable strings that
// callers can include in user-facing notices and grep for in logs:
//
//   - "system-keychain" — darwin, master key lives in the OS keychain
//   - "file-fallback"   — darwin, sandbox/CI path: master key on disk
//   - "file"            — linux, master key on disk under XDG dir
//   - "registry+dpapi"  — windows, registry value protected by DPAPI
//
// Knowing which backend is active matters because the security posture
// differs (system keychain prompts on access; file-fallback does not). The
// value is recomputed on every call so we always reflect the current
// on-disk state — relevant for tests that move files around mid-process.
func Backend(service string) string { return platformBackend(service) }

// PurgeService wipes ALL keychain state owned by the given service: the
// per-account encrypted blobs AND the master key (system-keychain entry on
// darwin, on-disk file on darwin/linux, registry hive on windows). Designed
// to be called when the last olares-cli profile is removed so we don't
// leave orphan secrets / files / registry values that would surface in
// security tooling and confuse users.
//
// Errors are wrapped through the same wrapError funnel for consistency,
// but callers (currently `profile remove`) are expected to log + continue
// on failure rather than abort: the user-facing config has already been
// updated and a leftover encrypted blob without a master key is harmless.
func PurgeService(service string) error {
	return wrapError("Purge", service, "*", platformPurge(service))
}
