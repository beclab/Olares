package tlsutil

import (
	"crypto/tls"
	"fmt"
)

// ValidateKeyPair checks that certPEM/keyPEM form a TLS certificate chain
// Envoy can load (parseable PEMs and a matching private key).
func ValidateKeyPair(certPEM, keyPEM string) error {
	if certPEM == "" || keyPEM == "" {
		return fmt.Errorf("empty cert or key")
	}
	if _, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM)); err != nil {
		return fmt.Errorf("invalid TLS key pair: %w", err)
	}
	return nil
}
