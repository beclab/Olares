package did

import (
	"github.com/beclab/Olares/cli/pkg/web5/dids/didcore"
	"github.com/beclab/Olares/cli/pkg/web5/jwk"
)

// PortableDID is a serializable BearerDID. VerificationMethod contains the private key
// of each verification method that the BearerDID's key manager contains
type PortableDID struct {
	// URI is the DID string as per https://www.w3.org/TR/did-core/#did-syntax
	URI string `json:"uri"`
	// PrivateKeys is an array of private keys associated to the BearerDID's verification methods
	// Note: PrivateKeys will be empty if the BearerDID was created using a KeyManager that does not
	// support exporting private keys (e.g. HSM based KeyManagers)
	PrivateKeys []jwk.JWK `json:"privateKeys"`
	// Document is the DID Document associated to the BearerDID
	Document didcore.Document `json:"document"`
	// Metadata is a map that can be used to store additional method specific data
	// that is necessary to inflate a BearerDID from a PortableDID
	Metadata map[string]interface{} `json:"metadata"`
}
