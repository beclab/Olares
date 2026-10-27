package meshinagent

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGeneratePlaceholderCertPEM(t *testing.T) {
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	certPEM, keyPEM, err := GeneratePlaceholderCertPEM(now, PlaceholderValidDays)
	if err != nil {
		t.Fatal(err)
	}
	block, _ := pem.Decode(certPEM)
	if block == nil {
		t.Fatal("cert pem decode failed")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if cert.Subject.CommonName != placeholderCN {
		t.Fatalf("CN=%q", cert.Subject.CommonName)
	}
	if len(cert.Subject.Organization) == 0 || cert.Subject.Organization[0] != placeholderOrg {
		t.Fatalf("O=%v", cert.Subject.Organization)
	}
	if len(cert.Subject.OrganizationalUnit) == 0 || cert.Subject.OrganizationalUnit[0] != placeholderOU {
		t.Fatalf("OU=%v", cert.Subject.OrganizationalUnit)
	}
	if cert.IsCA {
		t.Fatal("placeholder must not be a CA")
	}
	wantAfter := now.Add(PlaceholderValidDays * 24 * time.Hour)
	if cert.NotAfter.Sub(wantAfter) > time.Second || wantAfter.Sub(cert.NotAfter) > time.Second {
		t.Fatalf("NotAfter=%v want ~%v", cert.NotAfter, wantAfter)
	}
	if cert.NotBefore.After(now) {
		t.Fatalf("NotBefore=%v should be <= now (skew)", cert.NotBefore)
	}
	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil || keyBlock.Type != "EC PRIVATE KEY" {
		t.Fatalf("key pem = %#v", keyBlock)
	}
}

func TestEnsurePlaceholderCertIdempotent(t *testing.T) {
	dir := t.TempDir()
	if err := EnsurePlaceholderCert(dir); err != nil {
		t.Fatal(err)
	}
	cert1, err := os.ReadFile(filepath.Join(dir, "tls.crt"))
	if err != nil {
		t.Fatal(err)
	}
	key1, err := os.ReadFile(filepath.Join(dir, "tls.key"))
	if err != nil {
		t.Fatal(err)
	}
	if err := EnsurePlaceholderCert(dir); err != nil {
		t.Fatal(err)
	}
	cert2, err := os.ReadFile(filepath.Join(dir, "tls.crt"))
	if err != nil {
		t.Fatal(err)
	}
	key2, err := os.ReadFile(filepath.Join(dir, "tls.key"))
	if err != nil {
		t.Fatal(err)
	}
	if string(cert1) != string(cert2) || string(key1) != string(key2) {
		t.Fatal("second ensure must reuse existing files")
	}
}
