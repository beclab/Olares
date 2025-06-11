// Package eddsa implements the EdDSA signature schemes as per RFC 8032
// https://tools.ietf.org/html/rfc8032. Note: Currently only Ed25519 is supported
package eddsa

import (
	"errors"
	"fmt"

	"olares-cli/pkg/web5/jwk"
)

const (
	JWA     string = "EdDSA"
	KeyType string = "OKP"
)

var algorithmIDs = map[string]bool{
	ED25519AlgorithmID: true,
}

// GeneratePrivateKey generates an EdDSA private key for the given algorithm
func GeneratePrivateKey(algorithmID string) (jwk.JWK, error) {
	switch algorithmID {
	case ED25519AlgorithmID:
		return ED25519GeneratePrivateKey()
	default:
		return jwk.JWK{}, fmt.Errorf("unsupported algorithm: %s", algorithmID)
	}
}

// GetPublicKey builds an EdDSA public key from the given EdDSA private key
func GetPublicKey(privateKey jwk.JWK) jwk.JWK {
	return jwk.JWK{
		KTY: privateKey.KTY,
		CRV: privateKey.CRV,
		X:   privateKey.X,
	}
}

// Sign generates a cryptographic signature for the given payload with the given private key
//
// # Note
//
// The function will automatically detect the given EdDSA cryptographic curve from the given private key
func Sign(payload []byte, privateKey jwk.JWK) ([]byte, error) {
	if privateKey.D == "" {
		return nil, errors.New("d must be set")
	}

	switch privateKey.CRV {
	case ED25519JWACurve:
		return ED25519Sign(payload, privateKey)
	default:
		return nil, fmt.Errorf("unsupported curve: %s", privateKey.CRV)
	}
}

// Verify verifies the given signature over a given payload by the given public key
//
// # Note
//
// The function will automatically detect the given EdDSA cryptographic curve from the given public key
func Verify(payload []byte, signature []byte, publicKey jwk.JWK) (bool, error) {
	switch publicKey.CRV {
	case ED25519JWACurve:
		return ED25519Verify(payload, signature, publicKey)
	default:
		return false, fmt.Errorf("unsupported curve: %s", publicKey.CRV)
	}
}

// GetJWA returns the [JWA] for the given EdDSA key
//
// # Note
//
// The only supported [JWA] is "EdDSA"
//
// [JWA]: https://datatracker.ietf.org/doc/html/rfc7518
func GetJWA(jwk jwk.JWK) (string, error) {
	return JWA, nil
}

// BytesToPublicKey deserializes the given byte array into a jwk.JWK for the given cryptographic algorithm
func BytesToPublicKey(algorithmID string, input []byte) (jwk.JWK, error) {
	switch algorithmID {
	case ED25519AlgorithmID:
		return ED25519BytesToPublicKey(input)
	default:
		return jwk.JWK{}, fmt.Errorf("unsupported algorithm: %s", algorithmID)
	}
}

// PublicKeyToBytes serializes the given public key into a byte array
func PublicKeyToBytes(publicKey jwk.JWK) ([]byte, error) {
	switch publicKey.CRV {
	case ED25519JWACurve:
		return ED25519PublicKeyToBytes(publicKey)
	default:
		return nil, fmt.Errorf("unsupported curve: %s", publicKey.CRV)
	}
}

// SupportsAlgorithmID informs as to whether or not the given algorithm ID is supported by this package
func SupportsAlgorithmID(id string) bool {
	return algorithmIDs[id]
}

// AlgorithmID returns the algorithm ID for the given jwk.JWK
func AlgorithmID(jwk *jwk.JWK) (string, error) {
	switch jwk.CRV {
	case ED25519JWACurve:
		return ED25519AlgorithmID, nil
	default:
		return "", fmt.Errorf("unsupported curve: %s", jwk.CRV)
	}
}
