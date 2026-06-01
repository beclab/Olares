// Package keychain provides cross-platform secure storage for olares-cli secrets
// (currently used for the per-olaresId access/refresh token grants written by
// the `profile login` and `profile import` commands).
//
// The implementation is adapted from larksuite/cli's internal/keychain package
// (same Get/Set/Remove surface, same per-platform strategy split):
//
//   - macOS: a 32-byte AES-256 master key is kept in the system Keychain via
//     github.com/zalando/go-keyring; per-secret data is AES-GCM encrypted and
//     written to ~/Library/Application Support/<service>/<safeFileName>.enc.
//     If the system keychain is blocked (sandbox / CI) the master key falls
//     back to an on-disk master.key.file (mode 0600) under the same dir, so
//     the CLI keeps working at a Linux-equivalent security posture.
//   - Linux: pure file-based AES-GCM. The master key lives at
//     ~/.local/share/<service>/master.key (mode 0600); each secret lives at
//     <safeFileName>.enc next to it. Honors $OLARES_CLI_DATA_DIR when set to
//     an absolute path.
//   - Windows: DPAPI-protected blob (CryptProtectData/CryptUnprotectData)
//     persisted under HKCU\Software\OlaresCli\keychain\<service>, with
//     deterministic entropy bound to (service, account) to thwart swap/replay.
//
// olares-cli keeps the package internal-only on purpose: its consumers all
// live inside this repo and the package-level Get/Set/Remove surface
// intentionally mirrors lark-cli so future security upgrades can be ported
// back without renaming dance.
//
// Olares-side adaptations vs. the upstream lark-cli copy:
//   - dropped lark-cli's internal/vfs (we use stdlib os directly),
//     internal/output (we return plain wrapped errors with the same hint),
//     internal/validate (we do an explicit absolute-path check inline),
//   - dropped auth_log.go (audit logging is out of scope for now;
//     LogAuthError calls were removed),
//   - service constant renamed to OlaresCliService = "olares-cli",
//   - $LARKSUITE_CLI_DATA_DIR override renamed to $OLARES_CLI_DATA_DIR,
//   - Windows registry root changed to Software\OlaresCli\keychain,
//   - extra KeychainAccess interface + Default() seam + keychainfake
//     subpackage exist on the olares-cli side, even though lark-cli has no
//     equivalent. The reason is that olares-cli's keychainStore (in
//     pkg/auth/token_store_keychain.go) carries semantics lark-cli's
//     token_store.go does not — MarkInvalidated, List with
//     ProfileLister-driven enumeration, and InvalidatedAt — and we want
//     unit-test coverage for those without monkey-patching the platform
//     seams. Tests inject keychainfake.New() through NewTokenStoreWith;
//     production code only ever constructs the store via NewTokenStore
//     which wires keychain.Default(). lark-cli gets away without this
//     interface because its token_store.go has no unit tests of its own.
package keychain
