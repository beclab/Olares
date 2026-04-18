package wizard

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
)

// aesGCMEncrypt encrypts data using AES-GCM with a 16-byte nonce, matching
// the TS provider in apps/packages/sdk/src/core/container.ts which uses
// `additionalData` as AAD (and never appends it to the ciphertext).
func aesGCMEncrypt(key, plaintext, iv, additionalData []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, len(iv))
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}
	return gcm.Seal(nil, iv, plaintext, additionalData), nil
}

// aesGCMDecrypt is the inverse of aesGCMEncrypt; iv length must match the
// length used at encryption time (TS uses 16 bytes).
func aesGCMDecrypt(key, ciphertext, iv, additionalData []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}
	gcm, err := cipher.NewGCMWithNonceSize(block, len(iv))
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}
	plaintext, err := gcm.Open(nil, iv, ciphertext, additionalData)
	if err != nil {
		return nil, fmt.Errorf("AES-GCM decrypt failed: %w", err)
	}
	return plaintext, nil
}

// rsaOAEPEncrypt encrypts payload with the given DER-encoded RSA public key
// using RSA-OAEP/SHA-256, matching `RSAEncryptionParams` in TS.
func rsaOAEPEncrypt(publicKeyDER []byte, payload []byte) ([]byte, error) {
	pubAny, err := x509.ParsePKIXPublicKey(publicKeyDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA public key: %w", err)
	}
	pub, ok := pubAny.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, pub, payload, nil)
}

// rsaOAEPDecrypt decrypts an RSA-OAEP/SHA-256 payload using the provided
// PKCS1 DER-encoded private key (matching how Account.initialize stores it).
func rsaOAEPDecrypt(privateKeyDER []byte, ciphertext []byte) ([]byte, error) {
	priv, err := parseRSAPrivateKeyDER(privateKeyDER)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptOAEP(sha256.New(), rand.Reader, priv, ciphertext, nil)
}

// parseRSAPrivateKeyDER accepts both PKCS1 and PKCS8 encoded RSA private
// keys; falls back to PKCS8 if PKCS1 parsing fails.
func parseRSAPrivateKeyDER(privateKeyDER []byte) (*rsa.PrivateKey, error) {
	if priv, err := x509.ParsePKCS1PrivateKey(privateKeyDER); err == nil {
		return priv, nil
	}
	keyAny, err := x509.ParsePKCS8PrivateKey(privateKeyDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA private key (tried PKCS1 and PKCS8): %w", err)
	}
	priv, ok := keyAny.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("PKCS8 key is not an RSA key")
	}
	return priv, nil
}

// generateAESKey returns a fresh random 32-byte AES key (256 bits).
func generateAESKey() []byte {
	return generateRandomBytes(32)
}
