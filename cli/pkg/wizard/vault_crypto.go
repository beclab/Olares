package wizard

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

// hashFromAlgo returns a hash.Hash factory for the given SDK algo name.
// Only SHA-256 is supported (matches the TS SDK defaults).
func hashFromAlgo(name string) (func() hash.Hash, error) {
	if name == "" || name == "SHA-256" {
		return sha256.New, nil
	}
	return nil, fmt.Errorf("unsupported hash: %s", name)
}

// decodeBase64Loose accepts both standard and URL-safe base64
// (with or without padding); matches the lenient parsing in the TS server.
func decodeBase64Loose(s string) ([]byte, error) {
	if s == "" {
		return nil, nil
	}
	if b, err := base64.URLEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	if b, err := base64.RawURLEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	if b, err := base64.StdEncoding.DecodeString(s); err == nil {
		return b, nil
	}
	return base64.RawStdEncoding.DecodeString(s)
}

// Unlock derives the master key from `password` using PBKDF2 (account.KeyParams),
// AES-GCM-decrypts the account secrets blob, and returns a non-nil
// UnlockedAccount carrying the cleartext private/signing keys.
//
// Mirrors apps/packages/sdk/src/core/account.ts Account.unlock.
func (acc *Account) Unlock(password string) (*UnlockedAccount, error) {
	if acc == nil {
		return nil, fmt.Errorf("account is nil")
	}
	if acc.EncryptedData == "" {
		return nil, fmt.Errorf("account has no encrypted data — cannot unlock")
	}
	hashFn, err := hashFromAlgo(acc.KeyParams.Hash)
	if err != nil {
		return nil, err
	}

	salt, err := decodeBase64Loose(acc.KeyParams.Salt)
	if err != nil {
		return nil, fmt.Errorf("invalid keyParams.salt: %w", err)
	}
	iv, err := decodeBase64Loose(acc.EncryptionParams.IV)
	if err != nil {
		return nil, fmt.Errorf("invalid encryptionParams.iv: %w", err)
	}
	aad, err := decodeBase64Loose(acc.EncryptionParams.AdditionalData)
	if err != nil {
		return nil, fmt.Errorf("invalid encryptionParams.additionalData: %w", err)
	}
	ciphertext, err := decodeBase64Loose(acc.EncryptedData)
	if err != nil {
		return nil, fmt.Errorf("invalid encryptedData: %w", err)
	}

	keyLen := acc.KeyParams.KeySize / 8
	if keyLen == 0 {
		keyLen = 32
	}
	masterKey := pbkdf2.Key([]byte(password), salt, acc.KeyParams.Iterations, keyLen, hashFn)
	plaintext, err := aesGCMDecrypt(masterKey, ciphertext, iv, aad)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt account secrets (wrong password?): %w", err)
	}

	var secrets AccountSecrets
	if err := json.Unmarshal(plaintext, &secrets); err != nil {
		return nil, fmt.Errorf("failed to parse account secrets: %w", err)
	}

	return &UnlockedAccount{
		Account:    acc,
		MasterKey:  masterKey,
		PrivateKey: []byte(secrets.PrivateKey),
		SigningKey: []byte(secrets.SigningKey),
	}, nil
}

// CopySecrets transfers in-memory secrets and master key from `other` so
// that a freshly-fetched Account can be re-used without re-deriving them
// (mirrors Account.copySecrets in TS).
func (acc *Account) CopySecrets(unlocked *UnlockedAccount) {
	if unlocked == nil {
		return
	}
	// no-op for the dehydrated server account; the unlocked struct already
	// carries the secrets out of band, so callers should keep the
	// UnlockedAccount around in AppState.
	_ = acc
}

// =============================================================================
// Vault crypto
// =============================================================================

// Unlock decrypts the vault's shared key using the unlocked account's RSA
// private key, then AES-GCM-decrypts the items blob. Mirrors
// apps/packages/sdk/src/core/vault.ts Vault.unlock.
func (v *Vault) Unlock(unlocked *UnlockedAccount) error {
	if unlocked == nil || unlocked.Account == nil {
		return fmt.Errorf("vault.Unlock requires an unlocked account")
	}
	if v.aesKey != nil {
		return nil // already unlocked
	}

	if len(v.Accessors) == 0 {
		// Fresh vault: bootstrap with this account as the sole accessor.
		if err := v.UpdateAccessors([]*UnlockedAccount{unlocked}); err != nil {
			return err
		}
		return v.Commit()
	}

	var accessor *Accessor
	for i := range v.Accessors {
		if v.Accessors[i].ID == unlocked.Account.ID {
			accessor = &v.Accessors[i]
			break
		}
	}
	if accessor == nil || len(accessor.EncryptedKey) == 0 {
		return fmt.Errorf("no accessor entry for account %s", unlocked.Account.ID)
	}

	sharedKey, err := rsaOAEPDecrypt(unlocked.PrivateKey, []byte(accessor.EncryptedKey))
	if err != nil {
		return fmt.Errorf("failed to unwrap shared key: %w", err)
	}
	v.aesKey = sharedKey

	if len(v.EncryptedData) > 0 {
		iv, err := decodeBase64Loose(v.EncryptionParams.IV)
		if err != nil {
			return fmt.Errorf("invalid vault.encryptionParams.iv: %w", err)
		}
		aad, err := decodeBase64Loose(v.EncryptionParams.AdditionalData)
		if err != nil {
			return fmt.Errorf("invalid vault.encryptionParams.additionalData: %w", err)
		}
		plain, err := aesGCMDecrypt(sharedKey, []byte(v.EncryptedData), iv, aad)
		if err != nil {
			return fmt.Errorf("failed to decrypt vault data: %w", err)
		}
		coll := NewVaultItemCollection()
		if err := coll.FromBytes(plain); err != nil {
			return fmt.Errorf("failed to parse vault items: %w", err)
		}
		v.items = coll
	} else {
		v.items = NewVaultItemCollection()
	}
	return nil
}

