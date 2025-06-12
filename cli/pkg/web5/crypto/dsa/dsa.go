package dsa

import (
	"fmt"

	"github.com/beclab/Olares/cli/pkg/web5/crypto/dsa/ecdsa"
	"github.com/beclab/Olares/cli/pkg/web5/crypto/dsa/eddsa"
	"github.com/beclab/Olares/cli/pkg/web5/jwk"
)

const (
	AlgorithmIDSECP256K1 = ecdsa.SECP256K1AlgorithmID
	AlgorithmIDED25519   = eddsa.ED25519AlgorithmID
)

// GeneratePrivateKey generates a private key using the algorithm specified by algorithmID.
func GeneratePrivateKey(algorithmID string) (jwk.JWK, error) {
	if ecdsa.SupportsAlgorithmID(algorithmID) {
		return ecdsa.GeneratePrivateKey(algorithmID)
	} else if eddsa.SupportsAlgorithmID(algorithmID) {
		return eddsa.GeneratePrivateKey(algorithmID)
	}

	return jwk.JWK{}, fmt.Errorf("unsupported algorithm: %s", algorithmID)
}

// GetPublicKey returns the public key corresponding to the given private key.
func GetPublicKey(privateKey jwk.JWK) jwk.JWK {
	switch privateKey.KTY {
	case ecdsa.KeyType:
		return ecdsa.GetPublicKey(privateKey)
	case eddsa.KeyType:
		return eddsa.GetPublicKey(privateKey)
	default:
		return jwk.JWK{}
	}
}

// Sign signs the payload using the given private key.
func Sign(payload []byte, jwk jwk.JWK) ([]byte, error) {
	switch jwk.KTY {
	case ecdsa.KeyType:
		return ecdsa.Sign(payload, jwk)
	case eddsa.KeyType:
		return eddsa.Sign(payload, jwk)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", jwk.KTY)
	}
}

// Verify verifies the signature of the payload using the given public key.
func Verify(payload []byte, signature []byte, jwk jwk.JWK) (bool, error) {
	switch jwk.KTY {
	case ecdsa.KeyType:
		return ecdsa.Verify(payload, signature, jwk)
	case eddsa.KeyType:
		return eddsa.Verify(payload, signature, jwk)
	default:
		return false, fmt.Errorf("unsupported key type: %s", jwk.KTY)
	}
}

// GetJWA returns the JWA (JSON Web Algorithm) algorithm corresponding to the given key.
func GetJWA(jwk jwk.JWK) (string, error) {
	switch jwk.KTY {
	case ecdsa.KeyType:
		return ecdsa.GetJWA(jwk)
	case eddsa.KeyType:
		return eddsa.GetJWA(jwk)
	default:
		return "", fmt.Errorf("unsupported key type: %s", jwk.KTY)
	}
}

// BytesToPublicKey converts the given bytes to a public key based on the algorithm specified by algorithmID.
func BytesToPublicKey(algorithmID string, input []byte) (jwk.JWK, error) {
	if ecdsa.SupportsAlgorithmID(algorithmID) {
		return ecdsa.BytesToPublicKey(algorithmID, input)
	} else if eddsa.SupportsAlgorithmID(algorithmID) {
		return eddsa.BytesToPublicKey(algorithmID, input)
	}

	return jwk.JWK{}, fmt.Errorf("unsupported algorithm: %s", algorithmID)
}

// PublicKeyToBytes converts the provided public key to bytes
func PublicKeyToBytes(publicKey jwk.JWK) ([]byte, error) {
	switch publicKey.KTY {
	case ecdsa.KeyType:
		return ecdsa.PublicKeyToBytes(publicKey)
	case eddsa.KeyType:
		return eddsa.PublicKeyToBytes(publicKey)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", publicKey.KTY)
	}
}

// AlgorithmID returns the algorithm ID for the given jwk.JWK
func AlgorithmID(jwk *jwk.JWK) (string, error) {
	switch jwk.KTY {
	case ecdsa.KeyType:
		return ecdsa.AlgorithmID(jwk)
	case eddsa.KeyType:
		return eddsa.AlgorithmID(jwk)
	default:
		return "", fmt.Errorf("unsupported key type: %s", jwk.KTY)
	}
}
