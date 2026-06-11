package terminus

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

func TestIssuerNeedsRotation(t *testing.T) {
	now := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)

	issuer180, err := pemCertWithNotAfter(now.Add(179 * 24 * time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	need, _, err := issuerNeedsRotation(issuer180, now)
	if err != nil {
		t.Fatal(err)
	}
	if !need {
		t.Fatal("expected rotation when remaining < 180 days")
	}

	issuer200, err := pemCertWithNotAfter(now.Add(200 * 24 * time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	need, _, err = issuerNeedsRotation(issuer200, now)
	if err != nil {
		t.Fatal(err)
	}
	if need {
		t.Fatal("expected no rotation when remaining >= 180 days")
	}
}

func TestCertificateNotAfter(t *testing.T) {
	end := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	pemBytes, err := pemCertWithNotAfter(end)
	if err != nil {
		t.Fatal(err)
	}
	got, err := certificateNotAfter(pemBytes)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(end) {
		t.Fatalf("notAfter: got %v want %v", got, end)
	}
}

func TestRotateLinkerdIssuer(t *testing.T) {
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	caTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "root.linkerd.cluster.local"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour * 365 * 30),
		KeyUsage:     x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		IsCA:         true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	caCrtPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})
	caKeyDER, err := x509.MarshalECPrivateKey(caKey)
	if err != nil {
		t.Fatal(err)
	}
	caKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: caKeyDER})

	mat, err := rotateLinkerdIssuer(caCrtPEM, caKeyPEM)
	if err != nil {
		t.Fatal(err)
	}
	need, _, err := issuerNeedsRotation(mat.IssuerCrt, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if need {
		t.Fatal("new issuer should be valid for >= 6 months")
	}
}

func pemCertWithNotAfter(notAfter time.Time) ([]byte, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test.linkerd.cluster.local"},
		NotBefore:    notAfter.Add(-24 * time.Hour),
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), nil
}
