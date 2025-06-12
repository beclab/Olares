package crypto

import (
	"fmt"

	"github.com/beclab/Olares/cli/pkg/web5/crypto/dsa"
	"github.com/beclab/Olares/cli/pkg/web5/jwk"
)

// KeyManager is an abstraction that can be leveraged to manage/use keys (create, sign etc) as desired per the given use case
// examples of concrete implementations include: AWS KMS, Azure Key Vault, Google Cloud KMS, Hashicorp Vault etc
type KeyManager interface {
	// GeneratePrivateKey generates a new private key, stores it in the key store and returns the key id
	GeneratePrivateKey(algorithmID string) (string, error)

	// GetPublicKey returns the public key for the given key id
	GetPublicKey(keyID string) (jwk.JWK, error)

	// Sign signs the given payload with the private key for the given key id
	Sign(keyID string, payload []byte) ([]byte, error)
}

// KeyExporter is an abstraction that can be leveraged to implement types which intend to export keys
type KeyExporter interface {
	ExportKey(keyID string) (jwk.JWK, error)
}

// KeyImporter is an abstraction that can be leveraged to implement types which intend to import keys
type KeyImporter interface {
	ImportKey(key jwk.JWK) (string, error)
}

// LocalKeyManager is an implementation of KeyManager that stores keys in memory
type LocalKeyManager struct {
	keys map[string]jwk.JWK
}

// NewLocalKeyManager returns a new instance of InMemoryKeyManager
func NewLocalKeyManager() *LocalKeyManager {
	return &LocalKeyManager{
		keys: make(map[string]jwk.JWK),
	}
}

// GeneratePrivateKey generates a new private key using the algorithm provided,
// stores it in the key store and returns the key id
// Supported algorithms are available in [olares/github.com/beclab/Olares/cli/pkg/web5/crypto/dsa.AlgorithmID]
func (k *LocalKeyManager) GeneratePrivateKey(algorithmID string) (string, error) {
	var keyAlias string

	key, err := dsa.GeneratePrivateKey(algorithmID)
	if err != nil {
		return "", fmt.Errorf("failed to generate private key: %w", err)
	}

	keyAlias, err = key.ComputeThumbprint()
	if err != nil {
		return "", fmt.Errorf("failed to compute key alias: %w", err)
	}

	k.keys[keyAlias] = key

	return keyAlias, nil
}

// GetPublicKey returns the public key for the given key id
func (k *LocalKeyManager) GetPublicKey(keyID string) (jwk.JWK, error) {
	key, err := k.getPrivateJWK(keyID)
	if err != nil {
		return jwk.JWK{}, err
	}

	return dsa.GetPublicKey(key), nil

}

// Sign signs the payload with the private key for the given key id
func (k *LocalKeyManager) Sign(keyID string, payload []byte) ([]byte, error) {
	key, err := k.getPrivateJWK(keyID)
	if err != nil {
		return nil, err
	}

	return dsa.Sign(payload, key)
}

func (k *LocalKeyManager) getPrivateJWK(keyID string) (jwk.JWK, error) {
	key, ok := k.keys[keyID]

	if !ok {
		return jwk.JWK{}, fmt.Errorf("key with alias %s not found", keyID)
	}

	return key, nil
}

// ExportKey exports the key specific by the key ID from the [LocalKeyManager]
func (k *LocalKeyManager) ExportKey(keyID string) (jwk.JWK, error) {
	key, err := k.getPrivateJWK(keyID)
	if err != nil {
		return jwk.JWK{}, err
	}

	return key, nil
}

// ImportKey imports the key into the [LocalKeyManager] and returns the key alias
func (k *LocalKeyManager) ImportKey(key jwk.JWK) (string, error) {
	keyAlias, err := key.ComputeThumbprint()
	if err != nil {
		return "", fmt.Errorf("failed to compute key alias: %w", err)
	}

	k.keys[keyAlias] = key

	return keyAlias, nil
}
