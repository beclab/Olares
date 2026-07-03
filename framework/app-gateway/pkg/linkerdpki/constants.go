// Package linkerdpki provides the Linkerd identity issuer rotation and
// monitoring logic used by the app-gateway PKI guardian. The algorithm is
// migrated verbatim from cli/pkg/terminus without semantic change.
package linkerdpki

import "time"

const (
	// DefaultLinkerdNamespace is the namespace hosting the Linkerd control
	// plane and the PKI Secrets the guardian maintains.
	DefaultLinkerdNamespace = "linkerd"

	// PKISecretName is the single source of truth Secret storing ca.* and
	// issuer.* material for issuer rotation; access is restricted via RBAC.
	PKISecretName = "olares-linkerd-pki"

	// IssuerRotateThreshold triggers rotation when the issuer's remaining
	// validity drops below 6 months.
	IssuerRotateThreshold = 180 * 24 * time.Hour

	// IssuerLifetimeDays is the validity (3 years) of a freshly signed issuer.
	IssuerLifetimeDays = 1095
)

const (
	identityIssuerSecret  = "linkerd-identity-issuer"
	identityDeployment    = "linkerd-identity"
	identityIssuerCrtKey  = "crt.pem"
	identityIssuerKeyKey  = "key.pem"

	pkiCACrtKey     = "ca.crt"
	pkiCAKeyKey     = "ca.key"
	pkiIssuerCrtKey = "issuer.crt"
	pkiIssuerKeyKey = "issuer.key"
	pkiMetadataKey  = "metadata.json"
)
