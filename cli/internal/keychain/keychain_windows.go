//go:build windows

package keychain

import (
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// regRootPath is the HKCU subtree that holds all olares-cli-managed entries.
// Each service nests one level deeper (registryPathForService); each account
// becomes a value name inside the per-service key.
const regRootPath = `Software\OlaresCli\keychain`

// registryPathForService returns the per-service registry key path.
func registryPathForService(service string) string {
	return regRootPath + `\` + safeRegistryComponent(service)
}

var safeRegRe = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

// safeRegistryComponent strips characters that would either nest a registry
// path unintentionally ('\\') or trip up registry tooling (anything outside
// alphanumerics + ._-). The output is deterministic so existing entries keep
// being addressable across CLI versions.
func safeRegistryComponent(s string) string {
	s = strings.ReplaceAll(s, "\\", "_")
	return safeRegRe.ReplaceAllString(s, "_")
}

// valueNameForAccount keeps the registry value name fully alphanumeric (URL-
// safe base64). This avoids edge cases with '@' / '.' that olaresIds carry.
func valueNameForAccount(account string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(account))
}

// dpapiEntropy binds the ciphertext to (service, account) so a copied blob
// can't be decrypted under a different identity slot. DPAPI itself is keyed
// to the Windows user, so the blob is also useless if exfiltrated to another
// account on the same machine.
func dpapiEntropy(service, account string) *windows.DataBlob {
	data := []byte(service + "\x00" + account)
	if len(data) == 0 {
		return nil
	}
	return &windows.DataBlob{Size: uint32(len(data)), Data: &data[0]}
}

func dpapiProtect(plaintext []byte, entropy *windows.DataBlob) ([]byte, error) {
	var in windows.DataBlob
	if len(plaintext) > 0 {
		in = windows.DataBlob{Size: uint32(len(plaintext)), Data: &plaintext[0]}
	}
	var out windows.DataBlob
	err := windows.CryptProtectData(&in, nil, entropy, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out)
	if err != nil {
		return nil, err
	}
	defer freeDataBlob(&out)

	if out.Data == nil || out.Size == 0 {
		return []byte{}, nil
	}
	buf := unsafe.Slice(out.Data, int(out.Size))
	res := make([]byte, len(buf))
	copy(res, buf)
	return res, nil
}

func dpapiUnprotect(ciphertext []byte, entropy *windows.DataBlob) ([]byte, error) {
	var in windows.DataBlob
	if len(ciphertext) > 0 {
		in = windows.DataBlob{Size: uint32(len(ciphertext)), Data: &ciphertext[0]}
	}
	var out windows.DataBlob
	err := windows.CryptUnprotectData(&in, nil, entropy, 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out)
	if err != nil {
		return nil, err
	}
	defer freeDataBlob(&out)

	if out.Data == nil || out.Size == 0 {
		return []byte{}, nil
	}
	buf := unsafe.Slice(out.Data, int(out.Size))
	res := make([]byte, len(buf))
	copy(res, buf)
	return res, nil
}

// freeDataBlob releases the DPAPI-allocated buffer per the contract that
// CryptProtectData / CryptUnprotectData impose on their out-parameters.
func freeDataBlob(b *windows.DataBlob) {
	if b == nil || b.Data == nil {
		return
	}
	_, _ = windows.LocalFree(windows.Handle(unsafe.Pointer(b.Data)))
	b.Data = nil
	b.Size = 0
}

func platformGet(service, account string) (string, error) {
	v, ok := registryGet(service, account)
	if !ok {
		return "", nil
	}
	return v, nil
}

func platformSet(service, account, data string) error {
	entropy := dpapiEntropy(service, account)
	protected, err := dpapiProtect([]byte(data), entropy)
	if err != nil {
		return fmt.Errorf("dpapi protect failed: %w", err)
	}
	return registrySet(service, account, protected)
}

func platformRemove(service, account string) error {
	return registryRemove(service, account)
}

// registryGet pulls the base64-DPAPI blob from HKCU and unwraps it.
func registryGet(service, account string) (string, bool) {
	keyPath := registryPathForService(service)
	k, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.QUERY_VALUE)
	if err != nil {
		return "", false
	}
	defer k.Close()

	b64, _, err := k.GetStringValue(valueNameForAccount(account))
	if err != nil || b64 == "" {
		return "", false
	}
	blob, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", false
	}
	entropy := dpapiEntropy(service, account)
	plain, err := dpapiUnprotect(blob, entropy)
	if err != nil {
		return "", false
	}
	return string(plain), true
}

func registrySet(service, account string, protected []byte) error {
	keyPath := registryPathForService(service)
	k, _, err := registry.CreateKey(registry.CURRENT_USER, keyPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("registry create/open failed: %w", err)
	}
	defer k.Close()

	b64 := base64.StdEncoding.EncodeToString(protected)
	if err := k.SetStringValue(valueNameForAccount(account), b64); err != nil {
		return fmt.Errorf("registry set failed: %w", err)
	}
	return nil
}

func registryRemove(service, account string) error {
	keyPath := registryPathForService(service)
	k, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.SET_VALUE)
	if err != nil {
		return nil
	}
	defer k.Close()
	_ = k.DeleteValue(valueNameForAccount(account))
	return nil
}

// platformBackend on windows is always the per-user registry hive sealed
// with DPAPI — there is no fallback path so this is a constant.
func platformBackend(_ string) string { return "registry+dpapi" }

// platformPurge removes the entire per-service registry key under HKCU,
// which deletes every stored value (one per account) in one shot. DPAPI
// material itself is not user-visible state — wiping the registry value
// makes the ciphertext unreachable and OS-level cleanup is automatic.
//
// Missing keys are treated as success (idempotent purge): the registry
// package surfaces ERROR_FILE_NOT_FOUND for a key that was never created,
// which we map back to nil. Any other error bubbles up so wrapError can
// add the (service, "*") tag.
func platformPurge(service string) error {
	keyPath := registryPathForService(service)
	err := registry.DeleteKey(registry.CURRENT_USER, keyPath)
	if err == nil {
		return nil
	}
	if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
		return nil
	}
	return fmt.Errorf("registry delete %s: %w", keyPath, err)
}
