package eddsa

import (
	_ed25519 "crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"olares-cli/pkg/web5/jwk"
)

const (
	ED25519JWACurve    string = "Ed25519"
	ED25519AlgorithmID string = ED25519JWACurve
)

// ED25519GeneratePrivateKey generates a new ED25519 private key
func ED25519GeneratePrivateKey() (jwk.JWK, error) {
	publicKey, privateKey, err := _ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return jwk.JWK{}, err
	}

	privKeyJwk := jwk.JWK{
		KTY: KeyType,
		CRV: ED25519JWACurve,
		D:   base64.RawURLEncoding.EncodeToString(privateKey),
		X:   base64.RawURLEncoding.EncodeToString(publicKey),
	}

	return privKeyJwk, nil
}

// ED25519Sign signs the given payload with the given private key
func ED25519Sign(payload []byte, privateKey jwk.JWK) ([]byte, error) {
	privateKeyBytes, err := base64.RawURLEncoding.DecodeString(privateKey.D)
	if err != nil {
		return nil, fmt.Errorf("failed to decode d %w", err)
	}

	signature := _ed25519.Sign(privateKeyBytes, payload)
	return signature, nil
}

// ED25519Verify verifies the given signature against the given payload using the given public key
func ED25519Verify(payload []byte, signature []byte, publicKey jwk.JWK) (bool, error) {
	publicKeyBytes, err := base64.RawURLEncoding.DecodeString(publicKey.X)
	if err != nil {
		return false, err
	}

	legit := _ed25519.Verify(publicKeyBytes, payload, signature)
	return legit, nil
}

// ED25519BytesToPublicKey deserializes the byte array into a jwk.JWK public key
func ED25519BytesToPublicKey(input []byte) (jwk.JWK, error) {
	if len(input) != _ed25519.PublicKeySize {
		return jwk.JWK{}, errors.New("invalid public key")
	}

	return jwk.JWK{
		KTY: KeyType,
		CRV: ED25519JWACurve,
		X:   base64.RawURLEncoding.EncodeToString(input),
	}, nil
}

// ED25519PublicKeyToBytes serializes the given public key int a byte array
func ED25519PublicKeyToBytes(publicKey jwk.JWK) ([]byte, error) {
	if publicKey.X == "" {
		return nil, errors.New("x must be set")
	}

	publicKeyBytes, err := base64.RawURLEncoding.DecodeString(publicKey.X)
	if err != nil {
		return nil, fmt.Errorf("failed to decode x %w", err)
	}

	return publicKeyBytes, nil
}
