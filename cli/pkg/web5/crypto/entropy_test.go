package crypto_test

import (
	"encoding/hex"
	"testing"

	"github.com/beclab/Olares/cli/pkg/web5/crypto"

	"github.com/alecthomas/assert/v2"
)

func Test_GenerateEntropy(t *testing.T) {
	bytes, err := crypto.GenerateEntropy(crypto.Entropy128)
	assert.NoError(t, err)
	assert.Equal(t, int(crypto.Entropy128), len(bytes))
}

func Test_GenerateEntropy_CustomSize(t *testing.T) {
	customSize := 99
	bytes, err := crypto.GenerateEntropy(crypto.EntropySize(customSize))
	assert.NoError(t, err)
	assert.Equal(t, customSize, len(bytes))
}

func Test_GenerateEntropy_InvalidSize(t *testing.T) {
	bytes, err := crypto.GenerateEntropy(0)
	assert.Error(t, err)
	assert.Equal(t, nil, bytes)

	bytes, err = crypto.GenerateEntropy(-1)
	assert.Error(t, err)
	assert.Equal(t, nil, bytes)
}

func Test_GenerateNonce(t *testing.T) {
	nonce, err := crypto.GenerateNonce(crypto.Entropy128)
	assert.NoError(t, err)
	assert.Equal(t, int(crypto.Entropy128)*2, len(nonce))

	_, err = hex.DecodeString(nonce)
	assert.NoError(t, err)
}

func Test_GenerateNonce_CustomSize(t *testing.T) {
	customSize := 99
	nonce, err := crypto.GenerateNonce(crypto.EntropySize(99))
	assert.NoError(t, err)
	assert.Equal(t, customSize*2, len(nonce))

	_, err = hex.DecodeString(nonce)
	assert.NoError(t, err)
}

func Test_GenerateNonce_InvalidSize(t *testing.T) {
	nonce, err := crypto.GenerateNonce(0)
	assert.Error(t, err)
	assert.Equal(t, "", nonce)
}
