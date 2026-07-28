// Self-contained helper for the mesh-in-agent image build context.
// Keep behavior aligned with pkg/gateway/meshinagent.EnsurePlaceholderCert.
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultDir = "/tmp/olares/mesh-in-placeholder"
	validDays  = 90
	cn         = "mesh-in-placeholder.olares.local"
	org        = "Olares"
	ou         = "mesh-in-placeholder"
)

func main() {
	dir := flag.String("dir", defaultDir, "directory for tls.crt and tls.key")
	flag.Parse()
	if err := ensure(*dir); err != nil {
		fmt.Fprintf(os.Stderr, "ensure-placeholder-cert: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("mesh-in-agent: placeholder cert ensured (validDays=%d) under %s\n", validDays, *dir)
}

func ensure(dir string) error {
	certPath := filepath.Join(dir, "tls.crt")
	keyPath := filepath.Join(dir, "tls.key")
	if nonEmpty(certPath) && nonEmpty(keyPath) {
		return nil
	}
	certPEM, keyPEM, err := generate(time.Now())
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmpCert, tmpKey := certPath+".tmp", keyPath+".tmp"
	if err := os.WriteFile(tmpCert, certPEM, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(tmpKey, keyPEM, 0o600); err != nil {
		_ = os.Remove(tmpCert)
		return err
	}
	if err := os.Rename(tmpCert, certPath); err != nil {
		_ = os.Remove(tmpCert)
		_ = os.Remove(tmpKey)
		return err
	}
	if err := os.Rename(tmpKey, keyPath); err != nil {
		_ = os.Remove(tmpKey)
		return err
	}
	return nil
}

func generate(now time.Time) ([]byte, []byte, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName:         cn,
			Organization:       []string{org},
			OrganizationalUnit: []string{ou},
		},
		NotBefore:             now.Add(-5 * time.Minute),
		NotAfter:              now.Add(validDays * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
		DNSNames:              []string{cn, "*.olares.com"},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}),
		nil
}

func nonEmpty(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.Size() > 0
}
