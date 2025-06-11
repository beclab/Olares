package ecdsa

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"

	"olares-cli/pkg/web5/jwk"

	_secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

const (
	SECP256K1JWA         string = "ES256K"
	SECP256K1JWACurve    string = "secp256k1"
	SECP256K1AlgorithmID string = SECP256K1JWACurve
)

// SECP256K1GeneratePrivateKey generates a new private key
func SECP256K1GeneratePrivateKey() (jwk.JWK, error) {
	keyPair, err := _secp256k1.GeneratePrivateKey()
	if err != nil {
		return jwk.JWK{}, fmt.Errorf("failed to generate private key: %w", err)
	}

	dBytes := keyPair.Key.Bytes()
	pubKey := keyPair.PubKey()
	xBytes := pubKey.X().Bytes()
	yBytes := pubKey.Y().Bytes()

	privateKey := jwk.JWK{
		KTY: KeyType,
		CRV: SECP256K1JWACurve,
		D:   base64.RawURLEncoding.EncodeToString(dBytes[:]),
		X:   base64.RawURLEncoding.EncodeToString(xBytes),
		Y:   base64.RawURLEncoding.EncodeToString(yBytes),
	}

	return privateKey, nil
}

// SECP256K1Sign signs the given payload with the given private key
func SECP256K1Sign(payload []byte, privateKey jwk.JWK) ([]byte, error) {
	privateKeyBytes, err := base64.RawURLEncoding.DecodeString(privateKey.D)
	if err != nil {
		return nil, fmt.Errorf("failed to decode d %w", err)
	}

	key := _secp256k1.PrivKeyFromBytes(privateKeyBytes)

	hash := sha256.Sum256(payload)
	signature := ecdsa.SignCompact(key, hash[:], false)[1:]

	return signature, nil
}

// SECP256K1Verify verifies the given signature over the given payload with the given public key
func SECP256K1Verify(payload []byte, signature []byte, publicKey jwk.JWK) (bool, error) {
	if publicKey.X == "" || publicKey.Y == "" {
		return false, errors.New("x and y must be set")
	}

	hash := sha256.Sum256(payload)

	keyBytes, err := secp256k1PublicKeyToUncheckedBytes(publicKey)
	if err != nil {
		return false, fmt.Errorf("failed to convert public key to bytes: %w", err)
	}

	key, err := _secp256k1.ParsePubKey(keyBytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse public key: %w", err)
	}

	if len(signature) != 64 {
		return false, errors.New("signature must be 64 bytes")
	}

	r := new(_secp256k1.ModNScalar)
	r.SetByteSlice(signature[:32])

	s := new(_secp256k1.ModNScalar)
	s.SetByteSlice(signature[32:])

	sig := ecdsa.NewSignature(r, s)
	legit := sig.Verify(hash[:], key)

	return legit, nil
}

// SECP256K1BytesToPublicKey converts a secp256k1 public key to a JWK.
// Supports both Compressed and Uncompressed public keys described in
// https://www.secg.org/sec1-v2.pdf section 2.3.3
func SECP256K1BytesToPublicKey(input []byte) (jwk.JWK, error) {
	pubKey, err := _secp256k1.ParsePubKey(input)
	if err != nil {
		return jwk.JWK{}, fmt.Errorf("failed to parse public key: %w", err)
	}

	return jwk.JWK{
		KTY: KeyType,
		CRV: SECP256K1JWACurve,
		X:   base64.RawURLEncoding.EncodeToString(pubKey.X().Bytes()),
		Y:   base64.RawURLEncoding.EncodeToString(pubKey.Y().Bytes()),
	}, nil
}

// SECP256K1PublicKeyToBytes converts a secp256k1 public key JWK to bytes.
// Note: this function returns the uncompressed public key. compressed is not
// yet supported
func SECP256K1PublicKeyToBytes(publicKey jwk.JWK) ([]byte, error) {
	uncheckedBytes, err := secp256k1PublicKeyToUncheckedBytes(publicKey)
	if err != nil {
		return nil, err
	}

	key, err := _secp256k1.ParsePubKey(uncheckedBytes)
	if err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}

	return key.SerializeUncompressed(), nil
}

func secp256k1PublicKeyToUncheckedBytes(publicKey jwk.JWK) ([]byte, error) {
	if publicKey.X == "" || publicKey.Y == "" {
		return nil, errors.New("x and y must be set")
	}

	x, err := base64.RawURLEncoding.DecodeString(publicKey.X)
	if err != nil {
		return nil, fmt.Errorf("failed to decode x: %w", err)
	}

	y, err := base64.RawURLEncoding.DecodeString(publicKey.Y)
	if err != nil {
		return nil, fmt.Errorf("failed to decode y: %w", err)
	}

	// Prepend 0x04 to indicate an uncompressed public key format for secp256k1.
	// This byte is a prefix that distinguishes uncompressed keys, which include both X and Y coordinates,
	// from compressed keys which only include one coordinate and an indication of the other's parity.
	// The secp256k1 standard requires this prefix for uncompressed keys to ensure proper interpretation.
	keyBytes := []byte{0x04}
	keyBytes = append(keyBytes, x...)
	keyBytes = append(keyBytes, y...)

	return keyBytes, nil
}
