package ecdsa_test

import (
	"encoding/hex"
	"testing"

	"github.com/beclab/Olares/cli/pkg/web5/crypto/dsa/ecdsa"
	"github.com/beclab/Olares/cli/pkg/web5/jwk"

	"github.com/alecthomas/assert/v2"
)

func TestSECP256K1GeneratePrivateKey(t *testing.T) {
	key, err := ecdsa.SECP256K1GeneratePrivateKey()
	assert.NoError(t, err)

	assert.Equal(t, ecdsa.KeyType, key.KTY)
	assert.Equal(t, ecdsa.SECP256K1JWACurve, key.CRV)
	assert.True(t, key.D != "", "privateJwk.D is empty")
	assert.True(t, key.X != "", "privateJwk.X is empty")
	assert.True(t, key.Y != "", "privateJwk.Y is empty")
}

func TestSECP256K1BytesToPublicKey_Bad(t *testing.T) {
	_, err := ecdsa.SECP256K1BytesToPublicKey([]byte{0x00, 0x01, 0x02, 0x03})
	assert.Error(t, err)
}

func TestSECP256K1BytesToPublicKey_Uncompressed(t *testing.T) {
	// vector taken from https://github.com/decentralized-identity/web5-js/blob/dids-new-crypto/packages/crypto/tests/fixtures/test-vectors/secp256k1/bytes-to-public-key.json
	publicKeyHex := "0479be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b8"
	pubKeyBytes, err := hex.DecodeString(publicKeyHex)
	assert.NoError(t, err)

	jwk, err := ecdsa.SECP256K1BytesToPublicKey(pubKeyBytes)
	assert.NoError(t, err)

	assert.Equal(t, ecdsa.SECP256K1JWACurve, jwk.CRV)
	assert.Equal(t, ecdsa.KeyType, jwk.KTY)
	assert.Equal(t, "eb5mfvncu6xVoGKVzocLBwKb_NstzijZWfKBWxb4F5g", jwk.X)
	assert.Equal(t, "SDradyajxGVdpPv8DhEIqP0XtEimhVQZnEfQj_sQ1Lg", jwk.Y)
}

func TestSECP256K1PublicKeyToBytes(t *testing.T) {
	// vector taken from https://github.com/decentralized-identity/web5-js/blob/dids-new-crypto/packages/crypto/tests/fixtures/test-vectors/secp256k1/bytes-to-public-key.json
	jwk := jwk.JWK{
		KTY: "EC",
		CRV: ecdsa.SECP256K1JWACurve,
		X:   "eb5mfvncu6xVoGKVzocLBwKb_NstzijZWfKBWxb4F5g",
		Y:   "SDradyajxGVdpPv8DhEIqP0XtEimhVQZnEfQj_sQ1Lg",
	}

	pubKeyBytes, err := ecdsa.SECP256K1PublicKeyToBytes(jwk)
	assert.NoError(t, err)

	pubKeyHex := hex.EncodeToString(pubKeyBytes)
	expected := "0479be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b8"

	assert.Equal(t, expected, pubKeyHex)
}

func TestSECP256K1PublicKeyToBytes_Bad(t *testing.T) {
	vectors := []jwk.JWK{
		{
			KTY: "EC",
			CRV: ecdsa.SECP256K1JWACurve,
			X:   "eb5mfvncu6xVoGKVzocLBwKb_NstzijZWfKBWxb4F5g",
		},
		{
			KTY: "EC",
			CRV: ecdsa.SECP256K1JWACurve,
			Y:   "eb5mfvncu6xVoGKVzocLBwKb_NstzijZWfKBWxb4F5g",
		},
		{
			KTY: "EC",
			CRV: ecdsa.SECP256K1JWACurve,
			X:   "=///",
			Y:   "SDradyajxGVdpPv8DhEIqP0XtEimhVQZnEfQj_sQ1Lg",
		},
		{
			KTY: "EC",
			CRV: ecdsa.SECP256K1JWACurve,
			X:   "eb5mfvncu6xVoGKVzocLBwKb_NstzijZWfKBWxb4F5g",
			Y:   "=///",
		},
		{
			KTY: "EC",
			CRV: ecdsa.SECP256K1JWACurve,
			X:   "eb5mfvncu6xVoGKVzocLBwKb_NstzijZWfKBWxb4F5g",
			Y:   "SDradyajxGVdpPv8DhEIqP0XtEimhVQZnEfQj_sQ1Lg2",
		},
	}

	for _, vec := range vectors {
		pubKeyBytes, err := ecdsa.SECP256K1PublicKeyToBytes(vec)
		assert.Error(t, err)
		assert.Equal(t, nil, pubKeyBytes)
	}
}
