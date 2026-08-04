package meshinagent

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

const (
	// PlaceholderCertDir is the pod-local path for bootstrap TLS material.
	PlaceholderCertDir = "/tmp/olares/mesh-in-placeholder"
	// PlaceholderValidDays is the self-signed leaf lifetime.
	PlaceholderValidDays = 90
	placeholderCN        = "mesh-in-placeholder.olares.local"
	placeholderOrg       = "Olares"
	placeholderOU        = "mesh-in-placeholder"
)

// EnsurePlaceholderCert writes a pod-local ECDSA P-256 self-signed leaf when
// tls.crt/tls.key are missing or empty. Existing non-empty files are reused.
func EnsurePlaceholderCert(dir string) error {
	if dir == "" {
		dir = PlaceholderCertDir
	}
	certPath := filepath.Join(dir, "tls.crt")
	keyPath := filepath.Join(dir, "tls.key")
	if certOK, keyOK := fileNonEmpty(certPath), fileNonEmpty(keyPath); certOK && keyOK {
		return nil
	}

	certPEM, keyPEM, err := GeneratePlaceholderCertPEM(time.Now(), PlaceholderValidDays)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir placeholder cert dir: %w", err)
	}
	tmpCert := certPath + ".tmp"
	tmpKey := keyPath + ".tmp"
	if err := os.WriteFile(tmpCert, certPEM, 0o644); err != nil {
		return fmt.Errorf("write placeholder cert tmp: %w", err)
	}
	if err := os.WriteFile(tmpKey, keyPEM, 0o600); err != nil {
		_ = os.Remove(tmpCert)
		return fmt.Errorf("write placeholder key tmp: %w", err)
	}
	if err := os.Rename(tmpCert, certPath); err != nil {
		_ = os.Remove(tmpCert)
		_ = os.Remove(tmpKey)
		return fmt.Errorf("rename placeholder cert: %w", err)
	}
	if err := os.Rename(tmpKey, keyPath); err != nil {
		_ = os.Remove(tmpKey)
		return fmt.Errorf("rename placeholder key: %w", err)
	}
	return nil
}

// GeneratePlaceholderCertPEM creates a self-signed leaf (not a CA).
func GeneratePlaceholderCertPEM(now time.Time, validDays int) (certPEM, keyPEM []byte, err error) {
	if validDays <= 0 {
		validDays = PlaceholderValidDays
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate ecdsa key: %w", err)
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, fmt.Errorf("serial: %w", err)
	}
	notBefore := now.Add(-5 * time.Minute)
	notAfter := now.Add(time.Duration(validDays) * 24 * time.Hour)
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:         placeholderCN,
			Organization:       []string{placeholderOrg},
			OrganizationalUnit: []string{placeholderOU},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
		DNSNames: []string{
			placeholderCN,
			"*.olares.com",
		},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, nil, fmt.Errorf("create certificate: %w", err)
	}
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal key: %w", err)
	}
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	return certPEM, keyPEM, nil
}

func fileNonEmpty(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.Size() > 0
}
