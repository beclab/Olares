package utils

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"
)

// ValidateKeyPair checks that certPEM and keyPEM form a parseable, matching TLS key pair.
func ValidateKeyPair(certPEM, keyPEM string) error {
	if certPEM == "" || keyPEM == "" {
		return fmt.Errorf("empty cert or key")
	}
	if _, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM)); err != nil {
		return fmt.Errorf("invalid TLS key pair: %w", err)
	}
	return nil
}

// CheckSSLCertificate checks the validity of an SSL certificate and private key for a given hostname.
func CheckSSLCertificate(cert, key []byte, hostname string) error {
	if err := ValidateKeyPair(string(cert), string(key)); err != nil {
		return err
	}
	block, _ := pem.Decode(cert)
	if block == nil {
		return errors.New("certificate is invalid")
	}
	pub, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return errors.New("certificate is invalid")
	}
	// verify hostname
	err = pub.VerifyHostname(hostname)
	if err != nil {
		return err
	}

	// verify certificate whether valid or expired
	currentTime := time.Now()
	if currentTime.Before(pub.NotBefore) {
		return errors.New("certificate is not yet valid")
	}
	if currentTime.After(pub.NotAfter) {
		return errors.New("certificate has expired")
	}

	return nil
}
