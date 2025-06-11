package didkey

import (
	"encoding/base64"
	"fmt"

	"olares-cli/pkg/web5/jwk"

	"github.com/mr-tron/base58"
	"github.com/multiformats/go-varint"
)

// base58Alphabet is the alphabet used for base58btc encoding
// EncodeBase58BTC 对输入数据进行 Base58btc 编码
func EncodeBase58BTC(data []byte) string {
	// 调用 base58 库进行编码
	base58Encoded := base58.Encode(data)
	// 添加 Base58btc 的前缀 'z'
	return "z" + base58Encoded
}

// KeyToID converts a public key JWK to a did:key ID
func KeyToID(publicKey jwk.JWK) (string, error) {
	// Decode the public key X value from base64url
	pubKeyBytes, err := base64.RawURLEncoding.DecodeString(publicKey.X)
	if err != nil {
		return "", fmt.Errorf("failed to decode public key: %w", err)
	}

	ed25519CodecID := 0xed

	// Create the multicodec prefix
	prefix := varint.ToUvarint(uint64(ed25519CodecID))

	// Combine the prefix and public key bytes
	idBytes := make([]byte, len(prefix)+len(pubKeyBytes))
	copy(idBytes, prefix)
	copy(idBytes[len(prefix):], pubKeyBytes)

	// Encode to base58btc
	id := EncodeBase58BTC(idBytes)

	return id, nil
}
