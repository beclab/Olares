package tlsutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"
)

func TestValidateKeyPair(t *testing.T) {
	certPEM, keyPEM := mustGenerateKeyPair(t)

	t.Run("valid pair", func(t *testing.T) {
		if err := ValidateKeyPair(certPEM, keyPEM); err != nil {
			t.Fatalf("ValidateKeyPair: %v", err)
		}
	})

	t.Run("placeholder strings like test/test are rejected", func(t *testing.T) {
		if err := ValidateKeyPair("test", "test"); err == nil {
			t.Fatal("expected error for placeholder cert/key")
		}
	})

	t.Run("empty rejected", func(t *testing.T) {
		if err := ValidateKeyPair("", keyPEM); err == nil {
			t.Fatal("expected error for empty cert")
		}
		if err := ValidateKeyPair(certPEM, ""); err == nil {
			t.Fatal("expected error for empty key")
		}
	})

	t.Run("mismatched key rejected", func(t *testing.T) {
		_, otherKey := mustGenerateKeyPair(t)
		if err := ValidateKeyPair(certPEM, otherKey); err == nil {
			t.Fatal("expected error for mismatched key")
		}
	})
}

func mustGenerateKeyPair(t *testing.T) (certPEM, keyPEM string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test.example.com"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	certPEM = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
	keyBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	keyPEM = string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}))
	return certPEM, keyPEM
}
