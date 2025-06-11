package dsa_test

import (
	"encoding/hex"
	"testing"

	"olares-cli/pkg/web5/crypto/dsa"
	"olares-cli/pkg/web5/crypto/dsa/ecdsa"
	"olares-cli/pkg/web5/crypto/dsa/eddsa"
	"olares-cli/pkg/web5/jwk"

	"github.com/alecthomas/assert/v2"
)

func TestGeneratePrivateKeySECP256K1(t *testing.T) {
	privateJwk, err := dsa.GeneratePrivateKey(dsa.AlgorithmIDSECP256K1)

	assert.NoError(t, err)
	assert.Equal[string](t, ecdsa.SECP256K1JWACurve, privateJwk.CRV)
	assert.Equal[string](t, ecdsa.KeyType, privateJwk.KTY)
	assert.True(t, privateJwk.D != "", "privateJwk.D is empty")
	assert.True(t, privateJwk.X != "", "privateJwk.X is empty")
	assert.True(t, privateJwk.Y != "", "privateJwk.Y is empty")
}

func TestGeneratePrivateKeyED25519(t *testing.T) {
	privateJwk, err := dsa.GeneratePrivateKey(dsa.AlgorithmIDED25519)
	if err != nil {
		t.Errorf("failed to generate private key: %v", err.Error())
	}

	assert.NoError(t, err)
	assert.Equal[string](t, eddsa.ED25519JWACurve, privateJwk.CRV)
	assert.Equal[string](t, eddsa.KeyType, privateJwk.KTY)
	assert.True(t, privateJwk.D != "", "privateJwk.D is empty")
	assert.True(t, privateJwk.X != "", "privateJwk.X is empty")
}

func TestSignSECP256K1(t *testing.T) {
	privateJwk, err := dsa.GeneratePrivateKey(dsa.AlgorithmIDSECP256K1)
	assert.NoError(t, err)

	payload := []byte("hello world")
	signature, err := dsa.Sign(payload, privateJwk)

	assert.NoError(t, err)
	assert.True(t, len(signature) == 64, "invalid signature length")
}

func TestSignED25519(t *testing.T) {
	privateJwk, err := dsa.GeneratePrivateKey(dsa.AlgorithmIDED25519)
	assert.NoError(t, err)

	payload := []byte("hello world")
	signature, err := dsa.Sign(payload, privateJwk)

	assert.NoError(t, err)
	assert.True(t, len(signature) == 64, "invalid signature length")
}

func TestSignDeterministicSECP256K1(t *testing.T) {
	privateJwk, err := dsa.GeneratePrivateKey(dsa.AlgorithmIDSECP256K1)
	assert.NoError(t, err)

	payload := []byte("hello world")
	signature1, err := dsa.Sign(payload, privateJwk)
	assert.NoError(t, err, "failed to sign")

	signature2, err := dsa.Sign(payload, privateJwk)
	assert.NoError(t, err)

	assert.Equal(t, signature1, signature2, "signature is not deterministic")
}

func TestSignDeterministicED25519(t *testing.T) {
	privateJwk, err := dsa.GeneratePrivateKey(dsa.AlgorithmIDED25519)
	assert.NoError(t, err)

	payload := []byte("hello world")
	signature1, err := dsa.Sign(payload, privateJwk)
	assert.NoError(t, err, "failed to sign")

	signature2, err := dsa.Sign(payload, privateJwk)
	assert.NoError(t, err)

	assert.Equal(t, signature1, signature2, "signature is not deterministic")
}

func TestVerifySECP256K1(t *testing.T) {
	privateJwk, err := dsa.GeneratePrivateKey(dsa.AlgorithmIDSECP256K1)
	assert.NoError(t, err)

	payload := []byte("hello world")
	signature, err := dsa.Sign(payload, privateJwk)
	assert.NoError(t, err)

	publicJwk := dsa.GetPublicKey(privateJwk)

	legit, err := dsa.Verify(payload, signature, publicJwk)
	assert.NoError(t, err)

	assert.True(t, legit, "failed to verify signature")
}

func TestVerifyED25519(t *testing.T) {
	privateJwk, err := dsa.GeneratePrivateKey(dsa.AlgorithmIDED25519)
	assert.NoError(t, err)

	payload := []byte("hello world")
	signature, err := dsa.Sign(payload, privateJwk)

	assert.NoError(t, err)

	publicJwk := dsa.GetPublicKey(privateJwk)

	legit, err := dsa.Verify(payload, signature, publicJwk)
	assert.NoError(t, err)

	assert.True(t, legit, "failed to verify signature")
}

func TestBytesToPublicKey_BadAlgorithm(t *testing.T) {
	_, err := dsa.BytesToPublicKey("yolocrypto", []byte{0x00, 0x01, 0x02, 0x03})
	assert.Error(t, err)
}

func TestBytesToPublicKey_BadBytes(t *testing.T) {
	_, err := dsa.BytesToPublicKey(dsa.AlgorithmIDSECP256K1, []byte{0x00, 0x01, 0x02, 0x03})
	assert.Error(t, err)
}

func TestBytesToPublicKey_SECP256K1(t *testing.T) {
	// vector taken from https://github.com/decentralized-identity/web5-js/blob/dids-new-crypto/packages/crypto/tests/fixtures/test-vectors/secp256k1/bytes-to-public-key.json
	publicKeyHex := "0479be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798483ada7726a3c4655da4fbfc0e1108a8fd17b448a68554199c47d08ffb10d4b8"
	pubKeyBytes, err := hex.DecodeString(publicKeyHex)
	assert.NoError(t, err)

	jwk, err := dsa.BytesToPublicKey(dsa.AlgorithmIDSECP256K1, pubKeyBytes)
	assert.NoError(t, err)

	assert.Equal(t, ecdsa.SECP256K1JWACurve, jwk.CRV)
	assert.Equal(t, ecdsa.KeyType, jwk.KTY)
	assert.Equal(t, "eb5mfvncu6xVoGKVzocLBwKb_NstzijZWfKBWxb4F5g", jwk.X)
	assert.Equal(t, "SDradyajxGVdpPv8DhEIqP0XtEimhVQZnEfQj_sQ1Lg", jwk.Y)
}

func TestPublicKeyToBytes_UnsupportedKTY(t *testing.T) {
	_, err := dsa.PublicKeyToBytes(jwk.JWK{KTY: "yolocrypto"})
	assert.Error(t, err)
}
