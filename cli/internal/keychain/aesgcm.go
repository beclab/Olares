//go:build !windows

package keychain

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"os"
	"regexp"
)

// Shared AES-256-GCM + filename helpers for the file-backed platform
// implementations (currently darwin and linux). Windows uses DPAPI plus a
// base64 registry value name and does not import this file (the build tag
// gates it out).
//
// Centralising these guarantees the two file backends never drift on the
// crypto envelope (IV size, tag size, key size) or the on-disk filename
// scheme — a subtle drift there would silently strand existing users on the
// day someone touches just one platform's copy.

// AES-256-GCM parameters. The blob layout is `iv || ciphertext || tag`,
// where the tag is appended to the ciphertext by aesGCM.Seal — that's why
// decryptData gates on len >= ivBytes+tagBytes.
const (
	masterKeyBytes = 32
	ivBytes        = 12
	tagBytes       = 16
)

// safeFileNameRe / safeFileName turn an arbitrary account name (which is
// the olaresId, e.g. "alice@olares.com" — note the '@' and '.' which we
// keep, plus any custom alias) into a filename safe to land on the FS.
// Anything outside the whitelist collapses to '_'.
var safeFileNameRe = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

func safeFileName(account string) string {
	return safeFileNameRe.ReplaceAllString(account, "_") + ".enc"
}

// encryptData seals plaintext under AES-256-GCM with a fresh random IV.
func encryptData(plaintext string, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	iv := make([]byte, ivBytes)
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}
	ciphertext := aesGCM.Seal(nil, iv, []byte(plaintext), nil)
	result := make([]byte, 0, ivBytes+len(ciphertext))
	result = append(result, iv...)
	result = append(result, ciphertext...)
	return result, nil
}

// decryptData is the symmetric inverse of encryptData.
func decryptData(data []byte, key []byte) (string, error) {
	if len(data) < ivBytes+tagBytes {
		return "", os.ErrInvalid
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	iv := data[:ivBytes]
	ciphertext := data[ivBytes:]
	plaintext, err := aesGCM.Open(nil, iv, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
