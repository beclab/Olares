package wizard

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"fmt"

	"github.com/beclab/Olares/cli/pkg/web5/jwk"

	"github.com/mr-tron/base58"
	"github.com/multiformats/go-varint"
	"github.com/tyler-smith/go-bip39"
)

// Ed25519 multicodec identifier
const ED25519_CODEC_ID = 0xed

// DIDKeyResult represents the result of DID key generation
type DIDKeyResult struct {
	DID        string  `json:"did"`
	PublicJWK  jwk.JWK `json:"publicJwk"`
	PrivateJWK jwk.JWK `json:"privateJwk"`
}

// HDWalletGo is a pure Go HD wallet based on Trust Wallet Core implementation
type HDWalletGo struct {
	seed       []byte
	mnemonic   string
	passphrase string
}

// HDNode represents BIP32 hierarchical deterministic node
type HDNode struct {
	privateKey  []byte
	publicKey   []byte
	chainCode   []byte
	depth       uint8
	childNum    uint32
	fingerprint uint32
}

// NewHDWalletFromMnemonic creates HD wallet from mnemonic (simulates Trust Wallet Core implementation)
func NewHDWalletFromMnemonic(mnemonic, passphrase string) (*HDWalletGo, error) {
	// Validate mnemonic
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic")
	}

	// Generate seed (64 bytes) - using standard BIP39 implementation here
	seed := bip39.NewSeed(mnemonic, passphrase)

	return &HDWalletGo{
		seed:       seed,
		mnemonic:   mnemonic,
		passphrase: passphrase,
	}, nil
}

// getMasterNode generates master node from seed (simulates Trust Wallet Core's hdnode_from_seed)
func (w *HDWalletGo) getMasterNode() (*HDNode, error) {
	// BIP32 master key generation
	// Use "ed25519 seed" as HMAC-SHA512 key
	h := hmac.New(sha512.New, []byte("ed25519 seed"))
	h.Write(w.seed)
	hash := h.Sum(nil)

	// First 32 bytes as private key, last 32 bytes as chain code
	privateKey := hash[:32]
	chainCode := hash[32:]

	// Generate public key
	publicKey := ed25519.NewKeyFromSeed(privateKey).Public().(ed25519.PublicKey)

	return &HDNode{
		privateKey:  privateKey,
		publicKey:   publicKey,
		chainCode:   chainCode,
		depth:       0,
		childNum:    0,
		fingerprint: 0,
	}, nil
}

// GetMasterKeyEd25519 gets Ed25519 master key (simulates Trust Wallet Core's getMasterKey)
func (w *HDWalletGo) GetMasterKeyEd25519() (ed25519.PrivateKey, ed25519.PublicKey, error) {
	node, err := w.getMasterNode()
	if err != nil {
		return nil, nil, err
	}

	// Generate complete Ed25519 private key from private key seed
	privateKey := ed25519.NewKeyFromSeed(node.privateKey)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	return privateKey, publicKey, nil
}

// GetPrivateJWKTrustWalletCore generates private JWK using Trust Wallet Core compatible method
func (w *HDWalletGo) GetPrivateJWKTrustWalletCore() (*DIDKeyResult, error) {
	// Get Ed25519 key pair
	privateKey, publicKey, err := w.GetMasterKeyEd25519()
	if err != nil {
		return nil, fmt.Errorf("failed to get master key: %w", err)
	}

	// Create DID
	did, err := createDIDFromPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create DID: %w", err)
	}

	// Create key ID
	keyId := fmt.Sprintf("%s#z%s", did, base58.Encode(append(varint.ToUvarint(ED25519_CODEC_ID), publicKey...)))

	// Create public JWK
	publicJWK := jwk.JWK{
		ALG: "EdDSA",
		CRV: "Ed25519",
		KID: keyId,
		KTY: "OKP",
		USE: "sig",
		X:   base64.RawURLEncoding.EncodeToString(publicKey),
	}

	// Create private JWK
	privateJWK := jwk.JWK{
		ALG: "EdDSA",
		CRV: "Ed25519",
		KID: keyId,
		KTY: "OKP",
		USE: "sig",
		X:   base64.RawURLEncoding.EncodeToString(publicKey),
		D:   base64.RawURLEncoding.EncodeToString(privateKey), // Use complete 64-byte private key
	}

	return &DIDKeyResult{
		DID:        did,
		PublicJWK:  publicJWK,
		PrivateJWK: privateJWK,
	}, nil
}

// createDIDFromPublicKey creates a DID:key from Ed25519 public key
func createDIDFromPublicKey(publicKey ed25519.PublicKey) (string, error) {
	// Create multicodec identifier for Ed25519
	codecBytes := varint.ToUvarint(ED25519_CODEC_ID)

	// Combine codec + public key
	idBytes := make([]byte, len(codecBytes)+len(publicKey))
	copy(idBytes, codecBytes)
	copy(idBytes[len(codecBytes):], publicKey)

	// Encode with base58btc
	id := base58.Encode(idBytes)

	return fmt.Sprintf("did:key:z%s", id), nil
}

// GetPrivateJWK convenience function: generate private JWK from mnemonic
func GetPrivateJWK(mnemonic string) (*DIDKeyResult, error) {
	wallet, err := NewHDWalletFromMnemonic(mnemonic, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create HD wallet: %w", err)
	}

	return wallet.GetPrivateJWKTrustWalletCore()
}

// GetDID convenience function: generate DID from mnemonic
func GetDID(mnemonic string) (string, error) {
	result, err := GetPrivateJWK(mnemonic)
	if err != nil {
		return "", err
	}
	return result.DID, nil
}

// GetPublicJWK convenience function: generate public JWK from mnemonic
func GetPublicJWK(mnemonic string) (*jwk.JWK, error) {
	result, err := GetPrivateJWK(mnemonic)
	if err != nil {
		return nil, err
	}
	return &result.PublicJWK, nil
}

// GenerateMnemonic generates new BIP39 mnemonic
func GenerateMnemonic() string {
	// Generate 128 bits of entropy for 12-word mnemonic
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate entropy for mnemonic: %v", err))
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		panic(fmt.Sprintf("Failed to generate mnemonic from entropy: %v", err))
	}

	return mnemonic
}
