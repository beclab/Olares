package didkey

import (
	"context"
	"fmt"

	"github.com/beclab/Olares/cli/pkg/web5/crypto"
	"github.com/beclab/Olares/cli/pkg/web5/crypto/dsa"
	"github.com/beclab/Olares/cli/pkg/web5/dids/did"
	"github.com/beclab/Olares/cli/pkg/web5/dids/didcore"
	"github.com/beclab/Olares/cli/pkg/web5/jwk"
)

// createOptions is a struct that contains all options that can be passed to [Create]
type createOptions struct {
	keyManager  crypto.KeyManager
	algorithmID string
}

// CreateOption is a type returned by all [Create] options for variadic parameter support
type CreateOption func(o *createOptions)

// KeyManager is an option that can be passed to Create to provide a KeyManager
func KeyManager(k crypto.KeyManager) CreateOption {
	return func(o *createOptions) {
		o.keyManager = k
	}
}

// AlgorithmID is an option that can be passed to Create to specify a specific
// cryptographic algorithm to use to generate the private key
func AlgorithmID(id string) CreateOption {
	return func(o *createOptions) {
		o.algorithmID = id
	}
}

// Create can be used to create a new `did:jwk`. `did:jwk` is useful in scenarios where:
//   - Offline resolution is preferred
//   - Key rotation is not required
//   - Service endpoints are not necessary
//
// Spec: https://github.com/quartzjer/did-jwk/blob/main/spec.md
func Create(opts ...CreateOption) (did.BearerDID, error) {
	o := createOptions{
		keyManager:  crypto.NewLocalKeyManager(),
		algorithmID: dsa.AlgorithmIDED25519,
	}

	for _, opt := range opts {
		opt(&o)
	}

	keyMgr := o.keyManager

	keyID, err := keyMgr.GeneratePrivateKey(o.algorithmID)
	if err != nil {
		return did.BearerDID{}, fmt.Errorf("failed to generate private key: %w", err)
	}

	publicJWK, _ := keyMgr.GetPublicKey(keyID)

	id, err := KeyToID(publicJWK)
	if err != nil {
		return did.BearerDID{}, fmt.Errorf("failed to convert public key to ID: %w", err)
	}

	publicJWK.KID = "did:key:" + id
	publicJWK.USE = "sig"
	//	publicJWK.ALG = "EdDSA"

	didJWK := did.DID{
		Method: "key",
		URI:    "did:key:" + id,
		ID:     id,
	}

	bearerDID := did.BearerDID{
		DID:        didJWK,
		KeyManager: keyMgr,
		Document:   createDocument(didJWK, publicJWK),
	}

	return bearerDID, nil
}

// Resolver is a type to implement resolution
type Resolver struct{}

// ResolveWithContext the provided DID URI (must be a did:jwk) as per the wee bit of detail provided in the
// spec: https://github.com/quartzjer/did-jwk/blob/main/spec.md
func (r Resolver) ResolveWithContext(ctx context.Context, uri string) (didcore.ResolutionResult, error) {
	return r.Resolve(uri)
}

// Resolve the provided DID URI (must be a did:jwk) as per the wee bit of detail provided in the
// spec: https://github.com/quartzjer/did-jwk/blob/main/spec.md
func (r Resolver) Resolve(uri string) (didcore.ResolutionResult, error) {
	//
	return didcore.ResolutionResult{}, nil
}

// createDocument creates a DID document from a DID URI
func createDocument(did did.DID, publicKey jwk.JWK) didcore.Document {
	// Create the base document

	doc := didcore.Document{
		Context: []string{
			"https://www.w3.org/ns/did/v1",
			"https://w3id.org/security/suites/ed25519-2020/v1",
			"https://w3id.org/security/suites/x25519-2020/v1",
		},
		ID: did.URI,
	}

	// Create the key ID
	keyID := fmt.Sprintf("#%s", did.ID)

	// Create the verification method
	vm := didcore.VerificationMethod{
		ID:           keyID,
		Type:         "Ed25519VerificationKey2020",
		Controller:   did.URI,
		PublicKeyJwk: &publicKey,
	}

	// Add the verification method with all purposes
	doc.AddVerificationMethod(
		vm,
		didcore.Purposes(
			"assertionMethod",
			"authentication",
			"capabilityDelegation",
			"capabilityInvocation",
		),
	)

	return doc
}
