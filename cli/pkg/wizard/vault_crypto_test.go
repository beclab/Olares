package wizard

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"testing"

	"golang.org/x/crypto/pbkdf2"
)

// buildTestAccount mints a brand-new RSA-backed account whose secrets
// are sealed with the given password — exactly the shape Account.Unlock
// expects on the wire.
func buildTestAccount(t *testing.T, password string) *Account {
	t.Helper()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa keygen: %v", err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatalf("marshal pub: %v", err)
	}
	privDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("marshal PKCS8: %v", err)
	}
	signingKey := generateRandomBytes(32)

	salt := generateRandomBytes(16)
	iter := 1000 // smaller for tests
	masterKey := pbkdf2.Key([]byte(password), salt, iter, 32, sha256.New)

	secrets := AccountSecrets{
		PrivateKey: Base64Bytes(privDER),
		SigningKey: Base64Bytes(signingKey),
	}
	plain, err := json.Marshal(secrets)
	if err != nil {
		t.Fatalf("marshal secrets: %v", err)
	}

	iv := generateRandomBytes(16)
	aad := generateRandomBytes(16)
	ciphertext, err := aesGCMEncrypt(masterKey, plain, iv, aad)
	if err != nil {
		t.Fatalf("encrypt secrets: %v", err)
	}

	return &Account{
		ID:        "acc-1",
		DID:       "did:test:1",
		Name:      "tester",
		PublicKey: base64.RawURLEncoding.EncodeToString(pubDER),
		KeyParams: KeyParams{
			Algorithm:  "PBKDF2",
			Hash:       "SHA-256",
			KeySize:    256,
			Iterations: iter,
			Salt:       base64.RawURLEncoding.EncodeToString(salt),
			Version:    "4.0.0",
		},
		EncryptionParams: EncryptionParams{
			Algorithm:      "AES-GCM",
			TagSize:        128,
			KeySize:        256,
			IV:             base64.RawURLEncoding.EncodeToString(iv),
			AdditionalData: base64.RawURLEncoding.EncodeToString(aad),
			Kind:           "r",
			Version:        "4.0.0",
		},
		EncryptedData: base64.RawURLEncoding.EncodeToString(ciphertext),
	}
}

func TestAccountUnlock_RoundTrip(t *testing.T) {
	const password = "correct horse battery staple"
	acc := buildTestAccount(t, password)

	unlocked, err := acc.Unlock(password)
	if err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	if len(unlocked.PrivateKey) == 0 || len(unlocked.SigningKey) == 0 {
		t.Fatalf("expected non-empty secrets, got %d/%d", len(unlocked.PrivateKey), len(unlocked.SigningKey))
	}
	if _, err := parseRSAPrivateKeyDER(unlocked.PrivateKey); err != nil {
		t.Fatalf("private key did not round-trip: %v", err)
	}
}

func TestAccountUnlock_WrongPassword(t *testing.T) {
	acc := buildTestAccount(t, "right")
	if _, err := acc.Unlock("wrong"); err == nil {
		t.Fatalf("expected unlock with wrong password to fail")
	}
}

func TestVault_RoundTrip_UpdateAccessors_Commit_Unlock(t *testing.T) {
	acc := buildTestAccount(t, "pw")
	unlocked, err := acc.Unlock("pw")
	if err != nil {
		t.Fatalf("unlock account: %v", err)
	}

	v := &Vault{
		Kind:  "vault",
		ID:    "v-1",
		Name:  "main",
		Owner: acc.ID,
	}
	if err := v.UpdateAccessors([]*UnlockedAccount{unlocked}); err != nil {
		t.Fatalf("UpdateAccessors: %v", err)
	}
	if len(v.Accessors) != 1 || v.Accessors[0].ID != acc.ID {
		t.Fatalf("expected single accessor for %s, got %+v", acc.ID, v.Accessors)
	}

	v.AddItems(VaultItem{ID: "item-1", Name: "TOTP", Type: VaultTypeTerminusTotp, Fields: []Field{{Name: "code", Type: FieldTypeTotp, Value: "JBSWY3DPEHPK3PXP"}}})
	if err := v.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	if len(v.EncryptedData) == 0 {
		t.Fatalf("expected EncryptedData after Commit")
	}

	// Simulate a fresh fetch from the server: clear runtime state, then
	// Unlock should restore the items.
	cold := *v
	cold.aesKey = nil
	cold.items = nil

	if err := cold.Unlock(unlocked); err != nil {
		t.Fatalf("Unlock: %v", err)
	}
	got := cold.ItemsCollection()
	if len(got.Items) != 1 {
		t.Fatalf("expected 1 item after Unlock, got %d", len(got.Items))
	}
	if got.Items["item-1"].Name != "TOTP" {
		t.Fatalf("unexpected item: %+v", got.Items["item-1"])
	}
}

func TestVaultItemCollection_MergePreservesLocalChanges(t *testing.T) {
	local := NewVaultItemCollection()
	local.Update(VaultItem{ID: "a", Name: "local-a"})

	remote := NewVaultItemCollection()
	remote.Items["a"] = VaultItem{ID: "a", Name: "remote-a"}
	remote.Items["b"] = VaultItem{ID: "b", Name: "remote-b"}

	local.Merge(remote)

	if local.Items["a"].Name != "local-a" {
		t.Fatalf("locally-changed item should win, got %q", local.Items["a"].Name)
	}
	if local.Items["b"].Name != "remote-b" {
		t.Fatalf("expected remote-only item to be added, got %+v", local.Items["b"])
	}
}

func TestVaultItemCollection_RoundTrip(t *testing.T) {
	c := NewVaultItemCollection()
	c.Update(VaultItem{ID: "x", Name: "X"}, VaultItem{ID: "y", Name: "Y"})
	bytes, err := c.ToBytes()
	if err != nil {
		t.Fatalf("ToBytes: %v", err)
	}
	out := NewVaultItemCollection()
	if err := out.FromBytes(bytes); err != nil {
		t.Fatalf("FromBytes: %v", err)
	}
	if len(out.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(out.Items))
	}
	if out.Items["x"].Name != "X" || out.Items["y"].Name != "Y" {
		t.Fatalf("unexpected items after round-trip: %+v", out.Items)
	}
}

func TestDirKVStorage_PutGetListDelete(t *testing.T) {
	dir := t.TempDir()
	s, err := NewDirKVStorage(dir)
	if err != nil {
		t.Fatalf("NewDirKVStorage: %v", err)
	}

	in := map[string]string{"hello": "world"}
	if err := s.Put("test", "k1", in); err != nil {
		t.Fatalf("Put: %v", err)
	}

	out := map[string]string{}
	if err := s.Get("test", "k1", &out); err != nil {
		t.Fatalf("Get: %v", err)
	}
	if out["hello"] != "world" {
		t.Fatalf("expected world, got %q", out["hello"])
	}

	ids, err := s.List("test")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(ids) != 1 || ids[0] != "k1" {
		t.Fatalf("expected [k1], got %v", ids)
	}

	if err := s.Delete("test", "k1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	ids, _ = s.List("test")
	if len(ids) != 0 {
		t.Fatalf("expected empty list after delete, got %v", ids)
	}
}