// UpdateAccessors generates a fresh shared AES key, re-encrypts any
// existing data with it, and wraps the new key with each subject's RSA
// public key. Mirrors SharedContainer.updateAccessors in TS.
func (v *Vault) UpdateAccessors(subjects []*UnlockedAccount) error {
	var existing []byte
	if len(v.EncryptedData) > 0 {
		if v.aesKey == nil {
			return fmt.Errorf("non-empty vault must be unlocked before updating accessors")
		}
		iv, err := decodeBase64Loose(v.EncryptionParams.IV)
		if err != nil {
			return fmt.Errorf("invalid vault.encryptionParams.iv: %w", err)
		}
		aad, err := decodeBase64Loose(v.EncryptionParams.AdditionalData)
		if err != nil {
			return fmt.Errorf("invalid vault.encryptionParams.additionalData: %w", err)
		}
		existing, err = aesGCMDecrypt(v.aesKey, []byte(v.EncryptedData), iv, aad)
		if err != nil {
			return fmt.Errorf("failed to decrypt existing vault data: %w", err)
		}
	}

	v.aesKey = generateAESKey()

	if v.KeyParams.Algorithm == "" {
		v.KeyParams = NewRSAEncryptionParams()
	}
	if v.EncryptionParams.Algorithm == "" {
		v.EncryptionParams = EncryptionParams{
			Algorithm: "AES-GCM",
			TagSize:   128,
			KeySize:   256,
		}
	}

	if existing != nil {
		if err := v.setEncryptedData(existing); err != nil {
			return err
		}
	}

	v.Accessors = make([]Accessor, 0, len(subjects))
	for _, subj := range subjects {
		if subj == nil || subj.Account == nil {
			continue
		}
		pubDER, err := decodeBase64Loose(subj.Account.PublicKey)
		if err != nil {
			return fmt.Errorf("invalid public key for %s: %w", subj.Account.ID, err)
		}
		wrapped, err := rsaOAEPEncrypt(pubDER, v.aesKey)
		if err != nil {
			return fmt.Errorf("failed to wrap shared key for %s: %w", subj.Account.ID, err)
		}
		v.Accessors = append(v.Accessors, Accessor{
			ID:           subj.Account.ID,
			PublicKey:    Base64Bytes(pubDER),
			EncryptedKey: Base64Bytes(wrapped),
		})
	}
	return nil
}

// Commit re-encrypts the in-memory item collection into EncryptedData.
// Mirrors Vault.commit in TS.
func (v *Vault) Commit() error {
	if v.aesKey == nil {
		return fmt.Errorf("vault must be unlocked before commit")
	}
	if v.items == nil {
		v.items = NewVaultItemCollection()
	}
	plain, err := v.items.ToBytes()
	if err != nil {
		return err
	}
	return v.setEncryptedData(plain)
}

// setEncryptedData generates a fresh IV and AAD, then AES-GCM-encrypts.
func (v *Vault) setEncryptedData(plaintext []byte) error {
	iv := generateRandomBytes(16)
	aad := generateRandomBytes(16)
	ciphertext, err := aesGCMEncrypt(v.aesKey, plaintext, iv, aad)
	if err != nil {
		return err
	}
	v.EncryptionParams.IV = base64.URLEncoding.EncodeToString(iv)
	v.EncryptionParams.AdditionalData = base64.URLEncoding.EncodeToString(aad)
	v.EncryptionParams.Algorithm = "AES-GCM"
	if v.EncryptionParams.TagSize == 0 {
		v.EncryptionParams.TagSize = 128
	}
	if v.EncryptionParams.KeySize == 0 {
		v.EncryptionParams.KeySize = 256
	}
	v.EncryptedData = Base64Bytes(ciphertext)
	return nil
}

// Merge takes the items, name, accessors, etc from `other` while
// preserving locally-changed items. Mirrors Vault.merge in TS.
func (v *Vault) Merge(other *Vault) {
	if other == nil {
		return
	}
	if v.items == nil {
		v.items = NewVaultItemCollection()
	}
	v.items.Merge(other.items)
	if other.Name != "" {
		v.Name = other.Name
	}
	v.Revision = other.Revision
	v.Org = other.Org
	v.Accessors = other.Accessors
	if other.aesKey != nil {
		v.aesKey = other.aesKey
	}
	v.EncryptedData = other.EncryptedData
	v.EncryptionParams = other.EncryptionParams
	v.KeyParams = other.KeyParams
	if other.Updated != "" {
		v.Updated = other.Updated
	}
}

// AddItems inserts items into the vault's in-memory collection (the
// caller must subsequently Commit + push to the server).
func (v *Vault) AddItems(items ...VaultItem) {
	if v.items == nil {
		v.items = NewVaultItemCollection()
	}
	v.items.Update(items...)
}

// MarkSynced clears all change records up to `cutoff` (or all of them if
// the zero time is passed).
func (v *Vault) MarkSynced(cutoff time.Time) {
	if v.items != nil {
		v.items.ClearChanges(cutoff)
	}
}
