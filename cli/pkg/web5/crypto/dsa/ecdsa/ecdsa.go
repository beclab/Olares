package ecdsa

import (
	"errors"
	"fmt"

	"olares-cli/pkg/web5/jwk"
)

const (
	KeyType = "EC"
)

var algorithmIDs = map[string]bool{
	SECP256K1AlgorithmID: true,
}

// GeneratePrivateKey generates an ECDSA private key for the given algorithm
func GeneratePrivateKey(algorithmID string) (jwk.JWK, error) {
	switch algorithmID {
	case SECP256K1AlgorithmID:
		return SECP256K1GeneratePrivateKey()
	default:
		return jwk.JWK{}, fmt.Errorf("unsupported algorithm: %s", algorithmID)
	}
}

// GetPublicKey builds an ECDSA public key from the given ECDSA private key
func GetPublicKey(privateKey jwk.JWK) jwk.JWK {
	return jwk.JWK{
		KTY: privateKey.KTY,
		CRV: privateKey.CRV,
		X:   privateKey.X,
		Y:   privateKey.Y,
	}
}

// Sign generates a cryptographic signature for the given payload with the given private key
//
// # Note
//
// The function will automatically detect the given ECDSA cryptographic curve from the given private key
func Sign(payload []byte, privateKey jwk.JWK) ([]byte, error) {
	if privateKey.D == "" {
		return nil, errors.New("d must be set")
	}

	switch privateKey.CRV {
	case SECP256K1JWACurve:
		return SECP256K1Sign(payload, privateKey)
	default:
		return nil, fmt.Errorf("unsupported curve: %s", privateKey.CRV)
	}
}

// Verify verifies the given signature over a given payload by the given public key
//
// # Note
//
// The function will automatically detect the given ECDSA cryptographic curve from the given public key
func Verify(payload []byte, signature []byte, publicKey jwk.JWK) (bool, error) {
	switch publicKey.CRV {
	case SECP256K1JWACurve:
		return SECP256K1Verify(payload, signature, publicKey)
	default:
		return false, fmt.Errorf("unsupported curve: %s", publicKey.CRV)
	}
}

// GetJWA returns the [JWA] for the given ECDSA key
//
// [JWA]: https://datatracker.ietf.org/doc/html/rfc7518
func GetJWA(jwk jwk.JWK) (string, error) {
	switch jwk.CRV {
	case SECP256K1JWACurve:
		return SECP256K1JWA, nil
	default:
		return "", fmt.Errorf("unsupported curve: %s", jwk.CRV)
	}
}

// BytesToPublicKey deserializes the given byte array into a jwk.JWK for the given cryptographic algorithm
func BytesToPublicKey(algorithmID string, input []byte) (jwk.JWK, error) {
	switch algorithmID {
	case SECP256K1AlgorithmID:
		return SECP256K1BytesToPublicKey(input)
	default:
		return jwk.JWK{}, fmt.Errorf("unsupported algorithm: %s", algorithmID)
	}
}

// PublicKeyToBytes serializes the given public key into a byte array
func PublicKeyToBytes(publicKey jwk.JWK) ([]byte, error) {
	switch publicKey.CRV {
	case SECP256K1JWACurve:
		return SECP256K1PublicKeyToBytes(publicKey)
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
	case SECP256K1JWACurve:
		return SECP256K1AlgorithmID, nil
	default:
		return "", fmt.Errorf("unsupported curve: %s", jwk.CRV)
	}
}
